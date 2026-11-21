package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/domain"
	pkgerrors "github.com/mormm/boxing/internal/errors"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/service"
	"github.com/mormm/boxing/internal/store"
)

// ConsumableHandler handles consumable-related HTTP requests (MAT-153 Phase 6)
type ConsumableHandler struct {
	boxerStore      *store.BoxerStore
	consumableStore *store.ConsumableStore
	recoveryService *service.RecoveryService
}

// NewConsumableHandler creates a new ConsumableHandler instance
func NewConsumableHandler(
	boxerStore *store.BoxerStore,
	consumableStore *store.ConsumableStore,
	recoveryService *service.RecoveryService,
) *ConsumableHandler {
	return &ConsumableHandler{
		boxerStore:      boxerStore,
		consumableStore: consumableStore,
		recoveryService: recoveryService,
	}
}

// GetAllConsumables returns all available consumables (GET /consumables)
func (h *ConsumableHandler) GetAllConsumables(w http.ResponseWriter, r *http.Request) {
	if h.consumableStore == nil {
		pkgerrors.WriteJSON(w, http.StatusServiceUnavailable, []*model.Consumable{})
		return
	}

	consumables, err := h.consumableStore.GetAll(r.Context())
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve consumables"))
		return
	}

	if consumables == nil {
		consumables = []*model.Consumable{}
	}

	pkgerrors.WriteJSON(w, http.StatusOK, consumables)
}

// UseConsumableRequest represents the request body for using a consumable
type UseConsumableRequest struct {
	ConsumableID int `json:"consumable_id" binding:"required"`
}

// UseConsumableResponse represents the response after using a consumable
type UseConsumableResponse struct {
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	ConsumeResult *ConsumedItem `json:"consumed_item,omitempty"`
	CooldownInfo  *CooldownInfo `json:"cooldown_info,omitempty"`
}

// ConsumedItem represents the consumable that was used
type ConsumedItem struct {
	ID               int                  `json:"id"`
	Name             string               `json:"name"`
	Description      *string              `json:"description"`
	ConsumableType   model.ConsumableType `json:"consumable_type"`
	EnergyRestore    float64              `json:"energy_restore"`
	HealthRestore    float64              `json:"health_restore"`
	FatigueReduction float64              `json:"fatigue_reduction"`
	CooldownHours    int                  `json:"cooldown_hours"`
	IsUnlimited      bool                 `json:"is_unlimited"`
}

// CooldownInfo represents cooldown information for a consumable
type CooldownInfo struct {
	OnCooldown  bool   `json:"on_cooldown"`
	AvailableAt string `json:"available_at,omitempty"`
	Message     string `json:"message"`
}

// UseConsumable applies a consumable to a boxer (POST /boxers/{id}/use-consumable)
func (h *ConsumableHandler) UseConsumable(w http.ResponseWriter, r *http.Request) {
	if h.boxerStore == nil || h.consumableStore == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

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

	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Authentication failed"))
		return
	}

	ctx := r.Context()

	// Get boxer to verify ownership and existence
	boxer, err := h.boxerStore.GetByID(ctx, domain.BoxerID(boxerID))
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.NotFound("Boxer not found"))
		return
	}

	if boxer.UserID != user.ID {
		pkgerrors.WriteError(w, pkgerrors.Forbidden("You do not own this boxer"))
		return
	}

	// Parse request body
	var req UseConsumableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid request body"))
		return
	}

	if req.ConsumableID == 0 {
		pkgerrors.WriteError(w, pkgerrors.Validation("consumable_id", "consumable_id is required"))
		return
	}

	// Get consumable details
	consumable, err := h.consumableStore.GetByID(ctx, req.ConsumableID)
	if err != nil {
		if err == store.ErrConsumableNotFound {
			pkgerrors.WriteError(w, pkgerrors.NotFound("Consumable not found"))
			return
		}
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve consumable"))
		return
	}

	// Check cooldown if applicable
	if !consumable.IsUnlimited {
		onCooldown, availableAt, err := h.consumableStore.CheckCooldown(ctx, domain.BoxerID(boxerID), consumable)
		if err != nil {
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to check cooldown"))
			return
		}

		if onCooldown {
			response := UseConsumableResponse{
				Success: false,
				Message: "Consumable is on cooldown",
				CooldownInfo: &CooldownInfo{
					OnCooldown:  true,
					AvailableAt: availableAt.Format(time.RFC3339),
					Message:     "This consumable will be available again at the specified time",
				},
			}
			pkgerrors.WriteJSON(w, http.StatusOK, response)
			return
		}
	}

	// Use the consumable (decrement inventory if not unlimited)
	if !consumable.IsUnlimited {
		err = h.consumableStore.UseConsumable(ctx, domain.BoxerID(boxerID), req.ConsumableID)
		if err != nil {
			switch err {
			case store.ErrInsufficientQuantity:
				pkgerrors.WriteError(w, pkgerrors.Conflict("Insufficient quantity of this consumable"))
				return
			default:
				pkgerrors.WriteError(w, pkgerrors.Internal("Failed to use consumable"))
				return
			}
		}
	}

	// Apply the consumable effects to the boxer
	if h.recoveryService != nil {
		err = h.recoveryService.ApplyConsumable(ctx, domain.BoxerID(boxerID), consumable)
		if err != nil {
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to apply consumable effects"))
			return
		}

		// Construct response
		consumedItem := &ConsumedItem{
			ID:               consumable.ID,
			Name:             consumable.Name,
			Description:      consumable.Description,
			ConsumableType:   consumable.ConsumableType,
			EnergyRestore:    consumable.EnergyRestore,
			HealthRestore:    consumable.HealthRestore,
			FatigueReduction: consumable.FatigueReduction,
			CooldownHours:    consumable.CooldownHours,
			IsUnlimited:      consumable.IsUnlimited,
		}

		response := UseConsumableResponse{
			Success:       true,
			Message:       "Consumable used successfully",
			ConsumeResult: consumedItem,
		}

		if !consumable.IsUnlimited && consumable.CooldownHours > 0 {
			cooldownEnd := time.Now().Add(time.Duration(consumable.CooldownHours) * time.Hour)
			response.CooldownInfo = &CooldownInfo{
				OnCooldown:  false,
				AvailableAt: cooldownEnd.Format(time.RFC3339),
				Message:     "Consumable will be on cooldown until the specified time",
			}
		}

		pkgerrors.WriteJSON(w, http.StatusOK, response)
		return
	}

	pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
}

