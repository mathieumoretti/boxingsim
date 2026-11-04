package domain

import (
	"slices"
	"strings"
)

// Status defines a common interface for all status types in the boxing simulation.
// It provides shared functionality like string representation, terminal state detection,
// and transition validation.
//
// Usage:
//
//	// Check if any status is terminal
//	func processStatus(s Status) {
//	    if s.IsTerminal() {
//	        // Handle completion/failure/cancellation
//	    }
//	}
//
//	// Validate transitions
//	if current.CanTransitionTo(next) {
//	    updateStatus(next)
//	}
type Status interface {
	// String returns the string representation of the status.
	String() string

	// IsTerminal returns true if this is a final state with no further transitions.
	// Terminal states include: completed, failed, canceled, cancelled, stopped.
	IsTerminal() bool

	// CanTransitionTo returns true if a transition to the target status is valid.
	// Returns false for invalid transitions (e.g., completed -> pending).
	CanTransitionTo(target Status) bool

	// Equals returns true if this status equals another (case-insensitive comparison).
	Equals(other Status) bool
}

// TerminalStatuses is a slice of all terminal status strings across the domain.
// Used for generic status checking when type information is unavailable.
// Includes both spellings ("canceled"/"cancelled") for backward compatibility with legacy data.
var TerminalStatuses = []string{
	"completed",
	"failed",
	"canceled",
	"cancelled",
	"stopped",
}

// IsTerminalStatus checks if a string status value represents a terminal state.
// This helper works with raw strings when type information is not available.
//
// Examples:
//
//	IsTerminalStatus("completed") // true
//	IsTerminalStatus("pending")   // false
//	IsTerminalStatus("COMPLETED") // true (case-insensitive)
func IsTerminalStatus(status string) bool {
	return slices.Contains(TerminalStatuses, strings.ToLower(status))
}

// FightStatus implements Status for fight lifecycle management.
// Valid states: scheduled -> in_progress -> completed/canceled
type FightStatus string

const (
	FightStatusScheduled  FightStatus = "scheduled"
	FightStatusInProgress FightStatus = "in_progress"
	FightStatusCompleted  FightStatus = "completed"
	FightStatusCancelled  FightStatus = "cancelled"
)

// String returns the string representation.
func (s FightStatus) String() string {
	return string(s)
}

// IsTerminal returns true for completed or canceled fights.
func (s FightStatus) IsTerminal() bool {
	return s == FightStatusCompleted || s == FightStatusCancelled
}

// CanTransitionTo validates fight status transitions.
// Valid transitions:
//   - scheduled -> in_progress, canceled
//   - in_progress -> completed
//   - completed, canceled -> (none, terminal)
func (s FightStatus) CanTransitionTo(target Status) bool {
	t, ok := target.(FightStatus)
	if !ok {
		return false // Cannot transition to different status type
	}

	switch s {
	case FightStatusScheduled:
		return t == FightStatusInProgress || t == FightStatusCancelled
	case FightStatusInProgress:
		return t == FightStatusCompleted
	case FightStatusCompleted, FightStatusCancelled:
		return false // Terminal states
	default:
		return false // Unknown status
	}
}

// Equals compares two FightStatus values (case-insensitive).
func (s FightStatus) Equals(other Status) bool {
	if o, ok := other.(FightStatus); ok {
		return s == o
	}
	return string(s) == strings.ToLower(other.String())
}

// IsValid returns true if this is a recognized fight status.
func (s FightStatus) IsValid() bool {
	switch s {
	case FightStatusScheduled, FightStatusInProgress, FightStatusCompleted, FightStatusCancelled:
		return true
	default:
		return false
	}
}

// EventStatus implements Status for scheduled event lifecycle management.
// Valid states: pending -> processing -> completed/failed
type EventStatus string

const (
	EventStatusPending    EventStatus = "pending"
	EventStatusProcessing EventStatus = "processing"
	EventStatusCompleted  EventStatus = "completed"
	EventStatusFailed     EventStatus = "failed"
)

// String returns the string representation.
func (s EventStatus) String() string {
	return string(s)
}

// IsTerminal returns true for completed or failed events.
func (s EventStatus) IsTerminal() bool {
	return s == EventStatusCompleted || s == EventStatusFailed
}

