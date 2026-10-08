-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- MAT-113: Worker & Event Processing Hardening
-- Add columns for tracking event processing lifecycle and failures

-- Add processed_at timestamp to track when events were actually processed
ALTER TABLE scheduled_events
  ADD COLUMN IF NOT EXISTS processed_at TIMESTAMP WITH TIME ZONE;

-- Add status column for explicit state machine transitions
-- States: pending → processing → completed/failed
ALTER TABLE scheduled_events
  ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'pending'
  CHECK (status IN ('pending', 'processing', 'completed', 'failed'));

-- Update existing processed events to have proper status and timestamp
UPDATE scheduled_events
SET
  status = 'completed',
  processed_at = created_at
WHERE processed = TRUE;

-- Create index for quickly finding failed events (for debugging/monitoring)
CREATE INDEX IF NOT EXISTS idx_scheduled_events_failed
  ON scheduled_events(id)
  WHERE status = 'failed';

-- Create index for events currently being processed (for crash recovery)
CREATE INDEX IF NOT EXISTS idx_scheduled_events_processing
  ON scheduled_events(id)
  WHERE status = 'processing';

-- Update existing pending events to ensure they have explicit status
UPDATE scheduled_events
SET status = 'pending'
WHERE status IS NULL AND processed = FALSE;

-- Add column comments for documentation
COMMENT ON COLUMN scheduled_events.processed_at IS 'Timestamp when event was marked as processed by the worker';
COMMENT ON COLUMN scheduled_events.status IS 'Event lifecycle state: pending (awaiting processing), processing (currently being handled), completed (successfully processed), failed (processing error)';
COMMENT ON INDEX idx_scheduled_events_failed IS 'Quick lookup for failed events requiring attention';
COMMENT ON INDEX idx_scheduled_events_processing IS 'Identify events stuck in processing state (potential crash recovery)';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_scheduled_events_processing;
DROP INDEX IF EXISTS idx_scheduled_events_failed;
ALTER TABLE scheduled_events DROP COLUMN IF EXISTS status CASCADE;
ALTER TABLE scheduled_events DROP COLUMN IF EXISTS processed_at CASCADE;
