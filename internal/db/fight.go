package db

import (
	"database/sql"
	"errors"
	"time"

	"boxing/internal/model"
)

var (
	ErrBoxerNotFound = errors.New("boxer not found")
)

// GetFightByID retrieves a fight by ID
func GetFightByID(db *sql.DB, id int) (*model.Fight, error) {
	query := `
		SELECT id, attacker_id, defender_id, attacker_hp, defender_hp,
		       status, turns, started_at, end_time, created_at, updated_at
		FROM fights
		WHERE id = ?
	`
	fight := &model.Fight{}
	err := db.QueryRow(query, id).Scan(
		&fight.ID,
		&fight.AttackerID,
		&fight.DefenderID,
		&fight.AttackerHP,
		&fight.DefenderHP,
		&fight.Status,
		&fight.Turns,
		&fight.StartedAt,
		&fight.EndTime,
		&fight.CreatedAt,
		&fight.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFightNotFound
		}
		return nil, err
	}

	return fight, nil
}

// CreateFight creates a new fight
func CreateFight(db *sql.DB, fight *model.FightCreate) error {
	query := `
		INSERT INTO fights (attacker_id, defender_id, attacker_hp, defender_hp, status, turns)
		VALUES (?, ?, ?, ?, 'pending', 0)
	`
	result, err := db.Exec(query, fight.AttackerID, fight.DefenderID, fight.AttackerHP, fight.DefenderHP)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	fight.ID = int(id)
	return nil
}

// BoxerInFight checks if a boxer is currently in a fight
func BoxerInFight(db *sql.DB, boxerID int) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM fights
		WHERE (attacker_id = ? OR defender_id = ?)
		  AND status IN ('pending', 'in_progress')
	`
	var inFight bool
	err := db.QueryRow(query, boxerID, boxerID).Scan(&inFight)
	return inFight, err
}

// GetAvailableOpponents retrieves available opponents for a boxer
func GetAvailableOpponents(db *sql.DB, boxerID int) ([]*model.Boxer, error) {
	query := `
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility, experience, level,
		       created_at, updated_at
		FROM boxers
		WHERE id != ? AND user_id != (SELECT user_id FROM boxers WHERE id = ?)
	`
	rows, err := db.Query(query, boxerID, boxerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boxers := []*model.Boxer{}
	for rows.Next() {
		boxer := &model.Boxer{}
		err := rows.Scan(
			&boxer.ID,
			&boxer.UserID,
			&boxer.Name,
			&boxer.Nickname,
			&boxer.PositionX,
			&boxer.PositionY,
			&boxer.Health,
			&boxer.Energy,
			&boxer.Strength,
			&boxer.Defense,
			&boxer.Agility,
			&boxer.Experience,
			&boxer.Level,
			&boxer.CreatedAt,
			&boxer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		boxers = append(boxers, boxer)
	}

	return boxers, rows.Err()
}

// GetFightHistory retrieves fight history for a boxer
func GetFightHistory(db *sql.DB, boxerID int) ([]*model.Fight, error) {
	query := `
		SELECT id, attacker_id, defender_id, attacker_hp, defender_hp,
		       status, turns, started_at, end_time, created_at, updated_at
		FROM fights
		WHERE attacker_id = ? OR defender_id = ?
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := db.Query(query, boxerID, boxerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fights := []*model.Fight{}
	for rows.Next() {
		fight := &model.Fight{}
		err := rows.Scan(
			&fight.ID,
			&fight.AttackerID,
			&fight.DefenderID,
			&fight.AttackerHP,
			&fight.DefenderHP,
			&fight.Status,
			&fight.Turns,
			&fight.StartedAt,
			&fight.EndTime,
			&fight.CreatedAt,
			&fight.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		fights = append(fights, fight)
	}

	return fights, rows.Err()
}