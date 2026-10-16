-- MAT-138: Zombie Events and Fights Cleanup Script
-- =================================================
-- Purpose: Identify and fix zombie events/fights from failed worker processing
--
-- Problem Summary:
-- When a fight is scheduled, it creates:
--   1. A fight record with status='scheduled'
--   2. A scheduled_event with event_type='fight_simulate' and status='pending'
--
-- If the worker crashes or fails before processing the event:
--   - Event may stay in 'pending' or 'processing' state
--   - Event may transition to 'failed' but fight remains 'scheduled'
--   - Boxer appears unavailable for new fights (BoxerInFight check blocks booking)
--
-- This script identifies and safely remediates these zombie records.
--
-- USAGE:
--   1. Review the diagnostic queries first
--   2. Run remediation queries with care (they are safe but destructive)
--   3. Document any actions taken
--
-- PREREQUISITES:
--   - This script should be run by an administrator with database access
--   - Consider running in a transaction or on a backup first
--   - Ensure no worker is actively processing events during execution
--
-- VERSION: 1.0
-- DATE: 2026-10-15
-- AUTHOR: Mathieu Moretti

-----------------------------------------------------------------------------
SECTION 1: DIAGNOSTIC QUERIES (Run these first to understand the problem)
-----------------------------------------------------------------------------

-- 1a: Count fights in 'scheduled' status that have passed their scheduled_time
-- These are potential zombie fights waiting for events that may never be processed
SELECT
    COUNT(*) AS total_zombie_fights,
    COUNT(*) FILTER (WHERE se.id IS NOT NULL) AS has_events,
    COUNT(*) FILTER (WHERE se.id IS NULL) AS orphaned_no_event,
    COUNT(*) FILTER (WHERE se.status = 'pending') AS event_pending,
    COUNT(*) FILTER (WHERE se.status = 'processing') AS event_processing,
    COUNT(*) FILTER (WHERE se.status = 'failed') AS event_failed,
    COUNT(*) FILTER (WHERE se.status = 'completed') AS event_completed,
    MIN(f.scheduled_time) AS earliest_scheduled_time,
    MAX(f.scheduled_time) AS latest_scheduled_time
FROM fights f
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND f.scheduled_time < NOW();

-- 1b: Detailed list of zombie fights with their event status
-- Use this to understand what's happening before running cleanup
SELECT
    f.id AS fight_id,
    f.boxer1_id,
    b1.name AS boxer1_name,
    f.boxer2_id,
    b2.name AS boxer2_name,
    f.scheduled_time,
    NOW() - f.scheduled_time AS time_since_scheduled,
    se.id AS event_id,
    se.status AS event_status,
    se.event_type,
    se.error_message,
    se.processed_at,
    CASE
        WHEN se.status = 'completed' THEN 'Event completed but fight not marked - data inconsistency'
        WHEN se.status = 'failed' THEN 'Event failed - should cancel fight'
        WHEN se.status IN ('pending', 'processing') AND f.scheduled_time < NOW() THEN 'Event not processed past scheduled time - zombie'
        WHEN se.id IS NULL THEN 'No event created - orphaned fight'
        ELSE 'Unknown state'
    END AS diagnosis
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND (
      f.scheduled_time < NOW() OR  -- Past scheduled time
      se.status = 'failed' OR     -- Event failed
      se.id IS NULL               -- No event exists
  )
ORDER BY f.scheduled_time ASC;

-- 1c: Count failed events that are associated with scheduled fights
SELECT
    COUNT(*) AS total_failed_events,
    COUNT(DISTINCT se.event_data::jsonb->>'fight_id'::int) AS affected_fights
FROM scheduled_events se
WHERE se.status = 'failed'
  AND se.event_type IN ('fight_simulate', 'competition')
  AND EXISTS (
      SELECT 1 FROM fights f
      WHERE f.id = se.event_data::jsonb->>'fight_id'::int
      AND f.status = 'scheduled'
  );

-- 1d: Count events stuck in 'processing' state (potential crash victims)
SELECT
    COUNT(*) AS stuck_events,
    MIN(se.created_at) AS oldest_stuck_event,
    NOW() - MAX(se.created_at) AS max_time_stuck
FROM scheduled_events se
WHERE se.status = 'processing'
  AND se.event_type = 'fight_simulate';

-----------------------------------------------------------------------------
SECTION 2: SAFE REMEDIATION QUERIES (Review diagnostics before running)
-----------------------------------------------------------------------------

-- 2a: Cancel fights with failed events
-- These fights cannot be completed because their event processing failed
-- The fight record is cancelled, boxer becomes available again
-- ROLLBACK OPTION: Restore from backup or manually re-insert if needed

