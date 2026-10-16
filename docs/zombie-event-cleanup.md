# Zombie Events and Fights Cleanup Guide (MAT-138)

## Overview

This document describes the zombie events and fights problem in the boxing simulation system, its root causes, and how to safely clean up affected records using the provided SQL script.

## Problem Description

### What Are Zombie Events/Fights?

A **zombie fight** is a fight record that remains in `scheduled` status indefinitely because its associated scheduled event failed to process properly. This typically happens when:

1. **Worker crashes** during event processing (event stuck in `processing` state)
2. **Event processing fails** due to errors (event transitions to `failed` state)
3. **Event creation fails** but fight record was already created (orphaned fight with no event)

### The Problem Flow

```
Normal Flow:
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Book Fight  │ ──► │ Create Fight     │ ──► │ Worker          │
│ Request     │     │ Record (sched)   │     │ Processes Event │
└─────────────┘     └──────────────────┘     └─────────────────┘
                               │                        │
                               ▼                        ▼
                    ┌──────────────────┐     ┌─────────────────┐
                    │ Create Scheduled │     │ Update Fight    │
                    │ Event (pending)  │     │ Status (completed)│
                    └──────────────────┘     └─────────────────┘

Zombie Flow (Worker Fails):
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Book Fight  │ ──► │ Create Fight     │ ──► │ Worker          │ ❌ CRASH/ERROR
│ Request     │     │ Record (sched)   │     │ Fails           │
└─────────────┘     └──────────────────┘     └─────────────────┘
                               │                        │
                               ▼                        ▼
                    ┌──────────────────┐     ┌─────────────────┐
                    │ Create Scheduled │     │ Fight stays     │ ⚠️ ZOMBIE
                    │ Event (failed)   │     │ in 'scheduled'  │
                    └──────────────────┘     └─────────────────┘
```

### Impact

**Severity: HIGH** - Complete functional breakdown of fight booking system

- **Boxers become unavailable**: `BoxerInFight()` check blocks new bookings for boxers with zombie fights
- **Data inconsistency**: Fight status doesn't reflect actual state
- **User frustration**: Cannot book fights, no clear error message
- **Accumulation**: Failed events pile up over time

## Root Cause Analysis

### Primary Cause: No Rollback Mechanism

When a scheduled event fails to process:

1. The `scheduled_events.status` transitions to `'failed'` (tracked by MAT-113)
2. The associated `fights.status` remains `'scheduled'` forever
3. No automatic cleanup or cancellation occurs
4. Boxer validation in `BookFight()` checks for active fights and blocks new bookings

### Related Code

**Fight Booking Validation** (`internal/store/fight_store.go:481-491`):
```go
func (s *FightStore) BoxerInFight(ctx context.Context, boxerID int) (bool, error) {
    query := `
        SELECT COUNT(*) > 0
        FROM fights
        WHERE (boxer1_id = $1 OR boxer2_id = $1)
          AND status IN ('scheduled', 'in_progress')`
    // This check blocks booking if zombie fight exists
}
```

**Event Processing** (`internal/service/event_processor.go:76-98`):
```go
case model.EventTypeFightSimulate:
    processErr = p.processFightSimulate(ctx, event)
// If this fails, event is marked as 'failed' but fight remains 'scheduled'
```

## Cleanup Script

### Location

Script: `db/fix_zombie_events_fights.sql`

### Script Structure

The script is organized into 4 sections:

1. **Diagnostic Queries** - Identify affected records
2. **Safe Remediation Queries** - Fix the issues
3. **Verification Queries** - Confirm cleanup success
4. **Preventive Measures** - Optional indexes and views

### Usage Instructions

#### Step 1: Review Your Environment

```bash
# Ensure you're connected to the correct database
psql -h localhost -p 5433 -U boxing -d boxing

# Or via Docker
docker-compose exec postgres psql -U boxing -d boxing
```

**Important**: 
- Stop any running workers before cleanup
- Consider backing up your database first
- Run in a transaction or review each query carefully

#### Step 2: Run Diagnostic Queries

Copy and execute Section 1 queries to understand the current state:

```sql
-- 1a: Get counts of zombie fights
SELECT 
    COUNT(*) AS total_zombie_fights,
    COUNT(*) FILTER (WHERE se.status = 'failed') AS event_failed,
    COUNT(*) FILTER (WHERE se.status IN ('pending', 'processing')) AS event_stuck
FROM fights f
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND f.scheduled_time < NOW();
```

**Expected Output (Healthy System)**:
```
total_zombie_fights | event_failed | event_stuck
-------------------+--------------+-------------
0                  | 0           | 0
```

**Example Output (With Zombies)**:
```
total_zombie_fights | event_failed | event_stuck
-------------------+--------------+-------------
3                  | 2           | 1
```

#### Step 3: Review Detailed List

Run query `1b` to see the specific fights and their state:

```sql
SELECT 
    f.id AS fight_id,
    b1.name AS boxer1_name,
    b2.name AS boxer2_name,
    f.scheduled_time,
    se.status AS event_status,
    se.error_message,
    diagnosis
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND (f.scheduled_time < NOW() OR se.status = 'failed' OR se.id IS NULL)
ORDER BY f.scheduled_time ASC;
```

