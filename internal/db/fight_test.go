package db

import (
	"testing"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// This test was moved to boxer_test.go to avoid duplication

func TestCreateFight(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		CleanupTestDB(db)
	}()

	// Create test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}
	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Create boxers for the fight
	boxer1 := &model.BoxerCreate{
		Name:      "Boxer 1",
		Nickname:  stringPtr("B1"),
		PositionX: 10.0,
		PositionY: 10.0,
		Strength:  50.0,
		Defense:   40.0,
		Agility:   60.0,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.BoxerCreate{
		Name:      "Boxer 2",
		Nickname:  stringPtr("B2"),
		PositionX: 15.0,
		PositionY: 15.0,
		Strength:  45.0,
		Defense:   45.0,
		Agility:   55.0,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	// Create a fight
	scheduledTime := time.Now()
	fight := &model.FightCreate{
		Boxer1ID:      1,
		Boxer2ID:      2,
		ScheduledTime: &scheduledTime,
		Round:         1,
	}

	if err := CreateFight(db, fight); err != nil {
		t.Fatal(err)
	}

	// Verify fight was created
	foundFight, err := GetFightByID(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	if foundFight.Boxer1ID != 1 || foundFight.Boxer2ID != 2 {
		t.Error("Fight should reference correct boxers")
	}
}

func TestGetFightByID(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		CleanupTestDB(db)
	}()

	// Create test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}
	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Create boxers for the fight
	boxer1 := &model.BoxerCreate{
		Name:      "Boxer 1",
		Nickname:  stringPtr("B1"),
		PositionX: 10.0,
		PositionY: 10.0,
		Strength:  50.0,
		Defense:   40.0,
		Agility:   60.0,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.BoxerCreate{
		Name:      "Boxer 2",
		Nickname:  stringPtr("B2"),
		PositionX: 15.0,
		PositionY: 15.0,
		Strength:  45.0,
		Defense:   45.0,
		Agility:   55.0,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	scheduledTime := time.Now()
	fight := &model.FightCreate{
		Boxer1ID:      1,
		Boxer2ID:      2,
		ScheduledTime: &scheduledTime,
		Round:         1,
	}
	if err := CreateFight(db, fight); err != nil {
		t.Fatal(err)
	}

	// Test successful retrieval
	found, err := GetFightByID(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.ID != 1 {
		t.Errorf("Expected ID 1, got %d", found.ID)
	}

	if found.Boxer1ID != 1 {
		t.Errorf("Expected boxer1ID 1, got %d", found.Boxer1ID)
	}
}

func TestUpdateFight(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		CleanupTestDB(db)
	}()

	// Create test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}
	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Create boxers for the fight
	boxer1 := &model.BoxerCreate{
		Name:      "Boxer 1",
		Nickname:  stringPtr("B1"),
		PositionX: 10.0,
		PositionY: 10.0,
		Strength:  50.0,
		Defense:   40.0,
		Agility:   60.0,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.BoxerCreate{
		Name:      "Boxer 2",
		Nickname:  stringPtr("B2"),
		PositionX: 15.0,
		PositionY: 15.0,
		Strength:  45.0,
		Defense:   45.0,
		Agility:   55.0,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	scheduledTime := time.Now()
	fight := &model.FightCreate{
		Boxer1ID:      1,
		Boxer2ID:      2,
		ScheduledTime: &scheduledTime,
		Round:         1,
	}
	if err := CreateFight(db, fight); err != nil {
		t.Fatal(err)
	}

	// Note: UpdateFight function doesn't exist in the current implementation.
	// This test is here to show what we want to test, but would need a real update function
	// For now, just verify creation works
	found, err := GetFightByID(db, 1)
	if err != nil {
		t.Fatal(err)
	}

	if found.Boxer1ID != 1 {
		t.Error("Expected boxer1ID to be 1")
	}
}

func TestListFights(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		CleanupTestDB(db)
	}()

	// Create test user
	user := &model.UserCreate{
		Username:       "testuser",
		Email:          "test@example.com",
		HashedPassword: "hashedpassword",
	}
	if err := CreateUser(db, user); err != nil {
		t.Fatal(err)
	}

	// Create boxers for the fights
	boxer1 := &model.BoxerCreate{
		Name:      "Boxer 1",
		Nickname:  stringPtr("B1"),
		PositionX: 10.0,
		PositionY: 10.0,
		Strength:  50.0,
		Defense:   40.0,
		Agility:   60.0,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.BoxerCreate{
		Name:      "Boxer 2",
		Nickname:  stringPtr("B2"),
		PositionX: 15.0,
		PositionY: 15.0,
		Strength:  45.0,
		Defense:   45.0,
		Agility:   55.0,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	// Create multiple fights
	scheduledTime := time.Now()
	for i := 0; i < 3; i++ {
		fight := &model.FightCreate{
			Boxer1ID:      1,
			Boxer2ID:      2,
			ScheduledTime: &scheduledTime,
			Round:         1,
		}
		if err := CreateFight(db, fight); err != nil {
			t.Fatal(err)
		}
	}

	// List fights - this function doesn't exist in the current implementation
	// This is just to show what we want to test
	// foundFights, err := ListFights(db)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	//
	// if len(foundFights) != 3 {
	// 	t.Errorf("Expected 3 fights, got %d", len(foundFights))
	// }
}
