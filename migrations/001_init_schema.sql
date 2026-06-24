-- Boxing Game Database Schema

-- Users table for authentication
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Boxers table
CREATE TABLE IF NOT EXISTS boxers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    nickname VARCHAR(100),
    position_x DECIMAL(4,2) NOT NULL DEFAULT 0.0,
    position_y DECIMAL(4,2) NOT NULL DEFAULT 0.0,
    health DECIMAL(3,1) NOT NULL CHECK (health >= 0 AND health <= 100),
    energy DECIMAL(3,1) NOT NULL CHECK (energy >= 0 AND energy <= 100),
    strength DECIMAL(4,2) NOT NULL CHECK (strength >= 0),
    defense DECIMAL(4,2) NOT NULL CHECK (defense >= 0),
    agility DECIMAL(4,2) NOT NULL CHECK (agility >= 0),
    experience DECIMAL(10,2) NOT NULL DEFAULT 0,
    level INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

-- Scheduled events table
CREATE TABLE IF NOT EXISTS scheduled_events (
    id SERIAL PRIMARY KEY,
    boxer_id INTEGER REFERENCES boxers(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_time TIMESTAMP NOT NULL,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Fights table
CREATE TABLE IF NOT EXISTS fights (
    id SERIAL PRIMARY KEY,
    boxer1_id INTEGER REFERENCES boxers(id) ON DELETE SET NULL,
    boxer2_id INTEGER REFERENCES boxers(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    scheduled_time TIMESTAMP,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    winner_id INTEGER REFERENCES boxers(id) ON DELETE SET NULL,
    round INTEGER NOT NULL DEFAULT 1,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Training sessions table
CREATE TABLE IF NOT EXISTS training_sessions (
    id SERIAL PRIMARY KEY,
    boxer_id INTEGER REFERENCES boxers(id) ON DELETE CASCADE,
    session_type VARCHAR(50) NOT NULL,
    duration_minutes INTEGER NOT NULL,
    strength_gain DECIMAL(4,2) DEFAULT 0,
    defense_gain DECIMAL(4,2) DEFAULT 0,
    agility_gain DECIMAL(4,2) DEFAULT 0,
    experience_gain INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_boxers_user_id ON boxers(user_id);
CREATE INDEX IF NOT EXISTS idx_boxers_name ON boxers(name);
CREATE INDEX IF NOT EXISTS idx_scheduled_events_boxer_id ON scheduled_events(boxer_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_events_event_time ON scheduled_events(event_time);
CREATE INDEX IF NOT EXISTS idx_scheduled_events_event_type ON scheduled_events(event_type);
CREATE INDEX IF NOT EXISTS idx_fights_boxer1_id ON fights(boxer1_id);
CREATE INDEX IF NOT EXISTS idx_fights_boxer2_id ON fights(boxer2_id);
CREATE INDEX IF NOT EXISTS idx_fights_status ON fights(status);
CREATE INDEX IF NOT EXISTS idx_fights_scheduled_time ON fights(scheduled_time);
CREATE INDEX IF NOT EXISTS idx_training_sessions_boxer_id ON training_sessions(boxer_id);
CREATE INDEX IF NOT EXISTS idx_training_sessions_created_at ON training_sessions(created_at);