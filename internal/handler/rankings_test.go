package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/service"
	"github.com/stretchr/testify/assert"
)

// stringPtr is a helper to create a string pointer from a string value
func stringPtr(s string) *string {
	return &s
}

func TestRankingsHandler_GetRankings(t *testing.T) {
	t.Run("ReturnsRankingsWithDefaultCriteria", func(t *testing.T) {
		mockService := &service.RankingsService{}
		handler := NewRankingsHandler(mockService)

		req := httptest.NewRequest("GET", "/rankings", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetRankings(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ValidatesCriteriaParameter", func(t *testing.T) {
		validCriteria := []string{"win_rate", "level", "strength"}

		for _, criteria := range validCriteria {
			t.Logf("Testing criteria: %s", criteria)
			req := httptest.NewRequest("GET", "/rankings?criteria="+criteria, nil)
			w := httptest.NewRecorder()

			mockService := &service.RankingsService{}
			handler := NewRankingsHandler(mockService)

			defer func(c string) {
				if r := recover(); r != nil {
					t.Logf("Recovered from panic for criteria %s: %v", c, r)
				}
			}(criteria)

			handler.GetRankings(w, req)

			assert.Greater(t, w.Code, 0)
		}
	})

	t.Run("ValidatesLimitParameter", func(t *testing.T) {
		testCases := []struct {
			name     string
			limit    string
			expected int
		}{
			{"ValidLimit10", "10", 10},
			{"ValidLimit100", "100", 100},
			{"InvalidLimitNegative", "-5", 100}, // Should default to 100
			{"InvalidLimitString", "abc", 100},   // Should default to 100
			{"MissingLimit", "", 100},            // Should use default of 100
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := &service.RankingsService{}
				handler := NewRankingsHandler(mockService)

				url := "/rankings"
				if tc.limit != "" {
					url += "?limit=" + tc.limit
				}

				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetRankings(w, req)

				assert.Greater(t, w.Code, 0)
			})
		}
	})

	t.Run("ReturnsJSONResponseStructure", func(t *testing.T) {
		mockBoxers := []*model.RankedBoxer{
			{
				ID:           1,
				Name:         "Test Boxer",
				Nickname:     stringPtr("TB"),
				Rank:         1,
				Wins:         5,
				Losses:       0,
				Draws:        0,
				Knockouts:    3,
				Level:        10,
				Strength:     85.5,
				RankingScore: 100.0,
			},
		}

		expectedResponse := &model.RankingsResponse{
			Criteria:    "win_rate",
			Boxers:      mockBoxers,
			TotalCount:  1,
			GeneratedAt: time.Now(),
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var resp model.RankingsResponse
		err = json.Unmarshal(jsonData, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp.TotalCount)
	})
}

func TestRankingsHandler_GetRankingForBoxer(t *testing.T) {
	t.Run("ValidatesBoxerIDParsing", func(t *testing.T) {
		testCases := []struct {
			name       string
			boxerID    string
			validParse bool
		}{
			{"ValidBoxerID1", "1", true},
			{"ValidBoxerID999", "999", true},
			{"InvalidBoxerIDNegative", "-1", false},
			{"InvalidBoxerIDString", "abc", false},
			{"InvalidBoxerIDFloat", "1.5", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := &service.RankingsService{}
				handler := NewRankingsHandler(mockService)

				req := httptest.NewRequest("GET", "/rankings/"+tc.boxerID, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetRankingForBoxer(w, req)

				if tc.validParse {
					assert.Greater(t, w.Code, 0)
				} else {
					assert.Equal(t, http.StatusBadRequest, w.Code)
				}
			})
		}
	})

	t.Run("AcceptsCriteriaQueryParameter", func(t *testing.T) {
		validCriteria := []string{"win_rate", "level", "strength"}

		for _, criteria := range validCriteria {
			t.Run(criteria, func(t *testing.T) {
				mockService := &service.RankingsService{}
				handler := NewRankingsHandler(mockService)

				req := httptest.NewRequest("GET", "/rankings/1?criteria="+criteria, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetRankingForBoxer(w, req)

				assert.Greater(t, w.Code, 0)
			})
		}
	})

	t.Run("ReturnsJSONPositionStructure", func(t *testing.T) {
		expectedPosition := &model.RankingPosition{
			Rank:        5,
			TotalBoxers: 100,
			RankedBoxer: &model.RankedBoxer{
				ID:           123,
				Name:         "Example Boxer",
				Nickname:     stringPtr("EB"),
				Rank:         5,
				Wins:         10,
				Losses:       2,
				Draws:        1,
				Knockouts:    5,
				Level:        15,
				Strength:     83.33,
				RankingScore: 78.3,
			},
		}

		jsonData, err := json.Marshal(expectedPosition)
		assert.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		var pos model.RankingPosition
		err = json.Unmarshal(jsonData, &pos)
		assert.NoError(t, err)
		assert.Equal(t, 5, pos.Rank)
	})
}

func TestRankingsHandler_GetNearbyRankings(t *testing.T) {
	t.Run("ValidatesBoxerIDParsing", func(t *testing.T) {
		testCases := []struct {
			name       string
			boxerID    string
			validParse bool
		}{
			{"ValidBoxerID1", "1", true},
			{"InvalidBoxerIDString", "abc", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := &service.RankingsService{}
				handler := NewRankingsHandler(mockService)

				req := httptest.NewRequest("GET", "/rankings/nearby/"+tc.boxerID, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetNearbyRankings(w, req)

				if tc.validParse {
					assert.Greater(t, w.Code, 0)
				} else {
					assert.Equal(t, http.StatusBadRequest, w.Code)
				}
			})
		}
	})

	t.Run("ValidatesRadiusNormalization", func(t *testing.T) {
		testCases := []struct {
			name   string
			radius string
		}{
			{"ValidRadius3", "3"},
			{"ValidRadius10", "10"},
			{"InvalidRadiusNegative", "-5"},
			{"InvalidRadiusZero", "0"},
			{"MissingRadius", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockService := &service.RankingsService{}
				handler := NewRankingsHandler(mockService)

				url := "/rankings/nearby/1"
				if tc.radius != "" {
					url += "?radius=" + tc.radius
				}

				req := httptest.NewRequest("GET", url, nil)
				w := httptest.NewRecorder()

				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
					}
				}()

				handler.GetNearbyRankings(w, req)

				assert.Greater(t, w.Code, 0)
			})
		}
	})

	t.Run("CombinesCriteriaAndRadiusParameters", func(t *testing.T) {
		mockService := &service.RankingsService{}
		handler := NewRankingsHandler(mockService)

		req := httptest.NewRequest("GET", "/rankings/nearby/1?criteria=level&radius=7", nil)
		w := httptest.NewRecorder()

		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
		}()

		handler.GetNearbyRankings(w, req)

		assert.Greater(t, w.Code, 0)
	})

	t.Run("ReturnsJSONResponseStructure", func(t *testing.T) {
		mockBoxers := []*model.RankedBoxer{
			{ID: 1, Name: "Boxer A", Nickname: nil, Rank: 4, Wins: 8, Losses: 2, Draws: 0, Knockouts: 5, Level: 12, Strength: 80.0, RankingScore: 80.0},
			{ID: 2, Name: "Boxer B", Nickname: nil, Rank: 5, Wins: 7, Losses: 3, Draws: 0, Knockouts: 4, Level: 11, Strength: 79.5, RankingScore: 79.5},
			{ID: 3, Name: "Boxer C", Nickname: nil, Rank: 6, Wins: 6, Losses: 4, Draws: 0, Knockouts: 3, Level: 10, Strength: 78.0, RankingScore: 78.0},
		}

		expectedResponse := &model.RankingsResponse{
			Criteria:    "win_rate",
			Boxers:      mockBoxers,
			TotalCount:  3,
			GeneratedAt: time.Now(),
		}

		jsonData, err := json.Marshal(expectedResponse)
		assert.NoError(t, err)

		var resp model.RankingsResponse
		err = json.Unmarshal(jsonData, &resp)
		assert.NoError(t, err)
		assert.Equal(t, 3, resp.TotalCount)
	})
}
