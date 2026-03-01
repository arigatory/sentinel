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

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP server address")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Poll interval in seconds")
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

	return cfg
}
