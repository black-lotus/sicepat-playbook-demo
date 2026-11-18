// Package config membaca konfigurasi aplikasi dari environment variable.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config adalah konfigurasi bersama untuk binary api, worker, dan consumer.
type Config struct {
	HTTPAddr          string
	DatabaseURL       string
	RedisAddr         string
	KafkaBrokers      []string
	KafkaTopic        string
	KafkaGroupID      string
	CacheTTL          time.Duration
	AggregateInterval time.Duration
	LogLevel          string
}

// Load membaca konfigurasi dari environment. DATABASE_URL wajib diisi.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "sigesit.delivery-events"),
		KafkaGroupID: getEnv("KAFKA_GROUP_ID", "sigesit-performance-consumer"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL wajib diisi")
	}

	var err error
	if cfg.CacheTTL, err = getDuration("CACHE_TTL", 5*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.AggregateInterval, err = getDuration("AGGREGATE_INTERVAL", 5*time.Minute); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return d, nil
}
