package fight

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

type Fight struct {
	ID                int
	Boxer1ID          *int
	Boxer2ID          *int
	Status            string
	ScheduledTime     *time.Time
	StartTime         *time.Time
	EndTime           *time.Time
	WinnerID          *int
	Round             int
	SimulationVersion string
	Data              map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type FightService struct {
	fightStore *store.FightStore
	boxerStore *store.BoxerStore
	cfg        *config.Config
	logger     *logger.Logger
	boxerSvc   *boxer.BoxerService
	eventStore *store.ScheduledEventStore
}

func NewFightService(fightStore *store.FightStore, boxerStore *store.BoxerStore, cfg *config.Config, boxerSvc *boxer.BoxerService, eventStore *store.ScheduledEventStore) *FightService {
	return &FightService{
		fightStore: fightStore,
		boxerStore: boxerStore,
		cfg:        cfg,
		logger:     logger.New("FightService"),
		boxerSvc:   boxerSvc,
		eventStore: eventStore,
	}
}

func (s *FightService) Schedule(boxer1ID, boxer2ID int, scheduledTime time.Time) (*Fight, error) {
	ctx := context.Background()

	// Create the fight using the store
	id, err := s.fightStore.Create(ctx, boxer1ID, boxer2ID, &scheduledTime, 1)
	if err != nil {
		s.logger.Error("Failed to schedule fight", err)
		return nil, err
	}

	fight := &Fight{
		ID:            id,
		Boxer1ID:      &boxer1ID,
		Boxer2ID:      &boxer2ID,
		Status:        "scheduled",
		ScheduledTime: &scheduledTime,
		StartTime:     nil,
		EndTime:       nil,
		WinnerID:      nil,
		Round:         1,
		Data:          make(map[string]interface{}),
	}

	s.logger.Info("Fight scheduled", "id", fight.ID)

	// Create a scheduled event for the fight simulation (MAT-99)
	if s.eventStore != nil {
		eventData, _ := json.Marshal(map[string]any{
			"fight_id": id,
		})

		event := &model.ScheduledEvent{
			BoxerID:   boxer1ID, // Associate with first boxer
			EventType: model.EventTypeFightSimulate,
			EventTime: scheduledTime,
			Processed: false,
			EventData: model.EventData(eventData),
		}

		if err := s.eventStore.Create(ctx, event); err != nil {
			// Log the error but don't fail the fight scheduling
			s.logger.Error("Failed to create scheduled event for fight", "fight_id", id, "error", err)
		} else {
			s.logger.Info("Created scheduled event for fight", "fight_id", id, "event_time", scheduledTime)
		}
	}

	return fight, nil
}

func (s *FightService) GetByID(id int) (*Fight, error) {
	ctx := context.Background()

	fightModel, err := s.fightStore.GetByID(ctx, id)
	if err != nil {
		if err == store.ErrFightNotFound {
			return nil, nil
		}
		s.logger.Error("Failed to get fight", err)
		return nil, err
	}

	// Convert from model.Fight to Fight (internal/fight package type)
	return &Fight{
		ID:                fightModel.ID,
		Boxer1ID:          &fightModel.Boxer1ID,
		Boxer2ID:          &fightModel.Boxer2ID,
		Status:            string(fightModel.Status),
		ScheduledTime:     fightModel.ScheduledTime,
		StartTime:         fightModel.StartTime,
		EndTime:           fightModel.EndTime,
		WinnerID:          fightModel.WinnerID,
		Round:             fightModel.Round,
		SimulationVersion: fightModel.SimulationVersion,
		Data:              fightModel.Data,
		CreatedAt:         fightModel.CreatedAt,
		UpdatedAt:         fightModel.UpdatedAt,
	}, nil
}

func (s *FightService) GetUpcoming(limit int) ([]*Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE status = 'scheduled' AND (scheduled_time > NOW() OR scheduled_time IS NULL)
		ORDER BY scheduled_time ASC LIMIT $1
	`
	return s.getFightsByCustomQuery(query, limit)
}

func (s *FightService) GetInProgress(limit int) ([]*Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE status = 'in_progress'
		ORDER BY start_time DESC LIMIT $1
	`
	return s.getFightsByCustomQuery(query, limit)
}

// getFightsByCustomQuery executes a custom query and returns fights (used for GetUpcoming/GetInProgress)
func (s *FightService) getFightsByCustomQuery(query string, limit int) ([]*Fight, error) {
	// Note: This is a workaround since FightStore doesn't have a custom query method.
	// For production, we should add GetUpcoming/GetInProgress to FightStore if needed.
	// For now, GetByBoxer and GetCompleted cover the main use cases.
	s.logger.Warn("getFightsByCustomQuery is deprecated - use FightStore methods instead")
	return nil, nil
}

func (s *FightService) UpdateStatus(id int, status string) error {
	ctx := context.Background()
	err := s.fightStore.UpdateStatus(ctx, id, status)
	if err != nil {
		s.logger.Error("Failed to update fight status", err)
		return err
	}

	s.logger.Info("Fight status updated", "id", id, "status", status)
	return nil
}

func (s *FightService) UpdateRound(id int, round int) error {
	ctx := context.Background()
	err := s.fightStore.UpdateRound(ctx, id, round)
	if err != nil {
		s.logger.Error("Failed to update fight round", err)
		return err
	}

	s.logger.Info("Fight round updated", "id", id, "round", round)
	return nil
}

func (s *FightService) SetWinner(id int, winnerID int) error {
	ctx := context.Background()
	err := s.fightStore.SetWinner(ctx, id, winnerID)
	if err != nil {
		s.logger.Error("Failed to set fight winner", err)
		return err
	}

	s.logger.Info("Fight winner set", "id", id, "winner_id", winnerID)
	return nil
}

func (s *FightService) GetByBoxer(boxerID int, limit int) ([]*Fight, error) {
	ctx := context.Background()

	fightsModel, err := s.fightStore.GetByBoxer(ctx, boxerID, limit)
	if err != nil {
		s.logger.Error("Failed to get fights by boxer", err)
		return nil, err
	}

	// Convert from model.Fight to Fight (internal/fight package type)
	fights := make([]*Fight, len(fightsModel))
	for i, f := range fightsModel {
		fights[i] = &Fight{
			ID:                f.ID,
			Boxer1ID:          &f.Boxer1ID,
			Boxer2ID:          &f.Boxer2ID,
			Status:            string(f.Status),
			ScheduledTime:     f.ScheduledTime,
			StartTime:         f.StartTime,
			EndTime:           f.EndTime,
			WinnerID:          f.WinnerID,
			Round:             f.Round,
			SimulationVersion: f.SimulationVersion,
			Data:              f.Data,
			CreatedAt:         f.CreatedAt,
			UpdatedAt:         f.UpdatedAt,
		}
	}

	return fights, nil
}

func (s *FightService) GetCompleted(limit int) ([]*Fight, error) {
	ctx := context.Background()

	fightsModel, err := s.fightStore.GetCompleted(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to get completed fights", err)
		return nil, err
	}

	// Convert from model.Fight to Fight (internal/fight package type)
	return convertFightsModel(fightsModel), nil
}

func (s *FightService) Delete(id int) error {
	ctx := context.Background()
	err := s.fightStore.Delete(ctx, id)
	if err != nil {
		s.logger.Error("Failed to delete fight", err)
		return err
	}

	s.logger.Info("Fight deleted", "id", id)
	return nil
}

// convertFightsModel converts a slice of model.Fight to Fight (internal/fight package type)
func convertFightsModel(fightsModel []*model.Fight) []*Fight {
	fights := make([]*Fight, len(fightsModel))
	for i, f := range fightsModel {
		fights[i] = &Fight{
			ID:                f.ID,
			Boxer1ID:          &f.Boxer1ID,
			Boxer2ID:          &f.Boxer2ID,
			Status:            string(f.Status),
			ScheduledTime:     f.ScheduledTime,
			StartTime:         f.StartTime,
			EndTime:           f.EndTime,
			WinnerID:          f.WinnerID,
			Round:             f.Round,
			SimulationVersion: f.SimulationVersion,
			Data:              f.Data,
			CreatedAt:         f.CreatedAt,
			UpdatedAt:         f.UpdatedAt,
		}
	}
	return fights
}

func (s *FightService) Serialize(fight *Fight) ([]byte, error) {
	return json.Marshal(fight)
}

func (s *FightService) Deserialize(data []byte) (*Fight, error) {
	var fight Fight
	if err := json.Unmarshal(data, &fight); err != nil {
		return nil, err
	}
	return &fight, nil
}

// SimulateFight simulates a fight between two boxers using database transactions.
// Uses SELECT FOR UPDATE to prevent concurrent modifications and ensures atomic updates.
func (s *FightService) SimulateFight(fightID int) error {
	ctx := context.Background()

	// Begin transaction with serializable isolation for maximum consistency
	tx, err := s.fightStore.BeginFightTx(ctx)
	if err != nil {
		s.logger.Error("Failed to begin transaction", "fight_id", fightID, "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Safe to call even after Commit; only rolls back if not committed
	}()

	// Lock the fight row for update (prevents concurrent simulations of same fight)
	fight, err := s.fightStore.GetByIDWithLockTx(ctx, tx, fightID)
	if err != nil {
		s.logger.Error("Failed to lock fight", "fight_id", fightID, "error", err)
		return fmt.Errorf("failed to lock fight: %w", err)
	}

	// Verify fight is scheduled before processing
	if string(fight.Status) != "scheduled" {
		s.logger.Debug("Fight not scheduled, skipping", "fight_id", fightID, "status", fight.Status)
		return nil
	}

	// Lock both boxer rows for update (prevents concurrent training/fights)
	boxer1, err := s.boxerStore.GetByIDWithLockTx(ctx, tx, fight.Boxer1ID)
	if err != nil {
		s.logger.Error("Failed to lock boxer1", "boxer_id", fight.Boxer1ID, "error", err)
		return fmt.Errorf("failed to lock boxer1: %w", err)
	}

	boxer2, err := s.boxerStore.GetByIDWithLockTx(ctx, tx, fight.Boxer2ID)
	if err != nil {
		s.logger.Error("Failed to lock boxer2", "boxer_id", fight.Boxer2ID, "error", err)
		return fmt.Errorf("failed to lock boxer2: %w", err)
	}

	s.logger.Info("Simulating fight", "boxer1", boxer1.Name, "boxer2", boxer2.Name)

	// Update fight status to in_progress within transaction
	if err := s.fightStore.UpdateStatusTx(ctx, tx, fightID, "in_progress"); err != nil {
		return fmt.Errorf("failed to update fight status: %w", err)
	}

	// MAT-114: Use configurable simulation parameters
	maxRounds := s.cfg.Fight.Simulation.MaxRounds
	damageMultiplier := s.cfg.Fight.Simulation.DamageMultiplier
	evasionThreshold := s.cfg.Fight.Simulation.EvasionThreshold
	energyDrain := s.cfg.Fight.Simulation.EnergyDrain
	experienceGain := s.cfg.Fight.Simulation.ExperienceGain
	simVersion := s.cfg.Fight.Simulation.Version

	// MAT-114: Create deterministic seed for reproducible simulations
	// Seed is derived from fight data to ensure same fight always produces same result
	seed := int64(fightID*1000 + fight.Boxer1ID + fight.Boxer2ID)
	rng := rand.New(rand.NewSource(seed))

	// Run the fight simulation logic (still within transaction)
	for currentRound := 1; currentRound <= maxRounds; currentRound++ {
		if boxer1.Health <= 0 || boxer2.Health <= 0 {
			break
		}

		// Process attacks (modifies boxer state in memory only)
		s.processAttack(currentRound, boxer1, boxer2, damageMultiplier, evasionThreshold, energyDrain, rng)

		// Recover some energy
		boxer1.Energy = math.Min(boxer1.Energy+20, 100)
		boxer2.Energy = math.Min(boxer2.Energy+20, 100)

		// Update boxer stats in database (within transaction)
		if err := s.boxerStore.UpdateTx(ctx, tx, boxer1); err != nil {
			return fmt.Errorf("failed to update boxer1: %w", err)
		}
		if err := s.boxerStore.UpdateTx(ctx, tx, boxer2); err != nil {
			return fmt.Errorf("failed to update boxer2: %w", err)
		}

		// Update fight data and round (within transaction)
		fightData := map[string]any{
			"round":         currentRound,
			"boxer1_health": boxer1.Health,
			"boxer1_energy": boxer1.Energy,
			"boxer2_health": boxer2.Health,
			"boxer2_energy": boxer2.Energy,
		}

		if err := s.fightStore.UpdateRoundTx(ctx, tx, fightID, currentRound); err != nil {
			return fmt.Errorf("failed to update round: %w", err)
		}
		if err := s.fightStore.SetDataTx(ctx, tx, fightID, fightData); err != nil {
			return fmt.Errorf("failed to set fight data: %w", err)
		}

		s.logger.Debug("Round completed", "round", currentRound, "boxer1", boxer1.Name, "boxer1_hp", boxer1.Health, "boxer2", boxer2.Name, "boxer2_hp", boxer2.Health)
	}

	// Determine winner and finalize fight (within transaction)
	var winnerID *int
	if boxer1.Health > boxer2.Health {
		winnerID = &fight.Boxer1ID
		s.logger.Info("Fight winner", "winner", boxer1.Name)
	} else if boxer2.Health > boxer1.Health {
		winnerID = &fight.Boxer2ID
		s.logger.Info("Fight winner", "winner", boxer2.Name)
	}

	if winnerID != nil {
		// Set winner and mark as completed (atomic update)
		if err := s.fightStore.SetWinnerTx(ctx, tx, fightID, *winnerID); err != nil {
			return fmt.Errorf("failed to set winner: %w", err)
		}

		// Award experience to winner (within transaction)
		boxerToUpdate := boxer1
		if *winnerID == fight.Boxer2ID {
			boxerToUpdate = boxer2
		}
		boxerToUpdate.Experience += experienceGain

		// Check for level up
		oldLevel := boxerToUpdate.Level
		newLevel := int(boxerToUpdate.Experience/100.0) + 1
		if newLevel > oldLevel {
			boxerToUpdate.Level = newLevel
			s.logger.Info("Boxer leveled up", "boxer", boxerToUpdate.Name, "old_level", oldLevel, "new_level", newLevel)
		}

		// Update winner boxer stats (within transaction)
		if err := s.boxerStore.UpdateTx(ctx, tx, boxerToUpdate); err != nil {
			return fmt.Errorf("failed to update winner experience: %w", err)
		}
	} else {
		// Draw - just mark as completed
		if err := s.fightStore.UpdateStatusTx(ctx, tx, fightID, "completed"); err != nil {
			return fmt.Errorf("failed to mark draw: %w", err)
		}
		s.logger.Info("Fight ended in a draw")
	}

	// MAT-114: Record simulation version for replay compatibility
	if err := s.fightStore.SetSimulationVersionTx(ctx, tx, fightID, simVersion); err != nil {
		s.logger.Error("Failed to record simulation version", "fight_id", fightID, "version", simVersion, "error", err)
		// Don't fail the entire transaction - just log the error
	}

	// Commit the entire transaction - all updates happen atomically
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit fight transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info("Fight simulation completed successfully", "fight_id", fightID, "version", simVersion, "seed", seed)
	return nil
}

// processAttack handles the attack logic for both boxers using seeded RNG for determinism (MAT-114)
func (s *FightService) processAttack(
	round int,
	boxer1, boxer2 *model.Boxer,
	damageMultiplier, evasionThreshold, energyDrain float64,
	rng *rand.Rand,
) {
	// Boxer 1 attacks
	attack1 := boxer1.Strength * damageMultiplier
	evasion1 := boxer2.Agility / 100.0

	// MAT-114: Use seeded RNG for deterministic evasion checks
	if evasion1 > evasionThreshold && round%3 != 0 {
		// Boxer 2 evades (deterministic based on stats and round)
		s.logger.Debug("Boxer 2 evaded attack", "round", round)
	} else {
		// Add randomness to damage within a small range for variety (MAT-114)
		damageVariance := rng.Float64() * 0.2   // ±10% variance
		damageVariance = (damageVariance - 0.1) // Center around 0

		damage := attack1 * (1 - boxer2.Defense/100.0) * (1 + damageVariance)
		boxer2.Health -= damage
		boxer2.Energy -= energyDrain

		s.logger.Debug("Boxer 1 hit Boxer 2",
			"damage", damage,
			"boxer2_health", boxer2.Health,
			"boxer2_energy", boxer2.Energy,
			"round", round)
	}

	// Boxer 2 attacks
	attack2 := boxer2.Strength * damageMultiplier
	evasion2 := boxer1.Agility / 100.0

	// MAT-114: Use seeded RNG for deterministic evasion checks
	if evasion2 > evasionThreshold && round%3 != 0 {
		// Boxer 1 evades (deterministic based on stats and round)
		s.logger.Debug("Boxer 1 evaded attack", "round", round)
	} else {
		// Add randomness to damage within a small range for variety (MAT-114)
		damageVariance := rng.Float64() * 0.2   // ±10% variance
		damageVariance = (damageVariance - 0.1) // Center around 0

		damage := attack2 * (1 - boxer1.Defense/100.0) * (1 + damageVariance)
		boxer1.Health -= damage
		boxer1.Energy -= energyDrain

		s.logger.Debug("Boxer 2 hit Boxer 1",
			"damage", damage,
			"boxer1_health", boxer1.Health,
			"boxer1_energy", boxer1.Energy,
			"round", round)
	}
}
