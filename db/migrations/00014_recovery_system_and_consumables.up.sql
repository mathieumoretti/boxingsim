-- +goose Up
-- MAT-153 through MAT-158: Recovery System + Consumables Foundation
--
-- This migration adds:
-- 1. Boxer recovery columns (forced_recovery_until, next_available_at)
-- 2. Consumables system tables (consumables, boxer_consumables)
-- 3. Supporting indexes for queries

-------------------------------------------
-- 1. Recovery System Columns on boxers table
-------------------------------------------
ALTER TABLE boxers
ADD COLUMN forced_recovery_until TIMESTAMPTZ NULL,
ADD COLUMN next_available_at TIMESTAMPTZ NULL;

CREATE INDEX idx_boxers_forced_recovery ON boxers(forced_recovery_until)
WHERE forced_recovery_until IS NOT NULL;

CREATE INDEX idx_boxers_next_available ON boxers(next_available_at)
WHERE next_available_at IS NOT NULL;

COMMENT ON COLUMN boxers.forced_recovery_until IS
'Timestamp until which boxer is on mandatory recovery from fight injuries';

COMMENT ON COLUMN boxers.next_available_at IS
'Earliest timestamp when boxer can take next action (fight, training, etc.)';

-------------------------------------------
-- 2. Consumables Reference Table
-------------------------------------------
CREATE TABLE consumables (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    consumable_type TEXT NOT NULL, -- 'energy', 'health', 'recovery'
    energy_restore REAL NOT NULL DEFAULT 0,
    health_restore REAL NOT NULL DEFAULT 0,
    fatigue_reduction REAL NOT NULL DEFAULT 0,
    cooldown_hours INTEGER NOT NULL DEFAULT 0,
    is_unlimited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE consumables IS
'Reference data for energy/health/recovery consumable items';

COMMENT ON COLUMN consumables.consumable_type IS
'Type: energy (restores energy), health (heals HP), recovery (reduces fatigue)';

COMMENT ON COLUMN consumables.is_unlimited IS
'If true, no inventory tracking needed; always available to use';

-------------------------------------------
-- 3. Boxer Consumables Inventory Table
-------------------------------------------
CREATE TABLE boxer_consumables (
    id SERIAL PRIMARY KEY,
    boxer_id INTEGER NOT NULL,
    consumable_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (boxer_id) REFERENCES boxers(id) ON DELETE CASCADE,
    FOREIGN KEY (consumable_id) REFERENCES consumables(id) ON DELETE CASCADE,
    UNIQUE(boxer_id, consumable_id)
);

CREATE INDEX idx_boxer_consumables_boxer ON boxer_consumables(boxer_id);
CREATE INDEX idx_boxer_consumables_last_used ON boxer_consumables(last_used_at);

COMMENT ON TABLE boxer_consumables IS
'Inventory tracking for boxer-owned consumables';

-------------------------------------------
-- 4. Seed Default Consumables (reference data)
-------------------------------------------
-- Basic rest consumables
INSERT INTO consumables (name, description, consumable_type, energy_restore, health_restore, fatigue_reduction, cooldown_hours, is_unlimited) VALUES
('Quick Nap', 'Take a short nap to recover some energy', 'energy', 50.0, 0.0, 20.0, 1, FALSE),
('Full Rest', 'Get a full night of rest for maximum recovery', 'energy', 100.0, 0.0, 40.0, 8, FALSE),
('Ice Pack', 'Reduce fatigue and inflammation from recent activity', 'recovery', 0.0, 5.0, 30.0, 2, FALSE),
('Sports Drink', 'Hydrate and restore energy levels', 'energy', 30.0, 0.0, 10.0, 0, TRUE),
('Protein Shake', 'Aid muscle recovery after training', 'recovery', 20.0, 0.0, 15.0, 0, TRUE)
ON CONFLICT (name) DO NOTHING;

-- +goose Down
-- Rollback in reverse order

DELETE FROM consumables WHERE name IN ('Quick Nap', 'Full Rest', 'Ice Pack', 'Sports Drink', 'Protein Shake');
DROP TABLE IF EXISTS boxer_consumables;
DROP TABLE IF EXISTS consumables;
DROP INDEX IF EXISTS idx_boxers_next_available;
DROP INDEX IF EXISTS idx_boxers_forced_recovery;
ALTER TABLE boxers DROP COLUMN IF EXISTS next_available_at;
ALTER TABLE boxers DROP COLUMN IF EXISTS forced_recovery_until;
