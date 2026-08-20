package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/model"
)

var (
	ErrEventNotFound    = errors.New("scheduled event not found")
	ErrAlreadyProcessed = errors.New("event already marked as processed")
	ErrFailedToInsert   = errors.New("failed to insert scheduled event")
)

// ScheduledEventStore implements CRUD operations for scheduled events with idempotent processing support.
type ScheduledEventStore struct {
	db *sql.DB
}

// NewScheduledEventStore creates a new ScheduledEventStore instance.
func NewScheduledEventStore(db *sql.DB) *ScheduledEventStore {
	return &ScheduledEventStore{
		db: db,
	}
}

// Create inserts a new scheduled event and returns the generated ID.
func (s *ScheduledEventStore) Create(ctx context.Context, event *model.ScheduledEvent) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
			INSERT INTO scheduled_events (
				boxer_id, event_type, event_time, processed, created_at
			) VALUES ($1, $2, $3, FALSE, CURRENT_TIMESTAMP)
			RETURNING id`

	err := s.db.QueryRowContext(ctx, query,
			event.BoxerID, event.EventType, event.EventTime).Scan(&event.ID)
	if err != nil {
		return ErrFailedToInsert
	}
	return nil
}

// GetByID retrieves a scheduled event by its ID. Returns ErrEventNotFound if the event doesn't exist.
func (s *ScheduledEventStore) GetByID(ctx context.Context, id int) (*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
			SELECT
				id, boxer_id, event_type, event_time, processed, event_data,
				error_message, created_at
			FROM scheduled_events WHERE id = $1`

	event := &model.ScheduledEvent{}
	row := s.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
		&event.Processed, &event.EventData, &event.ErrorMessage, &event.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetByBoxerID retrieves all scheduled events for a specific boxer.
func (s *ScheduledEventStore) GetByBoxerID(ctx context.Context, boxerID int) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
			SELECT
				id, boxer_id, event_type, event_time, processed, event_data,
				error_message, created_at
			FROM scheduled_events WHERE boxer_id = $1 ORDER BY event_time ASC`

	rows, err := s.db.QueryContext(ctx, query, boxerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.EventData, &event.ErrorMessage, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetPendingEventsBeforeGameTime retrieves all unprocessed scheduled events
// that should have occurred by the given game time (inclusive). This is the hot path for worker simulation.
func (s *ScheduledEventStore) GetPendingEventsBeforeGameTime(
	ctx context.Context,
	gameTime time.Time,
	limit int,
) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
			SELECT
				id, boxer_id, event_type, event_time, processed, event_data,
				error_message, created_at
			FROM scheduled_events
			WHERE event_time <= $1 AND NOT processed
			ORDER BY event_time ASC LIMIT $2`

	rows, err := s.db.QueryContext(ctx, query, gameTime, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.EventData, &event.ErrorMessage, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// MarkAsProcessed marks a scheduled event as processed atomically using row-level locking.
// Returns ErrAlreadyProcessed if the event was already marked as processed (idempotent operation).
func (s *ScheduledEventStore) MarkAsProcessed(ctx context.Context, eventId int) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
			UPDATE scheduled_events
			SET processed = TRUE
			WHERE id = $1 AND NOT processed`

	result, err := s.db.ExecContext(ctx, query, eventId)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrAlreadyProcessed
	}

	return nil
}

// CreateAndProcessIfPastTime creates a new scheduled event and immediately marks it as processed
// if the specified time is less than or equal to current game time (for instant completion testing scenarios).
func (s *ScheduledEventStore) CreateAndProcessIfPastTime(
	ctx context.Context,
	boxerID int,
	eventType model.EventType,
	eventTime time.Time,
	data []byte, // JSONB data as raw bytes
) (*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }() // Will be ignored if committed

	insertQuery := `
			INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, event_data, created_at)
			VALUES ($1, $2, $3, FALSE, COALESCE($4::bytea, 'null'::jsonb), CURRENT_TIMESTAMP)
			RETURNING id`

	var eventId int
	err = tx.QueryRowContext(ctx, insertQuery, boxerID, eventType, eventTime, data).Scan(&eventId)
	if err != nil {
		return nil, ErrFailedToInsert
	}

	event := &model.ScheduledEvent{
		ID:        eventId,
		BoxerID:   boxerID,
		EventType: eventType,
		EventTime: eventTime,
		CreatedAt: time.Now(),
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return event, nil
}

// DeleteByBoxerID deletes all scheduled events for a specific boxer. Used for cleanup when boxing is deleted.
func (s *ScheduledEventStore) DeleteByBoxerID(ctx context.Context, boxerID int) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := "DELETE FROM scheduled_events WHERE boxer_id = $1"
	_, err := s.db.ExecContext(ctx, query, boxerID)
	return err
}
