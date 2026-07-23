package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/store"
)

// BoxerHandler handles boxer-related HTTP requests
type BoxerHandler struct {
	boxerStore *store.BoxerStore
}

func NewBoxerHandler(boxerStore *store.BoxerStore) *BoxerHandler {
	return &BoxerHandler{
		boxerStore: boxerStore,
	}
}

// CreateBoxer handles creating a new boxer
func (h *BoxerHandler) CreateBoxer(w http.ResponseWriter, r *http.Request) {
	var boxerCreate model.BoxerCreate
	if err := json.NewDecoder(r.Body).Decode(&boxerCreate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert user-id from header to int
	userIDStr := r.Header.Get("user-id")
	if userIDStr == "" {
		http.Error(w, "User ID header is required - make sure the frontend sets the 'user-id' header", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Validate the boxer creation request
	if boxerCreate.Name == "" {
		http.Error(w, "Boxer name is required", http.StatusBadRequest)
		return
	}

	if boxerCreate.Strength < 0 || boxerCreate.Defense < 0 || boxerCreate.Agility < 0 {
		http.Error(w, "Strength, defense, and agility must be non-negative", http.StatusBadRequest)
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		http.Error(w, "Database connection not available", http.StatusServiceUnavailable)
		return
	}

	// Create the boxer in the database using the boxerStore
	boxer := &model.Boxer{
		UserID:     userID,
		Name:       boxerCreate.Name,
		Nickname:   boxerCreate.Nickname,
		PositionX:  boxerCreate.PositionX,
		PositionY:  boxerCreate.PositionY,
		Health:     100.0,
		Energy:     100.0,
		Strength:   boxerCreate.Strength,
		Defense:    boxerCreate.Defense,
		Agility:    boxerCreate.Agility,
		Experience: 0.0,
		Level:      1,
	}

	err = h.boxerStore.Create(r.Context(), boxer)
	if err != nil {
		http.Error(w, "Failed to create boxer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Boxer created successfully",
		"boxer":   boxer,
	})
}

// GetBoxer handles retrieving a boxer by ID
func (h *BoxerHandler) GetBoxer(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL path
	idStr := r.URL.Path[len("/boxers/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid boxer ID", http.StatusBadRequest)
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		http.Error(w, "Database connection not available", http.StatusServiceUnavailable)
		return
	}

	boxer, getErr := h.boxerStore.GetByID(r.Context(), id)
	if getErr != nil {
		if getErr.Error() == "no rows in result set" {
			http.Error(w, "Boxer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve boxer", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(boxer)
}

// UpdateBoxer handles updating a boxer
func (h *BoxerHandler) UpdateBoxer(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL path
	idStr := r.URL.Path[len("/boxers/"):]
	id, parseErr := strconv.Atoi(idStr)
	if parseErr != nil {
		http.Error(w, "Invalid boxer ID", http.StatusBadRequest)
		return
	}

	var boxerUpdate model.BoxerUpdate
	if decodeErr := json.NewDecoder(r.Body).Decode(&boxerUpdate); decodeErr != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		http.Error(w, "Database connection not available", http.StatusServiceUnavailable)
		return
	}

	// Get the existing boxer to update
	boxer, getErr := h.boxerStore.GetByID(r.Context(), id)
	if getErr != nil {
		if getErr.Error() == "no rows in result set" {
			http.Error(w, "Boxer not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve boxer", http.StatusInternalServerError)
		}
		return
	}

	// Update the boxer fields with provided values or keep existing ones
	if boxerUpdate.Name != nil {
		boxer.Name = *boxerUpdate.Name
	}
	if boxerUpdate.Nickname != nil {
		boxer.Nickname = boxerUpdate.Nickname
	}
	if boxerUpdate.PositionX != nil {
		boxer.PositionX = *boxerUpdate.PositionX
	}
	if boxerUpdate.PositionY != nil {
		boxer.PositionY = *boxerUpdate.PositionY
	}
	if boxerUpdate.Strength != nil {
		boxer.Strength = *boxerUpdate.Strength
	}
	if boxerUpdate.Defense != nil {
		boxer.Defense = *boxerUpdate.Defense
	}
	if boxerUpdate.Agility != nil {
		boxer.Agility = *boxerUpdate.Agility
	}

	// Update the boxer in the database
	updateErr := h.boxerStore.Update(r.Context(), boxer)
	if updateErr != nil {
		http.Error(w, "Failed to update boxer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Boxer updated successfully",
		"id":      id,
	})
}

// GetBoxersByUserID handles retrieving all boxers for a specific user
func (h *BoxerHandler) GetBoxersByUserID(w http.ResponseWriter, r *http.Request) {
	// Parse user ID from URL path
	userIDStr := r.URL.Path[len("/users/"):]
	userIDStr = userIDStr[:len(userIDStr)-len("/boxers")]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		// Return empty array if no database connection
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]model.Boxer{})
		return
	}

	// Get the boxers from the database using the boxerStore
	boxers, err := h.boxerStore.GetByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to retrieve boxers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(boxers)
}
