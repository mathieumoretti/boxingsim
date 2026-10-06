package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/db"
	pkgerrors "github.com/mormm/boxing/internal/errors"
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
	cfg, err := config.Load()
	if err != nil {
		logger := logger.New("auth")
		logger.Error("Failed to load configuration: %v", err)
		panic(err) // In production, handle gracefully; for now panic on startup failure
	}
	logger := logger.New("auth")
	logger.Info("Initializing AuthHandler")
	logger.Info("Database connection provided", "dbNil", db == nil)
	if db != nil {
		logger.Info("Database connection details", "dbStructNil", db.DB == nil)
	}
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
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid JSON"))
		return
	}

	logger.Info("Registering user: %s", registerReq.Username)

	// Validate input
	if registerReq.Password != registerReq.ConfirmPassword {
		logger.Error("Passwords do not match for user: %s", registerReq.Username)
		pkgerrors.WriteError(w, pkgerrors.Validation("password", "Passwords do not match"))
		return
	}

	// Check if user already exists
	if h.db != nil && h.db.DB != nil {
		logger.Info("Checking if user exists in database")
		_, err := db.GetUserByUsername(h.db.DB, registerReq.Username)
		if err == nil {
			logger.Error("User already exists: %s", registerReq.Username)
			pkgerrors.WriteError(w, pkgerrors.Conflict("User already exists"))
			return
		}
	} else {
		logger.Info("No database connection - skipping user existence check")
	}

	// Hash the password
	hashedPassword, err := h.authService.HashPassword(registerReq.Password)
	if err != nil {
		logger.Error("Failed to hash password for user %s: %v", registerReq.Username, err)
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to hash password"))
		return
	}

	// Save user to database
	if h.db != nil && h.db.DB != nil {
		logger.Info("Creating user in database")
		userCreate := &model.UserCreate{
			Username:       registerReq.Username,
			Email:          registerReq.Email,
			HashedPassword: hashedPassword,
		}
		err = db.CreateUser(h.db.DB, userCreate)
		if err != nil {
			logger.Error("Failed to create user %s: %v", registerReq.Username, err)
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to create user"))
			return
		}
	} else {
		logger.Info("No database connection - skipping user creation")
	}

	pkgerrors.WriteSuccess(w, http.StatusOK, map[string]string{
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
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid JSON"))
		return
	}

	logger.Info("Attempting login for user: %s", loginReq.Username)

	// Find user in database
	var modelUser *model.User
	if h.db != nil && h.db.DB != nil {
		logger.Info("Looking up user in database")
		foundUser, err := db.GetUserByUsername(h.db.DB, loginReq.Username)
		if err != nil {
			logger.Error("User not found: %s", loginReq.Username)
			pkgerrors.WriteError(w, pkgerrors.Unauthorized("Invalid credentials"))
			return
		}
		modelUser = foundUser
	} else {
		logger.Info("No database connection - using mock user for development")
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
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Invalid credentials"))
		return
	}

	// Generate token pair
	tokenPair, err := h.authService.GenerateTokenPair(modelUser)
	if err != nil {
		logger.Error("Failed to generate tokens for user %s: %v", loginReq.Username, err)
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to generate authentication token"))
		return
	}

	response := map[string]any{
		"token": tokenPair.AccessToken,
		"user": map[string]any{
			"id":       modelUser.ID,
			"username": modelUser.Username,
			"email":    modelUser.Email,
		},
	}
	pkgerrors.WriteJSON(w, http.StatusOK, response)
	logger.Info("LoginUser completed successfully for user: %s", loginReq.Username)
}
