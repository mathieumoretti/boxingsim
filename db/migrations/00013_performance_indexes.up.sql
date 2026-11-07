-- Migration 00013: Performance Indexes (MAT-132)
-- Adds composite indexes for fight history queries and opponent discovery optimization

-- Composite index for fight history queries filtering by boxer_id AND status
-- Partial index to cover the most common query patterns (scheduled/completed fights)
CREATE INDEX idx_fights_boxer1_status ON fights(boxer1_id, status) WHERE status IN ('scheduled', 'completed');
CREATE INDEX idx_fights_boxer2_status ON fights(boxer2_id, status) WHERE status IN ('scheduled', 'completed');

-- Composite index for opponent discovery queries (status + level filtering)
-- Helps queries that filter by availability and order by level difference
CREATE INDEX idx_boxers_status_level ON boxers(status, level) WHERE status = 'available';

-- Index on scheduled_events for boxer-specific pending event lookups
-- Covers GetPendingByBoxerIDAndType queries in the worker
CREATE INDEX idx_scheduled_events_boxer_pending ON scheduled_events(boxer_id, event_type, processed)
WHERE processed = FALSE AND status IN ('pending', 'processing');

-- Covering index for training session completion queries
-- Includes scheduled_completion_time and status to avoid table lookups
CREATE INDEX idx_training_sessions_completion ON training_sessions(scheduled_completion_time, status, boxer_id)
WHERE status = 'in_progress';
