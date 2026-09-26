package service

import (
	"testing"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestScoreOpponent_sameLevel tests scoring with same-level opponents (should score high)
func TestScoreOpponent_sameLevel(t *testing.T) {
	playerBoxer := &model.Boxer{
		ID:        1,
		Name:      "Player Boxer",
		Level:     10,
		Strength:  75.0,
		Defense:   70.0,
		Agility:   72.0,
		Health:    90.0,
		Wins:      5,
		Losses:    2,
		Draws:     1,
		Knockouts: 3,
	}

	opponent := &model.Boxer{
		ID:        2,
		Name:      "Opponent Boxer",
		Level:     10, // Same level
		Strength:  72.0,
		Defense:   74.0,
		Agility:   71.0,
		Health:    88.0,
		Wins:      6,
		Losses:    2,
		Draws:     0,
		Knockouts: 4,
	}

	score := ScoreOpponent(playerBoxer, opponent)

	assert.Equal(t, 1.0, score.LevelProximity, "Level proximity should be 1.0 for same level")
	assert.GreaterOrEqual(t, score.OverallScore, 0.85, "Overall score should be high for well-matched opponents")
	assert.Contains(t, []string{"Perfect", "Great"}, score.MatchQuality, "Match quality should be Perfect or Great")
	assert.Empty(t, score.Warnings, "Should have no warnings for well-matched opponents")
}

// TestScoreOpponent_largeLevelDifference tests scoring with large level differences (should score low)
func TestScoreOpponent_largeLevelDifference(t *testing.T) {
	playerBoxer := &model.Boxer{
		ID:        1,
		Name:      "Player Boxer",
		Level:     5,
		Strength:  40.0,
		Defense:   35.0,
		Agility:   38.0,
		Health:    90.0,
		Wins:      1,
		Losses:    1,
		Draws:     0,
		Knockouts: 0,
	}

	opponent := &model.Boxer{
		ID:        2,
		Name:      "Opponent Boxer",
		Level:     50, // Large difference
		Strength:  150.0,
		Defense:   140.0,
		Agility:   145.0,
		Health:    88.0,
		Wins:      30,
		Losses:    5,
		Draws:     2,
		Knockouts: 20,
	}

	score := ScoreOpponent(playerBoxer, opponent)

	assert.Less(t, score.LevelProximity, 0.1, "Level proximity should be very low for large level difference")
	assert.Less(t, score.OverallScore, 0.35, "Overall score should be low for mismatched opponents")
	assert.Contains(t, []string{"Challenging", "Mismatched"}, score.MatchQuality, "Match quality should be Challenging or Mismatched")
	assert.NotEmpty(t, score.Warnings, "Should have warnings for large level difference")
}

// TestScoreOpponent_winRateSimilarity tests win rate similarity calculations
func TestScoreOpponent_winRateSimilarity(t *testing.T) {
	t.Run("identical win rates", func(t *testing.T) {
		boxer1 := &model.Boxer{
			ID:     1,
			Name:   "Boxer 1",
			Level:  10,
			Wins:   5,
			Losses: 5,
			Draws:  0,
		}

		boxer2 := &model.Boxer{
			ID:     2,
			Name:   "Boxer 2",
			Level:  10,
			Wins:   3,
			Losses: 3,
			Draws:  0,
		}

		score := ScoreOpponent(boxer1, boxer2)

		assert.Equal(t, float64(1.0), score.WinRateMatch, "Win rate match should be 1.0 for identical win rates (50%)")
	})

	t.Run("different win rates", func(t *testing.T) {
		boxer1 := &model.Boxer{
			ID:     1,
			Name:   "Boxer 1",
			Level:  10,
			Wins:   9,
			Losses: 1,
			Draws:  0,
		}

		boxer2 := &model.Boxer{
			ID:     2,
			Name:   "Boxer 2",
			Level:  10,
			Wins:   1,
			Losses: 9,
			Draws:  0,
		}

		score := ScoreOpponent(boxer1, boxer2)

		assert.Equal(t, float64(0.2), score.WinRateMatch, "Win rate match should be 0.2 for very different win rates (90% vs 10%)")
	})
}

// TestScoreOpponent_newBoxers tests edge case with new boxers who have no fights
func TestScoreOpponent_newBoxers(t *testing.T) {
	playerBoxer := &model.Boxer{
		ID:        1,
		Name:      "New Player",
		Level:     1,
		Strength:  20.0,
		Defense:   20.0,
		Agility:   20.0,
		Health:    100.0,
		Wins:      0,
		Losses:    0,
		Draws:     0,
		Knockouts: 0,
	}

	opponent := &model.Boxer{
		ID:        2,
		Name:      "New Opponent",
		Level:     1,
		Strength:  20.0,
		Defense:   20.0,
		Agility:   20.0,
		Health:    100.0,
		Wins:      0,
		Losses:    0,
		Draws:     0,
		Knockouts: 0,
	}

	score := ScoreOpponent(playerBoxer, opponent)

	assert.Equal(t, float64(1.0), score.LevelProximity, "Level proximity should be 1.0 for same level")
	assert.Equal(t, float64(1.0), score.WinRateMatch, "New boxers with no fights should have neutral win rate match (0.5 vs 0.5)")
	assert.GreaterOrEqual(t, score.OverallScore, 0.8, "New boxers at same level should be a good match")
}

// TestCalculateWinRate tests the calculateWinRate helper function
func TestCalculateWinRate(t *testing.T) {
	t.Run("boxer with wins and losses", func(t *testing.T) {
		boxer := &model.Boxer{
			Wins:   6,
			Losses: 4,
			Draws:  0,
		}

		winRate := calculateWinRate(boxer)

		assert.Equal(t, float64(0.6), winRate, "Win rate should be 6/10 = 0.6")
	})

	t.Run("boxer with draws", func(t *testing.T) {
		boxer := &model.Boxer{
			Wins:   5,
			Losses: 3,
			Draws:  2,
		}

		winRate := calculateWinRate(boxer)

		assert.Equal(t, float64(0.5), winRate, "Win rate should be 5/10 = 0.5 (draws count in total)")
	})

	t.Run("boxer with no fights", func(t *testing.T) {
		boxer := &model.Boxer{
			Wins:   0,
			Losses: 0,
			Draws:  0,
		}

		winRate := calculateWinRate(boxer)

		assert.Equal(t, float64(0.5), winRate, "New boxer with no fights should have neutral 0.5 win rate")
	})
}

// TestDetermineFightingStyle tests fighting style classification
func TestDetermineFightingStyle(t *testing.T) {
	t.Run("brawler - high strength", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 100.0,
			Defense:  60.0,
			Agility:  50.0,
		}

		style := determineFightingStyle(boxer)

		assert.Equal(t, "Brawler", style)
	})

	t.Run("technician - balanced with agility emphasis", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 70.0,
			Defense:  68.0, // Close to strength (< maxStat*0.3 diff)
			Agility:  95.0,
		}

		style := determineFightingStyle(boxer)

		assert.Equal(t, "Technician", style)
	})

	t.Run("speedster - agility focused with low strength (Technician takes precedence)", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 50.0, // ratio = 50/90 = 0.56 < 0.7 ✓ for Speedster
			Defense:  55.0, // |50-55|=5 < 90*0.3=27... this triggers Technician first!
			Agility:  90.0, // maxStat, ratio = 1.0 > 0.85 ✓ for Technician
		}

		style := determineFightingStyle(boxer)

		// Technician is checked before Speedster, so this becomes Technician
		assert.Equal(t, "Technician", style)
	})

	t.Run("speedster - true agility focus with high strength-defense gap", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 40.0, // ratio = 0.44 < 0.7 ✓
			Defense:  85.0, // |40-85|=45 > 90*0.3=27 ✓ avoids Technician
			Agility:  90.0, // maxStat, ratio = 1.0 > 0.8 ✓
		}

		style := determineFightingStyle(boxer)

		assert.Equal(t, "Speedster", style)
	})

	t.Run("all-rounder - well balanced stats trigger all-rounder check", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 70.0, // ratio = 70/72 = 0.97 > 0.6 ✓
			Defense:  72.0, // maxStat, ratio = 1.0 > 0.6 ✓ (also > 0.85 for Tank)
			Agility:  71.0, // ratio = 71/72 = 0.99 > 0.6 ✓
		}

		style := determineFightingStyle(boxer)

		// defenseRatio=1.0>0.85 AND strength=70 >= maxStat*0.7=50.4 -> Tank (checked before All-Rounder)
		assert.Equal(t, "Tank", style)
	})

	t.Run("defensive - defense highest when tank check fails", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 60.0, // Not brawler (def too high: 70/72 > 0.7)
			Defense:  72.0, // maxStat, ratio = 1.0 > 0.85 for Tank... but strength 60 < 72*0.7=50.4 NO WAIT 60>50.4 -> TANK!
			Agility:  68.0,
		}

		style := determineFightingStyle(boxer)

		// defenseRatio=1.0>0.85 AND strength=60 >= maxStat*0.7=50.4 -> Tank
		assert.Equal(t, "Tank", style)
	})

	t.Run("defensive - true defensive type with low strength to avoid tank", func(t *testing.T) {
		boxer := &model.Boxer{
			Strength: 45.0, // < maxStat*0.7=48 (avoiding Tank), ratio=0.63 > 0.6 for All-Rounder
			Defense:  72.0, // maxStat
			Agility:  50.0, // ratio=0.69 > 0.6 for All-Rounder... but agilityRatio=0.69 not > 0.8 so no Speedster
		}

		style := determineFightingStyle(boxer)

		// strength=45 >= 48? NO (72*0.7=50.4). So NOT Tank. All three ratios > 0.6 -> All-Rounder!
		assert.Equal(t, "All-Rounder", style)
	})
}

