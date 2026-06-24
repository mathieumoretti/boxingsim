package model

import "time"

// EventType represents the type of scheduled event
type EventType string

const (
	EventTypeTraining  EventType = "training"
	EventTypeRest      EventType = "rest"
	EventTypeCompetition EventType = "competition"
	EventTypeOther     EventType = "other"
)

// ScheduledEvent represents a scheduled event for a boxer
type ScheduledEvent struct {
	ID        int            `json:"id"`
	BoxerID   int            `json:"boxer_id"`
	EventType EventType     `json:"event_type"`
	EventTime time.Time      `json:"event_time"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

// ScheduledEventCreate represents a request to create a new scheduled event
type ScheduledEventCreate struct {
	BoxerID   int            `json:"boxer_id" binding:"required"`
	EventType EventType     `json:"event_type" binding:"required"`
	EventTime time.Time      `json:"event_time" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

// ScheduledEventUpdate represents a request to update a scheduled event
type ScheduledEventUpdate struct {
	EventType *EventType     `json:"event_type"`
	EventTime *time.Time     `json:"event_time"`
	Data      *map[string]interface{} `json:"data"`
}

// ScheduledEventResponse represents a scheduled event for API responses
type ScheduledEventResponse struct {
	ID        int            `json:"id"`
	BoxerID   int            `json:"boxer_id"`
	EventType EventType     `json:"event_type"`
	EventTime time.Time      `json:"event_time"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}