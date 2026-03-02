package repository

import (
	"context"
	"fmt"
	"time"

	models "github.com/arigatory/sentinel/internal/model"
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

func (s *PostgresStorage) UpdateCounter(name string, delta int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO counters (name, delta, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (name)
		DO UPDATE SET
			delta = counters.delta + EXCLUDED.delta,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.pool.Exec(ctx, query, name, delta)
	if err != nil {
		fmt.Printf("Error updating counter %s: %v\n", name, err)
	}
}

func (s *PostgresStorage) UpdateGauge(name string, value float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO gauges (name, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (name)
		DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.pool.Exec(ctx, query, name, value)
	if err != nil {
		fmt.Printf("Error updating gauge %s: %v\n", name, err)
	}
}

func (s *PostgresStorage) GetCounter(name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var delta int64
	err := s.pool.QueryRow(ctx, "SELECT delta FROM counters WHERE name = $1", name).Scan(&delta)
	if err != nil {
		return 0, false
	}

	return delta, true
}

func (s *PostgresStorage) GetGauge(name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var value float64
	err := s.pool.QueryRow(ctx, "SELECT value FROM gauges WHERE name = $1", name).Scan(&value)
	if err != nil {
		return 0, false
	}

	return value, true
}

func (s *PostgresStorage) GetAllGauges() map[string]float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make(map[string]float64)

	rows, err := s.pool.Query(ctx, "SELECT name, value FROM gauges")
	if err != nil {
		fmt.Printf("Error getting all gauges: %v\n", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			fmt.Printf("Error scanning gauge: %v\n", err)
			continue
		}
		result[name] = value
	}

	return result
}

func (s *PostgresStorage) GetAllCounters() map[string]int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make(map[string]int64)

	rows, err := s.pool.Query(ctx, "SELECT name, delta FROM counters")
	if err != nil {
		fmt.Printf("Error getting all counters: %v\n", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var delta int64
		if err := rows.Scan(&name, &delta); err != nil {
			fmt.Printf("Error scanning counter: %v\n", err)
			continue
		}
		result[name] = delta
	}

	return result
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
}