BEGIN;

-- First, log what we're about to cancel (run this as a SELECT first)
SELECT
    f.id AS fight_id,
    f.boxer1_id,
    b1.name AS boxer1_name,
    f.boxer2_id,
    b2.name AS boxer2_name,
    f.scheduled_time,
    se.error_message,
    'CANCELLED: Event processing failed' AS reason
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND se.status = 'failed'
  AND se.event_type = 'fight_simulate';

-- Now execute the cancellation (uncomment to run)
UPDATE fights f
SET
    status = 'cancelled',
    updated_at = CURRENT_TIMESTAMP,
    data = jsonb_build_object(
        'cancellation_reason', 'Event processing failed',
        'cancelled_at', CURRENT_TIMESTAMP,
        'original_scheduled_time', f.scheduled_time,
        'error_message', se.error_message
    )
FROM fights f
JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND se.status = 'failed'
  AND se.event_type = 'fight_simulate';

-- Verify the update
SELECT
    COUNT(*) AS fights_cancelled,
    status,
    COUNT(*) AS count_per_status
FROM fights
WHERE status = 'cancelled'
GROUP BY status;

COMMIT;

-- 2b: Cancel orphaned fights (scheduled past their time with no event)
-- These are fights where the event creation failed or was never called
-- The fight is cancelled, boxer becomes available again

BEGIN;

-- Log what we're about to cancel (run this as a SELECT first)
SELECT
    f.id AS fight_id,
    f.boxer1_id,
    b1.name AS boxer1_name,
    f.boxer2_id,
    b2.name AS boxer2_name,
    f.scheduled_time,
    NOW() - f.scheduled_time AS time_overdue,
    'CANCELLED: No associated event found' AS reason
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND f.scheduled_time < NOW()
  AND se.id IS NULL;

-- Execute the cancellation (uncomment to run)
UPDATE fights
SET
    status = 'cancelled',
    updated_at = CURRENT_TIMESTAMP,
    data = jsonb_build_object(
        'cancellation_reason', 'No associated event found (orphaned)',
        'cancelled_at', CURRENT_TIMESTAMP,
        'original_scheduled_time', scheduled_time
    )
WHERE id IN (
    SELECT f.id
    FROM fights f
    LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
    WHERE f.status = 'scheduled'
      AND f.scheduled_time < NOW()
      AND se.id IS NULL
);

-- Verify the update
SELECT COUNT(*) AS orphaned_fights_cancelled
FROM fights
WHERE status = 'cancelled'
  AND data::jsonb->>'cancellation_reason' LIKE '%orphaned%';

COMMIT;

-- 2c: Cancel zombie fights (past scheduled time with pending/processing events)
-- These are fights where the worker crashed or is stuck
-- The fight is cancelled, event can be safely ignored or retried manually if needed

BEGIN;

-- Log what we're about to cancel (run this as a SELECT first)
SELECT
    f.id AS fight_id,
    f.boxer1_id,
    b1.name AS boxer1_name,
    f.boxer2_id,
    b2.name AS boxer2_name,
    f.scheduled_time,
    se.status AS event_status,
    'CANCELLED: Event stuck in ' || se.status || ' state past scheduled time' AS reason
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND f.scheduled_time < NOW()
  AND se.status IN ('pending', 'processing')
  AND se.event_type = 'fight_simulate';

-- Execute the cancellation (uncomment to run)
UPDATE fights f
SET
    status = 'cancelled',
    updated_at = CURRENT_TIMESTAMP,
    data = jsonb_build_object(
        'cancellation_reason', format('Event stuck in %s state past scheduled time', se.status),
        'cancelled_at', CURRENT_TIMESTAMP,
        'original_scheduled_time', f.scheduled_time,
        'event_id', se.id::text,
        'event_status_at_cancellation', se.status
    )
FROM fights f
JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND f.scheduled_time < NOW()
  AND se.status IN ('pending', 'processing')
  AND se.event_type = 'fight_simulate';

-- Also mark the stuck events as failed to prevent re-processing
UPDATE scheduled_events
SET
    status = 'failed',
    error_message = 'Fight cancelled due to timeout - event not processed past scheduled time',
    processed_at = CURRENT_TIMESTAMP
WHERE id IN (
    SELECT se.id
    FROM fights f
    JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
    WHERE f.status = 'cancelled'  -- Just cancelled above
      AND se.status IN ('pending', 'processing')
);

-- Verify the update
SELECT
    COUNT(*) AS zombie_fights_cancelled,
    COUNT(DISTINCT se.id) AS events_marked_failed
FROM fights f
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'cancelled'
  AND f.updated_at >= CURRENT_TIMESTAMP - INTERVAL '5 minutes';

