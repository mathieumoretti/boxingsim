package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mormm/boxing/internal/model"
)

// FightHandler handles fight-related HTTP requests
type FightHandler struct{}

func NewFightHandler() *FightHandler {
	return &FightHandler{}
}

// BookFight schedules a new fight between two boxers (POST /fights/book)
func (h *FightHandler) BookFight(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Not implemented yet",
	})
}

// GetActiveFights returns all active (scheduled/in_progress) fights (GET /fights/active)
func (h *FightHandler) GetActiveFights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]model.Fight{})
}
