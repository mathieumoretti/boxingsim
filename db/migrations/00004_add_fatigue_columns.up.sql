-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Add fatigue tracking columns to boxers table
ALTER TABLE boxers ADD COLUMN IF NOT EXISTS fatigue_score DECIMAL(5,2) DEFAULT 0.0;
ALTER TABLE boxers ADD COLUMN IF NOT EXISTS forced_rest_until TIMESTAMP;

-- Add constraint to ensure fatigue_score stays in valid range (0-100) using DO block for idempotency
DO $do$ BEGIN ALTER TABLE boxers ADD CONSTRAINT chk_fatigue_range CHECK (fatigue_score >= 0 AND fatigue_score <= 100); EXCEPTION WHEN duplicate_object THEN NULL; END $do$;

-- Create training_history table for tracking completed training sessions
CREATE TABLE IF NOT EXISTS training_history (id SERIAL PRIMARY KEY, boxer_id INTEGER NOT NULL REFERENCES boxers(id) ON DELETE CASCADE, training_type_id INTEGER NOT NULL REFERENCES training_types(id), scheduled_event_id INTEGER REFERENCES scheduled_events(id), completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, duration_hours DOUBLE PRECISION NOT NULL CHECK (duration_hours > 0), fatigue_increase DECIMAL(5,2) NOT NULL DEFAULT 0.0, energy_cost DECIMAL(5,2) NOT NULL DEFAULT 0.0, strength_gain DOUBLE PRECISION NOT NULL DEFAULT 0, defense_gain DOUBLE PRECISION NOT NULL DEFAULT 0, agility_gain DOUBLE PRECISION NOT NULL DEFAULT 0, stats_before JSONB DEFAULT '{}'::jsonb, stats_after JSONB DEFAULT '{}'::jsonb, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);

-- Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_training_history_boxer_id ON training_history(boxer_id);
CREATE INDEX IF NOT EXISTS idx_training_history_completed_at ON training_history(completed_at);
CREATE INDEX IF NOT EXISTS idx_training_history_training_type_id ON training_history(training_type_id);
CREATE INDEX IF NOT EXISTS idx_boxers_fatigue_score ON boxers(fatigue_score);
CREATE INDEX IF NOT EXISTS idx_boxers_forced_rest_until ON boxers(forced_rest_until) WHERE forced_rest_until IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN boxers.fatigue_score IS 'Cumulative fatigue level (0.0-100.0). Increases with training, decreases with rest. >= 80 triggers exhaustion.';
COMMENT ON COLUMN boxers.forced_rest_until IS 'Timestamp until which boxer is on mandatory forced rest due to exhaustion. Cannot train until this time passes.';
COMMENT ON TABLE training_history IS 'Historical record of completed training sessions for analytics and progression tracking';
COMMENT ON COLUMN training_history.fatigue_increase IS 'Fatigue points added from this training session (duration_hours * 15)';
COMMENT ON COLUMN training_history.stats_before IS 'Boxer stats snapshot before training completion (JSONB)';
COMMENT ON COLUMN training_history.stats_after IS 'Boxer stats snapshot after training completion (JSONB)';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_boxers_forced_rest_until;
DROP INDEX IF EXISTS idx_boxers_fatigue_score;
DROP INDEX IF EXISTS idx_training_history_training_type_id;
DROP INDEX IF EXISTS idx_training_history_completed_at;
DROP INDEX IF EXISTS idx_training_history_boxer_id;
DROP TABLE IF EXISTS training_history;
ALTER TABLE boxers DROP CONSTRAINT IF EXISTS chk_fatigue_range;
ALTER TABLE boxers DROP COLUMN IF EXISTS forced_rest_until;
ALTER TABLE boxers DROP COLUMN IF EXISTS fatigue_score;
