package repository

import (
	"sync"

	models "github.com/arigatory/sentinel/internal/model"
)

var _ Storage = (*MemStorage)(nil)

type MemStorage struct {
	mu       sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (m *MemStorage) UpdateCounter(name string, delta int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
	return nil
}

func (m *MemStorage) UpdateGauge(name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
	return nil
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.counters[name]
	return value, exists
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.gauges[name]
	return value, exists
}

func (m *MemStorage) GetAllGauges() (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]float64, len(m.gauges))
	for name, value := range m.gauges {
		result[name] = value
	}
	return result, nil
}

func (m *MemStorage) GetAllCounters() (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]int64, len(m.counters))
	for name, value := range m.counters {
		result[name] = value
	}
	return result, nil
}

func (m *MemStorage) UpdateBatch(metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			if metric.Delta != nil {
				m.counters[metric.ID] += *metric.Delta
			}
		case models.Gauge:
			if metric.Value != nil {
				m.gauges[metric.ID] = *metric.Value
			}
		}
	}
	return nil
}
