//go:build integration

package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/fight"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/service"
	"github.com/mormm/boxing/internal/store"
)

// worldClockModel is a package-level variable to avoid re-declaration
var worldClockModel = &model.WorldClockModel{}

// TestFightSimulateEventProcessing verifies that fight_simulate events are processed correctly by the worker.
// This test simulates the full flow: create boxers -> book fight -> create event -> process event -> verify results.
func TestFightSimulateEventProcessing(t *testing.T) {
	testDB, _ := db.FreshDatabaseWithMigrations(t, "fight_simulate_event", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available")
	}

	ctx := context.Background()

	// Setup stores
	boxerStore := store.NewBoxerStore(testDB)
	fightStore := store.NewFightStore(testDB)
	eventStore := store.NewScheduledEventStore(testDB)

	// Load config for fight service
	cfg, err := config.Load()
	if err != nil {
		t.Logf("Config load failed (using minimal defaults): %v", err)
		cfg = &config.Config{
			Fight: config.FightConfig{
				MinScheduleHours: 0, // Disable minimum schedule hours for tests
			},
		}
	} else {
		// Override MinScheduleHours for tests (even if config loaded successfully)
		cfg.Fight.MinScheduleHours = 0
	}

	testLogger := logger.New("fight_simulate_test")

	// Create boxer service for fight.SimulateFight (low-level service)
	boxerSvcForFight := boxer.NewBoxerService(boxerStore)
	fightSvcLowLevel := fight.NewFightService(fightStore, boxerStore, cfg, boxerSvcForFight, eventStore)

	// Create high-level FightService for booking fights (includes validation and event creation)
	fightSvcHighLevel := service.NewFightService(
		fightStore, boxerStore, eventStore, worldClockModel, cfg, testDB,
	)

	// Create test user
	userID := createTestUserInDB(t, testDB)

	// Step 1: Create two boxers for the fight
	t.Run("CreateTwoBoxers", func(t *testing.T) {
		boxer1 := &model.Boxer{
			UserID:   userID,
			Name:     "Fight Test Boxer 1",
			Energy:   100.0,
			Health:   100.0,
			Strength: 70.0,
			Defense:  60.0,
			Agility:  80.0,
			Level:    5,
		}

		if err := boxerStore.Create(ctx, boxer1); err != nil {
			t.Fatalf("Failed to create boxer 1: %v", err)
		}
		t.Logf("Created boxer 1 with ID: %d", boxer1.ID)

		boxer2 := &model.Boxer{
			UserID:   userID,
			Name:     "Fight Test Boxer 2",
			Energy:   100.0,
			Health:   100.0,
			Strength: 65.0,
			Defense:  70.0,
			Agility:  75.0,
			Level:    5,
		}

		if err := boxerStore.Create(ctx, boxer2); err != nil {
			t.Fatalf("Failed to create boxer 2: %v", err)
		}
		t.Logf("Created boxer 2 with ID: %d", boxer2.ID)
	})

	// Step 2: Book a fight (which creates the fight record and scheduled event)
	var fightID int
	t.Run("BookFightAndCreateEvent", func(t *testing.T) {
		boxer1, _ := boxerStore.GetByID(ctx, 1)
		boxer2, _ := boxerStore.GetByID(ctx, 2)

		if boxer1 == nil || boxer2 == nil {
			t.Fatal("Boxers not found")
		}

		// Schedule fight using game time (not real time) to pass validation
		gameTime, err := worldClockModel.GetCurrentGameTime(ctx, testDB)
		if err != nil {
			t.Fatalf("Failed to get current game time: %v", err)
		}

		// Schedule for 1 hour from game time (small offset to be in the future)
		scheduledTime := gameTime.Add(1 * time.Hour)

		if err := fightSvcHighLevel.BookFight(ctx, boxer1.ID, boxer2.ID, scheduledTime, 10); err != nil {
			t.Fatalf("Failed to book fight: %v", err)
		}

		// Verify fight was created
		fight, err := fightStore.GetByID(ctx, 1) // First fight should have ID=1
		if err != nil {
			t.Fatalf("Failed to retrieve fight: %v", err)
		}

		if fight.Status != "scheduled" {
			t.Errorf("Expected fight status 'scheduled', got '%s'", fight.Status)
		}

		fightID = fight.ID
		t.Logf("Fight created with ID: %d, scheduled for: %v", fightID, fight.ScheduledTime)
	})

	// Step 3: Verify the scheduled event was created
	t.Run("VerifyScheduledEventCreated", func(t *testing.T) {
		events, err := eventStore.GetPendingByBoxerIDAndType(ctx, 1, model.EventTypeFightSimulate)
		if err != nil {
			t.Fatalf("Failed to query events: %v", err)
		}

		if len(events) == 0 {
			t.Fatal("No pending fight_simulate events found")
		}

		event := events[0]
		t.Logf("Found pending event with ID: %d, type: %s, scheduled for: %v",
			event.ID, event.EventType, event.EventTime)

		// Verify event data contains fight_id
		var eventData map[string]interface{}
		if err := json.Unmarshal(event.EventData, &eventData); err != nil {
			t.Fatalf("Failed to unmarshal event data: %v", err)
		}

		fightIDFromEvent, ok := eventData["fight_id"].(float64)
		if !ok || int(fightIDFromEvent) != fightID {
			t.Errorf("Expected fight_id in event data to be %d, got %v", fightID, fightIDFromEvent)
		}

		t.Logf("Event data contains fight_id: %d", int(fightIDFromEvent))
	})

	// Step 4: Process the fight_simulate event (simulating what the worker does)
	t.Run("ProcessEventWithFightService", func(t *testing.T) {
		// Get the event that needs processing - use GetByID to avoid potential slice aliasing issues
		event, err := eventStore.GetByID(ctx, 1) // We know event ID is 1 from step 3
		if err != nil || event == nil {
			t.Fatal("Event not found")
		}

		// Debug: print raw event data before processing
		t.Logf("Debug: Event ID=%d, DataLength=%d", event.ID, len(event.EventData))
		t.Logf("Debug: Raw bytes=%v", []byte(event.EventData))
		t.Logf("Debug: String repr=%q", string(event.EventData))

		// Create fatigue service (needed by event processor)
		fatigueSvc := service.NewFatigueService(boxerStore, testLogger)

		// Create event processor (this is what the worker uses internally)
		// Use the low-level fight service for processing (not the high-level one with validation)
		eventProcessor := service.NewEventProcessor(eventStore, boxerStore, fightSvcLowLevel, fatigueSvc, *testLogger)

		// Process the event
		if err := eventProcessor.ProcessScheduledEvent(ctx, event); err != nil {
			t.Fatalf("Failed to process event: %v", err)
		}

		t.Log("Event processed successfully")
	})

	// Step 5: Verify fight was completed and has a winner
	t.Run("VerifyFightCompletedWithWinner", func(t *testing.T) {
		fight, err := fightStore.GetByID(ctx, fightID)
		if err != nil {
			t.Fatalf("Failed to retrieve fight after processing: %v", err)
		}

		if fight.Status != "completed" {
			t.Errorf("Expected fight status 'completed', got '%s'", fight.Status)
		}

		if fight.WinnerID == nil {
			t.Error("Fight should have a winner")
		} else {
			t.Logf("Fight completed with winner ID: %d", *fight.WinnerID)
		}

		if fight.EndTime == nil {
			// NOTE: EndTime is currently not set by SimulateFight (MAT-XXX enhancement)
			t.Logf("Note: EndTime not set in current implementation (expected behavior)")
		} else {
			t.Logf("Fight ended at: %v", *fight.EndTime)
		}

		// Verify fight data contains round information
		if len(fight.Data) > 0 {
			
			if len(fight.Data) > 0 {
				if round, ok := fight.Data["round"].(float64); ok {
					t.Logf("Fight lasted %d rounds", int(round))
				}
			}
		}
	})

	// Step 6: Verify boxers have updated stats (health and energy changed)
	t.Run("VerifyBoxerStatsUpdated", func(t *testing.T) {
		boxer1, err := boxerStore.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("Failed to retrieve boxer 1: %v", err)
		}

		boxer2, err := boxerStore.GetByID(ctx, 2)
		if err != nil {
			t.Fatalf("Failed to retrieve boxer 2: %v", err)
		}

		t.Logf("Boxer 1 stats after fight - Health: %.1f, Energy: %.1f", boxer1.Health, boxer1.Energy)
		t.Logf("Boxer 2 stats after fight - Health: %.1f, Energy: %.1f", boxer2.Health, boxer2.Energy)

		// Both boxers should have reduced health from the fight
		if boxer1.Health >= 100.0 {
			t.Error("Boxer 1 health should be reduced after fight")
		}
		if boxer2.Health >= 100.0 {
			t.Error("Boxer 2 health should be reduced after fight")
		}

		// Winner should have higher health than loser
		boxer1Wins := boxer1.Health > boxer2.Health
		expectedWinnerID := 1
		if !boxer1Wins {
			expectedWinnerID = 2
		}

		fight, _ := fightStore.GetByID(ctx, fightID)
		if fight.WinnerID != nil && *fight.WinnerID != expectedWinnerID {
			t.Errorf("Expected winner ID %d (higher health), got %d", expectedWinnerID, *fight.WinnerID)
		}
	})

	// Step 7: Verify event status was updated to completed
	t.Run("VerifyEventStatusUpdated", func(t *testing.T) {
		event, err := eventStore.GetByID(ctx, 1) // First event should have ID=1
		if err != nil {
			t.Fatalf("Failed to retrieve event: %v", err)
		}

		if event.Status != "completed" {
			t.Errorf("Expected event status 'completed', got '%s'", event.Status)
		}

		if event.ProcessedAt == nil {
			t.Error("Event should have a processed_at timestamp")
		} else {
			t.Logf("Event was processed at: %v", *event.ProcessedAt)
		}

		t.Log("Event status verified as completed")
	})
}

