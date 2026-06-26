package db

import (
	"database/sql"
	"errors"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrBoxerNotFound = errors.New("boxer not found")
	ErrFightNotFound = errors.New("fight not found")
)

// GetFightByID retrieves a fight by ID
func GetFightByID(db *sql.DB, id int) (*model.Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE id = $1
	`
	fight := &model.Fight{}
	err := db.QueryRow(query, id).Scan(
		&fight.ID,
		&fight.Boxer1ID,
		&fight.Boxer2ID,
		&fight.Status,
		&fight.ScheduledTime,
		&fight.StartTime,
		&fight.EndTime,
		&fight.WinnerID,
		&fight.Round,
		&fight.Data,
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
		INSERT INTO fights (boxer1_id, boxer2_id, scheduled_time, round)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(query, fight.Boxer1ID, fight.Boxer2ID, fight.ScheduledTime, fight.Round)
	if err != nil {
		return err
	}

	return nil
}

// BoxerInFight checks if a boxer is currently in a fight
func BoxerInFight(db *sql.DB, boxerID int) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM fights
		WHERE (boxer1_id = $1 OR boxer2_id = $1)
		  AND status IN ('scheduled', 'in_progress')
	`
	var inFight bool
	err := db.QueryRow(query, boxerID).Scan(&inFight)
	return inFight, err
}

// GetAvailableOpponents retrieves available opponents for a boxer
func GetAvailableOpponents(db *sql.DB, boxerID int) ([]*model.Boxer, error) {
	query := `
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility, experience, level,
		       created_at, updated_at
		FROM boxers
		WHERE id != $1 AND user_id != (SELECT user_id FROM boxers WHERE id = $1)
	`
	rows, err := db.Query(query, boxerID)
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
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE boxer1_id = $1 OR boxer2_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`
	rows, err := db.Query(query, boxerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fights := []*model.Fight{}
	for rows.Next() {
		fight := &model.Fight{}
		err := rows.Scan(
			&fight.ID,
			&fight.Boxer1ID,
			&fight.Boxer2ID,
			&fight.Status,
			&fight.ScheduledTime,
			&fight.StartTime,
			&fight.EndTime,
			&fight.WinnerID,
			&fight.Round,
			&fight.Data,
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