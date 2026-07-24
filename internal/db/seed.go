package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/model"
)

// SeedData contains sample data for seeding the database
type SeedData struct {
	Users  []UserSeedData
	Boxers []BoxerSeedData
}

// UserSeedData represents a user for seeding purposes
type UserSeedData struct {
	Username string
	Email    string
	Password string
}

// BoxerSeedData represents a boxer for seeding purposes
type BoxerSeedData struct {
	Name      string
	Nickname  *string
	PositionX float64
	PositionY float64
	Strength  float64
	Defense   float64
	Agility   float64
}

// SeedDatabase populates the database with sample data
func SeedDatabase(db *sql.DB, authService *auth.AuthService) error {
	seedData := getSampleSeedData()

	// Create users first
	for _, userData := range seedData.Users {
		userCreate := &model.UserCreate{
			Username:       userData.Username,
			Email:          userData.Email,
			HashedPassword: userData.Password,
		}

		// Hash the password if it's not already hashed
		if !isPasswordHashed(userData.Password) {
			hashedPassword, err := authService.HashPassword(userData.Password)
			if err != nil {
				return fmt.Errorf("failed to hash password for user %s: %w", userData.Username, err)
			}
			userCreate.HashedPassword = hashedPassword
		}

		err := CreateUser(db, userCreate)
		if err != nil {
			log.Printf("Warning: failed to create user %s: %v", userData.Username, err)
			// Continue with other users even if one fails
		}
	}

	// Create boxers
	for _, boxerData := range seedData.Boxers {
		boxerCreate := &model.BoxerCreate{
			Name:      boxerData.Name,
			Nickname:  boxerData.Nickname,
			PositionX: boxerData.PositionX,
			PositionY: boxerData.PositionY,
			Strength:  boxerData.Strength,
			Defense:   boxerData.Defense,
			Agility:   boxerData.Agility,
		}

		err := CreateBoxer(db, boxerCreate)
		if err != nil {
			log.Printf("Warning: failed to create boxer %s: %v", boxerData.Name, err)
			// Continue with other boxers even if one fails
		}
	}

	return nil
}

// getSampleSeedData returns sample data for seeding the database
func getSampleSeedData() SeedData {
	users := []UserSeedData{
		{
			Username: "boxingfan",
			Email:    "boxingfan@example.com",
			Password: "password123",
		},
		{
			Username: "champ",
			Email:    "champ@example.com",
			Password: "champion123",
		},
		{
			Username: "puncher",
			Email:    "puncher@example.com",
			Password: "punchmaster",
		},
	}

	boxers := []BoxerSeedData{
		{
			Name:      "Mike Tyson",
			Nickname:  stringPtr("The Baddest Man on the Planet"),
			PositionX: 10.0,
			PositionY: 20.0,
			Strength:  85.0,
			Defense:   75.0,
			Agility:   90.0,
		},
		{
			Name:      "Muhammad Ali",
			Nickname:  stringPtr("The Greatest"),
			PositionX: 15.0,
			PositionY: 25.0,
			Strength:  80.0,
			Defense:   85.0,
			Agility:   95.0,
		},
		{
			Name:      "Floyd Mayweather",
			Nickname:  stringPtr("The Matrix"),
			PositionX: 5.0,
			PositionY: 10.0,
			Strength:  70.0,
			Defense:   95.0,
			Agility:   85.0,
		},
		{
			Name:      "Sugar Ray Leonard",
			Nickname:  stringPtr("The Lion"),
			PositionX: 12.0,
			PositionY: 18.0,
			Strength:  75.0,
			Defense:   80.0,
			Agility:   90.0,
		},
		{
			Name:      "Joe Frazier",
			Nickname:  stringPtr("The Executioner"),
			PositionX: 8.0,
			PositionY: 15.0,
			Strength:  90.0,
			Defense:   70.0,
			Agility:   75.0,
		},
	}

	return SeedData{
		Users:  users,
		Boxers: boxers,
	}
}

// isPasswordHashed checks if a password is already hashed
func isPasswordHashed(password string) bool {
	// Simple check - hashed passwords typically start with "$2a$" or "$2b$"
	return len(password) > 10 && (password[:4] == "$2a$" || password[:4] == "$2b$")
}

// stringPtr is a helper to create pointer to string
func stringPtr(s string) *string {
	return &s
}
