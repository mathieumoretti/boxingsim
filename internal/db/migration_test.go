package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeSchema(t *testing.T) {
	t.Skip("integration test - requires live PostgreSQL database")

	// Integration test using PostgreSQL database
	db := SetupTestDB(t)

	// Test schema initialization
	err := InitializeSchema(db)
	assert.NoError(t, err)

	// Test that schema can be run multiple times (idempotent)
	err = InitializeSchema(db)
	assert.NoError(t, err)
}
