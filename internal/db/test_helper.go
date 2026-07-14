package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"

	"github.com/mormm/boxing/internal/platform/config"
)

// TestDBConfig returns database configuration for testing
func TestDBConfig() *config.Config {
	return &config.Config{
		DBHost:     getEnv("TEST_DB_HOST", "localhost"),
		DBPort:     parseIntEnv("TEST_DB_PORT", 5432),
		DBUser:     getEnv("TEST_DB_USER", "test_user"),
		DBPassword: getEnv("TEST_DB_PASSWORD", "test_password"),
		DBName:     getEnv("TEST_DB_NAME", "test_boxing"),
	}
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseIntEnv returns parsed integer environment variable or default value
func parseIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intVal int
		if _, err := fmt.Sscanf(value, "%d", &intVal); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// SetupTestDB creates a test database connection and initializes schema
func SetupTestDB(t *testing.T) *sql.DB {
	cfg := TestDBConfig()

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal(err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		// If we can't connect to the test database, skip the test
		t.Skipf("Cannot connect to test database: %v", err)
	}

	// Initialize schema for testing
	if err := InitializeSchema(db); err != nil {
		t.Fatal(err)
	}

	return db
}

// CleanupTestDB closes the database connection
func CleanupTestDB(db *sql.DB) {
	if db != nil {
		_ = db.Close()
	}
}
