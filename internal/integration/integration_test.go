package integration

import (
	"testing"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/platform/config"
)

// TestDBConnection verifies that we can connect to the isolated test database.
func TestDBConnection(t *testing.T) {
	testDB, _ := db.FreshDatabaseWithMigrations(t, "integration_dbconn", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at DATABASE config")
	}
	// Note: cleanup is handled automatically via t.Cleanup in FreshDatabase

	var result int
	err := testDB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected SELECT 1 to return 1, got %d", result)
	}

	t.Log("OK: Database connection verified with isolated database")
}

// TestAuthIntegration tests the authentication service without database.
func TestAuthIntegration(t *testing.T) {
	// This test verifies auth logic without requiring database migrations
	// TODO: Implement once auth package is available

	t.Skip("Skipping until JWT refresh token implementation completes")

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "integration-test-secret-key"}}
	t.Logf("Config loaded with JWT secret length: %d", len(cfg.JWT.Secret))
}

// TestDatabaseOperations tests basic CRUD operations on isolated database.
func TestDatabaseOperations(t *testing.T) {
	testDB, _ := db.FreshDatabaseWithMigrations(t, "integration_dbops", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at DATABASE config")
	}
	// Note: cleanup is handled automatically via t.Cleanup in FreshDatabase

	// Verify we're connected to the correct isolated database
	var currentDB string
	err := testDB.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		t.Fatalf("Failed to query current database: %v", err)
	}

	t.Logf("TEST IS CONNECTED TO DATABASE: %s", currentDB)

	// Verify the database name starts with db_test_ prefix (our naming convention after sanitization)
	if len(currentDB) < 8 || currentDB[:8] != "db_test_" {
		t.Errorf("Expected database name to start with 'db_test_', got: %s", currentDB)
	}

	// Test basic table operations (tables should exist after migrations)
	var tableExists bool
	err = testDB.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'users'
		)
	`).Scan(&tableExists)
	if err != nil {
		t.Fatalf("Failed to check if users table exists: %v", err)
	}

	if !tableExists {
		t.Error("Users table does not exist after migrations")
	} else {
		t.Log("OK: Users table exists after migrations")
	}
}

// TestCompleteFlow tests end-to-end flow combining auth service and database.
func TestCompleteFlow(t *testing.T) {
	t.Skip("Skipping until JWT refresh token implementation completes")

	// TODO: Once auth and user services are complete, test full flow:
	// 1. Create isolated database with migrations
	// 2. Register new user via auth service
	// 3. Verify user exists in database
	// 4. Login and get JWT token
	// 5. Use token to access protected endpoint
}
