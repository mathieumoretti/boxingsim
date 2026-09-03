//go:build integration

package model

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
)

func init() {
	_ = pq.Driver{} // Register PostgreSQL driver
}

// TestGetAnchors_InitializeDefault tests that GetAnchors initializes defaults on first run.
func TestGetAnchors_InitializeDefault(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Clear world_clock table to simulate fresh migration
	_, err := db.Exec("DELETE FROM world_clock WHERE id = 1")
	if err != nil {
		t.Fatalf("Failed to clear world_clock: %v", err)
	}

	// Create model and fetch anchors (should initialize defaults)
	lg := logger.New("test")
	model := NewWorldClockModel(lg)
	ctx := context.Background()

	wc, err := model.GetAnchors(ctx, db)
	if err != nil {
		t.Fatalf("GetAnchors() failed: %v", err)
	}

	// Verify defaults were initialized
	if wc.ID != 1 {
		t.Errorf("Expected ID=1, got %d", wc.ID)
	}

	if wc.SpeedFactor != DefaultSpeedFactor {
		t.Errorf("Expected SpeedFactor=%f, got %f", DefaultSpeedFactor, wc.SpeedFactor)
	}

	if wc.Status != WorldClockRunning {
		t.Errorf("Expected Status=running, got %s", wc.Status)
	}

	// Verify game anchor matches default start time
	expectedGameAnchor, _ := time.Parse("2006-01-02 15:04:05", DefaultGameStart)
	if !wc.GameAnchor.Equal(expectedGameAnchor) {
		t.Errorf("Expected GameAnchor=%v, got %v", expectedGameAnchor, wc.GameAnchor)
	}

	// Real anchor should be close to now (within 1 second)
	timeDiff := time.Since(wc.RealAnchor)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("RealAnchor %v is too far from now", wc.RealAnchor)
	}
}

// TestGetCurrentGameTime tests the full calculation pipeline with database access.
func TestGetCurrentGameTime(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lg := logger.New("test")
	model := NewWorldClockModel(lg)
	ctx := context.Background()

	// Get current game time
	gameTime, err := model.GetCurrentGameTime(ctx, db)
	if err != nil {
		t.Fatalf("GetCurrentGameTime() failed: %v", err)
	}

	// Game time should be valid (not zero)
	if gameTime.IsZero() {
		t.Error("GetCurrentGameTime() returned zero time")
	}

	// Get anchors and verify calculation matches
	wc, _ := model.GetAnchors(ctx, db)
	expectedGameTime := wc.CalculateGameTime(time.Now())

	// Allow 1 second tolerance for clock drift
	diff := gameTime.Sub(expectedGameTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("GetCurrentGameTime() = %v, expected ~%v", gameTime, expectedGameTime)
	}
}

// TestPause_Resume tests pause and resume functionality.
func TestPause_Resume(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lg := logger.New("test")
	model := NewWorldClockModel(lg)
	ctx := context.Background()

	// Get initial game time
	initialGameTime, _ := model.GetCurrentGameTime(ctx, db)

	// Pause the clock
	err := model.Pause(ctx, db)
	if err != nil {
		t.Fatalf("Pause() failed: %v", err)
	}

	// Verify status is paused
	wc, _ := model.GetAnchors(ctx, db)
	if wc.Status != WorldClockPaused {
		t.Errorf("Expected Status=paused after Pause(), got %s", wc.Status)
	}

	// Game time should return the paused anchor (which is when we captured initialGameTime)
	pausedGameTime, _ := model.GetCurrentGameTime(ctx, db)
	// Due to pause handling returning game_anchor, it should be very close to initial time
	// The difference should be minimal (within execution time of a few ms)
	diff := pausedGameTime.Sub(initialGameTime)
	if diff < 0 {
		diff = -diff
	}
	// Allow 5 seconds tolerance for test execution and clock drift
	if diff > 5*time.Second {
		t.Errorf("Paused game time %v should be close to initial time %v (diff: %v)", pausedGameTime, initialGameTime, diff)
	}

	// Resume the clock
	err = model.Resume(ctx, db)
	if err != nil {
		t.Fatalf("Resume() failed: %v", err)
	}

	// Verify status is running again
	wc, _ = model.GetAnchors(ctx, db)
	if wc.Status != WorldClockRunning {
		t.Errorf("Expected Status=running after Resume(), got %s", wc.Status)
	}

	// Game time should advance after resume
	time.Sleep(100 * time.Millisecond)
	resumedGameTime, _ := model.GetCurrentGameTime(ctx, db)
	if !resumedGameTime.After(pausedGameTime) {
		t.Errorf("Resumed game time %v should be after paused time %v", resumedGameTime, pausedGameTime)
	}
}

