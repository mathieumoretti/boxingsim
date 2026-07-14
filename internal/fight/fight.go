package fight

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"time"

	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
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
	db       *sql.DB
	cfg      *config.Config
	logger   *logger.Logger
	boxerSvc *boxer.BoxerService
}

func NewFightService(db *sql.DB, cfg *config.Config, boxerSvc *boxer.BoxerService) *FightService {
	return &FightService{
		db:       db,
		cfg:      cfg,
		logger:   logger.New("FightService"),
		boxerSvc: boxerSvc,
	}
}

func (s *FightService) Schedule(boxer1ID, boxer2ID int, scheduledTime time.Time) (*Fight, error) {
	fight := &Fight{
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

	var err error

	_, err = s.db.Exec(`
		INSERT INTO fights (boxer1_id, boxer2_id, status, scheduled_time, round, data)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, boxer1ID, boxer2ID, "scheduled", scheduledTime, 1, fight.Data)

	if err != nil {
		s.logger.Error("Failed to schedule fight", err)
		return nil, err
	}

	// Get the last insert ID by querying the database
	var id int
	err = s.db.QueryRow(`
		SELECT id FROM fights
		WHERE boxer1_id = $1 AND boxer2_id = $2 AND status = 'scheduled'
		ORDER BY created_at DESC LIMIT 1
	`, boxer1ID, boxer2ID).Scan(&id)
	if err != nil {
		s.logger.Error("Failed to get fight ID", err)
		return nil, err
	}

	fight.ID = id
	fight.Boxer1ID = &boxer1ID
	fight.Boxer2ID = &boxer2ID

	s.logger.Info("Fight scheduled", "id", fight.ID)
	return fight, nil
}

func (s *FightService) GetByID(id int) (*Fight, error) {
	var fighter1ID sql.NullInt64
	var fighter2ID sql.NullInt64
	var scheduledTime sql.NullTime
	var startTime sql.NullTime
	var endTime sql.NullTime
	var winnerID sql.NullInt64
	var data interface{}

	var fight Fight
	err := s.db.QueryRow(`
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE id = $1
	`, id).Scan(
		&fight.ID,
		&fighter1ID,
		&fighter2ID,
		&fight.Status,
		&scheduledTime,
		&startTime,
		&endTime,
		&winnerID,
		&fight.Round,
		&data,
		&fight.CreatedAt,
		&fight.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.Error("Failed to get fight", err)
		return nil, err
	}

	if fighter1ID.Valid {
		boxer1ID := int(fighter1ID.Int64)
		fight.Boxer1ID = &boxer1ID
	}

	if fighter2ID.Valid {
		boxer2ID := int(fighter2ID.Int64)
		fight.Boxer2ID = &boxer2ID
	}

	if scheduledTime.Valid {
		fight.ScheduledTime = &scheduledTime.Time
	}

	if startTime.Valid {
		fight.StartTime = &startTime.Time
	}

	if endTime.Valid {
		fight.EndTime = &endTime.Time
	}

	if winnerID.Valid {
		winner := int(winnerID.Int64)
		fight.WinnerID = &winner
	}

	if data != nil {
		fight.Data = data.(map[string]interface{})
	}

	return &fight, nil
}

func (s *FightService) GetUpcoming(limit int) ([]*Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE status = 'scheduled' AND (scheduled_time > NOW() OR scheduled_time IS NULL)
		ORDER BY scheduled_time ASC LIMIT $1
	`
	return s.getFightsByQuery(query, limit)
}

func (s *FightService) GetInProgress(limit int) ([]*Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE status = 'in_progress'
		ORDER BY start_time DESC LIMIT $1
	`
	return s.getFightsByQuery(query, limit)
}

func (s *FightService) UpdateStatus(id int, status string) error {
	if status == "completed" {
		result, err := s.db.Exec(`
			UPDATE fights SET status = $1, end_time = CURRENT_TIMESTAMP
			WHERE id = $2
		`, status, id)
		if err != nil {
			s.logger.Error("Failed to update fight status", err)
			return err
		}
		if _, err := result.RowsAffected(); err != nil {
			return err
		}
	} else {
		_, err := s.db.Exec(`
			UPDATE fights SET status = $1
			WHERE id = $2
		`, status, id)
		if err != nil {
			s.logger.Error("Failed to update fight status", err)
			return err
		}
	}

	s.logger.Info("Fight status updated", "id", id, "status", status)
	return nil
}

func (s *FightService) UpdateRound(id int, round int) error {
	_, err := s.db.Exec(`
		UPDATE fights SET round = $1
		WHERE id = $2
	`, round, id)

	if err != nil {
		s.logger.Error("Failed to update fight round", err)
		return err
	}

	s.logger.Info("Fight round updated", "id", id, "round", round)
	return nil
}

func (s *FightService) SetWinner(id int, winnerID int) error {
	_, err := s.db.Exec(`
		UPDATE fights SET winner_id = $1, status = 'completed'
		WHERE id = $2
	`, winnerID, id)

	if err != nil {
		s.logger.Error("Failed to set fight winner", err)
		return err
	}

	s.logger.Info("Fight winner set", "id", id, "winner_id", winnerID)
	return nil
}

func (s *FightService) GetByBoxer(boxerID int, limit int) ([]*Fight, error) {
	rows, err := s.db.Query(`
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE boxer1_id = $1 OR boxer2_id = $1
		ORDER BY scheduled_time DESC LIMIT $2
	`, boxerID, limit)

	if err != nil {
		s.logger.Error("Failed to get fights by boxer", err)
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var fights []*Fight
	for rows.Next() {
		fight, err := s.scanFightRow(rows)
		if err != nil {
			s.logger.Error("Failed to scan fight row", err)
			continue
		}
		fights = append(fights, fight)
	}

	return fights, nil
}

func (s *FightService) GetCompleted(limit int) ([]*Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights WHERE status = 'completed'
		ORDER BY end_time DESC LIMIT $1
	`
	return s.getFightsByQuery(query, limit)
}

func (s *FightService) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM fights WHERE id = $1", id)
	if err != nil {
		s.logger.Error("Failed to delete fight", err)
		return err
	}

	s.logger.Info("Fight deleted", "id", id)
	return nil
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

// scanFightRow scans a single row from the fights table into a Fight struct
func (s *FightService) scanFightRow(rows *sql.Rows) (*Fight, error) {
	var fighter1ID sql.NullInt64
	var fighter2ID sql.NullInt64
	var scheduledTime sql.NullTime
	var startTime sql.NullTime
	var endTime sql.NullTime
	var winnerID sql.NullInt64
	var data interface{}

	var fight Fight
	if err := rows.Scan(
		&fight.ID,
		&fighter1ID,
		&fighter2ID,
		&fight.Status,
		&scheduledTime,
		&startTime,
		&endTime,
		&winnerID,
		&fight.Round,
		&data,
		&fight.CreatedAt,
		&fight.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if fighter1ID.Valid {
		boxer1ID := int(fighter1ID.Int64)
		fight.Boxer1ID = &boxer1ID
	}

	if fighter2ID.Valid {
		boxer2ID := int(fighter2ID.Int64)
		fight.Boxer2ID = &boxer2ID
	}

	if scheduledTime.Valid {
		fight.ScheduledTime = &scheduledTime.Time
	}

	if startTime.Valid {
		fight.StartTime = &startTime.Time
	}

	if endTime.Valid {
		fight.EndTime = &endTime.Time
	}

	if winnerID.Valid {
		winner := int(winnerID.Int64)
		fight.WinnerID = &winner
	}

	if data != nil {
		fight.Data = data.(map[string]interface{})
	}

	return &fight, nil
}

// getFightsByQuery executes a query and returns fights using scanFightRow
func (s *FightService) getFightsByQuery(query string, args ...interface{}) ([]*Fight, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.logger.Error("Failed to get fights by query", err)
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var fights []*Fight
	for rows.Next() {
		fight, err := s.scanFightRow(rows)
		if err != nil {
			s.logger.Error("Failed to scan fight row", err)
			continue
		}
		fights = append(fights, fight)
	}

	return fights, nil
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

// updateBoxers updates the boxers' stats in the database
func (s *FightService) updateBoxers(boxer1, boxer2 model.Boxer, boxer1ID, boxer2ID *int) error {
	_, err := s.db.Exec(`
		UPDATE boxers
		SET health = $1, energy = $2, position_x = $3, position_y = $4
		WHERE id = $5
	`, boxer1.Health, boxer1.Energy, boxer1.PositionX, boxer1.PositionY, *boxer1ID)

	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE boxers
		SET health = $1, energy = $2, position_x = $3, position_y = $4
		WHERE id = $5
	`, boxer2.Health, boxer2.Energy, boxer2.PositionX, boxer2.PositionY, *boxer2ID)

	return err
}

// updateFightData updates the fight data in the database
func (s *FightService) updateFightData(fight *Fight, boxer1, boxer2 model.Boxer, fightID int) error {
	// Update fight data
	fight.Data = map[string]interface{}{
		"round":         fight.Round,
		"boxer1_health": boxer1.Health,
		"boxer1_energy": boxer1.Energy,
		"boxer2_health": boxer2.Health,
		"boxer2_energy": boxer2.Energy,
	}

	// Convert data to JSON for storage
	var dataJSON string
	if dataBytes, marshalErr := json.Marshal(fight.Data); marshalErr == nil {
		dataJSON = string(dataBytes)
	}

	_, err := s.db.Exec(`UPDATE fights SET round = $1, data = $2 WHERE id = $3`, fight.Round, dataJSON, fightID)
	return err
}

// determineWinnerAndStatus determines the winner and updates the fight status
func (s *FightService) determineWinnerAndStatus(
	fightID int,
	fight *Fight,
	boxer1, boxer2 *model.Boxer,
	ctx context.Context) error {
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
