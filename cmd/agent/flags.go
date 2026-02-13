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
}

func parseFlags() *Config {
	cfg := &Config{}

	defaultAddress := getEnv("ADDRESS", "localhost:8080")
	defaultReportInterval := getEnvAsInt("REPORT_INTERVAL", 10)
	defaultPollInterval := getEnvAsInt("POLL_INTERVAL", 2)

	flag.StringVar(&cfg.Address, "a", defaultAddress, "HTTP server address")
	flag.IntVar(&cfg.ReportInterval, "r", defaultReportInterval, "Report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", defaultPollInterval, "Poll interval in seconds")
	flag.Parse()

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
