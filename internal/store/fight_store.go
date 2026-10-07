package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
)

var (
	// ErrFightNotFound is returned when a fight does not exist.
	ErrFightNotFound = errors.New("fight not found")

	// ErrBoxerInUse is returned when a boxer is already involved in another fight.
	ErrBoxerInUse = errors.New("boxer is currently involved in another fight")
)

// FightStore implements data access operations for fights.
type FightStore struct {
	db *sql.DB
}

// NewFightStore creates a new FightStore instance.
func NewFightStore(db *sql.DB) *FightStore {
	return &FightStore{db: db}
}

// Create inserts a new fight into the database and returns the generated ID.
func (s *FightStore) Create(ctx context.Context, boxer1ID, boxer2ID int, scheduledTime *time.Time, round int) (int, error) {
	query := `
		INSERT INTO fights (boxer1_id, boxer2_id, scheduled_time, round)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id int
	err := s.db.QueryRowContext(ctx, query, boxer1ID, boxer2ID, scheduledTime, round).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetByID retrieves a fight by its ID.
func (s *FightStore) GetByID(ctx context.Context, id int) (*model.Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE id = $1`

	row := s.db.QueryRowContext(ctx, query, id)
	fight, err := s.scanFight(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFightNotFound
		}
		return nil, err
	}

	return fight, nil
}

// GetByIDWithLock retrieves a fight by its ID with an exclusive row lock (SELECT FOR UPDATE).
// This prevents concurrent modifications during fight simulation.
// Must be called within a transaction to properly hold the lock.
func (s *FightStore) GetByIDWithLock(ctx context.Context, id int) (*model.Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE id = $1
		FOR UPDATE`

	row := s.db.QueryRowContext(ctx, query, id)
	fight, err := s.scanFight(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFightNotFound
		}
		return nil, err
	}

	return fight, nil
}

// GetByIDWithLockTx retrieves a fight by its ID with an exclusive row lock using the provided transaction.
// This is the preferred method when performing fight simulation within a transaction.
func (s *FightStore) GetByIDWithLockTx(ctx context.Context, tx *sql.Tx, id int) (*model.Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE id = $1
		FOR UPDATE SKIP LOCKED`

	row := tx.QueryRowContext(ctx, query, id)
	fight, err := s.scanFightTx(tx, row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFightNotFound
		}
		return nil, err
	}

	return fight, nil
}

