package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	boxerdb "github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/store"
)

type FightService struct {
	db         *sql.DB
	eventStore *store.ScheduledEventStore
}

func NewFightService(db interface{}, eventStore *store.ScheduledEventStore) *FightService {
	pdb := db.(*PostgresDBWrapper)
	return &FightService{
		db:         pdb.Conn,
		eventStore: eventStore,
	}
}

type PostgresDBWrapper struct {
	Conn *sql.DB
}

func (s *FightService) BookFight(ctx context.Context, boxer1ID int,
	boxer2ID int, scheduledTime time.Time, round int,
) error {
	if boxer1ID <= 0 || boxer2ID <= 0 {
		return errors.New("invalid request parameters")
	}

	exists, err := boxerdb.BoxerExists(s.db, boxer1ID)
	if err != nil || !exists {
		return fmt.Errorf("boxer does not exist: ID %d", boxer1ID)
	}

	exists2, err := boxerdb.BoxerExists(s.db, boxer2ID)
	if err != nil || !exists2 {
		return fmt.Errorf("boxer does not exist: ID %d", boxer2ID)
	}

	inUse, _ := boxerdb.BoxerInFight(s.db, boxer1ID)
	if inUse {
		return fmt.Errorf("%w: boxer %d is currently involved in another fight", boxerdb.ErrBoxerInUse, boxer1ID)
	}

	inUse2, _ := boxerdb.BoxerInFight(s.db, boxer2ID)
	if inUse2 {
		return fmt.Errorf("%w: boxer %d is currently involved in another fight", boxerdb.ErrBoxerInUse, boxer2ID)
	}

	// Validate opponent match quality (MAT-102)
	validation := s.ValidateOpponentMatch(boxer1ID, boxer2ID)
	if !validation.Valid {
		warningsStr := strings.Join(validation.Warnings, ", ")
		return fmt.Errorf("invalid matchup: status=%s, warnings=[%s]", validation.Status, warningsStr)
	}

	st := scheduledTime
	fight := &model.FightCreate{
		Boxer1ID:      boxer1ID,
		Boxer2ID:      boxer2ID,
		ScheduledTime: &st,
		Round:         round,
	}

	// Create the fight and get the ID
	fightID, err := boxerdb.CreateFight(s.db, fight)
	if err != nil {
		return fmt.Errorf("failed to create fight: %w", err)
	}

	// Create a scheduled event for the fight simulation (MAT-99)
	if s.eventStore != nil && scheduledTime.After(time.Now()) {
		eventData, _ := json.Marshal(map[string]interface{}{
			"fight_id": fightID,
		})

		event := &model.ScheduledEvent{
			BoxerID:   boxer1ID, // Associate with first boxer
			EventType: model.EventTypeFightSimulate,
			EventTime: scheduledTime,
			Processed: false,
			EventData: model.EventData(eventData),
		}

		if err := s.eventStore.Create(ctx, event); err != nil {
			// Log the error but don't fail the fight creation
			// The fight exists and can be simulated manually if needed
			fmt.Printf("Warning: failed to create scheduled event for fight %d: %v\n", fightID, err)
		}
	}

	return nil
}

func (s *FightService) GetActiveFights(ctx context.Context, statuses []string) ([]*model.Fight, error) {
	if len(statuses) == 0 {
		statuses = []string{"scheduled", "in_progress"}
	}
	return boxerdb.GetActiveFights(s.db, statuses)
}

func (s *FightService) GetFightByID(ctx context.Context, id int) (*model.Fight, error) {
	if id <= 0 {
		return nil, errors.New("invalid fight id")
	}
	return boxerdb.GetFightByID(s.db, id)
}

// GetUpcomingFightForBoxer retrieves the next upcoming fight for a specific boxer (MAT-106)
func (s *FightService) GetUpcomingFightForBoxer(ctx context.Context, boxerID int) (*boxerdb.UpcomingFightResponse, error) {
	if boxerID <= 0 {
		return nil, errors.New("invalid boxer id")
	}
	return boxerdb.GetUpcomingFightForBoxer(s.db, boxerID)
}

// ValidateOpponentMatch validates a potential matchup between two boxers (MAT-102).
func (s *FightService) ValidateOpponentMatch(boxer1ID, boxer2ID int) MatchValidation {
	validation := MatchValidation{
		Valid:       false,
		Status:      "unknown",
		Warnings:    []string{},
		Suggestions: []string{},
	}

	// Retrieve both boxers
	boxer1, err := boxerdb.GetBoxerByID(s.db, boxer1ID)
	if err != nil {
		validation.Warnings = append(validation.Warnings, "Boxer 1 not found")
		return validation
	}

	boxer2, err := boxerdb.GetBoxerByID(s.db, boxer2ID)
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
	levelDiff := abs(boxer1.Level - boxer2.Level)
	if levelDiff > MaxLevelDifference {
		validation.Warnings = append(validation.Warnings, fmt.Sprintf("Level difference exceeds maximum allowed (%d)", levelDiff))
		validation.Status = "mismatched"
		return validation
	}

	// Check health levels
	if boxer1.Health < MinHealthThreshold {
		validation.Warnings = append(validation.Warnings, fmt.Sprintf("Boxer 1 health is critically low (%.0f%%)", boxer1.Health))
	}
	if boxer2.Health < MinHealthThreshold {
		validation.Warnings = append(validation.Warnings, fmt.Sprintf("Boxer 2 health is critically low (%.0f%%)", boxer2.Health))
	}

	// Calculate match score using opponent scoring service
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
	if levelDiff > WarnLevelDifference {
		if boxer1.Level < boxer2.Level {
			validation.Suggestions = append(validation.Suggestions, "Consider finding a lower-level opponent for fairer matches")
		} else {
			validation.Suggestions = append(validation.Suggestions, "This opponent may be too weak - consider a higher-level challenge")
		}
	}

	if len(validation.Warnings) > 0 {
		validation.Suggestions = append(validation.Suggestions, "Address warnings before confirming fight")
	}

	if score.MatchQuality == "Perfect" {
		validation.Suggestions = append(validation.Suggestions, "Excellent matchup - both fighters are well-matched!")
	}

	return validation
}

// MatchValidation represents the result of opponent matchup validation.
type MatchValidation struct {
	Valid       bool         `json:"valid"`
	Status      string       `json:"status"` // "perfect", "good", "challenging", "mismatched"
	Warnings    []string     `json:"warnings"`
	Suggestions []string     `json:"suggestions"`
	Boxer1      *model.Boxer `json:"boxer1,omitempty"`
	Boxer2      *model.Boxer `json:"boxer2,omitempty"`
	MatchScore  float64      `json:"match_score"`
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
