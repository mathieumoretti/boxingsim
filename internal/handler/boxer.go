package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/model"
)

// BoxerHandler handles boxer-related HTTP requests
type BoxerHandler struct {
	boxerService boxer.BoxerService
}

func NewBoxerHandler(boxerService boxer.BoxerService) *BoxerHandler {
	return &BoxerHandler{
		boxerService: boxerService,
	}
}

// CreateBoxer handles creating a new boxer
func (h *BoxerHandler) CreateBoxer(w http.ResponseWriter, r *http.Request) {
	var boxerCreate model.BoxerCreate
	if err := json.NewDecoder(r.Body).Decode(&boxerCreate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call service
	boxer, err := h.boxerService.CreateBoxer(r.Context(), r.Header.Get("user-id"), &boxerCreate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(boxer)
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

	// Call service
	boxer, err := h.boxerService.GetBoxer(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxer)
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

	// Call service
	boxer, err := h.boxerService.UpdateBoxer(r.Context(), id, &boxerUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxer)
}
