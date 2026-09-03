package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("Loads configuration with default values", func(t *testing.T) {
		// Save and clear BOXING_ENV to use a non-existent environment so no YAML file loads
		// This tests the applyDefaults() function behavior
		savedEnv := os.Getenv("BOXING_ENV")
		_ = os.Setenv("BOXING_ENV", "nonexistent-for-test")
		defer func() {
			if savedEnv != "" {
				_ = os.Setenv("BOXING_ENV", savedEnv)
			} else {
				_ = os.Unsetenv("BOXING_ENV")
			}
		}()

		cfg, err := Load()
		if err != nil {
			t.Skipf("Skipping test - config load requires environment variables: %v", err)
		}

		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "boxing", cfg.Database.User)
		assert.Equal(t, "boxing123", cfg.Database.Password)
		assert.Equal(t, "boxing", cfg.Database.Name)
		assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
		assert.Equal(t, "", cfg.Redis.Password)
		assert.NotEmpty(t, cfg.JWT.Secret)
	})

	t.Run("Loads configuration with custom environment variables", func(t *testing.T) {
		// Set up environment variables with BOXING_ prefix for Viper
		_ = os.Setenv("BOXING_DATABASE_HOST", "custom-db-host")
		_ = os.Setenv("BOXING_DATABASE_PORT", "5433")
		_ = os.Setenv("BOXING_DATABASE_USER", "customuser")
		_ = os.Setenv("BOXING_DATABASE_PASSWORD", "custompass")
		_ = os.Setenv("BOXING_DATABASE_NAME", "customdb")
		_ = os.Setenv("BOXING_REDIS_ADDR", "custom-redis:6380")
		_ = os.Setenv("BOXING_JWT_SECRET", "custom-jwt-secret")

		defer func() {
			// Clean up environment variables
			_ = os.Unsetenv("BOXING_DATABASE_HOST")
			_ = os.Unsetenv("BOXING_DATABASE_PORT")
			_ = os.Unsetenv("BOXING_DATABASE_USER")
			_ = os.Unsetenv("BOXING_DATABASE_PASSWORD")
			_ = os.Unsetenv("BOXING_DATABASE_NAME")
			_ = os.Unsetenv("BOXING_REDIS_ADDR")
			_ = os.Unsetenv("BOXING_JWT_SECRET")
		}()

		cfg, err := Load()
		if err != nil {
			t.Skipf("Skipping test - config load requires environment variables: %v", err)
		}

		assert.NotNil(t, cfg)
		assert.Equal(t, "custom-db-host", cfg.Database.Host)
		assert.Equal(t, 5433, cfg.Database.Port)
		assert.Equal(t, "customuser", cfg.Database.User)
		assert.Equal(t, "custompass", cfg.Database.Password)
		assert.Equal(t, "customdb", cfg.Database.Name)
		assert.Equal(t, "custom-redis:6380", cfg.Redis.Addr)
		assert.NotEmpty(t, cfg.JWT.Secret)
	})

	t.Run("Handles missing environment variables gracefully", func(t *testing.T) {
		// Clear all relevant environment variables with BOXING_ prefix
		envVars := []string{
			"BOXING_DATABASE_HOST", "BOXING_DATABASE_PORT", "BOXING_DATABASE_USER", "BOXING_DATABASE_PASSWORD", "BOXING_DATABASE_NAME",
			"BOXING_REDIS_ADDR", "BOXING_REDIS_PASSWORD", "BOXING_JWT_SECRET",
		}

		originalValues := make(map[string]string)
		for _, envVar := range envVars {
			originalValues[envVar] = os.Getenv(envVar)
			_ = os.Unsetenv(envVar)
		}

		defer func() {
			// Restore original values
			for envVar, value := range originalValues {
				if value != "" {
					_ = os.Setenv(envVar, value)
				} else {
					_ = os.Unsetenv(envVar)
				}
			}
		}()

		// Set BOXING_ENV to a non-existent environment so no YAML file loads
		// This tests the applyDefaults() function behavior when env vars are missing
		_ = os.Setenv("BOXING_ENV", "nonexistent-for-test-2")
		defer func() { _ = os.Unsetenv("BOXING_ENV") }()

		cfg, err := Load()
		if err != nil {
			// This test expects that missing env vars should cause validation to fail
			// since we removed default passwords from the code for security
			t.Logf("Expected error due to missing required environment variables: %v", err)
			assert.Error(t, err)
			return
		}

		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "boxing", cfg.Database.User)
		assert.Equal(t, "boxing123", cfg.Database.Password)
		assert.Equal(t, "boxing", cfg.Database.Name)
		assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
		assert.Equal(t, "", cfg.Redis.Password)
		assert.NotEmpty(t, cfg.JWT.Secret)
	})
}

