package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mormm/boxing/internal/auth"
	pkgerrors "github.com/mormm/boxing/internal/errors"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/service"
)

// RankingsHandler handles ranking-related HTTP requests
type RankingsHandler struct {
	rankingsService *service.RankingsService
}

// NewRankingsHandler creates a new RankingsHandler
func NewRankingsHandler(rankingsService *service.RankingsService) *RankingsHandler {
	return &RankingsHandler{
		rankingsService: rankingsService,
	}
}

// GetRankings handles retrieving rankings by criteria
// GET /rankings?criteria=win_rate&limit=100
func (h *RankingsHandler) GetRankings(w http.ResponseWriter, r *http.Request) {
	if h.rankingsService == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

	// Parse query parameters
	criteriaStr := r.URL.Query().Get("criteria")
	if criteriaStr == "" {
		criteriaStr = string(model.RankingByWinRate) // Default to win rate
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get rankings from service
	boxers, err := h.rankingsService.GetTopRanked(r.Context(), criteriaStr, limit)
	if err != nil {
		if err.Error() == "invalid ranking criteria: "+criteriaStr {
			pkgerrors.WriteError(w, pkgerrors.Validation("criteria", "Invalid ranking criteria"))
			return
		}
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve rankings"))
		return
	}

	// Build response
	response := &model.RankingsResponse{
		Criteria:    model.RankingCriteria(criteriaStr),
		GeneratedAt: time.Now(),
		Boxers:      boxers,
		TotalCount:  len(boxers),
	}

	pkgerrors.WriteJSON(w, http.StatusOK, response)
}

// GetRankingForBoxer handles retrieving a specific boxer's ranking
// GET /rankings/{boxer_id}?criteria=win_rate
func (h *RankingsHandler) GetRankingForBoxer(w http.ResponseWriter, r *http.Request) {
	if h.rankingsService == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

	// Parse boxer ID from URL path
	idStr := r.URL.Path[len("/rankings/"):]
	boxerID, err := strconv.Atoi(idStr)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "Invalid boxer ID"))
		return
	}

	// Parse criteria from query parameter
	criteriaStr := r.URL.Query().Get("criteria")
	if criteriaStr == "" {
		criteriaStr = string(model.RankingByWinRate) // Default to win rate
	}

	// Get boxer ranking from service
	position, err := h.rankingsService.GetRankingForBoxer(r.Context(), boxerID, criteriaStr)
	if err != nil {
		if err.Error() == "invalid ranking criteria: "+criteriaStr {
			pkgerrors.WriteError(w, pkgerrors.Validation("criteria", "Invalid ranking criteria"))
			return
		}
		if err.Error() == "failed to get boxer ranking: sql: no rows in result set" {
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
			return
		}
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve boxer ranking"))
		return
	}

	pkgerrors.WriteJSON(w, http.StatusOK, position)
}

// GetNearbyRankings handles retrieving boxers ranked near a specific boxer
// GET /rankings/nearby/{boxer_id}?criteria=win_rate&radius=5
func (h *RankingsHandler) GetNearbyRankings(w http.ResponseWriter, r *http.Request) {
	if h.rankingsService == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

	// Parse boxer ID from URL path
	idStr := r.URL.Path[len("/rankings/nearby/"):]
	boxerID, err := strconv.Atoi(idStr)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "Invalid boxer ID"))
		return
	}

	// Parse query parameters
	criteriaStr := r.URL.Query().Get("criteria")
	if criteriaStr == "" {
		criteriaStr = string(model.RankingByWinRate) // Default to win rate
	}

	radiusStr := r.URL.Query().Get("radius")
	radius := 5 // Default radius
	if radiusStr != "" {
		parsedRadius, err := strconv.Atoi(radiusStr)
		if err == nil && parsedRadius > 0 {
			radius = parsedRadius
		}
	}

	// Get nearby rankings from service
	boxers, err := h.rankingsService.GetNearbyRankings(r.Context(), boxerID, criteriaStr, radius)
	if err != nil {
		if err.Error() == "invalid ranking criteria: "+criteriaStr {
			pkgerrors.WriteError(w, pkgerrors.Validation("criteria", "Invalid ranking criteria"))
			return
		}
		if err.Error() == "failed to get nearby rankings: sql: no rows in result set" {
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
			return
		}
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve nearby rankings"))
		return
	}

	response := &model.RankingsResponse{
		Criteria:    model.RankingCriteria(criteriaStr),
		GeneratedAt: time.Now(),
		Boxers:      boxers,
		TotalCount:  len(boxers),
	}

	pkgerrors.WriteJSON(w, http.StatusOK, response)
}

// InvalidateRankings handles cache invalidation for rankings (admin endpoint)
// POST /rankings/invalidate
func (h *RankingsHandler) InvalidateRankings(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Authentication failed"))
		return
	}

	// TODO: Add admin check here

	if h.rankingsService == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

	err := h.rankingsService.InvalidateRankingsCache(r.Context())
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to invalidate rankings cache"))
		return
	}

	pkgerrors.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Rankings cache invalidated successfully",
	})
}
