//go:build functional

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// TestCheckZombieEventsForBoxer tests the zombie event detection (MAT-139)
func TestCheckZombieEventsForBoxer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupEvents      func(ctx context.Context, db *sql.DB, boxerID int)
		expectedHasZombie bool
		expectError      bool
	}{
		{
			name: "no scheduled events returns false",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				// No setup - no events exist
			},
			expectedHasZombie: false,
			expectError:       false,
		},
		{
			name: "completed scheduled event returns false",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				// Create a completed fight
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForTest(ctx, db, boxerID, boxerID+100, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				// Create a completed scheduled event for that fight
				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createCompletedEventForTest(ctx, db, boxerID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectedHasZombie: false,
			expectError:       false,
		},
		{
			name: "pending scheduled event with valid fight returns true",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForTest(ctx, db, boxerID, boxerID+100, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createPendingEventForTest(ctx, db, boxerID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectedHasZombie: true,
			expectError:       false,
		},
		{
			name: "failed scheduled event with valid fight returns true",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForTest(ctx, db, boxerID, boxerID+100, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createFailedEventForTest(ctx, db, boxerID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectedHasZombie: true,
			expectError:       false,
		},
		{
			name: "processing scheduled event returns true",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForTest(ctx, db, boxerID, boxerID+100, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createProcessingEventForTest(ctx, db, boxerID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectedHasZombie: true,
			expectError:       false,
		},
		{
			name: "pending event for completed fight returns false",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForTest(ctx, db, boxerID, boxerID+100, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				// Mark the fight as completed
				updateFightStatusForTest(ctx, db, fightID, "completed")

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createPendingEventForTest(ctx, db, boxerID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectedHasZombie: false,
			expectError:       false,
		},
		{
			name: "event for different boxer returns false",
			setupEvents: func(ctx context.Context, db *sql.DB, boxerID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				// Create fight with different boxers
				fightID, err := createFightForTest(ctx, db, boxerID+50, boxerID+60, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createPendingEventForTest(ctx, db, boxerID+50, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectedHasZombie: false,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := setupTestDatabase(t)
			defer cleanup()

			store := NewFightStore(testDB)
			boxerID := 1001

			// Create test boxers
			createTestBoxer(t, ctx, testDB, boxerID)
			createTestBoxer(t, ctx, testDB, boxerID+100)

			// Setup test events
			tt.setupEvents(ctx, testDB, boxerID)

			// Run the check
			hasZombie, err := store.CheckZombieEventsForBoxer(ctx, boxerID, time.Now())

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if hasZombie != tt.expectedHasZombie {
				t.Errorf("expected hasZombie=%v, got %v", tt.expectedHasZombie, hasZombie)
			}
		})
	}
}

// Helper functions for test setup

func createTestBoxer(t *testing.T, ctx context.Context, db *sql.DB, boxerID int) {
	query := `
		INSERT INTO boxers (id, name, nickname, strength, defense, agility, level, health, fatigue, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING`

	_, err := db.ExecContext(ctx, query,
		boxerID,
		"Test Boxer",
		"Test Nickname",
		50, 50, 50,
		10, 100, 0, "active",
	)
	if err != nil {
		t.Logf("Could not create boxer %d (may already exist): %v", boxerID, err)
	}
}

func createFightForTest(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int, scheduledTime *time.Time) (int, error) {
	query := `
		INSERT INTO fights (boxer1_id, boxer2_id, status, scheduled_time, round)
		VALUES ($1, $2, 'scheduled', $3, $4)
		RETURNING id`

	var fightID int
	err := db.QueryRowContext(ctx, query, boxer1ID, boxer2ID, scheduledTime, 12).Scan(&fightID)
	return fightID, err
}

func updateFightStatusForTest(ctx context.Context, db *sql.DB, fightID int, status string) error {
	query := `UPDATE fights SET status = $1 WHERE id = $2`
	_, err := db.ExecContext(ctx, query, status, fightID)
	return err
}

func createPendingEventForTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data)
		VALUES ($1, $2, $3, FALSE, 'pending', $4)`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData)
	if err != nil {
		// Event may already exist
	}
}

func createFailedEventForTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data, error_message)
		VALUES ($1, $2, $3, FALSE, 'failed', $4, 'Simulated failure')`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData, "Simulated failure")
	if err != nil {
		// Event may already exist
	}
}

func createProcessingEventForTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data)
		VALUES ($1, $2, $3, FALSE, 'processing', $4)`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData)
	if err != nil {
		// Event may already exist
	}
}

func createCompletedEventForTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data)
		VALUES ($1, $2, $3, TRUE, 'completed', $4)`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData)
	if err != nil {
		// Event may already exist
	}
}

func setupTestDatabase(t *testing.T) (*sql.DB, func()) {
	testDBHost := mustGetEnv("TEST_DB_HOST", "localhost")
	testDBPort := mustGetEnv("TEST_DB_PORT", "5433")
	testDBUser := mustGetEnv("TEST_DB_USER", "testuser")
	testDBPass := mustGetEnv("TEST_DB_PASSWORD", "testpass123")
	testDBName := mustGetEnv("TEST_DB_NAME", "boxing_test")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPass, testDBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	return db, func() { db.Close() }
}

func mustGetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
