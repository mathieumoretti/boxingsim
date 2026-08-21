package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// getEnv returns environment variable or default value.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// parseIntEnv parses integer env var with fallback to defaultVal.
func parseIntEnv(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		var x int
		n, _ := fmt.Sscanf(v, "%d", &x)
		if n == 1 {
			return x
		}
	}
	return defaultVal
}

// buildDSN builds a PostgreSQL connection string.
func buildDSN(host string, port int, user, password, dbname string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

// buildBaseDSN builds a connection to the postgres template database.
func buildBaseDSN(host string, port int, user, password string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password,
	)
}

// TestDBConfig returns database configuration for testing.
func testDBConfig() *dbConfig {
	return &dbConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     parseIntEnv("TEST_DB_PORT", 5432),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", ""),
	}
}

// dbConfig holds database connection parameters.
type dbConfig struct {
	Host, User, Password string
	Port                 int
}

// sanitizeIdentifier makes a safe SQL identifier (starts with letter/underscore).
func sanitizeIdentifier(name string) string {
	result := "db_"
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			result += string(r)
		case r == '_':
			result += "_"
		default:
			continue // skip invalid chars
		}
	}
	return result + "_test"
}

// dropExistingDB drops a database if it exists (ignores errors).
func dropExistingDB(conn *sql.DB, dbName string) {
	query := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, sanitizeIdentifier(dbName))
	_, _ = conn.Exec(query) // ignore error - might not exist yet
}

// FreshDatabase creates a fresh isolated PostgreSQL database for testing.
// It: (1) drops existing DB if present; (2) creates new empty one with timestamp suffix;
// (3) runs MigrateUp from migrationsDir; returns connection and cleanup function.
func FreshDatabase(t *testing.T, dbPrefix string, migrationsDir string) (*sql.DB, func()) {
	cfg := testDBConfig()
	testDBName := fmt.Sprintf("%s_%d", strings.ToLower(dbPrefix), time.Now().UnixNano())

	// Connect to postgres template DB for creating new databases
	baseDSN := buildBaseDSN(cfg.Host, cfg.Port, cfg.User, cfg.Password)
	baseConn, err := sql.Open("postgres", baseDSN)
	if err != nil {
		t.Fatalf("FreshDatabase: failed to open postgres connection: %v", err)
	}

	dropExistingDB(baseConn, testDBName) // ignore errors if not exists yet

	// Create fresh database
	createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, sanitizeIdentifier(testDBName))
	if _, execErr := baseConn.Exec(createQuery); execErr != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: create database failed: %v", execErr)
	}

	testDSN := buildDSN(cfg.Host, cfg.Port, cfg.User, cfg.Password, testDBName)
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: open connection failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		baseConn.Close()
		t.Fatalf("FreshDatabase: cannot connect to %s: %v", testDBName, err)
	}

	err = MigrateUp(db, migrationsDir)
	if err != nil {
		db.Close()
		baseConn.Close()
		t.Fatalf("FreshDatabase: migrations failed from %s: %v", migrationsDir, err)
	}

	// Setup cleanup function to drop DB on test complete
	cleanup := func() {
		if db != nil {
			db.Close()
		}
		dropExistingDB(baseConn, testDBName)
		baseConn.Close()
	}

	t.Cleanup(cleanup)

	return db, nil
}

// TestDBConfig returns database configuration for testing (exported).
func TestDBConfig() *dbConfig {
	return testDBConfig()
}

// SetupTestDB creates a fresh isolated PostgreSQL database with migrations applied.
// Use this helper in tests that need an empty schema without full migration history.
// It registers cleanup via t.Cleanup and returns the opened connection ready for queries.
func SetupTestDB(t *testing.T) *sql.DB {
	db, _ := FreshDatabase(t, "test", "") // no migrations - just fresh DB with InitSchema needed manually
	return db
}

// CleanupTestDB drops a test database after cleanup is registered via t.Cleanup in FreshDatabase.
// This function exists for backwards compatibility but should not be called directly since
// FreshDatabase already handles teardown automatically through the testing.T interface itself.
func CleanupTestDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
