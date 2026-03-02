package repository

import (
	"context"
	"fmt"
	"time"

	models "github.com/arigatory/sentinel/internal/model"
	"github.com/arigatory/sentinel/pkg/retry"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Storage = (*PostgresStorage)(nil)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{
		pool: pool,
	}
}

func (s *PostgresStorage) UpdateCounter(name string, delta int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO counters (name, delta, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (name)
		DO UPDATE SET
			delta = counters.delta + EXCLUDED.delta,
			updated_at = CURRENT_TIMESTAMP
	`

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		_, err := s.pool.Exec(ctx, query, name, delta)
		return err
	})

	return err
}

func (s *PostgresStorage) UpdateGauge(name string, value float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := `
		INSERT INTO gauges (name, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (name)
		DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = CURRENT_TIMESTAMP
	`

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		_, err := s.pool.Exec(ctx, query, name, value)
		return err
	})

	return err
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var delta int64

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		return s.pool.QueryRow(ctx, "SELECT delta FROM counters WHERE name = $1", name).Scan(&delta)
	})

	if err != nil {
		return 0, false
	}

	return delta, true
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var value float64

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		return s.pool.QueryRow(ctx, "SELECT value FROM gauges WHERE name = $1", name).Scan(&value)
	})

	if err != nil {
		return 0, false
	}

	return value, true
}

func (s *PostgresStorage) GetAllGauges() (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result := make(map[string]float64)

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		rows, err := s.pool.Query(ctx, "SELECT name, value FROM gauges")
		if err != nil {
			return fmt.Errorf("error querying gauges: %w", err)
		}
		defer rows.Close()

		result = make(map[string]float64)

		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err != nil {
				return fmt.Errorf("error scanning gauge: %w", err)
			}
			result[name] = value
		}

		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all gauges: %w", err)
	}

	return result, nil
}

func (s *PostgresStorage) GetAllCounters() (map[string]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result := make(map[string]int64)

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		rows, err := s.pool.Query(ctx, "SELECT name, delta FROM counters")
		if err != nil {
			return fmt.Errorf("error querying counters: %w", err)
		}
		defer rows.Close()

		result = make(map[string]int64)

		for rows.Next() {
			var name string
			var delta int64
			if err := rows.Scan(&name, &delta); err != nil {
				return fmt.Errorf("error scanning counter: %w", err)
			}
			result[name] = delta
		}

		return rows.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get all counters: %w", err)
	}

	return result, nil
}

func (s *PostgresStorage) Save(path string) error {
	return nil
}

func (s *PostgresStorage) Load(path string) error {
	return nil
}

func (s *PostgresStorage) UpdateBatch(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewPostgresErrorClassifier()

	err := retry.Do(ctx, cfg, func() error {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		counterQuery := `
			INSERT INTO counters (name, delta, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (name)
			DO UPDATE SET
				delta = counters.delta + EXCLUDED.delta,
				updated_at = CURRENT_TIMESTAMP
		`

		gaugeQuery := `
			INSERT INTO gauges (name, value, updated_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (name)
			DO UPDATE SET
				value = EXCLUDED.value,
				updated_at = CURRENT_TIMESTAMP
		`

		for _, metric := range metrics {
			switch metric.MType {
			case models.Counter:
				if metric.Delta == nil {
					continue
				}
				_, err = tx.Exec(ctx, counterQuery, metric.ID, *metric.Delta)
				if err != nil {
					return fmt.Errorf("failed to update counter %s: %w", metric.ID, err)
				}
			case models.Gauge:
				if metric.Value == nil {
					continue
				}
				_, err = tx.Exec(ctx, gaugeQuery, metric.ID, *metric.Value)
				if err != nil {
					return fmt.Errorf("failed to update gauge %s: %w", metric.ID, err)
				}
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update batch after retries: %w", err)
	}

	return nil
}
