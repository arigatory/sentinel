package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Address        string
	ReportInterval int
	PollInterval   int
	Key            string // ключ подписи; пустой — подпись выключена
	RateLimit      int    // максимум одновременно исходящих запросов
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "Key for request signing (HMAC-SHA256)")
	flag.IntVar(&cfg.RateLimit, "l", 1, "Maximum number of concurrent outgoing requests")
	flag.Parse()

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if value, err := strconv.Atoi(envReportInterval); err == nil {
			cfg.ReportInterval = value
		}
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if value, err := strconv.Atoi(envPollInterval); err == nil {
			cfg.PollInterval = value
		}
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		if value, err := strconv.Atoi(envRateLimit); err == nil {
			cfg.RateLimit = value
		}
	}

	// воркер-пул не имеет смысла при нулевом или отрицательном размере
	if cfg.RateLimit < 1 {
		cfg.RateLimit = 1
	}

	return cfg
}
