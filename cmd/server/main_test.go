package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arigatory/sentinel/internal/repository"
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

			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			// Act
			srv.updateHandler(w, request)

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
