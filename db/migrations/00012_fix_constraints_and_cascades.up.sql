-- +goose Up
-- MAT-116: Database & Migration Audit - Constraint and Cascade Fixes
--
-- This migration addresses issues found during comprehensive database audit:
-- 1. Adds CHECK constraint on fights.status to enforce valid state machine
-- 2. Adds range constraints on boxer stats (strength, defense, agility)
-- 3. Fixes winner_id foreign key to use ON DELETE SET NULL
-- 4. Adds unique constraint on boxer names per user

-------------------------------------------
-- 1. CHECK constraint on fights.status
-------------------------------------------
-- Enforces valid status values: scheduled, in_progress, completed, cancelled
-- Prevents data corruption from invalid status strings

ALTER TABLE fights
ADD CONSTRAINT chk_fights_status
CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled'));

COMMENT ON CONSTRAINT chk_fights_status ON fights IS
'Enforces fight status state machine: scheduled → in_progress → completed, or scheduled → cancelled';

-------------------------------------------
-- 2. Range constraints on boxer stats
-------------------------------------------
-- Ensures stats stay within valid 0-100 range
-- Protects against application bugs or manual corruption

ALTER TABLE boxers
ADD CONSTRAINT chk_boxer_strength CHECK (strength >= 0 AND strength <= 100);

ALTER TABLE boxers
ADD CONSTRAINT chk_boxer_defense CHECK (defense >= 0 AND defense <= 100);

ALTER TABLE boxers
ADD CONSTRAINT chk_boxer_agility CHECK (agility >= 0 AND agility <= 100);

COMMENT ON CONSTRAINT chk_boxer_strength ON boxers IS 'Strength stat must be between 0 and 100';
COMMENT ON CONSTRAINT chk_boxer_defense ON boxers IS 'Defense stat must be between 0 and 100';
COMMENT ON CONSTRAINT chk_boxer_agility ON boxers IS 'Agility stat must be between 0 and 100';

-------------------------------------------
-- 3. Fix winner_id cascade behavior
-------------------------------------------
-- Allows boxer deletion without breaking fight history
-- Winner is set to NULL if champion boxer is deleted
-- Fight record is preserved for historical integrity

ALTER TABLE fights DROP CONSTRAINT IF EXISTS fights_winner_id_fkey;

ALTER TABLE fights
ADD CONSTRAINT fights_winner_id_fkey
FOREIGN KEY (winner_id) REFERENCES boxers(id)
ON DELETE SET NULL;

COMMENT ON CONSTRAINT fights_winner_id_fkey ON fights IS
'Uses SET NULL to preserve fight records when winner boxer is deleted';

-------------------------------------------
-- 4. Unique constraint on boxer names per user
-------------------------------------------
-- Prevents same user from creating multiple boxers with identical name
-- Allows different users to have boxers with same name (common in boxing)

CREATE UNIQUE INDEX uq_boxers_user_name ON boxers(user_id, name);

COMMENT ON INDEX uq_boxers_user_name IS
'Ensures each user has unique boxer names, prevents duplicate entries';

-------------------------------------------
-- 5. Additional composite index for query optimization
-------------------------------------------
-- Supports general queries filtering by status + ordering by scheduled_time
-- Complements existing partial index idx_fights_zombie_detection (scheduled only)

CREATE INDEX IF NOT EXISTS idx_fights_status_scheduled_time
ON fights(status, scheduled_time DESC);

COMMENT ON INDEX idx_fights_status_scheduled_time IS
'Composite index for fight queries filtering by status and ordering by time';

-- Additional partial index for completed fights history queries
CREATE INDEX IF NOT EXISTS idx_fights_completed_history
ON fights(scheduled_time DESC)
WHERE status = 'completed';

COMMENT ON INDEX idx_fights_completed_history IS
'Fast retrieval of completed fights in chronological order for fight history pages';

-- +goose Down
-- Rollback changes in reverse order

DROP INDEX IF EXISTS idx_fights_completed_history;
DROP INDEX IF EXISTS idx_fights_status_scheduled_time;
DROP INDEX IF EXISTS uq_boxers_user_name;

ALTER TABLE fights DROP CONSTRAINT IF EXISTS fights_winner_id_fkey;
ALTER TABLE fights
ADD CONSTRAINT fights_winner_id_fkey
FOREIGN KEY (winner_id) REFERENCES boxers(id);

ALTER TABLE boxers DROP CONSTRAINT IF EXISTS chk_boxer_agility;
ALTER TABLE boxers DROP CONSTRAINT IF EXISTS chk_boxer_defense;
ALTER TABLE boxers DROP CONSTRAINT IF EXISTS chk_boxer_strength;

ALTER TABLE fights DROP CONSTRAINT IF EXISTS chk_fights_status;