COMMIT;

-- 2d: Reset processing events stuck from crashes (MAT-113 recovery)
-- These events are in 'processing' state but the worker crashed
-- They should be reset to 'pending' for the next worker startup to pick up
-- This is the normal crash recovery path - fights are NOT cancelled here

BEGIN;

-- Log what we're about to reset (run this as a SELECT first)
SELECT
    se.id AS event_id,
    se.event_type,
    se.boxer_id,
    b.name AS boxer_name,
    se.event_time,
    se.created_at,
    NOW() - se.created_at AS time_in_processing,
    se.event_data::jsonb->>'fight_id' AS fight_id,
    f.status AS fight_status,
    'RESET: Event stuck in processing state' AS action
FROM scheduled_events se
JOIN boxers b ON se.boxer_id = b.id
LEFT JOIN fights f ON se.event_data::jsonb->>'fight_id'::int = f.id
WHERE se.status = 'processing';

-- Execute the reset (uncomment to run)
UPDATE scheduled_events
SET
    status = 'pending',
    processed_at = NULL
WHERE status = 'processing';

-- Verify the update
SELECT COUNT(*) AS events_reset_to_pending
FROM scheduled_events
WHERE status = 'pending'
  AND processed_at IS NULL;

COMMIT;

-----------------------------------------------------------------------------
SECTION 3: VERIFICATION QUERIES (Run after cleanup to confirm success)
-----------------------------------------------------------------------------

-- 3a: Verify no zombie fights remain
SELECT
    COUNT(*) AS remaining_zombie_fights
FROM fights f
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND (
      f.scheduled_time < NOW() OR
      se.status = 'failed' OR
      se.status IN ('pending', 'processing') AND f.scheduled_time < NOW()
  );

-- 3b: Verify event status distribution
SELECT
    status,
    event_type,
    COUNT(*) AS count,
    MIN(created_at) AS earliest,
    MAX(created_at) AS latest
FROM scheduled_events
GROUP BY status, event_type
ORDER BY status, event_type;

-- 3c: Verify fight status distribution
SELECT
    status,
    COUNT(*) AS count,
    MIN(scheduled_time) AS earliest_scheduled,
    MAX(scheduled_time) AS latest_scheduled
FROM fights
GROUP BY status
ORDER BY status;

-- 3d: List recently cancelled fights for audit trail
SELECT
    id AS fight_id,
    boxer1_id,
    boxer2_id,
    scheduled_time,
    data::jsonb->>'cancellation_reason' AS reason,
    data::jsonb->>'cancelled_at' AS cancelled_at,
    updated_at
FROM fights
WHERE status = 'cancelled'
  AND updated_at >= CURRENT_TIMESTAMP - INTERVAL '1 hour'
ORDER BY updated_at DESC;

-----------------------------------------------------------------------------
SECTION 4: PREVENTIVE MEASURES (Optional enhancements)
-----------------------------------------------------------------------------

-- 4a: Add index for faster zombie detection queries
-- Uncomment to create if not exists
/*
CREATE INDEX IF NOT EXISTS idx_fights_zombie_detection
ON fights(status, scheduled_time)
WHERE status = 'scheduled';

CREATE INDEX IF NOT EXISTS idx_scheduled_events_fight_lookup
ON scheduled_events ((event_data::jsonb->>'fight_id')::int)
WHERE event_type = 'fight_simulate';
*/

-- 4e: Create a view for ongoing monitoring (optional)
-- Uncomment to create if not exists
/*
CREATE OR REPLACE VIEW v_zombie_fights AS
SELECT
    f.id AS fight_id,
    f.boxer1_id,
    b1.name AS boxer1_name,
    f.boxer2_id,
    b2.name AS boxer2_name,
    f.scheduled_time,
    NOW() - f.scheduled_time AS time_overdue,
    se.id AS event_id,
    se.status AS event_status,
    se.error_message,
    CASE
        WHEN se.status = 'completed' THEN 'Event completed but fight not marked'
        WHEN se.status = 'failed' THEN 'Event failed'
        WHEN se.status IN ('pending', 'processing') AND f.scheduled_time < NOW() THEN 'Zombie - event not processed'
        WHEN se.id IS NULL THEN 'Orphaned - no event'
        ELSE 'Unknown'
    END AS zombie_type
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND (f.scheduled_time < NOW() OR se.status = 'failed' OR se.id IS NULL);

COMMENT ON VIEW v_zombie_fights IS 'View for monitoring zombie fights that need cleanup';
*/

-----------------------------------------------------------------------------
END OF SCRIPT
-----------------------------------------------------------------------------
