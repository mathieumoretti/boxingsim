package service

import (
	"boxing/internal/db"
	"boxing/internal/model"
	"errors"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidUser    = errors.New("invalid user")
)

// UserService handles user-related business logic
type UserService struct {
	db *db.DB
}

// NewUserService creates a new UserService
func NewUserService(db *db.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *model.UserCreate) (*model.UserResponse, error) {
	// Validate required fields
	if req.Username == "" {
		return nil, ErrInvalidUser
	}

	// Create user in database
	user, err := db.CreateUser(s.db.DB, req)
	if err != nil {
		return nil, err
	}

	return userToResponse(user)
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id int) (*model.UserResponse, error) {
	user, err := db.GetUserByID(s.db.DB, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return userToResponse(user)
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*model.UserResponse, error) {
	user, err := db.GetUserByUsername(s.db.DB, username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return userToResponse(user)
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(id int, req *model.UserUpdate) (*model.UserResponse, error) {
	// Check if user exists
	_, err := db.GetUserByID(s.db.DB, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update user in database
	user, err := db.UpdateUser(s.db.DB, id, req)
	if err != nil {
		return nil, err
	}

	return userToResponse(user)
}

// UpdateUserStats updates a user's experience and level
func (s *UserService) UpdateUserStats(id int, stats *model.UserUpdate) (*model.UserResponse, error) {
	// Check if user exists
	_, err := db.GetUserByID(s.db.DB, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update user in database
	user, err := db.UpdateUser(s.db.DB, id, stats)
	if err != nil {
		return nil, err
	}

	return userToResponse(user)
}

// ListUsers retrieves all users
func (s *UserService) ListUsers(limit, offset int) ([]*model.UserResponse, error) {
	users, err := db.ListUsers(s.db.DB, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*model.UserResponse, 0, len(users))
	for _, user := range users {
		response, err := userToResponse(user)
		if err != nil {
			continue
		}
		result = append(result, response)
	}

	return result, nil
}

// userToResponse converts a user model to response format
func userToResponse(user *model.User) (*model.UserResponse, error) {
	return &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Experience: user.Experience,
		Level:     user.Level,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}