package main

import (
	"flag"
	"os"
)

type Config struct {
	Address string
}

// parseFlags парсит флаги командной строки и возвращает конфигурацию
func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Server address")
	flag.Parse()

	// Переменная окружения имеет высший приоритет
	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}

	return cfg
}
