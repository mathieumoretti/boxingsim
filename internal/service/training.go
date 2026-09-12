package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

var (
	ErrTrainingNotFound     = errors.New("training session not found")
	ErrTrainingNotPending   = errors.New("training session is not in pending status")
	ErrInsufficientEnergy   = errors.New("insufficient energy to complete training")
	ErrTrainingTypeNotFound = errors.New("training type not found")
)

// TrainingService orchestrates training session completion logic
type TrainingService struct {
	boxerStore            *store.BoxerStore
	trainingTypeStore     *store.TrainingTypeStore
	trainingSessionStore  *store.TrainingSessionStore
	scheduledEventStore   *store.ScheduledEventStore
	fatigueService        *FatigueService
	progressionService    *ProgressionService
	worldClockModel       *model.WorldClockModel
	logger                *logger.Logger
}

// NewTrainingService creates a new TrainingService instance
func NewTrainingService(
	boxerStore *store.BoxerStore,
	trainingTypeStore *store.TrainingTypeStore,
	trainingSessionStore *store.TrainingSessionStore,
	scheduledEventStore *store.ScheduledEventStore,
	fatigueService *FatigueService,
	progressionService *ProgressionService,
	worldClockModel *model.WorldClockModel,
	lg *logger.Logger,
) *TrainingService {
	return &TrainingService{
		boxerStore:            boxerStore,
		trainingTypeStore:     trainingTypeStore,
		trainingSessionStore:  trainingSessionStore,
		scheduledEventStore:   scheduledEventStore,
		fatigueService:        fatigueService,
		progressionService:    progressionService,
		worldClockModel:       worldClockModel,
		logger:                lg,
	}
}