// GetActiveFights retrieves fights with specified statuses (e.g., "scheduled", "in_progress").
func (s *FightStore) GetActiveFights(ctx context.Context, statuses []string) ([]*model.Fight, error) {
	if len(statuses) == 0 {
		statuses = []string{"scheduled", "in_progress"}
	}

	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE status = ANY($1)
		ORDER BY scheduled_time ASC`

	rows, err := s.db.QueryContext(ctx, query, statuses)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanFights(rows)
}

// GetByBoxer retrieves fights involving a specific boxer.
func (s *FightStore) GetByBoxer(ctx context.Context, boxerID int, limit int) ([]*model.Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE boxer1_id = $1 OR boxer2_id = $1
		ORDER BY scheduled_time DESC
		LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, boxerID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanFights(rows)
}

// GetCompleted retrieves completed fights ordered by end time.
func (s *FightStore) GetCompleted(ctx context.Context, limit int) ([]*model.Fight, error) {
	query := `
		SELECT id, boxer1_id, boxer2_id, status, scheduled_time, start_time, end_time,
		       winner_id, round, data, created_at, updated_at
		FROM fights
		WHERE status = 'completed'
		ORDER BY end_time DESC
		LIMIT $1`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return s.scanFights(rows)
}

// GetUpcomingFightForBoxer retrieves the next upcoming fight for a specific boxer (MAT-106).
func (s *FightStore) GetUpcomingFightForBoxer(ctx context.Context, boxerID int) (*db.UpcomingFightResponse, error) {
	query := `
		SELECT
			f.id,
			CASE WHEN f.boxer1_id = $1 THEN f.boxer2_id ELSE f.boxer1_id END as opponent_id,
			CASE WHEN f.boxer1_id = $1 THEN b2.name ELSE b1.name END as opponent_name,
			CASE WHEN f.boxer1_id = $1 THEN b2.nickname ELSE b1.nickname END as opponent_nickname,
			f.scheduled_time,
			f.status,
			COALESCE(f.round, 12) as rounds
		FROM fights f
		LEFT JOIN boxers b1 ON f.boxer1_id = b1.id
		LEFT JOIN boxers b2 ON f.boxer2_id = b2.id
		WHERE (f.boxer1_id = $1 OR f.boxer2_id = $1)
		  AND f.status IN ('scheduled', 'in_progress')
		ORDER BY f.scheduled_time ASC
		LIMIT 1`

	var resp db.UpcomingFightResponse
	err := s.db.QueryRowContext(ctx, query, boxerID).Scan(
		&resp.FightID,
		&resp.OpponentID,
		&resp.OpponentName,
		&resp.OpponentNickname,
		&resp.ScheduledTime,
		&resp.Status,
		&resp.Rounds,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrUpcomingFightNotFound
		}
		return nil, err
	}

	return &resp, nil
}

// GetFightHistoryWithOpponents retrieves fight history for a boxer with opponent names (MAT-103).
func (s *FightStore) GetFightHistoryWithOpponents(ctx context.Context, boxerID int) ([]*db.FightHistoryWithOpponent, error) {
	query := `
		SELECT f.id, f.boxer1_id, f.boxer2_id, f.status, f.scheduled_time, f.start_time, f.end_time,
		       f.winner_id, f.round, f.data, f.created_at, f.updated_at,
		       CASE
		        WHEN f.boxer1_id = $1 THEN b2.name
		        ELSE b1.name
		       END as opponent_name
		FROM fights f
		LEFT JOIN boxers b1 ON f.boxer1_id = b1.id
		LEFT JOIN boxers b2 ON f.boxer2_id = b2.id
		WHERE f.boxer1_id = $1 OR f.boxer2_id = $1
		ORDER BY f.end_time DESC, f.created_at DESC
		LIMIT 50`

	rows, err := s.db.QueryContext(ctx, query, boxerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var fights []*db.FightHistoryWithOpponent
	for rows.Next() {
		fight := &db.FightHistoryWithOpponent{}
		fight.Data = &db.NullJSONB{Value: nil, IsValid: false, Null: true}
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
			fight.Data,
			&fight.CreatedAt,
			&fight.UpdatedAt,
			&fight.OpponentName,
		)
		if err != nil {
			return nil, err
		}
		// Copy the scanned NullJSONB value to JSONData for proper serialization
		if fight.Data.IsValid {
			fight.JSONData = fight.Data.Value
		} else {
			fight.JSONData = nil
		}
		fights = append(fights, fight)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fights, nil
}

// UpdateStatus updates the status of a fight.
func (s *FightStore) UpdateStatus(ctx context.Context, id int, status string) error {
	var query string
	var args []interface{}

	if status == "completed" {
		query = `UPDATE fights SET status = $1, end_time = CURRENT_TIMESTAMP WHERE id = $2`
		args = []interface{}{status, id}
	} else {
		query = `UPDATE fights SET status = $1 WHERE id = $2`
		args = []interface{}{status, id}
	}

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// UpdateStatusTx updates the status of a fight within a transaction.
func (s *FightStore) UpdateStatusTx(ctx context.Context, tx *sql.Tx, id int, status string) error {
	var query string
	var args []interface{}

	if status == "completed" {
		query = `UPDATE fights SET status = $1, end_time = CURRENT_TIMESTAMP WHERE id = $2`
		args = []interface{}{status, id}
	} else {
		query = `UPDATE fights SET status = $1 WHERE id = $2`
		args = []interface{}{status, id}
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// UpdateRound updates the current round of a fight.
func (s *FightStore) UpdateRound(ctx context.Context, id, round int) error {
	query := `UPDATE fights SET round = $1 WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, round, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// UpdateRoundTx updates the current round of a fight within a transaction.
func (s *FightStore) UpdateRoundTx(ctx context.Context, tx *sql.Tx, id, round int) error {
	query := `UPDATE fights SET round = $1 WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, round, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// SetWinner sets the winner of a fight and marks it as completed.
func (s *FightStore) SetWinner(ctx context.Context, id, winnerID int) error {
	query := `UPDATE fights SET winner_id = $1, status = 'completed' WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, winnerID, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// SetWinnerTx sets the winner of a fight and marks it as completed within a transaction.
func (s *FightStore) SetWinnerTx(ctx context.Context, tx *sql.Tx, id, winnerID int) error {
	query := `UPDATE fights SET winner_id = $1, status = 'completed' WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, winnerID, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// SetData updates the data field of a fight (used for storing fight simulation results).
func (s *FightStore) SetData(ctx context.Context, id int, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `UPDATE fights SET data = $1 WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, dataJSON, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// SetDataTx updates the data field of a fight within a transaction.
func (s *FightStore) SetDataTx(ctx context.Context, tx *sql.Tx, id int, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `UPDATE fights SET data = $1 WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, dataJSON, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// Delete removes a fight from the database.
func (s *FightStore) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM fights WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrFightNotFound
	}

	return nil
}

// BeginFightTx begins a new transaction for fighting operations with serializable isolation.
func (s *FightStore) BeginFightTx(ctx context.Context) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
}

// BoxerInFight checks if a boxer is currently involved in an active fight.
func (s *FightStore) BoxerInFight(ctx context.Context, boxerID int) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM fights
		WHERE (boxer1_id = $1 OR boxer2_id = $1)
		  AND status IN ('scheduled', 'in_progress')`

	var inFight bool
	err := s.db.QueryRowContext(ctx, query, boxerID).Scan(&inFight)
	return inFight, err
}

// scanFight scans a single row into a Fight model.
func (s *FightStore) scanFight(row Scanner) (*model.Fight, error) {
	var data db.NullJSONB
	fight := &model.Fight{}

	err := row.Scan(
		&fight.ID,
		&fight.Boxer1ID,
		&fight.Boxer2ID,
		&fight.Status,
		&fight.ScheduledTime,
		&fight.StartTime,
		&fight.EndTime,
		&fight.WinnerID,
		&fight.Round,
		&data,
		&fight.CreatedAt,
		&fight.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Set Data field from NullJSONB if valid
	if data.IsValid {
		fight.Data = data.Value
	}

	return fight, nil
}

// scanFights scans multiple rows into a slice of Fight models.
func (s *FightStore) scanFights(rows *sql.Rows) ([]*model.Fight, error) {
	var fights []*model.Fight
	for rows.Next() {
		fight, err := s.scanFight(rows)
		if err != nil {
			return nil, err
		}
		fights = append(fights, fight)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fights, nil
}

// scanFightTx scans a single row from a transaction query into a Fight model.
func (s *FightStore) scanFightTx(_ *sql.Tx, row Scanner) (*model.Fight, error) {
	var data db.NullJSONB
	fight := &model.Fight{}

	err := row.Scan(
		&fight.ID,
		&fight.Boxer1ID,
		&fight.Boxer2ID,
		&fight.Status,
		&fight.ScheduledTime,
		&fight.StartTime,
		&fight.EndTime,
		&fight.WinnerID,
		&fight.Round,
		&data,
		&fight.CreatedAt,
		&fight.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Set Data field from NullJSONB if valid
	if data.IsValid {
		fight.Data = data.Value
	}

	return fight, nil
}

// Scanner is an interface for types that can scan database rows.
type Scanner interface {
	Scan(dest ...interface{}) error
}