// TestSetSpeedFactor tests speed factor change without discontinuity.
func TestSetSpeedFactor(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lg := logger.New("test")
	model := NewWorldClockModel(lg)
	ctx := context.Background()

	// Get initial state without using the value
	_, _ = model.GetAnchors(ctx, db)
	initialGameTime, _ := model.GetCurrentGameTime(ctx, db)

	// Wait a bit for time to advance
	time.Sleep(100 * time.Millisecond)

	// Change speed factor from 60x to 1x
	newSpeed := 1.0
	err := model.SetSpeedFactor(ctx, db, newSpeed)
	if err != nil {
		t.Fatalf("SetSpeedFactor() failed: %v", err)
	}

	// Verify new speed was applied
	newWc, _ := model.GetAnchors(ctx, db)
	if newWc.SpeedFactor != newSpeed {
		t.Errorf("Expected SpeedFactor=%f, got %f", newSpeed, newWc.SpeedFactor)
	}

	// Game time should have advanced (no discontinuity - no jump backward or forward)
	newGameTime, _ := model.GetCurrentGameTime(ctx, db)
	if !newGameTime.After(initialGameTime) {
		t.Errorf("Game time should advance after SetSpeedFactor: was %v, now %v", initialGameTime, newGameTime)
	}

	// The advancement should be minimal since we changed to 1x speed just before measuring
	// (the actual test of no-discontinuity is that game time never jumps)
}

// TestSetSpeedFactor_Negative tests error handling for negative speed.
func TestSetSpeedFactor_Negative(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lg := logger.New("test")
	model := NewWorldClockModel(lg)
	ctx := context.Background()

	err := model.SetSpeedFactor(ctx, db, -1.0)
	if err == nil {
		t.Error("SetSpeedFactor(-1.0) should return error for negative speed")
	}
}

// TestGetTimeUntilEvent tests real-world duration calculation until event.
func TestGetTimeUntilEvent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	lg := logger.New("test")
	model := NewWorldClockModel(lg)
	ctx := context.Background()

	// Get current world clock state to understand the speed factor
	wc, _ := model.GetAnchors(ctx, db)
	speedFactor := wc.SpeedFactor

	// Get current game time
	currentGameTime, _ := model.GetCurrentGameTime(ctx, db)

	// Calculate event 1 hour in the future (game time)
	eventGameTime := currentGameTime.Add(1 * time.Hour)

	// Get real-world duration until event
	duration, err := model.GetTimeUntilEvent(ctx, db, eventGameTime)
	if err != nil {
		t.Fatalf("GetTimeUntilEvent() failed: %v", err)
	}

	// At 60x speed (default), 1 game hour = 1 real minute (3600 game seconds / 60 = 60 real seconds)
	expectedRealSeconds := float64(time.Hour.Seconds()) / speedFactor
	expectedDuration := time.Duration(expectedRealSeconds * float64(time.Second))
	tolerance := time.Second

	diff := duration - expectedDuration
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("GetTimeUntilEvent() = %v (%.2f real seconds), expected ~%v (%.2f real seconds) at %.1fx speed",
			duration, duration.Seconds(), expectedDuration, expectedRealSeconds, speedFactor)
	}

	// Test past event (should return 0)
	pastEvent := currentGameTime.Add(-1 * time.Hour)
	durationPast, _ := model.GetTimeUntilEvent(ctx, db, pastEvent)
	if durationPast != 0 {
		t.Errorf("GetTimeUntilEvent() for past event should return 0, got %v", durationPast)
	}
}

// setupTestDB creates a test database connection using TEST_DB_* env vars.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	testCfg, err := config.LoadTestDBConfig()
	if err != nil {
		t.Fatalf("Failed to load test DB config: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		testCfg.Host,
		testCfg.Port,
		testCfg.User,
		testCfg.Password,
		"boxing", // Use main database for model tests (we delete world_clock row)
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database connection: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db, func() {
		db.Close()
	}
}
