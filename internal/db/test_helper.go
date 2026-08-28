package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/mormm/boxing/internal/platform/config"
)

// buildDSN builds a PostgreSQL connection string.
func buildDSN(cfg *config.TestDBConfig, dbname string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, dbname,
	)
}

// buildBaseDSN builds a connection to the postgres template database.
func buildBaseDSN(cfg *config.TestDBConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password,
	)
}

// loadTestDBConfig loads test database configuration using centralized config loader.
// All fields are required - no defaults to prevent silent connection to wrong PostgreSQL instance.
func loadTestDBConfig() (*config.TestDBConfig, error) {
	return config.LoadTestDBConfig()
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

// verifyConnectedDatabase checks that the connected database matches the expected name.
// This is a CRITICAL safety check to prevent tests from running against the wrong database.
func verifyConnectedDatabase(db *sql.DB, expectedDB, host string, port int, user string) error {
	var actualDB string
	if err := db.QueryRow("SELECT current_database()").Scan(&actualDB); err != nil {
		return fmt.Errorf("failed to verify connected database: %w", err)
	}

	if actualDB != expectedDB {
		return fmt.Errorf(
			"connected to WRONG DATABASE (safety check failed):\n"+
				"  expected database: %s\n"+
				"  actual database:   %s\n"+
				"  PostgreSQL server: %s:%d\n"+
				"  user:              %s\n"+
				"\nThis indicates TEST_DB_HOST or TEST_DB_PORT may be pointing to the wrong PostgreSQL instance.",
			expectedDB, actualDB, host, port, user,
		)
	}

	return nil
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
	t.Helper()
	cfg, err := loadTestDBConfig()
	if err != nil {
		t.Fatalf("FreshDatabase: failed to load test DB config: %v", err)
	}

	testDBName := fmt.Sprintf("%s_%d", strings.ToLower(dbPrefix), time.Now().UnixNano())

	// Connect to postgres template DB for creating new databases
	baseDSN := buildBaseDSN(cfg)
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

	testDSN := buildDSN(cfg, testDBName)
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

	// 🔴 CRITICAL SAFETY CHECK: Verify we connected to the correct database
	if err := verifyConnectedDatabase(db, testDBName, cfg.Host, cfg.Port, cfg.User); err != nil {
		db.Close()
		baseConn.Close()
		t.Fatalf("FreshDatabase: %v", err)
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

// TestDBConfig loads and returns database configuration for testing (exported).
// DEPRECATED: Use config.LoadTestDBConfig() directly instead.
func TestDBConfig() (*config.TestDBConfig, error) {
	return loadTestDBConfig()
}

// SetupTestDB creates a fresh isolated PostgreSQL database with the initial schema.
// It runs migrations from "./db/migrations" and registers cleanup via t.Cleanup.
// Returns the opened connection ready for queries or skips the test if DB unavailable.
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, _ := FreshDatabaseWithMigrations(t, "test", true)
	return db
}

// CleanupTestDB drops a test database after cleanup is registered via t.Cleanup in FreshDatabase*.
// This function exists for backwards compatibility but should not be called directly since
// FreshDatabase* already handles teardown automatically through the testing.T interface itself.
func CleanupTestDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

// FreshDatabaseWithMigrations is a convenience wrapper that creates a fresh database and applies migrations,
// or skips without migration if skipMigrations=true (for tests using InitializeSchema directly).
func FreshDatabaseWithMigrations(t *testing.T, prefix string, runMigrations bool) (*sql.DB, func()) {
	t.Helper()
	if !runMigrations {
		return FreshDatabaseWithoutMigrations(t, prefix)
	}
	migrationsDir := os.Getenv("TEST_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "db/migrations"
	}
	db, cleanupFn := FreshDatabase(t, prefix, migrationsDir)
	return db, cleanupFn
}

// FreshDatabaseWithoutMigrations creates a fresh isolated database without running any migrations.
// Use this for tests that manually call InitializeSchema or test migration behavior itself.
func FreshDatabaseWithoutMigrations(t *testing.T, prefix string) (*sql.DB, func()) {
	t.Helper()
	cfg, err := loadTestDBConfig()
	if err != nil {
		t.Fatalf("FreshDatabaseWithoutMigrations: failed to load test DB config: %v", err)
	}

	testDBName := fmt.Sprintf("%s_%d", strings.ToLower(prefix), time.Now().UnixNano())

	baseDSN := buildBaseDSN(cfg)
	baseConn, err := sql.Open("postgres", baseDSN)
	if err != nil {
		t.Fatalf("FreshDatabaseWithoutMigrations: failed to open postgres connection: %v", err)
	}

	dropExistingDB(baseConn, testDBName)

	createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, sanitizeIdentifier(testDBName))
	if _, execErr := baseConn.Exec(createQuery); execErr != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabaseWithoutMigrations: create database failed: %v", execErr)
	}

	testDSN := buildDSN(cfg, testDBName)
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabaseWithoutMigrations: open connection failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		baseConn.Close()
		t.Fatalf("FreshDatabaseWithoutMigrations: cannot connect to %s: %v", testDBName, err)
	}

	// 🔴 CRITICAL SAFETY CHECK: Verify we connected to the correct database
	if err := verifyConnectedDatabase(db, testDBName, cfg.Host, cfg.Port, cfg.User); err != nil {
		db.Close()
		baseConn.Close()
		t.Fatalf("FreshDatabaseWithoutMigrations: %v", err)
	}

	cleanupFn := func() {
		if db != nil {
			db.Close()
		}
		dropExistingDB(baseConn, testDBName)
		baseConn.Close()
	}

	t.Cleanup(cleanupFn)

	return db, cleanupFn
}
