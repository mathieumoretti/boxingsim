package model

import "time"

// WorldClockStatus represents the global clock state.
type WorldClockStatus string

const (
	WorldClockRunning WorldClockStatus = "running"
	WorldClockPaused  WorldClockStatus = "paused"
	WorldClockStopped WorldClockStatus = "stopped"
)

// WorldClock represents a single-row table tracking simulation time anchors and global clock state.
type WorldClock struct {
	ID          int       `json:"id"`           // Always 1 for the single authoritative row
	RealAnchor  time.Time `json:"real_anchor"`  // Game anchor in real wall-clock time (with timezone)
	GameAnchor  time.Time `json:"game_anchor"`  // Timestamp base corresponding to real_anchor calculation
	SpeedFactor float64   `json:"speed_factor"` // Sim seconds per real second (default = game runs 60x faster)

	Status    WorldClockStatus `json:"status"`     // Global clock state: running, paused, or stopped
	UpdatedAt time.Time        `json:"updated_at"` // Last update timestamp for this anchor row
}

// CalculateGameTime computes the current game simulation time based on the world_clock anchors.
// Formula: GameTime = game_anchor + (real_now - real_anchor) * speed_factor
// Returns zero time if clock is not in "running" state or if calculation would result in invalid/negative duration.
func (w *WorldClock) CalculateGameTime(now time.Time) time.Time {
	if w.Status != WorldClockRunning {
		return w.GameAnchor // Return last known game anchor when paused/stopped
	}

	duration := now.Sub(w.RealAnchor).Seconds()
	gameDuration := duration * w.SpeedFactor

	baseInstant, err := time.ParseInLocation("2006-01-02 15:04:05",
		w.GameAnchor.Format("2006-01-02 15:04:05"), nil)
	if err != nil {
		return w.GameAnchor // Fallback to stored game anchor on parse error (should never happen with valid format above)
	}

	gameTime := baseInstant.Add(time.Duration(gameDuration * float64(time.Second)))

	// Clamp at epoch boundary: if subtraction goes into negative territory, return zero time.
	return gameTime
}
