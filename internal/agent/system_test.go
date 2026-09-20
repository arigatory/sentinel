package agent

import (
	"context"
	"runtime"
	"strconv"
	"testing"
)

func TestCollectSystemMetrics(t *testing.T) {
	storage := NewMetricsStorage()

	if err := storage.CollectSystemMetrics(context.Background()); err != nil {
		t.Fatalf("CollectSystemMetrics failed: %v", err)
	}

	gauges := storage.GetGauges()

	for _, name := range []string{"TotalMemory", "FreeMemory"} {
		value, exists := gauges[name]
		if !exists {
			t.Errorf("Expected gauge %q to be collected", name)
			continue
		}
		if value <= 0 {
			t.Errorf("Expected gauge %q to be positive, got %f", name, value)
		}
	}

	// Число метрик загрузки должно совпадать с числом CPU в рантайме.
	want := runtime.NumCPU()
	for i := 1; i <= want; i++ {
		name := "CPUutilization" + strconv.Itoa(i)
		if _, exists := gauges[name]; !exists {
			t.Errorf("Expected gauge %q for %d CPUs", name, want)
		}
	}

	if _, exists := gauges["CPUutilization"+strconv.Itoa(want+1)]; exists {
		t.Errorf("Did not expect a gauge beyond CPUutilization%d", want)
	}
}
