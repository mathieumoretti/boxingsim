package main

import (
	"log"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	dbConn, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if dbConn != nil {
			_ = dbConn.Close()
		}
	}()

	// Initialize auth service for password hashing
	authService := auth.NewAuthService(cfg)

	// Seed the database
	log.Println("Seeding database with sample data...")
	err = db.SeedDatabase(dbConn.DB, authService)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Database seeding completed successfully!")
}