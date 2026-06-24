package handler

import (
	"boxing/internal/db"
	"boxing/internal/model"
	"boxing/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type UserHandler struct {
	userService *service.UserService
	boxerService *service.BoxerService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *service.UserService, boxerService *service.BoxerService) *UserHandler {
	return &UserHandler{
		userService: userService,
		boxerService: boxerService,
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response for authentication
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// User represents the user response
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username, password, and email are required"})
		return
	}

	user, err := h.userService.CreateUser(&model.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Login handles user login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	user, err := h.userService.Authenticate(req.Username, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid username or password"})
		return
	}

	token, err := h.userService.GenerateToken(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate token"})
		return
	}

	response := AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: User{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUser handles retrieving a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateUser handles updating a user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Email *string `json:"email"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	user, err := h.userService.UpdateUser(id, req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	err = h.userService.DeleteUser(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserBoxers handles retrieving all boxers for a user
func (h *UserHandler) GetUserBoxers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID is required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	boxers, err := h.boxerService.GetBoxersByUserID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxers)
}

// RegisterRequest represents the request body for registering a boxer
type RegisterBoxerRequest struct {
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
	Class   string `json:"class"`
}

// RegisterBoxer handles registering a new boxer for a user
func (h *UserHandler) RegisterBoxer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req RegisterBoxerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	boxer := &model.Boxer{
		UserID:  req.UserID,
		Name:    req.Name,
		Class:   req.Class,
		MaxHealth: 100,
		MaxEnergy: 100,
		Health:   100,
		Energy:   100,
		Strength: 10,
		Defense:  10,
		Agility:  10,
		SkillPoints: 0,
		Level: 1,
		XP: 0,
		Experience: 0,
		Status: "active",
	}

	err = h.boxerService.CreateBoxer(boxer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(boxer)
}

// RegisterRequest represents the request body for updating a boxer's level
type LevelUpRequest struct {
	BoxerID int `json:"boxer_id"`
	XP      float64 `json:"xp"`
}

// LevelUpBoxer handles leveling up a boxer
func (h *UserHandler) LevelUpBoxer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req LevelUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	boxer, err := h.boxerService.LevelUpBoxer(req.BoxerID, req.XP)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxer)
}