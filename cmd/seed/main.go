package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
)

func main() {
	fmt.Println("Starting database seeding...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get mode from command line argument or default to "reference"
	mode := getSeedModeFromArgs(os.Args[1:])
	if !isValidMode(mode) {
		validModesMsg := map[string]bool{"reference": true, "development": true}
		fmt.Println("Invalid mode:", mode+", using 'reference' as default")
		fmt.Printf("\nValid modes: ")
		for m := range validModesMsg {
			fmt.Print(m + " ")
		}
		fmt.Println()
		mode = "reference"
	}

	// Initialize database connection
	pgDB, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := pgDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Check if we can connect
	if err := pgDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Seed the database with mode parameter
	if err := db.SeedDatabase(pgDB.DB, mode); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("Database seeding completed successfully!")
}

// getSeedModeFromArgs extracts the mode from command line arguments.
func getSeedModeFromArgs(args []string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return "reference" // default mode
}

// isValidMode checks if a given mode is valid (idempotent seeding modes).
func isValidMode(mode string) bool {
	validModes := map[string]bool{
		"reference":   true,
		"development": true,
		"dev":         true, // alias for development
	}
	return validModes[mode]
}
