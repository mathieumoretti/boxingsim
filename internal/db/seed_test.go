package db

import (
	"testing"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedDatabase(t *testing.T) {
	// Skip test if running in CI environment (no database available)
	t.Skip("Skipping seed test - requires database connection")

	// Load config
	cfg := config.Load()

	// Initialize database
	dbConn, err := database.NewPostgresDB(cfg)
	require.NoError(t, err, "Failed to connect to database")
	defer func() {
		if dbConn != nil {
			_ = dbConn.Close()
		}
	}()

	// Create auth service for password hashing
	authService := auth.NewAuthService(cfg)

	// Test seeding
	err = SeedDatabase(dbConn.DB, authService)
	assert.NoError(t, err, "Seeding should not return an error")

	// Verify that users were created
	users, err := ListUsers(dbConn.DB)
	assert.NoError(t, err, "Should be able to list users")
	assert.GreaterOrEqual(t, len(users), 1, "Should have at least one user")

	// Verify that boxers were created
	// Note: We can't easily test boxer creation directly without a user ID,
	// but we can check the seed data structure is correct
	seedData := getSampleSeedData()
	assert.GreaterOrEqual(t, len(seedData.Users), 1, "Should have at least one user in seed data")
	assert.GreaterOrEqual(t, len(seedData.Boxers), 1, "Should have at least one boxer in seed data")
}

func TestGetSampleSeedData(t *testing.T) {
	seedData := getSampleSeedData()

	// Verify we have sample users
	assert.NotNil(t, seedData.Users)
	assert.GreaterOrEqual(t, len(seedData.Users), 1)

	// Verify we have sample boxers
	assert.NotNil(t, seedData.Boxers)
	assert.GreaterOrEqual(t, len(seedData.Boxers), 1)

	// Verify user data structure
	for _, user := range seedData.Users {
		assert.NotEmpty(t, user.Username, "Username should not be empty")
		assert.NotEmpty(t, user.Email, "Email should not be empty")
		assert.NotEmpty(t, user.Password, "Password should not be empty")
	}

	// Verify boxer data structure
	for _, boxer := range seedData.Boxers {
		assert.NotEmpty(t, boxer.Name, "Boxer name should not be empty")
		assert.NotNil(t, boxer.PositionX, "Position X should not be nil")
		assert.NotNil(t, boxer.PositionY, "Position Y should not be nil")
		assert.NotNil(t, boxer.Strength, "Strength should not be nil")
		assert.NotNil(t, boxer.Defense, "Defense should not be nil")
		assert.NotNil(t, boxer.Agility, "Agility should not be nil")
	}
}

func TestIsPasswordHashed(t *testing.T) {
	// Test with a plain text password
	plainPassword := "password123"
	isHashed := isPasswordHashed(plainPassword)
	assert.False(t, isHashed, "Plain text password should not be recognized as hashed")

	// Test with a bcrypt hashed password (simulated)
	hashedPassword := "$2a$10$examplehashedpassword"
	isHashed = isPasswordHashed(hashedPassword)
	assert.True(t, isHashed, "Bcrypt hashed password should be recognized as hashed")
}

func TestStringPtr(t *testing.T) {
	s := "test string"
	ptr := stringPtr(s)
	assert.NotNil(t, ptr, "Pointer should not be nil")
	assert.Equal(t, s, *ptr, "Pointer should point to the correct value")
}
