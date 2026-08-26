package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const envPrefix = "BOXING"

// newViper creates and configures a new Viper instance.
func newViper() *viper.Viper {
	v := viper.New()

	// Set environment variable prefix
	v.SetEnvPrefix(envPrefix)

	// Enable automatic environment variable binding
	// BOXING_DATABASE_HOST -> database.host (after replacer transforms it)
	v.AutomaticEnv()

	// Replace underscores with dots in env var names for nested struct support
	// BOXING_DATABASE_HOST becomes database.host instead of database_host
	v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))

	// Add config paths (search for config files)
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	v.SetConfigType("yaml")

	// Determine environment and config name
	env := getEnvironment()
	v.SetConfigName(env)

	// Try to read config file (optional - env vars override)
	// If no config file exists, this is fine - defaults + env vars will be used
	_ = v.ReadInConfig()

	return v
}

// getEnvironment returns the current environment based on BOXING_ENV or DEFAULTS to "development".
func getEnvironment() string {
	env := os.Getenv(envPrefix + "_ENV")
	if env == "" {
		return "development"
	}
	return env
}

// bindEnvVar explicitly binds an environment variable to a viper key.
// This is needed because AutomaticEnv() with nested keys doesn't work reliably.
func bindEnvVar(v *viper.Viper, key string, envVar string) {
	if val := os.Getenv(envVar); val != "" {
		v.Set(key, val)
	}
}

// bindIntEnvVar explicitly binds an integer environment variable to a viper key.
func bindIntEnvVar(v *viper.Viper, key string, envVar string, defaultVal int) {
	if val := os.Getenv(envVar); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			v.Set(key, parsed)
			return
		}
	}
	v.Set(key, defaultVal)
}

// readEnvConfig reads configuration from environment variables with BOXING_ prefix.
// This is a fallback when AutomaticEnv doesn't work properly with nested structs.
func readEnvConfig(v *viper.Viper) {
	// Database configuration
	bindEnvVar(v, "database.host", envPrefix+"_DATABASE_HOST")
	bindIntEnvVar(v, "database.port", envPrefix+"_DATABASE_PORT", 5432)
	bindEnvVar(v, "database.user", envPrefix+"_DATABASE_USER")
	bindEnvVar(v, "database.password", envPrefix+"_DATABASE_PASSWORD")
	bindEnvVar(v, "database.name", envPrefix+"_DATABASE_NAME")

	// Redis configuration
	bindEnvVar(v, "redis.addr", envPrefix+"_REDIS_ADDR")
	bindEnvVar(v, "redis.password", envPrefix+"_REDIS_PASSWORD")

	// JWT configuration
	bindEnvVar(v, "jwt.secret", envPrefix+"_JWT_SECRET")

	// Server configuration
	bindIntEnvVar(v, "server.port", envPrefix+"_SERVER_PORT", 8080)
	bindEnvVar(v, "server.host", envPrefix+"_SERVER_HOST")

	// Logging configuration
	bindEnvVar(v, "logging.level", envPrefix+"_LOGGING_LEVEL")
}
