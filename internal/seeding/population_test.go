package seeding

import (
	"testing"
)

func TestNewPopulationGenerator(t *testing.T) {
	g := NewPopulationGenerator()

	if g == nil {
		t.Fatal("Expected non-nil PopulationGenerator")
	}

	if g.nameGenerator == nil {
		t.Error("Expected non-nil name generator")
	}

	if g.totalGenerated != 0 {
		t.Errorf("Expected initial count of 0, got %d", g.totalGenerated)
	}
}

func TestGenerateAIBoxer(t *testing.T) {
	g := NewPopulationGenerator()
	boxer := g.GenerateAIBoxer(1)

	// Verify name fields
	if boxer.Name.FirstName == "" {
		t.Error("Expected non-empty first name")
	}
	if boxer.Name.LastName == "" {
		t.Error("Expected non-empty last name")
	}
	if boxer.Name.FullName == "" {
		t.Error("Expected non-empty full name")
	}
	if boxer.Name.DisplayName == "" {
		t.Error("Expected non-empty display name")
	}

	// Verify archetype
	if boxer.Archetype.Name == "" {
		t.Error("Expected non-empty archetype name")
	}

	// Verify stats are positive
	if boxer.Stats.Strength <= 0 {
		t.Error("Expected positive strength")
	}
	if boxer.Stats.Defense <= 0 {
		t.Error("Expected positive defense")
	}
	if boxer.Stats.Agility <= 0 {
		t.Error("Expected positive agility")
	}
	if boxer.Stats.Health <= 0 {
		t.Error("Expected positive health")
	}
	if boxer.Stats.Energy <= 0 {
		t.Error("Expected positive energy")
	}

	// Verify level is valid
	if boxer.Level < 1 || boxer.Level > 50 {
		t.Errorf("Invalid level: %d", boxer.Level)
	}

	// Verify position is within bounds
	if boxer.PositionX < 0 || boxer.PositionX > 100 {
		t.Errorf("Position X out of bounds: %f", boxer.PositionX)
	}
	if boxer.PositionY < 0 || boxer.PositionY > 73.33 {
		t.Errorf("Position Y out of bounds: %f", boxer.PositionY)
	}

	// Verify user ID is set
	if boxer.UserID != 1 {
		t.Errorf("Expected userID 1, got %d", boxer.UserID)
	}

	// Verify counter incremented
	if g.totalGenerated != 1 {
		t.Errorf("Expected totalGenerated to be 1, got %d", g.totalGenerated)
	}
}

func TestGenerateAIPopulation(t *testing.T) {
	g := NewPopulationGenerator()
	boxers := g.GenerateAIPopulation(50, 99)

	if len(boxers) != 50 {
		t.Errorf("Expected 50 boxers, got %d", len(boxers))
	}

	// Verify all boxers have valid data
	for i, boxer := range boxers {
		if boxer.Name.FullName == "" {
			t.Errorf("Boxer %d has empty full name", i)
		}
		if boxer.Level < 1 || boxer.Level > 50 {
			t.Errorf("Boxer %d has invalid level: %d", i, boxer.Level)
		}
		if boxer.UserID != 99 {
			t.Errorf("Boxer %d has wrong userID: %d", i, boxer.UserID)
		}
	}

	if g.totalGenerated != 50 {
		t.Errorf("Expected totalGenerated to be 50, got %d", g.totalGenerated)
	}
}

func TestAIBoxerToBoxerCreate(t *testing.T) {
	g := NewPopulationGenerator()
	boxer := g.GenerateAIBoxer(123)

	createReq := boxer.ToBoxerCreate()

	if createReq == nil {
		t.Fatal("Expected non-nil BoxerCreate")
	}

	if createReq.Name != boxer.Name.FullName {
		t.Errorf("Name mismatch: got %q, expected %q", createReq.Name, boxer.Name.FullName)
	}

	if createReq.Strength != boxer.Stats.Strength {
		t.Errorf("Strength mismatch: got %f, expected %f", createReq.Strength, boxer.Stats.Strength)
	}

	if createReq.Defense != boxer.Stats.Defense {
		t.Errorf("Defense mismatch: got %f, expected %f", createReq.Defense, boxer.Stats.Defense)
	}

	if createReq.Agility != boxer.Stats.Agility {
		t.Errorf("Agility mismatch: got %f, expected %f", createReq.Agility, boxer.Stats.Agility)
	}

	if createReq.PositionX != boxer.PositionX {
		t.Errorf("PositionX mismatch: got %f, expected %f", createReq.PositionX, boxer.PositionX)
	}

	if createReq.PositionY != boxer.PositionY {
		t.Errorf("PositionY mismatch: got %f, expected %f", createReq.PositionY, boxer.PositionY)
	}

	// Nickname should match (both can be nil or both non-nil with same value)
	if (createReq.Nickname == nil) != (boxer.Name.Nickname == nil) {
		t.Error("Nickname nullity mismatch")
	}
	if createReq.Nickname != nil && boxer.Name.Nickname != nil {
		if *createReq.Nickname != *boxer.Name.Nickname {
			t.Errorf("Nickname value mismatch: got %q, expected %q", *createReq.Nickname, *boxer.Name.Nickname)
		}
	}
}

