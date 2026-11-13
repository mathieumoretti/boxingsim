package service

import (
	"testing"

	"github.com/mormm/boxing/internal/db"
	"github.com/stretchr/testify/assert"
)

// TestOpponentFilter_ApplyDefaults tests default value application for opponent filters
func TestOpponentFilter_ApplyDefaults(t *testing.T) {
	t.Run("applies default level range based on boxer level", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1}

		filter.ApplyDefaults(10, 30.0)

		assert.Equal(t, 7, filter.MinLevel, "MinLevel should be boxer level - 3")
		assert.Equal(t, 13, filter.MaxLevel, "MaxLevel should be boxer level + 3")
		assert.Equal(t, 30.0, filter.MinHealth, "MinHealth should default to 30.0")
		assert.Equal(t, 50, filter.MaxResults, "MaxResults should default to 50")
	})

	t.Run("clamps MinLevel to 1 when below", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1}

		filter.ApplyDefaults(2, 30.0)

		assert.Equal(t, 1, filter.MinLevel, "MinLevel should be clamped to 1")
		assert.Equal(t, 5, filter.MaxLevel, "MaxLevel should still be calculated normally")
	})

	t.Run("treats equal MinLevel/MaxLevel as +/- range", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1, MinLevel: 2, MaxLevel: 2}

		filter.ApplyDefaults(10, 30.0)

		assert.Equal(t, 8, filter.MinLevel, "MinLevel should be boxer level - range")
		assert.Equal(t, 12, filter.MaxLevel, "MaxLevel should be boxer level + range")
	})

	t.Run("respects explicit MinLevel and MaxLevel when different", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1, MinLevel: 5, MaxLevel: 15}

		filter.ApplyDefaults(10, 30.0)

		assert.Equal(t, 5, filter.MinLevel, "Explicit MinLevel should be preserved")
		assert.Equal(t, 15, filter.MaxLevel, "Explicit MaxLevel should be preserved")
	})

	t.Run("applies custom default health when provided", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1}

		filter.ApplyDefaults(10, 50.0)

		assert.Equal(t, 50.0, filter.MinHealth, "MinHealth should use custom default")
	})

	t.Run("respects explicit MinHealth when set", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1, MinHealth: 75.0}

		filter.ApplyDefaults(10, 30.0)

		assert.Equal(t, 75.0, filter.MinHealth, "Explicit MinHealth should be preserved")
	})

	t.Run("applies default MaxResults when not set", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1}

		filter.ApplyDefaults(10, 30.0)

		assert.Equal(t, 50, filter.MaxResults, "MaxResults should default to 50")
	})

	t.Run("respects explicit MaxResults when set", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1, MaxResults: 20}

		filter.ApplyDefaults(10, 30.0)

		assert.Equal(t, 20, filter.MaxResults, "Explicit MaxResults should be preserved")
	})
}

// TestOpponentFilter_EdgeCases tests edge cases for opponent filtering
func TestOpponentFilter_EdgeCases(t *testing.T) {
	t.Run("handles level 1 boxer", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1}

		filter.ApplyDefaults(1, 30.0)

		assert.Equal(t, 1, filter.MinLevel, "MinLevel clamped to 1")
		assert.Equal(t, 4, filter.MaxLevel, "MaxLevel is 1 + 3 = 4")
	})

	t.Run("handles high level boxer", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1}

		filter.ApplyDefaults(100, 30.0)

		assert.Equal(t, 97, filter.MinLevel, "MinLevel is 100 - 3 = 97")
		assert.Equal(t, 103, filter.MaxLevel, "MaxLevel is 100 + 3 = 103")
	})

	t.Run("range exceeds boxer level for low levels", func(t *testing.T) {
		filter := db.OpponentFilter{BoxerID: 1, MinLevel: 5, MaxLevel: 5}

		filter.ApplyDefaults(3, 30.0)

		assert.Equal(t, 1, filter.MinLevel, "MinLevel clamped to 1 when below")
		assert.Equal(t, 8, filter.MaxLevel, "MaxLevel is 3 + 5 = 8")
	})
}

// TestScoringWeights verifies the scoring weights are correctly defined
func TestScoringWeights(t *testing.T) {
	t.Run("weights sum to 1.0", func(t *testing.T) {
		sum := WeightLevelDiff + WeightRanking + WeightWinRate + WeightStyleVariety
		assert.Equal(t, 1.0, sum, "All weights should sum to 1.0")
	})

	t.Run("level weight is 40%", func(t *testing.T) {
		assert.Equal(t, 0.40, WeightLevelDiff, "Level proximity should be 40% of score")
	})

	t.Run("ranking weight is 30%", func(t *testing.T) {
		assert.Equal(t, 0.30, WeightRanking, "Ranking should be 30% of score")
	})

	t.Run("win rate weight is 20%", func(t *testing.T) {
		assert.Equal(t, 0.20, WeightWinRate, "Win rate should be 20% of score")
	})

	t.Run("style variety weight is 10%", func(t *testing.T) {
		assert.Equal(t, 0.10, WeightStyleVariety, "Style variety should be 10% of score")
	})
}

// TestScoreThresholds verifies the score threshold constants
func TestScoreThresholds(t *testing.T) {
	t.Run("max level difference is defined", func(t *testing.T) {
		assert.Equal(t, 10, MaxLevelDifference, "Max level difference should be 10")
	})

	t.Run("warn level difference is defined", func(t *testing.T) {
		assert.Equal(t, 5, WarnLevelDifference, "Warning threshold should be at 5 level difference")
	})

	t.Run("ideal match score threshold", func(t *testing.T) {
		assert.Equal(t, 0.85, IdealMatchScore, "Ideal match threshold should be 0.85")
	})

	t.Run("acceptable match score threshold", func(t *testing.T) {
		assert.Equal(t, 0.60, AcceptableMatchScore, "Acceptable match threshold should be 0.60")
	})

	t.Run("min health threshold", func(t *testing.T) {
		assert.Equal(t, 30, MinHealthThreshold, "Min health threshold should be 30")
	})
}
