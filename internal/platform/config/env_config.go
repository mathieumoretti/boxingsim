package config

import (
	"os"
)

// EnvConfig holds environment-specific configuration
type EnvConfig struct {
	APIBasePath string
}

// LoadEnvConfig loads environment configuration
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