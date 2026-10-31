# Database Cascading Behavior

This document explains foreign key cascading behavior in the boxing simulation database. Understanding these relationships is critical for data integrity and proper cleanup operations.

## Overview

The database uses a mix of RESTRICT (default) and CASCADE behaviors depending on whether we want to preserve historical data or allow clean deletions.

## Cascading Matrix

| Parent Table | Child Table(s) | FK Column(s) | Behavior | Rationale |
|--------------|----------------|--------------|----------|-----------|
| **users** | boxers | id | RESTRICT | Prevent accidental user deletion with active boxers |
| **boxers** | fights | boxer1_id, boxer2_id | RESTRICT | Preserve fight history even if boxer deleted |
| **boxers** | fights | winner_id | SET NULL | Preserve fight but allow winner cleanup |
| **boxers** | scheduled_events | boxer_id | RESTRICT | Prevent orphan events, review before delete |
| **boxers** | training_sessions | boxer_id | RESTRICT | Training sessions may have related events |
| **boxers** | training_history | boxer_id | CASCADE | History follows boxer lifecycle |
| **training_types** | training_sessions | training_type_id | RESTRICT | Reference data rarely deleted |
| **scheduled_events** | training_sessions | scheduled_event_id | RESTRICT | Event may have other dependents |

---

## Detailed Behavior by Entity

### User Deletion

**Current Behavior:** CANNOT delete user if they have boxers.

```sql
-- Will fail if boxers exist:
DELETE FROM users WHERE id = 1;
-- ERROR: update or delete on table "users" violates foreign key constraint
```

**Rationale:** Users are the account owners. Boxers represent significant game progress (training, fights, stats). Preventing accidental deletion protects user data.

**Recommended Approach:** 
1. Delete all boxers first (requires cascading review below)
2. Then delete user
3. Or implement soft-delete with `deleted_at` column

---

### Boxer Deletion

**Current Behavior:** Complex cascading based on relationship type.

#### Fight Relationships

```sql
-- boxer1_id and boxer2_id: RESTRICT (cannot delete boxer if in fights)
-- winner_id: SET NULL (winner can be nullified)
```

**Example Scenarios:**

**Scenario 1: Delete boxer with scheduled fights**
```sql
DELETE FROM boxers WHERE id = 5;
-- ERROR: violates foreign key constraint on fights(boxer1_id)
```

**Scenario 2: Delete winner boxer after fight completed**
```sql
-- Before deletion:
-- fights.id=10, winner_id=7 (champion boxer)

DELETE FROM boxers WHERE id = 7;
-- Success! fights.winner_id is now NULL
-- Fight record preserved with NULL winner
```

**Rationale:** 
- **Preserve fight history:** Fights are immutable historical records
- **Allow cleanup:** SET NULL on winner allows removing retired/deleted champions
- **Prevent corruption:** RESTRICT on participants prevents orphaned fight data

**Recommended Workflow for Boxer Deletion:**
1. Cancel all scheduled fights involving the boxer
2. Delete boxer (completed fights remain with NULL participant if needed)
3. Run cleanup job to remove orphaned events

---

### Scheduled Events

**Current Behavior:** RESTRICT on boxer_id

```sql
-- Will fail if boxer has pending events:
DELETE FROM boxers WHERE id = 3;
-- ERROR: violates foreign key constraint on scheduled_events
```

**Rationale:** Events represent scheduled game actions (training completion, fights). Deleting boxer with active events could:
- Leave orphaned events that worker will try to process
- Cause race conditions if boxer deleted mid-processing

**Recommended Approach:**
1. Cancel/complete all pending events for boxer
2. Delete boxer
3. Or use MAT-143 cleanup job to abandon old events first

---

### Training History

**Current Behavior:** CASCADE on boxer_id ✅

```sql
-- Deletes boxer AND all training history records:
DELETE FROM boxers WHERE id = 2;
-- Success! training_history entries automatically removed
```

**Rationale:** Training history is analytical data. If boxer is deleted, their training logs lose meaning and can be safely removed.

---

### Training Types (Reference Data)

**Current Behavior:** RESTRICT on training_type_id

```sql
-- Will fail if training sessions use this type:
DELETE FROM training_types WHERE name = 'strength_training';
-- ERROR: violates foreign key constraint
```

**Rationale:** Reference data should rarely be deleted. If needed:
1. Update existing sessions to different type
2. Or add `is_active` flag instead of deletion

---

## Migration 00012 Changes

Migration `00012_fix_constraints_and_cascades.up.sql` implements:

### winner_id SET NULL Cascade

```sql
-- Changed from RESTRICT to SET NULL:
ALTER TABLE fights 
DROP CONSTRAINT fights_winner_id_fkey;

ALTER TABLE fights 
ADD CONSTRAINT fights_winner_id_fkey 
FOREIGN KEY (winner_id) REFERENCES boxers(id) 
ON DELETE SET NULL;
```

**Impact:** 
- Champion boxers can now be deleted without breaking fight records
- Fight history preserved with `winner_id = NULL`
- UI should handle NULL winner gracefully (show "Unknown" or "Deleted")

---

## Soft Delete Pattern (Recommended Future Enhancement)

Consider adding soft delete support for critical entities:

```sql
-- Add deleted_at columns:
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP;
ALTER TABLE boxers ADD COLUMN deleted_at TIMESTAMP;

-- Update queries to exclude deleted:
-- WHERE deleted_at IS NULL

-- Mark as deleted instead of removing:
UPDATE boxers SET deleted_at = NOW() WHERE id = 5;
```

**Benefits:**
- Preserves all foreign key relationships
- Enables "undelete" functionality  
- Maintains audit trail
- No cascade behavior needed (RESTRICT everywhere)

**Drawbacks:**
- All queries need `WHERE deleted_at IS NULL`
- Indexes needed on deleted_at for performance
- Data grows over time (need archival strategy)

---

## Cleanup Operations

### Safe User Deletion Script

```sql
BEGIN;

-- 1. Cancel all scheduled fights for user's boxers
UPDATE fights 
SET status = 'cancelled' 
WHERE boxer1_id IN (SELECT id FROM boxers WHERE user_id = :user_id)
   OR boxer2_id IN (SELECT id FROM boxers WHERE user_id = :user_id);

-- 2. Abandon all scheduled events for user's boxers  
UPDATE scheduled_events
SET 
  status = 'failed',
  error_message = 'Boxer deleted by user request'
WHERE boxer_id IN (SELECT id FROM boxers WHERE user_id = :user_id)
  AND status IN ('pending', 'processing');

-- 3. Delete training sessions (history cascades automatically)
DELETE FROM training_sessions 
WHERE boxer_id IN (SELECT id FROM boxers WHERE user_id = :user_id);

-- 4. Delete boxers (training_history cascades)
DELETE FROM boxers WHERE user_id = :user_id;

-- 5. Finally delete user
DELETE FROM users WHERE id = :user_id;

COMMIT;
```

---

## Related Documentation

- [Database Design](./database-design.md) - Schema overview and ERD
- [MAT-116 Audit Findings](../memory/mat-116-database-audit-findings.md) - Full audit report
- [MAT-143 Cleanup Job](../memory/mat-143-implementation.md) - Automatic event cleanup
