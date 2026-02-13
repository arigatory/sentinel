package service

import (
	"errors"
	"testing"

	"github.com/arigatory/sentinel/internal/repository"
)

func TestMetricsService_UpdateCounter(t *testing.T) {
	storage := repository.NewMemStorage()
	service := NewMetricsService(storage)

	service.UpdateCounter("test", 10)
	service.UpdateCounter("test", 5)

	value, err := service.GetCounter("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value != 15 {
		t.Errorf("Expected counter value 15, got %d", value)
	}
}

func TestMetricsService_UpdateGauge(t *testing.T) {
	storage := repository.NewMemStorage()
	service := NewMetricsService(storage)

	service.UpdateGauge("temperature", 36.6)
	service.UpdateGauge("temperature", 37.2)

	value, err := service.GetGauge("temperature")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if value != 37.2 {
		t.Errorf("Expected gauge value 37.2, got %f", value)
	}
}

func TestMetricsService_GetCounter_NotFound(t *testing.T) {
	storage := repository.NewMemStorage()
	service := NewMetricsService(storage)

	_, err := service.GetCounter("nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, ErrMetricNotFound) {
		t.Errorf("Expected ErrMetricNotFound, got %v", err)
	}
}

func TestMetricsService_GetGauge_NotFound(t *testing.T) {
	storage := repository.NewMemStorage()
	service := NewMetricsService(storage)

	_, err := service.GetGauge("nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, ErrMetricNotFound) {
		t.Errorf("Expected ErrMetricNotFound, got %v", err)
	}
}

func TestMetricsService_GetAllMetrics(t *testing.T) {
	storage := repository.NewMemStorage()
	service := NewMetricsService(storage)

	service.UpdateGauge("temp", 36.6)
	service.UpdateGauge("memory", 1024.0)
	service.UpdateCounter("requests", 100)
	service.UpdateCounter("errors", 5)

	gauges, counters := service.GetAllMetrics()

	if len(gauges) != 2 {
		t.Errorf("Expected 2 gauges, got %d", len(gauges))
	}
	if gauges["temp"] != 36.6 {
		t.Errorf("Expected temp=36.6, got %f", gauges["temp"])
	}
	if gauges["memory"] != 1024.0 {
		t.Errorf("Expected memory=1024.0, got %f", gauges["memory"])
	}

	if len(counters) != 2 {
		t.Errorf("Expected 2 counters, got %d", len(counters))
	}
	if counters["requests"] != 100 {
		t.Errorf("Expected requests=100, got %d", counters["requests"])
	}
	if counters["errors"] != 5 {
		t.Errorf("Expected errors=5, got %d", counters["errors"])
	}
}
