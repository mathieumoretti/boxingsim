package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// BoxerStore implements the BoxerRepository interface
type BoxerStore struct {
	db *sql.DB
}

func NewBoxerStore(db *sql.DB) *BoxerStore {
	return &BoxerStore{
		db: db,
	}
}

// Create creates a new boxer in the database
func (s *BoxerStore) Create(ctx context.Context, boxer *model.Boxer) error {
	query := `
		INSERT INTO boxers (
			user_id, name, nickname, position_x, position_y,
			health, energy, strength, defense, agility,
			experience, level, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`

	now := time.Now()
	boxer.CreatedAt = now
	boxer.UpdatedAt = now

	err := s.db.QueryRowContext(ctx, query,
		boxer.UserID, boxer.Name, boxer.Nickname, boxer.PositionX, boxer.PositionY,
		boxer.Health, boxer.Energy, boxer.Strength, boxer.Defense, boxer.Agility,
		boxer.Experience, boxer.Level, boxer.CreatedAt, boxer.UpdatedAt,
	).Scan(&boxer.ID)

	return err
}

// GetByID retrieves a boxer by ID
func (s *BoxerStore) GetByID(ctx context.Context, id int) (*model.Boxer, error) {
	query := `
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility,
		       experience, level, created_at, updated_at
		FROM boxers WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, id)

	boxer := &model.Boxer{}
	err := row.Scan(
		&boxer.ID, &boxer.UserID, &boxer.Name, &boxer.Nickname, &boxer.PositionX, &boxer.PositionY,
		&boxer.Health, &boxer.Energy, &boxer.Strength, &boxer.Defense, &boxer.Agility,
		&boxer.Experience, &boxer.Level, &boxer.CreatedAt, &boxer.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return boxer, nil
}

// GetByUserID retrieves all boxers for a user
func (s *BoxerStore) GetByUserID(ctx context.Context, userID int) ([]*model.Boxer, error) {
	query := `
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility,
		       experience, level, created_at, updated_at
		FROM boxers WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var boxers []*model.Boxer
	for rows.Next() {
		boxer := &model.Boxer{}
		err := rows.Scan(
			&boxer.ID, &boxer.UserID, &boxer.Name, &boxer.Nickname, &boxer.PositionX, &boxer.PositionY,
			&boxer.Health, &boxer.Energy, &boxer.Strength, &boxer.Defense, &boxer.Agility,
			&boxer.Experience, &boxer.Level, &boxer.CreatedAt, &boxer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		boxers = append(boxers, boxer)
	}

	return boxers, nil
}

// Update updates a boxer's information
func (s *BoxerStore) Update(ctx context.Context, boxer *model.Boxer) error {
	query := `
		UPDATE boxers SET
			name = $1, nickname = $2, position_x = $3, position_y = $4,
			health = $5, energy = $6, strength = $7, defense = $8, agility = $9,
			experience = $10, level = $11, updated_at = $12
		WHERE id = $13`

	now := time.Now()
	boxer.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		boxer.Name, boxer.Nickname, boxer.PositionX, boxer.PositionY,
		boxer.Health, boxer.Energy, boxer.Strength, boxer.Defense, boxer.Agility,
		boxer.Experience, boxer.Level, boxer.UpdatedAt, boxer.ID,
	)

	return err
}

// Delete deletes a boxer by ID
func (s *BoxerStore) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM boxers WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
