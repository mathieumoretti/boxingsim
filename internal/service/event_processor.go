package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

var (
	ErrEventNotProcessed = errors.New("event was not processed")
	ErrBoxerNotFound     = errors.New("boxer not found")
	ErrInvalidEventData  = errors.New("invalid event data")
	ErrUnknownEventType  = errors.New("unknown event type")
)

// EventProcessor handles the processing of scheduled events for boxers.
// It follows the same pattern as FightService with proper error wrapping and contextual logging.
type EventProcessor struct {
	eventStore     *store.ScheduledEventStore
	boxerStore     *store.BoxerStore
	fatigueService *FatigueService
	logger         logger.Logger
}

// NewEventProcessor creates a new EventProcessor instance.
func NewEventProcessor(
	eventStore *store.ScheduledEventStore,
	boxerStore *store.BoxerStore,
	fatigueService *FatigueService,
	lg logger.Logger,
) *EventProcessor {
	return &EventProcessor{
		eventStore:     eventStore,
		boxerStore:     boxerStore,
		fatigueService: fatigueService,
		logger:         lg,
	}
}

// ProcessScheduledEvent processes a single scheduled event based on its type.
// It applies the appropriate handler and marks the event as processed upon success.
func (p *EventProcessor) ProcessScheduledEvent(ctx context.Context, event *model.ScheduledEvent) error {
	p.logger.Info("Processing event ID=%d type=%s boxer_id=%d", event.ID, event.EventType, event.BoxerID)

	var processErr error
	switch event.EventType {
	case model.EventTypeTraining:
		processErr = p.processTrainingComplete(ctx, event)
	case model.EventTypeRest:
		processErr = p.processRecovery(ctx, event)
	case model.EventTypeCompetition:
		processErr = p.processCompetition(ctx, event)
	default:
		processErr = fmt.Errorf("%w: %q", ErrUnknownEventType, event.EventType)
	}

	if processErr != nil {
		p.logger.Error("Event processing failed ID=%d error=%v", event.ID, processErr)
		return processErr
	}

	// Mark event as processed only after successful handling (idempotent)
	if err := p.eventStore.MarkAsProcessed(ctx, event.ID); err != nil {
		if errors.Is(err, store.ErrAlreadyProcessed) {
			p.logger.Info("Event already marked as processed ID=%d", event.ID)
			return nil
		}
		p.logger.Error("Failed to mark event as processed ID=%d error=%v", event.ID, err)
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	p.logger.Info("Event processed successfully ID=%d", event.ID)
	return nil
}

// processTrainingComplete handles training completion events.
// It applies stat gains based on the training data (strength_gain, defense_gain, agility_gain).
func (p *EventProcessor) processTrainingComplete(ctx context.Context, event *model.ScheduledEvent) error {
	// Unmarshal training data
	var data map[string]any
	if len(event.EventData) > 0 {
		if err := json.Unmarshal(event.EventData, &data); err != nil {
			return fmt.Errorf("%w: failed to unmarshal training data: %w", ErrInvalidEventData, err)
		}
	}

	// Extract stat gains from data (with defaults)
	strengthGain := getFloatField(data, "strength_gain", 0.0)
	defenseGain := getFloatField(data, "defense_gain", 0.0)
	agilityGain := getFloatField(data, "agility_gain", 0.0)

	// Fetch boxer
	boxer, err := p.boxerStore.GetByID(ctx, event.BoxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: ID=%d", ErrBoxerNotFound, event.BoxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", event.BoxerID, err)
	}

	// Apply stat gains
	boxer.Strength += strengthGain
	boxer.Defense += defenseGain
	boxer.Agility += agilityGain

	// Update boxer in database
	if err := p.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to update boxer %d after training: %w", boxer.ID, err)
	}

	p.logger.Info("Training applied to boxer ID=%d strength_gain=%.2f defense_gain=%.2f agility_gain=%.2f",
		boxer.ID, strengthGain, defenseGain, agilityGain)
	return nil
}

// processRecovery handles rest/recovery events.
// It delegates to FatigueService.ApplyRecovery for comprehensive recovery handling.
func (p *EventProcessor) processRecovery(ctx context.Context, event *model.ScheduledEvent) error {
	// Unmarshal recovery data
	var data map[string]any
	if len(event.EventData) > 0 {
		if err := json.Unmarshal(event.EventData, &data); err != nil {
			return fmt.Errorf("%w: failed to unmarshal recovery data: %w", ErrInvalidEventData, err)
		}
	}

	// Extract rest duration from event data (default to 1 hour if not specified)
	restHours := getIntField(data, "rest_hours", 1)

	// Delegate to FatigueService for comprehensive recovery handling
	if p.fatigueService == nil {
		return fmt.Errorf("fatigueService is nil, cannot apply recovery")
	}

	if err := p.fatigueService.ApplyRecovery(ctx, event.BoxerID, restHours); err != nil {
		return fmt.Errorf("failed to apply recovery: %w", err)
	}

	return nil
}

// processCompetition handles competition events.
// Currently a placeholder for future fight/competition processing.
func (p *EventProcessor) processCompetition(ctx context.Context, event *model.ScheduledEvent) error {
	_ = ctx // Unused for now, will be used when competition logic is implemented
	// TODO: Implement competition processing logic
	p.logger.Info("Competition event received for boxer ID=%d (not yet implemented)", event.BoxerID)
	return nil
}

// getFloatField safely extracts a float64 value from a map with a default fallback.
func getFloatField(data map[string]any, key string, defaultValue float64) float64 {
	if data == nil {
		return defaultValue
	}
	val, ok := data[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultValue
	}
}

// getIntField safely extracts an int value from a map with a default fallback.
func getIntField(data map[string]any, key string, defaultValue int) int {
	if data == nil {
		return defaultValue
	}
	val, ok := data[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return defaultValue
	}
}
