package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeSchema(t *testing.T) {
	// Integration test using PostgreSQL database
	db := SetupTestDB(t)
	defer CleanupTestDB(db)

	// Test schema initialization
	err := InitializeSchema(db)
	assert.NoError(t, err)

	// Test that schema can be run multiple times (idempotent)
	err = InitializeSchema(db)
	assert.NoError(t, err)
}
