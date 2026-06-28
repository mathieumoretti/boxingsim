package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mormm/boxing/internal/model"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Run migrations
	InitializeSchema(db)

	return db
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}

	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Test successful retrieval
	found, err := GetUserByID(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.ID != 1 {
		t.Errorf("Expected ID %d, got %d", 1, found.ID)
	}

	if found.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", found.Username)
	}

	// Test not found
	_, err = GetUserByID(db, 9999)
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestGetUserByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}

	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Test successful retrieval
	found, err := GetUserByUsername(db, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	if found.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", found.Username)
	}

	// Test not found
	_, err = GetUserByUsername(db, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}

	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Update user
	updatedUser := &model.User{
		ID:             1,
		Username:       "updateduser",
		Email:          "updated@example.com",
		HashedPassword: "newhashedpassword",
	}

	if err := UpdateUser(db, updatedUser); err != nil {
		t.Fatal(err)
	}

	// Verify update
	found, err := GetUserByID(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.Username != "updateduser" {
		t.Errorf("Expected username updateduser, got %s", found.Username)
	}

	if found.Email != "updated@example.com" {
		t.Errorf("Expected email updated@example.com, got %s", found.Email)
	}
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create multiple test users
	users := []*model.UserCreate{
		{Username: "user1", Email: "user1@example.com", HashedPassword: "pass1"},
		{Username: "user2", Email: "user2@example.com", HashedPassword: "pass2"},
		{Username: "user3", Email: "user3@example.com", HashedPassword: "pass3"},
	}

	for _, user := range users {
		if err := CreateUser(db, user); err != nil {
			t.Fatal(err)
		}
	}

	// List users
	foundUsers, err := ListUsers(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(foundUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(foundUsers))
	}
}