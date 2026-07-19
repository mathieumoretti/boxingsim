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
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would:
	// 1. Validate the request
	// 2. Create the boxer in the database using the boxerStore
	// For now, we'll just return a stub response

	// Simulate creating a boxer with the user ID
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

	// In a real implementation, we would:
	// 1. Get the boxer from the database using the boxerStore
	// For now, we'll just return a stub response

	boxer := &model.Boxer{
		ID:         id,
		UserID:     1, // Placeholder - in real app this should come from DB
		Name:       "Test Boxer",
		Nickname:   nil,
		PositionX:  0.0,
		PositionY:  0.0,
		Health:     100.0,
		Energy:     100.0,
		Strength:   10.0,
		Defense:    10.0,
		Agility:    10.0,
		Experience: 0.0,
		Level:      1,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(boxer)
}

// UpdateBoxer handles updating a boxer
func (h *BoxerHandler) UpdateBoxer(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL path
	idStr := r.URL.Path[len("/boxers/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid boxer ID", http.StatusBadRequest)
		return
	}

	var boxerUpdate model.BoxerUpdate
	if err := json.NewDecoder(r.Body).Decode(&boxerUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would:
	// 1. Validate the request
	// 2. Update the boxer in the database using the boxerStore
	// For now, we'll just return a stub response

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
