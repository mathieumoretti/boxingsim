-- +goose Up
-- SQL in this section is executed when the migration is applied.

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

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE IF EXISTS training_sessions;
DROP TABLE IF EXISTS scheduled_events;
DROP TABLE IF EXISTS fights;
DROP TABLE IF EXISTS boxers;
DROP TABLE IF EXISTS users;