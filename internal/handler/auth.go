package handler

import (
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
	lg          *logger.Logger
}

func NewAuthHandler(db *database.PostgresDB) *AuthHandler {
	cfg, err := config.Load()
	if err != nil {
		lg := logger.New("auth")
		lg.Error("Failed to load configuration", "error", err.Error())
		panic(err) // In production, handle gracefully; for now panic on startup failure
	}
	lg := logger.New("auth")
	dbConnected := db != nil && db.DB != nil
	lg.Info("Initializing AuthHandler", "db_connected", dbConnected)
	return &AuthHandler{
		authService: auth.NewAuthService(cfg),
		db:          db,
		lg:          lg,
	}
}

// RegisterUser handles user registration (POST /auth/register)
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerReq model.UserRegister
	if err := BindJSON(r, &registerReq); err != nil {
		h.lg.Error("Invalid JSON in RegisterUser", "error", err.Error())
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid JSON"))
		return
	}

	h.lg.Info("Registering user", "username", registerReq.Username)

	// Validate input (NO PASSWORD LOGGING - security)
	if registerReq.Password != registerReq.ConfirmPassword {
		h.lg.Error("Passwords do not match", "username", registerReq.Username)
		pkgerrors.WriteError(w, pkgerrors.Validation("password", "Passwords do not match"))
		return
	}

	// Check if user already exists
	if h.db != nil && h.db.DB != nil {
		h.lg.Debug("Checking if user exists in database", "username", registerReq.Username)
		_, err := db.GetUserByUsername(h.db.DB, registerReq.Username)
		if err == nil {
			h.lg.Error("User already exists", "username", registerReq.Username)
			pkgerrors.WriteError(w, pkgerrors.Conflict("User already exists"))
			return
		}
	} else {
		h.lg.Info("No database connection - skipping user existence check")
	}

	// Hash the password (NEVER log passwords or hashed passwords)
	hashedPassword, err := h.authService.HashPassword(registerReq.Password)
	if err != nil {
		h.lg.Error("Failed to hash password", "username", registerReq.Username, "error", err.Error())
		pkgerrors.WriteError(w, pkgerrors.Internal("Failed to hash password"))
		return
	}

	// Save user to database
	if h.db != nil && h.db.DB != nil {
		h.lg.Info("Creating user in database", "username", registerReq.Username, "email", registerReq.Email)
		userCreate := &model.UserCreate{
			Username:       registerReq.Username,
			Email:          registerReq.Email,
			HashedPassword: hashedPassword,
		}
		err = db.CreateUser(h.db.DB, userCreate)
		if err != nil {
			h.lg.Error("Failed to create user", "username", registerReq.Username, "error", err.Error())
			pkgerrors.WriteError(w, pkgerrors.Internal("Failed to create user"))
			return
		}
	} else {
		h.lg.Info("No database connection - skipping user creation")
	}

	pkgerrors.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "User registered successfully",
	})
	h.lg.Info("User registered successfully", "username", registerReq.Username)
}

// LoginUser handles user login (POST /auth/login)
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginReq model.UserLogin
	if err := BindJSON(r, &loginReq); err != nil {
		h.lg.Error("Invalid JSON in LoginUser", "error", err.Error())
		pkgerrors.WriteError(w, pkgerrors.BadRequest("Invalid JSON"))
		return
	}

	h.lg.Info("Attempting login", "username", loginReq.Username)

	// Find user in database
	var modelUser *model.User
	if h.db != nil && h.db.DB != nil {
		foundUser, err := db.GetUserByUsername(h.db.DB, loginReq.Username)
		if err != nil {
			h.lg.Error("User not found", "username", loginReq.Username)
			pkgerrors.WriteError(w, pkgerrors.Unauthorized("Invalid credentials"))
			return
		}
		modelUser = foundUser
	} else {
		h.lg.Info("No database connection - using mock user for development")
		// For development purposes, create a mock user
		modelUser = &model.User{
			ID:             1,
			Username:       loginReq.Username,
			Email:          "user@example.com",
			HashedPassword: "$2a$10$examplehashedpassword", // This is just for development
		}
	}

	// Verify password (NEVER log passwords)
	if !h.authService.CheckPassword(loginReq.Password, modelUser.HashedPassword) {
		h.lg.Error("Invalid password", "username", loginReq.Username)
		pkgerrors.WriteError(w, pkgerrors.Unauthorized("Invalid credentials"))
		return
	}

	// Generate token pair (NEVER log tokens)
	tokenPair, err := h.authService.GenerateTokenPair(modelUser)
	if err != nil {
		h.lg.Error("Failed to generate tokens", "username", loginReq.Username, "user_id", modelUser.ID, "error", err.Error())
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
	h.lg.Info("Login successful", "username", loginReq.Username, "user_id", modelUser.ID)
}
