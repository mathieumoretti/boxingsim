package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	boxerdb "github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/store"
)

type FightService struct {
	db         *sql.DB
	eventStore *store.ScheduledEventStore
}

func NewFightService(db interface{}, eventStore *store.ScheduledEventStore) *FightService {
	pdb := db.(*PostgresDBWrapper)
	return &FightService{
		db:         pdb.Conn,
		eventStore: eventStore,
	}
}

type PostgresDBWrapper struct {
	Conn *sql.DB
}

func (s *FightService) BookFight(ctx context.Context, boxer1ID int,
	boxer2ID int, scheduledTime time.Time, round int,
) error {
	if boxer1ID <= 0 || boxer2ID <= 0 {
		return errors.New("invalid request parameters")
	}

	exists, err := boxerdb.BoxerExists(s.db, boxer1ID)
	if err != nil || !exists {
		return fmt.Errorf("boxer does not exist: ID %d", boxer1ID)
	}

	exists2, err := boxerdb.BoxerExists(s.db, boxer2ID)
	if err != nil || !exists2 {
		return fmt.Errorf("boxer does not exist: ID %d", boxer2ID)
	}

	inUse, _ := boxerdb.BoxerInFight(s.db, boxer1ID)
	if inUse {
		return fmt.Errorf("%w: boxer %d is currently involved in another fight", boxerdb.ErrBoxerInUse, boxer1ID)
	}

	inUse2, _ := boxerdb.BoxerInFight(s.db, boxer2ID)
	if inUse2 {
		return fmt.Errorf("%w: boxer %d is currently involved in another fight", boxerdb.ErrBoxerInUse, boxer2ID)
	}

	st := scheduledTime
	fight := &model.FightCreate{
		Boxer1ID:      boxer1ID,
		Boxer2ID:      boxer2ID,
		ScheduledTime: &st,
		Round:         round,
	}

	// Create the fight and get the ID
	fightID, err := boxerdb.CreateFight(s.db, fight)
	if err != nil {
		return fmt.Errorf("failed to create fight: %w", err)
	}

	// Create a scheduled event for the fight simulation (MAT-99)
	if s.eventStore != nil && scheduledTime.After(time.Now()) {
		eventData, _ := json.Marshal(map[string]interface{}{
			"fight_id": fightID,
		})

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
	return boxerdb.GetActiveFights(s.db, statuses)
}

func (s *FightService) GetFightByID(ctx context.Context, id int) (*model.Fight, error) {
	if id <= 0 {
		return nil, errors.New("invalid fight id")
	}
	return boxerdb.GetFightByID(s.db, id)
}
