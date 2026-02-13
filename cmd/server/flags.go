package main

import (
	"flag"
)

type Config struct {
	Address string
}

// parseFlags парсит флаги командной строки и возвращает конфигурацию
func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Server address")
	flag.Parse()

	return cfg
}
