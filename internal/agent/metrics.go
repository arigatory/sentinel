package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/arigatory/sentinel/internal/hash"
	models "github.com/arigatory/sentinel/internal/model"
	"github.com/arigatory/sentinel/pkg/retry"
)

// MetricsStorage хранит собранные метрики. Все методы безопасны для
// одновременного вызова из нескольких горутин: сбор runtime-метрик,
// сбор системных метрик и отправка работают параллельно.
type MetricsStorage struct {
	mu       sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MetricsStorage) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

// GetGauges возвращает копию — она переживает последующие изменения хранилища.
func (m *MetricsStorage) GetGauges() map[string]float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	gauges := make(map[string]float64, len(m.gauges))
	for name, value := range m.gauges {
		gauges[name] = value
	}
	return gauges
}

func (m *MetricsStorage) AddCounter(name string, delta int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += delta
}

// GetCounters возвращает копию — она переживает последующие изменения хранилища.
func (m *MetricsStorage) GetCounters() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	counters := make(map[string]int64, len(m.counters))
	for name, delta := range m.counters {
		counters[name] = delta
	}
	return counters
}

// Snapshot возвращает все метрики одним согласованным срезом — его репортер
// раздаёт воркерам.
func (m *MetricsStorage) Snapshot() []models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make([]models.Metrics, 0, len(m.gauges)+len(m.counters))

	for name, value := range m.gauges {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &v,
		})
	}

	for name, delta := range m.counters {
		d := delta
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &d,
		})
	}

	return metrics
}

func (m *MetricsStorage) CollectRuntimeMetrics() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.SetGauge("Alloc", float64(rtm.Alloc))
	m.SetGauge("BuckHashSys", float64(rtm.BuckHashSys))
	m.SetGauge("Frees", float64(rtm.Frees))
	m.SetGauge("GCCPUFraction", rtm.GCCPUFraction)
	m.SetGauge("GCSys", float64(rtm.GCSys))
	m.SetGauge("HeapAlloc", float64(rtm.HeapAlloc))
	m.SetGauge("HeapIdle", float64(rtm.HeapIdle))
	m.SetGauge("HeapInuse", float64(rtm.HeapInuse))
	m.SetGauge("HeapObjects", float64(rtm.HeapObjects))
	m.SetGauge("HeapReleased", float64(rtm.HeapReleased))
	m.SetGauge("HeapSys", float64(rtm.HeapSys))
	m.SetGauge("LastGC", float64(rtm.LastGC))
	m.SetGauge("Lookups", float64(rtm.Lookups))
	m.SetGauge("MCacheInuse", float64(rtm.MCacheInuse))
	m.SetGauge("MCacheSys", float64(rtm.MCacheSys))
	m.SetGauge("MSpanInuse", float64(rtm.MSpanInuse))
	m.SetGauge("MSpanSys", float64(rtm.MSpanSys))
	m.SetGauge("Mallocs", float64(rtm.Mallocs))
	m.SetGauge("NextGC", float64(rtm.NextGC))
	m.SetGauge("NumForcedGC", float64(rtm.NumForcedGC))
	m.SetGauge("NumGC", float64(rtm.NumGC))
	m.SetGauge("OtherSys", float64(rtm.OtherSys))
	m.SetGauge("PauseTotalNs", float64(rtm.PauseTotalNs))
	m.SetGauge("StackInuse", float64(rtm.StackInuse))
	m.SetGauge("StackSys", float64(rtm.StackSys))
	m.SetGauge("Sys", float64(rtm.Sys))
	m.SetGauge("TotalAlloc", float64(rtm.TotalAlloc))

	m.AddCounter("PollCount", 1)
	m.SetGauge("RandomValue", rand.Float64())
}

// SendMetricsBatch отправляет срез метрик одним запросом на POST /updates/:
// JSON, сжатый gzip, с подписью HMAC-SHA256 и повторами при сетевых ошибках.
// Именно эту функцию вызывают воркеры пула.
func SendMetricsBatch(serverAddr string, metrics []models.Metrics, key string) error {
	if len(metrics) == 0 {
		return nil
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err = gw.Write(body); err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}
	if err = gw.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	url := fmt.Sprintf("http://%s/updates/", serverAddr)

	cfg := retry.DefaultConfig()
	cfg.Classifier = retry.NewNetworkErrorClassifier()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// подпись считается от несжатого тела, до gzip
	var signature string
	if key != "" {
		signature = hash.Sign(body, key)
	}

	err = retry.Do(ctx, cfg, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf.Bytes()))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		if key != "" {
			req.Header.Set(hash.Header, signature)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned error status: %s", resp.Status)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to send metrics batch after retries: %w", err)
	}

	return nil
}