// CompleteTrainingSession completes a pending training session by:
// 1. Validating the session exists and is pending
// 2. Fetching boxer and validating ownership/constraints
// 3. Deducting energy cost from boxer
// 4. Applying planned stat gains to boxer
// 5. Updating boxer via boxerStore.Update()
// 6. Marking training session as completed
func (s *TrainingService) CompleteTrainingSession(ctx context.Context, sessionID int) error {
	// Step 1: Fetch and validate training session
	session, err := s.trainingSessionStore.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: ID=%d", ErrTrainingNotFound, sessionID)
		}
		return fmt.Errorf("failed to fetch training session %d: %w", sessionID, err)
	}

	if session.Status != model.TrainingSessionPending {
		return fmt.Errorf("%w: current status is %q", ErrTrainingNotPending, session.Status)
	}

	// Step 2: Fetch boxer
	boxer, err := s.boxerStore.GetByID(ctx, session.BoxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", session.BoxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", session.BoxerID, err)
	}

	// Step 3: Fetch training type to calculate energy cost
	trainingType, err := s.trainingTypeStore.GetByID(ctx, session.TrainingTypeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: ID=%d", ErrTrainingTypeNotFound, session.TrainingTypeID)
		}
		return fmt.Errorf("failed to fetch training type %d: %w", session.TrainingTypeID, err)
	}

	// Step 4: Calculate energy cost (energy_cost is per hour in training_types table)
	energyCost := float64(trainingType.EnergyCost) * session.DurationHours

	// Validate boxer has sufficient energy (defensive check - handler should have validated at creation)
	if boxer.Energy < energyCost {
		return fmt.Errorf("%w: boxer has %.1f energy but training costs %.1f",
			ErrInsufficientEnergy, boxer.Energy, energyCost)
	}

	// Step 5: Apply progression-aware stat gains (MAT-22)
	boxer.Energy -= energyCost // Deduct energy cost

	if s.progressionService != nil {
		// Calculate effective gains with diminishing returns and fatigue modifier
		effectiveGains := s.progressionService.CalculateEffectiveGains(
			boxer,
			session.PlannedStrengthGain,
			session.PlannedDefenseGain,
			session.PlannedAgilityGain,
			session.DurationHours,
		)

		// Apply effective stat gains
		boxer.Strength += effectiveGains.Strength
		boxer.Defense += effectiveGains.Defense
		boxer.Agility += effectiveGains.Agility

		// Add experience points
		boxer.Experience += effectiveGains.XP

		// Check for level up
		levelsGained := s.progressionService.ApplyLevelUp(boxer)
		if levelsGained > 0 {
			s.logger.Info("Boxer ID=%d gained %d level(s) after training session %d",
				boxer.ID, levelsGained, session.ID)
		}

		s.logger.Info("Training completed: session_id=%d boxer_id=%d energy_cost=%.1f str_gain=%.2f(.3f) def_gain=%.2f(/.3f) agi_gain=%.2f(/.3f) xp_gained=%.0f fatigue_mult=%.2f",
			session.ID, boxer.ID, energyCost,
			effectiveGains.Strength, session.PlannedStrengthGain,
			effectiveGains.Defense, session.PlannedDefenseGain,
			effectiveGains.Agility, session.PlannedAgilityGain,
			effectiveGains.XP, effectiveGains.FatigueMultiplier)
	} else {
		// Fallback to linear gains if progression service not available
		boxer.Strength += session.PlannedStrengthGain
		boxer.Defense += session.PlannedDefenseGain
		boxer.Agility += session.PlannedAgilityGain

		s.logger.Info("Training completed (no progression): session_id=%d boxer_id=%d energy_cost=%.1f str_gain=%.2f def_gain=%.2f agi_gain=%.2f",
			session.ID, boxer.ID, energyCost,
			session.PlannedStrengthGain, session.PlannedDefenseGain, session.PlannedAgilityGain)
	}

	// Ensure energy doesn't go negative (safety check)
	if boxer.Energy < 0 {
		boxer.Energy = 0
	}

	// Step 6: Update boxer in database
	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to update boxer %d after training: %w", boxer.ID, err)
	}

	// Step 7: Mark training session as completed
	if err := s.trainingSessionStore.MarkAsCompleted(ctx, sessionID); err != nil {
		// Note: boxer was already updated; this is a partial failure state
		// In production, we might want to rollback or track this for manual recovery
		s.logger.Error("Training session %d completion failed after boxer update: %v", sessionID, err)
		return fmt.Errorf("failed to mark training session %d as completed: %w", sessionID, err)
	}

	// Step 8: Apply fatigue increase from training (duration_hours × 15)
	if s.fatigueService != nil {
		fatigueIncrease := s.fatigueService.CalculateFatigueIncrease(session.DurationHours)
		if err := s.fatigueService.ApplyFatigueIncrease(ctx, boxer.ID, fatigueIncrease); err != nil {
			s.logger.Error("Failed to apply fatigue after training session %d: %v", sessionID, err)
			// Note: Training was completed but fatigue not tracked; log and continue
		}
	}

	// Step 9: Schedule automatic rest period based on fatigue (MAT-86)
	if s.fatigueService != nil && s.scheduledEventStore != nil {
		// Get updated boxer state after fatigue increase
		updatedBoxer, err := s.boxerStore.GetByID(ctx, boxer.ID)
		if err != nil {
			s.logger.Error("Failed to fetch boxer for rest scheduling: %v", err)
			// Don't fail training completion due to rest scheduling
		} else {
			// Calculate rest duration in hours: ceil(duration/2) with fatigue multiplier
			baseRestHours := int(math.Ceil(session.DurationHours / 2.0))
			fatigueMultiplier := math.Max(1.0, updatedBoxer.FatigueScore/60.0)
			finalRestHours := int(math.Ceil(float64(baseRestHours)*fatigueMultiplier))

			// Check if forced rest needed (exhaustion threshold = 80)
			needsForcedRest := updatedBoxer.FatigueScore >= ExhaustionThreshold

			// Calculate rest end time (in hours from now)
			restEndTime := time.Now().Add(time.Duration(finalRestHours) * time.Hour)

			// Create scheduled rest event data
			eventData, err := json.Marshal(map[string]interface{}{
				"rest_hours":  finalRestHours,
				"forced_rest": needsForcedRest,
				"training_id": session.ID,
			})
			if err != nil {
				s.logger.Error("Failed to marshal rest event data: %v", err)
			} else {
				restEvent := &model.ScheduledEvent{
					BoxerID:   boxer.ID,
					EventType: model.EventTypeRest,
					EventTime: restEndTime,
					EventData: eventData,
				}

				if err := s.scheduledEventStore.Create(ctx, restEvent); err != nil {
					s.logger.Error("Failed to schedule rest event after training session %d: %v", session.ID, err)
				} else {
					s.logger.Info("Rest period scheduled for boxer ID=%d: %d hours ending at %v (forced_rest=%v)",
						boxer.ID, finalRestHours, restEndTime, needsForcedRest)

					// If forced rest needed, update boxer.ForcedRestUntil
					if needsForcedRest {
						if err := s.fatigueService.ScheduleForcedRestHours(ctx, boxer.ID, finalRestHours); err != nil {
							s.logger.Error("Failed to schedule forced rest for boxer ID=%d: %v", boxer.ID, err)
						}
					}
				}
			}
		}
	}

	s.logger.Info("Training completed: session_id=%d boxer_id=%d energy_cost=%.1f strength_gain=%.2f defense_gain=%.2f agility_gain=%.2f",
		session.ID, boxer.ID, energyCost,
		session.PlannedStrengthGain, session.PlannedDefenseGain, session.PlannedAgilityGain)

	return nil
}

