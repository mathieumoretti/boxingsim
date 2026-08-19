-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- World Clock Table - Single-row simulation time anchor
CREATE TABLE IF NOT EXISTS world_clock (
    id SERIAL PRIMARY KEY,
    real_anchor TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    game_anchor TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT '2030-01-01 08:00:00',
    speed_factor DOUBLE PRECISION NOT NULL DEFAULT 60.0,
    status TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'paused', 'stopped')),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial world clock row with id=1
INSERT INTO world_clock (id) VALUES (1);

-- Add column comments for documentation
COMMENT ON COLUMN world_clock.game_anchor IS 'Game timestamp corresponding to real_anchor - used as base for GameTime calculation';
COMMENT ON TABLE world_clock IS 'Single-row table tracking simulation time anchors and global clock state';

-- Scheduled Events Idempotency Support Columns
ALTER TABLE scheduled_events ADD COLUMN IF NOT EXISTS processed BOOLEAN DEFAULT FALSE;
ALTER TABLE scheduled_events ADD COLUMN IF NOT EXISTS event_data JSONB;
ALTER TABLE scheduled_events ADD COLUMN IF NOT EXISTS error_message TEXT;

CREATE INDEX idx_scheduled_events_pending ON ONLY scheduled_events (event_time DESC) WHERE processed = FALSE;

COMMENT ON COLUMN scheduled_events.processed IS 'True when the world clock worker has consumed and handled this scheduled event';
COMMENT ON COLUMN scheduled_events.event_data IS 'JSONB payload containing event-specific data such as boxer IDs for a fight or promotion details';
COMMENT ON COLUMN scheduled_events.error_message IS 'Error message if processing failed during consumption; cleared on retry success';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_scheduled_events_pending CASCADE;
ALTER TABLE scheduled_events DROP COLUMN IF EXISTS error_message CASCADE;
ALTER TABLE scheduled_events DROP COLUMN IF EXISTS event_data CASCADE;
ALTER TABLE scheduled_events DROP COLUMN IF EXISTS processed CASCADE;
DROP TABLE IF EXISTS world_clock;
