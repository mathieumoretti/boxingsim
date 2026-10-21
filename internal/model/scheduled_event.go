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

func (e *EventData) Scan(value any) error {
	if value == nil {
		*e = EventData([]byte("{}")) // Empty object instead of nil
		return nil
	}

	// Handle string representation from PostgreSQL JSONB
	if str, ok := value.(string); ok {
		if str == "null" || str == "" {
			*e = EventData([]byte("{}")) // Empty object instead of nil
			return nil
		}
		// Validate it's actually valid JSON before accepting
		var temp any
		if err := json.Unmarshal([]byte(str), &temp); err != nil {
			// If string is not valid JSON, treat as empty (corrupted data protection)
			// This handles cases where the database has invalid/corrupted event_data
			*e = EventData([]byte("{}"))
			return nil // Don't fail - just use empty data
		}
		*e = EventData(append([]byte(nil), []byte(str)...)) // Defensive copy to prevent lib/pq buffer reuse
		return nil
	}

	// Handle []byte from PostgreSQL JSONB (most common)
	if b, ok := value.([]byte); ok {
		if len(b) == 0 {
			*e = EventData([]byte("{}")) // Empty bytes treated as empty object
			return nil
		}
		// Validate bytes are valid JSON before accepting
		var temp any
		if err := json.Unmarshal(b, &temp); err != nil {
			// Invalid JSON - treat as empty (corrupted data protection)
			*e = EventData([]byte("{}"))
			return nil // Don't fail - just use empty data
		}
		*e = EventData(append([]byte(nil), b...)) // Defensive copy to prevent lib/pq buffer reuse
		return nil
	}

	// Handle json.RawMessage directly (some drivers return this)
	if raw, ok := value.(json.RawMessage); ok {
		if len(raw) == 0 {
			*e = EventData([]byte("{}"))
			return nil
		}
		*e = EventData(raw)
		return nil
	}

	// Fallback: try to convert any other type to bytes
	b, ok := value.([]byte)
	if !ok {
		// Convert unknown types to string representation for debugging
		typeName := fmt.Sprintf("%T", value)
		return fmt.Errorf("failed to scan EventData: unsupported type %s (value: %#v)", typeName, value)
	}
	if len(b) == 0 {
		*e = EventData([]byte("{}"))
		return nil
	}
	var temp any
	if err := json.Unmarshal(b, &temp); err != nil {
		*e = EventData([]byte("{}"))
		return nil
	}
	*e = EventData(append([]byte(nil), b...)) // Defensive copy to prevent lib/pq buffer reuse
	return nil
}

// Value implements driver.Valuer for proper INSERT/UPDATE handling.
func (e EventData) Value() (driver.Value, error) {
	if len(e) == 0 {
		return nil, nil // Return NULL for empty data
	}
	return []byte(e), nil
}

// ToMap converts EventData to a map[string]any for API responses.
func (e *EventData) ToMap() map[string]any {
	if e == nil || len(*e) == 0 {
		return nil
	}
	var m map[string]any
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
	BoxerID   int            `json:"boxer_id" binding:"required"`
	EventType EventType      `json:"event_type" binding:"required"`
	EventTime time.Time      `json:"event_time" binding:"required"`
	Data      map[string]any `json:"data"`
}

// ScheduledEventUpdate represents a request to update an existing scheduled event.
type ScheduledEventUpdate struct {
	EventType *EventType      `json:"event_type,omitempty"`
	EventTime *time.Time      `json:"event_time,omitempty"`
	Data      *map[string]any `json:"data,omitempty"`
}

// ScheduledEventResponse represents a scheduled event for API responses.
type ScheduledEventResponse struct {
	ID        int            `json:"id"`
	BoxerID   int            `json:"boxer_id"`
	EventType string         `json:"event_type"`
	EventTime time.Time      `json:"event_time"`
	Data      map[string]any `json:"data"`

	CreatedAt time.Time `json:"created_at"`
}
