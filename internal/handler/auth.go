package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler() *AuthHandler {
	cfg := config.Load()
	logger := logger.New("auth")
	logger.Info("Initializing AuthHandler")
	return &AuthHandler{
		authService: auth.NewAuthService(cfg),
	}
}

// RegisterUser handles user registration
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logger := logger.New("auth")
	logger.Info("RegisterUser endpoint called - Method: %s, URL: %s", r.Method, r.URL.Path)

	var registerReq model.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		logger.Error("Invalid JSON in RegisterUser: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logger.Info("Registering user: %s", registerReq.Username)

	// Validate input
	if registerReq.Password != registerReq.ConfirmPassword {
		logger.Error("Passwords do not match for user: %s", registerReq.Username)
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Hash the password
	_, err := h.authService.HashPassword(registerReq.Password)
	if err != nil {
		logger.Error("Failed to hash password for user %s: %v", registerReq.Username, err)
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
	logger.Info("RegisterUser completed successfully for user: %s", registerReq.Username)
}

// LoginUser handles user login
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	logger := logger.New("auth")
	logger.Info("LoginUser endpoint called - Method: %s, URL: %s", r.Method, r.URL.Path)

	var loginReq model.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		logger.Error("Invalid JSON in LoginUser: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logger.Info("Attempting login for user: %s", loginReq.Username)

	// In a real implementation, we would:
	// 1. Find user by username/email in database
	// 2. Verify password using authService.CheckPassword
	// 3. Generate tokens
	// For now, we'll return a stub response with proper structure for frontend testing

	// Note: This is a simplified version for development purposes
	// In production, this would:
	// - Query database for user
	// - Verify password
	// - Generate proper JWT token

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token": "mock-jwt-token-for-dev-testing",
		"user": map[string]interface{}{
			"id":       1,
			"username": loginReq.Username,
			"email":    "user@example.com",
		},
	})
	logger.Info("LoginUser completed successfully for user: %s", loginReq.Username)
}
