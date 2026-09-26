package db

import (
	"database/sql"
	"errors"
	"time"
)

var ErrUpcomingFightNotFound = errors.New("no upcoming fight found for this boxer")

// UpcomingFightResponse represents the response for an upcoming fight query
type UpcomingFightResponse struct {
	FightID        int        `json:"fight_id"`
	OpponentID     int        `json:"opponent_id"`
	OpponentName   string     `json:"opponent_name"`
	OpponentNickname *string  `json:"opponent_nickname"`
	ScheduledTime  *time.Time `json:"scheduled_time"`
	Status         string     `json:"status"`
	Rounds         int        `json:"rounds"`
}

// GetUpcomingFightForBoxer retrieves the next scheduled or in-progress fight for a specific boxer
func GetUpcomingFightForBoxer(db *sql.DB, boxerID int) (*UpcomingFightResponse, error) {
	query := `
		SELECT
			f.id,
			CASE
				WHEN f.boxer1_id = $1 THEN f.boxer2_id
				ELSE f.boxer1_id
			END AS opponent_id,
			CASE
				WHEN f.boxer1_id = $1 THEN b2.name
				ELSE b1.name
			END AS opponent_name,
			CASE
				WHEN f.boxer1_id = $1 THEN b2.nickname
				ELSE b1.nickname
			END AS opponent_nickname,
			f.scheduled_time,
			f.status,
			COALESCE(f.round, 12) AS rounds
		FROM fights f
		LEFT JOIN boxers b1 ON f.boxer1_id = b1.id
		LEFT JOIN boxers b2 ON f.boxer2_id = b2.id
		WHERE (f.boxer1_id = $1 OR f.boxer2_id = $1)
		  AND f.status IN ('scheduled', 'in_progress')
		ORDER BY f.scheduled_time ASC
		LIMIT 1
	`

	var response UpcomingFightResponse
	err := db.QueryRow(query, boxerID).Scan(
		&response.FightID,
		&response.OpponentID,
		&response.OpponentName,
		&response.OpponentNickname,
		&response.ScheduledTime,
		&response.Status,
		&response.Rounds,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUpcomingFightNotFound
		}
		return nil, err
	}

	return &response, nil
}