// CompleteAllDueTrainingSessions processes all pending training sessions that are due for completion.
// A session is "due" when its scheduled_completion_time <= current_game_time.
// This method is called by the world clock worker to batch-process training completions.
func (s *TrainingService) CompleteAllDueTrainingSessions(ctx context.Context, db *sql.DB) (int, int, error) {
	// For backward compatibility, this method accepts a db parameter but doesn't use it
	// The stores have their own database connections

	// Get all pending training sessions
	sessions, err := s.trainingSessionStore.GetAllPending(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch pending training sessions: %w", err)
	}

	if len(sessions) == 0 {
		return 0, 0, nil
	}

	s.logger.Info("Processing %d pending training sessions", len(sessions))

	var completedCount int
	var failedCount int

	for _, session := range sessions {
		if err := s.CompleteTrainingSession(ctx, session.ID); err != nil {
			s.logger.Error("Failed to complete training session ID=%d: %v", session.ID, err)
			failedCount++
		} else {
			completedCount++
		}
	}

	if completedCount > 0 {
		s.logger.Info("Completed %d training sessions, %d failed", completedCount, failedCount)
	}

	return completedCount, failedCount, nil
}

// CompleteTrainingForBoxer completes all pending training sessions for a specific boxer.
// This is useful for manual control panel (development/testing).
func (s *TrainingService) CompleteTrainingForBoxer(ctx context.Context, boxerID int) (int, int, error) {
	// Fetch pending training sessions for the specific boxer
	sessions, err := s.trainingSessionStore.GetPendingByBoxerID(ctx, boxerID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch pending training sessions for boxer %d: %w", boxerID, err)
	}

	if len(sessions) == 0 {
		return 0, 0, nil
	}

	s.logger.Info("Processing %d pending training sessions for boxer ID=%d", len(sessions), boxerID)

	var completedCount int
	var failedCount int

	for _, session := range sessions {
		if err := s.CompleteTrainingSession(ctx, session.ID); err != nil {
			s.logger.Error("Failed to complete training session ID=%d for boxer %d: %v", session.ID, boxerID, err)
			failedCount++
		} else {
			completedCount++
		}
	}

	if completedCount > 0 {
		s.logger.Info("Completed %d training sessions for boxer ID=%d, %d failed", completedCount, boxerID, failedCount)
	}

	return completedCount, failedCount, nil
}

// CreateTrainingSession creates a new training session with scheduled completion time calculation.
// This method is called by the handler when scheduling training for a boxer.
func (s *TrainingService) CreateTrainingSession(
	ctx context.Context,
	boxerID int,
	trainingTypeID int,
	durationHours float64,
	plannedStrengthGain float64,
	plannedDefenseGain float64,
	plannedAgilityGain float64,
) (*model.TrainingSession, error) {
	// Get current game time from world clock to calculate scheduled completion time
	gameTime, err := s.worldClockModel.GetCurrentGameTime(ctx, nil) // db is optional, will use session
	if err != nil {
		s.logger.Warn("Failed to get current game time, using real time: %v", err)
		gameTime = time.Now()
	}

	// Calculate scheduled completion time: game_time + duration_hours
	completionTime := gameTime.Add(time.Duration(durationHours*float64(time.Hour)))

	session := &model.TrainingSession{
		BoxerID:                   boxerID,
		TrainingTypeID:            trainingTypeID,
		DurationHours:             durationHours,
		PlannedStrengthGain:       plannedStrengthGain,
		PlannedDefenseGain:        plannedDefenseGain,
		PlannedAgilityGain:        plannedAgilityGain,
		ScheduledCompletionTime:   &completionTime,
		Status:                    model.TrainingSessionPending,
	}

	if err := s.trainingSessionStore.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create training session: %w", err)
	}

	s.logger.Info("Training session created: id=%d boxer_id=%d duration=%.1fh completion_time=%v",
		session.ID, boxerID, durationHours, completionTime)

	return session, nil
}
