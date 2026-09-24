package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/service"
)

// OpponentHandler handles opponent discovery and matchmaking requests
type OpponentHandler struct {
	opponentDiscoveryService *service.OpponentDiscoveryService
	rand                     *rand.Rand
}

// NewOpponentHandler creates a new OpponentHandler
func NewOpponentHandler(opponentService *service.OpponentDiscoveryService) *OpponentHandler {
	return &OpponentHandler{
		opponentDiscoveryService: opponentService,
		rand:                     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// OpponentResponse represents the API response for opponent discovery
type OpponentResponse struct {
	Opponent *db.RankedOpponent  `json:"opponent"`
	Score    service.OpponentScore `json:"score"`
}

// ListOpponentsResponse represents the API response for listing opponents
type ListOpponentsResponse struct {
	Opportunities []*ScoredOpponentResponse `json:"opportunities"`
	TotalCount    int                       `json:"total_count"`
	Filters       map[string]interface{}    `json:"filters"`
}

// ScoredOpponentResponse combines opponent data with scoring information
type ScoredOpponentResponse struct {
	Opponent  *db.RankedOpponent    `json:"opponent"`
	Score     service.OpponentScore `json:"score"`
	Reasoning string                `json:"reasoning"`
}

// GetOpponents handles GET /boxers/{id}/opponents
// Returns a filtered and scored list of available opponents
func (h *OpponentHandler) GetOpponents(w http.ResponseWriter, r *http.Request) {
	// Parse boxer ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/boxers/")
	idStr := strings.Split(path, "/")[0]
	boxerID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid boxer ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	filters := parseOpponentFilters(r.URL.Query())

	// Get opponents
	opponents, err := h.opponentDiscoveryService.FindOpponents(boxerID, filters)
	if err != nil {
		if err == db.ErrNoOpponentsFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ListOpponentsResponse{
				Opportunities: []*ScoredOpponentResponse{},
				TotalCount:    0,
				Filters:       buildFilterMap(filters),
			})
			return
		}
		http.Error(w, "Failed to find opponents", http.StatusInternalServerError)
		return
	}

	// Get the requesting boxer for scoring
	boxer, err := h.opponentDiscoveryService.GetBoxerByID(boxerID)
	if err != nil {
		http.Error(w, "Boxer not found", http.StatusNotFound)
		return
	}

	// Score and sort opponents
	scoredOpponents := service.ScoreRankedOpponents(boxer, opponents)

	// Build response
	response := ListOpponentsResponse{
		Opportunities: make([]*ScoredOpponentResponse, len(scoredOpponents)),
		TotalCount:    len(scoredOpponents),
		Filters:       buildFilterMap(filters),
	}

	for i, scored := range scoredOpponents {
		response.Opportunities[i] = &ScoredOpponentResponse{
			Opponent:  scored.RankedOpponent,
			Score:     scored.Score,
			Reasoning: scored.Score.Reasoning,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetBestMatch handles GET /opponents/best_match
// Returns the single best match for a boxer
func (h *OpponentHandler) GetBestMatch(w http.ResponseWriter, r *http.Request) {
	// Parse boxer ID from query parameter
	boxerIDStr := r.URL.Query().Get("boxer_id")
	if boxerIDStr == "" {
		http.Error(w, "boxer_id parameter is required", http.StatusBadRequest)
		return
	}

	boxerID, err := strconv.Atoi(boxerIDStr)
	if err != nil {
		http.Error(w, "Invalid boxer_id", http.StatusBadRequest)
		return
	}

	// Parse optional filters
	filters := parseOpponentFilters(r.URL.Query())

	// Find opponents
	opponents, err := h.opponentDiscoveryService.FindOpponents(boxerID, filters)
	if err != nil {
		if err == db.ErrNoOpponentsFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "no_opponents_found",
				"message": "No available opponents match your criteria",
			})
			return
		}
		http.Error(w, "Failed to find opponents", http.StatusInternalServerError)
		return
	}

	// Get the requesting boxer for scoring
	boxer, err := h.opponentDiscoveryService.GetBoxerByID(boxerID)
	if err != nil {
		http.Error(w, "Boxer not found", http.StatusNotFound)
		return
	}

	// Find best match
	bestMatch, err := service.FindBestMatch(boxer, opponents)
	if err != nil {
		http.Error(w, "No suitable opponents available", http.StatusNotFound)
		return
	}

	response := OpponentResponse{
		Opponent: bestMatch.RankedOpponent,
		Score:    bestMatch.Score,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetRandomOpponent handles GET /opponents/random
// Returns a random opponent within specified constraints
func (h *OpponentHandler) GetRandomOpponent(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	minLevelStr := r.URL.Query().Get("min_level")
	maxLevelStr := r.URL.Query().Get("max_level")

	minLevel := 1
	maxLevel := 999

	if minLevelStr != "" {
		if val, err := strconv.Atoi(minLevelStr); err == nil && val > 0 {
			minLevel = val
		}
	}

	if maxLevelStr != "" {
		if val, err := strconv.Atoi(maxLevelStr); err == nil && val > 0 {
			maxLevel = val
		}
	}

	// Get boxer ID from query parameter
	boxerIDStr := r.URL.Query().Get("boxer_id")
	if boxerIDStr == "" {
		http.Error(w, "boxer_id parameter is required", http.StatusBadRequest)
		return
	}

	boxerID, err := strconv.Atoi(boxerIDStr)
	if err != nil {
		http.Error(w, "Invalid boxer_id", http.StatusBadRequest)
		return
	}

	// Build filter
	filters := db.OpponentFilter{
		BoxerID:       boxerID,
		MinLevel:      minLevel,
		MaxLevel:      maxLevel,
		IncludeAI:     true,
		ExcludeOwned:  true,
		AvailableOnly: true,
		HealthyOnly:   true,
		MinHealth:     30.0,
		MaxResults:    100, // Get more to pick from
	}

	// Find opponents
	opponents, err := h.opponentDiscoveryService.FindOpponents(boxerID, filters)
	if err != nil {
		if err == db.ErrNoOpponentsFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "no_opponents_found",
				"message": "No available opponents found within level range",
			})
			return
		}
		http.Error(w, "Failed to find opponents", http.StatusInternalServerError)
		return
	}

	// Pick random opponent
	idx := h.rand.Intn(len(opponents))
	selected := opponents[idx]

	// Get boxer for scoring
	boxer, err := h.opponentDiscoveryService.GetBoxerByID(boxerID)
	if err != nil {
		http.Error(w, "Boxer not found", http.StatusNotFound)
		return
	}

	// Score the opponent
	score := service.ScoreOpponent(boxer, selected.Boxer)

	response := OpponentResponse{
		Opponent: selected,
		Score:    score,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// parseOpponentFilters extracts opponent filter parameters from query string
func parseOpponentFilters(query url.Values) db.OpponentFilter {
	filters := db.OpponentFilter{
		IncludeAI:     true,
		ExcludeOwned:  true,
		AvailableOnly: true,
		HealthyOnly:   true,
		MinHealth:     30.0,
		MaxResults:    50,
	}

	// Parse level_range (symmetric range around boxer's level - applied in service)
	if levelRangeStr := query.Get("level_range"); levelRangeStr != "" {
		if val, err := strconv.Atoi(levelRangeStr); err == nil && val > 0 {
			// Store as comment - actual Min/Max calculated from boxer level in service
			filters.MaxResults = val // Placeholder
		}
	}

	// Parse ai_only (if true, exclude humans)
	if query.Get("ai_only") == "true" {
		filters.IncludeAI = true
		filters.ExcludeOwned = true
	}

	// Parse humans_only (if true, exclude AI)
	if query.Get("humans_only") == "true" {
		filters.IncludeAI = false
	}

	// Parse min_health
	if minHealthStr := query.Get("min_health"); minHealthStr != "" {
		if val, err := strconv.ParseFloat(minHealthStr, 64); err == nil && val >= 0 {
			filters.MinHealth = val
		}
	}

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			if val > 100 { // Cap at 100
				val = 100
			}
			filters.MaxResults = val
		}
	}

	// Parse prefer_ranked (sort by ranking proximity)
	if query.Get("prefer_ranked") == "true" {
		filters.PreferRankings = true
	}

	return filters
}

// buildFilterMap creates a map representation of filters for API response
func buildFilterMap(filters db.OpponentFilter) map[string]interface{} {
	return map[string]interface{}{
		"include_ai":      filters.IncludeAI,
		"exclude_owned":   filters.ExcludeOwned,
		"available_only":  filters.AvailableOnly,
		"healthy_only":    filters.HealthyOnly,
		"min_health":      filters.MinHealth,
		"max_results":     filters.MaxResults,
		"prefer_rankings": filters.PreferRankings,
	}
}

// ValidateMatch handles POST /opponents/validate
// Validates a potential matchup and returns warnings/suggestions
func (h *OpponentHandler) ValidateMatch(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Boxer1ID int `json:"boxer1_id"`
		Boxer2ID int `json:"boxer2_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Boxer1ID == 0 || request.Boxer2ID == 0 {
		http.Error(w, "Both boxer1_id and boxer2_id are required", http.StatusBadRequest)
		return
	}

	// Validate the matchup
	validation := h.opponentDiscoveryService.ValidateOpponentMatch(request.Boxer1ID, request.Boxer2ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(validation)
}
