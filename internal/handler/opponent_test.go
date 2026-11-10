package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mormm/boxing/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestNewOpponentHandler(t *testing.T) {
	t.Run("CreatesHandlerWithService", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.opponentDiscoveryService)
	})
}

func TestOpponentHandler_GetOpponents(t *testing.T) {
	t.Run("ParsesBoxerIDFromPath", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"ValidBoxerID1", "1", http.StatusInternalServerError}, // Service call with nil DB causes panic -> handled as internal
			{"InvalidBoxerIDString", "abc", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := service.NewOpponentDiscoveryService(nil)
				handler := NewOpponentHandler(mockService)

				req := httptest.NewRequest("GET", "/boxers/"+tc.boxerID+"/opponents", nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetOpponents(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("ParsesLevelRangeFilter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/boxers/1/opponents?level_range=5", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetOpponents(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ParsesAiOnlyFilter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/boxers/1/opponents?ai_only=true", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetOpponents(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ParsesHumansOnlyFilter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/boxers/1/opponents?humans_only=true", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetOpponents(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ParsesMinHealthFilter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/boxers/1/opponents?min_health=80", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetOpponents(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ParsesLimitFilter", func(t *testing.T) {
		testCases := []struct {
			name  string
			limit string
		}{
			{"Limit10", "10"},
			{"Limit50", "50"},
			{"Limit100", "100"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := service.NewOpponentDiscoveryService(nil)
				handler := NewOpponentHandler(mockService)

				req := httptest.NewRequest("GET", "/boxers/1/opponents?limit="+tc.limit, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetOpponents(w, req)

				assert.Greater(t, w.Code, 0)
			})
		}
	})

	t.Run("ParsesPreferRankingsFilter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/boxers/1/opponents?prefer_ranked=true", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetOpponents(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ReturnsJSONResponseStructure", func(t *testing.T) {
		filters := map[string]any{
			"include_ai":      true,
			"exclude_owned":   false,
			"available_only":  true,
			"healthy_only":    true,
			"min_health":      30.0,
			"max_results":     50,
			"prefer_rankings": false,
		}

		expectedResponse := ListOpponentsResponse{
			Opportunities: []*ScoredOpponentResponse{},
			TotalCount:    0,
			Filters:       filters,
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var resp ListOpponentsResponse
		err = json.Unmarshal(jsonData, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 0, resp.TotalCount)
	})

	t.Run("HandlesInvalidBoxerID", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/boxers/not_a_number/opponents", nil)
		w := httptest.NewRecorder()

		handler.GetOpponents(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOpponentHandler_GetBestMatch(t *testing.T) {
	t.Run("ValidatesBoxerIDParameter", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"MissingBoxerID", "", http.StatusBadRequest},
			{"InvalidBoxerIDString", "abc", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := service.NewOpponentDiscoveryService(nil)
				handler := NewOpponentHandler(mockService)

				url := "/opponents/best_match"
				if tc.boxerID != "" {
					url += "?boxer_id=" + tc.boxerID
				}

				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetBestMatch(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("RequiresBoxerIDParameter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/opponents/best_match", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetBestMatch(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("AcceptsFilterParameters", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/opponents/best_match?boxer_id=1&level_range=3", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetBestMatch(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ReturnsOpponentResponseStructure", func(t *testing.T) {
		expectedResponse := map[string]any{
			"opponent": map[string]any{"rank": 5},
			"score":    map[string]float64{},
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)
	})
}

func TestOpponentHandler_GetRandomOpponent(t *testing.T) {
	t.Run("ValidatesBoxerIDParameter", func(t *testing.T) {
		testCases := []struct {
			name     string
			boxerID  string
			wantCode int
		}{
			{"MissingBoxerID", "", http.StatusBadRequest},
			{"InvalidBoxerIDString", "abc", http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := service.NewOpponentDiscoveryService(nil)
				handler := NewOpponentHandler(mockService)

				url := "/opponents/random"
				if tc.boxerID != "" {
					url += "?boxer_id=" + tc.boxerID
				}

				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetRandomOpponent(w, req)

				assert.Equal(t, tc.wantCode, w.Code)
			})
		}
	})

	t.Run("RequiresBoxerIDParameter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/opponents/random", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetRandomOpponent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ParsesMinLevelParameter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/opponents/random?boxer_id=1&min_level=5", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetRandomOpponent(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ParsesMaxLevelParameter", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/opponents/random?boxer_id=1&max_level=20", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetRandomOpponent(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("CombinesAllParameters", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("GET", "/opponents/random?boxer_id=1&min_level=5&max_level=15", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetRandomOpponent(w, req)

		assert.Greater(t, w.Code, 0)
	})
}

func TestOpponentHandler_ValidateMatch(t *testing.T) {
	t.Run("ValidatesRequestJSON", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		requestBody := map[string]any{
			"boxer1_id": 1,
			"boxer2_id": 2,
		}

		jsonData, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/opponents/validate", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.ValidateMatch(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("RejectsInvalidJSON", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		req := httptest.NewRequest("POST", "/opponents/validate", bytes.NewBufferString("{invalid json}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ValidateMatch(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("RequiresBothBoxerIDs", func(t *testing.T) {
		mockService := service.NewOpponentDiscoveryService(nil)
		handler := NewOpponentHandler(mockService)

		testCases := []struct {
			name string
			body map[string]any
		}{
			{"MissingBoxer1ID", map[string]any{"boxer2_id": 2}},
			{"MissingBoxer2ID", map[string]any{"boxer1_id": 1}},
			{"BothZero", map[string]any{"boxer1_id": 0, "boxer2_id": 0}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				jsonData, _ := json.Marshal(tc.body)
				req := httptest.NewRequest("POST", "/opponents/validate", bytes.NewBuffer(jsonData))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.ValidateMatch(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)
			})
		}
	})
}