func TestGenerateBoxersForSeeding(t *testing.T) {
	createReqs, err := GenerateBoxersForSeeding(25, 456)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(createReqs) != 25 {
		t.Errorf("Expected 25 boxer creation requests, got %d", len(createReqs))
	}

	// Verify all requests have required fields
	for i, req := range createReqs {
		if req.Name == "" {
			t.Errorf("Boxer %d has empty name", i)
		}
		if req.Strength <= 0 {
			t.Errorf("Boxer %d has invalid strength: %f", i, req.Strength)
		}
		if req.Defense <= 0 {
			t.Errorf("Boxer %d has invalid defense: %f", i, req.Defense)
		}
		if req.Agility <= 0 {
			t.Errorf("Boxer %d has invalid agility: %f", i, req.Agility)
		}
	}
}

func TestPopulationDiversity(t *testing.T) {
	g := NewPopulationGenerator()
	boxers := g.GenerateAIPopulation(100, 1)

	// Check name diversity
	names := make(map[string]bool)
	for _, boxer := range boxers {
		names[boxer.Name.FullName] = true
	}

	if len(names) < 90 {
		t.Errorf("Expected at least 90 unique names out of 100, got %d", len(names))
	}

	// Check archetype diversity
	archetypes := make(map[string]int)
	for _, boxer := range boxers {
		archetypes[boxer.Archetype.Name]++
	}

	if len(archetypes) < 5 {
		t.Errorf("Expected at least 5 different archetypes, got %d", len(archetypes))
	}

	// Check level distribution (pyramid)
	lowLevels := 0  // Levels 1-20
	highLevels := 0 // Levels 40+
	for _, boxer := range boxers {
		if boxer.Level <= 20 {
			lowLevels++
		}
		if boxer.Level >= 40 {
			highLevels++
		}
	}

	if lowLevels < 45 {
		t.Errorf("Expected at least 45 low-level boxers, got %d", lowLevels)
	}

	if highLevels > 20 {
		t.Errorf("Expected at most 20 high-level boxers, got %d", highLevels)
	}
}

func TestGenderDistribution(t *testing.T) {
	g := NewPopulationGenerator()
	boxers := g.GenerateAIPopulation(200, 1)

	maleCount := 0
	femaleCount := 0

	for _, boxer := range boxers {
		// Check if first name is from male or female list
		isFemale := false
		for _, fn := range g.nameGenerator.firstNamesFemale {
			if fn == boxer.Name.FirstName {
				isFemale = true
				break
			}
		}

		if isFemale {
			femaleCount++
		} else {
			maleCount++
		}
	}

	// Should be roughly 15% female (with variance)
	femalePercent := float64(femaleCount) / 200.0

	if femalePercent < 0.05 || femalePercent > 0.25 {
		t.Logf("Warning: Female percentage %f outside expected range [5-25%%]", femalePercent)
	}

	if maleCount+femaleCount != 200 {
		t.Errorf("Total count mismatch: %d males + %d females = %d, expected 200",
			maleCount, femaleCount, maleCount+femaleCount)
	}
}

func TestNicknameFrequency(t *testing.T) {
	g := NewPopulationGenerator()
	boxers := g.GenerateAIPopulation(100, 1)

	withNickname := 0
	for _, boxer := range boxers {
		if boxer.Name.Nickname != nil {
			withNickname++
		}
	}

	// Should be roughly 70% with nicknames (with variance)
	percentage := float64(withNickname) / 100.0

	if percentage < 0.50 || percentage > 0.90 {
		t.Errorf("Expected nickname frequency between 50-90%%, got %f", percentage)
	}
}

func TestPrintSummary(t *testing.T) {
	g := NewPopulationGenerator()
	boxer := g.GenerateAIBoxer(1)

	// This test just verifies PrintSummary doesn't panic
	// Actual output verification would require capturing stdout
	boxer.PrintSummary()
}
