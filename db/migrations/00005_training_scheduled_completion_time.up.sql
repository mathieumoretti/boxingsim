-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Add scheduled_completion_time column to training_sessions with default value
-- This enables time-based training completion via world clock integration
ALTER TABLE training_sessions
ADD COLUMN IF NOT EXISTS scheduled_completion_time TIMESTAMP DEFAULT NOW();

-- Backfill existing pending sessions with a 4-hour duration from created_at
UPDATE training_sessions
SET scheduled_completion_time = created_at + INTERVAL '4 hours'
WHERE status = 'pending';

-- For completed/cancelled sessions, set to their completion or creation time
UPDATE training_sessions
SET scheduled_completion_time = COALESCE(completed_at, created_at + INTERVAL '4 hours')
WHERE status IN ('completed', 'cancelled');

-- Now safe to make NOT NULL since all rows have values
ALTER TABLE training_sessions
ALTER COLUMN scheduled_completion_time SET NOT NULL;

-- Remove the default since application will set it explicitly
ALTER TABLE training_sessions
ALTER COLUMN scheduled_completion_time DROP DEFAULT;

-- Create index for efficient querying of due training sessions
CREATE INDEX IF NOT EXISTS idx_training_sessions_scheduled_completion_time
ON training_sessions(scheduled_completion_time);

-- Create composite index for worker queries (status + completion time filter)
CREATE INDEX IF NOT EXISTS idx_training_sessions_pending_due
ON training_sessions(status, scheduled_completion_time)
WHERE status = 'pending';

-- Add comments for documentation
COMMENT ON COLUMN training_sessions.scheduled_completion_time IS 'Game timestamp when training should complete. Worker processes sessions where scheduled_completion_time <= current_game_time.';
COMMENT ON INDEX idx_training_sessions_scheduled_completion_time IS 'Index for time-based training session queries';
COMMENT ON INDEX idx_training_sessions_pending_due IS 'Partial index for efficiently finding pending due training sessions';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_training_sessions_pending_due;
DROP INDEX IF EXISTS idx_training_sessions_scheduled_completion_time;
ALTER TABLE training_sessions DROP COLUMN IF EXISTS scheduled_completion_time;
