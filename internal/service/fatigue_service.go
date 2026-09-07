package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
)

var (
	ErrExhausted        = errors.New("boxer is exhausted and cannot train")
	ErrForcedRestActive = errors.New("boxer is on mandatory forced rest")
	ErrFatigueOverflow  = errors.New("fatigue score would exceed maximum (100)")
)

const (
	// Fatigue increase per hour of training
	FatiguePerHour = 15.0

	// Natural fatigue decay per game day (no training)
	DailyFatigueDecay = 5.0

	// Exhaustion threshold - when exceeded, boxer cannot train
	ExhaustionThreshold = 80.0

	// Maximum fatigue score
	MaxFatigueScore = 100.0

	// Minimum fatigue score
	MinFatigueScore = 0.0

	// Default forced rest duration in days
	DefaultForcedRestDays = 3
)

// boxerRepository defines the interface for boxer data access
type boxerRepository interface {
	GetByID(ctx context.Context, id int) (*model.Boxer, error)
	Update(ctx context.Context, boxer *model.Boxer) error
}

// FatigueService manages boxer fatigue and recovery mechanics
type FatigueService struct {
	boxerStore boxerRepository
	logger     *logger.Logger
}

// NewFatigueService creates a new FatigueService instance
func NewFatigueService(
	boxerStore boxerRepository,
	lg *logger.Logger,
) *FatigueService {
	return &FatigueService{
		boxerStore: boxerStore,
		logger:     lg,
	}
}

// CalculateFatigueIncrease calculates the fatigue points gained from a training session
// Formula: duration_hours × 15
func (s *FatigueService) CalculateFatigueIncrease(durationHours float64) float64 {
	return durationHours * FatiguePerHour
}

// CheckCanTrain validates if a boxer can undergo training
// Returns (true, "") if training is allowed
// Returns (false, errorMessage) if training is blocked
func (s *FatigueService) CheckCanTrain(boxer *model.Boxer) (bool, string) {
	// Check exhaustion threshold
	if boxer.FatigueScore >= ExhaustionThreshold {
		return false, "Exhausted - needs rest"
	}

	// Check if forced rest is active
	if boxer.ForcedRestUntil != nil && time.Now().Before(*boxer.ForcedRestUntil) {
		return false, "Mandatory rest period active"
	}

	return true, ""
}

// ApplyFatigueIncrease adds fatigue points to a boxer after training
// Updates the boxer record and validates bounds
func (s *FatigueService) ApplyFatigueIncrease(ctx context.Context, boxerID int, increase float64) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	newFatigue := boxer.FatigueScore + increase

	// Clamp to maximum
	if newFatigue > MaxFatigueScore {
		newFatigue = MaxFatigueScore
		s.logger.Warn("Fatigue clamped to max for boxer ID=%d: was %.2f+%.2f", boxerID, boxer.FatigueScore, increase)
	}

	boxer.FatigueScore = newFatigue

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to update fatigue for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Fatigue increased for boxer ID=%d: %.2f → %.2f (+%.2f)", boxerID, boxer.FatigueScore-increase, boxer.FatigueScore, increase)

	// Check if boxer reached exhaustion threshold
	if newFatigue >= ExhaustionThreshold {
		s.logger.Warn("Boxer ID=%d reached exhaustion threshold (%.2f >= %.2f)", boxerID, newFatigue, ExhaustionThreshold)
	}

	return nil
}

// ReduceFatigue reduces fatigue points after recovery/rest
func (s *FatigueService) ReduceFatigue(ctx context.Context, boxerID int, reduction float64) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	oldFatigue := boxer.FatigueScore
	newFatigue := oldFatigue - reduction

	// Clamp to minimum
	if newFatigue < MinFatigueScore {
		newFatigue = MinFatigueScore
	}

	boxer.FatigueScore = newFatigue

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to update fatigue for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Fatigue reduced for boxer ID=%d: %.2f → %.2f (-%.2f)", boxerID, oldFatigue, newFatigue, reduction)
	return nil
}

// ApplyFatigueDecay applies daily natural fatigue decay to all boxers
// This is called by the world clock worker once per game day
func (s *FatigueService) ApplyFatigueDecay(ctx context.Context, boxerID int) error {
	return s.ReduceFatigue(ctx, boxerID, DailyFatigueDecay)
}

// ScheduleForcedRest sets a forced rest period for an exhausted boxer
func (s *FatigueService) ScheduleForcedRest(ctx context.Context, boxerID int, durationDays int) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	// Calculate forced rest end time
	forcedRestUntil := time.Now().AddDate(0, 0, durationDays)
	boxer.ForcedRestUntil = &forcedRestUntil

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to schedule forced rest for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Forced rest scheduled for boxer ID=%d: until %s (%d days)", boxerID, forcedRestUntil.Format(time.RFC3339), durationDays)
	return nil
}

