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
		log.Printf("Gauge %s set to %f", metricName, value)
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
		fmt.Fprintf(res, "%f", value)
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

func main() {
	storage := repository.NewMemStorage()
	srv := &server{storage: storage}

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/update/{type}/{name}/{value}", srv.updateHandler)
	r.Get("/value/{type}/{name}", srv.valueHandler)

	log.Println("Starting server on: 8080")
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		log.Fatal(err)
	}

}
