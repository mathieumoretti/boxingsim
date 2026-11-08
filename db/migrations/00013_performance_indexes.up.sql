-- +goose Up
-- MAT-132: Performance Indexes - Backend optimization indexes
--
-- This migration adds composite indexes to optimize common query patterns:
-- 1. Fight history queries filtering by boxer_id AND status
-- 2. Opponent discovery availability checks via NOT EXISTS on fights table
-- 3. Worker event lookups by boxer with pending status
-- 4. Training session completion hot path

-------------------------------------------
-- 1. Fight history indexes
-------------------------------------------
-- Composite indexes for fight history queries filtering by boxer_id AND status
-- Partial indexes to cover the most common query patterns (scheduled/completed fights)
-- Significantly faster than separate indexes on boxer1_id and status

CREATE INDEX idx_fights_boxer1_status ON fights(boxer1_id, status) WHERE status IN ('scheduled', 'completed');
CREATE INDEX idx_fights_boxer2_status ON fights(boxer2_id, status) WHERE status IN ('scheduled', 'completed');

-------------------------------------------
-- 2. Opponent discovery availability index
-------------------------------------------
-- Index for NOT EXISTS subquery checking if boxer is in scheduled/in_progress fight
-- Used in opponent_discovery.go line 180-186 for AvailableOnly filter
-- Composite on (boxer_id, status) with partial WHERE matches the query pattern

CREATE INDEX idx_fights_availability_check ON fights(boxer1_id, status) WHERE status IN ('scheduled', 'in_progress');
CREATE INDEX idx_fights_availability_check_boxer2 ON fights(boxer2_id, status) WHERE status IN ('scheduled', 'in_progress');

-------------------------------------------
-- 3. Worker event processing indexes
-------------------------------------------
-- Index on scheduled_events for boxer-specific pending event lookups
-- Covers GetPendingByBoxerIDAndType queries in the worker
-- Improves crash recovery by efficiently finding unprocessed events

CREATE INDEX idx_scheduled_events_boxer_pending ON scheduled_events(boxer_id, event_type, processed)
WHERE processed = FALSE AND status IN ('pending', 'processing');

-------------------------------------------
-- 4. Training session completion index
-------------------------------------------
-- Covering index for training session completion queries
-- Includes scheduled_completion_time and status to avoid table lookups
-- Hot path for worker CompleteAllDueTrainingSessions

CREATE INDEX idx_training_sessions_completion ON training_sessions(scheduled_completion_time, status, boxer_id)
WHERE status = 'in_progress';

-- +goose Down
-- Rollback indexes in reverse order

DROP INDEX IF EXISTS idx_training_sessions_completion;
DROP INDEX IF EXISTS idx_scheduled_events_boxer_pending;
DROP INDEX IF EXISTS idx_fights_availability_check_boxer2;
DROP INDEX IF EXISTS idx_fights_availability_check;
DROP INDEX IF EXISTS idx_fights_boxer2_status;
DROP INDEX IF EXISTS idx_fights_boxer1_status;
