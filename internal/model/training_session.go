package model

import "time"

// TrainingType represents a type of training available in the system (reference data)
type TrainingType struct {
	ID                 int       `db:"id" json:"id"`
	Name               string    `db:"name" json:"name"`
	Description        *string   `db:"description" json:"description"`
	StrengthGainFactor float64   `db:"strength_gain_factor" json:"strength_gain_factor"`
	DefenseGainFactor  float64   `db:"defense_gain_factor" json:"defense_gain_factor"`
	AgilityGainFactor  float64   `db:"agility_gain_factor" json:"agility_gain_factor"`
	EnergyCost         int       `db:"energy_cost" json:"energy_cost"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// TrainingSessionStatus represents the status of a training session
type TrainingSessionStatus string

const (
	TrainingSessionPending   TrainingSessionStatus = "pending"
	TrainingSessionCompleted TrainingSessionStatus = "completed"
	TrainingSessionCancelled TrainingSessionStatus = "cancelled"
)

// TrainingSession represents an individual training session for a boxer
type TrainingSession struct {
	ID                  int                 `db:"id" json:"id"`
	BoxerID             int                 `db:"boxer_id" json:"boxer_id"`
	TrainingTypeID      int                 `db:"training_type_id" json:"training_type_id"`
	ScheduledEventID    *int                `db:"scheduled_event_id" json:"scheduled_event_id,omitempty"`
	DurationHours       float64             `db:"duration_hours" json:"duration_hours"`
	PlannedStrengthGain float64             `db:"planned_strength_gain" json:"planned_strength_gain"`
	PlannedDefenseGain  float64             `db:"planned_defense_gain" json:"planned_defense_gain"`
	PlannedAgilityGain  float64             `db:"planned_agility_gain" json:"planned_agility_gain"`
	Status              TrainingSessionStatus `db:"status" json:"status"`
	CompletedAt         *time.Time          `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt           time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time           `db:"updated_at" json:"updated_at"`

	// Joined data (optional, for query results with JOIN)
	TrainingType *TrainingType `json:"training_type,omitempty"`
}

// TrainingSessionCreate represents a request to create a new training session
type TrainingSessionCreate struct {
	BoxerID          int     `json:"boxer_id" binding:"required,min=1"`
	TrainingTypeID   int     `json:"training_type_id" binding:"required,min=1"`
	ScheduledEventID *int    `json:"scheduled_event_id,omitempty"`
	DurationHours    float64 `json:"duration_hours" binding:"required,gt=0"`
}

// TrainingSessionUpdate represents a request to update an existing training session
type TrainingSessionUpdate struct {
	Status      *TrainingSessionStatus `json:"status,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}
