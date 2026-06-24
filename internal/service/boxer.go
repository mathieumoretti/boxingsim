package service

import (
	"boxing/internal/db"
	"boxing/internal/model"
	"errors"
	"time"
)

var (
	ErrBoxerNotFound = errors.New("boxer not found")
	ErrInvalidStats  = errors.New("invalid stats")
	ErrHealthBelowZero = errors.New("health cannot be below zero")
)

// BoxerService handles boxer-related business logic
type BoxerService struct {
	db *db.DB
}

// NewBoxerService creates a new BoxerService
func NewBoxerService(db *db.DB) *BoxerService {
	return &BoxerService{db: db}
}

// CreateBoxer creates a new boxer for a user
func (s *BoxerService) CreateBoxer(user *model.UserCreate) (*model.Boxer, error) {
	// Create user first
	if err := db.CreateUser(s.db.DB, user); err != nil {
		return nil, err
	}

	// Create boxer for the user
	boxer := &model.BoxerCreate{
		UserID:      user.ID,
		Name:        user.BoxerName,
		PositionX:   0,
		PositionY:   0,
		Strength:    0,
		Defense:     0,
		Agility:     0,
	}

	if err := db.CreateBoxer(s.db.DB, boxer); err != nil {
		return nil, err
	}

	return boxerToResponse(boxer)
}

// GetBoxer retrieves a boxer by ID
func (s *BoxerService) GetBoxer(id int) (*model.BoxerResponse, error) {
	boxer, err := db.GetBoxerByID(s.db.DB, id)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	return boxerToResponse(boxer)
}

// GetBoxerByUserID retrieves a boxer by user ID
func (s *BoxerService) GetBoxerByUserID(userID int) (*model.BoxerResponse, error) {
	boxer, err := db.GetBoxerByUserID(s.db.DB, userID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	return boxerToResponse(boxer)
}

// UpdateBoxerStats updates boxer attributes
func (s *BoxerService) UpdateBoxerStats(id int, stats *model.BoxerUpdate) (*model.BoxerResponse, error) {
	boxer, err := db.GetBoxerByID(s.db.DB, id)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	// Apply updates
	if stats.Health != nil {
		if *stats.Health < 0 {
			return nil, ErrHealthBelowZero
		}
		boxer.Health = *stats.Health
	}
	if stats.Energy != nil {
		if *stats.Energy < 0 {
			return nil, ErrHealthBelowZero
		}
		boxer.Energy = *stats.Energy
	}
	if stats.PositionX != nil {
		boxer.PositionX = *stats.PositionX
	}
	if stats.PositionY != nil {
		boxer.PositionY = *stats.PositionY
	}

	// Save to database
	query := `
		UPDATE boxers SET
		    health = COALESCE(?, health),
		    energy = COALESCE(?, energy),
		    position_x = COALESCE(?, position_x),
		    position_y = COALESCE(?, position_y),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.DB.Exec(query, stats.Health, stats.Energy, stats.PositionX, stats.PositionY, id)
	if err != nil {
		return nil, err
	}

	return boxerToResponse(boxer)
}

// AddStats adds to boxer stats (strength, defense, agility)
func (s *BoxerService) AddStats(id int, stats *model.StatsAdd) (*model.BoxerResponse, error) {
	boxer, err := db.GetBoxerByID(s.db.DB, id)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	if stats.Strength != nil {
		boxer.Strength += *stats.Strength
	}
	if stats.Defense != nil {
		boxer.Defense += *stats.Defense
	}
	if stats.Agility != nil {
		boxer.Agility += *stats.Agility
	}

	// Calculate level based on total stats
	totalStats := float64(boxer.Strength + boxer.Defense + boxer.Agility)
	boxer.Level = int(totalStats / 10) + 1
	if boxer.Level < 1 {
		boxer.Level = 1
	}

	// Update experience
	if stats.Experience != nil {
		boxer.Experience += *stats.Experience
	}

	// Save to database
	query := `
		UPDATE boxers SET
		    strength = COALESCE(?, strength),
		    defense = COALESCE(?, defense),
		    agility = COALESCE(?, agility),
		    experience = COALESCE(?, experience),
		    level = COALESCE(?, level),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.DB.Exec(query, stats.Strength, stats.Defense, stats.Agility,
		stats.Experience, boxer.Level, id)
	if err != nil {
		return nil, err
	}

	return boxerToResponse(boxer)
}

// RecoverHealth recovers health for a boxer
func (s *BoxerService) RecoverHealth(id int, recoveryPercent float64) (*model.BoxerResponse, error) {
	boxer, err := db.GetBoxerByID(s.db.DB, id)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	recovered := int(float64(boxer.Health) * recoveryPercent)
	boxer.Health += recovered
	if boxer.Health > 100 {
		boxer.Health = 100
	}

	// Save to database
	query := `
		UPDATE boxers SET
		    health = COALESCE(?, health),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.DB.Exec(query, boxer.Health, id)
	if err != nil {
		return nil, err
	}

	return boxerToResponse(boxer)
}

// boxerToResponse converts a boxer model to response format
func boxerToResponse(boxer *model.Boxer) (*model.BoxerResponse, error) {
	totalStats := float64(boxer.Strength + boxer.Defense + boxer.Agility)
	level := int(totalStats / 10) + 1
	if level < 1 {
		level = 1
	}

	return &model.BoxerResponse{
		ID:            boxer.ID,
		UserID:        boxer.UserID,
		Name:          boxer.Name,
		Nickname:      boxer.Nickname,
		PositionX:     boxer.PositionX,
		PositionY:     boxer.PositionY,
		Health:        boxer.Health,
		Energy:        boxer.Energy,
		Strength:      boxer.Strength,
		Defense:       boxer.Defense,
		Agility:       boxer.Agility,
		Experience:    boxer.Experience,
		Level:         level,
		CreatedAt:     boxer.CreatedAt,
		UpdatedAt:     boxer.UpdatedAt,
	}, nil
}

// ValidateStats validates boxer stats
func ValidateStats(strength, defense, agility float64) error {
	if strength < 0 || defense < 0 || agility < 0 {
		return ErrInvalidStats
	}
	return nil
}