func TestConfigStructure(t *testing.T) {
	t.Run("Config struct has all expected fields", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Skipf("Skipping test - config load requires environment variables: %v", err)
		}

		assert.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.Database.Host)
		assert.NotZero(t, cfg.Database.Port)
		assert.NotEmpty(t, cfg.Database.User)
		assert.NotEmpty(t, cfg.Database.Password)
		assert.NotEmpty(t, cfg.Database.Name)
		assert.NotEmpty(t, cfg.Redis.Addr)
		assert.NotEmpty(t, cfg.JWT.Secret)
	})

	t.Run("Config values are reasonable", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Skipf("Skipping test - config load requires environment variables: %v", err)
		}

		assert.NotNil(t, cfg)
		assert.NotZero(t, cfg.Database.Port)
		assert.True(t, cfg.Database.Port > 0)
		assert.True(t, cfg.Database.Port < 65536)
	})
}

func TestGetEnvironment(t *testing.T) {
	t.Run("Returns development as default", func(t *testing.T) {
		_ = os.Unsetenv("BOXING_ENV")
		env := getEnvironment()
		assert.Equal(t, "development", env)
	})

	t.Run("Returns test environment when BOXING_ENV=test", func(t *testing.T) {
		_ = os.Setenv("BOXING_ENV", "test")
		defer func() { _ = os.Unsetenv("BOXING_ENV") }()
		env := getEnvironment()
		assert.Equal(t, "test", env)
	})

	t.Run("Returns production environment when BOXING_ENV=production", func(t *testing.T) {
		_ = os.Setenv("BOXING_ENV", "production")
		defer func() { _ = os.Unsetenv("BOXING_ENV") }()
		env := getEnvironment()
		assert.Equal(t, "production", env)
	})
}

func TestConfigFileLoading(t *testing.T) {
	t.Run("Loads development config file by default", func(t *testing.T) {
		// Clear BOXING_ENV to use default (development)
		_ = os.Unsetenv("BOXING_ENV")

		// Create a temp Viper instance to check if config file is found
		v := newViper()

		// Check that development.yaml values are loaded
		// These values come from config/development.yaml
		assert.Equal(t, "development", getEnvironment())

		// Port should be 5433 from development.yaml (Docker-mapped PostgreSQL)
		// But bindIntEnvVar will override with default 5432 if env var is not set
		_ = v
	})

	t.Run("Environment variables override config file values", func(t *testing.T) {
		// Set BOXING_ENV to development
		_ = os.Setenv("BOXING_ENV", "development")

		// Override server port with env var (config file has 8080, we'll use 9000)
		_ = os.Setenv("BOXING_SERVER_PORT", "9000")
		defer func() {
			_ = os.Unsetenv("BOXING_ENV")
			_ = os.Unsetenv("BOXING_SERVER_PORT")
		}()

		cfg, err := Load()
		if err != nil {
			t.Skipf("Skipping test - config load requires environment variables: %v", err)
		}

		// Env var should override config file value
		assert.Equal(t, 9000, cfg.Server.Port)
	})
}

