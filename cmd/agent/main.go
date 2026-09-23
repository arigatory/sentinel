package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/arigatory/sentinel/internal/agent"
	models "github.com/arigatory/sentinel/internal/model"
)

func main() {
	cfg := parseFlags()

	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	storage := agent.NewMetricsStorage()
	// Буфер по числу воркеров: репортёр отдаёт задание и возвращается к своему
	// тику, не дожидаясь освобождения конкретного воркера.
	jobs := make(chan []models.Metrics, cfg.RateLimit)

	log.Printf("Agent started. Server: %s, Poll: %ds, Report: %ds, Rate limit: %d",
		cfg.Address, cfg.PollInterval, cfg.ReportInterval, cfg.RateLimit)

	// Пул отправителей: одновременно в сеть уходит не более cfg.RateLimit запросов.
	workers := agent.StartWorkers(jobs, cfg.Address, cfg.Key, cfg.RateLimit)

	// Сбор метрик и их отправка разнесены по разным горутинам.
	var producers sync.WaitGroup
	producers.Add(3)

	go func() {
		defer producers.Done()
		agent.PollRuntime(ctx, storage, pollInterval)
	}()

	go func() {
		defer producers.Done()
		agent.PollSystem(ctx, storage, pollInterval)
	}()

	go func() {
		defer producers.Done()
		agent.Report(ctx, storage, jobs, reportInterval)
	}()

	// Report закроет jobs, после чего воркеры доработают остаток очереди.
	producers.Wait()
	workers.Wait()

	log.Println("Agent stopped")
}
