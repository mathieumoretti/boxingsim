package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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
						boxer_id, event_type, event_time, processed, status, event_data, created_at
					) VALUES ($1, $2, $3, $4, 'pending', NULLIF($5::jsonb, '{}'::jsonb), CURRENT_TIMESTAMP)
					RETURNING id`

	err := s.db.QueryRowContext(ctx, query,
		event.BoxerID, event.EventType, event.EventTime, event.Processed, event.EventData).Scan(&event.ID)
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
						id, boxer_id, event_type, event_time, processed, status, event_data,
						error_message, processed_at, retry_count, created_at
					FROM scheduled_events WHERE id = $1`

	event := &model.ScheduledEvent{}
	row := s.db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
		&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
		&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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
						id, boxer_id, event_type, event_time, processed, status, event_data,
						error_message, processed_at, retry_count, created_at
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
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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
						id, boxer_id, event_type, event_time, processed, status, event_data,
						error_message, processed_at, retry_count, created_at
					FROM scheduled_events
					WHERE event_time <= $1 AND processed = FALSE AND status IN ('pending', 'processing')
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
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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

// MarkAsProcessing marks an event as currently being processed (for crash recovery tracking).
// Returns ErrAlreadyProcessed if the event was already processed or failed.
func (s *ScheduledEventStore) MarkAsProcessing(ctx context.Context, eventId int) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
					UPDATE scheduled_events
					SET status = 'processing', processed_at = NULL
					WHERE id = $1 AND status IN ('pending', 'processing')`

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

// MarkAsCompleted marks a scheduled event as successfully processed atomically.
// Returns ErrAlreadyProcessed if the event was already processed or failed.
func (s *ScheduledEventStore) MarkAsCompleted(ctx context.Context, eventId int) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
					UPDATE scheduled_events
					SET processed = TRUE, status = 'completed', processed_at = CURRENT_TIMESTAMP, error_message = NULL
					WHERE id = $1 AND status IN ('pending', 'processing')`

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

// MarkAsFailed marks a scheduled event as failed with an error message.
// This allows for debugging and potential retry mechanisms.
func (s *ScheduledEventStore) MarkAsFailed(ctx context.Context, eventId int, errorMessage string) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
					UPDATE scheduled_events
					SET status = 'failed', processed_at = CURRENT_TIMESTAMP, error_message = $2
					WHERE id = $1 AND status IN ('pending', 'processing')`

	result, err := s.db.ExecContext(ctx, query, eventId, errorMessage)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrAlreadyProcessed
	}

	return nil
}

// MarkAsProcessed marks a scheduled event as processed atomically using row-level locking.
// Returns ErrAlreadyProcessed if the event was already marked as processed (idempotent operation).
// Deprecated: Use MarkAsCompleted or MarkAsFailed instead for better visibility.
func (s *ScheduledEventStore) MarkAsProcessed(ctx context.Context, eventId int) error {
	return s.MarkAsCompleted(ctx, eventId)
}

// GetFailedEvents retrieves all events that have failed processing for debugging/monitoring.
func (s *ScheduledEventStore) GetFailedEvents(ctx context.Context, limit int) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
					SELECT
						id, boxer_id, event_type, event_time, processed, status, event_data,
						error_message, processed_at, retry_count, created_at
					FROM scheduled_events
					WHERE status = 'failed'
					ORDER BY processed_at DESC
					LIMIT $1`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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

// GetProcessingEvents retrieves all events currently in 'processing' state (for crash recovery).
func (s *ScheduledEventStore) GetProcessingEvents(ctx context.Context) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
					SELECT
						id, boxer_id, event_type, event_time, processed, status, event_data,
						error_message, processed_at, retry_count, created_at
					FROM scheduled_events
					WHERE status = 'processing'
					ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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

// ResetToPending resets an event from 'processing' state back to 'pending' for crash recovery.
func (s *ScheduledEventStore) ResetToPending(ctx context.Context, eventId int) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `UPDATE scheduled_events SET status = 'pending', processed_at = NULL WHERE id = $1 AND status = 'processing'`

	result, err := s.db.ExecContext(ctx, query, eventId)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrEventNotFound
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
					INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data, created_at)
					VALUES ($1, $2, $3, FALSE, 'pending', COALESCE($4::bytea, 'null'::jsonb), CURRENT_TIMESTAMP)
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
		Status:    model.EventStatusPending,
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

