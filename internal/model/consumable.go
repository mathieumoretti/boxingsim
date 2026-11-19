package model

import "time"

// ConsumableType represents the type of consumable effect
type ConsumableType string

const (
	ConsumableTypeEnergy   ConsumableType = "energy"
	ConsumableTypeHealth   ConsumableType = "health"
	ConsumableTypeRecovery ConsumableType = "recovery"
)

// Consumable represents a consumable item in the system
type Consumable struct {
	ID               int            `json:"id"`
	Name             string         `json:"name"`
	Description      *string        `json:"description"`
	ConsumableType   ConsumableType `json:"consumable_type"`
	EnergyRestore    float64        `json:"energy_restore"`
	HealthRestore    float64        `json:"health_restore"`
	FatigueReduction float64        `json:"fatigue_reduction"`
	CooldownHours    int            `json:"cooldown_hours"`
	IsUnlimited      bool           `json:"is_unlimited"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// BoxerConsumable represents a boxer's inventory of consumables
type BoxerConsumable struct {
	ID           int        `json:"id"`
	BoxerID      int        `json:"boxer_id"`
	ConsumableID int        `json:"consumable_id"`
	Quantity     int        `json:"quantity"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ConsumableUseRequest represents a request to use a consumable
type ConsumableUseRequest struct {
	ConsumableID int `json:"consumable_id" binding:"required,min=1"`
}

// ConsumableResponse represents a consumable for API responses with effective values
type ConsumableResponse struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	Description       *string `json:"description"`
	ConsumableType    string  `json:"consumable_type"`
	EnergyRestore     float64 `json:"energy_restore"`
	HealthRestore     float64 `json:"health_restore"`
	FatigueReduction  float64 `json:"fatigue_reduction"`
	CooldownHours     int     `json:"cooldown_hours"`
	IsUnlimited       bool    `json:"is_unlimited"`
	AvailableNow      *bool   `json:"available_now,omitempty"`      // For inventory items, checks cooldown
	RemainingQuantity *int    `json:"remaining_quantity,omitempty"` // Current quantity in inventory
}
