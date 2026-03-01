package service

import (
	"errors"
	"log"

	"github.com/arigatory/sentinel/internal/repository"
)

var (
	ErrMetricNotFound = errors.New("metric not found")
)

type MetricsService struct {
	storage   repository.Storage
	filePath  string
	syncWrite bool
}

func NewMetricsService(storage repository.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) ConfigurePersistence(filePath string, syncWrite bool) {
	s.filePath = filePath
	s.syncWrite = syncWrite
}

func (s *MetricsService) save() {
	if s.filePath == "" {
		return
	}
	if err := s.storage.Save(s.filePath); err != nil {
		log.Printf("Error saving metrics to file: %v", err)
	}
}

func (s *MetricsService) UpdateCounter(name string, delta int64) {
	s.storage.UpdateCounter(name, delta)
	if s.syncWrite {
		s.save()
	}
}

func (s *MetricsService) UpdateGauge(name string, value float64) {
	s.storage.UpdateGauge(name, value)
	if s.syncWrite {
		s.save()
	}
}

func (s *MetricsService) Save() {
	s.save()
}

func (s *MetricsService) Load() error {
	if s.filePath == "" {
		return nil
	}
	return s.storage.Load(s.filePath)
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
