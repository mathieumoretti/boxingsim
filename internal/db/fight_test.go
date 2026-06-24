package db

import (
	"database/sql"
	"testing"

	"boxing/internal/model"
)

func TestCreateBoxer(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test users for the boxer
	owner := &model.User{
		Username:    "owner",
		Email:       "owner@example.com",
		PasswordHash: "pass",
		Role:        "user",
	}
	if err := CreateUser(db, owner); err != nil {
		t.Fatal(err)
	}

	// Create a boxer
	boxer := &model.Boxer{
		Name:        "Test Boxer",
		WeightClass: "featherweight",
		Height:      170,
		Reach:       172,
		Record:      "10-5-0",
		OwnerID:     owner.ID,
	}

	if err := CreateBoxer(db, boxer); err != nil {
		t.Fatal(err)
	}

	if boxer.ID == 0 {
		t.Error("Boxer ID should be set")
	}

	if boxer.Name != "Test Boxer" {
		t.Errorf("Expected name Test Boxer, got %s", boxer.Name)
	}
}

func TestCreateFight(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test boxers
	owner := &model.User{
		Username:    "owner",
		Email:       "owner@example.com",
		PasswordHash: "pass",
		Role:        "user",
	}
	if err := CreateUser(db, owner); err != nil {
		t.Fatal(err)
	}

	boxer1 := &model.Boxer{
		Name:        "Boxer 1",
		WeightClass: "featherweight",
		Height:      170,
		Reach:       172,
		Record:      "10-5-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.Boxer{
		Name:        "Boxer 2",
		WeightClass: "featherweight",
		Height:      175,
		Reach:       178,
		Record:      "8-3-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	// Create a fight
	fight := &model.Fight{
		Boxer1ID:     boxer1.ID,
		Boxer2ID:     boxer2.ID,
		EventID:      1, // Will be set after event creation
		WeightClass:  "featherweight",
		Stance:       "orthodox",
		Date:         "2025-01-15T10:00:00Z",
		Location:     "Las Vegas",
		Predictions:  0,
		HasResult:    false,
		Boxer1Win:    false,
		Boxer2Win:    false,
	}

	if err := CreateFight(db, fight); err != nil {
		t.Fatal(err)
	}

	if fight.ID == 0 {
		t.Error("Fight ID should be set")
	}

	if fight.Boxer1ID != boxer1.ID || fight.Boxer2ID != boxer2.ID {
		t.Error("Fight should reference correct boxers")
	}
}

func TestGetFightByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	owner := &model.User{
		Username:    "owner",
		Email:       "owner@example.com",
		PasswordHash: "pass",
		Role:        "user",
	}
	if err := CreateUser(db, owner); err != nil {
		t.Fatal(err)
	}

	boxer1 := &model.Boxer{
		Name:        "Boxer 1",
		WeightClass: "featherweight",
		Height:      170,
		Reach:       172,
		Record:      "10-5-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.Boxer{
		Name:        "Boxer 2",
		WeightClass: "featherweight",
		Height:      175,
		Reach:       178,
		Record:      "8-3-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	fight := &model.Fight{
		Boxer1ID:     boxer1.ID,
		Boxer2ID:     boxer2.ID,
		EventID:      1,
		WeightClass:  "featherweight",
		Stance:       "orthodox",
		Date:         "2025-01-15T10:00:00Z",
		Location:     "Las Vegas",
		Predictions:  0,
		HasResult:    false,
		Boxer1Win:    false,
		Boxer2Win:    false,
	}
	if err := CreateFight(db, fight); err != nil {
		t.Fatal(err)
	}

	// Test successful retrieval
	found, err := GetFightByID(db, fight.ID)
	if err != nil {
		t.Fatal(err)
	}

	if found.ID != fight.ID {
		t.Errorf("Expected ID %d, got %d", fight.ID, found.ID)
	}

	if found.Boxer1ID != boxer1.ID {
		t.Errorf("Expected boxer1ID %d, got %d", boxer1.ID, found.Boxer1ID)
	}
}

func TestUpdateFight(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	owner := &model.User{
		Username:    "owner",
		Email:       "owner@example.com",
		PasswordHash: "pass",
		Role:        "user",
	}
	if err := CreateUser(db, owner); err != nil {
		t.Fatal(err)
	}

	boxer1 := &model.Boxer{
		Name:        "Boxer 1",
		WeightClass: "featherweight",
		Height:      170,
		Reach:       172,
		Record:      "10-5-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.Boxer{
		Name:        "Boxer 2",
		WeightClass: "featherweight",
		Height:      175,
		Reach:       178,
		Record:      "8-3-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	fight := &model.Fight{
		Boxer1ID:     boxer1.ID,
		Boxer2ID:     boxer2.ID,
		EventID:      1,
		WeightClass:  "featherweight",
		Stance:       "orthodox",
		Date:         "2025-01-15T10:00:00Z",
		Location:     "Las Vegas",
		Predictions:  0,
		HasResult:    false,
		Boxer1Win:    false,
		Boxer2Win:    false,
	}
	if err := CreateFight(db, fight); err != nil {
		t.Fatal(err)
	}

	// Update fight with result
	updatedFight := &model.Fight{
		ID:           fight.ID,
		Boxer1ID:     boxer1.ID,
		Boxer2ID:     boxer2.ID,
		EventID:      1,
		WeightClass:  "featherweight",
		Stance:       "orthodox",
		Date:         "2025-01-15T10:00:00Z",
		Location:     "Las Vegas",
		Predictions:  100,
		HasResult:    true,
		Boxer1Win:    true,
		Boxer2Win:    false,
	}

	if err := UpdateFight(db, updatedFight); err != nil {
		t.Fatal(err)
	}

	// Verify update
	found, err := GetFightByID(db, fight.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !found.HasResult {
		t.Error("Expected HasResult to be true")
	}

	if found.Boxer1Win != true {
		t.Error("Expected Boxer1Win to be true")
	}

	if found.Boxer2Win != false {
		t.Error("Expected Boxer2Win to be false")
	}
}

func TestListFights(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	owner := &model.User{
		Username:    "owner",
		Email:       "owner@example.com",
		PasswordHash: "pass",
		Role:        "user",
	}
	if err := CreateUser(db, owner); err != nil {
		t.Fatal(err)
	}

	boxer1 := &model.Boxer{
		Name:        "Boxer 1",
		WeightClass: "featherweight",
		Height:      170,
		Reach:       172,
		Record:      "10-5-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer1); err != nil {
		t.Fatal(err)
	}

	boxer2 := &model.Boxer{
		Name:        "Boxer 2",
		WeightClass: "featherweight",
		Height:      175,
		Reach:       178,
		Record:      "8-3-0",
		OwnerID:     owner.ID,
	}
	if err := CreateBoxer(db, boxer2); err != nil {
		t.Fatal(err)
	}

	// Create multiple fights
	for i := 0; i < 3; i++ {
		fight := &model.Fight{
			Boxer1ID:     boxer1.ID,
			Boxer2ID:     boxer2.ID,
			EventID:      1,
			WeightClass:  "featherweight",
			Stance:       "orthodox",
			Date:         "2025-01-15T10:00:00Z",
			Location:     "Las Vegas",
			Predictions:  0,
			HasResult:    false,
			Boxer1Win:    false,
			Boxer2Win:    false,
		}
		if err := CreateFight(db, fight); err != nil {
			t.Fatal(err)
		}
	}

	// List fights
	foundFights, err := ListFights(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(foundFights) != 3 {
		t.Errorf("Expected 3 fights, got %d", len(foundFights))
	}
}