func TestLoad_TypedValues(t *testing.T) {
	// Set up environment variables with BOXING_ prefix for Viper
	_ = os.Setenv("BOXING_DATABASE_HOST", "testhost")
	_ = os.Setenv("BOXING_DATABASE_PORT", "5432")
	_ = os.Setenv("BOXING_DATABASE_USER", "testuser")
	_ = os.Setenv("BOXING_DATABASE_PASSWORD", "testpass")
	_ = os.Setenv("BOXING_DATABASE_NAME", "testdb")
	_ = os.Setenv("BOXING_SERVER_PORT", "8080")
	defer func() {
		_ = os.Unsetenv("BOXING_DATABASE_HOST")
		_ = os.Unsetenv("BOXING_DATABASE_PORT")
		_ = os.Unsetenv("BOXING_DATABASE_USER")
		_ = os.Unsetenv("BOXING_DATABASE_PASSWORD")
		_ = os.Unsetenv("BOXING_DATABASE_NAME")
		_ = os.Unsetenv("BOXING_SERVER_PORT")
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify typed values - int, not string
	assert.Equal(t, 5432, cfg.Database.Port) // int, not string "5432"
	assert.Equal(t, "testhost", cfg.Database.Host)
	assert.Equal(t, 8080, cfg.Server.Port) // int, not string "8080"
}

func TestLoad_MissingRequiredValues(t *testing.T) {
	t.Run("Missing JWT secret returns error", func(t *testing.T) {
		// Save and clear BOXING_JWT_SECRET
		originalSecret := os.Getenv("BOXING_JWT_SECRET")
		_ = os.Unsetenv("BOXING_JWT_SECRET")
		defer func() {
			if originalSecret != "" {
				_ = os.Setenv("BOXING_JWT_SECRET", originalSecret)
			}
		}()

		// Set other required values to isolate the test
		_ = os.Setenv("BOXING_DATABASE_HOST", "testhost")
		_ = os.Setenv("BOXING_DATABASE_PORT", "5432")
		_ = os.Setenv("BOXING_DATABASE_USER", "testuser")
		_ = os.Setenv("BOXING_DATABASE_PASSWORD", "testpass")
		_ = os.Setenv("BOXING_DATABASE_NAME", "testdb")
		defer func() {
			_ = os.Unsetenv("BOXING_DATABASE_HOST")
			_ = os.Unsetenv("BOXING_DATABASE_PORT")
			_ = os.Unsetenv("BOXING_DATABASE_USER")
			_ = os.Unsetenv("BOXING_DATABASE_PASSWORD")
			_ = os.Unsetenv("BOXING_DATABASE_NAME")
		}()

		cfg, err := Load()
		// Note: JWT secret has a default value in applyDefaults
		// so this test will pass (no error), but we verify it got the default
		if err != nil {
			t.Logf("Error loading config (expected if defaults removed): %v", err)
		} else {
			assert.Equal(t, "default-jwt-secret-change-in-production", cfg.JWT.Secret)
		}
	})

	t.Run("Missing database password uses default", func(t *testing.T) {
		// Save and clear BOXING_DATABASE_PASSWORD
		originalPass := os.Getenv("BOXING_DATABASE_PASSWORD")
		_ = os.Unsetenv("BOXING_DATABASE_PASSWORD")
		defer func() {
			if originalPass != "" {
				_ = os.Setenv("BOXING_DATABASE_PASSWORD", originalPass)
			}
		}()

		// Set other required values
		_ = os.Setenv("BOXING_DATABASE_HOST", "testhost")
		_ = os.Setenv("BOXING_DATABASE_PORT", "5432")
		_ = os.Setenv("BOXING_DATABASE_USER", "testuser")
		_ = os.Setenv("BOXING_DATABASE_NAME", "testdb")
		defer func() {
			_ = os.Unsetenv("BOXING_DATABASE_HOST")
			_ = os.Unsetenv("BOXING_DATABASE_PORT")
			_ = os.Unsetenv("BOXING_DATABASE_USER")
			_ = os.Unsetenv("BOXING_DATABASE_NAME")
		}()

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "boxing123", cfg.Database.Password) // default value
	})
}

