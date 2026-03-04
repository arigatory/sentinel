package retry

import (
	"context"
	"errors"
	"log"
	"time"
)

type ErrorClassification int

const (
	NonRetriable ErrorClassification = iota
	Retriable
)

type Classifier interface {
	Classify(err error) ErrorClassification
}

type Config struct {
	MaxAttempts int
	Delays      []time.Duration
	Classifier  Classifier
}

func DefaultConfig() Config {
	return Config{
		MaxAttempts: 4,
		Delays:      []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
		Classifier:  nil,
	}
}

func Do(ctx context.Context, cfg Config, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if cfg.Classifier != nil {
			classification := cfg.Classifier.Classify(err)
			if classification == NonRetriable {
				return err
			}
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		var delay time.Duration
		if attempt < len(cfg.Delays) {
			delay = cfg.Delays[attempt]
		} else {
			delay = cfg.Delays[len(cfg.Delays)-1]
		}

		log.Printf("Retry attempt %d/%d after error: %v (waiting %v)",
			attempt+1, cfg.MaxAttempts-1, err, delay)

		select {
		case <-ctx.Done():
			return errors.Join(lastErr, ctx.Err())
		case <-time.After(delay):
		}
	}

	return lastErr
}
