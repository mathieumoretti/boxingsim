package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Config represents the complete application configuration.
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Server   ServerConfig   `mapstructure:"server"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// TestDBConfig represents test database configuration for integration tests.
type TestDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// Load initializes and returns application configuration from Viper.
func Load() (*Config, error) {
	v := newViper()

	// Read environment variables explicitly for reliable nested struct support
	readEnvConfig(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults for non-secret fields (only if not set by env vars or config file)
	applyDefaults(cfg)

	// Validate required fields
	if err := validateRequiredFields(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "boxing"
	}
	if cfg.Database.Password == "" {
		cfg.Database.Password = "boxing123"
	}
	if cfg.Database.Name == "" {
		cfg.Database.Name = "boxing"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}
	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "localhost:6379"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "default-jwt-secret-change-in-production"
	}
}

func validateRequiredFields(cfg *Config) error {
	// Validate port ranges
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("DATABASE_PORT must be a valid port number (1-65535), got: %d", cfg.Database.Port)
	}

	if cfg.Server.Port != 0 && (cfg.Server.Port <= 0 || cfg.Server.Port > 65535) {
		return fmt.Errorf("SERVER_PORT must be a valid port number (1-65535), got: %d", cfg.Server.Port)
	}

	return nil
}

// LoadTestDBConfig loads test database configuration.
// Priority: TEST_DB_* env vars → main DATABASE config from file/env → defaults.
// Using TEST_DB_* allows isolating tests from the main development database.
func LoadTestDBConfig() (*TestDBConfig, error) {
	cfg := &TestDBConfig{}

	// Check if explicit test DB is configured via env vars
	testHost := os.Getenv("TEST_DB_HOST")
	testPort := os.Getenv("TEST_DB_PORT")
	testUser := os.Getenv("TEST_DB_USER")
	testPassword := os.Getenv("TEST_DB_PASSWORD")

	// If any TEST_DB_* var is set, all must be set (partial config is error-prone)
	if testHost != "" || testPort != "" || testUser != "" || testPassword != "" {
		if testHost == "" {
			return nil, errors.New("TEST_DB_HOST environment variable is required when using TEST_DB_* configuration")
		}
		cfg.Host = testHost

		if testPort == "" {
			return nil, errors.New("TEST_DB_PORT environment variable is required when using TEST_DB_* configuration")
		}
		port, err := strconv.Atoi(testPort)
		if err != nil {
			return nil, fmt.Errorf("invalid TEST_DB_PORT value: %w", err)
		}
		cfg.Port = port

		if testUser == "" {
			return nil, errors.New("TEST_DB_USER environment variable is required when using TEST_DB_* configuration")
		}
		cfg.User = testUser

		if testPassword == "" {
			return nil, errors.New("TEST_DB_PASSWORD environment variable is required when using TEST_DB_* configuration")
		}
		cfg.Password = testPassword

		return cfg, nil
	}

	// Fall back to main DATABASE config (from config file or BOXING_DATABASE_* env vars)
	mainCfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load main config for tests: %w", err)
	}

	cfg.Host = mainCfg.Database.Host
	cfg.Port = mainCfg.Database.Port
	cfg.User = mainCfg.Database.User
	cfg.Password = mainCfg.Database.Password

	return cfg, nil
}

// func isDockerEnv() bool {
//	_, err := os.Stat("/.dockerenv")
//	if err == nil {
//		return true
//	}

//	dockerHost := os.Getenv("DOCKER_HOST")
//	if dockerHost != "" {
//		return true
//	}

//	return false
//}
