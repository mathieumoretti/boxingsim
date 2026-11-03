package domain

import (
	"testing"
)

func TestValidateBoxerStats(t *testing.T) {
	tests := []struct {
		name     string
		strength float64
		defense  float64
		agility  float64
		wantErr  bool
	}{
		{"valid stats all 50", 50, 50, 50, false},
		{"valid stats at min", 0, 0, 0, false},
		{"valid stats at max", 100, 100, 100, false},
		{"valid mixed stats", 25, 75, 50, false},
		{"invalid strength negative", -1, 50, 50, true},
		{"invalid strength over max", 101, 50, 50, true},
		{"invalid defense negative", 50, -1, 50, true},
		{"invalid defense over max", 50, 101, 50, true},
		{"invalid agility negative", 50, 50, -1, true},
		{"invalid agility over max", 50, 50, 101, true},
		{"multiple invalid stats", -1, 101, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoxerStats(tt.strength, tt.defense, tt.agility)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBoxerStats(%v, %v, %v) error = %v, wantErr %v",
					tt.strength, tt.defense, tt.agility, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTrainingDuration(t *testing.T) {
	tests := []struct {
		name    string
		hours   float64
		wantErr bool
	}{
		{"valid minimum duration", 0.1, false},
		{"valid maximum duration", 24.0, false},
		{"valid one hour", 1.0, false},
		{"valid eight hours", 8.0, false},
		{"invalid below minimum", 0.05, true},
		{"invalid zero duration", 0.0, true},
		{"invalid negative duration", -1.0, true},
		{"invalid above maximum", 24.1, true},
		{"invalid very large duration", 100.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTrainingDuration(tt.hours)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTrainingDuration(%v) error = %v, wantErr %v",
					tt.hours, err, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid single char", "A", false},
		{"valid typical name", "John Smith", false},
		{"valid max length", func() string {
			s := ""
			for i := 0; i < 100; i++ {
				s += "a"
			}
			return s
		}(), false},
		{"invalid empty string", "", true},
		{"invalid over max length", func() string {
			s := ""
			for i := 0; i < 101; i++ {
				s += "a"
			}
			return s
		}(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidationConstants(t *testing.T) {
	// Verify constants have expected values
	if MinStatValue != 0.0 {
		t.Errorf("MinStatValue = %v, want 0.0", MinStatValue)
	}
	if MaxStatValue != 100.0 {
		t.Errorf("MaxStatValue = %v, want 100.0", MaxStatValue)
	}
	if MinDurationHrs != 0.1 {
		t.Errorf("MinDurationHrs = %v, want 0.1", MinDurationHrs)
	}
	if MaxDurationHrs != 24.0 {
		t.Errorf("MaxDurationHrs = %v, want 24.0", MaxDurationHrs)
	}
	if MinNameLength != 1 {
		t.Errorf("MinNameLength = %v, want 1", MinNameLength)
	}
	if MaxNameLength != 100 {
		t.Errorf("MaxNameLength = %v, want 100", MaxNameLength)
	}
}
