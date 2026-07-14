package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mormm/boxing/internal/model"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	// In a real implementation, this would be injected with proper service
	return &AuthHandler{}
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerReq model.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if registerReq.Password != registerReq.ConfirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Call service - for now we'll just return a stub response since we don't have the real implementation yet
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{})
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginReq model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call service - for now we'll just return a stub response since we don't have the real implementation yet
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{})
}
