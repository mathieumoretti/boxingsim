package db

import (
	"database/sql"
	"fmt"

	"github.com/mormm/boxing/internal/model"
)

// InitializeSchema creates all database tables
func InitializeSchema(db *sql.DB) error {
	schema := `
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Boxers table
CREATE TABLE IF NOT EXISTS boxers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    nickname TEXT,
    position_x REAL NOT NULL,
    position_y REAL NOT NULL,
    health REAL NOT NULL DEFAULT 100,
    energy REAL NOT NULL DEFAULT 100,
    strength REAL NOT NULL DEFAULT 0,
    defense REAL NOT NULL DEFAULT 0,
    agility REAL NOT NULL DEFAULT 0,
    experience REAL NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Fights table
CREATE TABLE IF NOT EXISTS fights (
    id SERIAL PRIMARY KEY,
    boxer1_id INTEGER NOT NULL,
    boxer2_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled',
    scheduled_time TIMESTAMP,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    winner_id INTEGER,
    round INTEGER NOT NULL DEFAULT 1,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (boxer1_id) REFERENCES boxers(id),
    FOREIGN KEY (boxer2_id) REFERENCES boxers(id),
    FOREIGN KEY (winner_id) REFERENCES boxers(id)
);

-- Scheduled events table
CREATE TABLE IF NOT EXISTS scheduled_events (
    id SERIAL PRIMARY KEY,
    boxer_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    event_time TIMESTAMP NOT NULL,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (boxer_id) REFERENCES boxers(id)
);

-- Training sessions table
CREATE TABLE IF NOT EXISTS training_sessions (
    id SERIAL PRIMARY KEY,
    boxer_id INTEGER NOT NULL,
    session_type TEXT NOT NULL,
    duration_minutes INTEGER NOT NULL,
    strength_gain REAL NOT NULL DEFAULT 0,
    defense_gain REAL NOT NULL DEFAULT 0,
    agility_gain REAL NOT NULL DEFAULT 0,
    experience_gain INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (boxer_id) REFERENCES boxers(id)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_boxers_user_id ON boxers(user_id);
CREATE INDEX IF NOT EXISTS idx_fights_boxer1_id ON fights(boxer1_id);
CREATE INDEX IF NOT EXISTS idx_fights_boxer2_id ON fights(boxer2_id);
CREATE INDEX IF NOT EXISTS idx_fights_status ON fights(status);
CREATE INDEX IF NOT EXISTS idx_scheduled_events_boxer_id ON scheduled_events(boxer_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_events_event_time ON scheduled_events(event_time);
CREATE INDEX IF NOT EXISTS idx_training_sessions_boxer_id ON training_sessions(boxer_id);
`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// GetBoxerByID retrieves a boxer by ID
func GetBoxerByID(db *sql.DB, id int) (*model.Boxer, error) {
	var boxer model.Boxer
	err := db.QueryRow(`
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility,
		       experience, level, created_at, updated_at
		FROM boxers
		WHERE id = ?
	`, id).Scan(
		&boxer.ID, &boxer.UserID, &boxer.Name, &boxer.Nickname,
		&boxer.PositionX, &boxer.PositionY, &boxer.Health, &boxer.Energy,
		&boxer.Strength, &boxer.Defense, &boxer.Agility,
		&boxer.Experience, &boxer.Level, &boxer.CreatedAt, &boxer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &boxer, nil
}

// CreateUser creates a new user
func CreateUser(db *sql.DB, user *model.UserCreate) error {
	query := `
		INSERT INTO users (username, email, hashed_password)
		VALUES ($1, $2, $3)
	`
	_, err := db.Exec(query, user.Username, user.Email, user.HashedPassword)
	if err != nil {
		return err
	}

	return nil
}

// GetBoxerByUserID retrieves a boxer by user ID
func GetBoxerByUserID(db *sql.DB, userID int) (*model.Boxer, error) {
	var boxer model.Boxer
	err := db.QueryRow(`
		SELECT id, user_id, name, nickname, position_x, position_y,
		       health, energy, strength, defense, agility,
		       experience, level, created_at, updated_at
		FROM boxers
		WHERE user_id = ?
	`, userID).Scan(
		&boxer.ID, &boxer.UserID, &boxer.Name, &boxer.Nickname,
		&boxer.PositionX, &boxer.PositionY, &boxer.Health, &boxer.Energy,
		&boxer.Strength, &boxer.Defense, &boxer.Agility,
		&boxer.Experience, &boxer.Level, &boxer.CreatedAt, &boxer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &boxer, nil
}

// CreateBoxer creates a new boxer for a user
func CreateBoxer(db *sql.DB, boxer *model.BoxerCreate) error {
	query := `
		INSERT INTO boxers (user_id, name, nickname, position_x, position_y,
		                    strength, defense, agility)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	// For now, we'll use user_id = 1 as a placeholder. In real implementation,
	// this would be passed in or retrieved from the authenticated context.
	_, err := db.Exec(query, 1, boxer.Name, boxer.Nickname, boxer.PositionX, boxer.PositionY,
		boxer.Strength, boxer.Defense, boxer.Agility)
	if err != nil {
		return err
	}

	return nil
}

// CreateScheduledEvent creates a new scheduled event
func CreateScheduledEvent(db *sql.DB, event *model.ScheduledEventCreate) error {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, data)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(query, event.BoxerID, event.EventType, event.EventTime, event.Data)
	if err != nil {
		return err
	}

	return nil
}

// CreateTrainingSession creates a new training session
func CreateTrainingSession(db *sql.DB, session *model.TrainingSessionCreate) error {
	query := `
		INSERT INTO training_sessions (boxer_id, session_type, duration_minutes,
		                               strength_gain, defense_gain, agility_gain, experience_gain)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query, session.BoxerID, session.SessionType, session.DurationMinutes,
		session.StrengthGain, session.DefenseGain, session.AgilityGain, session.ExperienceGain)
	if err != nil {
		return err
	}

	return nil
}