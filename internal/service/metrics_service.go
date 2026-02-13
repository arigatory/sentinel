package service

import (
	"errors"

	"github.com/arigatory/sentinel/internal/repository"
)

var (
	ErrMetricNotFound = errors.New("metric not found")
)

type MetricsService struct {
	storage *repository.MemStorage
}

func NewMetricsService(storage *repository.MemStorage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) UpdateCounter(name string, delta int64) {
	s.storage.UpdateCounter(name, delta)
}

func (s *MetricsService) UpdateGauge(name string, value float64) {
	s.storage.UpdateGauge(name, value)
}

func (s *MetricsService) GetCounter(name string) (int64, error) {
	value, exists := s.storage.GetCounter(name)
	if !exists {
		return 0, ErrMetricNotFound
	}
	return value, nil
}

func (s *MetricsService) GetGauge(name string) (float64, error) {
	value, exists := s.storage.GetGauge(name)
	if !exists {
		return 0, ErrMetricNotFound
	}
	return value, nil
}

func (s *MetricsService) GetAllMetrics() (map[string]float64, map[string]int64) {
	gauges := s.storage.GetAllGauges()
	counters := s.storage.GetAllCounters()
	return gauges, counters
}