// GetPendingByBoxerIDAndType returns pending unprocessed events for a boxer filtered by event type.
// This is used to check if a boxer has active rest periods before allowing new training sessions.
func (s *ScheduledEventStore) GetPendingByBoxerIDAndType(ctx context.Context, boxerID int, eventType model.EventType) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
						SELECT
							id, boxer_id, event_type, event_time, processed, status, event_data,
							error_message, processed_at, retry_count, created_at
						FROM scheduled_events
						WHERE boxer_id = $1
						AND event_type = $2
						AND processed = FALSE
						AND status IN ('pending', 'processing')
						ORDER BY event_time ASC`

	rows, err := s.db.QueryContext(ctx, query, boxerID, eventType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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

// GetByFightID retrieves scheduled events for a specific fight (for cancellation cleanup).
func (s *ScheduledEventStore) GetByFightID(ctx context.Context, fightID int) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
						SELECT
							id, boxer_id, event_type, event_time, processed, status, event_data,
							error_message, processed_at, retry_count, created_at
						FROM scheduled_events
						WHERE (event_data->>'fight_id')::int = $1
						AND processed = FALSE
						AND status IN ('pending', 'processing')
						ORDER BY event_time ASC`

	rows, err := s.db.QueryContext(ctx, query, fightID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
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

// GetFailedEventsEligibleForRetry retrieves failed events that are eligible for retry.
// An event is eligible if:
// 1. It has a retry_count less than maxRetries
// 2. The configured delay period (exponential backoff) has passed since last failure
// 3. The error message does not indicate a permanent failure
func (s *ScheduledEventStore) GetFailedEventsEligibleForRetry(
	ctx context.Context,
	maxRetries int,
	baseDelaySeconds int,
	maxDelaySeconds int,
	currentTime time.Time,
) ([]*model.ScheduledEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
					SELECT
						id, boxer_id, event_type, event_time, processed, status, event_data,
						error_message, processed_at, retry_count, created_at
					FROM scheduled_events
					WHERE status = 'failed'
					AND retry_count < $1
					AND processed_at IS NOT NULL
					ORDER BY processed_at ASC`

	rows, err := s.db.QueryContext(ctx, query, maxRetries)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var allEvents []*model.ScheduledEvent
	for rows.Next() {
		event := &model.ScheduledEvent{}
		err := rows.Scan(
			&event.ID, &event.BoxerID, &event.EventType, &event.EventTime,
			&event.Processed, &event.Status, &event.EventData, &event.ErrorMessage,
			&event.ProcessedAt, &event.RetryCount, &event.CreatedAt)
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Filter events based on backoff delay and permanent errors
	var eligibleEvents []*model.ScheduledEvent
	for _, event := range allEvents {
		// Skip if error indicates permanent failure
		if isPermanentFailure(event.ErrorMessage) {
			continue
		}

		// Calculate required delay with exponential backoff
		delaySeconds := calculateBackoffDelay(event.RetryCount, baseDelaySeconds, maxDelaySeconds)

		// Check if enough time has passed since last attempt
		if event.ProcessedAt != nil {
			timeSinceFailure := currentTime.Sub(*event.ProcessedAt)
			if timeSinceFailure.Seconds() < float64(delaySeconds) {
				continue // Not yet eligible for retry
			}
		}

		eligibleEvents = append(eligibleEvents, event)
	}

	return eligibleEvents, nil
}

// isPermanentFailure checks if an error message indicates a permanent (non-recoverable) failure.
// These errors should not be retried as they will fail again.
func isPermanentFailure(errorMessage *string) bool {
	if errorMessage == nil {
		return false
	}

	errLower := strings.ToLower(*errorMessage)
	permanentKeywords := []string{
		"boxer not found",
		"fight not found",
		"constraint violation",
		"violates check constraint",
		"unique constraint",
		"foreign key",
		"does not exist",
	}

	for _, keyword := range permanentKeywords {
		if strings.Contains(errLower, keyword) {
			return true
		}
	}
	return false
}

// calculateBackoffDelay calculates the delay in seconds using exponential backoff.
// Formula: min(baseDelay * 2^attempt, maxDelay)
func calculateBackoffDelay(attempt int, baseDelaySeconds int, maxDelaySeconds int) int {
	if attempt <= 0 || baseDelaySeconds <= 0 {
		return baseDelaySeconds
	}

	// Calculate exponential backoff: baseDelay * 2^attempt
	delay := baseDelaySeconds << attempt // Efficient left shift for power of 2

	// Cap at max delay
	if delay > maxDelaySeconds {
		delay = maxDelaySeconds
	}

	return delay
}

// IncrementRetryCount increments the retry count for a failed event and resets it to pending.
// This should be called before reprocessing a failed event.
func (s *ScheduledEventStore) IncrementRetryCount(ctx context.Context, eventId int) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
					UPDATE scheduled_events
					SET retry_count = retry_count + 1, status = 'pending', processed_at = NULL, error_message = NULL
					WHERE id = $1 AND status = 'failed'`

	result, err := s.db.ExecContext(ctx, query, eventId)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}

// MarkAsPermanentlyFailed marks an event as permanently failed and increments retry_count.
// This is used for events that should not be retried (e.g., boxer not found).
func (s *ScheduledEventStore) MarkAsPermanentlyFailed(ctx context.Context, eventId int, reason string) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
					UPDATE scheduled_events
					SET retry_count = retry_count + 1, processed_at = CURRENT_TIMESTAMP, error_message = $2
					WHERE id = $1 AND status = 'failed'`

	result, err := s.db.ExecContext(ctx, query, eventId, reason)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrEventNotFound
	}

	return nil
}
