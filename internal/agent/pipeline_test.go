package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	models "github.com/arigatory/sentinel/internal/model"
)

// TestStartWorkersRespectsRateLimit — главная проверка инкремента: сколько бы
// заданий ни стояло в очереди, одновременно в полёте не больше rateLimit запросов.
func TestStartWorkersRespectsRateLimit(t *testing.T) {
	const (
		rateLimit  = 3
		totalJobs  = 30
		handlerLag = 20 * time.Millisecond
	)

	var (
		inFlight int64
		maxSeen  int64
		handled  int64
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt64(&inFlight, 1)
		for {
			seen := atomic.LoadInt64(&maxSeen)
			if current <= seen || atomic.CompareAndSwapInt64(&maxSeen, seen, current) {
				break
			}
		}

		// держим соединение, чтобы запросы реально пересекались во времени
		time.Sleep(handlerLag)

		atomic.AddInt64(&inFlight, -1)
		atomic.AddInt64(&handled, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	addr := strings.TrimPrefix(server.URL, "http://")

	jobs := make(chan models.Metrics)
	workers := StartWorkers(jobs, addr, "", rateLimit)

	for i := 0; i < totalJobs; i++ {
		v := float64(i)
		jobs <- models.Metrics{ID: "Gauge", MType: models.Gauge, Value: &v}
	}
	close(jobs)
	workers.Wait()

	if got := atomic.LoadInt64(&handled); got != totalJobs {
		t.Errorf("Expected %d requests to reach the server, got %d", totalJobs, got)
	}
	if got := atomic.LoadInt64(&maxSeen); got > rateLimit {
		t.Errorf("Expected at most %d concurrent requests, observed %d", rateLimit, got)
	}
	if got := atomic.LoadInt64(&maxSeen); got < 2 {
		t.Errorf("Expected the pool to actually work in parallel, peak concurrency was %d", got)
	}
}

func TestReportSendsSnapshotAndClosesJobs(t *testing.T) {
	storage := NewMetricsStorage()
	storage.SetGauge("Alloc", 42)
	storage.AddCounter("PollCount", 1)

	jobs := make(chan models.Metrics, 16)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		Report(ctx, storage, jobs, 10*time.Millisecond)
		close(done)
	}()

	got := make(map[string]bool)
	for i := 0; i < 2; i++ {
		select {
		case m := <-jobs:
			got[m.ID] = true
		case <-time.After(2 * time.Second):
			t.Fatal("Timed out waiting for the reporter to emit metrics")
		}
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Report did not stop after context cancellation")
	}

	if _, ok := <-jobs; ok {
		// канал мог содержать остаток снимка — вычитываем до закрытия
		for range jobs {
		}
	}

	if !got["Alloc"] || !got["PollCount"] {
		t.Errorf("Expected both Alloc and PollCount in the snapshot, got %v", got)
	}
}

// TestStorageConcurrentAccess проверяет хранилище под -race в том режиме,
// в котором его использует агент: два сборщика и читатель одновременно.
func TestStorageConcurrentAccess(t *testing.T) {
	storage := NewMetricsStorage()
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			storage.CollectRuntimeMetrics()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			if err := storage.CollectSystemMetrics(ctx); err != nil {
				t.Errorf("CollectSystemMetrics failed: %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = storage.Snapshot()
			_ = storage.GetGauges()
			_ = storage.GetCounters()
		}
	}()

	wg.Wait()
}

// GetGauges должен отдавать копию, иначе читатель ловит гонку с коллектором.
func TestGetGaugesReturnsCopy(t *testing.T) {
	storage := NewMetricsStorage()
	storage.SetGauge("Alloc", 1)

	gauges := storage.GetGauges()
	gauges["Alloc"] = 999
	gauges["Injected"] = 1

	if got := storage.GetGauges()["Alloc"]; got != 1 {
		t.Errorf("Expected storage to keep 1, got %f", got)
	}
	if _, exists := storage.GetGauges()["Injected"]; exists {
		t.Error("Mutating the returned map must not affect the storage")
	}
}
