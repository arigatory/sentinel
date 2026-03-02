package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/arigatory/sentinel/internal/config/db"
	customMiddleware "github.com/arigatory/sentinel/internal/middleware"
	models "github.com/arigatory/sentinel/internal/model"
	"github.com/arigatory/sentinel/internal/repository"
	"github.com/arigatory/sentinel/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type server struct {
	metricsService *service.MetricsService
	db             *db.DB
}

func (s *server) updateHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	metricName := chi.URLParam(req, "name")
	metricValue := chi.URLParam(req, "value")

	if metricType == "" {
		http.Error(res, "Metric type is required", http.StatusBadRequest)
		return
	}

	if metricType != models.Counter && metricType != models.Gauge {
		http.Error(res, "Metric type must be 'counter' or 'gauge'", http.StatusBadRequest)
		return
	}

	if metricType == models.Counter {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(res, "Invalid counter value", http.StatusBadRequest)
			return
		}
		log.Printf("Counter %s updated by %d", metricName, value)
		s.metricsService.UpdateCounter(metricName, value)
	}

	if metricType == models.Gauge {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		log.Printf("Gauge %s set to %g", metricName, value)
		s.metricsService.UpdateGauge(metricName, value)
	}

	log.Printf("Type: %s, Name: %s, Value: %s", metricType, metricName, metricValue)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func (s *server) valueHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	metricName := chi.URLParam(req, "name")

	if metricType == models.Gauge {
		value, err := s.metricsService.GetGauge(metricName)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.Error(res, "Gauge not found", http.StatusNotFound)
				return
			}
			http.Error(res, "Internal error", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%g", value)
		return
	}

	if metricType == models.Counter {
		value, err := s.metricsService.GetCounter(metricName)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.Error(res, "Counter not found", http.StatusNotFound)
				return
			}
			http.Error(res, "Internal error", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%d", value)
		return
	}

	http.Error(res, "Unknown metric type", http.StatusBadRequest)
}

func (s *server) updateJSONHandler(res http.ResponseWriter, req *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case models.Counter:
		if m.Delta == nil {
			http.Error(res, "delta is required for counter", http.StatusBadRequest)
			return
		}
		s.metricsService.UpdateCounter(m.ID, *m.Delta)
		updated, _ := s.metricsService.GetCounter(m.ID)
		m.Delta = &updated
	case models.Gauge:
		if m.Value == nil {
			http.Error(res, "value is required for gauge", http.StatusBadRequest)
			return
		}
		s.metricsService.UpdateGauge(m.ID, *m.Value)
	default:
		http.Error(res, "unknown metric type", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(m); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *server) valueJSONHandler(res http.ResponseWriter, req *http.Request) {
	var m models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&m); err != nil {
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}

	switch m.MType {
	case models.Gauge:
		value, err := s.metricsService.GetGauge(m.ID)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.Error(res, "Gauge not found", http.StatusNotFound)
				return
			}
			http.Error(res, "Internal error", http.StatusInternalServerError)
			return
		}
		m.Value = &value
	case models.Counter:
		value, err := s.metricsService.GetCounter(m.ID)
		if err != nil {
			if errors.Is(err, service.ErrMetricNotFound) {
				http.Error(res, "Counter not found", http.StatusNotFound)
				return
			}
			http.Error(res, "Internal error", http.StatusInternalServerError)
			return
		}
		m.Delta = &value
	default:
		http.Error(res, "unknown metric type", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(m); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *server) rootHandler(res http.ResponseWriter, req *http.Request) {
	gauges, counters := s.metricsService.GetAllMetrics()

	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)

	fmt.Fprintln(res, "<!DOCTYPE html>")
	fmt.Fprintln(res, "<html>")
	fmt.Fprintln(res, "<head><title>Metrics</title></head>")
	fmt.Fprintln(res, "<body>")
	fmt.Fprintln(res, "<h1>All Metrics</h1>")

	fmt.Fprintln(res, "<h2>Gauges</h2>")
	fmt.Fprintln(res, "<ul>")
	for name, value := range gauges {
		fmt.Fprintf(res, "<li>%s: %g</li>\n", name, value)
	}
	fmt.Fprintln(res, "</ul>")

	fmt.Fprintln(res, "<h2>Counters</h2>")
	fmt.Fprintln(res, "<ul>")
	for name, value := range counters {
		fmt.Fprintf(res, "<li>%s: %d</li>\n", name, value)
	}
	fmt.Fprintln(res, "</ul>")

	fmt.Fprintln(res, "</body>")
	fmt.Fprintln(res, "</html>")
}

