package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arigatory/sentinel/internal/repository"
	"github.com/go-chi/chi/v5"
)

func TestUpdateHandler(t *testing.T) {
	type want struct {
		statusCode   int
		checkStorage bool
		metricType   string
		metricName   string
		gaugeValue   float64
		counterValue int64
	}

	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "valid gauge",
			method: http.MethodPost,
			url:    "/update/gauge/temperature/36.6",
			want: want{
				statusCode:   http.StatusOK,
				checkStorage: true,
				metricType:   "gauge",
				metricName:   "temperature",
				gaugeValue:   36.6,
			},
		},
		{
			name:   "valid counter",
			method: http.MethodPost,
			url:    "/update/counter/requests/5",
			want: want{
				statusCode:   http.StatusOK,
				checkStorage: true,
				metricType:   "counter",
				metricName:   "requests",
				counterValue: 5,
			},
		},
		{
			name:   "invalid method GET",
			method: http.MethodGet,
			url:    "/update/gauge/temperature/36.6",
			want:   want{statusCode: http.StatusMethodNotAllowed},
		},
		{
			name:   "invalid type",
			method: http.MethodPost,
			url:    "/update/unknown/metric/123",
			want:   want{statusCode: http.StatusBadRequest},
		},
		{
			name:   "invalid gauge value",
			method: http.MethodPost,
			url:    "/update/gauge/temperature/invalid",
			want:   want{statusCode: http.StatusBadRequest},
		},
		{
			name:   "invalid counter value",
			method: http.MethodPost,
			url:    "/update/counter/requests/3.14",
			want:   want{statusCode: http.StatusBadRequest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			storage := repository.NewMemStorage()
			srv := &server{storage: storage}

			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", srv.updateHandler)

			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			// Assert
			result := w.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.want.statusCode {
				t.Errorf("Expected status %d, got %d",
					tt.want.statusCode, result.StatusCode)
			}

			if tt.want.checkStorage {
				switch tt.want.metricType {
				case "gauge":
					value, exists := storage.GetGauge(tt.want.metricName)
					if !exists {
						t.Errorf("Gauge '%s' should exist", tt.want.metricName)
					}
					if value != tt.want.gaugeValue {
						t.Errorf("Expected gauge value %f, got %f",
							tt.want.gaugeValue, value)
					}
				case "counter":
					value, exists := storage.GetCounter(tt.want.metricName)
					if !exists {
						t.Errorf("Counter '%s' should exist", tt.want.metricName)
					}
					if value != tt.want.counterValue {
						t.Errorf("Expected counter value %d, got %d",
							tt.want.counterValue, value)
					}
				}
			}
		})
	}
}

func TestValueHandler(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}

	tests := []struct {
		name      string
		setupData func(*repository.MemStorage) // Функция для подготовки данных
		url       string
		want      want
	}{
		{
			name: "get existing gauge",
			setupData: func(s *repository.MemStorage) {
				s.UpdateGauge("temperature", 36.6)
			},
			url: "/value/gauge/temperature",
			want: want{
				statusCode: http.StatusOK,
				body:       "36.600000",
			},
		},
		{
			name: "get existing counter",
			setupData: func(s *repository.MemStorage) {
				s.UpdateCounter("requests", 42)
			},
			url: "/value/counter/requests",
			want: want{
				statusCode: http.StatusOK,
				body:       "42",
			},
		},
		{
			name:      "get non-existent gauge",
			setupData: func(s *repository.MemStorage) {},
			url:       "/value/gauge/unknown",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "get non-existent counter",
			setupData: func(s *repository.MemStorage) {},
			url:       "/value/counter/unknown",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:      "get unknown metric type",
			setupData: func(s *repository.MemStorage) {},
			url:       "/value/unknown/metric",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			storage := repository.NewMemStorage()
			tt.setupData(storage) // Подготовка данных
			srv := &server{storage: storage}

			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", srv.valueHandler)

			request := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			// Act
			r.ServeHTTP(w, request)

			// Assert
			result := w.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.want.statusCode {
				t.Errorf("Expected status %d, got %d",
					tt.want.statusCode, result.StatusCode)
			}

			if tt.want.body != "" {
				bodyBytes, err := io.ReadAll(result.Body)
				if err != nil {
					t.Fatal(err)
				}
				bodyString := string(bodyBytes)
				if bodyString != tt.want.body {
					t.Errorf("Expected body '%s', got '%s'",
						tt.want.body, bodyString)
				}
			}
		})
	}
}


func TestRootHandler(t *testing.T) {
    // Arrange
    storage := repository.NewMemStorage()
    
    storage.UpdateGauge("temperature", 36.6)
    storage.UpdateGauge("memory", 1024.5)
    storage.UpdateCounter("requests", 42)
    storage.UpdateCounter("errors", 5)
    
    srv := &server{storage: storage}
    
    r := chi.NewRouter()
    r.Get("/", srv.rootHandler)
    
    request := httptest.NewRequest(http.MethodGet, "/", nil)
    w := httptest.NewRecorder()
    
    // Act
    r.ServeHTTP(w, request)
    
    // Assert
    result := w.Result()
    defer result.Body.Close()
    
    if result.StatusCode != http.StatusOK {
        t.Errorf("Expected status 200, got %d", result.StatusCode)
    }
    
    contentType := result.Header.Get("Content-Type")
    if contentType != "text/html; charset=utf-8" {
        t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
    }
    
    body, err := io.ReadAll(result.Body)
    if err != nil {
        t.Fatal(err)
    }
    bodyString := string(body)
    
    expectedStrings := []string{
        "temperature",
        "36.6",
        "memory",
        "1024.5",
        "requests",
        "42",
        "errors",
        "5",
        "<h2>Gauges</h2>",
        "<h2>Counters</h2>",
        "<h1>All Metrics</h1>",
    }
    
    for _, expected := range expectedStrings {
        if !strings.Contains(bodyString, expected) {
            t.Errorf("Expected body to contain '%s', but it was not found", expected)
        }
    }
}