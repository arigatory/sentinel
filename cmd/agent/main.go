package main

import (
	"log"
	"time"

	"github.com/arigatory/sentinel/internal/agent"
	models "github.com/arigatory/sentinel/internal/model"
)

func main() {
	cfg := parseFlags()

	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	storage := agent.NewMetricsStorage()

	log.Printf("Agent started. Server: %s, Poll: %ds, Report: %ds",
		cfg.Address, cfg.PollInterval, cfg.ReportInterval)

	for {
		select {
		case <-pollTicker.C:
			// Сбор метрик (включая PollCount и RandomValue)
			storage.CollectRuntimeMetrics()

		case <-reportTicker.C:
			// Отправка метрик
			log.Println("Sending metrics to server...")

			for name, value := range storage.GetGauges() {
				agent.SendMetric(cfg.Address, models.Gauge, name, value)
			}

			for name, value := range storage.GetCounters() {
				agent.SendMetric(cfg.Address, models.Counter, name, value)
			}
		}
	}
}
