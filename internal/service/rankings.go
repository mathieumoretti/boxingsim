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
