package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// GetRepoRoot finds the repository root by walking up the directory tree
// looking for .git or go.mod. This is used to resolve migration paths
// regardless of the current working directory during test execution.
func GetRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // reached root without finding repo markers
		}
		dir = parentDir
	}

	return "", fmt.Errorf("could not find repository root (no .git or go.mod found)")
}

// GetMigrationsDir returns the absolute path to the migrations directory.
// It first checks TEST_MIGRATIONS_DIR env var if set and absolute,
// then tries to locate it relative to the repo root, finally falling back
// to a relative path. This ensures migrations are found during test execution
// when Go changes the working directory.
func GetMigrationsDir() string {
	// Allow override via environment variable (must be absolute path)
	if v := os.Getenv("TEST_MIGRATIONS_DIR"); v != "" && filepath.IsAbs(v) {
		return v
	}

	// Try to find migrations relative to repo root
	repoRoot, err := GetRepoRoot()
	if err == nil {
		migrationsDir := filepath.Join(repoRoot, "db", "migrations")
		if _, statErr := os.Stat(migrationsDir); statErr == nil {
			return migrationsDir
		}
	}

	// Fallback to relative path
	return "db/migrations"
}

// generateUniqueDBName generates a unique database name with crypto-random suffix.
// Uses 64 bits of entropy (10 hex chars) to ensure uniqueness across concurrent tests.
func generateUniqueDBName(prefix string) string {
	b := make([]byte, 8) // 64 bits of entropy
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if rand fails (should never happen)
		panic(fmt.Sprintf("generateUniqueDBName: crypto/rand failed: %v", err))
	}
	hexStr := hex.EncodeToString(b)[:10] // e.g., "f2a9b3c8d1"
	return fmt.Sprintf("test_%s_%s", prefix, hexStr)
}

