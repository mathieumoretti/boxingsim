package model

import "time"

// SessionType represents the type of training session
type SessionType string

const (
	SessionTypeStrength SessionType = "strength"
	SessionTypeDefense  SessionType = "defense"
	SessionTypeAgility  SessionType = "agility"
	SessionTypeMixed    SessionType = "mixed"
	SessionTypeOther    SessionType = "other"
)

// TrainingSession represents a training session
type TrainingSession struct {
	ID                 int            `json:"id"`
	BoxerID            int            `json:"boxer_id"`
	SessionType        SessionType    `json:"session_type"`
	DurationMinutes    int            `json:"duration_minutes"`
	StrengthGain       float64        `json:"strength_gain"`
	DefenseGain        float64        `json:"defense_gain"`
	AgilityGain        float64        `json:"agility_gain"`
	ExperienceGain     int            `json:"experience_gain"`
	CreatedAt          time.Time      `json:"created_at"`
}

// TrainingSessionCreate represents a request to create a new training session
type TrainingSessionCreate struct {
	BoxerID            int            `json:"boxer_id" binding:"required"`
	SessionType        SessionType    `json:"session_type" binding:"required"`
	DurationMinutes    int            `json:"duration_minutes" binding:"required,min=1,max=180"`
	StrengthGain       *float64       `json:"strength_gain"`
	DefenseGain        *float64       `json:"defense_gain"`
	AgilityGain        *float64       `json:"agility_gain"`
	ExperienceGain     *int           `json:"experience_gain"`
}

// TrainingSessionUpdate represents a request to update a training session
type TrainingSessionUpdate struct {
	SessionType        *SessionType    `json:"session_type"`
	DurationMinutes    *int            `json:"duration_minutes"`
	StrengthGain       *float64        `json:"strength_gain"`
	DefenseGain        *float64        `json:"defense_gain"`
	AgilityGain        *float64        `json:"agility_gain"`
	ExperienceGain     *int            `json:"experience_gain"`
}

// TrainingSessionResponse represents a training session for API responses
type TrainingSessionResponse struct {
	ID                 int            `json:"id"`
	BoxerID            int            `json:"boxer_id"`
	SessionType        SessionType    `json:"session_type"`
	DurationMinutes    int            `json:"duration_minutes"`
	StrengthGain       float64        `json:"strength_gain"`
	DefenseGain        float64        `json:"defense_gain"`
	AgilityGain        float64        `json:"agility_gain"`
	ExperienceGain     int            `json:"experience_gain"`
	CreatedAt          time.Time      `json:"created_at"`
}