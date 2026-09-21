package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

	// Get count for population mode (only used for population seeding)
	count := getPopulationCountFromArgs(os.Args[1:])

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
	if err := db.SeedDatabase(pgDB.DB, mode, count); err != nil {
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
		"population":  true, // AI boxer population generation
	}
	return validModes[mode]
}

// getPopulationCountFromArgs extracts the --count parameter from arguments.
// Supports both "--count 50" and "--count=50" formats.
func getPopulationCountFromArgs(args []string) int {
	for i := 0; i < len(args); i++ {
		if args[i] == "--count" && i+1 < len(args) {
			// Format: --count 50
			count, err := strconv.Atoi(args[i+1])
			if err != nil {
				return 100 // default count
			}
			return count
		}
		// Handle format: --count=50
		if len(args[i]) > 8 && args[i][:7] == "--count" {
			countStr := args[i][8:] // Remove "--count=" prefix
			count, err := strconv.Atoi(countStr)
			if err != nil {
				return 100 // default count
			}
			return count
		}
	}
	return 100 // default count
}
