package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arigatory/sentinel/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type server struct {
	storage *repository.MemStorage
}

func (s *server) updateHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	metricName := chi.URLParam(req, "name")
	metricValue := chi.URLParam(req, "value")

	if metricType == "" {
		http.Error(res, "Metric type is required", http.StatusBadRequest)
		return
	}

	if metricType != "counter" && metricType != "gauge" {
		http.Error(res, "Metric type must be 'counter' or 'gauge'", http.StatusBadRequest)
		return
	}

	if metricType == "counter" {
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(res, "Invalid counter value", http.StatusBadRequest)
			return
		}
		log.Printf("Counter %s updated by %d", metricName, value)
		s.storage.UpdateCounter(metricName, value)
	}

	if metricType == "gauge" {
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		log.Printf("Gauge %s set to %g", metricName, value)
		s.storage.UpdateGauge(metricName, value)
	}

	log.Printf("Type: %s, Name: %s, Value: %s", metricType, metricName, metricValue)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func (s *server) valueHandler(res http.ResponseWriter, req *http.Request) {
	metricType := chi.URLParam(req, "type")
	metricName := chi.URLParam(req, "name")

	if metricType == "gauge" {
		value, exists := s.storage.GetGauge(metricName)
		if !exists {
			http.Error(res, "Gauge not found", http.StatusNotFound)
			return
		}
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%g", value)
		return
	}

	if metricType == "counter" {
		value, exists := s.storage.GetCounter(metricName)
		if !exists {
			http.Error(res, "Counter not found", http.StatusNotFound)
			return
		}

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "%d", value)
		return
	}

	http.Error(res, "Unknown metric type", http.StatusBadRequest)
}

func (s *server) rootHandler(res http.ResponseWriter, req *http.Request) {
	gauges := s.storage.GetAllGauges()
	counters := s.storage.GetAllCounters()

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

	storage := repository.NewMemStorage()
	srv := &server{storage: storage}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/update/{type}/{name}/{value}", srv.updateHandler)
	r.Get("/value/{type}/{name}", srv.valueHandler)
	r.Get("/", srv.rootHandler)

	log.Printf("Starting server on: %s", cfg.Address)
	err := http.ListenAndServe(cfg.Address, r)
	if err != nil {
		log.Fatal(err)
	}

}
