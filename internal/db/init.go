package db

import (
	"fmt"

	"github.com/mormm/boxing/internal/platform/database"
)

// InitDB initializes the database with required schema
func InitDB(db *database.PostgresDB) error {
	// Initialize the database schema
	err := InitializeSchema(db.DB)
	if err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return nil
}