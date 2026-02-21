package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	customMiddleware "github.com/arigatory/sentinel/internal/middleware"
	models "github.com/arigatory/sentinel/internal/model"
	"github.com/arigatory/sentinel/internal/repository"
	"github.com/arigatory/sentinel/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type server struct {
	metricsService *service.MetricsService
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

	// Сформировать HTML
	fmt.Fprintln(res, "<!DOCTYPE html>")
	fmt.Fprintln(res, "<html>")
	fmt.Fprintln(res, "<head><title>Metrics</title></head>")
	fmt.Fprintln(res, "<body>")
	fmt.Fprintln(res, "<h1>All Metrics</h1>")

	// Выводим gauges
	fmt.Fprintln(res, "<h2>Gauges</h2>")
	fmt.Fprintln(res, "<ul>")
	for name, value := range gauges {
		fmt.Fprintf(res, "<li>%s: %g</li>\n", name, value)
	}
	fmt.Fprintln(res, "</ul>")

	// counters аналогично
	fmt.Fprintln(res, "<h2>Counters</h2>")
	fmt.Fprintln(res, "<ul>")
	for name, value := range counters {
		fmt.Fprintf(res, "<li>%s: %d</li>\n", name, value)
	}
	fmt.Fprintln(res, "</ul>")

	fmt.Fprintln(res, "</body>")
	fmt.Fprintln(res, "</html>")
}

func main() {
	cfg := parseFlags()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	storage := repository.NewMemStorage()
	metricsService := service.NewMetricsService(storage)

	// Настраиваем персистентность
	if cfg.FileStoragePath != "" {
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

	srv := &server{metricsService: metricsService}

	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)
	r.Use(customMiddleware.Logger(logger))
	r.Use(customMiddleware.GzipMiddleware)

	r.Post("/update", srv.updateJSONHandler)
	r.Post("/value", srv.valueJSONHandler)
	r.Post("/update/{type}/{name}/{value}", srv.updateHandler)
	r.Get("/value/{type}/{name}", srv.valueHandler)
	r.Get("/", srv.rootHandler)

	httpServer := &http.Server{
		Addr:         cfg.Address,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Периодическое сохранение (если интервал > 0)
	if cfg.FileStoragePath != "" && cfg.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				metricsService.Save()
				log.Printf("Metrics saved to %s", cfg.FileStoragePath)
			}
		}()
	}

	log.Printf("Starting metrics server on %s (read: 5s, write: 10s, idle: 60s)", cfg.Address)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
