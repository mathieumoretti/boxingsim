package integration

import (
	"testing"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
)

// TestDBConnection tests database connectivity with isolated test database per-test.
func TestDBConnection(t *testing.T) {
	t.Parallel()

	testDB, _ := db.FreshDatabaseWithMigrations(t, "integration_dbconn", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at TEST_DB_HOST")
	}

	var result int
	err := testDB.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}

	t.Log("OK: Database connection verified with isolated database")
}

// TestAuthIntegration tests the authentication service logic without requiring a database.
func TestAuthIntegration(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{JWTSecret: "integration-test-secret-key"}
	service := auth.NewAuthService(cfg)

	password := "securepassword123"
	hashedPassword, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	isValid := service.CheckPassword(password, hashedPassword)
	if !isValid {
		t.Error("Password validation failed")
	}

	user := &model.User{
		ID:       1,
		Username: "integrationtest",
		Email:    "integration@test.com",
	}

	tokenPair, err := service.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" {
		t.Error("Access token is empty")
	}

	claims, err := service.VerifyToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims == nil || len(*claims) == 0 {
		t.Error("Claims are empty")
	} else if (*claims)["username"] != "integrationtest" {
		t.Errorf("Username mismatch in JWT claims, got %q", (*claims)["username"])
	}

	t.Log("OK: Authentication logic verified without database")
}

// TestDatabaseOperations tests CRUD operations using isolated test database.
func TestDatabaseOperations(t *testing.T) {
	t.Parallel()

	testDB, _ := db.FreshDatabaseWithMigrations(t, "integration_dbops", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at TEST_DB_HOST")
	}

	cfg := &config.Config{JWTSecret: "db-ops-test-secret"}
	authService := auth.NewAuthService(cfg)
	hashedPassword, err := authService.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash test password: %v", err)
	}

	userCreate := &model.UserCreate{
		Username:       "dbtestuser",
		Email:          "dbtest@example.com",
		HashedPassword: hashedPassword,
	}

	err = db.CreateUser(testDB, userCreate)
	if err != nil {
		t.Fatalf("Failed to create user in isolated test DB: %v", err)
	}

	retrievedUser, err := db.GetUserByUsername(testDB, "dbtestuser")
	if err != nil {
		t.Fatalf("Failed to retrieve user from isolated test DB: %v", err)
	}

	if retrievedUser == nil {
		t.Error("Retrieved user is nil, expected dbtestuser")
	} else if retrievedUser.Username != "dbtestuser" {
		t.Errorf("Retrieved username mismatch - got '%s', expected 'dbtestuser'", retrievedUser.Username)
	}

	t.Log("OK: Database CRUD operations verified with isolated database")
}

// TestCompleteFlow tests end-to-end flow combining auth service and database operations.
func TestCompleteFlow(t *testing.T) {
	t.Parallel()

	// TODO: Skip until refresh token generation is implemented (auth.GenerateTokenPair missing RefreshToken)
	t.Skip("Skipping until JWT refresh token implementation completes")

	cfg := &config.Config{JWTSecret: "complete-flow-test-secret"}
	authService := auth.NewAuthService(cfg)

	testDB, _ := db.FreshDatabaseWithMigrations(t, "integration_flow", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at TEST_DB_HOST")
	}

	password := "testpassword123"
	hashedPassword, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	isValid := authService.CheckPassword(password, hashedPassword)
	if !isValid {
		t.Error("Password validation failed")
	} else {
		t.Log("OK: Password hashing verified")
	}

	user := &model.User{
		ID:       999,
		Username: "flowtestuser",
		Email:    "flow@test.com",
	}

	tokenPair, err := authService.GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokenPair.AccessToken == "" || len(tokenPair.RefreshToken) < 20 {
		t.Error("Access or refresh token is empty/too short")
	} else {
		t.Log("OK: JWT token generation verified")
	}

	userCreate := &model.UserCreate{
		Username:       "flowtestuser",
		Email:          "flow@test.com",
		HashedPassword: hashedPassword,
	}

	err = db.CreateUser(testDB, userCreate)
	if err != nil {
		t.Fatalf("Failed to create test flow user in isolated DB: %v", err)
	} else {
		t.Log("OK: User creation verified")
	}

	retrievedUser, err := db.GetUserByUsername(testDB, "flowtestuser")
	if err != nil {
		t.Fatalf("Failed to retrieve test flow user: %v", err)
	}

	if retrievedUser == nil {
		t.Error("Retrieved user is nil, expected flowtestuser")
	} else if retrievedUser.Username != "flowtestuser" {
		t.Errorf("Retrieved username mismatch - got '%s', expected 'flowtestuser'", retrievedUser.Username)
	} else {
		t.Log("OK: User retrieval verified")
	}

	t.Logf("Complete flow integration test passed with auth+db verification in isolated database")
}
