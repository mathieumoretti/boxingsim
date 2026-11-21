package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/mormm/boxing/internal/domain"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

var (
	ErrInsufficientHealthForFight = errors.New("boxer health too low to fight")
	ErrForcedRecoveryActive       = errors.New("boxer is on mandatory forced recovery")
)

const (
	// Minimum health required to enter a fight
	MinHealthToFight = 50.0

	// Health damage thresholds for recovery calculation
	MajorDamageThreshold    = 40.0 // >40 HP lost = major injury
	ModerateDamageThreshold = 20.0 // >20 HP lost = moderate injury

	// Recovery duration multipliers (in hours)
	MajorInjuryRecoveryHours    = 72 // 3 days for major injuries
	ModerateInjuryRecoveryHours = 48 // 2 days for moderate injuries
	MinorInjuryRecoveryHours    = 24 // 1 day for minor damage

	// Max health (for restoration calculations)
	MaxHealth = 100.0
)

// RecoveryService manages boxer recovery from fight injuries
type RecoveryService struct {
	boxerStore      *store.BoxerStore
	consumableStore *store.ConsumableStore
	logger          *logger.Logger
	db              *sql.DB
}

// NewRecoveryService creates a new RecoveryService instance
func NewRecoveryService(
	boxerStore *store.BoxerStore,
	consumableStore *store.ConsumableStore,
	lg *logger.Logger,
	db *sql.DB,
) *RecoveryService {
	return &RecoveryService{
		boxerStore:      boxerStore,
		consumableStore: consumableStore,
		logger:          lg,
		db:              db,
	}
}

// CalculateRecoveryDuration determines required recovery time based on health damage
func (s *RecoveryService) CalculateRecoveryDuration(healthLost float64) int {
	switch {
	case healthLost > MajorDamageThreshold:
		return MajorInjuryRecoveryHours
	case healthLost > ModerateDamageThreshold:
		return ModerateInjuryRecoveryHours
	default:
		return MinorInjuryRecoveryHours
	}
}

// ScheduleForcedRecovery sets a mandatory recovery period after fight damage
func (s *RecoveryService) ScheduleForcedRecovery(ctx context.Context, boxerID domain.BoxerID, healthLost float64, gameTime time.Time) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	// Calculate recovery duration based on damage taken
	recoveryHours := s.CalculateRecoveryDuration(healthLost)

	// Set forced recovery until (using game time for consistency with events)
	forcedRecoveryUntil := gameTime.Add(time.Duration(recoveryHours) * time.Hour)
	boxer.ForcedRecoveryUntil = &forcedRecoveryUntil

	// Also set next available at for quick checks
	boxer.NextAvailableAt = &forcedRecoveryUntil

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to schedule forced recovery for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Forced recovery scheduled for boxer ID=%d: health lost=%.1f, recovery hours=%d, until=%s",
		boxerID, healthLost, recoveryHours, forcedRecoveryUntil.Format(time.RFC3339))

	return nil
}

// ClearForcedRecovery removes the forced recovery restriction
func (s *RecoveryService) ClearForcedRecovery(ctx context.Context, boxerID domain.BoxerID) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	if boxer.ForcedRecoveryUntil == nil {
		s.logger.Debug("No forced recovery active for boxer ID=%d", boxerID)
		return nil
	}

	boxer.ForcedRecoveryUntil = nil
	boxer.NextAvailableAt = nil

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to clear forced recovery for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Forced recovery cleared for boxer ID=%d", boxerID)
	return nil
}

// IsOnForcedRecovery returns true if boxer is currently in forced recovery
func (s *RecoveryService) IsOnForcedRecovery(boxer *model.Boxer, gameTime time.Time) bool {
	if boxer.ForcedRecoveryUntil == nil {
		return false
	}
	return gameTime.Before(*boxer.ForcedRecoveryUntil)
}

// CheckCanFight validates if boxer can enter a fight based on health and recovery status
func (s *RecoveryService) CheckCanFight(boxer *model.Boxer, gameTime time.Time) (bool, string) {
	// Check minimum health threshold
	if boxer.Health < MinHealthToFight {
		return false, fmt.Sprintf("Health too low (%.0f%%). Minimum required: %.0f%%", boxer.Health, MinHealthToFight)
	}

	// Check forced recovery status
	if s.IsOnForcedRecovery(boxer, gameTime) {
		return false, "Boxer is on mandatory forced recovery from fight injuries"
	}

	return true, ""
}

// ApplyHealthRestoration restores health during a recovery rest event (MAT-154)
func (s *RecoveryService) ApplyHealthRestoration(ctx context.Context, boxerID domain.BoxerID, restHours int) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	oldHealth := boxer.Health

	// Calculate health restoration based on rest duration (constants from package-level in fatigue_service.go)
	healthRestore := math.Min(float64(restHours)*HealthRestorePerRecoveryHour, MaxHealthRestoreInOneRest)

	// Apply restoration (cap at max health)
	newHealth := math.Min(oldHealth+healthRestore, MaxHealth)

	boxer.Health = newHealth

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to restore health for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Health restored for boxer ID=%d: %.1f→%.1f (+%.1f)", boxerID, oldHealth, newHealth, healthRestore)

	// Clear forced recovery if completed and health is sufficient
	if s.IsOnForcedRecovery(boxer, time.Now()) && newHealth >= MinHealthToFight {
		if err := s.ClearForcedRecovery(ctx, boxerID); err != nil {
			s.logger.Warn("Failed to clear forced recovery for boxer ID=%d: %v", boxerID, err)
		}
	}

	return nil
}

// ApplyConsumable applies a consumable's effects to a boxer (MAT-153 Phase 6)
func (s *RecoveryService) ApplyConsumable(ctx context.Context, boxerID domain.BoxerID, consumable *model.Consumable) error {
	boxer, err := s.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("boxer not found: ID=%d", boxerID)
		}
		return fmt.Errorf("failed to fetch boxer %d: %w", boxerID, err)
	}

	oldHealth := boxer.Health
	oldEnergy := boxer.Energy
	oldFatigue := boxer.FatigueScore

	// Apply health restoration (cap at MaxHealth)
	if consumable.HealthRestore > 0 {
		boxer.Health = math.Min(boxer.Health+consumable.HealthRestore, MaxHealth)
	}

	// Apply energy restoration (cap at 100.0)
	if consumable.EnergyRestore > 0 {
		boxer.Energy = math.Min(boxer.Energy+consumable.EnergyRestore, 100.0)
	}

	// Apply fatigue reduction (constants from package-level in fatigue_service.go)
	if consumable.FatigueReduction > 0 {
		boxer.FatigueScore = math.Max(boxer.FatigueScore-consumable.FatigueReduction, MinFatigueScore)
	}

	if err := s.boxerStore.Update(ctx, boxer); err != nil {
		return fmt.Errorf("failed to apply consumable effects for boxer %d: %w", boxerID, err)
	}

	s.logger.Info("Consumable '%s' applied for boxer ID=%d: health %.1f→%.1f, energy %.1f→%.1f, fatigue %.2f→%.2f",
		consumable.Name, boxerID, oldHealth, boxer.Health, oldEnergy, boxer.Energy, oldFatigue, boxer.FatigueScore)

	return nil
}