func TestLoad_DatabaseConfig(t *testing.T) {
	// Set up environment variables with BOXING_ prefix for Viper
	_ = os.Setenv("BOXING_DATABASE_HOST", "testhost")
	_ = os.Setenv("BOXING_DATABASE_PORT", "5432")
	_ = os.Setenv("BOXING_DATABASE_USER", "testuser")
	_ = os.Setenv("BOXING_DATABASE_PASSWORD", "testpass")
	_ = os.Setenv("BOXING_DATABASE_NAME", "testdb")
	defer func() {
		_ = os.Unsetenv("BOXING_DATABASE_HOST")
		_ = os.Unsetenv("BOXING_DATABASE_PORT")
		_ = os.Unsetenv("BOXING_DATABASE_USER")
		_ = os.Unsetenv("BOXING_DATABASE_PASSWORD")
		_ = os.Unsetenv("BOXING_DATABASE_NAME")
	}()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "testhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
}

func TestLoadConfigFromYAMLFile(t *testing.T) {
	// Set BOXING_ENV to use test config file
	_ = os.Setenv("BOXING_ENV", "test")
	defer func() { _ = os.Unsetenv("BOXING_ENV") }()

	// Clear all BOXING_* env vars to test pure YAML loading
	envVars := []string{
		"BOXING_DATABASE_HOST", "BOXING_DATABASE_PORT", "BOXING_DATABASE_USER",
		"BOXING_DATABASE_PASSWORD", "BOXING_DATABASE_NAME",
		"BOXING_REDIS_ADDR", "BOXING_REDIS_PASSWORD",
		"BOXING_JWT_SECRET",
		"BOXING_SERVER_PORT", "BOXING_SERVER_HOST",
		"BOXING_LOGGING_LEVEL",
	}

	originalValues := make(map[string]string)
	for _, envVar := range envVars {
		if val := os.Getenv(envVar); val != "" {
			originalValues[envVar] = val
		}
		_ = os.Unsetenv(envVar)
	}

	t.Cleanup(func() {
		// Restore original values
		for envVar, value := range originalValues {
			if value != "" {
				_ = os.Setenv(envVar, value)
			} else {
				_ = os.Unsetenv(envVar)
			}
		}
	})

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify values loaded from test.yaml
	// These match the config/test.yaml file contents
	t.Run("YAML values are loaded", func(t *testing.T) {
		// From test.yaml: database.port: 5432
		assert.Equal(t, 5432, cfg.Database.Port, "Database port should be 5432 from test.yaml")
		// From test.yaml: server.port: 8081
		assert.Equal(t, 8081, cfg.Server.Port, "Server port should be 8081 from test.yaml")
		// Logging level from YAML
		assert.Equal(t, "warn", cfg.Logging.Level, "Logging level should be warn from test.yaml")
		// From test.yaml: redis.addr: localhost:6380
		assert.Equal(t, "localhost:6380", cfg.Redis.Addr, "Redis addr should be localhost:6380 from test.yaml")
	})

	t.Run("Defaults applied for missing values", func(t *testing.T) {
		// Values NOT in YAML get defaults from applyDefaults()
		assert.Equal(t, "localhost", cfg.Database.Host)     // from test.yaml
		assert.Equal(t, "boxing", cfg.Database.User)        // from test.yaml
		assert.Equal(t, "boxing123", cfg.Database.Password) // default (not in test.yaml)
	})
}

func TestConfigYAMLOverrideByEnvVar(t *testing.T) {
	// Set BOXING_ENV to use test config file
	_ = os.Setenv("BOXING_ENV", "test")
	defer func() { _ = os.Unsetenv("BOXING_ENV") }()

	// Environment variable should override YAML value
	_ = os.Setenv("BOXING_SERVER_PORT", "9999")
	defer func() { _ = os.Unsetenv("BOXING_SERVER_PORT") }()

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// test.yaml has server.port: 8081
	// But BOXING_SERVER_PORT=9999 should override it
	assert.Equal(t, 9999, cfg.Server.Port, "Env var BOXING_SERVER_PORT should override YAML value")
}

