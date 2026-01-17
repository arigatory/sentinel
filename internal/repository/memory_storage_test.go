package repository

import "testing"

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()

	if storage == nil {
		t.Fatal("Expected storage to be not nil")
	}
	if storage.counters == nil {
		t.Error("Expected counters map to be initialized")
	}
	if storage.gauges == nil {
		t.Error("Expected gauges map to be initialized")
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	storage.UpdateCounter("requests", 5)
	value, exists := storage.GetCounter("requests")
	if !exists {
		t.Fatal("Counter 'requests' should exist")
	}
	if value != 5 {
		t.Errorf("Expected 5, got %d", value)
	}

	storage.UpdateCounter("requests", 3)
	value, exists = storage.GetCounter("requests")
	if !exists {
		t.Fatal("Counter 'requests' should exist")
	}
	if value != 8 {
		t.Errorf("Expected 8 (5+3), got %d", value)
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	storage.UpdateGauge("temperature", 36.6)
	value, exists := storage.GetGauge("temperature")
	if !exists {
		t.Fatal("Gauge 'temperature' should exist")
	}
	if value != 36.6 {
		t.Errorf("Expected 36.6, got %f", value)
	}

	storage.UpdateGauge("temperature", 37.2)
	value, exists = storage.GetGauge("temperature")
	if !exists {
		t.Fatal("Gauge 'temperature' should exist")
	}
	if value != 37.2 {
		t.Errorf("Expected 37.2, got %f", value)
	}
}

func TestMemStorage_GetCounter_NotFound(t *testing.T) {
	storage := NewMemStorage()
	_, exists := storage.GetCounter("nonexistent")
	if exists {
		t.Error("Expected counter 'nonexistent' to not exist")
	}
}

func TestMemStorage_GetGauge_NotFound(t *testing.T) {
	storage := NewMemStorage()
	_, exists := storage.GetGauge("nonexistent")
	if exists {
		t.Error("Expected gauge 'nonexistent' to not exist")
	}
}