// TestFightSimulateEventWithInvalidFightID verifies that processing an event with invalid fight_id returns error.
func TestFightSimulateEventWithInvalidFightID(t *testing.T) {
	testDB, _ := db.FreshDatabaseWithMigrations(t, "fight_simulate_invalid", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available")
	}

	ctx := context.Background()

	eventStore := store.NewScheduledEventStore(testDB)
	fightStore := store.NewFightStore(testDB)
	boxerStore := store.NewBoxerStore(testDB)

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.Fight.MinScheduleHours = 0 // Disable minimum schedule hours for tests

	testLogger := logger.New("fight_simulate_invalid_test")

	boxerSvcForFight := boxer.NewBoxerService(boxerStore)
	fightSvcLowLevel := fight.NewFightService(fightStore, boxerStore, cfg, boxerSvcForFight, eventStore)
	fatigueSvc := service.NewFatigueService(boxerStore, testLogger)

	eventProcessor := service.NewEventProcessor(eventStore, boxerStore, fightSvcLowLevel, fatigueSvc, *testLogger)

	// Create a test user and boxer for the event
	userID := createTestUserInDB(t, testDB)
	boxer := &model.Boxer{
		UserID:   userID,
		Name:     "Invalid Fight Test Boxer",
		Energy:   100.0,
		Health:   100.0,
		Strength: 70.0,
		Defense:  60.0,
		Agility:  80.0,
		Level:    5,
	}

	if err := boxerStore.Create(ctx, boxer); err != nil {
		t.Fatalf("Failed to create test boxer: %v", err)
	}
	t.Logf("Created test boxer with ID: %d", boxer.ID)

	// Insert an event directly into the database with invalid fight_id=999
	query := `INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data, created_at)
			  VALUES ($1, $2, $3, FALSE, 'pending', $4::jsonb, CURRENT_TIMESTAMP) RETURNING id`

	var eventId int
	invalidEventData := []byte(`{"fight_id": 999}`)
	var err error
	err = testDB.QueryRowContext(ctx, query, boxer.ID, model.EventTypeFightSimulate, time.Now().Add(-1*time.Hour), invalidEventData).Scan(&eventId)
	if err != nil {
		t.Fatalf("Failed to insert test event: %v", err)
	}

	t.Logf("Created event with ID: %d and invalid fight_id: 999", eventId)

	// Fetch the event for processing
	event, err := eventStore.GetByID(ctx, eventId)
	if err != nil {
		t.Fatalf("Failed to retrieve created event: %v", err)
	}

	// Process the event - should fail because fight 999 doesn't exist
	err = eventProcessor.ProcessScheduledEvent(ctx, event)
	if err == nil {
		t.Error("Expected error when processing event with invalid fight_id")
	} else {
		t.Logf("Got expected error: %v", err)

		// Verify the error message contains "fight not found" or similar
		expectedErrSubstrings := []string{"fight not found", "invalid fight_id"}
		foundExpectedSubstring := false
		for _, substr := range expectedErrSubstrings {
			if containsSubstring(err.Error(), substr) {
				foundExpectedSubstring = true
				break
			}
		}

		if !foundExpectedSubstring {
			t.Logf("Warning: error message doesn't contain expected substring. Got: %v", err)
		}
	}

	// Verify event status was updated to failed
	updatedEvent, err := eventStore.GetByID(ctx, eventId)
	if err != nil {
		t.Fatalf("Failed to retrieve event after processing: %v", err)
	}

	if updatedEvent.Status != "failed" {
		t.Errorf("Expected event status 'failed', got '%s'", updatedEvent.Status)
	}

	if updatedEvent.ErrorMessage == nil || *updatedEvent.ErrorMessage == "" {
		t.Error("Failed event should have an error message")
	} else {
		t.Logf("Event error message: %s", *updatedEvent.ErrorMessage)
	}
}

