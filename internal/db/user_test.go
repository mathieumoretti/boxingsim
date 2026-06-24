package db

import (
	"database/sql"
	"testing"

	"boxing/internal/model"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Run migrations
	CreateTables()

	return db
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	user := &model.User{
		Username:    "testuser",
		Email:       "test@example.com",
		PasswordHash: "hashedpassword",
		Role:        "user",
	}

	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Test successful retrieval
	found, err := GetUserByID(db, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if found.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, found.ID)
	}

	if found.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, found.Username)
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
	user := &model.User{
		Username:    "testuser",
		Email:       "test@example.com",
		PasswordHash: "hashedpassword",
		Role:        "user",
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
	user := &model.User{
		Username:    "testuser",
		Email:       "test@example.com",
		PasswordHash: "hashedpassword",
		Role:        "user",
	}

	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Update user
	updatedUser := &model.User{
		ID:           user.ID,
		Username:    "updateduser",
		Email:       "updated@example.com",
		PasswordHash: "newhashedpassword",
		Role:        "admin",
	}

	if err := UpdateUser(db, updatedUser); err != nil {
		t.Fatal(err)
	}

	// Verify update
	found, err := GetUserByID(db, user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if found.Username != "updateduser" {
		t.Errorf("Expected username updateduser, got %s", found.Username)
	}

	if found.Email != "updated@example.com" {
		t.Errorf("Expected email updated@example.com, got %s", found.Email)
	}

	if found.Role != "admin" {
		t.Errorf("Expected role admin, got %s", found.Role)
	}
}

func TestListUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create multiple test users
	users := []*model.User{
		{Username: "user1", Email: "user1@example.com", PasswordHash: "pass1", Role: "user"},
		{Username: "user2", Email: "user2@example.com", PasswordHash: "pass2", Role: "user"},
		{Username: "user3", Email: "user3@example.com", PasswordHash: "pass3", Role: "admin"},
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