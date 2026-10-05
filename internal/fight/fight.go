package fight

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"
)

type Fight struct {
	ID            int
	Boxer1ID      *int
	Boxer2ID      *int
	Status        string
	ScheduledTime *time.Time
	StartTime     *time.Time
	EndTime       *time.Time
	WinnerID      *int
	Round         int
	Data          map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
			s.logger.Error("Failed to create scheduled event for fight %d: %v", id, err)
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
		ID:            fightModel.ID,
		Boxer1ID:      &fightModel.Boxer1ID,
		Boxer2ID:      &fightModel.Boxer2ID,
		Status:        string(fightModel.Status),
		ScheduledTime: fightModel.ScheduledTime,
		StartTime:     fightModel.StartTime,
		EndTime:       fightModel.EndTime,
		WinnerID:      fightModel.WinnerID,
		Round:         fightModel.Round,
		Data:          fightModel.Data,
		CreatedAt:     fightModel.CreatedAt,
		UpdatedAt:     fightModel.UpdatedAt,
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
			ID:            f.ID,
			Boxer1ID:      &f.Boxer1ID,
			Boxer2ID:      &f.Boxer2ID,
			Status:        string(f.Status),
			ScheduledTime: f.ScheduledTime,
			StartTime:     f.StartTime,
			EndTime:       f.EndTime,
			WinnerID:      f.WinnerID,
			Round:         f.Round,
			Data:          f.Data,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
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
			ID:            f.ID,
			Boxer1ID:      &f.Boxer1ID,
			Boxer2ID:      &f.Boxer2ID,
			Status:        string(f.Status),
			ScheduledTime: f.ScheduledTime,
			StartTime:     f.StartTime,
			EndTime:       f.EndTime,
			WinnerID:      f.WinnerID,
			Round:         f.Round,
			Data:          f.Data,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
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

// SimulateFight simulates a fight between two boxers
func (s *FightService) SimulateFight(fightID int) error {
	fight, err := s.GetByID(fightID)
	if err != nil {
		return err
	}

	if fight.Status != "scheduled" {
		return nil
	}

	// Update status to in_progress
	if err := s.UpdateStatus(fightID, "in_progress"); err != nil {
		return err
	}

	ctx := context.Background()
	boxer1, errGetBoxer1 := s.boxerSvc.GetBoxer(ctx, *fight.Boxer1ID)
	if errGetBoxer1 != nil {
		return errGetBoxer1
	}

	boxer2, errGetBoxer2 := s.boxerSvc.GetBoxer(ctx, *fight.Boxer2ID)
	if errGetBoxer2 != nil {
		return errGetBoxer2
	}

	s.logger.Info("Simulating fight", "boxer1", boxer1.Name, "boxer2", boxer2.Name)

	// Fight simulation logic - reduced complexity by splitting into smaller functions
	maxRounds := 12
	minHealth := 0.0
	damageMultiplier := 0.1
	evasionThreshold := 0.4
	energyDrain := 10.0

	for fight.Round = 1; fight.Round <= maxRounds; fight.Round++ {
		if boxer1.Health <= minHealth || boxer2.Health <= minHealth {
			break
		}

		// Process attacks
		s.processAttack(fight, boxer1, boxer2, damageMultiplier, evasionThreshold, energyDrain)

		// Recover some energy
		boxer1.Energy = math.Min(boxer1.Energy+20, 100)
		boxer2.Energy = math.Min(boxer2.Energy+20, 100)

		// Update fighters
		if err := s.updateBoxers(*boxer1, *boxer2, fight.Boxer1ID, fight.Boxer2ID); err != nil {
			return err
		}

		// Update fight data and database
		if err := s.updateFightData(fight, *boxer1, *boxer2, fightID); err != nil {
			return err
		}

		// Sleep a bit between rounds for visualization
		time.Sleep(500 * time.Millisecond)
	}

	// Determine winner and update status
	return s.determineWinnerAndStatus(fightID, fight, boxer1, boxer2, ctx)
}

// processAttack handles the attack logic for both boxers
func (s *FightService) processAttack(
	fight *Fight,
	boxer1, boxer2 *model.Boxer,
	damageMultiplier, evasionThreshold, energyDrain float64,
) {
	// Boxer 1 attacks
	attack1 := boxer1.Strength * damageMultiplier
	evasion1 := boxer2.Agility / 100.0

	if evasion1 > evasionThreshold && fight.Round%3 != 0 {
		// Boxer 2 evades
		s.logger.Debug("Boxer 2 evaded attack", "round", fight.Round)
	} else {
		damage := attack1 * (1 - boxer2.Defense/100.0)
		boxer2.Health -= damage
		boxer2.Energy -= energyDrain

		s.logger.Debug("Boxer 1 hit Boxer 2",
			"damage", damage,
			"boxer2_health", boxer2.Health,
			"boxer2_energy", boxer2.Energy,
			"round", fight.Round)
	}

	// Boxer 2 attacks
	attack2 := boxer2.Strength * damageMultiplier
	evasion2 := boxer1.Agility / 100.0

	if evasion2 > evasionThreshold && fight.Round%3 != 0 {
		// Boxer 1 evades
		s.logger.Debug("Boxer 1 evaded attack", "round", fight.Round)
	} else {
		damage := attack2 * (1 - boxer1.Defense/100.0)
		boxer1.Health -= damage
		boxer1.Energy -= energyDrain

		s.logger.Debug("Boxer 2 hit Boxer 1",
			"damage", damage,
			"boxer1_health", boxer1.Health,
			"boxer1_energy", boxer1.Energy,
			"round", fight.Round)
	}
}

// updateBoxers updates the boxers' health and energy using boxerStore
func (s *FightService) updateBoxers(boxer1, boxer2 model.Boxer, boxer1ID, boxer2ID *int) error {
	ctx := context.Background()

	// Update boxer 1
	boxer1.ID = *boxer1ID
	if err := s.boxerStore.Update(ctx, &boxer1); err != nil {
		return fmt.Errorf("failed to update boxer1 %d: %w", *boxer1ID, err)
	}

	// Update boxer 2
	boxer2.ID = *boxer2ID
	if err := s.boxerStore.Update(ctx, &boxer2); err != nil {
		return fmt.Errorf("failed to update boxer2 %d: %w", *boxer2ID, err)
	}

	return nil
}

// updateFightData updates the fight data in the database using fightStore
func (s *FightService) updateFightData(fight *Fight, boxer1, boxer2 model.Boxer, fightID int) error {
	ctx := context.Background()

	// Update fight data
	fightData := map[string]interface{}{
		"round":         fight.Round,
		"boxer1_health": boxer1.Health,
		"boxer1_energy": boxer1.Energy,
		"boxer2_health": boxer2.Health,
		"boxer2_energy": boxer2.Energy,
	}

	// Update round
	if err := s.fightStore.UpdateRound(ctx, fightID, fight.Round); err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}

	// Set data
	if err := s.fightStore.SetData(ctx, fightID, fightData); err != nil {
		return fmt.Errorf("failed to set fight data: %w", err)
	}

	return nil
}

