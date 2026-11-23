package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppPort, AppEnv                                       string
	DBHost, DBPort, DBUser, DBPassword, DBName, DBSSLMode string
}

func Load() (Config, error) {
	cfg := Config{
		AppPort: get("APP_PORT", "8080"), AppEnv: get("APP_ENV", "development"),
		DBHost: os.Getenv("DB_HOST"), DBPort: get("DB_PORT", "5432"),
		DBUser: os.Getenv("DB_USER"), DBPassword: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"), DBSSLMode: get("DB_SSLMODE", "disable"),
	}
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return Config{}, fmt.Errorf("DB_HOST, DB_USER, and DB_NAME must be set")
	}
	return cfg, nil
}

func get(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
