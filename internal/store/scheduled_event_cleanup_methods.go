package store

import (
	"context"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// OldFailedEvent represents a failed event with associated fight data for cleanup (MAT-143).
type OldFailedEvent struct {
	EventID      int
	EventType    model.EventType
	EventTime    time.Time
	ProcessedAt  time.Time
	ErrorMessage *string
	FightID      *int
	FightStatus  *string
	Boxer1ID     *int
	Boxer2ID     *int
	BoxerID      int
	HoursOld     float64
}

// GetFailedEventsOlderThan retrieves failed events older than the specified threshold.
// For fight_simulate events, it joins with fights table to check fight status.
// This is used by the automatic cleanup job (MAT-143).
func (s *ScheduledEventStore) GetFailedEventsOlderThan(ctx context.Context, hoursOld int) ([]*OldFailedEvent, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT
			se.id as event_id,
			se.event_type,
			se.event_time,
			se.processed_at,
			se.error_message,
			se.boxer_id,
			f.id as fight_id,
			f.status as fight_status,
			f.boxer1_id,
			f.boxer2_id,
			EXTRACT(epoch FROM CURRENT_TIMESTAMP - se.processed_at) / 3600 as hours_old
		FROM scheduled_events se
		LEFT JOIN fights f ON f.id = (se.event_data->>'fight_id')::int
		WHERE se.status = 'failed'
			AND se.processed_at <= CURRENT_TIMESTAMP - ($1 * INTERVAL '1 hour')
		ORDER BY se.processed_at ASC`

	rows, err := s.db.QueryContext(ctx, query, hoursOld)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []*OldFailedEvent
	for rows.Next() {
		event := &OldFailedEvent{}
		err := rows.Scan(
			&event.EventID, &event.EventType, &event.EventTime, &event.ProcessedAt,
			&event.ErrorMessage, &event.BoxerID, &event.FightID, &event.FightStatus,
			&event.Boxer1ID, &event.Boxer2ID, &event.HoursOld)
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

// MarkAsCompletedWithReason marks an event as completed with a custom reason message.
// This is used by the cleanup job to record why an event was auto-resolved (MAT-143).
func (s *ScheduledEventStore) MarkAsCompletedWithReason(ctx context.Context, eventId int, reason string) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		UPDATE scheduled_events
		SET processed = TRUE, status = 'completed', processed_at = CURRENT_TIMESTAMP,
			error_message = COALESCE(error_message, '') || ' | ' || $2
		WHERE id = $1 AND status = 'failed'`

	result, err := s.db.ExecContext(ctx, query, eventId, reason)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrAlreadyProcessed
	}

	return nil
}

// GetFailedEventsCount returns the count of failed events for monitoring purposes (MAT-143).
func (s *ScheduledEventStore) GetFailedEventsCount(ctx context.Context) (int, error) {
	if s.db == nil {
		return 0, errors.New("database connection is nil")
	}

	var count int
	query := `SELECT COUNT(*) FROM scheduled_events WHERE status = 'failed'`
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// GetOldFailedEventsCount returns the count of failed events older than the threshold (MAT-143).
func (s *ScheduledEventStore) GetOldFailedEventsCount(ctx context.Context, hoursOld int) (int, error) {
	if s.db == nil {
		return 0, errors.New("database connection is nil")
	}

	var count int
	query := `
		SELECT COUNT(*) FROM scheduled_events
		WHERE status = 'failed'
			AND processed_at <= CURRENT_TIMESTAMP - ($1 * INTERVAL '1 hour')`

	err := s.db.QueryRowContext(ctx, query, hoursOld).Scan(&count)
	return count, err
}

// GetScheduledFightsWithFailedEvents returns fights that are still scheduled but have failed events (MAT-143).
func (s *ScheduledEventStore) GetScheduledFightsWithFailedEvents(ctx context.Context) ([]*struct {
	FightID     int
	Boxer1ID    int
	Boxer2ID    int
	ScheduledAt time.Time
	EventID     int
	Error       string
}, error,
) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT
			f.id as fight_id,
			f.boxer1_id,
			f.boxer2_id,
			f.scheduled_time,
			se.id as event_id,
			COALESCE(se.error_message, 'no error') as error
		FROM scheduled_events se
		JOIN fights f ON f.id = (se.event_data->>'fight_id')::int
		WHERE se.status = 'failed'
			AND se.event_type = 'fight_simulate'
			AND f.status = 'scheduled'`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*struct {
		FightID     int
		Boxer1ID    int
		Boxer2ID    int
		ScheduledAt time.Time
		EventID     int
		Error       string
	}
	for rows.Next() {
		r := &struct {
			FightID     int
			Boxer1ID    int
			Boxer2ID    int
			ScheduledAt time.Time
			EventID     int
			Error       string
		}{}
		err := rows.Scan(&r.FightID, &r.Boxer1ID, &r.Boxer2ID, &r.ScheduledAt, &r.EventID, &r.Error)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
