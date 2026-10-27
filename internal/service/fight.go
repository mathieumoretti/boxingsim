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
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/store"
)

type FightService struct {
	fightStore      *store.FightStore
	boxerStore      *store.BoxerStore
	eventStore      *store.ScheduledEventStore
	worldClockModel *model.WorldClockModel
	cfg             *config.Config
	db              *sql.DB
}

func NewFightService(fightStore *store.FightStore, boxerStore *store.BoxerStore, eventStore *store.ScheduledEventStore, worldClockModel *model.WorldClockModel, cfg *config.Config, db *sql.DB) *FightService {
	return &FightService{
		fightStore:      fightStore,
		boxerStore:      boxerStore,
		eventStore:      eventStore,
		worldClockModel: worldClockModel,
		cfg:             cfg,
		db:              db,
	}
}

func (s *FightService) BookFight(ctx context.Context, boxer1ID int,
	boxer2ID int, scheduledTime time.Time, round int,
) error {
	if boxer1ID <= 0 || boxer2ID <= 0 {
		return errors.New("invalid request parameters")
	}

	// Validate fight is scheduled in the future relative to game time (MAT-FIGHT-SCHEDULING)
	if s.worldClockModel != nil && s.db != nil && s.cfg != nil {
		gameTime, err := s.worldClockModel.GetCurrentGameTime(ctx, s.db)
		if err != nil {
			return fmt.Errorf("failed to get current game time: %w", err)
		}

		minScheduleHours := s.cfg.Fight.MinScheduleHours
		minScheduleTime := gameTime.Add(time.Duration(minScheduleHours) * time.Hour)

		if scheduledTime.Before(minScheduleTime) || scheduledTime.Before(gameTime) {
			return fmt.Errorf("fight must be scheduled at least %d hours from current game time (game time: %v, minimum: %v, requested: %v)",
				minScheduleHours, gameTime.Format("2006-01-02 15:04"), minScheduleTime.Format("2006-01-02 15:04"), scheduledTime.Format("2006-01-02 15:04"))
		}
	}

	// Validate boxers exist using boxerStore
	boxer1, err := s.boxerStore.GetByID(ctx, boxer1ID)
	if err != nil || boxer1 == nil {
		return fmt.Errorf("boxer does not exist: ID %d", boxer1ID)
	}

	boxer2, err := s.boxerStore.GetByID(ctx, boxer2ID)
	if err != nil || boxer2 == nil {
		return fmt.Errorf("boxer does not exist: ID %d", boxer2ID)
	}

	// Check if boxers are in active fights using fightStore
	inUse, _ := s.fightStore.BoxerInFight(ctx, boxer1ID)
	if inUse {
		return fmt.Errorf("%w: boxer %d is currently involved in another fight", store.ErrBoxerInUse, boxer1ID)
	}

	inUse2, _ := s.fightStore.BoxerInFight(ctx, boxer2ID)
	if inUse2 {
		return fmt.Errorf("%w: boxer %d is currently involved in another fight", store.ErrBoxerInUse, boxer2ID)
	}

	// Validate opponent match quality (MAT-102)
	validation := s.ValidateOpponentMatch(boxer1ID, boxer2ID)
	if !validation.Valid {
		warningsStr := strings.Join(validation.Warnings, ", ")
		return fmt.Errorf("invalid matchup: status=%s, warnings=[%s]", validation.Status, warningsStr)
	}

	// Create the fight and get the ID using fightStore
	fightID, err := s.fightStore.Create(ctx, boxer1ID, boxer2ID, &scheduledTime, round)
	if err != nil {
		return fmt.Errorf("failed to create fight: %w", err)
	}

	// Create a scheduled event for the fight simulation (MAT-99)
	if s.eventStore != nil && scheduledTime.After(time.Now()) {
		eventData, err := json.Marshal(map[string]interface{}{
			"fight_id": fightID,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal fight event data: %w", err)
		}

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
	return s.fightStore.GetActiveFights(ctx, statuses)
}

func (s *FightService) GetFightByID(ctx context.Context, id int) (*model.Fight, error) {
	if id <= 0 {
		return nil, errors.New("invalid fight id")
	}
	return s.fightStore.GetByID(ctx, id)
}

// GetUpcomingFightForBoxer retrieves the next upcoming fight for a specific boxer (MAT-106)
func (s *FightService) GetUpcomingFightForBoxer(ctx context.Context, boxerID int) (*boxerdb.UpcomingFightResponse, error) {
	if boxerID <= 0 {
		return nil, errors.New("invalid boxer id")
	}
	return s.fightStore.GetUpcomingFightForBoxer(ctx, boxerID)
}

// GetFightHistoryWithOpponents retrieves fight history for a boxer with opponent names (MAT-103).
func (s *FightService) GetFightHistoryWithOpponents(ctx context.Context, boxerID int) ([]*boxerdb.FightHistoryWithOpponent, error) {
	if boxerID <= 0 {
		return nil, errors.New("invalid boxer id")
	}
	return s.fightStore.GetFightHistoryWithOpponents(ctx, boxerID)
}

// CancelFight cancels a scheduled fight (MAT-140).
// Validates that:
// - Fight exists and is in 'scheduled' status
// - Fight scheduled_time is in the future relative to game time
// Cleans up associated scheduled events.
func (s *FightService) CancelFight(ctx context.Context, fightID int, reason string) error {
	if fightID <= 0 {
		return errors.New("invalid fight id")
	}

	// Get current game time for validation
	var gameTime time.Time
	if s.worldClockModel != nil && s.db != nil {
		gt, err := s.worldClockModel.GetCurrentGameTime(ctx, s.db)
		if err != nil {
			return fmt.Errorf("failed to get current game time: %w", err)
		}
		gameTime = gt
	}

	// Retrieve the fight using a transaction for consistency
	tx, err := s.fightStore.BeginFightTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // Will be ignored if committed

	fight, err := s.fightStore.GetByIDWithLockTx(ctx, tx, fightID)
	if err != nil {
		if errors.Is(err, store.ErrFightNotFound) {
			return fmt.Errorf("fight not found: %w", boxerdb.ErrBoxerNotExists)
		}
		return fmt.Errorf("failed to retrieve fight: %w", err)
	}

	// Validate fight is in 'scheduled' status
	if fight.Status != model.FightStatusScheduled {
		return fmt.Errorf("fight cannot be cancelled: current status is '%s' (only 'scheduled' fights can be cancelled)", fight.Status)
	}

	// Validate fight is scheduled in the future relative to game time
	if fight.ScheduledTime != nil && !gameTime.IsZero() {
		if !fight.ScheduledTime.After(gameTime) {
			return fmt.Errorf("fight cannot be cancelled: scheduled time (%v) is in the past relative to game time (%v)",
				fight.ScheduledTime.Format("2006-01-02 15:04"), gameTime.Format("2006-01-02 15:04"))
		}
	}

	// Cancel the fight (updates status to 'cancelled')
	cancelReason := reason
	if cancelReason == "" {
		cancelReason = "User requested cancellation"
	}

	if err := s.fightStore.CancelFight(ctx, fightID, cancelReason); err != nil {
		if errors.Is(err, store.ErrFightCannotCancel) {
			return fmt.Errorf("fight cannot be cancelled: %w", err)
		}
		return fmt.Errorf("failed to cancel fight: %w", err)
	}

	// Clean up associated scheduled events (mark as completed to prevent processing)
	if s.eventStore != nil {
		events, err := s.eventStore.GetByFightID(ctx, fightID)
		if err != nil {
			// Log but don't fail - the fight is cancelled even if event cleanup fails
			fmt.Printf("Warning: failed to retrieve events for fight %d: %v\n", fightID, err)
		} else {
			for _, event := range events {
				if err := s.eventStore.MarkAsCompleted(ctx, event.ID); err != nil {
					// Log but don't fail - other events may still be cleaned up
					fmt.Printf("Warning: failed to mark event %d as completed: %v\n", event.ID, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ValidateOpponentMatch validates a potential matchup between two boxers (MAT-102).
func (s *FightService) ValidateOpponentMatch(boxer1ID, boxer2ID int) MatchValidation {
	validation := MatchValidation{
		Valid:       false,
		Status:      "unknown",
		Warnings:    []string{},
		Suggestions: []string{},
	}

	ctx := context.Background()

	// Retrieve both boxers using boxerStore
	boxer1, err := s.boxerStore.GetByID(ctx, boxer1ID)
	if err != nil || boxer1 == nil {
		validation.Warnings = append(validation.Warnings, "Boxer 1 not found")
		return validation
	}

	boxer2, err := s.boxerStore.GetByID(ctx, boxer2ID)
	if err != nil || boxer2 == nil {
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
