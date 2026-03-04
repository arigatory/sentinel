package service

import (
	"errors"
	"log"

	models "github.com/arigatory/sentinel/internal/model"
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

func (s *MetricsService) UpdateCounter(name string, delta int64) error {
	if err := s.storage.UpdateCounter(name, delta); err != nil {
		return err
	}
	if s.syncWrite {
		s.save()
	}
	return nil
}

func (s *MetricsService) UpdateGauge(name string, value float64) error {
	if err := s.storage.UpdateGauge(name, value); err != nil {
		return err
	}
	if s.syncWrite {
		s.save()
	}
	return nil
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

func (s *MetricsService) GetAllMetrics() (map[string]float64, map[string]int64, error) {
	gauges, err := s.storage.GetAllGauges()
	if err != nil {
		return nil, nil, err
	}
	counters, err := s.storage.GetAllCounters()
	if err != nil {
		return nil, nil, err
	}
	return gauges, counters, nil
}

func (s *MetricsService) UpdateBatch(metrics []models.Metrics) error {
	if err := s.storage.UpdateBatch(metrics); err != nil {
		return err
	}
	if s.syncWrite {
		s.save()
	}
	return nil
}
