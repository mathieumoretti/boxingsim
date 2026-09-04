package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrTrainingSessionNotFound   = errors.New("training session not found")
	ErrTrainingTypeNotFound      = errors.New("training type not found")
	ErrTrainingSessionNotPending = errors.New("training session is not in pending status")
)

// TrainingSessionStore implements CRUD operations for training sessions
type TrainingSessionStore struct {
	db *sql.DB
}

// NewTrainingSessionStore creates a new TrainingSessionStore instance
func NewTrainingSessionStore(db *sql.DB) *TrainingSessionStore {
	return &TrainingSessionStore{db: db}
}

// TrainingTypeStore implements CRUD operations for training types (reference data)
type TrainingTypeStore struct {
	db *sql.DB
}

// NewTrainingTypeStore creates a new TrainingTypeStore instance
func NewTrainingTypeStore(db *sql.DB) *TrainingTypeStore {
	return &TrainingTypeStore{db: db}
}

// ==================== TrainingTypeStore Methods ====================

// GetAll retrieves all training types
func (s *TrainingTypeStore) GetAll(ctx context.Context) ([]*model.TrainingType, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT id, name, description, strength_gain_factor, defense_gain_factor,
		       agility_gain_factor, energy_cost, created_at, updated_at
		FROM training_types
		ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var types []*model.TrainingType
	for rows.Next() {
		t := &model.TrainingType{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.StrengthGainFactor, &t.DefenseGainFactor,
			&t.AgilityGainFactor, &t.EnergyCost, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return types, nil
}

// GetByID retrieves a training type by ID
func (s *TrainingTypeStore) GetByID(ctx context.Context, id int) (*model.TrainingType, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT id, name, description, strength_gain_factor, defense_gain_factor,
		       agility_gain_factor, energy_cost, created_at, updated_at
		FROM training_types WHERE id = $1`

	t := &model.TrainingType{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Description, &t.StrengthGainFactor, &t.DefenseGainFactor,
		&t.AgilityGainFactor, &t.EnergyCost, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrTrainingTypeNotFound
	}
	if err != nil {
		return nil, err
	}

	return t, nil
}

// GetName retrieves a training type by name
func (s *TrainingTypeStore) GetByName(ctx context.Context, name string) (*model.TrainingType, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT id, name, description, strength_gain_factor, defense_gain_factor,
		       agility_gain_factor, energy_cost, created_at, updated_at
		FROM training_types WHERE name = $1`

	t := &model.TrainingType{}
	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&t.ID, &t.Name, &t.Description, &t.StrengthGainFactor, &t.DefenseGainFactor,
		&t.AgilityGainFactor, &t.EnergyCost, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrTrainingTypeNotFound
	}
	if err != nil {
		return nil, err
	}

	return t, nil
}

// ==================== TrainingSessionStore Methods ====================

// GetByID retrieves a training session by ID with optional joins
func (s *TrainingSessionStore) GetByID(ctx context.Context, id int) (*model.TrainingSession, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT ts.id, ts.boxer_id, ts.training_type_id, ts.scheduled_event_id,
		       ts.duration_hours, ts.planned_strength_gain, ts.planned_defense_gain,
		       ts.planned_agility_gain, ts.status, ts.completed_at,
		       ts.created_at, ts.updated_at
		FROM training_sessions ts
		WHERE ts.id = $1`

	session := &model.TrainingSession{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.BoxerID, &session.TrainingTypeID, &session.ScheduledEventID,
		&session.DurationHours, &session.PlannedStrengthGain, &session.PlannedDefenseGain,
		&session.PlannedAgilityGain, &session.Status, &session.CompletedAt,
		&session.CreatedAt, &session.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrTrainingSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetByIDWithType retrieves a training session by ID with joined training type data
func (s *TrainingSessionStore) GetByIDWithType(ctx context.Context, id int) (*model.TrainingSession, error) {
	session, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	trainingType, err := (&TrainingTypeStore{db: s.db}).GetByID(ctx, session.TrainingTypeID)
	if err != nil {
		return nil, err
	}

	session.TrainingType = trainingType
	return session, nil
}

// GetPendingByBoxerID retrieves all pending training sessions for a boxer
func (s *TrainingSessionStore) GetPendingByBoxerID(ctx context.Context, boxerID int) ([]*model.TrainingSession, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT ts.id, ts.boxer_id, ts.training_type_id, ts.scheduled_event_id,
		       ts.duration_hours, ts.planned_strength_gain, ts.planned_defense_gain,
		       ts.planned_agility_gain, ts.status, ts.completed_at,
		       ts.created_at, ts.updated_at
		FROM training_sessions ts
		WHERE ts.boxer_id = $1 AND ts.status = 'pending'
		ORDER BY ts.created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, boxerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var sessions []*model.TrainingSession
	for rows.Next() {
		session := &model.TrainingSession{}
		err := rows.Scan(
			&session.ID, &session.BoxerID, &session.TrainingTypeID, &session.ScheduledEventID,
			&session.DurationHours, &session.PlannedStrengthGain, &session.PlannedDefenseGain,
			&session.PlannedAgilityGain, &session.Status, &session.CompletedAt,
			&session.CreatedAt, &session.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetDueBeforeGameTime retrieves training sessions that should be processed by the given game time.
// This is the hot path for worker simulation - it joins with scheduled_events to find sessions due.
func (s *TrainingSessionStore) GetDueBeforeGameTime(
	ctx context.Context,
	gameTime time.Time,
	limit int,
) ([]*model.TrainingSession, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT DISTINCT ts.id, ts.boxer_id, ts.training_type_id, ts.scheduled_event_id,
		       ts.duration_hours, ts.planned_strength_gain, ts.planned_defense_gain,
		       ts.planned_agility_gain, ts.status, ts.completed_at,
		       ts.created_at, ts.updated_at
		FROM training_sessions ts
		JOIN scheduled_events se ON ts.scheduled_event_id = se.id
		WHERE ts.status = 'pending'
		  AND se.event_time <= $1
		  AND NOT se.processed
		ORDER BY se.event_time ASC
		LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, gameTime, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var sessions []*model.TrainingSession
	for rows.Next() {
		session := &model.TrainingSession{}
		err := rows.Scan(
			&session.ID, &session.BoxerID, &session.TrainingTypeID, &session.ScheduledEventID,
			&session.DurationHours, &session.PlannedStrengthGain, &session.PlannedDefenseGain,
			&session.PlannedAgilityGain, &session.Status, &session.CompletedAt,
			&session.CreatedAt, &session.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// Create inserts a new training session and returns the generated ID
func (s *TrainingSessionStore) Create(ctx context.Context, session *model.TrainingSession) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		INSERT INTO training_sessions (
			boxer_id, training_type_id, scheduled_event_id,
			duration_hours, planned_strength_gain, planned_defense_gain,
			planned_agility_gain, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9)
		RETURNING id`

	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now
	session.Status = model.TrainingSessionPending

	err := s.db.QueryRowContext(ctx, query,
		session.BoxerID, session.TrainingTypeID, session.ScheduledEventID,
		session.DurationHours, session.PlannedStrengthGain, session.PlannedDefenseGain,
		session.PlannedAgilityGain, now, now).Scan(&session.ID)

	return err
}

// UpdateStatus updates the status of a training session (e.g., to completed or cancelled)
func (s *TrainingSessionStore) UpdateStatus(ctx context.Context, id int, status model.TrainingSessionStatus) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	// First check if the session exists and is pending (only pending sessions can be updated)
	query := `
		SELECT status FROM training_sessions WHERE id = $1`

	var currentStatus string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&currentStatus)
	if err == sql.ErrNoRows {
		return ErrTrainingSessionNotFound
	}
	if err != nil {
		return err
	}

	// If completing a session, set completed_at timestamp
	var completedAt *time.Time
	if status == model.TrainingSessionCompleted {
		now := time.Now()
		completedAt = &now
	}

	// Update the status
	updateQuery := `
		UPDATE training_sessions
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'pending'`

	if completedAt != nil {
		updateQuery = `
			UPDATE training_sessions
			SET status = $2, completed_at = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND status = 'pending'`

		_, err = s.db.ExecContext(ctx, updateQuery, id, status, completedAt)
	} else {
		_, err = s.db.ExecContext(ctx, updateQuery, id, status)
	}

	return err
}

// MarkAsCompleted marks a training session as completed with the current timestamp
func (s *TrainingSessionStore) MarkAsCompleted(ctx context.Context, id int) error {
	return s.UpdateStatus(ctx, id, model.TrainingSessionCompleted)
}

// MarkAsCancelled marks a training session as cancelled
func (s *TrainingSessionStore) MarkAsCancelled(ctx context.Context, id int) error {
	return s.UpdateStatus(ctx, id, model.TrainingSessionCancelled)
}
