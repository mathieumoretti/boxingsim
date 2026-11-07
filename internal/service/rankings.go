package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

const (
	defaultRankingCacheTTL = 5 * time.Minute
	defaultRankingLimit    = 100
	maxRankingLimit        = 1000
)

// RankingsService handles ranking-related business logic with caching
type RankingsService struct {
	rankingsStore *store.RankingsStore
	redisClient   *redis.Client
	logger        *logger.Logger
	cacheTTL      time.Duration
}

// NewRankingsService creates a new RankingsService
func NewRankingsService(rankingsStore *store.RankingsStore, redisClient *redis.Client, logger *logger.Logger) *RankingsService {
	return &RankingsService{
		rankingsStore: rankingsStore,
		redisClient:   redisClient,
		logger:        logger,
		cacheTTL:      defaultRankingCacheTTL,
	}
}

// cacheKey generates a Redis cache key for rankings
func (s *RankingsService) cacheKey(criteria model.RankingCriteria, limit int) string {
	return fmt.Sprintf("rankings:%s:limit:%d", criteria, limit)
}

// GetTopRanked retrieves the top ranked boxers by specified criteria with caching
func (s *RankingsService) GetTopRanked(ctx context.Context, criteriaStr string, limit int) ([]*model.RankedBoxer, error) {
	// Validate criteria
	criteria := model.RankingCriteria(criteriaStr)
	if !criteria.IsValid() {
		return nil, fmt.Errorf("invalid ranking criteria: %s", criteriaStr)
	}

	// Validate and normalize limit
	if limit <= 0 {
		limit = defaultRankingLimit
	}
	if limit > maxRankingLimit {
		limit = maxRankingLimit
	}

	// Try to get from cache
	cacheKey := s.cacheKey(criteria, limit)
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var boxers []*model.RankedBoxer
		if err := json.Unmarshal(cachedData, &boxers); err == nil {
			s.logger.Debug("Rankings cache hit", "criteria", criteria, "limit", limit)
			return boxers, nil
		}
	}

	// Cache miss or error - fetch from database
	s.logger.Debug("Rankings cache miss", "criteria", criteria, "limit", limit)
	boxers, err := s.rankingsStore.GetRankings(ctx, criteria, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get rankings: %w", err)
	}

	// Cache the result
	cachedJSON, err := json.Marshal(boxers)
	if err == nil {
		_, cacheErr := s.redisClient.Set(ctx, cacheKey, cachedJSON, s.cacheTTL).Result()
		if cacheErr != nil {
			s.logger.Warn("Failed to cache rankings", "error", cacheErr)
		}
	}

	return boxers, nil
}

// GetRankingForBoxer retrieves a specific boxer's ranking position
func (s *RankingsService) GetRankingForBoxer(ctx context.Context, boxerID int, criteriaStr string) (*model.RankingPosition, error) {
	// Validate criteria
	criteria := model.RankingCriteria(criteriaStr)
	if !criteria.IsValid() {
		return nil, fmt.Errorf("invalid ranking criteria: %s", criteriaStr)
	}

	cacheKey := fmt.Sprintf("ranking:boxer:%d:%s", boxerID, criteria)

	// Try to get from cache
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var position *model.RankingPosition
		if err := json.Unmarshal(cachedData, &position); err == nil {
			s.logger.Debug("Boxer ranking cache hit", "boxerID", boxerID, "criteria", criteria)
			return position, nil
		}
	}

	// Cache miss - fetch from database
	s.logger.Debug("Boxer ranking cache miss", "boxerID", boxerID, "criteria", criteria)
	position, err := s.rankingsStore.GetRankingByPosition(ctx, boxerID, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to get boxer ranking: %w", err)
	}

	// Cache the result
	cachedJSON, err := json.Marshal(position)
	if err == nil {
		_, cacheErr := s.redisClient.Set(ctx, cacheKey, cachedJSON, s.cacheTTL).Result()
		if cacheErr != nil {
			s.logger.Warn("Failed to cache boxer ranking", "error", cacheErr)
		}
	}

	return position, nil
}

// GetNearbyRankings retrieves boxers ranked near a specific boxer
func (s *RankingsService) GetNearbyRankings(ctx context.Context, boxerID int, criteriaStr string, radius int) ([]*model.RankedBoxer, error) {
	// Validate criteria
	criteria := model.RankingCriteria(criteriaStr)
	if !criteria.IsValid() {
		return nil, fmt.Errorf("invalid ranking criteria: %s", criteriaStr)
	}

	// Validate radius
	if radius <= 0 {
		radius = 5 // Default radius
	}
	if radius > 50 {
		radius = 50 // Max radius for performance
	}

	cacheKey := fmt.Sprintf("rankings:nearby:boxer:%d:%s:radius:%d", boxerID, criteria, radius)

	// Try to get from cache
	cachedData, err := s.redisClient.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var boxers []*model.RankedBoxer
		if err := json.Unmarshal(cachedData, &boxers); err == nil {
			s.logger.Debug("Nearby rankings cache hit", "boxerID", boxerID, "criteria", criteria)
			return boxers, nil
		}
	}

	// Cache miss - fetch from database
	s.logger.Debug("Nearby rankings cache miss", "boxerID", boxerID, "criteria", criteria)
	boxers, err := s.rankingsStore.GetNearbyRankings(ctx, boxerID, criteria, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby rankings: %w", err)
	}

	// Cache the result
	cachedJSON, err := json.Marshal(boxers)
	if err == nil {
		_, cacheErr := s.redisClient.Set(ctx, cacheKey, cachedJSON, s.cacheTTL).Result()
		if cacheErr != nil {
			s.logger.Warn("Failed to cache nearby rankings", "error", cacheErr)
		}
	}

	return boxers, nil
}

