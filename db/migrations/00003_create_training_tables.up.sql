-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Drop old training_sessions table from migration 1 (different schema)
DROP TABLE IF EXISTS training_sessions;

-- Training Types Table (training_types)
-- Defines available training categories with their base stat gains
CREATE TABLE IF NOT EXISTS training_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    strength_gain_factor DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (strength_gain_factor >= 0 AND strength_gain_factor <= 1),
    defense_gain_factor DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (defense_gain_factor >= 0 AND defense_gain_factor <= 1),
    agility_gain_factor DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (agility_gain_factor >= 0 AND agility_gain_factor <= 1),
    energy_cost INTEGER NOT NULL DEFAULT 0 CHECK (energy_cost >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Training Sessions Table (training_sessions)
-- Tracks individual boxer training with integration to scheduled events
CREATE TABLE IF NOT EXISTS training_sessions (
    id SERIAL PRIMARY KEY,
    boxer_id INTEGER NOT NULL,
    training_type_id INTEGER NOT NULL,
    scheduled_event_id INTEGER,
    duration_hours DOUBLE PRECISION NOT NULL DEFAULT 1 CHECK (duration_hours > 0),
    planned_strength_gain DOUBLE PRECISION NOT NULL DEFAULT 0,
    planned_defense_gain DOUBLE PRECISION NOT NULL DEFAULT 0,
    planned_agility_gain DOUBLE PRECISION NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'cancelled')),
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (boxer_id) REFERENCES boxers(id),
    FOREIGN KEY (training_type_id) REFERENCES training_types(id),
    FOREIGN KEY (scheduled_event_id) REFERENCES scheduled_events(id)
);

-- Seed reference data for training types
INSERT INTO training_types (name, description, strength_gain_factor, defense_gain_factor, agility_gain_factor, energy_cost) VALUES
    ('strength_training', 'Intense weightlifting and resistance training to build raw power', 1.0, 0.1, 0.05, 15),
    ('defense_drill', 'Technical drills focusing on blocking, parrying, and defensive positioning', 0.1, 1.0, 0.1, 12),
    ('agility_workout', 'Speed and footwork exercises including ladder drills and sprints', 0.05, 0.1, 1.0, 14),
    ('sparring', 'Controlled practice fights to improve timing, reactions, and ring experience', 0.6, 0.5, 0.7, 25),
    ('cardio', 'Stamina building through running, cycling, or other endurance activities', 0.2, 0.2, 0.4, 18)
ON CONFLICT (name) DO NOTHING;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_training_sessions_boxer_id ON training_sessions(boxer_id);
CREATE INDEX IF NOT EXISTS idx_training_sessions_status ON training_sessions(status);
CREATE INDEX IF NOT EXISTS idx_training_sessions_scheduled_event_id ON training_sessions(scheduled_event_id);

COMMENT ON COLUMN training_types.name IS 'Unique identifier: strength_training, defense_drill, agility_workout, sparring, cardio';
COMMENT ON COLUMN training_types.strength_gain_factor IS 'Multiplier for strength improvement (0.0-1.0)';
COMMENT ON COLUMN training_types.defense_gain_factor IS 'Multiplier for defense improvement (0.0-1.0)';
COMMENT ON COLUMN training_types.agility_gain_factor IS 'Multiplier for agility improvement (0.0-1.0)';
COMMENT ON COLUMN training_types.energy_cost IS 'Energy points consumed per hour of training';
COMMENT ON TABLE training_types IS 'Reference data defining available training categories and their effects';
COMMENT ON COLUMN training_sessions.status IS 'pending, completed, or cancelled';
COMMENT ON TABLE training_sessions IS 'Individual training records linked to boxers and scheduled events';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_training_sessions_scheduled_event_id;
DROP INDEX IF EXISTS idx_training_sessions_status;
DROP INDEX IF EXISTS idx_training_sessions_boxer_id;
DROP TABLE IF EXISTS training_sessions;
DROP TABLE IF EXISTS training_types;
