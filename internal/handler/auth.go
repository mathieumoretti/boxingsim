package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler() *AuthHandler {
	cfg := config.Load()
	return &AuthHandler{
		authService: auth.NewAuthService(cfg),
	}
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

	// Hash the password
	_, err := h.authService.HashPassword(registerReq.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// In a real implementation, we would:
	// 1. Check if user already exists
	// 2. Save to database using proper repository
	// For now, we'll just return a success response

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
	})
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginReq model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would:
	// 1. Find user by username
	// 2. Verify password using authService.CheckPassword
	// 3. Generate tokens
	// For now, we'll return a stub response with mock token but with proper structure

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token": "mock-jwt-token",
		"user": map[string]interface{}{
			"id":       1,
			"username": loginReq.Username,
			"email":    "user@example.com",
		},
	})
}
