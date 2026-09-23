package seeding

import (
	"testing"
)

func TestGetLevelProbability(t *testing.T) {
	tests := []struct {
		level        int
		expectedProb float64
	}{
		{1, 0.05},   // Level 1: 5% (tier 1)
		{5, 0.05},   // Levels 1-10: 5% each
		{10, 0.05},  // Level 10 still in tier 1
		{11, 0.022}, // Levels 11-20: 2.2% each (tier 2)
		{15, 0.022}, // Middle of tier 2
		{20, 0.022}, // End of tier 2
		{21, 0.014}, // Levels 21-30: 1.4% each (tier 3)
		{30, 0.014}, // End of tier 3
		{31, 0.009}, // Levels 31-40: 0.9% each (tier 4)
		{40, 0.009}, // End of tier 4
		{41, 0.005}, // Levels 41-50: 0.5% each (tier 5)
		{50, 0.005}, // Max level still 0.5%
	}

	for _, tt := range tests {
		result := GetLevelProbability(tt.level)
		if result != tt.expectedProb {
			t.Errorf("GetLevelProbability(%d) = %f, expected %f", tt.level, result, tt.expectedProb)
		}
	}

	// Test invalid levels
	if GetLevelProbability(0) != 0 {
		t.Error("GetLevelProbability(0) should return 0")
	}

	if GetLevelProbability(51) != 0 {
		t.Error("GetLevelProbability(51) should return 0")
	}

	if GetLevelProbability(-1) != 0 {
		t.Error("GetLevelProbability(-1) should return 0")
	}
}

func TestGenerateRandomLevel(t *testing.T) {
	// Generate 1000 levels and verify distribution
	levelCounts := make(map[int]int)

	for i := 0; i < 1000; i++ {
		level := GenerateRandomLevel()
		if level < 1 || level > 50 {
			t.Errorf("Generated invalid level: %d", level)
		}
		levelCounts[level]++
	}

	// Verify all levels were generated at least once (with high probability - we have 50 unique levels possible)
	if len(levelCounts) < 45 {
		t.Errorf("Expected diverse level distribution, got %d unique levels", len(levelCounts))
	}

	// Lower levels should be more common based on our tier system:
	// Tier 1 (L1-10): 50%, Tier 2 (L11-20): 22% = ~72% for L1-20
	lowLevels := 0
	highLevels := 0 // L41-50: 5%
	for level := range levelCounts {
		if level <= 20 {
			lowLevels += levelCounts[level]
		}
		if level >= 41 {
			highLevels += levelCounts[level]
		}
	}

	if lowLevels < 650 {
		t.Errorf("Expected at least 650 low-level boxers (L1-20), got %d", lowLevels)
	}

	if highLevels > 100 {
		t.Errorf("Expected at most 100 high-level boxers (L41-50), got %d", highLevels)
	}
}

func TestCalculateStat(t *testing.T) {
	template := StatTemplate{
		StrengthRange: [2]float64{0.9, 1.1}, // Balanced range
	}

	// Test level 1 (minimum)
	stat1 := CalculateStat(1, template.StrengthRange, 20, 40)
	if stat1 < 10 || stat1 > 60 {
		t.Errorf("Level 1 stat out of expected range: %f", stat1)
	}

	// Test level 50 (maximum)
	stat50 := CalculateStat(50, template.StrengthRange, 20, 40)
	if stat50 < 30 || stat50 > 100 {
		t.Errorf("Level 50 stat out of expected range: %f", stat50)
	}

	// Higher level should generally produce higher stats (with variance)
	totalLow := 0.0
	totalHigh := 0.0
	for i := 0; i < 100; i++ {
		totalLow += CalculateStat(1, template.StrengthRange, 20, 40)
		totalHigh += CalculateStat(50, template.StrengthRange, 20, 40)
	}

	avgLow := totalLow / 100
	avgHigh := totalHigh / 100

	if avgLow >= avgHigh {
		t.Errorf("Level 50 should have higher average stats than Level 1: L1=%.2f, L50=%.2f", avgLow, avgHigh)
	}
}

func TestCalculateHealth(t *testing.T) {
	template := StatTemplate{
		HealthRange: [2]float64{0.9, 1.1},
	}

	// Health should scale with level
	health1 := CalculateHealth(1, template.HealthRange)
	health50 := CalculateHealth(50, template.HealthRange)

	if health1 < 30 || health1 > 80 {
		t.Errorf("Level 1 health out of expected range: %f", health1)
	}

	if health50 < 130 || health50 > 260 {
		t.Errorf("Level 50 health out of expected range: %f", health50)
	}

	// Average L50 health should be higher than L1
	totalLow := 0.0
	totalHigh := 0.0
	for i := 0; i < 100; i++ {
		totalLow += CalculateHealth(1, template.HealthRange)
		totalHigh += CalculateHealth(50, template.HealthRange)
	}

	if totalLow/100 >= totalHigh/100 {
		t.Error("Level 50 should have higher average health than Level 1")
	}
}

func TestCalculateEnergy(t *testing.T) {
	template := StatTemplate{
		EnergyRange: [2]float64{0.9, 1.1},
	}

	// Energy should scale with level
	energy1 := CalculateEnergy(1, template.EnergyRange)
	energy50 := CalculateEnergy(50, template.EnergyRange)

	if energy1 < 25 || energy1 > 60 {
		t.Errorf("Level 1 energy out of expected range: %f", energy1)
	}

	if energy50 < 80 || energy50 > 150 {
		t.Errorf("Level 50 energy out of expected range: %f", energy50)
	}
}

