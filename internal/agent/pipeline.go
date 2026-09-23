package agent

import (
	"context"
	"log"
	"sync"
	"time"

	models "github.com/arigatory/sentinel/internal/model"
)

// PollRuntime периодически опрашивает runtime и складывает метрики в хранилище.
// Работает до отмены ctx.
func PollRuntime(ctx context.Context, storage *MetricsStorage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			storage.CollectRuntimeMetrics()
		case <-ctx.Done():
			return
		}
	}
}

// PollSystem периодически собирает метрики операционной системы через gopsutil.
// Работает независимо от PollRuntime, в своей горутине.
func PollSystem(ctx context.Context, storage *MetricsStorage, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := storage.CollectSystemMetrics(ctx); err != nil {
				log.Printf("Error collecting system metrics: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Report раз в interval снимает срез метрик и отдаёт его воркерам одним заданием.
//
// Отправкой занимаются воркеры, поэтому репортер никогда не ходит в сеть сам.
// При завершении закрывает jobs — это сигнал воркерам остановиться.
func Report(ctx context.Context, storage *MetricsStorage, jobs chan<- []models.Metrics, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(jobs)

	for {
		select {
		case <-ticker.C:
			metrics := storage.Snapshot()
			if len(metrics) == 0 {
				continue
			}

			log.Printf("Sending %d metrics to server...", len(metrics))

			select {
			case jobs <- metrics:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// StartWorkers поднимает пул из rateLimit воркеров — столько одновременно
// исходящих батч-запросов агент допускает максимум. Воркеры живут, пока не
// закроют jobs, и дорабатывают уже принятые задания.
func StartWorkers(jobs <-chan []models.Metrics, serverAddr, key string, rateLimit int) *sync.WaitGroup {
	var wg sync.WaitGroup

	for i := 1; i <= rateLimit; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			for batch := range jobs {
				if err := SendMetricsBatch(serverAddr, batch, key); err != nil {
					log.Printf("Worker %d: failed to send batch of %d metrics: %v", id, len(batch), err)
				}
			}
		}(i)
	}

	return &wg
}
