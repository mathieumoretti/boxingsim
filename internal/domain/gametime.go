package domain

import (
	"fmt"
	"time"
)

// GameTime represents a point in simulated game time, distinct from real wall-clock time.
// It wraps time.Time and provides type safety to prevent mixing real time with game time.
//
// GameTime is used for:
//   - Scheduled event times (training completion, fight scheduling)
//   - World clock calculations
//   - Comparing timestamps within the game simulation
//
// Important: GameTime should NEVER be used for real-world deadlines or timeouts.
// Always use time.Time (context.Context with real deadlines) when interacting with external systems.
type GameTime struct {
	time time.Time
}

// NowGameTime creates a GameTime from the provided time.Time value.
// The input is assumed to already be in game time (calculated by WorldClock).
func NowGameTime(t time.Time) GameTime {
	return GameTime{time: t}
}

// FromTime converts a real wall-clock time.Time to GameTime.
// Use only when you explicitly want to treat real time as game time (e.g., test fixtures).
// For normal operation, use NowGameTime() with times from WorldClock.CalculateGameTime().
func FromTime(t time.Time) GameTime {
	return GameTime{time: t}
}

// AsTime returns the underlying time.Time value.
// Use this when you need to pass the value to SQL or JSON serialization.
// The returned time.Time should be treated as immutable game time.
func (g GameTime) AsTime() time.Time {
	return g.time
}

// IsZero returns true if the GameTime is the zero value.
func (g GameTime) IsZero() bool {
	return g.time.IsZero()
}

// Before reports whether the GameTime g is before o.
func (g GameTime) Before(o GameTime) bool {
	return g.time.Before(o.time)
}

// After reports whether the GameTime g is after o.
func (g GameTime) After(o GameTime) bool {
	return g.time.After(o.time)
}

// Equal reports whether the GameTime g equals o.
func (g GameTime) Equal(o GameTime) bool {
	return g.time.Equal(o.time)
}

// Compare compares two GameTimes:
// Returns -1 if g < o, 0 if g == o, +1 if g > o
func (g GameTime) Compare(o GameTime) int {
	return g.time.Compare(o.time)
}

// Add returns a GameTime with duration d added.
// Useful for calculating future game times (e.g., training completion time).
func (g GameTime) Add(d time.Duration) GameTime {
	return GameTime{time: g.time.Add(d)}
}

// Sub subtracts duration d and returns the resulting GameTime.
// Useful for calculating past game times.
func (g GameTime) Sub(d time.Duration) GameTime {
	return GameTime{time: g.time.Add(-d)}
}

// SubGameTime calculates the duration between two GameTimes.
// Returns positive if g is after o, negative if before.
func (g GameTime) SubGameTime(o GameTime) time.Duration {
	return g.time.Sub(o.time)
}

// Format formats the GameTime using RFC3339 format.
func (g GameTime) Format() string {
	return g.time.Format(time.RFC3339)
}

// String returns a human-readable representation for debugging.
func (g GameTime) String() string {
	if g.IsZero() {
		return "GameTime{<zero>}"
	}
	return fmt.Sprintf("GameTime{%s}", g.time.Format("2006-01-02 15:04:05"))
}

// Hour returns the hour component of the GameTime (0-23).
func (g GameTime) Hour() int {
	return g.time.Hour()
}

// Minute returns the minute component of the GameTime (0-59).
func (g GameTime) Minute() int {
	return g.time.Minute()
}

// Second returns the second component of the GameTime (0-59).
func (g GameTime) Second() int {
	return g.time.Second()
}

// Day returns the day of the month (1-31).
func (g GameTime) Day() int {
	return g.time.Day()
}

// Month returns the month of the year.
func (g GameTime) Month() time.Month {
	return g.time.Month()
}

// Year returns the year.
func (g GameTime) Year() int {
	return g.time.Year()
}

// TruncateSecond truncates to second precision (removes nanoseconds).
// Useful for database comparisons where nanosecond differences matter.
func (g GameTime) TruncateSecond() GameTime {
	return GameTime{time: g.time.Truncate(time.Second)}
}

// IsToday returns true if the GameTime is on the same calendar day as the given reference time.
func (g GameTime) IsToday(ref time.Time) bool {
	refDate := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
	gDate := time.Date(g.time.Year(), g.time.Month(), g.time.Day(), 0, 0, 0, 0, g.time.Location())
	return refDate.Equal(gDate)
}

// In returns a GameTime in the specified location (timezone).
func (g GameTime) In(loc *time.Location) GameTime {
	return GameTime{time: g.time.In(loc)}
}

// UTC converts the GameTime to UTC.
func (g GameTime) UTC() GameTime {
	return GameTime{time: g.time.UTC()}
}
