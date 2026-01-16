package agent

import (
	"math/rand"
	"runtime"
)

type MetricsStorage struct {
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
	m.gauges[name] = value
}

func (m *MetricsStorage) GetGauges() map[string]float64 {
	return m.gauges
}

func (m *MetricsStorage) AddCounter(name string, delta int64) {
	m.counters[name] += delta
}

func (m *MetricsStorage) GetCounters() map[string]int64 {
	return m.counters
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