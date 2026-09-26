package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	boxerdb "github.com/mormm/boxing/internal/db"
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
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	boxerIDStr := vars["id"]

	if boxerIDStr == "" {
		http.Error(w, `{"error": "boxer id required"}`, http.StatusBadRequest)
		return
	}

	boxerID, err := strconv.Atoi(boxerIDStr)
	if err != nil {
		http.Error(w, `{"error": "invalid boxer id format"}`, http.StatusBadRequest)
		return
	}

	var req BookFightForBoxerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.OpponentID == 0 {
		http.Error(w, `{"error": "opponent_id is required"}`, http.StatusBadRequest)
		return
	}

	// Schedule fight for current game time (immediate booking)
	scheduledTime := time.Now()

	err = h.fightService.BookFight(r.Context(), boxerID, req.OpponentID, scheduledTime, 0)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrBoxerNotExists):
			http.Error(w, `{"error": "boxer not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusConflict)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Fight booked successfully",
	}); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
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
	w.Header().Set("Content-Type", "application/json")

	var req BookFightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
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
			http.Error(w, `{"error": "boxer not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusConflict)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Fight booked successfully"}); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
}

// GetActiveFights returns all active (scheduled/in_progress) fights (GET /fights/active)
func (h *FightHandler) GetActiveFights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	statuses := []string{"scheduled", "in_progress"}
	fights, err := h.fightService.GetActiveFights(r.Context(), statuses)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(fights); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
}

// GetFightByID returns a specific fight by ID (GET /fights/{id})
func (h *FightHandler) GetFightByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]

	if idStr == "" {
		http.Error(w, `{"error": "fight id required"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid fight id format"}`, http.StatusBadRequest)
		return
	}

	fight, err := h.fightService.GetFightByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, boxerdb.ErrBoxerNotExists):
			http.Error(w, `{"error": "fight not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(fight); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
}