// CanTransitionTo validates event status transitions.
// Valid transitions:
//   - pending -> processing, completed (idempotent direct completion)
//   - processing -> completed, failed
//   - completed, failed -> (none, terminal)
func (s EventStatus) CanTransitionTo(target Status) bool {
	t, ok := target.(EventStatus)
	if !ok {
		return false // Cannot transition to different status type
	}

	switch s {
	case EventStatusPending:
		// Allow direct completion for idempotent processing
		// (worker may complete event in single step without explicit "processing" state)
		return t == EventStatusProcessing || t == EventStatusCompleted
	case EventStatusProcessing:
		return t == EventStatusCompleted || t == EventStatusFailed
	case EventStatusCompleted, EventStatusFailed:
		return false // Terminal states
	default:
		return false // Unknown status
	}
}

// Equals compares two EventStatus values (case-insensitive).
func (s EventStatus) Equals(other Status) bool {
	if o, ok := other.(EventStatus); ok {
		return s == o
	}
	return string(s) == strings.ToLower(other.String())
}

// IsValid returns true if this is a recognized event status.
func (s EventStatus) IsValid() bool {
	switch s {
	case EventStatusPending, EventStatusProcessing, EventStatusCompleted, EventStatusFailed:
		return true
	default:
		return false
	}
}

// TrainingSessionStatus implements Status for training session lifecycle management.
// Valid states: pending -> completed/cancelled
type TrainingSessionStatus string

const (
	TrainingSessionPending   TrainingSessionStatus = "pending"
	TrainingSessionCompleted TrainingSessionStatus = "completed"
	TrainingSessionCancelled TrainingSessionStatus = "cancelled"
)

// String returns the string representation.
func (s TrainingSessionStatus) String() string {
	return string(s)
}

// IsTerminal returns true for completed or cancelled sessions.
func (s TrainingSessionStatus) IsTerminal() bool {
	return s == TrainingSessionCompleted || s == TrainingSessionCancelled
}

// CanTransitionTo validates training session status transitions.
// Valid transitions:
//   - pending -> completed, cancelled
//   - completed, cancelled -> (none, terminal)
func (s TrainingSessionStatus) CanTransitionTo(target Status) bool {
	t, ok := target.(TrainingSessionStatus)
	if !ok {
		return false // Cannot transition to different status type
	}

	switch s {
	case TrainingSessionPending:
		return t == TrainingSessionCompleted || t == TrainingSessionCancelled
	case TrainingSessionCompleted, TrainingSessionCancelled:
		return false // Terminal states
	default:
		return false // Unknown status
	}
}

// Equals compares two TrainingSessionStatus values (case-insensitive).
func (s TrainingSessionStatus) Equals(other Status) bool {
	if o, ok := other.(TrainingSessionStatus); ok {
		return s == o
	}
	return string(s) == strings.ToLower(other.String())
}

// IsValid returns true if this is a recognized training session status.
func (s TrainingSessionStatus) IsValid() bool {
	switch s {
	case TrainingSessionPending, TrainingSessionCompleted, TrainingSessionCancelled:
		return true
	default:
		return false
	}
}

// WorldClockStatus implements Status for the global simulation clock.
// Valid states: running <-> paused, running/paused -> stopped
type WorldClockStatus string

const (
	WorldClockRunning WorldClockStatus = "running"
	WorldClockPaused  WorldClockStatus = "paused"
	WorldClockStopped WorldClockStatus = "stopped"
)

// String returns the string representation.
func (s WorldClockStatus) String() string {
	return string(s)
}

// IsTerminal returns true only for stopped clock.
// Note: Paused is NOT terminal - it can be resumed.
func (s WorldClockStatus) IsTerminal() bool {
	return s == WorldClockStopped
}

// CanTransitionTo validates world clock status transitions.
// Valid transitions:
//   - running -> paused, stopped
//   - paused -> running, stopped
//   - stopped -> running (restart), paused (resume from stopped state)
func (s WorldClockStatus) CanTransitionTo(target Status) bool {
	t, ok := target.(WorldClockStatus)
	if !ok {
		return false // Cannot transition to different status type
	}

	switch s {
	case WorldClockRunning:
		return t == WorldClockPaused || t == WorldClockStopped
	case WorldClockPaused:
		return t == WorldClockRunning || t == WorldClockStopped
	case WorldClockStopped:
		// From stopped, can restart directly or go to paused state
		return t == WorldClockRunning || t == WorldClockPaused
	default:
		return false // Unknown status
	}
}

// Equals compares two WorldClockStatus values (case-insensitive).
func (s WorldClockStatus) Equals(other Status) bool {
	if o, ok := other.(WorldClockStatus); ok {
		return s == o
	}
	return string(s) == strings.ToLower(other.String())
}

// IsValid returns true if this is a recognized world clock status.
func (s WorldClockStatus) IsValid() bool {
	switch s {
	case WorldClockRunning, WorldClockPaused, WorldClockStopped:
		return true
	default:
		return false
	}
}
