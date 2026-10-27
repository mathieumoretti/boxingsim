-- +goose Up
-- Convert fights.data from TEXT to JSONB for proper JSON handling
--
-- This is a development database migration. In production, you'd need to handle
-- existing data carefully. Here we simply drop the column and recreate it as JSONB.

-- Drop the existing TEXT column (data will be lost)
ALTER TABLE fights DROP COLUMN IF EXISTS data;

-- Recreate as JSONB with default empty object
ALTER TABLE fights ADD COLUMN data jsonb DEFAULT '{}'::jsonb;

-- +goose Down
-- Revert to TEXT type

ALTER TABLE fights DROP COLUMN IF EXISTS data;
ALTER TABLE fights ADD COLUMN data text DEFAULT '';
