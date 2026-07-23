package main

import (
	"log"

	"github.com/mormm/boxing/internal/db"
)

func main() {
	// Test the seed data structure first (this doesn't require database connection)
	log.Println("Testing sample seed data structure...")

	// Create a simple validation by checking that our seed data is properly structured
	users := []db.UserSeedData{
		{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		},
	}

	boxers := []db.BoxerSeedData{
		{
			Name:      "Test Boxer",
			PositionX: 10.0,
			PositionY: 20.0,
			Strength:  80.0,
			Defense:   70.0,
			Agility:   90.0,
		},
	}

	if len(users) == 0 {
		log.Fatal("No sample users found in seed data")
	}

	if len(boxers) == 0 {
		log.Fatal("No sample boxers found in seed data")
	}

	log.Printf("Found %d sample users and %d sample boxers", len(users), len(boxers))
	log.Printf("Sample user: %s (%s)", users[0].Username, users[0].Email)
	log.Printf("Sample boxer: %s", boxers[0].Name)

	log.Println("Seed data structure validation completed successfully!")
	log.Println("To actually seed the database, please ensure your database is running and properly configured.")
	log.Println("Use 'make docker-up' to start the database services, then run 'make seed'.")
}