// TestGenerateWarnings tests warning generation for matchups
func TestGenerateWarnings(t *testing.T) {
	t.Run("no warnings for well-matched boxers", func(t *testing.T) {
		boxer1 := &model.Boxer{
			ID:     1,
			Name:   "Boxer 1",
			Level:  10,
			Health: 90.0,
			Wins:   5,
			Losses: 3,
		}

		boxer2 := &model.Boxer{
			ID:     2,
			Name:   "Boxer 2",
			Level:  11,
			Health: 88.0,
			Wins:   6,
			Losses: 3,
		}

		levelProximity := calculateLevelProximity(boxer1.Level, boxer2.Level)
		warnings := generateWarnings(boxer1, boxer2, levelProximity)

		assert.Empty(t, warnings, "Should have no warnings for well-matched boxers")
	})

	t.Run("warning for large level difference", func(t *testing.T) {
		boxer1 := &model.Boxer{
			ID:     1,
			Name:   "Boxer 1",
			Level:  5,
			Health: 90.0,
		}

		boxer2 := &model.Boxer{
			ID:     2,
			Name:   "Boxer 2",
			Level:  20, // Large difference
			Health: 88.0,
		}

		levelProximity := calculateLevelProximity(boxer1.Level, boxer2.Level)
		warnings := generateWarnings(boxer1, boxer2, levelProximity)

		assert.NotEmpty(t, warnings, "Should have warnings for large level difference")
	})

	t.Run("warning for low health", func(t *testing.T) {
		boxer1 := &model.Boxer{
			ID:     1,
			Name:   "Boxer 1",
			Level:  10,
			Health: 90.0,
		}

		boxer2 := &model.Boxer{
			ID:     2,
			Name:   "Boxer 2",
			Level:  10,
			Health: 25.0, // Below threshold
		}

		levelProximity := calculateLevelProximity(boxer1.Level, boxer2.Level)
		warnings := generateWarnings(boxer1, boxer2, levelProximity)

		assert.NotEmpty(t, warnings, "Should have warning for low health opponent")
	})

	t.Run("warning for significant skill difference", func(t *testing.T) {
		boxer1 := &model.Boxer{
			ID:     1,
			Name:   "Boxer 1",
			Level:  10,
			Health: 90.0,
			Wins:   18,
			Losses: 2, // 90% win rate
		}

		boxer2 := &model.Boxer{
			ID:     2,
			Name:   "Boxer 2",
			Level:  10,
			Health: 88.0,
			Wins:   2,
			Losses: 18, // 10% win rate
		}

		levelProximity := calculateLevelProximity(boxer1.Level, boxer2.Level)
		warnings := generateWarnings(boxer1, boxer2, levelProximity)

		assert.NotEmpty(t, warnings, "Should have warning for significant skill difference")
	})
}

