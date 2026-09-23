package service

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
)

var (
	ErrOpponentNotFound     = errors.New("opponent not found")
	ErrInvalidMatchup       = errors.New("invalid matchup configuration")
	ErrLevelMismatchTooHigh = errors.New("level difference exceeds maximum allowed")
)

// OpponentDiscoveryService handles opponent discovery and matchmaking business logic.
type OpponentDiscoveryService struct {
	db     *sql.DB
	rand   *rand.Rand
}

// NewOpponentDiscoveryService creates a new OpponentDiscoveryService.
func NewOpponentDiscoveryService(database *sql.DB) *OpponentDiscoveryService {
	return &OpponentDiscoveryService{
		db:   database,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// FindOpponents retrieves and scores available opponents for a boxer.
func (s *OpponentDiscoveryService) FindOpponents(boxerID int, filters db.OpponentFilter) ([]*db.RankedOpponent, error) {
	filters.BoxerID = boxerID
	return db.FindOpponents(s.db, boxerID, filters)
}

// GetBoxerByID retrieves a boxer by ID.
func (s *OpponentDiscoveryService) GetBoxerByID(id int) (*model.Boxer, error) {
	return db.GetBoxerByID(s.db, id)
}

// ValidateOpponentMatch validates a potential matchup between two boxers.
// Note: This is now primarily handled by FightService.ValidateOpponentMatch for fight booking.
// This method provides a lightweight wrapper for external callers.
func (s *OpponentDiscoveryService) ValidateOpponentMatch(boxer1ID, boxer2ID int) MatchValidation {
	validation := MatchValidation{
		Valid:       false,
		Status:      "unknown",
		Warnings:    []string{},
		Suggestions: []string{},
	}

	// Retrieve both boxers
	boxer1, err := s.GetBoxerByID(boxer1ID)
	if err != nil {
		validation.Warnings = append(validation.Warnings, "Boxer 1 not found")
		return validation
	}

	boxer2, err := s.GetBoxerByID(boxer2ID)
	if err != nil {
		validation.Warnings = append(validation.Warnings, "Boxer 2 not found")
		return validation
	}

	validation.Boxer1 = boxer1
	validation.Boxer2 = boxer2

	// Check if boxers are the same
	if boxer1ID == boxer2ID {
		validation.Warnings = append(validation.Warnings, "Cannot fight yourself")
		validation.Status = "invalid"
		return validation
	}

	// Check level difference
	levelDiff := int(math.Abs(float64(boxer1.Level - boxer2.Level)))
	if levelDiff > MaxLevelDifference {
		validation.Warnings = append(validation.Warnings, "Level difference exceeds maximum allowed")
		validation.Status = "mismatched"
		return validation
	}

	// Check health levels
	if boxer1.Health < MinHealthThreshold {
		validation.Warnings = append(validation.Warnings, "Boxer 1 health is critically low")
	}
	if boxer2.Health < MinHealthThreshold {
		validation.Warnings = append(validation.Warnings, "Boxer 2 health is critically low")
	}

	// Calculate match score
	score := ScoreOpponent(boxer1, boxer2)
	validation.MatchScore = score.OverallScore

	// Determine match status based on score and warnings
	if len(validation.Warnings) == 0 && score.OverallScore >= IdealMatchScore {
		validation.Status = "perfect"
		validation.Valid = true
	} else if len(validation.Warnings) <= 1 && score.OverallScore >= AcceptableMatchScore {
		validation.Status = "good"
		validation.Valid = true
	} else if score.OverallScore >= 0.4 {
		validation.Status = "challenging"
		validation.Valid = true
	} else {
		validation.Status = "mismatched"
		validation.Valid = false
	}

	// Add warnings from scoring
	validation.Warnings = append(validation.Warnings, score.Warnings...)

	// Generate suggestions
	validation.Suggestions = s.generateSuggestions(boxer1, boxer2, levelDiff, score)

	return validation
}

// generateSuggestions creates helpful suggestions based on the matchup analysis.
func (s *OpponentDiscoveryService) generateSuggestions(boxer1, boxer2 *model.Boxer, levelDiff int, score OpponentScore) []string {
	var suggestions []string

	// Level mismatch suggestions
	if levelDiff > WarnLevelDifference {
		if boxer1.Level < boxer2.Level {
			suggestions = append(suggestions, "Consider finding a lower-level opponent for fairer matches")
		} else {
			suggestions = append(suggestions, "This opponent may be too weak - consider a higher-level challenge")
		}
	}

	// Health suggestions
	if boxer1.Health < MinHealthThreshold || boxer2.Health < MinHealthThreshold {
		suggestions = append(suggestions, "Consider scheduling rest events before fighting")
	}

	// Win rate mismatch suggestions
	playerWinRate := calculateWinRate(boxer1)
	opponentWinRate := calculateWinRate(boxer2)
	winRateDiff := math.Abs(playerWinRate - opponentWinRate)

	if winRateDiff > 0.4 {
		if playerWinRate < opponentWinRate {
			suggestions = append(suggestions, "This is a challenging match against an experienced fighter")
		} else {
			suggestions = append(suggestions, "Good opportunity to build confidence and gain experience")
		}
	}

	// Perfect match celebration
	if score.MatchQuality == "Perfect" {
		suggestions = append(suggestions, "Excellent matchup - both fighters are well-matched!")
	}

	return suggestions
}

// FindRandomOpponent returns a random opponent from the available pool.
func (s *OpponentDiscoveryService) FindRandomOpponent(boxerID int, minLevel, maxLevel int) (*db.RankedOpponent, error) {
	filters := db.OpponentFilter{
		BoxerID:       boxerID,
		MinLevel:      minLevel,
		MaxLevel:      maxLevel,
		IncludeAI:     true,
		ExcludeOwned:  true,
		AvailableOnly: true,
		HealthyOnly:   true,
		MinHealth:     30.0,
		MaxResults:    100,
	}

	opponents, err := s.FindOpponents(boxerID, filters)
	if err != nil {
		return nil, err
	}

	if len(opponents) == 0 {
		return nil, db.ErrNoOpponentsFound
	}

	// Pick random opponent
	idx := s.rand.Intn(len(opponents))
	return opponents[idx], nil
}

// GetScoredOpponents returns a scored and sorted list of opponents.
func (s *OpponentDiscoveryService) GetScoredOpponents(boxerID int, filters db.OpponentFilter) ([]*ScoredOpponent, error) {
	// Get the boxer for scoring
	boxer, err := s.GetBoxerByID(boxerID)
	if err != nil {
		return nil, err
	}

	// Find opponents
	opponents, err := s.FindOpponents(boxerID, filters)
	if err != nil {
		return nil, err
	}

	// Score and sort
	return ScoreRankedOpponents(boxer, opponents), nil
}

// GetTopMatches returns the top N matches for a boxer.
func (s *OpponentDiscoveryService) GetTopMatches(boxerID int, count int) ([]*ScoredOpponent, error) {
	if count <= 0 {
		count = 5 // Default to top 5
	}
	if count > 20 {
		count = 20 // Cap at 20
	}

	filters := db.OpponentFilter{
		BoxerID:       boxerID,
		IncludeAI:     true,
		ExcludeOwned:  true,
		AvailableOnly: true,
		HealthyOnly:   true,
		MinHealth:     30.0,
		MaxResults:    count * 2, // Get more to score and filter
		PreferRankings: true,
	}

	scored, err := s.GetScoredOpponents(boxerID, filters)
	if err != nil {
		return nil, err
	}

	// Return top N (already sorted by score)
	if len(scored) > count {
		scored = scored[:count]
	}

	return scored, nil
}

// CanFight checks if two boxers can currently fight (availability check).
func (s *OpponentDiscoveryService) CanFight(ctx context.Context, boxer1ID, boxer2ID int) (bool, error) {
	// Check if boxer1 is in a fight
	inFight1, err := db.BoxerInFight(s.db, boxer1ID)
	if err != nil {
		return false, err
	}
	if inFight1 {
		return false, nil
	}

	// Check if boxer2 is in a fight
	inFight2, err := db.BoxerInFight(s.db, boxer2ID)
	if err != nil {
		return false, err
	}
	if inFight2 {
		return false, nil
	}

	return true, nil
}
