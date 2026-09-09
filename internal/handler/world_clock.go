package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mormm/boxing/internal/model"
)

// WorldClockHandler handles world clock related requests.
type WorldClockHandler struct {
	db *sql.DB
}

// NewWorldClockHandler creates a new WorldClockHandler instance.
func NewWorldClockHandler(db *sql.DB) *WorldClockHandler {
	return &WorldClockHandler{
		db: db,
	}
}

// GetCurrentGameTimeResponse represents the response structure for the /world/time endpoint.
type GetCurrentGameTimeResponse struct {
	CurrentGameTime    string  `json:"current_game_time"`
	FormattedTime      string  `json:"formatted_time"`
	Status             string  `json:"status"`
	SpeedFactor        float64 `json:"speed_factor"`
	GameAnchor         string  `json:"game_anchor"`
	RealAnchor         string  `json:"real_anchor"`
	UpdatedAt          string  `json:"updated_at"`
	ClockRunning       bool    `json:"clock_running"`
	SecondsPerGameHour int     `json:"seconds_per_game_hour"`
	TimeSinceStart     string  `json:"time_since_start"`
}

// GetCurrentGameTime handles GET /world/time requests.
// Returns the current simulated game time and clock status.
func (h *WorldClockHandler) GetCurrentGameTime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Initialize world clock model
	clockModel := &model.WorldClockModel{}

	// Get current game time
	currentTime, err := clockModel.GetCurrentGameTime(ctx, h.db)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get game time: " + err.Error(),
		})
		return
	}

	// Get clock anchors for status and metadata
	anchors, err := clockModel.GetAnchors(ctx, h.db)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get clock status: " + err.Error(),
		})
		return
	}

	// Calculate time since game start
	timeSinceStart := currentTime.Sub(anchors.GameAnchor)
	var timeSinceStartStr string
	if timeSinceStart.Hours() >= 24 {
		days := int(timeSinceStart.Hours() / 24)
		hours := int(timeSinceStart.Hours()) % 24
		timeSinceStartStr = "Day " + itoa(days) + ", " + itoa(hours) + "h"
	} else {
		hours := int(timeSinceStart.Hours())
		minutes := int(timeSinceStart.Minutes()) % 60
		timeSinceStartStr = itoa(hours) + "h " + itoa(minutes) + "m"
	}

	// Calculate real seconds per game hour (for UI to show timing)
	secondsPerGameHour := int(3600 / anchors.SpeedFactor)

	// Format the response
	response := GetCurrentGameTimeResponse{
		CurrentGameTime:    currentTime.Format(time.RFC3339),
		FormattedTime:      formatReadableTime(currentTime),
		Status:             string(anchors.Status),
		SpeedFactor:        anchors.SpeedFactor,
		GameAnchor:         anchors.GameAnchor.Format(time.RFC3339),
		RealAnchor:         anchors.RealAnchor.Format(time.RFC3339),
		UpdatedAt:          anchors.UpdatedAt.Format(time.RFC3339),
		ClockRunning:       anchors.Status == model.WorldClockRunning,
		SecondsPerGameHour: secondsPerGameHour,
		TimeSinceStart:     timeSinceStartStr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// formatReadableTime formats a time into human-readable form like "March 15, 2030 at 2:30 PM"
func formatReadableTime(t time.Time) string {
	monthNames := []string{
		"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	month := monthNames[t.Month()]
	day := t.Day()
	year := t.Year()

	// 12-hour format with AM/PM
	hour := t.Hour()
	minute := t.Minute()
	ampm := "AM"
	if hour >= 12 {
		ampm = "PM"
		if hour > 12 {
			hour -= 12
		}
	}
	if hour == 0 {
		hour = 12
	}

	return month + " " + itoa(day) + ", " + itoa(year) + " at " + itoa(hour) + ":" + padZero(itoa(minute)) + " " + ampm
}

// itoa converts an integer to string (simple implementation)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	negative := i < 0
	if negative {
		i = -i
	}

	buf := make([]byte, 0, 12)
	for i > 0 {
		digit := byte('0' + (i % 10))
		buf = append([]byte{digit}, buf...)
		i /= 10
	}

	if negative {
		buf = append([]byte{'-'}, buf...)
	}

	return string(buf)
}

// padZero pads a single digit with leading zero
func padZero(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}
