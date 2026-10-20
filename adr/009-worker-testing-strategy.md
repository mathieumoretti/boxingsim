# Worker Testing Strategy

**Date:** 2026-10-18  
**Status:** Accepted  
**Issue:** MAT-141

## Context

The worker process (`cmd/worker/main.go`) is event-driven and time-based, making it critical to test properly. It handles:

1. Event state machine transitions (pending → processing → completed/failed)
2. Crash recovery logic (events stuck in 'processing' state)
3. Idempotent event processing (re-processing shouldn't break things)
4. Failed event handling and visibility
5. Time-based event triggering (event_time <= game_time)

The worker orchestrates fight simulations, training completion, rest/recovery events, and future competition events. Proper testing ensures data consistency and system resilience.

## Decision

We implement a **two-tier testing strategy**:

### Tier 1: Unit Tests (`internal/service/event_processor_test.go`)

Test individual event handlers in isolation with mocked dependencies:

- `processTrainingComplete()` - validates stat gains applied correctly
- `processRecovery()` - delegates to FatigueService properly
- `processFightSimulate()` - extracts fight_id, calls FightService.SimulateFight
- `processCompetition()` - placeholder behavior
- `ProcessScheduledEvent()` - state machine transitions (pending → processing → completed/failed)

**Mocking Strategy:**
- Mock `ScheduledEventStore` for state transitions
- Mock `BoxerStore` for boxer fetch/update operations
- Mock `FightService` for fight simulation
- Mock `FatigueService` for recovery operations

**Coverage Target:** 80%+ on `event_processor.go`

### Tier 2: Integration Tests (`internal/integration/worker_test.go`)

Test worker with real database connection:

- Full event lifecycle from creation to completion/failure
- Worker crash simulation (process kill mid-processing, then restart)
- Multiple events processing concurrently
- Time-based triggering with world clock manipulation
- Idempotent processing verification (same event processed twice safely)

**Test Database:** Uses separate `TEST_DB_*` environment variables (isolated from production)

## Consequences

### Positive

1. **Fast Feedback:** Unit tests run in milliseconds, providing immediate feedback on code changes
2. **Isolated Debugging:** Can reproduce specific event failures without full system setup
3. **Crash Recovery Verified:** Integration tests ensure crash recovery logic works end-to-end
4. **Idempotency Proven:** Tests verify re-processing doesn't break data consistency

### Negative

1. **Mock Maintenance:** Mocks need updates when interfaces change
2. **Integration Test Fragility:** Real database tests can be flaky if test isolation isn't perfect
3. **Time Investment:** Initial test setup requires significant effort

### Neutral

1. **Documentation:** ADR serves as reference for future contributors
2. **Test Patterns:** Established patterns can be reused for other services

## Implementation Plan

### Phase 1: Infrastructure (Week 1)

1. Create mock interfaces for dependencies
2. Set up test helper functions and fixtures
3. Document integration test setup in `docs/testing-strategy.md`

### Phase 2: Unit Tests (Week 1-2)

1. Test `processTrainingComplete()` with various stat gain scenarios
2. Test `processRecovery()` with fatigue service delegation
3. Test `processFightSimulate()` with fight ID extraction and validation
4. Test `ProcessScheduledEvent()` state machine for all event types
5. Test error handling paths (boxer not found, invalid event data, unknown event type)

### Phase 3: Integration Tests (Week 2-3)

1. Full event lifecycle test (create → pending → processing → completed)
2. Crash recovery test (simulate worker crash mid-processing)
3. Idempotency test (process same event multiple times)
4. Concurrent event processing test
5. Failed event visibility test (error messages stored and retrievable)

### Phase 4: Documentation & CI (Week 3)

1. Update `docs/testing-strategy.md` with worker-specific guidance
2. Add tests to CI pipeline
3. Set up coverage reporting

## Test Structure Example

```go
// Unit test example
func TestEventProcessor_processFightSimulate_ValidEventData(t *testing.T) {
    t.Parallel()
    
    ctx := context.Background()
    
    // Arrange
    mockEventStore := &MockScheduledEventStore{}
    mockBoxerStore := &MockBoxerStore{}
    mockFightService := &MockFightService{
        SimulateFightFunc: func(fightID int) error {
            return nil // Successful simulation
        },
    }
    mockFatigueService := &MockFatigueService{}
    
    processor := NewEventProcessor(mockEventStore, mockBoxerStore, mockFightService, mockFatigueService, logger)
    
    event := &model.ScheduledEvent{
        ID:        1,
        BoxerID:   42,
        EventType: model.EventTypeFightSimulate,
        EventData: EventData([]byte(`{"fight_id": 123}`)),
        Status:    model.EventStatusPending,
    }
    
    // Act
    err := processor.ProcessScheduledEvent(ctx, event)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, mockEventStore.MarkAsProcessingCallCount)
    assert.Equal(t, 123, mockFightService.SimulateFightArgs)
    assert.Equal(t, 1, mockEventStore.MarkAsCompletedCallCount)
}

// Integration test example
func TestWorkerEventLifecycleIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Setup real database connection from TEST_DB_* env vars
    db := setupTestDatabase(t)
    
    // Create event → process → verify completion
    // Test crash recovery scenario
    // Verify idempotency
}
```

## Related Decisions

- [[adr/002-database-design.md]] - Database schema and transactions
- [[adr/006-architecture-patterns.md]] - Service layer patterns
- [[adr/007-docker-deployment.md]] - Docker-based test infrastructure
- [[memory/world-clock-architecture.md]] - Worker architecture decision