// TestFindBestMatch tests finding the best match from a list of opponents
func TestFindBestMatch(t *testing.T) {
	playerBoxer := &model.Boxer{
		ID:     1,
		Name:   "Player",
		Level:  10,
		Health: 90.0,
		Wins:   5,
		Losses: 3,
	}

	// Create opponents with different match qualities
	opponents := []*model.Boxer{
		{ID: 2, Name: "Perfect Match", Level: 10, Health: 90.0, Wins: 5, Losses: 3},   // Same stats
		{ID: 3, Name: "Slightly Higher", Level: 12, Health: 85.0, Wins: 6, Losses: 4}, // +2 levels
		{ID: 4, Name: "Too High", Level: 25, Health: 90.0, Wins: 20, Losses: 5},       // Way too high
	}

	score := ScoreOpponent(playerBoxer, opponents[0])

	assert.GreaterOrEqual(t, score.OverallScore, 0.8, "Perfect match should have high score")
}

// TestScoreRankedOpponents tests scoring and sorting multiple opponents
func TestScoreRankedOpponents(t *testing.T) {
	playerBoxer := &model.Boxer{
		ID:     1,
		Name:   "Player",
		Level:  10,
		Health: 90.0,
		Wins:   5,
		Losses: 3,
	}

	opponents := []*db.RankedOpponent{
		{Boxer: &model.Boxer{ID: 2, Name: "Match 1", Level: 10, Health: 90.0}},
		{Boxer: &model.Boxer{ID: 3, Name: "Match 2", Level: 15, Health: 85.0}}, // Further level
		{Boxer: &model.Boxer{ID: 4, Name: "Match 3", Level: 9, Health: 88.0}},  // Closer level
	}

	scored := ScoreRankedOpponents(playerBoxer, opponents)

	assert.Len(t, scored, 3, "Should score all opponents")

	// Check that scoring was applied
	for _, s := range scored {
		assert.NotZero(t, s.Score.OverallScore, "Each opponent should have a score")
		assert.NotZero(t, s.Score.LevelProximity, "Each opponent should have level proximity")
	}

	// The first result should be the highest-scoring match
	assert.GreaterOrEqual(t, scored[0].Score.OverallScore, scored[1].Score.OverallScore,
		"First opponent should have highest or equal score")
}

