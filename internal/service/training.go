package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	boxerStore           *store.BoxerStore
	trainingTypeStore    *store.TrainingTypeStore
	trainingSessionStore *store.TrainingSessionStore
	logger               *logger.Logger
}

// NewTrainingService creates a new TrainingService instance
func NewTrainingService(
	boxerStore *store.BoxerStore,
	trainingTypeStore *store.TrainingTypeStore,
	trainingSessionStore *store.TrainingSessionStore,
	lg *logger.Logger,
) *TrainingService {
	return &TrainingService{
		boxerStore:           boxerStore,
		trainingTypeStore:    trainingTypeStore,
		trainingSessionStore: trainingSessionStore,
		logger:               lg,
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

	// Step 5: Apply changes to boxer
	boxer.Energy -= energyCost                    // Deduct energy cost
	boxer.Strength += session.PlannedStrengthGain // Apply planned gains
	boxer.Defense += session.PlannedDefenseGain
	boxer.Agility += session.PlannedAgilityGain

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

	s.logger.Info("Training completed: session_id=%d boxer_id=%d energy_cost=%.1f strength_gain=%.2f defense_gain=%.2f agility_gain=%.2f",
		session.ID, boxer.ID, energyCost,
		session.PlannedStrengthGain, session.PlannedDefenseGain, session.PlannedAgilityGain)

	return nil
}

// CompleteAllDueTrainingSessions processes all pending training sessions.
// This method is called by the world clock worker to batch-process training completions.
// Currently completes all pending sessions immediately. Future iterations can add
// time-based filtering via scheduled_events integration.
func (s *TrainingService) CompleteAllDueTrainingSessions(ctx context.Context) (int, int, error) {
	// Fetch all pending training sessions
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
