package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduledEvent_NewScheduledEvent(t *testing.T) {
	// Test creating a new scheduled event
	eventType := EventTrainingComplete
	// Use UTC timezone to ensure consistent behavior across environments
	eventTime := time.Now().UTC()
	data := map[string]interface{}{
		"boxer_id": 1,
		"duration": 30,
	}

	scheduledEvent := NewScheduledEvent(1, eventType, eventTime, data)

	assert.Equal(t, 1, scheduledEvent.BoxerID)
	assert.Equal(t, eventType, scheduledEvent.EventType)
	assert.Equal(t, eventTime, scheduledEvent.EventTime)
	assert.Equal(t, data, scheduledEvent.Data)
}

func TestScheduledEvent_ToJSON(t *testing.T) {
	// Test JSON serialization
	eventType := EventFightComplete
	// Use UTC timezone to ensure consistent behavior across environments
	eventTime := time.Now().UTC().Truncate(time.Second) // Truncate to remove nanoseconds for comparison
	data := map[string]interface{}{
		"fighter1": "Boxer A",
		"fighter2": "Boxer B",
		"winner":   "Boxer A",
	}

	scheduledEvent := NewScheduledEvent(2, eventType, eventTime, data)

	// Serialize to JSON
	jsonBytes, err := scheduledEvent.ToJSON()
	assert.NoError(t, err)
	assert.NotNil(t, jsonBytes)

	// Deserialize back and verify
	var deserializedEvent ScheduledEvent
	err = json.Unmarshal(jsonBytes, &deserializedEvent)
	assert.NoError(t, err)

	assert.Equal(t, scheduledEvent.BoxerID, deserializedEvent.BoxerID)
	assert.Equal(t, scheduledEvent.EventType, deserializedEvent.EventType)
	assert.Equal(t, scheduledEvent.EventTime, deserializedEvent.EventTime)
	// For data comparison, we need to check individual elements since map order may vary
	assert.Equal(t, len(scheduledEvent.Data), len(deserializedEvent.Data))
	for k, v := range scheduledEvent.Data {
		assert.Equal(t, v, deserializedEvent.Data[k])
	}
}

func TestScheduledEvent_FromJSON(t *testing.T) {
	// Test JSON deserialization
	eventType := EventWorldTick
	// Use UTC timezone to ensure consistent behavior across environments
	eventTime := time.Now().UTC().Truncate(time.Second)
	data := map[string]interface{}{
		"tick":        100,
		"world_state": "active",
	}

	// Create original event
	originalEvent := NewScheduledEvent(3, eventType, eventTime, data)

	// Serialize to JSON
	jsonBytes, err := originalEvent.ToJSON()
	assert.NoError(t, err)

	// Create new event and deserialize from JSON
	newEvent := &ScheduledEvent{}
	err = newEvent.FromJSON(jsonBytes)
	assert.NoError(t, err)

	// Verify the deserialized event matches the original
	assert.Equal(t, originalEvent.BoxerID, newEvent.BoxerID)
	assert.Equal(t, originalEvent.EventType, newEvent.EventType)
	assert.Equal(t, originalEvent.EventTime, newEvent.EventTime)
	// For data comparison, we need to check individual elements since map order may vary
	assert.Equal(t, len(originalEvent.Data), len(newEvent.Data))
	for k, v := range originalEvent.Data {
		// Since JSON unmarshals numbers as float64, we need to handle the type conversion
		if expectedNum, ok := v.(int); ok {
			actualNum, ok2 := newEvent.Data[k].(float64)
			if ok2 {
				assert.Equal(t, float64(expectedNum), actualNum)
			} else {
				assert.Equal(t, v, newEvent.Data[k])
			}
		} else {
			assert.Equal(t, v, newEvent.Data[k])
		}
	}
}

func TestScheduledEvent_InvalidJSON(t *testing.T) {
	// Test handling invalid JSON during deserialization
	event := &ScheduledEvent{}
	err := event.FromJSON([]byte("{invalid json"))
	assert.Error(t, err)
}

func TestEventTypeConstants(t *testing.T) {
	// Test that event type constants are properly defined
	assert.Equal(t, EventType("training_complete"), EventTrainingComplete)
	assert.Equal(t, EventType("fight_complete"), EventFightComplete)
	assert.Equal(t, EventType("fight_result"), EventFightResult)
	assert.Equal(t, EventType("world_tick"), EventWorldTick)
}

func TestScheduledEvent_WithNilData(t *testing.T) {
	// Test creating scheduled event with nil data
	eventType := EventTrainingComplete
	eventTime := time.Now()

	scheduledEvent := NewScheduledEvent(1, eventType, eventTime, nil)

	assert.Equal(t, 1, scheduledEvent.BoxerID)
	assert.Equal(t, eventType, scheduledEvent.EventType)
	assert.Equal(t, eventTime, scheduledEvent.EventTime)
	// Data should be initialized to an empty map, not nil
	assert.NotNil(t, scheduledEvent.Data)
	assert.Len(t, scheduledEvent.Data, 0)
}

func TestScheduledEvent_EmptyDataMap(t *testing.T) {
	// Test creating scheduled event with empty data map
	eventType := EventFightComplete
	eventTime := time.Now()

	scheduledEvent := NewScheduledEvent(1, eventType, eventTime, map[string]interface{}{})

	assert.Equal(t, 1, scheduledEvent.BoxerID)
	assert.Equal(t, eventType, scheduledEvent.EventType)
	assert.Equal(t, eventTime, scheduledEvent.EventTime)
	assert.NotNil(t, scheduledEvent.Data)
	assert.Len(t, scheduledEvent.Data, 0)
}

func TestScheduledEvent_ToJSONWithNilData(t *testing.T) {
	// Test JSON serialization with nil data
	eventType := EventWorldTick
	// Use UTC timezone to ensure consistent behavior across environments
	eventTime := time.Now().UTC().Truncate(time.Second)

	scheduledEvent := NewScheduledEvent(1, eventType, eventTime, nil)

	jsonBytes, err := scheduledEvent.ToJSON()
	assert.NoError(t, err)
	assert.NotNil(t, jsonBytes)

	// Verify that we can deserialize it back
	var deserializedEvent ScheduledEvent
	err = json.Unmarshal(jsonBytes, &deserializedEvent)
	assert.NoError(t, err)
	assert.Equal(t, scheduledEvent.BoxerID, deserializedEvent.BoxerID)
	assert.Equal(t, scheduledEvent.EventType, deserializedEvent.EventType)
	assert.Equal(t, scheduledEvent.EventTime, deserializedEvent.EventTime)
	assert.NotNil(t, deserializedEvent.Data)
	assert.Len(t, deserializedEvent.Data, 0)
}
