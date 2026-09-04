package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// InitializeSchema creates all database tables.
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

// CreateUser creates a new user.
func CreateUser(db *sql.DB, user *model.UserCreate) error {
	query := `
		INSERT INTO users (username, email, hashed_password)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userID int
	err := db.QueryRow(query, user.Username, user.Email, user.HashedPassword).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("Created new User with ID=%d\n", userID)
	return nil
}

// CreateBoxer creates a new boxer for user ID 1 (default/owner).
func CreateBoxer(db *sql.DB, boxer *model.BoxerCreate) (*model.Boxer, error) {
	return CreateBoxerForUser(db, 1, boxer)
}

// CreateBoxerForUser creates a new boxer for a specific user by ID.
func CreateBoxerForUser(db *sql.DB, userID int, boxer *model.BoxerCreate) (*model.Boxer, error) {
	query := `
		INSERT INTO boxers (user_id, name, nickname, position_x, position_y, strength, defense, agility)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, name, nickname, position_x, position_y, health, energy,
		          strength, defense, agility, experience, level, created_at, updated_at
	`

	boxerModel := &model.Boxer{}
	var nullHealth, nullEnergy sql.NullFloat64

	err := db.QueryRow(query, userID, boxer.Name, boxer.Nickname, boxer.PositionX, boxer.PositionY,
		boxer.Strength, boxer.Defense, boxer.Agility).Scan(
		&boxerModel.ID,
		&boxerModel.UserID,
		&boxerModel.Name,
		&boxerModel.Nickname,
		&boxerModel.PositionX,
		&boxerModel.PositionY,
		&nullHealth, // health (uses default 100)
		&nullEnergy, // energy (uses default 100)
		&boxerModel.Strength,
		&boxerModel.Defense,
		&boxerModel.Agility,
		&boxerModel.Experience,
		&boxerModel.Level,
		&boxerModel.CreatedAt,
		&boxerModel.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create boxer: %w", err)
	}

	boxerModel.Health = 100 // Set default health
	boxerModel.Energy = 100 // Set default energy

	fmt.Printf("Created new Boxer with ID=%d\n", boxerModel.ID)
	return boxerModel, nil
}

// CreateScheduledEvent creates a new scheduled event.
func CreateScheduledEvent(db *sql.DB, event *model.ScheduledEventCreate) error {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, data)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(query, event.BoxerID, event.EventType, event.EventTime, event.Data)
	if err != nil {
		return fmt.Errorf("failed to create scheduled event: %w", err)
	}

	return nil
}

// CreateTrainingSession creates a new training session using the MAT-72 schema.
func CreateTrainingSession(db *sql.DB, session *model.TrainingSessionCreate) error {
	now := time.Now()
	query := `
		INSERT INTO training_sessions (boxer_id, training_type_id, scheduled_event_id,
		                               duration_hours, planned_strength_gain, planned_defense_gain,
		                               planned_agility_gain, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 0.0, 0.0, 0.0, 'pending', $5, $6)
	`

	var scheduledEventID *int
	if session.ScheduledEventID != nil {
		scheduledEventID = session.ScheduledEventID
	}

	_, err := db.Exec(query, session.BoxerID, session.TrainingTypeID, scheduledEventID,
		session.DurationHours, now, now)
	if err != nil {
		return fmt.Errorf("failed to create training session: %w", err)
	}

	return nil
}
