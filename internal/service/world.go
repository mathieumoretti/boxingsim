package service

import (
	"boxing/internal/db"
	"database/sql"
	"errors"
	"time"
	"math/rand"
)

var (
	ErrWorldTickNotFound = errors.New("world tick not found")
)

// WorldService handles world events and game ticks
type WorldService struct {
	db *sql.DB
}

// NewWorldService creates a new WorldService
func NewWorldService(db *sql.DB) *WorldService {
	return &WorldService{db: db}
}

// ProcessEvent processes a world event
func (s *WorldService) ProcessEvent(eventType string, boxerID int, eventData map[string]interface{}) error {
	switch eventType {
	case "boxer_moved":
		return s.processBoxerMove(boxerID, eventData)
	case "level_up":
		return s.processLevelUp(boxerID)
	case "xp_gain":
		return s.processXPGain(boxerID)
	default:
		return errors.New("unknown event type")
	}
}

// processBoxerMove handles boxer movement events
func (s *WorldService) processBoxerMove(boxerID int, eventData map[string]interface{}) error {
	x, okX := eventData["x"].(float64)
	y, okY := eventData["y"].(float64)

	if !okX || !okY {
		return errors.New("invalid position data")
	}

	query := `
		UPDATE boxers SET
		    position_x = ?,
		    position_y = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := s.db.Exec(query, x, y, boxerID)
	return err
}

// processLevelUp handles level up events
func (s *WorldService) processLevelUp(boxerID int) error {
	// Update boxer level
	query := `
		UPDATE boxers SET
		    level = level + 1,
		    strength = strength + 5,
		    defense = defense + 3,
		    agility = agility + 4,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := s.db.Exec(query, boxerID)
	return err
}

// processXPGain handles XP gain events
func (s *WorldService) processXPGain(boxerID int) error {
	// Calculate XP needed for next level
	boxer, err := db.GetBoxerByID(s.db.DB, boxerID)
	if err != nil {
		return err
	}

	xpNeeded := float64(boxer.Level * 100)
	if boxer.Experience < xpNeeded {
		// Not enough XP for level up yet
		return nil
	}

	// Remove XP needed and level up
	experience := boxer.Experience - xpNeeded
	query := `
		UPDATE boxers SET
		    experience = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.Exec(query, experience, boxerID)
	if err != nil {
		return err
	}

	return nil
}

// WorldTick represents a world game tick
type WorldTick struct {
	ID          int
	TickNumber  int
	StartTime   time.Time
	EndTime     time.Time
	ProcessedAt time.Time
}

// StartWorldTick starts a world tick
func (s *WorldService) StartWorldTick() (*WorldTick, error) {
	query := `
		INSERT INTO world_ticks (tick_number, start_time)
		VALUES ((SELECT COALESCE(MAX(tick_number), 0) + 1 FROM world_ticks), CURRENT_TIMESTAMP)
	`
	result, err := s.db.Exec(query)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	tickNumber := 1
	err = s.db.QueryRow("SELECT COALESCE(MAX(tick_number), 0) + 1 FROM world_ticks").Scan(&tickNumber)
	if err != nil {
		return nil, err
	}

	return &model.WorldTick{
		ID:        int(id),
		TickNumber: tickNumber,
		StartTime:  time.Now(),
	}, nil
}

// EndWorldTick ends a world tick
func (s *WorldService) EndWorldTick(tickID int) error {
	query := `
		UPDATE world_ticks SET
		    end_time = CURRENT_TIMESTAMP,
		    processed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := s.db.Exec(query, tickID)
	return err
}

// GetTick retrieves a world tick by ID
func (s *WorldService) GetTick(tickID int) (*WorldTick, error) {
	query := `
		SELECT id, tick_number, start_time, end_time, processed_at
		FROM world_ticks
		WHERE id = ?
	`
	tick := &WorldTick{}
	err := s.db.QueryRow(query, tickID).Scan(
		&tick.ID,
		&tick.TickNumber,
		&tick.StartTime,
		&tick.EndTime,
		&tick.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWorldTickNotFound
		}
		return nil, err
	}

	return tick, nil
}

// GetActiveTicks retrieves all active world ticks
func (s *WorldService) GetActiveTicks() ([]*WorldTick, error) {
	query := `
		SELECT id, tick_number, start_time, end_time, processed_at
		FROM world_ticks
		WHERE end_time IS NULL
		ORDER BY start_time DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ticks := []*WorldTick{}
	for rows.Next() {
		tick := &WorldTick{}
		err := rows.Scan(
			&tick.ID,
			&tick.TickNumber,
			&tick.StartTime,
			&tick.EndTime,
			&tick.ProcessedAt,
		)
		if err != nil {
			return nil, err
		}
		ticks = append(ticks, tick)
	}

	return ticks, rows.Err()
}

// AutoHealBoxers heals injured boxers at the end of each tick
func (s *WorldService) AutoHealBoxers() error {
	query := `
		UPDATE boxers SET
		    health = LEAST(health + 10, max_health),
		    updated_at = CURRENT_TIMESTAMP
		WHERE health < max_health
	`
	_, err := s.db.Exec(query)
	return err
}

// AutoEnergyRegen regenerates energy for boxers
func (s *WorldService) AutoEnergyRegen() error {
	query := `
		UPDATE boxers SET
		    energy = LEAST(energy + 15, max_energy),
		    updated_at = CURRENT_TIMESTAMP
		WHERE energy < max_energy
	`
	_, err := s.db.Exec(query)
	return err
}

// processRandomEvents generates and processes random world events
func (s *WorldService) processRandomEvents() error {
	// Get all boxers
	query := `
		SELECT id, experience, level FROM boxers WHERE status = 'active'
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var boxerID, level int
		var experience float64
		err := rows.Scan(&boxerID, &experience, &level)
		if err != nil {
			return err
		}

		// Random chance for level up
		if level > 0 && rand.Float64() < 0.1 {
			err := s.processLevelUp(boxerID)
			if err != nil {
				continue
			}
		}

		// Random chance for XP gain
		if rand.Float64() < 0.05 {
			experienceGain := rand.Float64() * 50
			query := `
				UPDATE boxers SET
				    experience = experience + ?,
				    updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`
			_, err := s.db.Exec(query, experienceGain, boxerID)
			if err != nil {
				continue
			}
		}
	}

	return rows.Err()
}

// StartWorld starts the world simulation
func (s *WorldService) StartWorld() error {
	// Start a new tick
	tick, err := s.StartWorldTick()
	if err != nil {
		return err
	}

	// Process events
	err = s.processRandomEvents()
	if err != nil {
		return err
	}

	// Heal boxers
	err = s.AutoHealBoxers()
	if err != nil {
		return err
	}

	// Regenerate energy
	err = s.AutoEnergyRegen()
	if err != nil {
		return err
	}

	// End the tick
	return s.EndWorldTick(tick.ID)
}

// StopWorld stops the world simulation
func (s *WorldService) StopWorld() error {
	// Stop all active ticks
	query := `
		UPDATE world_ticks
		SET end_time = CURRENT_TIMESTAMP, processed_at = CURRENT_TIMESTAMP
		WHERE end_time IS NULL
	`
	_, err := s.db.Exec(query)
	return err
}