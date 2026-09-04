package model

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mormm/boxing/internal/platform/logger"
)

// WorldClockStatus represents the global clock state.
type WorldClockStatus string

const (
	WorldClockRunning WorldClockStatus = "running"
	WorldClockPaused  WorldClockStatus = "paused"
	WorldClockStopped WorldClockStatus = "stopped"
)

// Default constants for world clock initialization.
const (
	DefaultGameStart   = "2030-01-01 08:00:00" // Boxing simulation year
	DefaultSpeedFactor = 60.0                   // 1 real minute = 1 game hour (60x speed)
)

// WorldClock represents a single-row table tracking simulation time anchors and global clock state.
type WorldClock struct {
	ID          int              `json:"id"`           // Always 1 for the single authoritative row
	RealAnchor  time.Time        `json:"real_anchor"`  // Game anchor in real wall-clock time (with timezone)
	GameAnchor  time.Time        `json:"game_anchor"`  // Timestamp base corresponding to real_anchor calculation
	SpeedFactor float64          `json:"speed_factor"` // Sim seconds per real second (default = game runs 60x faster)
	Status      WorldClockStatus `json:"status"`       // Global clock state: running, paused, or stopped
	UpdatedAt   time.Time        `json:"updated_at"`   // Last update timestamp for this anchor row
}

// WorldClockModel provides methods to calculate and manage game time from database anchors.
// It stores NO internal state - all data is persisted in the database and reconstructed per query.
type WorldClockModel struct {
	logger *logger.Logger
}

// NewWorldClockModel creates a new WorldClockModel instance.
func NewWorldClockModel(logger *logger.Logger) *WorldClockModel {
	return &WorldClockModel{
		logger: logger,
	}
}

// GetAnchors fetches the single authoritative row from world_clock table.
// Handles initialization on first run by inserting default values if table is empty.
func (w *WorldClockModel) GetAnchors(ctx context.Context, db *sql.DB) (*WorldClock, error) {
	query := `
		SELECT
			id, real_anchor, game_anchor, speed_factor, status, updated_at
		FROM world_clock
		WHERE id = 1
	`

	wc := &WorldClock{}
	row := db.QueryRowContext(ctx, query)

	err := row.Scan(&wc.ID, &wc.RealAnchor, &wc.GameAnchor, &wc.SpeedFactor, &wc.Status, &wc.UpdatedAt)
	if err == sql.ErrNoRows {
		// Table exists but no row - insert defaults
		return w.initializeDefaultAnchors(ctx, db)
	}
	if err != nil {
		// Check if table doesn't exist yet (migration not run)
		if containsTableNotFound(err.Error()) {
			w.logger.Debug("world_clock table not found, will be created by migration")
			return nil, errors.New("world_clock table not found - run migrations")
		}
		return nil, err
	}

	return wc, nil
}

// initializeDefaultAnchors inserts the default world clock row and returns it.
func (w *WorldClockModel) initializeDefaultAnchors(ctx context.Context, db *sql.DB) (*WorldClock, error) {
	gameAnchor, err := time.Parse("2006-01-02 15:04:05", DefaultGameStart)
	if err != nil {
		return nil, err
	}

	insertQuery := `
		INSERT INTO world_clock (id, real_anchor, game_anchor, speed_factor, status, updated_at)
		VALUES (1, NOW(), $1, $2, 'running', NOW())
		RETURNING id, real_anchor, game_anchor, speed_factor, status, updated_at
	`

	wc := &WorldClock{}
	err = db.QueryRowContext(ctx, insertQuery, gameAnchor, DefaultSpeedFactor).Scan(
		&wc.ID, &wc.RealAnchor, &wc.GameAnchor, &wc.SpeedFactor, &wc.Status, &wc.UpdatedAt)
	if err != nil {
		return nil, err
	}

	w.logger.Info("Initialized world clock with defaults: game_anchor=%v, speed_factor=%.1f", wc.GameAnchor, wc.SpeedFactor)
	return wc, nil
}

// GetCurrentGameTime calculates the current simulated world time using the anchor formula.
// Formula: GameTime = GameAnchor + (RealNow - RealAnchor) × SpeedFactor
// Returns the paused game time if status is 'paused'.
func (w *WorldClockModel) GetCurrentGameTime(ctx context.Context, db *sql.DB) (time.Time, error) {
	wc, err := w.GetAnchors(ctx, db)
	if err != nil {
		return time.Time{}, err
	}

	return wc.CalculateGameTime(time.Now()), nil
}

