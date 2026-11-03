package domain

import (
	"testing"
)

func TestBoxerID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   BoxerID
		want bool
	}{
		{"valid positive ID", BoxerID(1), true},
		{"valid large ID", BoxerID(9999), true},
		{"invalid zero ID", BoxerID(0), false},
		{"invalid negative ID", BoxerID(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("BoxerID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestFightID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   FightID
		want bool
	}{
		{"valid positive ID", FightID(1), true},
		{"valid large ID", FightID(9999), true},
		{"invalid zero ID", FightID(0), false},
		{"invalid negative ID", FightID(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("FightID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestUserID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   UserID
		want bool
	}{
		{"valid positive ID", UserID(1), true},
		{"valid large ID", UserID(9999), true},
		{"invalid zero ID", UserID(0), false},
		{"invalid negative ID", UserID(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("UserID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestTrainingSessionID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   TrainingSessionID
		want bool
	}{
		{"valid positive ID", TrainingSessionID(1), true},
		{"valid large ID", TrainingSessionID(9999), true},
		{"invalid zero ID", TrainingSessionID(0), false},
		{"invalid negative ID", TrainingSessionID(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("TrainingSessionID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestScheduledEventID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   ScheduledEventID
		want bool
	}{
		{"valid positive ID", ScheduledEventID(1), true},
		{"valid large ID", ScheduledEventID(9999), true},
		{"invalid zero ID", ScheduledEventID(0), false},
		{"invalid negative ID", ScheduledEventID(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("ScheduledEventID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestTrainingTypeID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   TrainingTypeID
		want bool
	}{
		{"valid positive ID", TrainingTypeID(1), true},
		{"valid large ID", TrainingTypeID(9999), true},
		{"invalid zero ID", TrainingTypeID(0), false},
		{"invalid negative ID", TrainingTypeID(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("TrainingTypeID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestInvalidConstants(t *testing.T) {
	tests := []struct {
		name string
		id   interface{ IsValid() bool }
	}{
		{"InvalidBoxerID", InvalidBoxerID},
		{"InvalidFightID", InvalidFightID},
		{"InvalidUserID", InvalidUserID},
		{"InvalidTrainingSessionID", InvalidTrainingSessionID},
		{"InvalidScheduledEventID", InvalidScheduledEventID},
		{"InvalidTrainingTypeID", InvalidTrainingTypeID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.id.IsValid() {
				t.Errorf("Expected invalid ID to return false for IsValid()")
			}
		})
	}
}
