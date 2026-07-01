package events

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTrainingComplete EventType = "training_complete"
	EventFightComplete    EventType = "fight_complete"
	EventFightResult      EventType = "fight_result"
	EventWorldTick        EventType = "world_tick"
)

type ScheduledEvent struct {
	ID        int
	BoxerID   int
	EventType EventType
	EventTime time.Time
	Data      map[string]interface{}
}

func NewScheduledEvent(
	boxerID int,
	eventType EventType,
	eventTime time.Time,
	data map[string]interface{},
) *ScheduledEvent {
	return &ScheduledEvent{
		BoxerID:   boxerID,
		EventType: eventType,
		EventTime: eventTime,
		Data:      data,
	}
}

func (e *ScheduledEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *ScheduledEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}
