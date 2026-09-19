-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Add career statistics columns to boxers table
-- All existing boxers will have 0 for all new fields due to DEFAULT 0
ALTER TABLE boxers ADD COLUMN wins INTEGER NOT NULL DEFAULT 0;
ALTER TABLE boxers ADD COLUMN losses INTEGER NOT NULL DEFAULT 0;
ALTER TABLE boxers ADD COLUMN draws INTEGER NOT NULL DEFAULT 0;
ALTER TABLE boxers ADD COLUMN knockouts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE boxers ADD COLUMN knockdowns_suffered INTEGER NOT NULL DEFAULT 0;

-- Create index on fight record fields for ranking queries
CREATE INDEX IF NOT EXISTS idx_boxers_fight_record ON boxers(wins, losses, draws);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX IF EXISTS idx_boxers_fight_record;
ALTER TABLE boxers DROP COLUMN IF EXISTS knockdowns_suffered;
ALTER TABLE boxers DROP COLUMN IF EXISTS knockouts;
ALTER TABLE boxers DROP COLUMN IF EXISTS draws;
ALTER TABLE boxers DROP COLUMN IF EXISTS losses;
ALTER TABLE boxers DROP COLUMN IF EXISTS wins;
