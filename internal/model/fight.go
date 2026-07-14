package model

import "time"

// FightStatus represents the status of a fight
type FightStatus string

const (
	FightStatusScheduled  FightStatus = "scheduled"
	FightStatusInProgress FightStatus = "in_progress"
	FightStatusCompleted  FightStatus = "completed"
	FightStatusCancelled  FightStatus = "canceled"
)

// Fight represents a boxing match
type Fight struct {
	ID            int                    `json:"id"`
	Boxer1ID      int                    `json:"boxer1_id"`
	Boxer2ID      int                    `json:"boxer2_id"`
	Status        FightStatus            `json:"status"`
	ScheduledTime *time.Time             `json:"scheduled_time"`
	StartTime     *time.Time             `json:"start_time"`
	EndTime       *time.Time             `json:"end_time"`
	WinnerID      *int                   `json:"winner_id"`
	Round         int                    `json:"round"`
	Data          map[string]interface{} `json:"data"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// FightCreate represents a request to create a new fight
type FightCreate struct {
	Boxer1ID      int        `json:"boxer1_id" binding:"required"`
	Boxer2ID      int        `json:"boxer2_id" binding:"required"`
	ScheduledTime *time.Time `json:"scheduled_time" binding:"required"`
	Round         int        `json:"round" binding:"min=1"`
}

// FightUpdate represents a request to update a fight
type FightUpdate struct {
	Status    *FightStatus            `json:"status"`
	StartTime *time.Time              `json:"start_time"`
	EndTime   *time.Time              `json:"end_time"`
	WinnerID  *int                    `json:"winner_id"`
	Round     *int                    `json:"round"`
	Data      *map[string]interface{} `json:"data"`
}

// FightResponse represents a fight for API responses
type FightResponse struct {
	ID            int                    `json:"id"`
	Boxer1ID      int                    `json:"boxer1_id"`
	Boxer2ID      int                    `json:"boxer2_id"`
	Status        FightStatus            `json:"status"`
	ScheduledTime *time.Time             `json:"scheduled_time"`
	StartTime     *time.Time             `json:"start_time"`
	EndTime       *time.Time             `json:"end_time"`
	WinnerID      *int                   `json:"winner_id"`
	Round         int                    `json:"round"`
	Data          map[string]interface{} `json:"data"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
