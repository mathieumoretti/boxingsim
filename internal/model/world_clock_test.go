package model

import (
	"testing"
	"time"
)

// TestCalculateGameTime_Running tests game time calculation when clock is running at various speeds.
func TestCalculateGameTime_Running(t *testing.T) {
	tests := []struct {
		name         string
		speedFactor  float64
		realNow      time.Time
		expectAfter  bool // Expect game time to be after anchor
	}{
		{
			name:        "60x speed advances game time faster",
			speedFactor: 60.0,
			realNow:     mustParseTime("2030-01-01 08:01:00"), // 1 real minute after anchor
		},
		{
			name:        "1x speed advances game time at real rate",
			speedFactor: 1.0,
			realNow:     mustParseTime("2030-01-01 08:01:00"), // 1 real minute after anchor
		},
		{
			name:        "0.5x speed advances game time slower",
			speedFactor: 0.5,
			realNow:     mustParseTime("2030-01-01 08:02:00"), // 2 real minutes after anchor
		},
		{
			name:        "100x speed for fast simulation",
			speedFactor: 100.0,
			realNow:     mustParseTime("2030-01-01 08:00:30"), // 30 real seconds after anchor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create world clock with known anchors
			realAnchor := mustParseTime("2030-01-01 08:00:00")
			gameAnchor := mustParseTime("2030-01-01 08:00:00")

			wc := &WorldClock{
				ID:          1,
				RealAnchor:  realAnchor,
				GameAnchor:  gameAnchor,
				SpeedFactor: tt.speedFactor,
				Status:      WorldClockRunning,
			}

			gameTime := wc.CalculateGameTime(tt.realNow)

			// Verify calculation
			elapsedRealSeconds := tt.realNow.Sub(realAnchor).Seconds()
			expectedGameSeconds := elapsedRealSeconds * tt.speedFactor
			expectedGameTime := gameAnchor.Add(time.Duration(expectedGameSeconds) * time.Second)

			if !gameTime.Equal(expectedGameTime) {
				t.Errorf("CalculateGameTime() = %v, want %v", gameTime, expectedGameTime)
			}

			// Verify game time advanced
			if tt.expectAfter && !gameTime.After(gameAnchor) {
				t.Error("Expected game time to be after anchor")
			}
		})
	}
}

// TestCalculateGameTime_Paused tests that paused clock returns last known anchor.
func TestCalculateGameTime_Paused(t *testing.T) {
	realAnchor := mustParseTime("2030-01-01 08:00:00")
	gameAnchor := mustParseTime("2030-01-01 08:00:00")

	wc := &WorldClock{
		ID:          1,
		RealAnchor:  realAnchor,
		GameAnchor:  gameAnchor,
		SpeedFactor: 60.0,
		Status:      WorldClockPaused,
	}

	futureTime := mustParseTime("2030-01-01 10:00:00") // 2 hours in the future
	gameTime := wc.CalculateGameTime(futureTime)

	// Should return game anchor when paused
	if !gameTime.Equal(gameAnchor) {
		t.Errorf("CalculateGameTime() = %v, want %v (paused should return anchor)", gameTime, gameAnchor)
	}
}

// TestCalculateGameTime_Stopped tests that stopped clock returns last known anchor.
func TestCalculateGameTime_Stopped(t *testing.T) {
	realAnchor := mustParseTime("2030-01-01 08:00:00")
	gameAnchor := mustParseTime("2030-01-01 08:00:00")

	wc := &WorldClock{
		ID:          1,
		RealAnchor:  realAnchor,
		GameAnchor:  gameAnchor,
		SpeedFactor: 60.0,
		Status:      WorldClockStopped,
	}

	futureTime := mustParseTime("2030-01-01 10:00:00")
	gameTime := wc.CalculateGameTime(futureTime)

	if !gameTime.Equal(gameAnchor) {
		t.Errorf("CalculateGameTime() = %v, want %v (stopped should return anchor)", gameTime, gameAnchor)
	}
}

// TestCalculateGameTime_NegativeElapsed tests handling of negative elapsed time.
func TestCalculateGameTime_NegativeElapsed(t *testing.T) {
	realAnchor := mustParseTime("2030-01-01 10:00:00") // Future anchor
	gameAnchor := mustParseTime("2030-01-01 08:00:00")

	wc := &WorldClock{
		ID:          1,
		RealAnchor:  realAnchor,
		GameAnchor:  gameAnchor,
		SpeedFactor: 60.0,
		Status:      WorldClockRunning,
	}

	pastTime := mustParseTime("2030-01-01 08:00:00") // Before real anchor
	gameTime := wc.CalculateGameTime(pastTime)

	// Should return game anchor when elapsed time is negative
	if !gameTime.Equal(gameAnchor) {
		t.Errorf("CalculateGameTime() = %v, want %v (negative elapsed should return anchor)", gameTime, gameAnchor)
	}
}

// TestCalculateGameTime_ZeroSpeed tests zero speed factor.
func TestCalculateGameTime_ZeroSpeed(t *testing.T) {
	realAnchor := mustParseTime("2030-01-01 08:00:00")
	gameAnchor := mustParseTime("2030-01-01 08:00:00")

	wc := &WorldClock{
		ID:          1,
		RealAnchor:  realAnchor,
		GameAnchor:  gameAnchor,
		SpeedFactor: 0.0,
		Status:      WorldClockRunning,
	}

	futureTime := mustParseTime("2030-01-01 10:00:00")
	gameTime := wc.CalculateGameTime(futureTime)

	// Zero speed means no advancement
	if !gameTime.Equal(gameAnchor) {
		t.Errorf("CalculateGameTime() = %v, want %v (zero speed should not advance)", gameTime, gameAnchor)
	}
}

// TestCalculateGameTime_LargeSpeed tests extreme speed factors.
func TestCalculateGameTime_LargeSpeed(t *testing.T) {
	realAnchor := mustParseTime("2030-01-01 08:00:00")
	gameAnchor := mustParseTime("2030-01-01 08:00:00")

	wc := &WorldClock{
		ID:          1,
		RealAnchor:  realAnchor,
		GameAnchor:  gameAnchor,
		SpeedFactor: 3600.0, // 1 hour = 1 day (3600x speed)
		Status:      WorldClockRunning,
	}

	// 1 second of real time at 3600x speed = 1 hour of game time
	oneSecondLater := realAnchor.Add(1 * time.Second)
	gameTime := wc.CalculateGameTime(oneSecondLater)

	expected := gameAnchor.Add(1 * time.Hour)
	if !gameTime.Equal(expected) {
		t.Errorf("CalculateGameTime() = %v, want %v", gameTime, expected)
	}
}

// Helper function to parse time without error handling in tests.
func mustParseTime(layout string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", layout)
	if err != nil {
		panic(err)
	}
	return t
}
