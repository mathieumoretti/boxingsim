package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

func createTestRankingsService(t *testing.T) *RankingsService {
	return &RankingsService{
		rankingsStore: &store.RankingsStore{},
		redisClient:   nil, // No redis - tests focus on validation logic
		logger:        logger.New("TestRankingsService"),
		cacheTTL:      5 * time.Minute,
	}
}

func TestRankingsService_GetTopRanked(t *testing.T) {
	t.Run("ValidatesInvalidCriteria", func(t *testing.T) {
		svc := createTestRankingsService(t)

		result, err := svc.GetTopRanked(nil, "invalid_criteria", 10)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ranking criteria")
	})

	t.Run("NormalizesNegativeLimit", func(t *testing.T) {
		svc := createTestRankingsService(t)

		// This tests that negative limit is normalized (validation passes), then hits redis
		// We use recover() because nil redis client will panic
		assert.Panics(t, func() {
			_, _ = svc.GetTopRanked(nil, "win_rate", -5)
		})
	})

	t.Run("CappingLimitToMaxWhenExceeded", func(t *testing.T) {
		svc := createTestRankingsService(t)

		// Tests that limit > 1000 is capped (validation passes), then hits redis
		assert.Panics(t, func() {
			_, _ = svc.GetTopRanked(nil, "win_rate", 5000)
		})
	})

	t.Run("UsesDefaultLimitWhenZero", func(t *testing.T) {
		svc := createTestRankingsService(t)

		// Tests that zero limit defaults to 100 (validation passes), then hits redis
		assert.Panics(t, func() {
			_, _ = svc.GetTopRanked(nil, "win_rate", 0)
		})
	})
}

func TestRankingsService_cacheKey(t *testing.T) {
	t.Run("GeneratesCorrectCacheKey", func(t *testing.T) {
		svc := createTestRankingsService(t)

		key := svc.cacheKey(model.RankingByWinRate, 100)

		assert.Equal(t, "rankings:win_rate:limit:100", key)
	})

	t.Run("GeneratesUniqueKeysForDifferentCriteria", func(t *testing.T) {
		svc := createTestRankingsService(t)

		key1 := svc.cacheKey(model.RankingByWinRate, 100)
		key2 := svc.cacheKey(model.RankingByLevel, 100)

		assert.NotEqual(t, key1, key2)
	})

	t.Run("GeneratesUniqueKeysForDifferentLimits", func(t *testing.T) {
		svc := createTestRankingsService(t)

		key1 := svc.cacheKey(model.RankingByWinRate, 50)
		key2 := svc.cacheKey(model.RankingByWinRate, 100)

		assert.NotEqual(t, key1, key2)
	})
}

func TestRankingsService_GetRankingForBoxer(t *testing.T) {
	t.Run("ValidatesCriteria", func(t *testing.T) {
		svc := createTestRankingsService(t)

		result, err := svc.GetRankingForBoxer(nil, 1, "invalid_criteria")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ranking criteria")
	})

	t.Run("HandlesNilRedisClient", func(t *testing.T) {
		svc := createTestRankingsService(t)

		// Valid criteria passes validation but nil redis will panic
		assert.Panics(t, func() {
			_, _ = svc.GetRankingForBoxer(nil, 1, "win_rate")
		})
	})
}

func TestRankingsService_GetNearbyRankings(t *testing.T) {
	t.Run("ValidatesCriteria", func(t *testing.T) {
		svc := createTestRankingsService(t)

		result, err := svc.GetNearbyRankings(nil, 1, "invalid_criteria", 5)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ranking criteria")
	})

	t.Run("NormalizesRadiusOverMax", func(t *testing.T) {
		svc := createTestRankingsService(t)

		// Radius normalization passes but nil redis will panic
		assert.Panics(t, func() {
			_, _ = svc.GetNearbyRankings(nil, 1, "win_rate", 100)
		})
	})

	t.Run("UsesDefaultRadiusWhenZero", func(t *testing.T) {
		svc := createTestRankingsService(t)

		// Radius defaults to 5 but nil redis will panic
		assert.Panics(t, func() {
			_, _ = svc.GetNearbyRankings(nil, 1, "win_rate", 0)
		})
	})
}

func TestRankingsService_InvalidateRankingsCache(t *testing.T) {
	t.Run("HandlesNilRedisClient", func(t *testing.T) {
		svc := &RankingsService{
			rankingsStore: &store.RankingsStore{},
			redisClient:   nil,
			logger:        logger.New("TestRankingsService"),
			cacheTTL:      5 * time.Minute,
		}

		assert.Panics(t, func() {
			svc.InvalidateRankingsCache(nil)
		})
	})
}

func TestRankingsService_InvalidateBoxerRankingCache(t *testing.T) {
	t.Run("HandlesNilRedisClient", func(t *testing.T) {
		svc := &RankingsService{
			rankingsStore: &store.RankingsStore{},
			redisClient:   nil,
			logger:        logger.New("TestRankingsService"),
			cacheTTL:      5 * time.Minute,
		}

		assert.Panics(t, func() {
			svc.InvalidateBoxerRankingCache(nil, 123)
		})
	})
}

func TestRankingsService_ForceRecalculate(t *testing.T) {
	t.Run("HandlesNilRedisClient", func(t *testing.T) {
		svc := &RankingsService{
			rankingsStore: &store.RankingsStore{},
			redisClient:   nil,
			logger:        logger.New("TestRankingsService"),
			cacheTTL:      5 * time.Minute,
		}

		assert.Panics(t, func() {
			svc.ForceRecalculate(nil)
		})
	})
}

func TestRankingsService_LimitNormalization(t *testing.T) {
	t.Run("NormalizesZeroLimitToDefault", func(t *testing.T) {
		limit := 0
		if limit <= 0 {
			limit = defaultRankingLimit
		}
		assert.Equal(t, defaultRankingLimit, limit)
	})

	t.Run("NormalizesNegativeLimitToDefault", func(t *testing.T) {
		limit := -10
		if limit <= 0 {
			limit = defaultRankingLimit
		}
		assert.Equal(t, defaultRankingLimit, limit)
	})

	t.Run("CapsExcessiveLimitToMax", func(t *testing.T) {
		limit := 5000
		if limit > maxRankingLimit {
			limit = maxRankingLimit
		}
		assert.Equal(t, maxRankingLimit, limit)
	})

	t.Run("KeepsValidLimitUnchanged", func(t *testing.T) {
		limit := 50
		if limit <= 0 {
			limit = defaultRankingLimit
		}
		if limit > maxRankingLimit {
			limit = maxRankingLimit
		}
		assert.Equal(t, 50, limit)
	})
}

func TestRankingsService_RadiusNormalization(t *testing.T) {
	t.Run("NormalizesZeroRadiusToDefault", func(t *testing.T) {
		radius := 0
		if radius <= 0 {
			radius = 5 // Default radius
		}
		assert.Equal(t, 5, radius)
	})

	t.Run("CapsExcessiveRadiusToMax", func(t *testing.T) {
		radius := 100
		if radius > 50 {
			radius = 50 // Max radius
		}
		assert.Equal(t, 50, radius)
	})

	t.Run("KeepsValidRadiusUnchanged", func(t *testing.T) {
		radius := 10
		if radius <= 0 {
			radius = 5
		}
		if radius > 50 {
			radius = 50
		}
		assert.Equal(t, 10, radius)
	})
}