// GetBoxerInventory returns all consumables in a boxer's inventory (GET /boxers/{id}/consumables)
type InventoryItem struct {
	ID               int                  `json:"id"`
	Name             string               `json:"name"`
	Description      *string              `json:"description"`
	ConsumableType   model.ConsumableType `json:"consumable_type"`
	EnergyRestore    float64              `json:"energy_restore"`
	HealthRestore    float64              `json:"health_restore"`
	FatigueReduction float64              `json:"fatigue_reduction"`
	CooldownHours    int                  `json:"cooldown_hours"`
	IsUnlimited      bool                 `json:"is_unlimited"`
	Quantity         int                  `json:"quantity"`
	LastUsedAt       *time.Time           `json:"last_used_at,omitempty"`
	OnCooldown       bool                 `json:"on_cooldown,omitempty"`
	AvailableAt      *string              `json:"available_at,omitempty"`
}

func (h *ConsumableHandler) GetBoxerInventory(w http.ResponseWriter, r *http.Request) {
	if h.boxerStore == nil || h.consumableStore == nil {
		pkgerrors.WriteError(w, pkgerrors.ServiceUnavailable())
		return
	}

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

	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Authentication failed"))
		return
	}

	ctx := r.Context()

	// Get boxer to verify ownership
	boxer, err := h.boxerStore.GetByID(ctx, domain.BoxerID(boxerID))
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.NotFound("Boxer not found"))
		return
	}

	if boxer.UserID != user.ID {
		pkgerrors.WriteError(w, pkgerrors.Forbidden("You do not own this boxer"))
		return
	}

	// Get inventory
	inventory, err := h.consumableStore.GetBoxerInventory(ctx, domain.BoxerID(boxerID))
	if err != nil {
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to retrieve inventory"))
		return
	}

	var inventoryItems []*InventoryItem

	if inventory == nil {
		pkgerrors.WriteJSON(w, http.StatusOK, []*InventoryItem{})
		return
	}

	for _, item := range inventory {
		consumable, err := h.consumableStore.GetByID(ctx, item.ConsumableID)
		if err != nil {
			// Skip items with missing consumable definitions
			continue
		}

		inventoryItem := &InventoryItem{
			ID:               consumable.ID,
			Name:             consumable.Name,
			Description:      consumable.Description,
			ConsumableType:   consumable.ConsumableType,
			EnergyRestore:    consumable.EnergyRestore,
			HealthRestore:    consumable.HealthRestore,
			FatigueReduction: consumable.FatigueReduction,
			CooldownHours:    consumable.CooldownHours,
			IsUnlimited:      consumable.IsUnlimited,
			Quantity:         item.Quantity,
			LastUsedAt:       item.LastUsedAt,
		}

		// Check cooldown status for limited items
		if !consumable.IsUnlimited && consumable.CooldownHours > 0 {
			onCooldown, availableAt, _ := h.consumableStore.CheckCooldown(ctx, domain.BoxerID(boxerID), consumable)
			inventoryItem.OnCooldown = onCooldown
			if onCooldown && availableAt != nil {
				availableStr := availableAt.Format(time.RFC3339)
				inventoryItem.AvailableAt = &availableStr
			}
		}

		inventoryItems = append(inventoryItems, inventoryItem)
	}

	if inventoryItems == nil {
		inventoryItems = []*InventoryItem{}
	}

	pkgerrors.WriteJSON(w, http.StatusOK, inventoryItems)
}
