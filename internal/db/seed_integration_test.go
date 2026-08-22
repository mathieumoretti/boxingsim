//go:build integration

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSeedDatabaseIntegration tests the seeding functionality with live database.
func TestSeedDatabaseIntegration(t *testing.T) {
	testDB, _ := FreshDatabaseWithMigrations(t, "integration_seed", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at TEST_DB_HOST")
	}

	err := SeedDatabase(testDB, "test")
	assert.NoError(t, err)

	t.Log("OK: Database seeding completed with live database and migrations applied")
}
