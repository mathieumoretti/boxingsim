//go:build integration

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitializeSchemaIntegration tests schema initialization with live database.
func TestInitializeSchemaIntegration(t *testing.T) {
	t.Parallel()

	testDB, _ := FreshDatabaseWithoutMigrations(t, "integration_schema")
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available at TEST_DB_HOST")
	}

	err := InitializeSchema(testDB)
	assert.NoError(t, err)

	// Test that schema can be run multiple times (idempotent)
	err = InitializeSchema(testDB)
	assert.NoError(t, err)

	t.Log("OK: Schema initialization is idempotent with live database")
}
