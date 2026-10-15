package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// EventType represents the type of scheduled event.
type EventType string

const (
	EventTypeTraining      EventType = "training"
	EventTypeRest          EventType = "rest"
	EventTypeCompetition   EventType = "competition"
	EventTypeFightSimulate EventType = "fight_simulate"
	EventTypeOther         EventType = "other"
)

// EventData wraps json.RawMessage for proper null handling from JSONB.
type EventData json.RawMessage

func (e *EventData) Scan(value interface{}) error {
	if value == nil {
		*e = EventData([]byte("{}")) // Empty object instead of nil
		return nil
	}

	// Handle string representation of null from PostgreSQL jsonb NULLIF
	if str, ok := value.(string); ok {
		if str == "null" || str == "" {
			*e = EventData([]byte("{}")) // Empty object instead of nil
			return nil
		}
		*e = EventData([]byte(str))
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		// Log what type we actually received for debugging
		return fmt.Errorf("failed to scan []byte: received type %T with value %#v", value, value)
	}
	*e = EventData(b)
	return nil
}

// Value implements driver.Valuer for proper INSERT/UPDATE handling.
func (e EventData) Value() (driver.Value, error) {
	if len(e) == 0 {
		return nil, nil // Return NULL for empty data
	}
	return []byte(e), nil
}

// ToMap converts EventData to a map[string]interface{} for API responses.
func (e *EventData) ToMap() map[string]interface{} {
	if e == nil || len(*e) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(*e, &m); err != nil {
		return nil
	}
	return m
}

// EventStatus represents the lifecycle state of a scheduled event.
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"    // Awaiting processing by worker
	EventStatusProcessing EventStatus = "processing" // Currently being handled
	EventStatusCompleted  EventStatus = "completed"  // Successfully processed
	EventStatusFailed     EventStatus = "failed"     // Processing error occurred
)

// ScheduledEvent represents a scheduled event for a boxer with idempotent processing support.
type ScheduledEvent struct {
	ID           int         `db:"id" json:"id"`
	BoxerID      int         `db:"boxer_id" json:"boxer_id"`
	EventType    EventType   `db:"event_type" json:"event_type"`
	EventTime    time.Time   `db:"event_time" json:"event_time"`
	Processed    bool        `db:"processed" json:"processed"`
	Status       EventStatus `db:"status" json:"status"`
	EventData    EventData   `db:"event_data" dbtype:"jsonb" json:"-"`
	ErrorMessage *string     `db:"error_message" json:"error_message,omitempty"`
	ProcessedAt  *time.Time  `db:"processed_at" json:"processed_at,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// ScheduledEventCreate represents a request to create a new scheduled event.
type ScheduledEventCreate struct {
	BoxerID   int                    `json:"boxer_id" binding:"required"`
	EventType EventType              `json:"event_type" binding:"required"`
	EventTime time.Time              `json:"event_time" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

// ScheduledEventUpdate represents a request to update an existing scheduled event.
type ScheduledEventUpdate struct {
	EventType *EventType              `json:"event_type,omitempty"`
	EventTime *time.Time              `json:"event_time,omitempty"`
	Data      *map[string]interface{} `json:"data,omitempty"`
}

// ScheduledEventResponse represents a scheduled event for API responses.
type ScheduledEventResponse struct {
	ID        int                    `json:"id"`
	BoxerID   int                    `json:"boxer_id"`
	EventType string                 `json:"event_type"`
	EventTime time.Time              `json:"event_time"`
	Data      map[string]interface{} `json:"data"`

	CreatedAt time.Time `json:"created_at"`
}