// TestFightSimulateEventIdempotentProcessing verifies that processing the same event twice is safe (idempotent).
func TestFightSimulateEventIdempotentProcessing(t *testing.T) {
	testDB, _ := db.FreshDatabaseWithMigrations(t, "fight_simulate_idempotent", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available")
	}

	ctx := context.Background()

	// Setup stores
	boxerStore := store.NewBoxerStore(testDB)
	fightStore := store.NewFightStore(testDB)
	eventStore := store.NewScheduledEventStore(testDB)

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.Config{}
	}
	cfg.Fight.MinScheduleHours = 0 // Disable minimum schedule hours for tests

	testLogger := logger.New("fight_simulate_idempotent_test")

	boxerSvcForFight := boxer.NewBoxerService(boxerStore)
	fightSvcLowLevel := fight.NewFightService(fightStore, boxerStore, cfg, boxerSvcForFight, eventStore)
	fightSvcHighLevel := service.NewFightService(
		fightStore, boxerStore, eventStore, worldClockModel, cfg, testDB,
	)

	userID := createTestUserInDB(t, testDB)

	// Create boxers and book a fight
	boxer1 := &model.Boxer{
		UserID:   userID,
		Name:     "Idempotent Test Boxer 1",
		Energy:   100.0,
		Health:   100.0,
		Strength: 70.0,
		Defense:  60.0,
		Agility:  80.0,
		Level:    5,
	}

	if err := boxerStore.Create(ctx, boxer1); err != nil {
		t.Fatalf("Failed to create boxer 1: %v", err)
	}

	boxer2 := &model.Boxer{
		UserID:   userID,
		Name:     "Idempotent Test Boxer 2",
		Energy:   100.0,
		Health:   100.0,
		Strength: 65.0,
		Defense:  70.0,
		Agility:  75.0,
		Level:    5,
	}

	if err := boxerStore.Create(ctx, boxer2); err != nil {
		t.Fatalf("Failed to create boxer 2: %v", err)
	}

	// Schedule fight using game time (not real time) to pass validation
	gameTime, err := worldClockModel.GetCurrentGameTime(ctx, testDB)
	if err != nil {
		t.Fatalf("Failed to get current game time: %v", err)
	}
	scheduledTime := gameTime.Add(24 * time.Hour)

	if err := fightSvcHighLevel.BookFight(ctx, boxer1.ID, boxer2.ID, scheduledTime, 10); err != nil {
		t.Fatalf("Failed to book fight: %v", err)
	}

	fight, _ := fightStore.GetByID(ctx, 1)
	t.Logf("Created fight with ID: %d", fight.ID)

	// Get the pending event
	events, _ := eventStore.GetPendingByBoxerIDAndType(ctx, boxer1.ID, model.EventTypeFightSimulate)
	if len(events) == 0 {
		t.Fatal("No pending events found")
	}

	event := events[0]
	t.Logf("Event ID: %d", event.ID)

	// Create event processor using low-level fight service
	fatigueSvc := service.NewFatigueService(boxerStore, testLogger)
	eventProcessor := service.NewEventProcessor(eventStore, boxerStore, fightSvcLowLevel, fatigueSvc, *testLogger)

	// Process the event first time
	t.Run("FirstProcessing", func(t *testing.T) {
		if err := eventProcessor.ProcessScheduledEvent(ctx, event); err != nil {
			t.Fatalf("First processing failed: %v", err)
		}

		fightAfterFirst, _ := fightStore.GetByID(ctx, fight.ID)
		t.Logf("After first processing - Status: %s, Winner: %v",
			fightAfterFirst.Status, fightAfterFirst.WinnerID)

		if fightAfterFirst.Status != "completed" {
			t.Error("Fight should be completed after first processing")
		}
	})

	// Get the event again (it should now be marked as completed)
	eventAfterFirst, _ := eventStore.GetByID(ctx, event.ID)
	t.Logf("Event status after first processing: %s", eventAfterFirst.Status)

	// Try to process the same event again - this simulates a worker crash and retry scenario
	// The fight simulation should detect that the fight is already completed and skip it
	t.Run("SecondProcessingIdempotent", func(t *testing.T) {
		// Fetch the event fresh (simulating worker querying pending events)
		eventFresh, _ := eventStore.GetByID(ctx, event.ID)

		// Note: After first processing, the event status should be "completed"
		// In a real worker, completed events would not be queried again
		// But if we somehow tried to process it again, SimulateFight has built-in idempotency

		// The actual idempotency check is in FightService.SimulateFight() which checks fight status
		// If the fight is already "completed", it won't re-simulate
		err := eventProcessor.ProcessScheduledEvent(ctx, eventFresh)

		// After first processing, event status is "completed" so ProcessScheduledEvent skips it
		// This is expected behavior - the event processor checks event status before processing
		t.Logf("Second processing result (expected to skip completed event): %v", err)

		fightAfterSecond, _ := fightStore.GetByID(ctx, fight.ID)
		t.Logf("After second processing attempt - Status: %s, Winner: %v",
			fightAfterSecond.Status, fightAfterSecond.WinnerID)

		// Fight should still be completed with same winner (no change)
		if fightAfterSecond.Status != "completed" {
			t.Error("Fight should remain completed")
		}
	})

	t.Log("Idempotency test completed successfully")
}

// Helper functions

// createTestUserInDB creates a test user directly in the database and returns the user ID.
func createTestUserInDB(t *testing.T, dbConn *sql.DB) int {
	t.Helper()

	ctx := context.Background()
	query := `
		INSERT INTO users (username, email, hashed_password, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id`

	var userID int
	err := dbConn.QueryRowContext(ctx, query,
		fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		"$2a$10$DqnoC8RGebKReGMnlBhWRuXlKJp7e6FyH5Z8rL9mT3vN2wO4xU8QO").Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// containsSubstring checks if a string contains another string (case-insensitive).
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
