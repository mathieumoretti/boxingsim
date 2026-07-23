package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.AuthService
	db          *database.PostgresDB
}

func NewAuthHandler(db *database.PostgresDB) *AuthHandler {
	cfg := config.Load()
	logger := logger.New("auth")
	logger.Info("Initializing AuthHandler")
	return &AuthHandler{
		authService: auth.NewAuthService(cfg),
		db:          db,
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

	// Check if user already exists
	if h.db != nil {
		_, err := db.GetUserByUsername(h.db.DB, registerReq.Username)
		if err == nil {
			logger.Error("User already exists: %s", registerReq.Username)
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
	}

	// Hash the password
	hashedPassword, err := h.authService.HashPassword(registerReq.Password)
	if err != nil {
		logger.Error("Failed to hash password for user %s: %v", registerReq.Username, err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Save user to database
	if h.db != nil {
		userCreate := &model.UserCreate{
			Username:       registerReq.Username,
			Email:          registerReq.Email,
			HashedPassword: hashedPassword,
		}
		err = db.CreateUser(h.db.DB, userCreate)
		if err != nil {
			logger.Error("Failed to create user %s: %v", registerReq.Username, err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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

	// Find user in database
	var modelUser *model.User
	if h.db != nil {
		foundUser, err := db.GetUserByUsername(h.db.DB, loginReq.Username)
		if err != nil {
			logger.Error("User not found: %s", loginReq.Username)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		modelUser = foundUser
	} else {
		// For development purposes, create a mock user
		modelUser = &model.User{
			ID:             1,
			Username:       loginReq.Username,
			Email:          "user@example.com",
			HashedPassword: "$2a$10$examplehashedpassword", // This is just for development
		}
	}

	// Verify password
	if !h.authService.CheckPassword(loginReq.Password, modelUser.HashedPassword) {
		logger.Error("Invalid password for user: %s", loginReq.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token pair
	tokenPair, err := h.authService.GenerateTokenPair(modelUser)
	if err != nil {
		logger.Error("Failed to generate tokens for user %s: %v", loginReq.Username, err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenPair.AccessToken,
		"user": map[string]interface{}{
			"id":       modelUser.ID,
			"username": modelUser.Username,
			"email":    modelUser.Email,
		},
	})
	logger.Info("LoginUser completed successfully for user: %s", loginReq.Username)
}