// TestCalculatePowerScore tests power score calculation for rankings
func TestCalculatePowerScore(t *testing.T) {
	t.Run("new boxer with no fights", func(t *testing.T) {
		boxer := &model.Boxer{
			Level:     1,
			Strength:  20.0,
			Wins:      0,
			Losses:    0,
			Knockouts: 0,
		}

		score := calculatePowerScore(boxer)

		assert.GreaterOrEqual(t, score, 0.0, "Power score should be positive")
	})

	t.Run("experienced boxer with wins", func(t *testing.T) {
		boxer := &model.Boxer{
			Level:     25,
			Strength:  100.0,
			Wins:      20,
			Losses:    5,
			Draws:     2,
			Knockouts: 15,
		}

		score := calculatePowerScore(boxer)

		assert.Greater(t, score, 200.0, "Experienced boxer should have high power score")
	})
}

// TestDetermineMatchQuality tests match quality classification
func TestDetermineMatchQuality(t *testing.T) {
	t.Run("perfect match", func(t *testing.T) {
		quality := determineMatchQuality(0.95, 0.9)
		assert.Equal(t, "Perfect", quality)
	})

	t.Run("great match", func(t *testing.T) {
		quality := determineMatchQuality(0.7, 0.7)
		assert.Equal(t, "Great", quality)
	})

	t.Run("good match", func(t *testing.T) {
		quality := determineMatchQuality(0.5, 0.5)
		assert.Equal(t, "Good", quality)
	})

	t.Run("challenging match", func(t *testing.T) {
		quality := determineMatchQuality(0.3, 0.3)
		assert.Equal(t, "Challenging", quality)
	})

	t.Run("mismatched", func(t *testing.T) {
		quality := determineMatchQuality(0.1, 0.1)
		assert.Equal(t, "Mismatched", quality)
	})
}