Review the output to understand:
- Which boxers are affected
- What type of zombies exist (failed, stuck, orphaned)
- Any patterns in error messages

#### Step 4: Execute Remediation Queries

**Important**: Each subsection `2a`, `2b`, etc. is a separate transaction:

1. Run the SELECT query first to see what will be affected
2. Uncomment and run the UPDATE/DELETE queries
3. Verify with the included verification query
4. Proceed to next subsection only after confirmation

**Example for cancelling fights with failed events (Section 2a)**:

```sql
BEGIN;

-- First, review what will be cancelled
SELECT 
    f.id AS fight_id,
    b1.name AS boxer1_name,
    b2.name AS boxer2_name,
    se.error_message,
    'CANCELLED: Event processing failed' AS reason
FROM fights f
JOIN boxers b1 ON f.boxer1_id = b1.id
JOIN boxers b2 ON f.boxer2_id = b2.id
JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND se.status = 'failed'
  AND se.event_type = 'fight_simulate';

-- If output looks correct, execute the cancellation (uncomment):
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

-- Verify the cancellation
SELECT COUNT(*) AS fights_cancelled FROM fights WHERE status = 'cancelled';

COMMIT;
```

#### Step 5: Run Verification Queries

Execute Section 3 queries to confirm cleanup:

```sql
-- Should return 0 if all zombies cleaned up
SELECT COUNT(*) AS remaining_zombie_fights
FROM fights f
LEFT JOIN scheduled_events se ON f.id = se.event_data::jsonb->>'fight_id'::int
WHERE f.status = 'scheduled'
  AND (f.scheduled_time < NOW() OR se.status = 'failed');
```

### Remediation Options Summary

| Issue Type | Query Section | Action | Reversible? |
|------------|---------------|--------|-------------|
| Failed events | 2a | Cancel fights, keep event records | Yes (restore from backup) |
| Orphaned fights | 2b | Cancel fights with no events | Yes (restore from backup) |
| Stuck events (zombies) | 2c | Cancel fights + mark events failed | Yes (restore from backup) |
| Processing state crashes | 2d | Reset to pending (worker will retry) | No (events stay in system) |

## Prevention Strategies

### Short-Term

1. **Regular monitoring**: Use diagnostic queries to check for zombies weekly
2. **Worker stability**: Ensure worker process is stable and monitored
3. **Event processing timeouts**: Consider adding timeout mechanisms to prevent indefinite processing states

### Long-Term (Future Enhancements)

1. **Automatic cleanup job**: Create a periodic job that:
   - Detects zombie fights older than X hours
   - Automatically cancels them with appropriate reasons
   - Logs cancellations for audit trail

2. **Transaction boundaries**: Consider wrapping fight creation + event creation in a transaction to prevent orphans

3. **Event retry mechanism**: Add exponential backoff for failed events before cancelling associated fights

4. **Health check endpoint**: Create an admin endpoint that reports zombie counts and status

## Monitoring View (Optional)

Create the monitoring view from Section 4e:

```sql
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
```

Then monitor with:
```sql
SELECT * FROM v_zombie_fights ORDER BY scheduled_time ASC;
```

## Related Issues

- **MAT-113**: Worker & Event Processing Hardening - Added status tracking and crash recovery
- **MAT-FIGHT-SCHEDULING**: Fight scheduling validation with minimum advance notice
- **Future**: Automatic cleanup job, event retry mechanisms

## Troubleshooting

### "No rows affected" during cleanup

This is normal if there are no zombie fights. Run diagnostic queries first to confirm the issue exists.

### Worker processing events during cleanup

Stop all workers before running cleanup to prevent race conditions:
```bash
# Stop worker if running
pkill -f "boxing-worker"
```

### Events reset to pending but still not processed (Section 2d)

This indicates the worker is not running or there's a deeper issue:
1. Verify worker is running
2. Check worker logs for errors
3. Ensure game time is advancing properly

### Fight booking still fails after cleanup

Possible causes:
- Not all zombie fights were cleaned (re-run diagnostics)
- Boxer has other active issues (check health, energy, fatigue)
- Database connection/transaction isolation issues

## Appendix: Event State Machine

```
┌─────────┐    ┌──────────┐    ┌─────────┐
│ pending │ ─► │ processing│ ─► │ completed│
└─────────┘    └──────────┘    └─────────┘
                  │
                  ▼
             ┌─────────┐
             │  failed │
             └─────────┘

Transitions:
- pending → processing: Worker picks up event (MarkAsProcessing)
- processing → completed: Event handled successfully (MarkAsCompleted)
- processing → failed: Error during handling (MarkAsFailed)
- processing → pending: Crash recovery (ResetToPending in Section 2d)
```

## Appendix: Fight State Machine

```
┌──────────┐    ┌─────────────┐    ┌──────────┐
│ scheduled│ ─► │ in_progress │ ─► │ completed│
└──────────┘    └─────────────┘    └──────────┘
    │                                    ▲
    │ ┌──────────┐                      │
    └│ cancelled │◄─────────────────────┘
      └──────────┘

Transitions:
- scheduled → in_progress: Fight simulation starts
- in_progress → completed: Fight ends with winner
- scheduled → cancelled: Event failed or timeout (this cleanup script)
```
