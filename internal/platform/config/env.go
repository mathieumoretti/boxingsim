package config

import (
	"fmt"
	"os"
)

// Environment configuration for API path handling
type EnvConfig struct {
	APIBasePath string
}

func LoadEnvConfig() *EnvConfig {
	return &EnvConfig{
		APIBasePath: getEnv("API_BASE_PATH", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}