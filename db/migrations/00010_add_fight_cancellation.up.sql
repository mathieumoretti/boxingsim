-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- MAT-140: Add Fight Cancellation with Proper Cleanup
-- This migration documents the cancellation feature and ensures proper indexing.
-- The fights table already supports 'cancelled' status (no CHECK constraint).

-- Note: Migration 00011 will convert data column to JSONB for proper indexing

-- Add comment documenting the cancelled status
COMMENT ON COLUMN fights.status IS
'Fight status: scheduled, in_progress, completed, cancelled.
Use "cancelled" for user-initiated fight cancellations (MAT-140).';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
