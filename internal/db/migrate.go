package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

// MigrateDatabase runs database migrations
func MigrateDatabase(db *sql.DB) error {
	// Run migrations
	log.Println("Running database migrations...")
	if err := goose.Up(db, "db/migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// ResetDatabase resets the database by rolling back all migrations and re-running them
func ResetDatabase(db *sql.DB) error {
	log.Println("Resetting database...")

	// Rollback all migrations
	if err := goose.DownTo(db, "db/migrations", 0); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	// Re-run all migrations
	if err := goose.Up(db, "db/migrations"); err != nil {
		return fmt.Errorf("failed to re-run migrations: %w", err)
	}

	log.Println("Database reset completed successfully")
	return nil
}

// StatusDatabase shows the current migration status
func StatusDatabase(db *sql.DB) error {
	log.Println("Database migration status:")
	if err := goose.Status(db, "db/migrations"); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(name string) error {
	// Generate up and down files with timestamp
	upFileName := fmt.Sprintf("db/migrations/%s_up.sql", name)
	downFileName := fmt.Sprintf("db/migrations/%s_down.sql", name)

	// Create empty up file
	upFile, err := os.Create(upFileName)
	if err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}
	defer upFile.Close()

	// Create empty down file
	downFile, err := os.Create(downFileName)
	if err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}
	defer downFile.Close()

	log.Printf("Created new migration files: %s and %s", upFileName, downFileName)
	return nil
}