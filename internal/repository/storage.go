package repository

import (
	models "github.com/arigatory/sentinel/internal/model"
)

type Storage interface {
	UpdateCounter(name string, delta int64) error

	UpdateGauge(name string, value float64) error

	GetCounter(name string) (int64, bool)

	GetGauge(name string) (float64, bool)

	GetAllGauges() (map[string]float64, error)

	GetAllCounters() (map[string]int64, error)

	UpdateBatch(metrics []models.Metrics) error

	Save(path string) error

	Load(path string) error
}
