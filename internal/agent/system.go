package agent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// CollectSystemMetrics снимает метрики операционной системы через gopsutil:
// общий и свободный объём памяти, а также загрузку каждого логического CPU.
//
// Счётчики загрузки именуются CPUutilization1..CPUutilizationN, где N —
// число CPU, определяемое во время исполнения.
func (m *MetricsStorage) CollectSystemMetrics(ctx context.Context) error {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to read virtual memory stats: %w", err)
	}

	m.SetGauge("TotalMemory", float64(vm.Total))
	m.SetGauge("FreeMemory", float64(vm.Free))

	// interval = 0 — загрузка считается относительно предыдущего вызова,
	// то есть за прошедший интервал опроса; percpu = true — по каждому CPU.
	utilization, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return fmt.Errorf("failed to read cpu utilization: %w", err)
	}

	for i, value := range utilization {
		m.SetGauge("CPUutilization"+strconv.Itoa(i+1), value)
	}

	return nil
}