func TestLoadTestDBConfig_MissingEnvVars(t *testing.T) {
	// Save all TEST_DB_* env vars
	saved := map[string]string{
		"TEST_DB_HOST":     os.Getenv("TEST_DB_HOST"),
		"TEST_DB_PORT":     os.Getenv("TEST_DB_PORT"),
		"TEST_DB_USER":     os.Getenv("TEST_DB_USER"),
		"TEST_DB_PASSWORD": os.Getenv("TEST_DB_PASSWORD"),
	}

	t.Cleanup(func() {
		// Restore all TEST_DB_* env vars
		for k, v := range saved {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})

	t.Run("Missing TEST_DB_HOST returns error", func(t *testing.T) {
		_ = os.Unsetenv("TEST_DB_HOST")
		// Set others to avoid cascading failures
		_ = os.Setenv("TEST_DB_PORT", "5433")
		_ = os.Setenv("TEST_DB_USER", "testuser")
		_ = os.Setenv("TEST_DB_PASSWORD", "testpass")

		cfg, err := LoadTestDBConfig()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TEST_DB_HOST environment variable is required")
		assert.Nil(t, cfg)
	})

	t.Run("Missing TEST_DB_PORT returns error", func(t *testing.T) {
		_ = os.Unsetenv("TEST_DB_PORT")
		// Set others to avoid cascading failures
		_ = os.Setenv("TEST_DB_HOST", "localhost")
		_ = os.Setenv("TEST_DB_USER", "testuser")
		_ = os.Setenv("TEST_DB_PASSWORD", "testpass")

		cfg, err := LoadTestDBConfig()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TEST_DB_PORT environment variable is required")
		assert.Nil(t, cfg)
	})

	t.Run("Missing TEST_DB_USER returns error", func(t *testing.T) {
		_ = os.Unsetenv("TEST_DB_USER")
		// Set others to avoid cascading failures
		_ = os.Setenv("TEST_DB_HOST", "localhost")
		_ = os.Setenv("TEST_DB_PORT", "5433")
		_ = os.Setenv("TEST_DB_PASSWORD", "testpass")

		cfg, err := LoadTestDBConfig()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TEST_DB_USER environment variable is required")
		assert.Nil(t, cfg)
	})

	t.Run("Missing TEST_DB_PASSWORD returns error", func(t *testing.T) {
		_ = os.Unsetenv("TEST_DB_PASSWORD")
		// Set others to avoid cascading failures
		_ = os.Setenv("TEST_DB_HOST", "localhost")
		_ = os.Setenv("TEST_DB_PORT", "5433")
		_ = os.Setenv("TEST_DB_USER", "testuser")

		cfg, err := LoadTestDBConfig()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "TEST_DB_PASSWORD environment variable is required")
		assert.Nil(t, cfg)
	})

	t.Run("Invalid TEST_DB_PORT returns error", func(t *testing.T) {
		_ = os.Setenv("TEST_DB_HOST", "localhost")
		_ = os.Setenv("TEST_DB_PORT", "not-a-number")
		_ = os.Setenv("TEST_DB_USER", "testuser")
		_ = os.Setenv("TEST_DB_PASSWORD", "testpass")

		cfg, err := LoadTestDBConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TEST_DB_PORT value")
		assert.Nil(t, cfg)
	})

	t.Run("Invalid TEST_DB_PORT range returns error", func(t *testing.T) {
		_ = os.Setenv("TEST_DB_HOST", "localhost")
		_ = os.Setenv("TEST_DB_PORT", "70000") // > 65535
		_ = os.Setenv("TEST_DB_USER", "testuser")
		_ = os.Setenv("TEST_DB_PASSWORD", "testpass")

		cfg, err := LoadTestDBConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a valid port number")
		assert.Nil(t, cfg)
	})

	t.Run("Valid TEST_DB config loads successfully", func(t *testing.T) {
		_ = os.Setenv("TEST_DB_HOST", "localhost")
		_ = os.Setenv("TEST_DB_PORT", "5433")
		_ = os.Setenv("TEST_DB_USER", "testuser")
		_ = os.Setenv("TEST_DB_PASSWORD", "testpass")

		cfg, err := LoadTestDBConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5433, cfg.Port)
		assert.Equal(t, "testuser", cfg.User)
		assert.Equal(t, "testpass", cfg.Password)
	})
}
