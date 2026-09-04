package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser(t *testing.T) {
	t.Run("User struct creation", func(t *testing.T) {
		user := &User{
			ID:             1,
			Username:       "testuser",
			Email:          "test@example.com",
			HashedPassword: "hashedpassword",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		assert.NotNil(t, user)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("User JSON marshaling", func(t *testing.T) {
		user := &User{
			ID:             1,
			Username:       "testuser",
			Email:          "test@example.com",
			HashedPassword: "hashedpassword",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		data, err := json.Marshal(user)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check that HashedPassword is not included in JSON (it has -json:"-" tag)
		assert.NotContains(t, string(data), "hashedpassword")
		assert.Contains(t, string(data), "testuser")
		assert.Contains(t, string(data), "test@example.com")
	})

	t.Run("User JSON unmarshaling", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"username": "testuser",
			"email": "test@example.com",
			"created_at": "2023-01-01T00:00:00Z",
			"updated_at": "2023-01-01T00:00:00Z"
		}`

		var user User
		err := json.Unmarshal([]byte(jsonData), &user)
		assert.NoError(t, err)
		assert.Equal(t, 1, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})
}

func TestBoxer(t *testing.T) {
	t.Run("Boxer struct creation", func(t *testing.T) {
		boxer := &Boxer{
			ID:         1,
			UserID:     1,
			Name:       "Test Boxer",
			Nickname:   stringPtr("TB"),
			PositionX:  0.0,
			PositionY:  0.0,
			Health:     100.0,
			Energy:     100.0,
			Strength:   10.0,
			Defense:    10.0,
			Agility:    10.0,
			Experience: 0.0,
			Level:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		assert.NotNil(t, boxer)
		assert.Equal(t, "Test Boxer", boxer.Name)
		assert.NotNil(t, boxer.Nickname)
		assert.Equal(t, "TB", *boxer.Nickname)
	})

	t.Run("Boxer JSON marshaling", func(t *testing.T) {
		boxer := &Boxer{
			ID:         1,
			UserID:     1,
			Name:       "Test Boxer",
			Nickname:   stringPtr("TB"),
			PositionX:  0.0,
			PositionY:  0.0,
			Health:     100.0,
			Energy:     100.0,
			Strength:   10.0,
			Defense:    10.0,
			Agility:    10.0,
			Experience: 0.0,
			Level:      1,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		data, err := json.Marshal(boxer)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check that the JSON contains expected fields
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "Test Boxer")
		assert.Contains(t, jsonStr, "TB")
		assert.Contains(t, jsonStr, "100")
	})

	t.Run("Boxer JSON unmarshaling with nil nickname", func(t *testing.T) {
		jsonData := `{
			"id": 1,
			"user_id": 1,
			"name": "Test Boxer",
			"nickname": null,
			"position_x": 0.0,
			"position_y": 0.0,
			"health": 100.0,
			"energy": 100.0,
			"strength": 10.0,
			"defense": 10.0,
			"agility": 10.0,
			"experience": 0.0,
			"level": 1,
			"created_at": "2023-01-01T00:00:00Z",
			"updated_at": "2023-01-01T00:00:00Z"
		}`

		var boxer Boxer
		err := json.Unmarshal([]byte(jsonData), &boxer)
		assert.NoError(t, err)
		assert.Equal(t, "Test Boxer", boxer.Name)
		assert.Nil(t, boxer.Nickname)
	})
}

func TestFight(t *testing.T) {
	t.Run("FightStatus constants", func(t *testing.T) {
		assert.Equal(t, FightStatusScheduled, FightStatus("scheduled"))
		assert.Equal(t, FightStatusInProgress, FightStatus("in_progress"))
		assert.Equal(t, FightStatusCompleted, FightStatus("completed"))
		assert.Equal(t, FightStatusCancelled, FightStatus("canceled"))
	})

	t.Run("Fight struct creation", func(t *testing.T) {
		now := time.Now()
		fight := &Fight{
			ID:            1,
			Boxer1ID:      1,
			Boxer2ID:      2,
			Status:        FightStatusScheduled,
			ScheduledTime: &now,
			StartTime:     &now,
			EndTime:       &now,
			WinnerID:      intPtr(1),
			Round:         1,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		assert.NotNil(t, fight)
		assert.Equal(t, FightStatusScheduled, fight.Status)
		assert.NotNil(t, fight.WinnerID)
		assert.Equal(t, 1, *fight.WinnerID)
	})

	t.Run("Fight JSON marshaling", func(t *testing.T) {
		now := time.Now()
		fight := &Fight{
			ID:            1,
			Boxer1ID:      1,
			Boxer2ID:      2,
			Status:        FightStatusScheduled,
			ScheduledTime: &now,
			StartTime:     &now,
			EndTime:       &now,
			WinnerID:      intPtr(1),
			Round:         1,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		data, err := json.Marshal(fight)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check that the JSON contains expected fields
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "scheduled")
		assert.Contains(t, jsonStr, "1")
	})
}

func TestScheduledEvent(t *testing.T) {
	t.Run("EventType constants", func(t *testing.T) {
		assert.Equal(t, EventTypeTraining, EventType("training"))
		assert.Equal(t, EventTypeRest, EventType("rest"))
		assert.Equal(t, EventTypeCompetition, EventType("competition"))
		assert.Equal(t, EventTypeOther, EventType("other"))
	})

	t.Run("ScheduledEvent struct creation", func(t *testing.T) {
		now := time.Now()
		event := &ScheduledEvent{
			ID:        1,
			BoxerID:   1,
			EventType: EventTypeTraining,
			EventTime: now,
			CreatedAt: now,
		}

		assert.NotNil(t, event)
		assert.Equal(t, EventTypeTraining, event.EventType)
	})

	t.Run("ScheduledEvent JSON marshaling", func(t *testing.T) {
		now := time.Now()
		event := &ScheduledEvent{
			ID:        1,
			BoxerID:   1,
			EventType: EventTypeTraining,
			EventTime: now,
			CreatedAt: now,
		}

		data, err := json.Marshal(event)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check that the JSON contains expected fields
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "training")
	})
}

func TestTrainingSession(t *testing.T) {
	t.Run("TrainingSessionStatus constants", func(t *testing.T) {
		assert.Equal(t, TrainingSessionPending, TrainingSessionStatus("pending"))
		assert.Equal(t, TrainingSessionCompleted, TrainingSessionStatus("completed"))
		assert.Equal(t, TrainingSessionCancelled, TrainingSessionStatus("cancelled"))
	})

	t.Run("TrainingType struct creation", func(t *testing.T) {
		now := time.Now()
		trainingType := &TrainingType{
			ID:                 1,
			Name:               "strength_training",
			Description:        stringPtr("Build raw power"),
			StrengthGainFactor: 1.0,
			DefenseGainFactor:  0.1,
			AgilityGainFactor:  0.05,
			EnergyCost:         15,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		assert.NotNil(t, trainingType)
		assert.Equal(t, "strength_training", trainingType.Name)
		assert.Equal(t, 1.0, trainingType.StrengthGainFactor)
		assert.Equal(t, 15, trainingType.EnergyCost)
	})

	t.Run("TrainingSession struct creation", func(t *testing.T) {
		now := time.Now()
		session := &TrainingSession{
			ID:                  1,
			BoxerID:             1,
			TrainingTypeID:      1,
			DurationHours:       1.0,
			PlannedStrengthGain: 5.0,
			PlannedDefenseGain:  2.0,
			PlannedAgilityGain:  3.0,
			Status:              TrainingSessionPending,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		assert.NotNil(t, session)
		assert.Equal(t, 1, session.TrainingTypeID)
		assert.Equal(t, 1.0, session.DurationHours)
		assert.Equal(t, 5.0, session.PlannedStrengthGain)
	})

	t.Run("TrainingSession JSON marshaling", func(t *testing.T) {
		now := time.Now()
		session := &TrainingSession{
			ID:                  1,
			BoxerID:             1,
			TrainingTypeID:      1,
			DurationHours:       1.0,
			PlannedStrengthGain: 5.0,
			PlannedDefenseGain:  2.0,
			PlannedAgilityGain:  3.0,
			Status:              TrainingSessionPending,
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		data, err := json.Marshal(session)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		// Check that the JSON contains expected fields
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "pending")
		assert.Contains(t, jsonStr, "1")
		assert.Contains(t, jsonStr, "5")
	})
}

// Helper functions to create pointers for tests
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
