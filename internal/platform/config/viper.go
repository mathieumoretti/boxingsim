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
	// Use absolute path to handle tests running from different directories
	configPath := getProjectConfigPath()
	v.AddConfigPath(configPath)
	v.SetConfigType("yaml")

	// Determine environment and config name
	env := getEnvironment()
	v.SetConfigName(env)

	// Try to read config file (optional - env vars override)
	// If no config file exists, this is fine - defaults + env vars will be used
	_ = v.ReadInConfig() // Ignore error - config file is optional

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

// getProjectConfigPath returns the absolute path to the config directory.
// This handles tests running from different directories by finding the project root.
func getProjectConfigPath() string {
	// Try to find the project root by looking for go.mod
	// Go changes working directory during test execution, so relative paths fail
	currentDir, err := os.Getwd()
	if err != nil {
		return "./config" // fallback
	}

	// Search upwards for go.mod (project root marker)
	searchDir := currentDir
	for {
		goModPath := strings.Join([]string{searchDir, "go.mod"}, string(rune(os.PathSeparator)))
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod - this is the project root
			return strings.Join([]string{searchDir, "config"}, string(rune(os.PathSeparator)))
		}

		// Move up one directory
		parentDir := strings.Split(searchDir, string(rune(os.PathSeparator)))
		if len(parentDir) <= 1 {
			// Reached filesystem root, use default
			return "./config"
		}
		searchDir = strings.Join(parentDir[:len(parentDir)-1], string(rune(os.PathSeparator)))
	}
}

// bindEnvVar explicitly binds an environment variable to a viper key.
// This is needed because AutomaticEnv() with nested keys doesn't work reliably.
func bindEnvVar(v *viper.Viper, key string, envVar string) {
	if val := os.Getenv(envVar); val != "" {
		v.Set(key, val)
	}
}

// bindIntEnvVar explicitly binds an integer environment variable to a viper key.
// Only sets the value if env var is present; defaults are applied in applyDefaults().
func bindIntEnvVar(v *viper.Viper, key string, envVar string, _ int) {
	if val := os.Getenv(envVar); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			v.Set(key, parsed)
		}
	}
	// If env var not set or invalid, don't set anything - let YAML file value persist
	// and applyDefaults() will fill in if still empty after unmarshal.
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
