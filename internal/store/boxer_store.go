package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/mormm/boxing/internal/boxer"
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
			experience, level, fatigue_score, forced_rest_until,
			wins, losses, draws, knockouts, knockdowns_suffered,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
		           $15, $16, $17, $18, $19, $20, $21)
		RETURNING id`

	now := time.Now()
	boxer.CreatedAt = now
	boxer.UpdatedAt = now

	err := s.db.QueryRowContext(ctx, query,
		boxer.UserID, boxer.Name, boxer.Nickname, boxer.PositionX, boxer.PositionY,
		boxer.Health, boxer.Energy, boxer.Strength, boxer.Defense, boxer.Agility,
		boxer.Experience, boxer.Level, boxer.FatigueScore, boxer.ForcedRestUntil,
		boxer.Wins, boxer.Losses, boxer.Draws, boxer.Knockouts, boxer.KnockdownsSuffered,
		boxer.CreatedAt, boxer.UpdatedAt,
	).Scan(&boxer.ID)

	return err
}

// GetByID retrieves a boxer by ID
func (s *BoxerStore) GetByID(ctx context.Context, id int) (*model.Boxer, error) {
	query := `
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility,
		       experience, level, fatigue_score, forced_rest_until,
		       wins, losses, draws, knockouts, knockdowns_suffered,
		       created_at, updated_at
		FROM boxers WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, id)

	boxer := &model.Boxer{}
	err := row.Scan(
		&boxer.ID, &boxer.UserID, &boxer.Name, &boxer.Nickname, &boxer.PositionX, &boxer.PositionY,
		&boxer.Health, &boxer.Energy, &boxer.Strength, &boxer.Defense, &boxer.Agility,
		&boxer.Experience, &boxer.Level, &boxer.FatigueScore, &boxer.ForcedRestUntil,
		&boxer.Wins, &boxer.Losses, &boxer.Draws, &boxer.Knockouts, &boxer.KnockdownsSuffered,
		&boxer.CreatedAt, &boxer.UpdatedAt,
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
		       experience, level, fatigue_score, forced_rest_until,
		       wins, losses, draws, knockouts, knockdowns_suffered,
		       created_at, updated_at
		FROM boxers WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var boxers []*model.Boxer = make([]*model.Boxer, 0)
	for rows.Next() {
		boxer := &model.Boxer{}
		err := rows.Scan(
			&boxer.ID, &boxer.UserID, &boxer.Name, &boxer.Nickname, &boxer.PositionX, &boxer.PositionY,
			&boxer.Health, &boxer.Energy, &boxer.Strength, &boxer.Defense, &boxer.Agility,
			&boxer.Experience, &boxer.Level, &boxer.FatigueScore, &boxer.ForcedRestUntil,
			&boxer.Wins, &boxer.Losses, &boxer.Draws, &boxer.Knockouts, &boxer.KnockdownsSuffered,
			&boxer.CreatedAt, &boxer.UpdatedAt,
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

// Update updates a boxer's information
func (s *BoxerStore) Update(ctx context.Context, boxer *model.Boxer) error {
	query := `
		UPDATE boxers SET
			name = $1, nickname = $2, position_x = $3, position_y = $4,
			health = $5, energy = $6, strength = $7, defense = $8, agility = $9,
			experience = $10, level = $11, fatigue_score = $12, forced_rest_until = $13,
			wins = $14, losses = $15, draws = $16, knockouts = $17, knockdowns_suffered = $18,
			updated_at = $19
		WHERE id = $20`

	now := time.Now()
	boxer.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		boxer.Name, boxer.Nickname, boxer.PositionX, boxer.PositionY,
		boxer.Health, boxer.Energy, boxer.Strength, boxer.Defense, boxer.Agility,
		boxer.Experience, boxer.Level, boxer.FatigueScore, boxer.ForcedRestUntil,
		boxer.Wins, boxer.Losses, boxer.Draws, boxer.Knockouts, boxer.KnockdownsSuffered,
		boxer.UpdatedAt, boxer.ID,
	)

	return err
}

// Delete deletes a boxer by ID
func (s *BoxerStore) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM boxers WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// UpdateFightResult atomically updates fight statistics for two boxers
func (s *BoxerStore) UpdateFightResult(
	ctx context.Context,
	boxer1ID, boxer2ID int,
	result boxer.FightResult,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Build update query for boxer 1
	var boxer1SQL string
	var boxer1Args []interface{}

	if result.IsDraw {
		boxer1SQL = "UPDATE boxers SET draws = draws + $1, updated_at = $2 WHERE id = $3"
		boxer1Args = []interface{}{1, time.Now(), boxer1ID}
	} else if result.Boxer1Wins {
		boxer1SQL = "UPDATE boxers SET wins = wins + 1"
		if result.IsKnockout {
			boxer1SQL += ", knockouts = knockouts + 1"
		}
		boxer1SQL += ", updated_at = $1 WHERE id = $2"
		boxer1Args = []interface{}{time.Now(), boxer1ID}
	} else {
		boxer1SQL = "UPDATE boxers SET losses = losses + 1"
		if result.Boxer1Knockdowned {
			boxer1SQL += ", knockdowns_suffered = knockdowns_suffered + 1"
		}
		boxer1SQL += ", updated_at = $1 WHERE id = $2"
		boxer1Args = []interface{}{time.Now(), boxer1ID}
	}

	// Execute boxer 1 update
	_, err = tx.ExecContext(ctx, boxer1SQL, boxer1Args...)
	if err != nil {
		return err
	}

	// Build update query for boxer 2
	var boxer2SQL string
	var boxer2Args []interface{}

	if result.IsDraw {
		boxer2SQL = "UPDATE boxers SET draws = draws + $1, updated_at = $2 WHERE id = $3"
		boxer2Args = []interface{}{1, time.Now(), boxer2ID}
	} else if !result.Boxer1Wins { // boxer 2 wins
		boxer2SQL = "UPDATE boxers SET wins = wins + 1"
		if result.IsKnockout {
			boxer2SQL += ", knockouts = knockouts + 1"
		}
		boxer2SQL += ", updated_at = $1 WHERE id = $2"
		boxer2Args = []interface{}{time.Now(), boxer2ID}
	} else {
		boxer2SQL = "UPDATE boxers SET losses = losses + 1"
		if result.Boxer2Knockdowned {
			boxer2SQL += ", knockdowns_suffered = knockdowns_suffered + 1"
		}
		boxer2SQL += ", updated_at = $1 WHERE id = $2"
		boxer2Args = []interface{}{time.Now(), boxer2ID}
	}

	// Execute boxer 2 update
	_, err = tx.ExecContext(ctx, boxer2SQL, boxer2Args...)
	if err != nil {
		return err
	}

	return tx.Commit()
}
