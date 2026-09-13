package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/store"
)

// BoxerHandler handles boxer-related HTTP requests
type BoxerHandler struct {
	boxerStore          *store.BoxerStore
	scheduledEventStore *store.ScheduledEventStore
}

func NewBoxerHandler(boxerStore *store.BoxerStore, scheduledEventStore *store.ScheduledEventStore) *BoxerHandler {
	return &BoxerHandler{
		boxerStore:          boxerStore,
		scheduledEventStore: scheduledEventStore,
	}
}

// enrichBoxerResponse converts a Boxer entity to BoxerResponse and adds rest status fields (MAT-89)
func (h *BoxerHandler) enrichBoxerResponse(ctx context.Context, boxer *model.Boxer) (*model.BoxerResponse, error) {
	response := &model.BoxerResponse{
		ID:                    boxer.ID,
		UserID:                boxer.UserID,
		Name:                  boxer.Name,
		Nickname:              boxer.Nickname,
		PositionX:             boxer.PositionX,
		PositionY:             boxer.PositionY,
		Health:                boxer.Health,
		Energy:                boxer.Energy,
		Strength:              boxer.Strength,
		Defense:               boxer.Defense,
		Agility:               boxer.Agility,
		Experience:            boxer.Experience,
		Level:                 boxer.Level,
		FatigueScore:          boxer.FatigueScore,
		ForcedRestUntil:       boxer.ForcedRestUntil,
		HasActiveRest:         false,
		NextAvailableTraining: nil,
		CreatedAt:             boxer.CreatedAt,
		UpdatedAt:             boxer.UpdatedAt,
	}

	now := time.Now()

	// Check for active rest events (MAT-89)
	if h.scheduledEventStore != nil {
		pendingRest, err := h.scheduledEventStore.GetPendingByBoxerIDAndType(ctx, boxer.ID, model.EventTypeRest)
		if err == nil && len(pendingRest) > 0 {
			for _, event := range pendingRest {
				if !event.Processed && event.EventTime.After(now) {
					response.HasActiveRest = true
					response.NextAvailableTraining = &event.EventTime
					break
				}
			}
		}
	}

	// Override with forced rest if applicable (more restrictive)
	if boxer.ForcedRestUntil != nil && now.Before(*boxer.ForcedRestUntil) {
		response.HasActiveRest = true
		response.NextAvailableTraining = boxer.ForcedRestUntil
	}

	return response, nil
}

// enrichBoxersResponse enriches a slice of Boxer entities with rest status fields (MAT-89)
func (h *BoxerHandler) enrichBoxersResponse(ctx context.Context, boxers []*model.Boxer) ([]*model.BoxerResponse, error) {
	responses := make([]*model.BoxerResponse, 0, len(boxers))
	for _, boxer := range boxers {
		response, err := h.enrichBoxerResponse(ctx, boxer)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

// CreateBoxer handles creating a new boxer
func (h *BoxerHandler) CreateBoxer(w http.ResponseWriter, r *http.Request) {
	var boxerCreate model.BoxerCreate
	if err := json.NewDecoder(r.Body).Decode(&boxerCreate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		http.Error(w, `{"error": "Authentication failed"}`, http.StatusUnauthorized)
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
		UserID:     user.ID,
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

	err := h.boxerStore.Create(r.Context(), boxer)
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

	// Enrich boxer response with rest status (MAT-89)
	enrichedResponse, err := h.enrichBoxerResponse(r.Context(), boxer)
	if err != nil {
		http.Error(w, "Failed to enrich boxer response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(enrichedResponse)
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

// GetBoxersByUserID handles retrieving all boxers for the authenticated user
func (h *BoxerHandler) GetBoxersByUserID(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		http.Error(w, `{"error": "Authentication failed"}`, http.StatusUnauthorized)
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
	boxers, err := h.boxerStore.GetByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve boxers", http.StatusInternalServerError)
		return
	}

	// Enrich boxer responses with rest status (MAT-89)
	enrichedBoxers, err := h.enrichBoxersResponse(r.Context(), boxers)
	if err != nil {
		http.Error(w, "Failed to enrich boxer responses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(enrichedBoxers)
}
