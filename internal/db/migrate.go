package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"
)

const defaultMigrationsDir = "db/migrations"

// MigrateDatabase runs all pending migrations using the default migrations directory.
// This is the entry point used by main.go for automatic migration on server startup.
func MigrateDatabase(db *sql.DB) error {
	return MigrateUp(db, defaultMigrationsDir)
}

// MigrateUp runs all pending migrations up to the latest one.
func MigrateUp(db *sql.DB, dir string) error {
	log.Println("Running database migrations...")
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// MigrateDownToZero rolls back all applied migrations.
func MigrateDownToZero(db *sql.DB, dir string) error {
	log.Println("Rolling back all migrations...")
	if err := goose.DownTo(db, dir, 0); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	log.Println("All migrations rolled back successfully")
	return nil
}

// ResetDatabase resets the database by rolling back all migrations and re-running them.
func ResetDatabase(db *sql.DB, dir string) error {
	log.Println("Resetting database...")

	if err := goose.DownTo(db, dir, 0); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("failed to re-run migrations: %w", err)
	}

	log.Println("Database reset completed successfully")
	return nil
}

// ShowStatus shows the current migration status.
func ShowStatus(db *sql.DB, dir string) error {
	if err := goose.Status(db, dir); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// MigrateUpOne runs the next single pending migration.
func MigrateUpOne(db *sql.DB, dir string) error {
	log.Println("Running one pending migration (up)...")
	if err := goose.UpByOne(db, dir); err != nil {
		return fmt.Errorf("failed to run migration up by one: %w", err)
	}

	log.Println("Migration completed successfully")
	return nil
}

// MigrateDownOne rolls back the last applied migration.
func MigrateDownOne(db *sql.DB, dir string) error {
	log.Println("Rolling back one migration (down)...")
	if err := goose.Down(db, dir); err != nil {
		return fmt.Errorf("failed to run migration down by one: %w", err)
	}

	log.Println("Migration rollback completed successfully")
	return nil
}

// CreateMigration creates a new empty timestamped migration file.
func CreateMigration(name string, migrationsDir string) error {
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", timestamp, name)

	fullPath := filepath.Join(migrationsDir, filename)

	// Create empty file with goose up/down markers
	content := `-- +goose Up
-- SQL in this section is executed when the migration is applied.


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

`
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	log.Printf("Created new migration file at: %s", fullPath)
	return nil
}
