package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mormm/boxing/internal/auth"
	pkgerrors "github.com/mormm/boxing/internal/errors"
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

	// Check for active rest events (MAT-89, MAT-96)
	// If there's any unprocessed rest event, the boxer is currently resting
	if h.scheduledEventStore != nil {
		pendingRest, err := h.scheduledEventStore.GetPendingByBoxerIDAndType(ctx, boxer.ID, model.EventTypeRest)
		if err == nil && len(pendingRest) > 0 {
			response.HasActiveRest = true
			if len(pendingRest) > 0 {
				response.NextAvailableTraining = &pendingRest[0].EventTime
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
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid JSON"))
		return
	}

	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Authentication failed"))
		return
	}

	// Validate the boxer creation request
	if boxerCreate.Name == "" {
		pkgerrors.WriteError(w, pkgerrors.Validation("name", "Boxer name is required"))
		return
	}

	if boxerCreate.Strength < 0 || boxerCreate.Defense < 0 || boxerCreate.Agility < 0 {
		pkgerrors.WriteError(w, pkgerrors.Validation("stats", "Strength, defense, and agility must be non-negative"))
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
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
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to create boxer"))
		return
	}

	pkgerrors.WriteSuccess(w, http.StatusCreated, map[string]any{
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
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "Invalid boxer ID"))
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

	boxer, getErr := h.boxerStore.GetByID(r.Context(), id)
	if getErr != nil {
		if getErr.Error() == "no rows in result set" {
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
		} else {
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve boxer"))
		}
		return
	}

	// Enrich boxer response with rest status (MAT-89)
	enrichedResponse, err := h.enrichBoxerResponse(r.Context(), boxer)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to enrich boxer response"))
		return
	}

	pkgerrors.WriteJSON(w, http.StatusOK, enrichedResponse)
}

// UpdateBoxer handles updating a boxer
func (h *BoxerHandler) UpdateBoxer(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL path
	idStr := r.URL.Path[len("/boxers/"):]
	id, parseErr := strconv.Atoi(idStr)
	if parseErr != nil {
		pkgerrors.WriteError(w, pkgerrors.Validation("id", "Invalid boxer ID"))
		return
	}

	var boxerUpdate model.BoxerUpdate
	if decodeErr := json.NewDecoder(r.Body).Decode(&boxerUpdate); decodeErr != nil {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid JSON"))
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

	// Get the existing boxer to update
	boxer, getErr := h.boxerStore.GetByID(r.Context(), id)
	if getErr != nil {
		if getErr.Error() == "no rows in result set" {
			pkgerrors.WriteError(w, pkgerrors.NotFound("boxer"))
		} else {
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve boxer"))
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
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to update boxer"))
		return
	}

	pkgerrors.WriteSuccess(w, http.StatusOK, map[string]any{
		"message": "Boxer updated successfully",
		"id":      id,
	})
}

// GetBoxersByUserID handles retrieving all boxers for the authenticated user
func (h *BoxerHandler) GetBoxersByUserID(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Authentication failed"))
		return
	}

	// Check if database connection is available
	if h.boxerStore == nil {
		// Return empty array if no database connection
		pkgerrors.WriteJSON(w, http.StatusOK, []*model.BoxerResponse{})
		return
	}

	// Get the boxers from the database using the boxerStore
	boxers, err := h.boxerStore.GetByUserID(r.Context(), user.ID)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve boxers"))
		return
	}

	// Enrich boxer responses with rest status (MAT-89)
	enrichedBoxers, err := h.enrichBoxersResponse(r.Context(), boxers)
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to enrich boxer responses"))
		return
	}

	pkgerrors.WriteJSON(w, http.StatusOK, enrichedBoxers)
}
