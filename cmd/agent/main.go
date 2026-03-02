package main

import (
	"log"
	"time"

	"github.com/arigatory/sentinel/internal/agent"
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
			storage.CollectRuntimeMetrics()

		case <-reportTicker.C:
			log.Println("Sending metrics to server...")

			if err := storage.SendAllMetricsBatch(cfg.Address); err != nil {
				log.Printf("Error sending metrics batch: %v", err)
			}
		}
	}
}