func (s *server) pingHandler(res http.ResponseWriter, req *http.Request) {
	if s.db == nil {
		http.Error(res, "Database not configured", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		log.Printf("Database ping failed: %v", err)
		http.Error(res, "Database connection failed", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func (s *server) updateBatchHandler(res http.ResponseWriter, req *http.Request) {
	var metrics []models.Metrics
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&metrics); err != nil {
		http.Error(res, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode([]models.Metrics{})
		return
	}

	// Валидация метрик
	for i, m := range metrics {
		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				http.Error(res, fmt.Sprintf("delta is required for counter at index %d", i), http.StatusBadRequest)
				return
			}
		case models.Gauge:
			if m.Value == nil {
				http.Error(res, fmt.Sprintf("value is required for gauge at index %d", i), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, fmt.Sprintf("unknown metric type at index %d", i), http.StatusBadRequest)
			return
		}
	}

	// Обновляем все метрики в одной транзакции
	if err := s.metricsService.UpdateBatch(metrics); err != nil {
		log.Printf("Error updating batch: %v", err)
		http.Error(res, "Failed to update metrics", http.StatusInternalServerError)
		return
	}

	// Обновляем значения для ответа (для счетчиков возвращаем текущее значение)
	for i, m := range metrics {
		if m.MType == models.Counter {
			updated, _ := s.metricsService.GetCounter(m.ID)
			metrics[i].Delta = &updated
		}
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(metrics); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func main() {
	cfg := parseFlags()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	var storage repository.Storage
	var database *db.DB
	var metricsService *service.MetricsService

	// Логика выбора хранилища:
	// 1. PostgreSQL (если есть DATABASE_DSN)
	// 2. File storage (если есть FILE_STORAGE_PATH)
	// 3. Memory storage (fallback)

	if cfg.DatabaseDSN != "" {
		// Пытаемся подключиться к PostgreSQL
		var err error
		database, err = db.NewDB(cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Warning: failed to connect to database: %v", err)
			log.Println("Falling back to file or memory storage")
		} else {
			log.Println("Successfully connected to PostgreSQL")

			// Выполняем миграции
			if err := database.RunMigrations(cfg.DatabaseDSN); err != nil {
				log.Printf("Warning: failed to run migrations: %v", err)
				log.Println("Falling back to file or memory storage")
				database.Close()
				database = nil
			} else {
				log.Println("Database migrations completed successfully")
				// Используем PostgreSQL storage
				storage = repository.NewPostgresStorage(database.Pool())
				log.Println("Using PostgreSQL storage for metrics")
				defer database.Close()
			}
		}
	}

	// Если PostgreSQL недоступен, используем Memory storage
	if storage == nil {
		storage = repository.NewMemStorage()
		log.Println("Using in-memory storage for metrics")
	}

	metricsService = service.NewMetricsService(storage)

	// Настраиваем персистентность для file storage (только если не используем PostgreSQL)
	if database == nil && cfg.FileStoragePath != "" {
		syncWrite := cfg.StoreInterval == 0
		metricsService.ConfigurePersistence(cfg.FileStoragePath, syncWrite)

		if cfg.Restore {
			if err := metricsService.Load(); err != nil {
				log.Printf("Warning: could not load metrics from %s: %v", cfg.FileStoragePath, err)
			} else {
				log.Printf("Metrics loaded from %s", cfg.FileStoragePath)
			}
		}
	}

	srv := &server{
		metricsService: metricsService,
		db:             database,
	}

	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)
	r.Use(customMiddleware.Logger(logger))
	r.Use(customMiddleware.GzipMiddleware)

	r.Post("/updates", srv.updateBatchHandler)
	r.Post("/updates/", srv.updateBatchHandler)
	r.Post("/update", srv.updateJSONHandler)
	r.Post("/value", srv.valueJSONHandler)
	r.Post("/update/{type}/{name}/{value}", srv.updateHandler)
	r.Get("/value/{type}/{name}", srv.valueHandler)
	r.Get("/ping", srv.pingHandler)
	r.Get("/", srv.rootHandler)

	httpServer := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Периодическое сохранение в файл (только если не используем PostgreSQL)
	if database == nil && cfg.FileStoragePath != "" && cfg.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					metricsService.Save()
					log.Printf("Metrics saved to %s", cfg.FileStoragePath)
				case <-ctx.Done():
					log.Println("Stopping periodic save goroutine")
					return
				}
			}
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		log.Printf("Starting metrics server on %s (read: 5s, write: 10s, idle: 60s)", cfg.Address)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal: %v, initiating graceful shutdown", sig)

	cancel()

	// Финальное сохранение в файл (только если не используем PostgreSQL)
	if database == nil && cfg.FileStoragePath != "" {
		metricsService.Save()
		log.Printf("Final metrics save to %s completed", cfg.FileStoragePath)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}
}