// InvalidateRankingsCache invalidates all ranking caches (call after fights or significant changes)
func (s *RankingsService) InvalidateRankingsCache(ctx context.Context) error {
	// Get all keys matching the pattern
	iter := s.redisClient.Scan(ctx, 0, "rankings:*", 100).Iterator()
	count := 0
	for iter.Next(ctx) {
		if err := s.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			s.logger.Warn("Failed to delete cache key", "key", iter.Val(), "error", err)
		}
		count++
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error scanning cache keys: %w", err)
	}

	s.logger.Info("Invalidated rankings cache", "keys_deleted", count)
	return nil
}

// InvalidateBoxerRankingCache invalidates a specific boxer's ranking caches
func (s *RankingsService) InvalidateBoxerRankingCache(ctx context.Context, boxerID int) error {
	// Delete individual boxer ranking keys
	iter := s.redisClient.Scan(ctx, 0, fmt.Sprintf("ranking:boxer:%d:*", boxerID), 100).Iterator()
	count := 0
	for iter.Next(ctx) {
		if err := s.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			s.logger.Warn("Failed to delete cache key", "key", iter.Val(), "error", err)
		}
		count++
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error scanning cache keys: %w", err)
	}

	s.logger.Debug("Invalidated boxer ranking cache", "boxerID", boxerID, "keys_deleted", count)
	return nil
}

// RecalculationResult holds the result of a force recalculation operation (MAT-115)
type RecalculationResult struct {
	Criteria  string   `json:"criteria"`
	CacheKeys int      `json:"cache_keys_invalidated"`
	BoxerCount int     `json:"boxer_count"`
	SampleRankings []*model.RankedBoxer `json:"sample_rankings"`
	Error     error    `json:"error,omitempty"`
}

// ForceRecalculate invalidates all ranking caches and fetches fresh rankings to verify calculation works (MAT-115)
// This is primarily for debugging/audit purposes to ensure cached rankings match database state.
func (s *RankingsService) ForceRecalculate(ctx context.Context) ([]RecalculationResult, error) {
	var results []RecalculationResult

	// Define criteria to recalculate
	criteriaList := []model.RankingCriteria{
		model.RankingByWinRate,
		model.RankingByTotalFights,
		model.RankingByLevel,
		model.RankingByStrength,
		model.RankingByPowerScore,
	}

	s.logger.Info("Starting force recalculation", "criteria_count", len(criteriaList))

	// First, invalidate all caches
	cacheKeysDeleted, err := s.invalidateAllCachesWithCount(ctx)
	if err != nil {
		s.logger.Error("Failed to invalidate caches during force recalculation", "error", err)
	}

	// Recalculate for each criteria
	for _, criteria := range criteriaList {
		result := RecalculationResult{
			Criteria: string(criteria),
		}

		// Fetch top 10 rankings for this criteria to verify calculation works
		boxers, err := s.GetTopRanked(ctx, string(criteria), 10)
		if err != nil {
			result.Error = fmt.Errorf("failed to fetch rankings for %s: %w", criteria, err)
			results = append(results, result)
			continue
		}

		result.BoxerCount = len(boxers)
		result.SampleRankings = boxers[:min(3, len(boxers))] // Include top 3 for verification

		s.logger.Info("Force recalculation successful", "criteria", criteria, "boxer_count", result.BoxerCount)
		results = append(results, result)
	}

	s.logger.Info("Force recalculation complete", "total_criteria", len(results), "cache_keys_invalidated", cacheKeysDeleted)
	return results, nil
}

// invalidateAllCachesWithCount invalidates all ranking caches and returns the count of deleted keys
func (s *RankingsService) invalidateAllCachesWithCount(ctx context.Context) (int, error) {
	// Get all keys matching the pattern (both rankings:* and ranking:boxer:*:)
	iter := s.redisClient.Scan(ctx, 0, "ranking*", 100).Iterator()
	count := 0
	for iter.Next(ctx) {
		if err := s.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			s.logger.Warn("Failed to delete cache key", "key", iter.Val(), "error", err)
		}
		count++
	}

	if err := iter.Err(); err != nil {
		return count, fmt.Errorf("error scanning cache keys: %w", err)
	}

	return count, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
