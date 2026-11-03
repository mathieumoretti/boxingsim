package domain

// BoxerID represents a boxer entity identifier with type safety.
// Database IDs start at 1, so 0 is invalid.
type BoxerID int

// IsValid returns true if the ID is a valid database identifier (> 0).
func (b BoxerID) IsValid() bool {
	return b > 0
}

// FightID represents a fight entity identifier with type safety.
type FightID int

// IsValid returns true if the ID is a valid database identifier (> 0).
func (f FightID) IsValid() bool {
	return f > 0
}

// UserID represents a user entity identifier with type safety.
type UserID int

// IsValid returns true if the ID is a valid database identifier (> 0).
func (u UserID) IsValid() bool {
	return u > 0
}

// TrainingSessionID represents a training session entity identifier with type safety.
type TrainingSessionID int

// IsValid returns true if the ID is a valid database identifier (> 0).
func (t TrainingSessionID) IsValid() bool {
	return t > 0
}

// ScheduledEventID represents a scheduled event entity identifier with type safety.
type ScheduledEventID int

// IsValid returns true if the ID is a valid database identifier (> 0).
func (s ScheduledEventID) IsValid() bool {
	return s > 0
}

// TrainingTypeID represents a training type reference data identifier with type safety.
type TrainingTypeID int

// IsValid returns true if the ID is a valid database identifier (> 0).
func (t TrainingTypeID) IsValid() bool {
	return t > 0
}

// Invalid* constants represent invalid zero-value IDs for error handling and tests.
var (
	InvalidBoxerID           = BoxerID(0)
	InvalidFightID           = FightID(0)
	InvalidUserID            = UserID(0)
	InvalidTrainingSessionID = TrainingSessionID(0)
	InvalidScheduledEventID  = ScheduledEventID(0)
	InvalidTrainingTypeID    = TrainingTypeID(0)
)
