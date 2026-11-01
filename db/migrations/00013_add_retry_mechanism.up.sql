-- +goose Up
-- MAT-142: Add Retry Mechanism for Failed Scheduled Events
--
-- This migration adds retry tracking infrastructure to enable automatic retry
-- of failed scheduled events with exponential backoff.

-------------------------------------------
-- 1. Add retry_count column
-------------------------------------------
-- Tracks the number of retry attempts for each failed event
-- Starts at 0 (initial attempt), increments on each retry

ALTER TABLE scheduled_events
ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0;

-- Ensure existing rows have a default value
UPDATE scheduled_events SET retry_count = 0 WHERE retry_count IS NULL;

-- Add CHECK constraint to prevent negative values
ALTER TABLE scheduled_events
ADD CONSTRAINT chk_scheduled_events_retry_count CHECK (retry_count >= 0);

COMMENT ON COLUMN scheduled_events.retry_count IS 'Number of retry attempts for this event (0 = initial attempt)';
COMMENT ON CONSTRAINT chk_scheduled_events_retry_count ON scheduled_events IS 'Retry count must be non-negative';

-------------------------------------------
-- 2. Add index for efficient retry queries
-------------------------------------------
-- Optimizes querying failed events eligible for retry
-- Composite index on (status, processed_at) with partial index for 'failed' status

CREATE INDEX IF NOT EXISTS idx_scheduled_events_retry_lookup
  ON scheduled_events(status, processed_at)
  WHERE status = 'failed';

COMMENT ON INDEX idx_scheduled_events_retry_lookup IS 'Efficient lookup of failed events for retry mechanism';

-------------------------------------------
-- 3. Update existing failed events to have explicit retry_count
-------------------------------------------
-- This ensures backward compatibility with any pre-existing failed events

UPDATE scheduled_events
SET retry_count = 0
WHERE status = 'failed' AND retry_count IS NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_scheduled_events_retry_lookup;
ALTER TABLE scheduled_events DROP CONSTRAINT IF EXISTS chk_scheduled_events_retry_count;
ALTER TABLE scheduled_events DROP COLUMN IF EXISTS retry_count;
