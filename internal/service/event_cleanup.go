package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

// CleanupResult represents the result of processing a single event.
type CleanupResult int

const (
	CleanupSkipped CleanupResult = iota
	CleanupFightCancelled
	CleanupEventCompleted
)

// CleanupSummary holds the results of a cleanup run.
type CleanupSummary struct {
	FightsCancelled       int
	EventsMarkedCompleted int
	EventsSkipped         int
	TotalProcessed        int
}

// EventCleanupService handles automatic cleanup of old failed events (MAT-143).
type EventCleanupService struct {
	eventStore *store.ScheduledEventStore
	fightStore *store.FightStore
	config     config.CleanupConfig
	logger     *logger.Logger
}

// NewEventCleanupService creates a new EventCleanupService instance.
func NewEventCleanupService(
	eventStore *store.ScheduledEventStore,
	fightStore *store.FightStore,
	cfg config.CleanupConfig,
	lg *logger.Logger,
) *EventCleanupService {
	return &EventCleanupService{
		eventStore: eventStore,
		fightStore: fightStore,
		config:     cfg,
		logger:     lg,
	}
}

// Run executes the cleanup job for old failed events.
// Returns a summary of actions taken.
func (s *EventCleanupService) Run(ctx context.Context, db *sql.DB) (CleanupSummary, error) {
	startTime := time.Now()
	summary := CleanupSummary{}

	s.logger.Info("Cleanup job started", "threshold_hours", s.config.AbandonAgeHours)

	// Query for failed events older than threshold
	oldFailedEvents, err := s.eventStore.GetFailedEventsOlderThan(ctx, s.config.AbandonAgeHours)
	if err != nil {
		return summary, fmt.Errorf("failed to query old failed events: %w", err)
	}

	summary.TotalProcessed = len(oldFailedEvents)

	for _, event := range oldFailedEvents {
		result, err := s.processOldFailedEvent(ctx, event, db)
		if err != nil {
			s.logger.Error("Failed to process old failed event", "event_id", event.EventID, "error", err.Error())
			summary.EventsSkipped++
			continue
		}

		// Update summary based on result
		switch result {
		case CleanupFightCancelled:
			summary.FightsCancelled++
		case CleanupEventCompleted:
			summary.EventsMarkedCompleted++
		case CleanupSkipped:
			summary.EventsSkipped++
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("Cleanup job completed",
		"processed", summary.TotalProcessed,
		"fights_cancelled", summary.FightsCancelled,
		"events_completed", summary.EventsMarkedCompleted,
		"events_skipped", summary.EventsSkipped,
		"duration_ms", duration.Milliseconds())

	return summary, nil
}

// processOldFailedEvent handles cleanup for a single old failed event.
func (s *EventCleanupService) processOldFailedEvent(ctx context.Context, event *store.OldFailedEvent, db *sql.DB) (CleanupResult, error) {
	// Log event age for observability
	eventMsg := "unknown error"
	if event.ErrorMessage != nil {
		eventMsg = *event.ErrorMessage
	}

	s.logger.Debug("Processing old failed event",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"boxer_id", event.BoxerID,
		"hours_old", fmt.Sprintf("%.1f", event.HoursOld),
		"error", eventMsg)

	switch event.EventType {
	case model.EventTypeFightSimulate:
		return s.cleanupFightEvent(ctx, event, db)
	case model.EventTypeTraining, model.EventTypeRest:
		return s.cleanupTrainingOrRestEvent(ctx, event, db)
	case model.EventTypeCompetition:
		// Competition events can be marked as completed (no-op)
		return s.cleanupTrainingOrRestEvent(ctx, event, db)
	default:
		// Unknown event types - mark as completed to prevent blocking
		s.logger.Debug("Unknown event type, marking as completed", "event_id", event.EventID, "event_type", event.EventType)
		return s.markEventAsCompleted(ctx, event.EventID, fmt.Sprintf("Auto-cleaned: unknown event type after %.0f hours", event.HoursOld), db)
	}
}

// cleanupFightEvent handles failed fight_simulate events.
// Cancels the associated fight if it's still in 'scheduled' status.
func (s *EventCleanupService) cleanupFightEvent(ctx context.Context, event *store.OldFailedEvent, db *sql.DB) (CleanupResult, error) {
	// Safety check: ensure fight exists and is still scheduled
	if event.FightID == nil {
		s.logger.Warn("Fight event has no associated fight_id", "event_id", event.EventID)
		return CleanupSkipped, nil
	}

	fightStatus := "unknown"
	if event.FightStatus != nil {
		fightStatus = *event.FightStatus
	}

	// Only cancel if fight is still scheduled
	if fightStatus != "scheduled" {
		s.logger.Debug("Fight already not in scheduled state, skipping",
			"event_id", event.EventID,
			"fight_id", *event.FightID,
			"fight_status", fightStatus)
		return CleanupSkipped, nil
	}

	// Cancel the fight with auto-cleanup reason
	reason := fmt.Sprintf("Auto-cleaned: fight_simulate event failed after %.0f hours", event.HoursOld)
	if err := s.fightStore.CancelZombieFight(ctx, *event.FightID, reason); err != nil {
		if err == store.ErrFightNotFound {
			s.logger.Debug("Fight not found during cancellation (already resolved)",
				"event_id", event.EventID,
				"fight_id", *event.FightID)
			return CleanupSkipped, nil
		}
		return CleanupSkipped, fmt.Errorf("failed to cancel fight %d: %w", *event.FightID, err)
	}

	s.logger.Info("Cancelled fight via auto-cleanup",
		"event_id", event.EventID,
		"fight_id", *event.FightID,
		"boxer1_id", event.Boxer1ID,
		"boxer2_id", event.Boxer2ID,
		"reason", reason)

	// Mark the event as completed to prevent re-processing
	if result, err := s.markEventAsCompleted(ctx, event.EventID, reason, db); err != nil {
		s.logger.Warn("Failed to mark event as completed after fight cancellation",
			"event_id", event.EventID,
			"error", err.Error())
	} else if result == CleanupSkipped {
		// Event was already resolved
		return CleanupSkipped, nil
	}

	return CleanupFightCancelled, nil
}

// cleanupTrainingOrRestEvent handles failed training or rest events.
// Marks them as completed since boxer state has likely already changed.
func (s *EventCleanupService) cleanupTrainingOrRestEvent(ctx context.Context, event *store.OldFailedEvent, db *sql.DB) (CleanupResult, error) {
	reason := fmt.Sprintf("Auto-cleaned: %s event failed after %.0f hours", event.EventType, event.HoursOld)

	result, err := s.markEventAsCompleted(ctx, event.EventID, reason, db)
	if err != nil {
		return CleanupSkipped, fmt.Errorf("failed to mark event as completed: %w", err)
	}

	s.logger.Debug("Marked training/rest event as completed via auto-cleanup",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"boxer_id", event.BoxerID,
		"reason", reason)

	return result, nil
}

// markEventAsCompleted marks an event as completed with a reason.
func (s *EventCleanupService) markEventAsCompleted(ctx context.Context, eventId int, reason string, db *sql.DB) (CleanupResult, error) {
	// Use a custom update that sets status to completed and adds reason to error_message
	query := `
		UPDATE scheduled_events
		SET processed = TRUE, status = 'completed', processed_at = CURRENT_TIMESTAMP,
			error_message = COALESCE(error_message, '') || ' | Auto-cleaned: ' || $2
		WHERE id = $1 AND status = 'failed'`

	result, err := db.ExecContext(ctx, query, eventId, reason)
	if err != nil {
		return CleanupSkipped, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		s.logger.Debug("Event already resolved (no longer failed)", "event_id", eventId)
		return CleanupSkipped, nil
	}

	return CleanupEventCompleted, nil
}
