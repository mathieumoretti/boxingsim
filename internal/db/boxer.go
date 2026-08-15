package db

import (
	"database/sql"
	"errors"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrBoxerNotFound = errors.New("boxer not found")
)

// boxerSelectColumns defines the standard SELECT clause for boxing records.
const boxerSelectColumns = `
		id, user_id, name, nickname, position_x, position_y, health, energy, strength, defense,
		agility, experience, level, created_at, updated_at
`

// scanBoxers scans sql.Rows into a slice of Boxer pointers.
func scanBoxers(rows *sql.Rows) ([]*model.Boxer, error) {
	defer func() { _ = rows.Close() }()

	var boxers []*model.Boxer
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return boxers, nil
}

// GetBoxerByID retrieves a boxer by ID
func GetBoxerByID(db *sql.DB, id int) (*model.Boxer, error) {
	query := `
		SELECT id, user_id, name, nickname, position_x, position_y, health, energy,
		       strength, defense, agility, experience, level, created_at, updated_at
		FROM boxers
		WHERE id = $1
	`

	boxer := &model.Boxer{}
	err := db.QueryRow(query, id).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBoxerNotFound
		}
		return nil, err
	}

	return boxer, nil
}

// ListBoxersByUserID retrieves all boxers for a user by ID
func ListBoxersByUserID(db *sql.DB, userID int) ([]*model.Boxer, error) {
	query := `SELECT ` + boxerSelectColumns + ` FROM boxers WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}

	return scanBoxers(rows)
}

// ListAllBoxers retrieves all boxers in the system
func ListAllBoxers(db *sql.DB) ([]*model.Boxer, error) {
	query := `SELECT ` + boxerSelectColumns + ` FROM boxers ORDER BY created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	return scanBoxers(rows)
}

// ListBoxerByName retrieves a boxer by name (for checking duplicates during seeding).
func ListBoxerByName(db *sql.DB, name string) ([]*model.Boxer, error) {
	query := `SELECT ` + boxerSelectColumns + ` FROM boxers WHERE LOWER(name) = LOWER($1) ORDER BY created_at DESC`

	rows, err := db.Query(query, name)
	if err != nil {
		return nil, err
	}

	return scanBoxers(rows)
}

// BoxerExists checks if a boxer with the given ID exists.
func BoxerExists(db *sql.DB, id int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM boxers WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

