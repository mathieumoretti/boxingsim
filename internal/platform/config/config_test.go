package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Loads configuration with default values", func(t *testing.T) {
		cfg := Load()

		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.DBHost)
		assert.Equal(t, 5432, cfg.DBPort)
		assert.Equal(t, "boxing", cfg.DBUser)
		assert.Equal(t, "boxing123", cfg.DBPassword)
		assert.Equal(t, "boxing", cfg.DBName)
		assert.Equal(t, "localhost:6379", cfg.RedisAddr)
		assert.Equal(t, "", cfg.RedisPassword)
		assert.NotEmpty(t, cfg.JWTSecret)
	})

	t.Run("Loads configuration with custom environment variables", func(t *testing.T) {
		// Set up environment variables
		os.Setenv("DB_HOST", "custom-db-host")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_USER", "customuser")
		os.Setenv("DB_PASSWORD", "custompass")
		os.Setenv("DB_NAME", "customdb")
		os.Setenv("REDIS_ADDR", "custom-redis:6380")
		os.Setenv("JWT_SECRET", "custom-jwt-secret")

		defer func() {
			// Clean up environment variables
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("REDIS_ADDR")
			os.Unsetenv("JWT_SECRET")
		}()

		cfg := Load()

		assert.NotNil(t, cfg)
		assert.Equal(t, "custom-db-host", cfg.DBHost)
		assert.Equal(t, 5433, cfg.DBPort)
		assert.Equal(t, "customuser", cfg.DBUser)
		assert.Equal(t, "custompass", cfg.DBPassword)
		assert.Equal(t, "customdb", cfg.DBName)
		assert.Equal(t, "custom-redis:6380", cfg.RedisAddr)
		assert.NotEmpty(t, cfg.JWTSecret)
	})

	t.Run("Handles missing environment variables gracefully", func(t *testing.T) {
		// Clear all relevant environment variables
		envVars := []string{
			"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
			"REDIS_ADDR", "REDIS_PASSWORD", "JWT_SECRET",
		}

		originalValues := make(map[string]string)
		for _, envVar := range envVars {
			originalValues[envVar] = os.Getenv(envVar)
			os.Unsetenv(envVar)
		}

		defer func() {
			// Restore original values
			for envVar, value := range originalValues {
				if value != "" {
					os.Setenv(envVar, value)
				} else {
					os.Unsetenv(envVar)
				}
			}
		}()

		cfg := Load()

		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.DBHost)
		assert.Equal(t, 5432, cfg.DBPort)
		assert.Equal(t, "boxing", cfg.DBUser)
		assert.Equal(t, "boxing123", cfg.DBPassword)
		assert.Equal(t, "boxing", cfg.DBName)
		assert.Equal(t, "localhost:6379", cfg.RedisAddr)
		assert.Equal(t, "", cfg.RedisPassword)
		assert.NotEmpty(t, cfg.JWTSecret)
	})
}

func TestGetEnv(t *testing.T) {
	t.Run("Returns default value when environment variable is not set", func(t *testing.T) {
		value := getEnv("NON_EXISTENT_VAR", "default_value")
		assert.Equal(t, "default_value", value)
	})

	t.Run("Returns environment variable value when set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		defer os.Unsetenv("TEST_VAR")

		value := getEnv("TEST_VAR", "default_value")
		assert.Equal(t, "test_value", value)
	})

	t.Run("Returns environment variable value when set with empty default", func(t *testing.T) {
		os.Setenv("EMPTY_TEST_VAR", "test_value")
		defer os.Unsetenv("EMPTY_TEST_VAR")

		value := getEnv("EMPTY_TEST_VAR", "")
		assert.Equal(t, "test_value", value)
	})
}

func TestParseIntEnv(t *testing.T) {
	t.Run("Returns default value when environment variable is not set", func(t *testing.T) {
		value := parseIntEnv("NON_EXISTENT_INT_VAR", 42)
		assert.Equal(t, 42, value)
	})

	t.Run("Returns parsed integer value when valid", func(t *testing.T) {
		os.Setenv("INT_TEST_VAR", "123")
		defer os.Unsetenv("INT_TEST_VAR")

		value := parseIntEnv("INT_TEST_VAR", 42)
		assert.Equal(t, 123, value)
	})

	t.Run("Returns default when parsing fails", func(t *testing.T) {
		os.Setenv("INVALID_INT_VAR", "not_a_number")
		defer os.Unsetenv("INVALID_INT_VAR")

		value := parseIntEnv("INVALID_INT_VAR", 42)
		assert.Equal(t, 42, value)
	})

	t.Run("Returns default when environment variable is empty", func(t *testing.T) {
		os.Setenv("EMPTY_INT_VAR", "")
		defer os.Unsetenv("EMPTY_INT_VAR")

		value := parseIntEnv("EMPTY_INT_VAR", 42)
		assert.Equal(t, 42, value)
	})
}

func TestConfigStructure(t *testing.T) {
	t.Run("Config struct has all expected fields", func(t *testing.T) {
		cfg := Load()

		assert.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.DBHost)
		assert.NotZero(t, cfg.DBPort)
		assert.NotEmpty(t, cfg.DBUser)
		assert.NotEmpty(t, cfg.DBPassword)
		assert.NotEmpty(t, cfg.DBName)
		assert.NotEmpty(t, cfg.RedisAddr)
		assert.NotEmpty(t, cfg.JWTSecret)
	})

	t.Run("Config values are reasonable", func(t *testing.T) {
		cfg := Load()

		assert.NotNil(t, cfg)
		assert.NotZero(t, cfg.DBPort)
		assert.True(t, cfg.DBPort > 0)
		assert.True(t, cfg.DBPort < 65536)
	})
}