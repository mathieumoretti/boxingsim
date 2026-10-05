package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	boxerdb "github.com/mormm/boxing/internal/db"
	pkgerrors "github.com/mormm/boxing/internal/errors"
	"github.com/mormm/boxing/internal/service"
)

type FightHandler struct {
	fightService *service.FightService
}

func NewFightHandler(fightService *service.FightService) *FightHandler {
	return &FightHandler{fightService: fightService}
}

// BookFightForBoxerRequest represents the JSON body for booking a fight for a specific boxer
type BookFightForBoxerRequest struct {
	OpponentID int `json:"opponent_id" binding:"required"`
}

// BookFightForBoxer schedules a fight for a specific boxer with an opponent (POST /boxers/{id}/fights)
func (h *FightHandler) BookFightForBoxer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boxerIDStr := vars["id"]

	if boxerIDStr == "" {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("boxer id required"))
		return
	}

	boxerID, err := strconv.Atoi(boxerIDStr)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "invalid boxer id format"))
		return
	}

	var req BookFightForBoxerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("invalid request body"))
		return
	}

	if req.OpponentID == 0 {
		pkgerrors.WriteError(w, pkgerrors.Validation("opponent_id", "opponent_id is required"))
		return
	}

	// Schedule fight for current game time (immediate booking)
	scheduledTime := time.Now()

	err = h.fightService.BookFight(r.Context(), boxerID, req.OpponentID, scheduledTime, 0)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrBoxerNotExists):
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
		default:
			pkgerrors.WriteError(w, pkgerrors.Conflict(err.Error()))
		}
		return
	}

	pkgerrors.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Fight booked successfully",
	})
}

// BookFightRequest represents the JSON body for booking a fight
type BookFightRequest struct {
	Boxer1ID      int    `json:"boxer1_id" binding:"required"`
	Boxer2ID      int    `json:"boxer2_id" binding:"required"`
	ScheduledTime string `json:"scheduled_time" binding:"required"`
	Round         int    `json:"round,omitempty"`
}

// BookFight schedules a new fight between two boxers (POST /fights/book)
func (h *FightHandler) BookFight(w http.ResponseWriter, r *http.Request) {
	var req BookFightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("invalid request body"))
		return
	}

	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledTime)
	if err != nil {
		scheduledTime = time.Now()
	}

	err = h.fightService.BookFight(r.Context(), req.Boxer1ID, req.Boxer2ID, scheduledTime, req.Round)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrBoxerNotExists):
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
		default:
			pkgerrors.WriteError(w, pkgerrors.Conflict(err.Error()))
		}
		return
	}

	pkgerrors.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "Fight booked successfully",
	})
}

// GetActiveFights returns all active (scheduled/in_progress) fights (GET /fights/active)
func (h *FightHandler) GetActiveFights(w http.ResponseWriter, r *http.Request) {
	statuses := []string{"scheduled", "in_progress"}
	fights, err := h.fightService.GetActiveFights(r.Context(), statuses)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to fetch active fights"))
		return
	}

	pkgerrors.WriteJSON(w, http.StatusOK, fights)
}

// GetFightByID returns a specific fight by ID (GET /fights/{id})
func (h *FightHandler) GetFightByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	if idStr == "" {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("fight id required"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "invalid fight id format"))
		return
	}

	fight, err := h.fightService.GetFightByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrBoxerNotExists):
			pkgerrors.WriteError(w, pkgerrors.NotFound("fight"))
		default:
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to fetch fight"))
		}
		return
	}

	pkgerrors.WriteJSON(w, http.StatusOK, fight)
}

// GetUpcomingFightForBoxer returns the next upcoming fight for a specific boxer (GET /boxers/{id}/upcoming-fight) (MAT-106)
func (h *FightHandler) GetUpcomingFightForBoxer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	if idStr == "" {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("boxer id required"))
		return
	}

	boxerID, err := strconv.Atoi(idStr)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "invalid boxer id format"))
		return
	}

	fight, err := h.fightService.GetUpcomingFightForBoxer(r.Context(), boxerID)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrUpcomingFightNotFound):
			pkgerrors.WriteError(w, pkgerrors.NotFound("upcoming fight"))
		default:
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to fetch upcoming fight"))
		}
		return
	}

	pkgerrors.WriteJSON(w, http.StatusOK, fight)
}

// GetFightHistoryWithOpponents returns the fight history for a specific boxer with opponent names (GET /boxers/{id}/fights-history) (MAT-103)
func (h *FightHandler) GetFightHistoryWithOpponents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	if idStr == "" {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("boxer id required"))
		return
	}

	boxerID, err := strconv.Atoi(idStr)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "invalid boxer id format"))
		return
	}

	fights, err := h.fightService.GetFightHistoryWithOpponents(r.Context(), boxerID)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrBoxerNotExists):
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
		default:
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to fetch fight history"))
		}
		return
	}

	// Return empty array if no fights found
	if fights == nil {
		fights = []*boxerdb.FightHistoryWithOpponent{}
	}

	pkgerrors.WriteJSON(w, http.StatusOK, fights)
}