// ClearForcedRest removes the forced rest restriction from a boxer
func (s *FatigueService) ClearForcedRest(ctx context.Context, boxerID int) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	if boxer.ForcedRestUntil == nil {
		s.logger.Debug("No forced rest active for boxer ID=%d", boxerID)
		return nil
	}

	boxer.ForcedRestUntil = nil

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to clear forced rest for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Forced rest cleared for boxer ID=%d", boxerID)
	return nil
}

// CheckExhaustionThreshold returns true if boxer is at or above exhaustion threshold
func (s *FatigueService) CheckExhaustionThreshold(boxer *model.Boxer) bool {
	return boxer.FatigueScore >= ExhaustionThreshold
}

// IsOnForcedRest returns true if boxer currently has an active forced rest period
func (s *FatigueService) IsOnForcedRest(boxer *model.Boxer) bool {
	if boxer.ForcedRestUntil == nil {
		return false
	}
	return time.Now().Before(*boxer.ForcedRestUntil)
}

// GetRecoveryBenefits returns the energy recovery, fatigue reduction, and stat decay risk for a given rest duration
type RecoveryBenefits struct {
	EnergyRecoveryPercent float64 `json:"energy_recovery_percent"` // Percentage of max energy restored (50-100)
	FatigueReduction      float64 `json:"fatigue_reduction"`       // Fatigue points reduced
	StatDecayRisk         float64 `json:"stat_decay_risk"`         // Percentage stat decay if any (0 for short rests)
}

func (s *FatigueService) GetRecoveryBenefits(restDays int) RecoveryBenefits {
	switch {
	case restDays == 1:
		return RecoveryBenefits{
			EnergyRecoveryPercent: 50.0,
			FatigueReduction:      20.0,
			StatDecayRisk:         0.0,
		}
	case restDays == 2:
		return RecoveryBenefits{
			EnergyRecoveryPercent: 80.0,
			FatigueReduction:      40.0,
			StatDecayRisk:         0.0,
		}
	case restDays >= 3 && restDays < 7:
		return RecoveryBenefits{
			EnergyRecoveryPercent: 100.0,
			FatigueReduction:      60.0,
			StatDecayRisk:         0.0,
		}
	case restDays >= 7:
		return RecoveryBenefits{
			EnergyRecoveryPercent: 100.0,
			FatigueReduction:      100.0,
			StatDecayRisk:         5.0, // Long rest causes minor stat decay
		}
	default:
		return RecoveryBenefits{
			EnergyRecoveryPercent: 0.0,
			FatigueReduction:      0.0,
			StatDecayRisk:         0.0,
		}
	}
}

// ApplyRecovery applies recovery benefits based on rest duration to a boxer
func (s *FatigueService) ApplyRecovery(ctx context.Context, boxerID int, restDays int) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	benefits := s.GetRecoveryBenefits(restDays)

	// Store old values for logging
	oldEnergy := boxer.Energy
	oldFatigue := boxer.FatigueScore
	oldStrength := boxer.Strength
	oldDefense := boxer.Defense
	oldAgility := boxer.Agility

	// Apply energy recovery (cap at 100)
	boxer.Energy = math.Min(boxer.Energy+benefits.EnergyRecoveryPercent, 100.0)

	// Apply fatigue reduction
	if err := s.ReduceFatigue(ctx, boxerID, benefits.FatigueReduction); err != nil {
		return err
	}

	// Refresh boxer after fatigue update
	boxer, err = s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		return fmt.Errorf("failed to refresh boxer %d: %w", boxerID, err)
	}

	// Apply stat decay for long rests (7+ days)
	if benefits.StatDecayRisk > 0 {
		decayFactor := 1.0 - (benefits.StatDecayRisk / 100.0)
		boxer.Strength *= decayFactor
		boxer.Defense *= decayFactor
		boxer.Agility *= decayFactor

		s.logger.Info("Stat decay applied for boxer ID=%d after %d day rest: strength %.2f→%.2f, defense %.2f→%.2f, agility %.2f→%.2f",
			boxerID, restDays, oldStrength, boxer.Strength, oldDefense, boxer.Defense, oldAgility, boxer.Agility)
	}

	// Update boxer (energy and stats)
	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to update boxer %d after recovery: %w", boxerID, err)
	}

	s.logger.Info("Recovery applied for boxer ID=%d (%d days): energy %.1f→%.1f, fatigue %.2f (old) with -%.1f reduction, stat_decay=%.1f%%",
		boxerID, restDays, oldEnergy, boxer.Energy, oldFatigue, benefits.FatigueReduction, benefits.StatDecayRisk)

	// Check if forced rest period should be cleared
	if s.IsOnForcedRest(boxer) && boxer.FatigueScore < ExhaustionThreshold {
		if err := s.ClearForcedRest(ctx, boxerID); err != nil {
			s.logger.Warn("Failed to clear forced rest for boxer ID=%d: %v", boxerID, err)
		}
	}

	return nil
}