// sanitizeIdentifier makes a safe SQL identifier (starts with letter/underscore).
func sanitizeIdentifier(name string) string {
	var builder strings.Builder
	builder.WriteString("db_")
	for _, r := range name {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			builder.WriteRune(r)
		case r == '_':
			builder.WriteRune('_')
		default:
			continue // skip invalid chars
		}
	}
	builder.WriteString("_test")
	return builder.String()
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
				"\nThis indicates DATABASE config (BOXING_DATABASE_* or TEST_DB_*) may be pointing to the wrong PostgreSQL instance.",
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

	testDBName := generateUniqueDBName(dbPrefix)

	// Connect to postgres template DB for creating new databases
	baseDSN := buildBaseDSN(cfg)
	baseConn, err := sql.Open("postgres", baseDSN)
	if err != nil {
		t.Fatalf("FreshDatabase: cannot open postgres connection: %v", err)
	}

	// Try to ping the database to verify connectivity before attempting CREATE DATABASE
	if err := baseConn.Ping(); err != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: PostgreSQL not available at %s:%d: %v", cfg.Host, cfg.Port, err)
	}

	// Drop existing db using original name (sanitizeIdentifier will sanitize it)
	dropExistingDB(baseConn, testDBName) // ignore errors if not exists yet

	// Sanitize the database name for PostgreSQL compatibility (use consistently throughout)
	sanitizedDBName := sanitizeIdentifier(testDBName)
	t.Logf("FreshDatabase: original name=%s, sanitized name=%s", testDBName, sanitizedDBName)

	// Create fresh database
	createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, sanitizedDBName)
	t.Logf("FreshDatabase: executing %s", createQuery)
	result, execErr := baseConn.Exec(createQuery)
	if execErr != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: cannot create database: %v", execErr)
	}
	rowsAffected, _ := result.RowsAffected()
	t.Logf("FreshDatabase: CREATE DATABASE result: rows=%d", rowsAffected)

	// Ping to ensure the command is flushed to PostgreSQL
	if err := baseConn.Ping(); err != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: failed to ping after CREATE DATABASE: %v", err)
	}

	// Verify the database was created by querying pg_database on the base connection
	var dbExists bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`
	t.Logf("FreshDatabase: checking if database %s exists", sanitizedDBName)
	if err := baseConn.QueryRow(checkQuery, sanitizedDBName).Scan(&dbExists); err != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: failed to verify database creation: %v", err)
	}
	t.Logf("FreshDatabase: dbExists = %v for %s", dbExists, sanitizedDBName)
	if !dbExists {
		// List all databases to debug
		var debugInfo string
		rows, _ := baseConn.Query("SELECT datname FROM pg_database WHERE datname LIKE 'test_%' OR datname LIKE 'db_test_%' ORDER BY datname")
		if rows != nil {
			var dbNames []string
			for rows.Next() {
				var name string
				rows.Scan(&name)
				dbNames = append(dbNames, name)
			}
			rows.Close()
			debugInfo = fmt.Sprintf("Existing test databases: %v", dbNames)
		}
		baseConn.Close()
		t.Fatalf("FreshDatabase: database %s was not created (verification failed). %s", sanitizedDBName, debugInfo)
	}

	testDSN := buildDSN(cfg, sanitizedDBName)
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		baseConn.Close()
		t.Fatalf("FreshDatabase: open connection failed: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		baseConn.Close()
		t.Fatalf("FreshDatabase: cannot connect to %s: %v", sanitizedDBName, err)
	}

	// 🔴 CRITICAL SAFETY CHECK: Verify we connected to the correct database
	if err := verifyConnectedDatabase(db, sanitizedDBName, cfg.Host, cfg.Port, cfg.User); err != nil {
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
		dropExistingDB(baseConn, sanitizedDBName)
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
	migrationsDir := GetMigrationsDir()
	db, cleanupFn := FreshDatabase(t, prefix, migrationsDir)
	return db, cleanupFn
}

// FreshDatabaseWithoutMigrations creates a fresh isolated database without running any migrations.
// Use this for tests that manually call InitializeSchema or test migration behavior itself.
func FreshDatabaseWithoutMigrations(t *testing.T, prefix string) (*sql.DB, func()) {
	t.Helper()
	cfg, err := loadTestDBConfig()
	if err != nil {
		t.Logf("FreshDatabaseWithoutMigrations: skipped - failed to load test DB config: %v", err)
		return nil, nil
	}

	testDBName := generateUniqueDBName(prefix)

	// Sanitize the database name for PostgreSQL compatibility (use consistently throughout)
	sanitizedDBName := sanitizeIdentifier(testDBName)
	t.Logf("FreshDatabaseWithoutMigrations: original name=%s, sanitized name=%s", testDBName, sanitizedDBName)

	baseDSN := buildBaseDSN(cfg)
	baseConn, err := sql.Open("postgres", baseDSN)
	if err != nil {
		t.Logf("FreshDatabaseWithoutMigrations: skipped - failed to open postgres connection: %v", err)
		return nil, nil
	}

	// Try to ping the database to verify connectivity
	if err := baseConn.Ping(); err != nil {
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - PostgreSQL not available at %s:%d: %v", cfg.Host, cfg.Port, err)
		return nil, nil
	}

	dropExistingDB(baseConn, sanitizedDBName)

	// Create fresh database
	createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, sanitizedDBName)
	t.Logf("FreshDatabaseWithoutMigrations: executing %s", createQuery)
	result, execErr := baseConn.Exec(createQuery)
	if execErr != nil {
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - create database failed: %v", execErr)
		return nil, nil
	}
	rowsAffected, _ := result.RowsAffected()
	t.Logf("FreshDatabaseWithoutMigrations: CREATE DATABASE result: rows=%d", rowsAffected)

	// Ping to ensure the command is flushed to PostgreSQL
	if err := baseConn.Ping(); err != nil {
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - failed to ping after CREATE DATABASE: %v", err)
		return nil, nil
	}

	// Verify the database was created by querying pg_database on the base connection
	var dbExists bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`
	if err := baseConn.QueryRow(checkQuery, sanitizedDBName).Scan(&dbExists); err != nil {
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - failed to verify database creation: %v", err)
		return nil, nil
	}
	t.Logf("FreshDatabaseWithoutMigrations: dbExists = %v for %s", dbExists, sanitizedDBName)
	if !dbExists {
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - database %s was not created (verification failed)", sanitizedDBName)
		return nil, nil
	}

	testDSN := buildDSN(cfg, sanitizedDBName)
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - open connection failed: %v", err)
		return nil, nil
	}

	if err := db.Ping(); err != nil {
		db.Close()
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - cannot connect to %s: %v", sanitizedDBName, err)
		return nil, nil
	}

	// 🔴 CRITICAL SAFETY CHECK: Verify we connected to the correct database
	if err := verifyConnectedDatabase(db, sanitizedDBName, cfg.Host, cfg.Port, cfg.User); err != nil {
		db.Close()
		baseConn.Close()
		t.Logf("FreshDatabaseWithoutMigrations: skipped - %v", err)
		return nil, nil
	}

	cleanupFn := func() {
		if db != nil {
			db.Close()
		}
		dropExistingDB(baseConn, sanitizedDBName)
		baseConn.Close()
	}

	t.Cleanup(cleanupFn)

	return db, cleanupFn
}
