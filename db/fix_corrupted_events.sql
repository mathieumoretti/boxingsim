-- Fix corrupted scheduled events with invalid JSONB data
-- This script finds and fixes events where event_data is NULL or empty

-- First, let's see what's wrong with event ID 44
SELECT
    id,
    boxer_id,
    event_type,
    event_time,
    status,
    error_message,
    event_data,
    jsonb_typeof(event_data) as data_type,
    length(event_data::text) as data_length
FROM scheduled_events
WHERE id = 44;

-- Check all events with NULL or problematic event_data
SELECT
    id,
    boxer_id,
    event_type,
    event_time,
    status,
    CASE
        WHEN event_data IS NULL THEN 'NULL'
        WHEN event_data = '{}'::jsonb THEN 'EMPTY_OBJECT'
        ELSE 'HAS_DATA'
    END as data_status
FROM scheduled_events
WHERE event_data IS NULL OR event_data = '{}'::jsonb
ORDER BY created_at DESC
LIMIT 20;

-- Fix events with NULL event_data by setting to empty JSON object
UPDATE scheduled_events
SET event_data = '{}'::jsonb
WHERE event_data IS NULL
AND status IN ('pending', 'processing');

-- Alternatively, delete orphaned events (if you want to remove them instead)
-- DELETE FROM scheduled_events WHERE event_data IS NULL AND status = 'failed';