// CalculateGameTime computes the current game simulation time based on the world_clock anchors.
// This is a pure calculation method that doesn't require database access - use with a fetched WorldClock instance.
// Formula: GameTime = game_anchor + (real_now - real_anchor) * speed_factor
// Returns zero time if clock is not in "running" state or if calculation would result in invalid/negative duration.
func (wc *WorldClock) CalculateGameTime(realNow time.Time) time.Time {
	if wc.Status != WorldClockRunning {
		return wc.GameAnchor // Return last known game anchor when paused/stopped
	}

	// Calculate elapsed real time since anchor
	elapsedRealSeconds := realNow.Sub(wc.RealAnchor).Seconds()

	// Handle negative elapsed time (if system clock went backward or anchors are in future)
	if elapsedRealSeconds < 0 {
		return wc.GameAnchor
	}

	// Apply speed factor to get game time elapsed
	elapsedGameSeconds := elapsedRealSeconds * wc.SpeedFactor

	// Calculate current game time
	gameTime := wc.GameAnchor.Add(time.Duration(elapsedGameSeconds * float64(time.Second)))

	// Safety check: if calculation results in a time before the anchor, return zero time
	if gameTime.Before(wc.GameAnchor) {
		return time.Time{}
	}

	return gameTime
}

// Pause sets the world clock status to 'paused' and records the current time as the pause point.
func (w *WorldClockModel) Pause(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		UPDATE world_clock
		SET status = 'paused', updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`

	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("failed to pause world clock - no row found")
	}

	return tx.Commit()
}

// Resume sets the world clock status back to 'running' and creates fresh anchors
// to ensure no discontinuity in the timeline.
func (w *WorldClockModel) Resume(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// First, get current anchors to calculate the resume point
	var pausedGameAnchor time.Time
	var pausedRealAnchor time.Time
	var speedFactor float64

	err = tx.QueryRowContext(ctx, `
		SELECT game_anchor, real_anchor, speed_factor
		FROM world_clock WHERE id = 1
	`).Scan(&pausedGameAnchor, &pausedRealAnchor, &speedFactor)
	if err != nil {
		return err
	}

	// Calculate what the current game time would be if we hadn't paused
	realNow := time.Now()
	elapsedRealSeconds := realNow.Sub(pausedRealAnchor).Seconds()
	gameTimeAtResume := pausedGameAnchor.Add(time.Duration(elapsedRealSeconds*speedFactor) * time.Second)

	// Update with fresh anchors and running status
	query := `
		UPDATE world_clock
		SET real_anchor = $1, game_anchor = $2, status = 'running', updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`

	_, err = tx.ExecContext(ctx, query, realNow, gameTimeAtResume)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// SetSpeedFactor updates the simulation speed factor while maintaining game time continuity.
// Creates fresh anchor pair at transition moment to ensure no jump in game time.
func (w *WorldClockModel) SetSpeedFactor(ctx context.Context, db *sql.DB, newSpeed float64) error {
	if newSpeed < 0 {
		return errors.New("speed factor cannot be negative")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Get current anchors
	var oldRealAnchor, oldGameAnchor time.Time
	var oldSpeedFactor float64
	var status WorldClockStatus

	err = tx.QueryRowContext(ctx, `
		SELECT real_anchor, game_anchor, speed_factor, status
		FROM world_clock WHERE id = 1
	`).Scan(&oldRealAnchor, &oldGameAnchor, &oldSpeedFactor, &status)
	if err != nil {
		return err
	}

	// Calculate current game time at transition (using old speed)
	realNow := time.Now()
	elapsedRealSeconds := realNow.Sub(oldRealAnchor).Seconds()
	currentGameTime := oldGameAnchor.Add(time.Duration(elapsedRealSeconds*oldSpeedFactor) * time.Second)

	// Update with fresh anchors and new speed - no discontinuity!
	query := `
		UPDATE world_clock
		SET real_anchor = $1, game_anchor = $2, speed_factor = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`

	_, err = tx.ExecContext(ctx, query, realNow, currentGameTime, newSpeed)
	if err != nil {
		return err
	}

	w.logger.Info("World clock speed changed: %.1f -> %.1f", oldSpeedFactor, newSpeed)
	return tx.Commit()
}

// GetTimeUntilEvent returns the remaining real-world duration until a scheduled event becomes due.
// Uses current game time calculation and converts back to real seconds using speed factor.
func (w *WorldClockModel) GetTimeUntilEvent(ctx context.Context, db *sql.DB, eventGameTime time.Time) (time.Duration, error) {
	wc, err := w.GetAnchors(ctx, db)
	if err != nil {
		return 0, err
	}

	currentGameTime := wc.CalculateGameTime(time.Now())

	// Calculate difference in game time
	diff := eventGameTime.Sub(currentGameTime)

	// Convert game time difference to real time using speed factor
	realSeconds := diff.Seconds() / wc.SpeedFactor

	// Handle edge cases
	if realSeconds < 0 {
		return 0, nil // Event is in the past
	}

	return time.Duration(realSeconds * float64(time.Second)), nil
}

// containsTableNotFound checks if error indicates table doesn't exist.
func containsTableNotFound(errMsg string) bool {
	return len(errMsg) > 0 &&
		(errMsg == "relation \"world_clock\" does not exist" ||
			containsSubstring(errMsg, "does not exist"))
}

// containsSubstring is a simple substring checker.
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
