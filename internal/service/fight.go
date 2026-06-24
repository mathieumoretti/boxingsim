package service

import (
	"boxing/internal/db"
	"boxing/internal/model"
	"errors"
	"math/rand"
	"time"
)

var (
	ErrFightNotFound   = errors.New("fight not found")
	ErrInvalidFight    = errors.New("invalid fight")
	ErrFightInProgress = errors.New("fight in progress")
	ErrNoAvailableOpponent = errors.New("no available opponent")
)

// FightService handles fight-related business logic
type FightService struct {
	db *db.DB
}

// NewFightService creates a new FightService
func NewFightService(db *db.DB) *FightService {
	return &FightService{db: db}
}

// CreateFight creates a new fight between two boxers
func (s *FightService) CreateFight(attackerID, defenderID int) (*model.Fight, error) {
	// Check if either boxer doesn't exist
	attacker, err := db.GetBoxerByID(s.db.DB, attackerID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	defender, err := db.GetBoxerByID(s.db.DB, defenderID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	// Check if fighters are the same
	if attackerID == defenderID {
		return nil, ErrInvalidFight
	}

	// Check if either fighter is in a fight
	inAttackerFight, _ := db.BoxerInFight(s.db.DB, attackerID)
	if inAttackerFight {
		return nil, ErrFightInProgress
	}

	inDefenderFight, _ := db.BoxerInFight(s.db.DB, defenderID)
	if inDefenderFight {
		return nil, ErrFightInProgress
	}

	// Calculate fight outcome based on stats
	attackerStrength := attacker.Strength * attacker.Level
	defenderStrength := defender.Strength * defender.Level

	// Add agility modifier
	attackerStrength += attacker.Agility * 0.5
	defenderStrength += defender.Agility * 0.5

	// Create fight
	fight := &model.FightCreate{
		AttackerID:   attackerID,
		DefenderID:   defenderID,
		AttackerHP:   attacker.Health,
		DefenderHP:   defender.Health,
		Status:       "pending",
		Turns:        0,
	}

	if err := db.CreateFight(s.db.DB, fight); err != nil {
		return nil, err
	}

	return fightToResponse(fight, attacker, defender)
}

// GetFight retrieves a fight by ID
func (s *FightService) GetFight(id int) (*model.FightResponse, error) {
	fight, err := db.GetFightByID(s.db.DB, id)
	if err != nil {
		return nil, ErrFightNotFound
	}

	attacker, _ := db.GetBoxerByID(s.db.DB, fight.AttackerID)
	defender, _ := db.GetBoxerByID(s.db.DB, fight.DefenderID)

	return fightToResponse(fight, attacker, defender)
}

// StartFight starts a pending fight
func (s *FightService) StartFight(id int) (*model.FightResponse, error) {
	fight, err := db.GetFightByID(s.db.DB, id)
	if err != nil {
		return nil, ErrFightNotFound
	}

	if fight.Status != "pending" {
		return nil, ErrFightInProgress
	}

	// Mark as in progress
	query := `
		UPDATE fights SET
		    status = 'in_progress',
		    started_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.DB.Exec(query, id)
	if err != nil {
		return nil, err
	}

	// Get fighters
	attacker, err := db.GetBoxerByID(s.db.DB, fight.AttackerID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	defender, err := db.GetBoxerByID(s.db.DB, fight.DefenderID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	return fightToResponse(fight, attacker, defender)
}

// FightRound performs one round of combat
func (s *FightService) FightRound(id int) (*model.FightResponse, error) {
	fight, err := db.GetFightByID(s.db.DB, id)
	if err != nil {
		return nil, ErrFightNotFound
	}

	if fight.Status != "in_progress" {
		return nil, ErrFightInProgress
	}

	// Get fighters
	attacker, err := db.GetBoxerByID(s.db.DB, fight.AttackerID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	defender, err := db.GetBoxerByID(s.db.DB, fight.DefenderID)
	if err != nil {
		return nil, ErrBoxerNotFound
	}

	// Check if either fighter has fainted
	if fight.AttackerHP <= 0 || fight.DefenderHP <= 0 {
		return s.EndFight(id)
	}

	// Calculate round damage
	attackerDamage := s.calculateDamage(attacker, defender)
	defenderDamage := s.calculateDamage(defender, attacker)

	// Apply damage
	fight.AttackerHP -= defenderDamage
	fight.DefenderHP -= attackerDamage
	fight.Turns++

	// Ensure health doesn't go negative
	if fight.AttackerHP < 0 {
		fight.AttackerHP = 0
	}
	if fight.DefenderHP < 0 {
		fight.DefenderHP = 0
	}

	// Save progress
	query := `
		UPDATE fights SET
		    attacker_hp = ?,
		    defender_hp = ?,
		    turns = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = s.db.DB.Exec(query, fight.AttackerHP, fight.DefenderHP, fight.Turns, id)
	if err != nil {
		return nil, err
	}

	// Check if fight is over
	if fight.AttackerHP <= 0 || fight.DefenderHP <= 0 {
		return s.EndFight(id)
	}

	// Check for tie after reasonable number of turns
	if fight.Turns >= 100 {
		return s.EndFight(id)
	}

	return fightToResponse(fight, attacker, defender)
}

// EndFight ends a fight and updates boxer stats
func (s *FightService) EndFight(id int) (*model.FightResponse, error) {
	fight, err := db.GetFightByID(s.db.DB, id)
	if err != nil {
		return nil, ErrFightNotFound
	}

	if fight.Status == "completed" {
		return nil, ErrFightInProgress
	}

	// Mark as completed
	query := `
		UPDATE fights SET
		    status = 'completed',
		    winner_id = ?,
		    end_time = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	var winnerID *int
	if fight.AttackerHP > fight.DefenderHP {
		winnerID = &fight.AttackerID
	} else if fight.DefenderHP > fight.AttackerHP {
		winnerID = &fight.DefenderID
	}

	_, err = s.db.DB.Exec(query, winnerID, id)
	if err != nil {
		return nil, err
	}

	// Update winner stats
	if winnerID != nil {
		experience := float64(100) * fight.Turns * 0.1
		stats := &StatsAdd{
			Experience: &experience,
		}
		_, err = s.UpdateBoxerStats(*winnerID, stats)
		if err != nil {
			return nil, err
		}
	}

	// Get fighters for response
	attacker, _ := db.GetBoxerByID(s.db.DB, fight.AttackerID)
	defender, _ := db.GetBoxerByID(s.db.DB, fight.DefenderID)

	return fightToResponse(fight, attacker, defender)
}

// GetAvailableOpponents retrieves available opponents for a boxer
func (s *FightService) GetAvailableOpponents(boxerID int) ([]*model.Boxer, error) {
	return db.GetAvailableOpponents(s.db.DB, boxerID)
}

// calculateDamage calculates damage for a round based on fighter stats
func (s *FightService) calculateDamage(attacker, defender *model.Boxer) float64 {
	// Base damage calculation
	damage := (attacker.Strength * attacker.Level) *
		(1.0 - defender.Defense*0.02)

	// Add some randomness
	damage *= rand.Float64() * 0.4 + 0.8

	// Clamp damage to reasonable range
	if damage < 5 {
		damage = 5
	}
	if damage > 50 {
		damage = 50
	}

	return damage
}

// fightToResponse converts a fight model to response format
func fightToResponse(fight *model.Fight, attacker, defender *model.Boxer) (*model.FightResponse, error) {
	var winnerID *int
	if fight.Status == "completed" {
		if fight.AttackerHP > fight.DefenderHP {
			winnerID = &fight.AttackerID
		} else if fight.DefenderHP > fight.AttackerHP {
			winnerID = &fight.DefenderID
		}
	}

	return &model.FightResponse{
		ID:           fight.ID,
		AttackerID:   fight.AttackerID,
		DefenderID:   fight.DefenderID,
		AttackerHP:   fight.AttackerHP,
		DefenderHP:   fight.DefenderHP,
		Status:       fight.Status,
		Turns:        fight.Turns,
		WinnerID:     winnerID,
		StartedAt:    fight.StartedAt,
		EndTime:      fight.EndTime,
		CreatedAt:    fight.CreatedAt,
		UpdatedAt:    fight.UpdatedAt,
		Attacker:     boxerToResponse(attacker),
		Defender:     boxerToResponse(defender),
	}, nil
}