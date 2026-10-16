-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- MAT-138: Zombie Fight Prevention and Detection
-- Add indexes to improve zombie fight detection query performance
-- This supports the cleanup script and future automatic cleanup jobs

-- Index for fast zombie fight detection (scheduled fights ordered by time)
-- Used in diagnostic queries to find fights past their scheduled_time
CREATE INDEX IF NOT EXISTS idx_fights_zombie_detection
ON fights(status, scheduled_time)
WHERE status = 'scheduled';

-- Index for lookup of events by fight_id from JSONB data
-- This speeds up joins between fights and scheduled_events during cleanup
CREATE INDEX IF NOT EXISTS idx_scheduled_events_fight_lookup
ON scheduled_events ((CAST(event_data->>'fight_id' AS int)))
WHERE event_type = 'fight_simulate'
  AND event_data->>'fight_id' IS NOT NULL;

-- Index for monitoring failed fight events specifically
CREATE INDEX IF NOT EXISTS idx_scheduled_events_failed_fights
ON scheduled_events(id, event_time)
WHERE status = 'failed'
  AND event_type = 'fight_simulate';

-- Add column comments for documentation
COMMENT ON INDEX idx_fights_zombie_detection IS 'Fast detection of zombie fights (scheduled past their time)';
COMMENT ON INDEX idx_scheduled_events_fight_lookup IS 'Lookup scheduled events by fight_id from JSONB event_data';
COMMENT ON INDEX idx_scheduled_events_failed_fights IS 'Monitoring failed fight simulation events';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_scheduled_events_failed_fights;
DROP INDEX IF EXISTS idx_scheduled_events_fight_lookup;
DROP INDEX IF EXISTS idx_fights_zombie_detection;
