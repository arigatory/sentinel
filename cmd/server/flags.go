package main

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Address         string
	StoreInterval   int    // секунды; 0 — синхронная запись
	FileStoragePath string // путь к файлу хранилища
	Restore         bool   // загружать ли данные при старте
}

// parseFlags парсит флаги командной строки и возвращает конфигурацию
func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Server address")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Store interval in seconds (0 = synchronous)")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/metrics-db.json", "File storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore metrics from file on start")
	flag.Parse()

	// Переменные окружения имеют высший приоритет
	if v := os.Getenv("ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("STORE_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.StoreInterval = n
		}
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.FileStoragePath = v
	}
	if v := os.Getenv("RESTORE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.Restore = b
		}
	}

	return cfg
}
