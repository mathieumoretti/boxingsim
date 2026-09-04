package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/store"
)

// TrainingHandler handles training-related HTTP requests
type TrainingHandler struct {
	boxerStore         *store.BoxerStore
	trainingTypeStore  *store.TrainingTypeStore
	trainingSessionStore *store.TrainingSessionStore
	scheduledEventStore  *store.ScheduledEventStore
}

// NewTrainingHandler creates a new TrainingHandler instance
func NewTrainingHandler(
	boxerStore *store.BoxerStore,
	trainingTypeStore *store.TrainingTypeStore,
	trainingSessionStore *store.TrainingSessionStore,
	scheduledEventStore *store.ScheduledEventStore,
) *TrainingHandler {
	return &TrainingHandler{
		boxerStore:         boxerStore,
		trainingTypeStore:  trainingTypeStore,
		trainingSessionStore: trainingSessionStore,
		scheduledEventStore:  scheduledEventStore,
	}
}

// GetAllTrainingTypes returns all available training types (reference data)
func (h *TrainingHandler) GetAllTrainingTypes(w http.ResponseWriter, r *http.Request) {
	if h.trainingTypeStore == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode([]model.TrainingType{})
		return
	}

	trainingTypes, err := h.trainingTypeStore.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve training types", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(trainingTypes)
}

// ScheduleTraining handles scheduling a new training session for a boxer
// POST /boxers/{id}/train
type ScheduleTrainingRequest struct {
	TrainingTypeID int     `json:"training_type_id" binding:"required,min=1"`
	DurationHours  float64 `json:"duration_hours" binding:"required,gt=0,lte=8"`
	ScheduledAt    string  `json:"scheduled_at" binding:"required"`
}

func (h *TrainingHandler) ScheduleTraining(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context (injected by middleware)
	user := auth.UserFromRequest(r)
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Authentication failed"})
		return
	}

	// Parse boxer ID from URL path
	idStr := r.URL.Path[len("/boxers/"):]
	boxerID, parseErr := strconv.Atoi(idStr)
	if parseErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid boxer ID"})
		return
	}

	// Decode request body
	var req ScheduleTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Check if stores are available
	if h.boxerStore == nil || h.trainingTypeStore == nil || h.trainingSessionStore == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Database connection not available"})
		return
	}

	ctx := r.Context()

	// 1. Validate boxer exists and belongs to user
	boxer, err := h.boxerStore.GetByID(ctx, boxerID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Boxer not found"})
		return
	}

	if boxer.UserID != user.ID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "You don't own this boxer"})
		return
	}

	// 2. Validate training type exists
	trainingType, err := h.trainingTypeStore.GetByID(ctx, req.TrainingTypeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Training type not found"})
		return
	}

	// 3. Validate boxer has sufficient energy
	energyCost := trainingType.EnergyCost * int(req.DurationHours)
	if int(boxer.Energy) < energyCost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":       "Insufficient energy",
			"required":    strconv.Itoa(energyCost),
			"available":   strconv.Itoa(int(boxer.Energy)),
		})
		return
	}

	// 4. Validate boxer is not in recovery (health < 50%)
	if boxer.Health < 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Boxer needs to recover before training (health < 50%)"})
		return
	}

	// 5. Validate duration constraints (1-8 hours, already enforced by binding tag)
	if req.DurationHours < 1 || req.DurationHours > 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Duration must be between 1 and 8 hours"})
		return
	}

	// 6. Check for pending training sessions (boxer can only do one training at a time)
	pendingSessions, err := h.trainingSessionStore.GetPendingByBoxerID(ctx, boxerID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to check pending training sessions"})
		return
	}

	if len(pendingSessions) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Boxer already has a pending training session"})
		return
	}

	// Calculate planned gains based on duration and training type factors
	plannedStrengthGain := trainingType.StrengthGainFactor * req.DurationHours
	plannedDefenseGain := trainingType.DefenseGainFactor * req.DurationHours
	plannedAgilityGain := trainingType.AgilityGainFactor * req.DurationHours

	// 7. Create training session
	trainingSession := &model.TrainingSession{
		BoxerID:             boxerID,
		TrainingTypeID:      req.TrainingTypeID,
		DurationHours:       req.DurationHours,
		PlannedStrengthGain: plannedStrengthGain,
		PlannedDefenseGain:  plannedDefenseGain,
		PlannedAgilityGain:  plannedAgilityGain,
		Status:              model.TrainingSessionPending,
	}

	err = h.trainingSessionStore.Create(ctx, trainingSession)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create training session"})
		return
	}

	// 8. Create scheduled event for training completion (if scheduled_at is provided in future)
	// For now, we'll schedule it to complete after the duration (instant + duration hours)
	// In a full implementation, this would use the World Clock worker
	if h.scheduledEventStore != nil && req.ScheduledAt != "" {
		// Parse scheduled time and create event
		// This is simplified - in production you'd use proper date parsing
		scheduledEvent := &model.ScheduledEvent{
			BoxerID:   boxerID,
			EventType: model.EventTypeTraining,
			// EventTime would be parsed from req.ScheduledAt
		}

		// Update training session with scheduled event ID after creation
		_ = scheduledEvent // Placeholder for future implementation
	}

	// Return success response with training session details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Training scheduled successfully",
		"session": map[string]interface{}{
			"id":                  trainingSession.ID,
			"boxer_id":            trainingSession.BoxerID,
			"training_type_id":    trainingSession.TrainingTypeID,
			"duration_hours":      trainingSession.DurationHours,
			"planned_gains": map[string]float64{
				"strength": plannedStrengthGain,
				"defense":  plannedDefenseGain,
				"agility":  plannedAgilityGain,
			},
			"energy_cost": energyCost,
			"status":      trainingSession.Status,
		},
		"training_type": map[string]interface{}{
			"name":           trainingType.Name,
			"description":    trainingType.Description,
			"energy_cost_per_hour": trainingType.EnergyCost,
		},
	})
}

// GetTrainingSessionsForBoxer returns all training sessions for a boxer
func (h *TrainingHandler) GetTrainingSessionsForBoxer(w http.ResponseWriter, r *http.Request) {
	// Parse boxer ID from URL path
	idStr := r.URL.Path[len("/boxers/"):]
	boxerID, parseErr := strconv.Atoi(idStr)
	if parseErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid boxer ID"})
		return
	}

	if h.trainingSessionStore == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode([]model.TrainingSession{})
		return
	}

	sessions, err := h.trainingSessionStore.GetPendingByBoxerID(r.Context(), boxerID)
	if err != nil {
		http.Error(w, "Failed to retrieve training sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sessions)
}
