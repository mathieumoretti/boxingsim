-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- MAT-114: Fight Simulation Audit - Versioning and Determinism
-- Add simulation_version column to track fight simulation algorithm version
-- This enables replay of historical fights with their original algorithm

ALTER TABLE fights ADD COLUMN simulation_version VARCHAR(20) DEFAULT '1.0.0';

COMMENT ON COLUMN fights.simulation_version IS 'Version of fight simulation algorithm (semver) for deterministic replay compatibility';

-- Backfill existing fights with default version
UPDATE fights SET simulation_version = '1.0.0' WHERE simulation_version IS NULL;

-- Make column NOT NULL after backfill
ALTER TABLE fights ALTER COLUMN simulation_version SET NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE fights DROP COLUMN IF EXISTS simulation_version;
