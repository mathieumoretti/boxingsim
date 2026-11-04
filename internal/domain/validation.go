package domain

import "fmt"

// Validation constants shared across the domain layer.
const (
	MinStatValue   = 0.0
	MaxStatValue   = 100.0
	MinDurationHrs = 0.1 // 6 minutes minimum
	MaxDurationHrs = 24.0
	MinNameLength  = 1
	MaxNameLength  = 100
)

// ValidateBoxerStats validates strength, defense, and agility are within acceptable ranges.
func ValidateBoxerStats(strength, defense, agility float64) error {
	if err := validateStat("strength", strength); err != nil {
		return err
	}
	if err := validateStat("defense", defense); err != nil {
		return err
	}
	if err := validateStat("agility", agility); err != nil {
		return err
	}
	return nil
}

func validateStat(name string, value float64) error {
	if value < MinStatValue || value > MaxStatValue {
		return fmt.Errorf("%s must be between %.0f and %.0f", name, MinStatValue, MaxStatValue)
	}
	return nil
}

// ValidateTrainingDuration ensures the training duration is within acceptable bounds.
func ValidateTrainingDuration(hours float64) error {
	if hours < MinDurationHrs || hours > MaxDurationHrs {
		return fmt.Errorf("duration must be between %.1f and %.1f hours", MinDurationHrs, MaxDurationHrs)
	}
	return nil
}

// ValidateName checks that a name string is within length constraints.
func ValidateName(name string) error {
	if len(name) < MinNameLength || len(name) > MaxNameLength {
		return fmt.Errorf("name must be between %d and %d characters", MinNameLength, MaxNameLength)
	}
	return nil
}
