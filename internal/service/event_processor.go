package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"

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
// It restores energy and health based on recovery rates in the event data.
func (p *EventProcessor) processRecovery(ctx context.Context, event *model.ScheduledEvent) error {
	// Unmarshal recovery data
	var data map[string]any
	if len(event.EventData) > 0 {
		if err := json.Unmarshal(event.EventData, &data); err != nil {
			return fmt.Errorf("%w: failed to unmarshal recovery data: %w", ErrInvalidEventData, err)
		}
	}

	// Extract rest duration from event data (default to 1 day if not specified)
	restDays := getIntField(data, "rest_days", 1)

	// Fetch boxer
	boxer, err := p.boxerStore.GetByID(ctx, event.BoxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: ID=%d", ErrBoxerNotFound, event.BoxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", event.BoxerID, err)
	}

	// Apply fatigue decay for each day of rest (5 points per day)
	if p.fatigueService != nil {
		fatigueReduction := float64(restDays) * FatigueDailyDecay
		oldFatigue := boxer.FatigueScore
		newFatigue := math.Max(0, oldFatigue-fatigueReduction)
		boxer.FatigueScore = newFatigue

		p.logger.Info("Fatigue decay applied to boxer ID=%d: %.2f → %.2f (%d days × 5.0)",
			boxer.ID, oldFatigue, newFatigue, restDays)
	}

	// Apply energy recovery based on rest duration (capped at 100)
	energyGain := getFloatField(data, "energy_gain", 100.0) // Full energy by default
	boxer.Energy = math.Min(boxer.Energy+energyGain, 100.0)

	// Apply health recovery based on rest duration (capped at 100)
	healthGain := getFloatField(data, "health_gain", 100.0) // Full health by default
	boxer.Health = math.Min(boxer.Health+healthGain, 100.0)

	// Update boxer in database
	if err := p.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to update boxer %d after recovery: %w", boxer.ID, err)
	}

	p.logger.Info("Recovery applied to boxer ID=%d energy_gain=%.2f health_gain=%.2f new_energy=%.2f new_health=%.2f new_fatigue=%.2f",
		boxer.ID, energyGain, healthGain, boxer.Energy, boxer.Health, boxer.FatigueScore)
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

// FatigueDailyDecay is the amount of fatigue reduced per day of rest (5.0 points/day).
const FatigueDailyDecay = 5.0
