package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Subj struct {
	Product string `json:"name"`
	Price   int    `json:"price"`
}

func JSONHandler(w http.ResponseWriter, req *http.Request) {
	// собираем данные
	subj := Subj{"Milk", 50}
	// кодируем в JSON
	resp, err := json.Marshal(subj)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// устанавливаем заголовок Content-Type
	// для передачи клиенту информации, кодированной в JSON
	w.Header().Set("content-type", "application/json")
	// устанавливаем код 200
	w.WriteHeader(http.StatusOK)
	// пишем тело ответа
	w.Write(resp)
}

func mainPage(res http.ResponseWriter, req *http.Request) {
	body := fmt.Sprintf("Method: %s\r\n", req.Method)
	body += "Header ===============\r\n"
	for k, v := range req.Header {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	body += "Query parameters ===============\r\n"
	for k, v := range req.URL.Query() {
		body += fmt.Sprintf("%s: %v\r\n", k, v)
	}
	res.Write([]byte(body))
}

func apiPage(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Это страница /api."))
}

func updateHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.Trim(req.URL.Path, "/")
	parts := strings.Split(path, "/")

	log.Printf("Received request: %s, parts: %v", path, parts)

	if len(parts) != 4 {
		http.Error(res, "Bad request. Need 4 parts: update/{type}/{name}/{value}", http.StatusNotFound)
		return
	}

	metricType := strings.TrimSpace(parts[1])
	metricName := strings.TrimSpace(parts[2])
	metricValue := strings.TrimSpace(parts[3])

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
	}

	if metricType == "gauge" {
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(res, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		log.Printf("Gauge %s set to %s", metricName, metricValue)
	}

	log.Printf("Type: %s, Name: %s, Value: %s", metricType, metricName, metricValue)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/api/`, apiPage)
	mux.HandleFunc(`/json/`, JSONHandler)
	mux.HandleFunc(`/`, mainPage)
	mux.HandleFunc("/update/", updateHandler)

	log.Println("Starting server on: 8080")
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}

}
