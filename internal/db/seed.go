package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/model"
)

// SeedDatabase populates the database with sample data for demonstration purposes.
func SeedDatabase(db *sql.DB, mode string) error {
	fmt.Println("Seeding database with sample data (mode:", mode+")...")

	authService := auth.NewAuthService(nil) // nil config is fine - we just need password hashing functionality

	switch mode {
	case "reference":
		return seedReferenceData(db, authService)
	case "development", "dev":
		return seedDevelopmentData(db, authService)
	default:
		fmt.Println("Unknown mode:", mode+", using 'reference' as default")
		return seedReferenceData(db, authService)
	}
}

// createAdminUser creates or retrieves the admin user for championship boxers.
func createAdminUser(db *sql.DB, authService *auth.AuthService) (*model.User, error) {
	adminUser, _ := GetUserByUsername(db, "admin")
	if adminUser != nil {
		fmt.Printf("  Admin user already exists (ID: %d)\n", adminUser.ID)
		return adminUser, nil
	}

	passwordHash, err := authService.HashPassword("BoxingSim123!")
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	err = CreateUser(db, &model.UserCreate{
		Username:       "admin",
		Email:          "admin@boxingsim.local",
		HashedPassword: passwordHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	adminUser, _ = GetUserByUsername(db, "admin")
	fmt.Printf("  Created admin user with ID: %d\n", adminUser.ID)
	return adminUser, nil
}

type championBoxer struct {
	name     string
	nickname string
	strength float64
	defense  float64
	agility  float64
}

// seedReferenceData creates championship boxer references.
func seedReferenceData(db *sql.DB, authService *auth.AuthService) error {
	fmt.Println("Seeding with championship boxer references...")

	adminUser, err := createAdminUser(db, authService)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	champions := []championBoxer{
		{"Mike Tyson", "Iron Mike", 85.0, 72.0, 91.0},
		{"Floyd Mayweather Jr.", "Money", 78.0, 95.0, 82.0},
		{"Muhammad Ali", "The Greatest", 83.0, 68.0, 94.0},
		{"Manny Pacquiao", "PacMan", 79.0, 71.0, 88.0},
		{"Sugar Ray Leonard", "High on Action", 75.0, 76.0, 93.0},
		{"Roberto Duran", "Hands of Stone", 74.0, 82.0, 81.0},
		{"Thomas Hearns", "The Hitman", 77.0, 58.0, 86.0},
		{"Marvelous Marvin Hagler", "Marvelous", 89.0, 73.0, 72.0},
		{"Oscar De La Hoya", "The Golden Boy", 71.0, 65.0, 84.0},
		{"George Foreman", "Big George", 96.0, 70.0, 45.0},
		{"Joe Frazier", "Smiling Joe", 82.0, 75.0, 73.0},
	}

	for _, champ := range champions {
		existingBoxers, _ := ListBoxerByName(db, champ.name)
		if len(existingBoxers) > 0 {
			fmt.Printf("  Skipping %s (already exists)\n", champ.name)
			continue
		}

		nicknameStr := champ.nickname
		boxer := &model.BoxerCreate{
			Name:      champ.name,
			Nickname:  &nicknameStr,
			PositionX: rand.Float64()*80 + 10,
			PositionY: rand.Float64()*53.33 + 10,
			Strength:  champ.strength,
			Defense:   champ.defense,
			Agility:   champ.agility,
		}

		createdBoxer, err := CreateBoxerForUser(db, adminUser.ID, boxer)
		if err != nil {
			log.Printf("Warning: Failed to create champion %s: %v", champ.name, err)
			continue
		}

		nicknameDisplay := "N/A"
		if createdBoxer.Nickname != nil && *createdBoxer.Nickname != "" {
			nicknameDisplay = *createdBoxer.Nickname
		}
		fmt.Printf("  Created championship boxer: %s (%s)\n", createdBoxer.Name, nicknameDisplay)
	}

	fmt.Println("Reference seeding completed!")
	return nil
}

type sampleUser struct {
	username   string
	email      string
	password   string
	boxerCount int
}

// seedDevelopmentData creates users with their personal boxers and training history.
func seedDevelopmentData(db *sql.DB, authService *auth.AuthService) error {
	seedErr := seedReferenceData(db, authService)
	if seedErr != nil && !strings.Contains(seedErr.Error(), "already exists") {
		log.Printf("Warning during reference seeding: %v", seedErr)
	}

	fmt.Println("\nCreating sample user accounts...")

	sampleUsers := []sampleUser{
		{"johndoe", "john.doe@example.com", "password123", 5},
		{"janedoe", "jane.doe@example.com", "securepass42", 3},
		{"boxerfan99", "fan@boxingsim.local", "fighting!champion", 7},
	}

	for _, su := range sampleUsers {
		err := createCompleteUserData(db, authService, &su)
		if err != nil {
			log.Printf("Warning: Failed to create user %s data: %v\n", su.username, err)
		} else {
			fmt.Println("  Created user:", su.username+" (password: "+su.password+")")
		}
	}

	return seedErr
}

// boxerTemplate defines a template for generating sample boxers.
type boxerTemplate struct {
	name     string
	nickname string
}

// createSampleBoxersForUser creates sample personal/fighter-style custom boxers.
func createSampleBoxersForUser(db *sql.DB, userID int, existingNames *map[string]bool, count int) error {
	templates := []boxerTemplate{
		{"Thunder Punch", "The Storm"},
		{"Iron Fist Fighter", ""},
		{"Storm Breaker", ""},
		{"Blaze Warrior", ""},
		{"Shadow Strike", ""},
		{"Lightning Bolt", "Quick Draw"},
		{"Steel Uppercut", ""},
		{"Night Hawk", ""},
		{"Golden Gloves", ""},
		{"Crimson Tide", ""},
	}

	var createdCount int
	for _, t := range templates {
		if createdCount >= count {
			break
		}

		name := fmt.Sprintf("%s %d", t.name, rand.Intn(100)+2000) // Add random suffix for uniqueness.

		nicknamePtr := (*string)(nil)
		if t.nickname != "" {
			nicknamePtr = &t.nickname
		}

		boxerCreate := &model.BoxerCreate{
			Name:      name,
			Nickname:  nicknamePtr,
			PositionX: rand.Float64()*80 + 10,
			PositionY: rand.Float64()*53.33 + 10,
			Strength:  rand.Float64()*50 + 25,
			Defense:   rand.Float64()*50 + 25,
			Agility:   rand.Float64()*50 + 25,
		}

		createdBoxer, err := CreateBoxerForUser(db, userID, boxerCreate)
		if err != nil {
			log.Printf("Warning: Failed to create boxer %s for user %d: %v", name, userID, err)
			continue
		}

		nicknameDisplay := "N/A"
		if createdBoxer.Nickname != nil && *createdBoxer.Nickname != "" {
			nicknameDisplay = *createdBoxer.Nickname
		}
		fmt.Printf("    - Created boxer: %s (%s)\n", createdBoxer.Name, nicknameDisplay)

		(*existingNames)[name] = true
		createdCount++
	}

	return nil
}

// createCompleteUserData creates a complete user profile with boxers and training.
func createCompleteUserData(db *sql.DB, authService *auth.AuthService, userData *sampleUser) error {
	existingUser, _ := GetUserByUsername(db, userData.username)

	if existingUser != nil {
		fmt.Printf("  User %s already exists (ID: %d)\n", userData.username, existingUser.ID)
		return updateExistingUserData(db, authService, userData, existingUser)
	}

	passwordHash, err := authService.HashPassword(userData.password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = CreateUser(db, &model.UserCreate{
		Username:       userData.username,
		Email:          userData.email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", userData.username, err)
	}

	// Get the newly created user ID
	newUser, _ := GetUserByUsername(db, userData.username)
	userID := newUser.ID

	existingNames := make(map[string]bool)
	err = createSampleBoxersForUser(db, userID, &existingNames, userData.boxerCount)
	if err != nil {
		return fmt.Errorf("failed to create boxers: %w", err)
	}

	boxers, _ := ListBoxersByUserID(db, userID)
	fmt.Printf("  Created user %s with %d boxers (password: %s)\n", userData.username, len(boxers), userData.password)
	return nil
}

// updateExistingUserData updates an existing user's data.
func updateExistingUserData(db *sql.DB, authService *auth.AuthService, userData *sampleUser,
	existingUser *model.User,
) error {
	boxers, err := ListBoxersByUserID(db, existingUser.ID)
	if err != nil {
		return fmt.Errorf("failed to list boxers: %w", err)
	}

	if len(boxers) >= userData.boxerCount {
		fmt.Printf("  User %s already has sufficient data (%d boxers)\n", userData.username, len(boxers))
		return nil
	}

	existingNames := make(map[string]bool)
	for _, b := range boxers {
		existingNames[b.Name] = true
	}

	newCount := userData.boxerCount - len(boxers)
	err = createSampleBoxersForUser(db, existingUser.ID, &existingNames, newCount)
	if err != nil {
		return fmt.Errorf("failed to add boxers: %w", err)
	}

	allBoxers, _ := ListBoxersByUserID(db, existingUser.ID)
	fmt.Printf("  Updated user %s to have %d boxers\n", userData.username, len(allBoxers))
	return nil
}