// determineWinnerAndStatus determines the winner and updates the fight status
func (s *FightService) determineWinnerAndStatus(
	fightID int,
	fight *Fight,
	boxer1, boxer2 *model.Boxer,
	ctx context.Context,
) error {
	var winnerID *int
	if boxer1.Health > boxer2.Health {
		winnerID = fight.Boxer1ID
	} else if boxer2.Health > boxer1.Health {
		winnerID = fight.Boxer2ID
	}

	// Update fight status
	if winnerID != nil {
		if err := s.SetWinner(fightID, *winnerID); err != nil {
			return err
		}

		// Award experience
		experienceGain := 50.0
		if err := s.boxerSvc.UpdateStats(ctx, *winnerID, model.BoxerStats{
			Experience: experienceGain,
			Level:      int(experienceGain/100.0) + 1,
		}); err != nil {
			s.logger.Error("Failed to update winner experience", err)
		}

		s.logger.Info("Fight completed", "winner", winnerID, "boxer1_health", boxer1.Health, "boxer2_health", boxer2.Health)
	} else {
		// Draw
		if err := s.UpdateStatus(fightID, "completed"); err != nil {
			return err
		}
		s.logger.Info("Fight ended in draw", "boxer1_health", boxer1.Health, "boxer2_health", boxer2.Health)
	}

	return nil
}
