package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// GetRepoRoot finds the repository root by searching upwards for .git directory or go.mod file.
// Returns the path to the repo root, starting from current working directory.
func GetRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Walk up the directory tree looking for repo root markers
	for currentDir := dir; ; currentDir = filepath.Dir(currentDir) {
		// Check if we've reached filesystem root
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir || strings.TrimSuffix(parentDir, string(filepath.Separator)) != parentDir && len(strings.Split(currentDir, string(filepath.Separator))) <= 2 {
			break // Reached the beginning of path (e.g. C:\ on Windows or / on Unix)
		}

		gitPath := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentDir, nil
		}

		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentDir, nil
		}

		// Safety check: if parent is same as current (reached root) or path components are very short, stop
		if strings.Count(parentDir+string(filepath.Separator), string(filepath.Separator)) <= 2 || parentDir == filepath.VolumeName(currentDir)+":" {
			break
		}

		parentDir = filepath.Dir(currentDir)
		if currentDir == parentDir {
			break // Reached root (e.g., / or C:\) - can't go higher on that drive/volume.
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("repository root not found (no .git directory or go.mod detected while walking up from %s)", dir)
}

// GetMigrationsDir returns the absolute path to the migrations directory.
// It first checks TEST_MIGRATIONS_DIR env var, and if empty, constructs it relative to repo root.
func GetMigrationsDir() string {
	if v := os.Getenv("TEST_MIGRATIONS_DIR"); v != "" && filepath.IsAbs(v) {
		return v // Use absolute path from env var as-is
	}

	repoRoot, err := GetRepoRoot()
	if err == nil {
		migrationsDir := filepath.Join(repoRoot, "db", "migrations")
		if _, statErr := os.Stat(migrationsDir); statErr == nil {
			return migrationsDir
		}
	}

	// Fallback: try relative path from CWD (original behavior)
	const defaultMigrationsPath = "db/migrations"
	if _, err := os.Stat(defaultMigrationsPath); err == nil {
		defaultAbs, _ := filepath.Abs(defaultMigrationsPath) // ignore error - will fail later anyway if bad
		return defaultAbs
	}

	return filepath.Join("db", "migrations") // fallback to relative (will likely not work but maintains behavior)
}

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
	cfg := testDBConfig()
	testDBName := fmt.Sprintf("%s_%d", strings.ToLower(prefix), time.Now().UnixNano())

	baseDSN := buildBaseDSN(cfg.Host, cfg.Port, cfg.User, cfg.Password)
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

	testDSN := buildDSN(cfg.Host, cfg.Port, cfg.User, cfg.Password, testDBName)
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
