package db

import (
	"testing"
)

func TestMigrations(t *testing.T) {
	// Test that migration functions exist
	t.Run("CreateTables", func(t *testing.T) {
		CreateTables()
	})

	t.Run("CreateDefaultAdmin", func(t *testing.T) {
		CreateDefaultAdmin()
	})
}