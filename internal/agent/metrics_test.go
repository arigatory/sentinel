package agent

import (
	"testing"
)

func TestNewMetricsStorage(t *testing.T) {
	storage := NewMetricsStorage()

	if storage == nil {
		t.Fatal("Expected storage to be not nil")
	}
	if storage.gauges == nil {
		t.Error("Expected gauges map to be initialized")
	}
	if storage.counters == nil {
		t.Error("Expected counters map to be initialized")
	}
}

func TestMetricsStorage_SetGauge(t *testing.T) {
	storage := NewMetricsStorage()
	storage.SetGauge("temperature", 36.6)

	value, exists := storage.GetGauges()["temperature"]
	if !exists {
		t.Fatal("Gauge 'temperature' should exist")
	}
	if value != 36.6 {
		t.Errorf("Expected 36.6, got %f", value)
	}
}

func TestMetricsStorage_AddCounter(t *testing.T) {
	storage := NewMetricsStorage()

	storage.AddCounter("requests", 5)
	value, exists := storage.GetCounters()["requests"]
	if !exists {
		t.Fatal("Counter 'requests' should exist")
	}
	if value != 5 {
		t.Errorf("Expected 5, got %d", value)
	}

	storage.AddCounter("requests", 3)
	value, exists = storage.GetCounters()["requests"]
	if !exists {
		t.Fatal("Counter 'requests' should exist")
	}
	if value != 8 {
		t.Errorf("Expected 8 (5+3), got %d", value)
	}
}

func TestMetricsStorage_CollectRuntimeMetrics(t *testing.T) {
	storage := NewMetricsStorage()
	storage.CollectRuntimeMetrics()

	gauges := storage.GetGauges()
	expectedMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs",
	}

	for _, metric := range expectedMetrics {
		if _, exists := gauges[metric]; !exists {
			t.Errorf("Expected gauge '%s' to be collected", metric)
		}
	}
}

func TestMetricsStorage_GetGauges(t *testing.T) {
	storage := NewMetricsStorage()
	storage.SetGauge("temperature", 36.6)
	storage.SetGauge("humidity", 75.0)

	gauges := storage.GetGauges()

	if len(gauges) != 2 {
		t.Errorf("Expected 2 gauges, got %d", len(gauges))
	}

	if gauges["temperature"] != 36.6 {
		t.Errorf("Expected temperature 36.6, got %f", gauges["temperature"])
	}

	if gauges["humidity"] != 75.0 {
		t.Errorf("Expected humidity 75.0, got %f", gauges["humidity"])
	}
}

func TestMetricsStorage_GetCounters(t *testing.T) {
	storage := NewMetricsStorage()
	storage.AddCounter("requests", 5)
	storage.AddCounter("errors", 2)

	counters := storage.GetCounters()

	if len(counters) != 2 {
		t.Errorf("Expected 2 counters, got %d", len(counters))
	}

	if counters["requests"] != 5 {
		t.Errorf("Expected requests 5, got %d", counters["requests"])
	}

	if counters["errors"] != 2 {
		t.Errorf("Expected errors 2, got %d", counters["errors"])
	}
}
