# MAT-82: Training Duration and World Clock Integration

**Date**: 2026-09-09  
**Status**: Accepted  
**Issue**: [MAT-82](https://linear.app/mathieu-moretti/issue/MAT-82/design-training-duration-and-world-clock-integration)

## Context

The current training system (`internal/service/training.go`) creates training sessions with `status='pending'` but completes them **instantly** via `CompleteAllDueTrainingSessions()`. There is no time-based filtering - the worker processes ALL pending sessions immediately without world clock integration.

### Current Architecture

1. **Frontend**: BoxerCard shows training badge when boxer has pending session (MAT-79 UI ready)
2. **Backend**: TrainingService creates sessions with `status='pending'` but no scheduled completion time
3. **Worker**: Processes ALL pending training sessions immediately (no time filtering)
4. **Gap**: `scheduled_events` table exists for time-based event tracking, but training doesn't use it

### Technical Evidence

```go
// internal/service/training.go:181-211
func (s *TrainingService) CompleteAllDueTrainingSessions(ctx context.Context) (int, int, error) {
    sessions, err := s.trainingSessionStore.GetAllPending(ctx)
    // Fetches ALL pending sessions - NO TIME-BASED FILTERING
    
    // Completes ALL sessions immediately
    for _, session := range sessions {
        if err := s.CompleteTrainingSession(ctx, session.ID); err != nil {
            failed++
        } else {
            completed++
        }
    }
}
```

Comment at lines 179-180: *"Currently completes all pending sessions immediately. Future iterations can add time-based filtering via scheduled_events integration."*

## Decision

**Selected Approach**: Option A - Training Duration Synced with World Clock

Training sessions will have a scheduled completion time based on duration and world clock. The worker will only complete sessions whose `scheduled_completion_time <= current_game_time`.

### Rationale

1. **Realistic Simulation**: Training takes measurable game time, matching athletic training reality
2. **Visual Feedback**: Players see countdown until completion via BoxerCard badge
3. **Strategic Depth**: Players can schedule training during optimal times (night, between fights)
4. **Minimal Complexity**: Single-phase approach vs. hybrid; recovery period can be added later
5. **Leverages Existing Infrastructure**: Uses world clock model and `scheduled_events` table pattern

## Implementation Details

### 1. Database Schema Changes

**Migration**: Add scheduled completion time to training sessions

```sql
ALTER TABLE training_sessions 
ADD COLUMN scheduled_completion_time TIMESTAMP NOT NULL DEFAULT NOW();
```

**Rationale**: Stores when the training should complete in game time, enabling worker to filter due sessions.

### 2. Training Session Model Enhancement

**File**: `internal/model/training.go`

```go
type TrainingSession struct {
    ID                     uuid.UUID      `json:"id" db:"id"`
    BoxerID                uuid.UUID      `json:"boxer_id" db:"boxer_id"`
    Type                   string         `json:"type" db:"type"` // strength, defense, agility
    Status                 string         `json:"status" db:"status"` // pending, in_progress, completed
    ScheduledCompletionTime time.Time     `json:"scheduled_completion_time" db:"scheduled_completion_time"`
    DurationHours          float64        `json:"duration_hours" db:"duration_hours"`
    CreatedAt              time.Time      `json:"created_at" db:"created_at"`
    CompletedAt            *time.Time     `json:"completed_at" db:"completed_at"`
}
```

### 3. Training Service Changes

**File**: `internal/service/training.go`

#### Create Training Session

Calculate scheduled completion time based on current game time + duration:

```go
func (s *TrainingService) CreateTrainingSession(ctx context.Context, boxerID uuid.UUID, trainingType string, durationHours float64) (*TrainingSession, error) {
    // Get current game time from world clock model
    gameTime := s.worldClockModel.GetCurrentGameTime()
    
    // Calculate completion time (game time + duration in hours)
    scheduledCompletion := gameTime.Add(time.Duration(durationHours * float64(time.Hour)))
    
    session := &TrainingSession{
        ID:                      uuid.New(),
        BoxerID:                 boxerID,
        Type:                    trainingType,
        Status:                  "pending",
        ScheduledCompletionTime: scheduledCompletion,
        DurationHours:           durationHours,
        CreatedAt:               time.Now(),
    }
    
    return s.trainingSessionStore.Create(ctx, session)
}
```

#### Complete Due Training Sessions

Filter by scheduled completion time:

```go
func (s *TrainingService) CompleteAllDueTrainingSessions(ctx context.Context) (int, int, error) {
    // Get current game time
    gameTime := s.worldClockModel.GetCurrentGameTime()
    
    // Fetch only sessions where scheduled_completion_time <= current_game_time
    sessions, err := s.trainingSessionStore.GetDueSessions(ctx, gameTime)
    if err != nil {
        return 0, 0, err
    }
    
    completed, failed := 0, 0
    for _, session := range sessions {
        if err := s.CompleteTrainingSession(ctx, session.ID); err != nil {
            failed++
        } else {
            completed++
        }
    }
    
    return completed, failed, nil
}
```

### 4. Training Session Store Changes

**File**: `internal/store/training_session.go`

```go
func (s *TrainingSessionStore) GetDueSessions(ctx context.Context, gameTime time.Time) ([]*model.TrainingSession, error) {
    const query = `
        SELECT id, boxer_id, type, status, scheduled_completion_time, 
               duration_hours, created_at, completed_at
        FROM training_sessions
        WHERE status = 'pending'
          AND scheduled_completion_time <= $1
    `
    
    sessions := []*model.TrainingSession{}
    err := s.db.Select(ctx, &sessions, query, gameTime)
    return sessions, err
}
```

### 5. Frontend Countdown Display

**File**: `src/components/BoxerCard.jsx`

Calculate remaining time and display countdown:

```javascript
const calculateRemainingTime = (scheduledCompletionTime, currentGameTime, secondsPerGameHour) => {
    const completionDate = new Date(scheduledCompletionTime);
    const currentDate = new Date(currentGameTime);
    const diffMs = completionDate - currentDate;
    const diffSeconds = Math.floor(diffMs / 1000);
    
    // Convert game seconds to real-time seconds
    const realSeconds = diffSeconds / (3600 / secondsPerGameHour);
    
    if (realSeconds <= 0) return "Ready to complete!";
    
    const hours = Math.floor(realSeconds / 3600);
    const minutes = Math.floor((realSeconds % 3600) / 60);
    
    if (hours > 0) {
        return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
};

// Display on training badge:
{trainingSession && (
    <span className="training-badge">
        Training... {calculateRemainingTime(
            trainingSession.scheduled_completion_time,
            worldClock.current_game_time,
            config.seconds_per_game_hour
        )}
    </span>
)}
```

### 6. API Response Enhancement

**File**: `internal/handler/boxer.go` (or training handler)

Include scheduled completion time in boxer/training session responses:

```go
type TrainingSessionResponse struct {
    ID                      string    `json:"id"`
    Type                    string    `json:"type"`
    Status                  string    `json:"status"`
    ScheduledCompletionTime time.Time `json:"scheduled_completion_time"`
    DurationHours           float64   `json:"duration_hours"`
}

// When fetching boxer details or training sessions, include this data
```

## Edge Cases Handled

### 1. Clock Paused/Stopped Behavior

**Scenario**: World clock is paused or stopped during active training.

**Handling**: Training sessions remain in "pending" status until clock resumes. No timeout or automatic completion.

**Code**: Worker checks `world_clock.status` before processing:

```go
func (w *Worker) processTrainingSessions(ctx context.Context) error {
    clockStatus := w.worldClockModel.GetStatus(ctx)
    if clockStatus.Status != "running" {
        // Skip training completion when clock not running
        return nil
    }
    // ... normal processing
}
```

### 2. Training Cancellation

**Scenario**: Player wants to cancel an ongoing training session.

**Handling**: Add `CancelTrainingSession()` method that deletes/soft-deletes the pending session.

**Code**:
```go
func (s *TrainingService) CancelTrainingSession(ctx context.Context, sessionID uuid.UUID) error {
    return s.trainingSessionStore.Delete(ctx, sessionID)
}
```

### 3. Fighting During Training

**Scenario**: Boxer scheduled for fight while in training.

**Handling**: Training completes normally; fighter can participate at fight time. No automatic conflict resolution (future enhancement).

**Rationale**: Simplifies MVP; players manage their own schedules.

### 4. Clock Speed Changes

**Scenario**: Admin changes `seconds_per_game_hour` config mid-training.

**Handling**: Scheduled completion time is absolute game timestamp, not relative duration. Config change affects future sessions only.

**Rationale**: Preserves data integrity; countdown display adjusts automatically via recalculation.

## Follow-Up Issues Created

| Issue | Title | Dependency |
|-------|-------|------------|
| MAT-80 | Manual Training Completion Control Panel (Development) | Depends on MAT-82 ✅ |
| MAT-83 | Design Energy Recuperation and Rest Period System | Depends on MAT-80 |
| TBD | Database Migration: Add scheduled_completion_time | Part of MAT-82 implementation |
| TBD | Training Service Implementation | Part of MAT-82 implementation |
| TBD | Frontend Countdown Display | Part of MAT-82 implementation |

## Related Issues

- [MAT-79](https://linear.app/mathieu-moretti/issue/MAT-79/training-status-indicator-on-boxercard-component): Training Status Indicator on BoxerCard (UI ready)
- [MAT-22](https://linear.app/mathieu-moretti/issue/MAT-22/implement-stat-progression-and-training-effectiveness-system): Stat Progression and Training Effectiveness System
- [MAT-75](https://linear.app/mathieu-moretti/issue/MAT-75/recovery-system-and-fatigue-management-mat-64): Recovery System & Fatigue Management

## Migration from Current State

**Current**: Instant training completion with no time tracking.

**Migration Strategy**:
1. Add `scheduled_completion_time` column with default value of `NOW()`
2. Backfill existing pending sessions: set `scheduled_completion_time = created_at + duration_hours`
3. Update worker to filter by scheduled completion time
4. Frontend updates to show countdown instead of instant "Training" badge

**Data Migration SQL**:
```sql
-- Add column
ALTER TABLE training_sessions 
ADD COLUMN scheduled_completion_time TIMESTAMP NOT NULL DEFAULT NOW();

-- Backfill existing pending sessions (assuming 4-hour default duration)
UPDATE training_sessions 
SET scheduled_completion_time = created_at + INTERVAL '4 hours'
WHERE status = 'pending' AND scheduled_completion_time = created_at;
```

## Verification Checklist

- [x] Design decision documented (Option A selected)
- [x] Database migration plan drafted
- [x] Worker loop modification outlined
- [x] Frontend countdown calculation spec defined
- [x] Edge cases addressed (paused clock, cancellation, fights, config changes)
- [ ] Follow-up implementation issues created

---

**References**:
- [ADR-007: Docker Deployment](007-docker-deployment.md)
- [World Clock Worker Architecture](../.claude/projects/C--Users-mormm-Git-boxingsim/memory/world-clock-architecture.md)
