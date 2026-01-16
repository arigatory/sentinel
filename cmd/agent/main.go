package main

import (
	"fmt"
	"time"

	"github.com/arigatory/sentinel/internal/agent"
)

func main() {
	storage := agent.NewMetricsStorage()

	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	fmt.Println("Starting agent...")
	fmt.Printf("Poll interval: %v\n", pollInterval)
	fmt.Printf("Report interval: %v\n", reportInterval)

	for {
		select {
		case <-pollTicker.C:
			fmt.Println("Collecting metrics...")
			storage.CollectRuntimeMetrics()
		case <-reportTicker.C:
			fmt.Println("Sending metrics...")
			storage.SendAllMetrics()
		}
	}
}