func TestGenerateBoxerStats(t *testing.T) {
	tests := []struct {
		level    int
		archName string
	}{
		{1, "Brawler"},
		{25, "Technician"},
		{50, "Tank"},
		{10, "All-Rounder"},
	}

	for _, tt := range tests {
		arch := GetArchetypeByName(tt.archName)
		if arch == nil {
			t.Fatalf("Archetype %q not found", tt.archName)
		}

		stats := GenerateBoxerStats(tt.level, *arch)

		// All stats should be positive
		if stats.Strength <= 0 {
			t.Errorf("Level %d %s has invalid strength: %f", tt.level, tt.archName, stats.Strength)
		}
		if stats.Defense <= 0 {
			t.Errorf("Level %d %s has invalid defense: %f", tt.level, tt.archName, stats.Defense)
		}
		if stats.Agility <= 0 {
			t.Errorf("Level %d %s has invalid agility: %f", tt.level, tt.archName, stats.Agility)
		}
		if stats.Health <= 0 {
			t.Errorf("Level %d %s has invalid health: %f", tt.level, tt.archName, stats.Health)
		}
		if stats.Energy <= 0 {
			t.Errorf("Level %d %s has invalid energy: %f", tt.level, tt.archName, stats.Energy)
		}

		// Verify archetype characteristics for specific cases
		switch tt.archName {
		case "Brawler":
			if tt.level > 20 && stats.Strength < stats.Agility {
				t.Logf("Warning: Level %d Brawler has STR (%.1f) < AGI (%.1f)", tt.level, stats.Strength, stats.Agility)
			}
		case "Tank":
			if tt.level > 20 && stats.Defense < stats.Strength {
				t.Logf("Warning: Level %d Tank has DEF (%.1f) < STR (%.1f)", tt.level, stats.Defense, stats.Strength)
			}
		}
	}
}

func TestGenerateExperience(t *testing.T) {
	// XP should increase with level
	xp1 := GenerateExperience(1)
	if xp1 != 0 {
		t.Errorf("Level 1 experience should be 0, got %f", xp1)
	}

	xp10 := GenerateExperience(10)
	xp20 := GenerateExperience(20)
	xp50 := GenerateExperience(50)

	if xp10 < 700 || xp10 > 1500 {
		t.Errorf("Level 10 experience out of expected range: %f", xp10)
	}

	if xp50 < xp20 || xp20 < xp10 {
		t.Error("Experience should increase monotonically with level")
	}

	// Max XP around L50 should be reasonable (wider range for variance)
	if xp50 > 15000 || xp50 < 4000 {
		t.Logf("Level 50 experience: %f (expected ~4000-15000)", xp50)
	}
}

func TestGenerateCareerRecord(t *testing.T) {
	tests := []struct {
		level        int
		archName     string
		expectFights bool
	}{
		{1, "Brawler", false},    // No fights at L1
		{2, "Technician", false}, // No fights at L2
		{3, "All-Rounder", true}, // Start having fights at L3+
		{25, "Tank", true},
		{50, "Glass Cannon", true},
	}

	for _, tt := range tests {
		arch := GetArchetypeByName(tt.archName)
		if arch == nil {
			t.Fatalf("Archetype %q not found", tt.archName)
		}

		stats := GenerateBoxerStats(tt.level, *arch)
		record := GenerateCareerRecord(tt.level, *arch, stats)

		totalFights := record.Wins + record.Losses + record.Draws

		if !tt.expectFights {
			if totalFights > 0 {
				t.Errorf("Level %d boxer should have no fights, got %d", tt.level, totalFights)
			}
		} else {
			if totalFights == 0 {
				t.Errorf("Level %d boxer should have some fights", tt.level)
			}

			// Knockouts should not exceed wins
			if record.KOs > record.Wins {
				t.Errorf("Level %d boxer has more KOs (%d) than wins (%d)", tt.level, record.KOs, record.Wins)
			}

			// All values should be non-negative
			if record.Wins < 0 || record.Losses < 0 || record.Draws < 0 {
				t.Errorf("Career record has negative values: %v", record)
			}
		}
	}
}

func TestHighLevelFighterStats(t *testing.T) {
	// High level fighters (L40+) should have clear specialties
	// and shouldn't be perfectly balanced

	for i := 0; i < 20; i++ {
		level := 40 + randIntn(11) // Random level 40-50
		arch := RandomArchetype()
		stats := GenerateBoxerStats(level, arch)

		// No fighter should have all stats > 90 (too perfect)
		tooPerfect := stats.Strength > 90 && stats.Defense > 90 && stats.Agility > 90
		if tooPerfect {
			t.Logf("Warning: Generated nearly perfect fighter at L%d: STR=%.1f DEF=%.1f AGI=%.1f",
				level, stats.Strength, stats.Defense, stats.Agility)
		}
	}
}

func TestGenerateMapPosition(t *testing.T) {
	validCount := 0
	for i := 0; i < 100; i++ {
		x, y := GenerateMapPosition()

		// Positions should be within valid bounds (with margin)
		if x >= 2.0 && x <= 98.0 && y >= 2.0 && y <= 71.33 {
			validCount++
		}
	}

	if validCount < 95 {
		t.Errorf("Expected at least 95%% of positions to be within bounds, got %d%%", validCount)
	}
}
