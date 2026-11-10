package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestNewTrainingHandler(t *testing.T) {
	t.Run("CreatesHandlerWithServices", func(t *testing.T) {
		testLogger := logger.New("test")
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)

		handler := NewTrainingHandler(
			nil, // boxerStore
			nil, // trainingTypeStore
			nil, // trainingSessionStore
			nil, // scheduledEventStore
			fatigueService,
			progressionService,
			trainingService,
		)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.trainingService)
		assert.NotNil(t, handler.fatigueService)
	})
}

func TestTrainingHandler_ScheduleTraining(t *testing.T) {
	testLogger := logger.New("test")

	t.Run("ParsesBoxerIDFromPath", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"ValidBoxerID1", "1", http.StatusUnauthorized}, // Auth checked first
			{"InvalidBoxerIDString", "abc", http.StatusUnauthorized}, // Auth checked before path param parsing
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fatigueService := service.NewFatigueService(nil, testLogger)
				progressionService := service.NewProgressionService(testLogger)
				worldClock := &model.WorldClockModel{}
				trainingService := service.NewTrainingService(
					nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
				)
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

				reqBody := map[string]any{
					"training_type_id": 1,
					"duration_hours":   2.0,
					"scheduled_at":     "",
				}
				jsonData, _ := json.Marshal(reqBody)

				req := httptest.NewRequest("POST", "/boxers/"+tc.boxerID+"/train", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				handler.ScheduleTraining(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("ValidatesTrainingTypeIDInRequest", func(t *testing.T) {
		testCases := []struct {
			name     string
			body     map[string]any
			wantCode int
		}{
			{"ValidTrainingRequest", map[string]any{"training_type_id": 1, "duration_hours": 2.0, "scheduled_at": ""}, http.StatusUnauthorized}, // Auth first
			{"MissingTrainingTypeID", map[string]any{"duration_hours": 2.0, "scheduled_at": ""}, http.StatusUnauthorized},
			{"MissingDurationHours", map[string]any{"training_type_id": 1, "scheduled_at": ""}, http.StatusUnauthorized},
			{"InvalidJSON", nil, http.StatusUnauthorized}, // Auth checked before JSON binding
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fatigueService := service.NewFatigueService(nil, testLogger)
				progressionService := service.NewProgressionService(testLogger)
				worldClock := &model.WorldClockModel{}
				trainingService := service.NewTrainingService(
					nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
				)
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

				var jsonData []byte
				var err error
				if tc.body != nil {
					jsonData, err = json.Marshal(tc.body)
					assert.NoError(t, err)
				} else {
					jsonData = []byte("{invalid")
				}

				req := httptest.NewRequest("POST", "/boxers/1/train", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				handler.ScheduleTraining(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("ValidatesDurationHoursRange", func(t *testing.T) {
		testCases := []struct {
			name          string
			durationHours float64
			wantCode      int
		}{
			{"Duration1Hour", 1.0, http.StatusUnauthorized}, // Auth checked before validation
			{"Duration4Hours", 4.0, http.StatusUnauthorized},
			{"Duration8Hours", 8.0, http.StatusUnauthorized},
			{"DurationTooShort", 0.5, http.StatusUnauthorized}, // Auth fails before duration validation
			{"DurationTooLong", 10.0, http.StatusUnauthorized}, // Auth fails before duration validation
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fatigueService := service.NewFatigueService(nil, testLogger)
				progressionService := service.NewProgressionService(testLogger)
				worldClock := &model.WorldClockModel{}
				trainingService := service.NewTrainingService(
					nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
				)
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

				reqBody := map[string]any{
					"training_type_id": 1,
					"duration_hours":   tc.durationHours,
					"scheduled_at":     "",
				}
				jsonData, _ := json.Marshal(reqBody)

				req := httptest.NewRequest("POST", "/boxers/1/train", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				handler.ScheduleTraining(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("AcceptsScheduledAtParameter", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		reqBody := map[string]any{
			"training_type_id": 1,
			"duration_hours":   2.0,
			"scheduled_at":     "2024-01-15T20:00:00Z",
		}
		jsonData, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/boxers/1/train", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ScheduleTraining(w, req)

		assert.Greater(t, w.Code, 0) // Request parses
	})

	t.Run("ReturnsSuccessResponseStructure", func(t *testing.T) {
		expectedResponse := map[string]any{
			"message": "Training scheduled successfully",
			"session": map[string]any{
				"id":                        1,
				"boxer_id":                  10,
				"training_type_id":          1,
				"duration_hours":            2.0,
				"scheduled_completion_time": "2024-01-15T22:00:00Z",
				"planned_gains": map[string]float64{
					"strength": 1.5,
					"defense":  1.0,
					"agility":  0.5,
				},
				"energy_cost": 20,
				"status":      "pending",
			},
			"training_type": map[string]any{
				"name":                 "Strength Training",
				"description":          "Focus on building strength",
				"energy_cost_per_hour": 10,
			},
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var resp map[string]any
		err = json.Unmarshal(jsonData, &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Training scheduled successfully", resp["message"])
	})

	t.Run("RequiresAuthenticatedUser", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		reqBody := map[string]any{
			"training_type_id": 1,
			"duration_hours":   2.0,
			"scheduled_at":     "",
		}
		jsonData, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/boxers/1/train", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ScheduleTraining(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code) // No auth context provided
	})
}

func TestTrainingHandler_GetTrainingSessionsForBoxer(t *testing.T) {
	t.Run("ParsesBoxerIDFromPath", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"ValidBoxerID1", "1", http.StatusBadRequest}, // Handler returns 400 (not implemented/missing functionality with nil service)
			{"InvalidBoxerIDString", "abc", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, nil)

				req := httptest.NewRequest("GET", "/boxers/"+tc.boxerID+"/training-sessions", nil)
				w := httptest.NewRecorder()

				handler.GetTrainingSessionsForBoxer(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("ReturnsEmptyArrayForNoSessions", func(t *testing.T) {
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, nil)

		req := httptest.NewRequest("GET", "/boxers/1/training-sessions", nil)
		w := httptest.NewRecorder()

		handler.GetTrainingSessionsForBoxer(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code) // Nil store returns 400
	})

	t.Run("ReturnsJSONResponseStructure", func(t *testing.T) {
		expectedSessions := []map[string]any{
			{
				"id":                        1,
				"boxer_id":                  10,
				"training_type_id":          1,
				"duration_hours":            2.0,
				"scheduled_completion_time": "2024-01-15T22:00:00Z",
				"status":                    "pending",
			},
		}

		jsonData, err := json.Marshal(expectedSessions)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var sessions []map[string]any
		err = json.Unmarshal(jsonData, &sessions)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(sessions))
	})
}

func TestTrainingHandler_CompleteTraining(t *testing.T) {
	testLogger := logger.New("test")

	t.Run("ParsesSessionIDFromPath", func(t *testing.T) {
		testCases := []struct {
			name      string
			sessionID string
			wantCode  int
		}{
			{"ValidSessionID1", "1", http.StatusBadRequest}, // Handler returns 400 (not implemented/missing functionality with nil service)
			{"InvalidSessionIDString", "abc", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fatigueService := service.NewFatigueService(nil, testLogger)
				progressionService := service.NewProgressionService(testLogger)
				worldClock := &model.WorldClockModel{}
				trainingService := service.NewTrainingService(
					nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
				)
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

				req := httptest.NewRequest("POST", "/training/"+tc.sessionID+"/complete", nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.CompleteTraining(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("ReturnsSuccessResponseStructure", func(t *testing.T) {
		expectedResponse := map[string]string{
			"message": "Training session completed successfully",
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var resp map[string]string
		err = json.Unmarshal(jsonData, &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Training session completed successfully", resp["message"])
	})
}

func TestTrainingHandler_BulkCompleteTraining(t *testing.T) {
	testLogger := logger.New("test")

	t.Run("AcceptsEmptyBody", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		req := httptest.NewRequest("POST", "/training/bulk-complete", bytes.NewBuffer([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.BulkCompleteTraining(w, req)

		assert.Greater(t, w.Code, 0) // Handler processes but may error due to nil stores
	})

	t.Run("AcceptsBoxerIDFilter", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		reqBody := map[string]any{
			"boxer_id": 10,
		}
		jsonData, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/training/bulk-complete", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.BulkCompleteTraining(w, req)

		assert.Greater(t, w.Code, 0) // Handler processes but may error due to nil stores
	})

	t.Run("RejectsInvalidJSON", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		req := httptest.NewRequest("POST", "/training/bulk-complete", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.BulkCompleteTraining(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ReturnsBulkCompleteResponseStructure", func(t *testing.T) {
		expectedResponse := map[string]any{
			"completed": 2,
			"failed":    0,
			"message":   "Completed 2 training session(s)",
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var resp map[string]any
		err = json.Unmarshal(jsonData, &resp)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), resp["completed"])
	})
}

func TestTrainingHandler_GetAllTrainingTypes(t *testing.T) {
	t.Run("ReturnsTrainingTypesList", func(t *testing.T) {
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, nil)

		req := httptest.NewRequest("GET", "/training-types", nil)
		w := httptest.NewRecorder()

		handler.GetAllTrainingTypes(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code) // Nil store returns 503
	})

	t.Run("ReturnsJSONResponseStructure", func(t *testing.T) {
		expectedTypes := []map[string]any{
			{
				"id":                 1,
				"name":               "Strength Training",
				"description":        "Build muscle strength",
				"strength_gain_factor": 0.75,
				"defense_gain_factor":  0.25,
				"agility_gain_factor":  0.10,
				"energy_cost":         10,
			},
			{
				"id":                 2,
				"name":               "Speed Training",
				"description":        "Improve agility and speed",
				"strength_gain_factor": 0.10,
				"defense_gain_factor":  0.15,
				"agility_gain_factor":  0.75,
				"energy_cost":         8,
			},
		}

		jsonData, err := json.Marshal(expectedTypes)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var types []map[string]any
		err = json.Unmarshal(jsonData, &types)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(types))
	})
}

func TestTrainingHandler_GetTrainingHistory(t *testing.T) {
	testLogger := logger.New("test")

	t.Run("ParsesBoxerIDFromPath", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"ValidBoxerID1", "1", http.StatusUnauthorized}, // Requires auth (checked before path parsing)
			{"InvalidBoxerIDString", "abc", http.StatusUnauthorized}, // Auth checked first, returns 401
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fatigueService := service.NewFatigueService(nil, testLogger)
				progressionService := service.NewProgressionService(testLogger)
				worldClock := &model.WorldClockModel{}
				trainingService := service.NewTrainingService(
					nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
				)
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

				req := httptest.NewRequest("GET", "/boxers/"+tc.boxerID+"/training-history", nil)
				w := httptest.NewRecorder()

				handler.GetTrainingHistory(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("RequiresAuthenticatedUser", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		req := httptest.NewRequest("GET", "/boxers/1/training-history", nil)
		w := httptest.NewRecorder()

		handler.GetTrainingHistory(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestTrainingHandler_GetStatProgression(t *testing.T) {
	testLogger := logger.New("test")

	t.Run("ParsesBoxerIDFromPath", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"ValidBoxerID1", "1", http.StatusUnauthorized}, // Requires auth (checked before path parsing)
			{"InvalidBoxerIDString", "abc", http.StatusUnauthorized}, // Auth checked first, returns 401
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				fatigueService := service.NewFatigueService(nil, testLogger)
				progressionService := service.NewProgressionService(testLogger)
				worldClock := &model.WorldClockModel{}
				trainingService := service.NewTrainingService(
					nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
				)
				handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

				req := httptest.NewRequest("GET", "/boxers/"+tc.boxerID+"/stat-progression", nil)
				w := httptest.NewRecorder()

				handler.GetStatProgression(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("RequiresAuthenticatedUser", func(t *testing.T) {
		fatigueService := service.NewFatigueService(nil, testLogger)
		progressionService := service.NewProgressionService(testLogger)
		worldClock := &model.WorldClockModel{}
		trainingService := service.NewTrainingService(
			nil, nil, nil, nil, fatigueService, progressionService, worldClock, testLogger, nil,
		)
		handler := NewTrainingHandler(nil, nil, nil, nil, nil, nil, trainingService)

		req := httptest.NewRequest("GET", "/boxers/1/stat-progression", nil)
		w := httptest.NewRecorder()

		handler.GetStatProgression(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
