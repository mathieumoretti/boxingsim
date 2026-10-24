//go:build functional

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/store"
)

// TestBookFightZombieEventPrevention tests that fight booking is prevented when zombie events exist (MAT-139)
func TestBookFightZombieEventPrevention(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupEvents   func(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int)
		expectError   bool
		errorContains string
	}{
		{
			name: "booking fight with no zombie events succeeds",
			setupEvents: func(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int) {
				// No setup - clean state
			},
			expectError:   false,
			errorContains: "",
		},
		{
			name: "booking fight with pending zombie event for boxer1 fails",
			setupEvents: func(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForServiceTest(ctx, db, boxer1ID, boxer2ID, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createPendingEventForServiceTest(ctx, db, boxer1ID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectError:   true,
			errorContains: "unresolved scheduled event",
		},
		{
			name: "booking fight with failed zombie event for boxer2 fails",
			setupEvents: func(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForServiceTest(ctx, db, boxer1ID, boxer2ID, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createFailedEventForServiceTest(ctx, db, boxer2ID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectError:   true,
			errorContains: "unresolved scheduled event",
		},
		{
			name: "booking fight with completed events succeeds",
			setupEvents: func(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int) {
				scheduledTime := time.Now().Add(24 * time.Hour)
				fightID, err := createFightForServiceTest(ctx, db, boxer1ID, boxer2ID, &scheduledTime)
				if err != nil {
					t.Fatalf("Failed to create fight: %v", err)
				}

				eventData, _ := json.Marshal(map[string]any{"fight_id": fightID})
				createCompletedEventForServiceTest(ctx, db, boxer1ID, model.EventTypeFightSimulate, scheduledTime, eventData)
			},
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testDB, cleanup := setupTestDatabaseForService(t)
			defer cleanup()

			fightStore := store.NewFightStore(testDB)
			boxerStore := store.NewBoxerStore(testDB)
			eventStore := store.NewScheduledEventStore(testDB)

			cfg := &config.Config{
				Fight: config.FightConfig{
					MinScheduleHours: 24,
				},
			}

			service := NewFightService(fightStore, boxerStore, eventStore, nil, cfg, testDB)

			boxer1ID := 2001
			boxer2ID := 2002

			// Create test boxers
			createTestBoxerForService(t, ctx, testDB, boxer1ID)
			createTestBoxerForService(t, ctx, testDB, boxer2ID)

			// Setup zombie events if needed
			tt.setupEvents(ctx, testDB, boxer1ID, boxer2ID)

			scheduledTime := time.Now().Add(48 * time.Hour)

			err := service.BookFight(ctx, boxer1ID, boxer2ID, scheduledTime, 12)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %v", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper functions for service tests

func createTestBoxerForService(t *testing.T, ctx context.Context, db *sql.DB, boxerID int) {
	query := `
		INSERT INTO boxers (id, name, nickname, strength, defense, agility, level, health, fatigue, status, wins, losses, draws, knockouts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO NOTHING`

	_, err := db.ExecContext(ctx, query,
		boxerID,
		fmt.Sprintf("Test Boxer %d", boxerID),
		fmt.Sprintf("Nickname %d", boxerID),
		50, 50, 50,
		10, 100, 0, "active",
		0, 0, 0, 0,
	)
	if err != nil {
		t.Logf("Could not create boxer %d (may already exist): %v", boxerID, err)
	}
}

func createFightForServiceTest(ctx context.Context, db *sql.DB, boxer1ID, boxer2ID int, scheduledTime *time.Time) (int, error) {
	query := `
		INSERT INTO fights (boxer1_id, boxer2_id, status, scheduled_time, round)
		VALUES ($1, $2, 'scheduled', $3, $4)
		RETURNING id`

	var fightID int
	err := db.QueryRowContext(ctx, query, boxer1ID, boxer2ID, scheduledTime, 12).Scan(&fightID)
	return fightID, err
}

func createPendingEventForServiceTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data)
		VALUES ($1, $2, $3, FALSE, 'pending', $4)`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData)
	if err != nil {
		// Event may already exist
	}
}

func createFailedEventForServiceTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data, error_message)
		VALUES ($1, $2, $3, FALSE, 'failed', $4, 'Simulated failure')`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData, "Simulated failure")
	if err != nil {
		// Event may already exist
	}
}

func createCompletedEventForServiceTest(ctx context.Context, db *sql.DB, boxerID int, eventType model.EventType, eventTime time.Time, eventData []byte) {
	query := `
		INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data)
		VALUES ($1, $2, $3, TRUE, 'completed', $4)`

	_, err := db.ExecContext(ctx, query, boxerID, eventType, eventTime, eventData)
	if err != nil {
		// Event may already exist
	}
}

func setupTestDatabaseForService(t *testing.T) (*sql.DB, func()) {
	testDBHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	testDBPort := getEnvOrDefault("TEST_DB_PORT", "5433")
	testDBUser := getEnvOrDefault("TEST_DB_USER", "testuser")
	testDBPass := getEnvOrDefault("TEST_DB_PASSWORD", "testpass123")
	testDBName := getEnvOrDefault("TEST_DB_NAME", "boxing_test")

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

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
