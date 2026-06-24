package handler

import (
	"boxing/internal/db"
	"boxing/internal/model"
	"boxing/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
)

type BoxerHandler struct {
	boxerService *service.BoxerService
}

// NewBoxerHandler creates a new BoxerHandler
func NewBoxerHandler(boxerService *service.BoxerService) *BoxerHandler {
	return &BoxerHandler{boxerService: boxerService}
}

// CreateBoxerRequest represents the request body for creating a boxer
type CreateBoxerRequest struct {
	Name        string  `json:"name"`
	Class       string  `json:"class"`
	MaxHealth   float64 `json:"max_health"`
	MaxEnergy   float64 `json:"max_energy"`
	Strength    float64 `json:"strength"`
	Defense     float64 `json:"defense"`
	Agility     float64 `json:"agility"`
	SkillPoints float64 `json:"skill_points"`
}

// UpdateBoxerRequest represents the request body for updating a boxer
type UpdateBoxerRequest struct {
	Name        *string  `json:"name"`
	Class       *string  `json:"class"`
	Health      *float64 `json:"health"`
	Energy      *float64 `json:"energy"`
	PositionX   *float64 `json:"position_x"`
	PositionY   *float64 `json:"position_y"`
	Strength    *float64 `json:"strength"`
	Defense     *float64 `json:"defense"`
	Agility     *float64 `json:"agility"`
	SkillPoints *float64 `json:"skill_points"`
	Level       *int     `json:"level"`
	XP          *float64 `json:"xp"`
}

// CreateBoxer handles creating a new boxer
func (h *BoxerHandler) CreateBoxer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req CreateBoxerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	boxer := &model.Boxer{
		Name:        req.Name,
		Class:       req.Class,
		MaxHealth:   req.MaxHealth,
		MaxEnergy:   req.MaxEnergy,
		Health:      req.MaxHealth,
		Energy:      req.MaxEnergy,
		Strength:    req.Strength,
		Defense:     req.Defense,
		Agility:     req.Agility,
		SkillPoints: req.SkillPoints,
		Level:       1,
		XP:          0,
		Experience:  0,
		Status:      "active",
	}

	err = h.boxerService.CreateBoxer(boxer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(boxer)
}

// GetBoxer handles retrieving a boxer by ID
func (h *BoxerHandler) GetBoxer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid boxer ID"})
		return
	}

	boxer, err := h.boxerService.GetBoxerByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxer)
}

// GetBoxers handles retrieving all boxers
func (h *BoxerHandler) GetBoxers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	boxers, err := h.boxerService.GetAllBoxers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxers)
}

// UpdateBoxer handles updating a boxer
func (h *BoxerHandler) UpdateBoxer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid boxer ID"})
		return
	}

	var req UpdateBoxerRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	boxer, err := h.boxerService.UpdateBoxer(id, req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxer)
}

// DeleteBoxer handles deleting a boxer
func (h *BoxerHandler) DeleteBoxer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid boxer ID"})
		return
	}

	err = h.boxerService.DeleteBoxer(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateBoxerPosition handles updating a boxer's position
func (h *BoxerHandler) UpdateBoxerPosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid boxer ID"})
		return
	}

	var req struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	boxer, err := h.boxerService.UpdateBoxerPosition(id, req.X, req.Y)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxer)
}

// FightRequest represents the request body for starting a fight
type FightRequest struct {
	Boxer1ID int    `json:"boxer1_id"`
	Boxer2ID int    `json:"boxer2_id"`
	Location string `json:"location"`
}

// StartFight handles starting a fight
func (h *BoxerHandler) StartFight(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req FightRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Boxer1ID == req.Boxer2ID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer 1 and Boxer 2 must be different"})
		return
	}

	boxer1, err := h.boxerService.GetBoxerByID(req.Boxer1ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer 1 not found"})
		return
	}

	boxer2, err := h.boxerService.GetBoxerByID(req.Boxer2ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Boxer 2 not found"})
		return
	}

	if boxer1.Status == "idle" || boxer2.Status == "idle" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Both boxers must be active to fight"})
		return
	}

	fight, err := h.boxerService.StartFight(req.Boxer1ID, req.Boxer2ID, req.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fight)
}