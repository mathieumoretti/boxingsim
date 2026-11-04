package domain

import (
	"testing"
)

// ==================== IsTerminalStatus Tests ====================

func TestIsTerminalStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"completed", "completed", true},
		{"failed", "failed", true},
		{"cancelled", "cancelled", true},
		{"stopped", "stopped", true},
		{"pending", "pending", false},
		{"processing", "processing", false},
		{"scheduled", "scheduled", false},
		{"in_progress", "in_progress", false},
		{"running", "running", false},
		{"paused", "paused", false},
		{"uppercase COMPLETED", "COMPLETED", true},
		{"uppercase FAILED", "FAILED", true},
		{"mixed case CoMpLeTeD", "CoMpLeTeD", true},
		{"empty string", "", false},
		{"unknown status", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTerminalStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsTerminalStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

// ==================== FightStatus Tests ====================

func TestFightStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    FightStatus
		want bool
	}{
		{"scheduled", FightStatusScheduled, true},
		{"in_progress", FightStatusInProgress, true},
		{"completed", FightStatusCompleted, true},
		{"canceled", FightStatusCancelled, true},
		{"empty", FightStatus(""), false},
		{"unknown", FightStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFightStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name string
		s    FightStatus
		want bool
	}{
		{"scheduled is not terminal", FightStatusScheduled, false},
		{"in_progress is not terminal", FightStatusInProgress, false},
		{"completed is terminal", FightStatusCompleted, true},
		{"canceled is terminal", FightStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFightStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current FightStatus
		target  Status
		want    bool
	}{
		// Valid transitions from scheduled
		{"scheduled -> in_progress", FightStatusScheduled, FightStatusInProgress, true},
		{"scheduled -> canceled", FightStatusScheduled, FightStatusCancelled, true},
		{"scheduled -> completed (skip)", FightStatusScheduled, FightStatusCompleted, false},

		// Valid transitions from in_progress
		{"in_progress -> completed", FightStatusInProgress, FightStatusCompleted, true},
		{"in_progress -> scheduled (back)", FightStatusInProgress, FightStatusScheduled, false},
		{"in_progress -> canceled", FightStatusInProgress, FightStatusCancelled, false},

		// Terminal states - no transitions allowed
		{"completed -> in_progress", FightStatusCompleted, FightStatusInProgress, false},
		{"canceled -> scheduled", FightStatusCancelled, FightStatusScheduled, false},

		// Self-transitions not allowed
		{"scheduled -> scheduled", FightStatusScheduled, FightStatusScheduled, false},

		// Wrong status type
		{"scheduled -> EventStatusPending", FightStatusScheduled, EventStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.current.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%v) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestFightStatus_Equals(t *testing.T) {
	tests := []struct {
		name  string
		s     FightStatus
		other Status
		want  bool
	}{
		{"same status", FightStatusScheduled, FightStatusScheduled, true},
		{"different status", FightStatusScheduled, FightStatusInProgress, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Equals(tt.other)
			if got != tt.want {
				t.Errorf("Equals(%v) = %v, want %v", tt.other, got, tt.want)
			}
		})
	}
}

func TestFightStatus_String(t *testing.T) {
	s := FightStatusScheduled
	got := s.String()
	want := "scheduled"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ==================== EventStatus Tests ====================

func TestEventStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    EventStatus
		want bool
	}{
		{"pending", EventStatusPending, true},
		{"processing", EventStatusProcessing, true},
		{"completed", EventStatusCompleted, true},
		{"failed", EventStatusFailed, true},
		{"empty", EventStatus(""), false},
		{"unknown", EventStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name string
		s    EventStatus
		want bool
	}{
		{"pending is not terminal", EventStatusPending, false},
		{"processing is not terminal", EventStatusProcessing, false},
		{"completed is terminal", EventStatusCompleted, true},
		{"failed is terminal", EventStatusFailed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current EventStatus
		target  Status
		want    bool
	}{
		// Valid transitions from pending
		{"pending -> processing", EventStatusPending, EventStatusProcessing, true},
		{"pending -> completed (idempotent)", EventStatusPending, EventStatusCompleted, true},
		{"pending -> failed", EventStatusPending, EventStatusFailed, false},

		// Valid transitions from processing
		{"processing -> completed", EventStatusProcessing, EventStatusCompleted, true},
		{"processing -> failed", EventStatusProcessing, EventStatusFailed, true},
		{"processing -> pending (back)", EventStatusProcessing, EventStatusPending, false},

		// Terminal states - no transitions allowed
		{"completed -> processing", EventStatusCompleted, EventStatusProcessing, false},
		{"failed -> pending", EventStatusFailed, EventStatusPending, false},

		// Self-transitions not allowed
		{"pending -> pending", EventStatusPending, EventStatusPending, false},

		// Wrong status type
		{"pending -> FightStatusScheduled", EventStatusPending, FightStatusScheduled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.current.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%v) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestEventStatus_Equals(t *testing.T) {
	tests := []struct {
		name  string
		s     EventStatus
		other Status
		want  bool
	}{
		{"same status", EventStatusPending, EventStatusPending, true},
		{"different status", EventStatusPending, EventStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Equals(tt.other)
			if got != tt.want {
				t.Errorf("Equals(%v) = %v, want %v", tt.other, got, tt.want)
			}
		})
	}
}

// ==================== TrainingSessionStatus Tests ====================

func TestTrainingSessionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    TrainingSessionStatus
		want bool
	}{
		{"pending", TrainingSessionPending, true},
		{"completed", TrainingSessionCompleted, true},
		{"cancelled", TrainingSessionCancelled, true},
		{"empty", TrainingSessionStatus(""), false},
		{"unknown", TrainingSessionStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrainingSessionStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name string
		s    TrainingSessionStatus
		want bool
	}{
		{"pending is not terminal", TrainingSessionPending, false},
		{"completed is terminal", TrainingSessionCompleted, true},
		{"cancelled is terminal", TrainingSessionCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrainingSessionStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current TrainingSessionStatus
		target  Status
		want    bool
	}{
		// Valid transitions from pending
		{"pending -> completed", TrainingSessionPending, TrainingSessionCompleted, true},
		{"pending -> cancelled", TrainingSessionPending, TrainingSessionCancelled, true},

		// Terminal states - no transitions allowed
		{"completed -> pending", TrainingSessionCompleted, TrainingSessionPending, false},
		{"cancelled -> completed", TrainingSessionCancelled, TrainingSessionCompleted, false},

		// Self-transitions not allowed
		{"pending -> pending", TrainingSessionPending, TrainingSessionPending, false},

		// Wrong status type
		{"pending -> EventStatusPending", TrainingSessionPending, EventStatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.current.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%v) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestTrainingSessionStatus_Equals(t *testing.T) {
	tests := []struct {
		name  string
		s     TrainingSessionStatus
		other Status
		want  bool
	}{
		{"same status", TrainingSessionPending, TrainingSessionPending, true},
		{"different status", TrainingSessionPending, TrainingSessionCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Equals(tt.other)
			if got != tt.want {
				t.Errorf("Equals(%v) = %v, want %v", tt.other, got, tt.want)
			}
		})
	}
}

// ==================== WorldClockStatus Tests ====================

func TestWorldClockStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    WorldClockStatus
		want bool
	}{
		{"running", WorldClockRunning, true},
		{"paused", WorldClockPaused, true},
		{"stopped", WorldClockStopped, true},
		{"empty", WorldClockStatus(""), false},
		{"unknown", WorldClockStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorldClockStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name string
		s    WorldClockStatus
		want bool
	}{
		{"running is not terminal", WorldClockRunning, false},
		{"paused is NOT terminal (can resume)", WorldClockPaused, false},
		{"stopped is terminal", WorldClockStopped, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsTerminal(); got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorldClockStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name    string
		current WorldClockStatus
		target  Status
		want    bool
	}{
		// Valid transitions from running
		{"running -> paused", WorldClockRunning, WorldClockPaused, true},
		{"running -> stopped", WorldClockRunning, WorldClockStopped, true},

		// Valid transitions from paused
		{"paused -> running (resume)", WorldClockPaused, WorldClockRunning, true},
		{"paused -> stopped", WorldClockPaused, WorldClockStopped, true},

		// Valid transitions from stopped
		{"stopped -> running (restart)", WorldClockStopped, WorldClockRunning, true},
		{"stopped -> paused", WorldClockStopped, WorldClockPaused, true},

		// Self-transitions not allowed
		{"running -> running", WorldClockRunning, WorldClockRunning, false},
		{"paused -> paused", WorldClockPaused, WorldClockPaused, false},

		// Wrong status type
		{"running -> FightStatusScheduled", WorldClockRunning, FightStatusScheduled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.current.CanTransitionTo(tt.target)
			if got != tt.want {
				t.Errorf("CanTransitionTo(%v) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestWorldClockStatus_Equals(t *testing.T) {
	tests := []struct {
		name  string
		s     WorldClockStatus
		other Status
		want  bool
	}{
		{"same status", WorldClockRunning, WorldClockRunning, true},
		{"different status", WorldClockRunning, WorldClockPaused, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Equals(tt.other)
			if got != tt.want {
				t.Errorf("Equals(%v) = %v, want %v", tt.other, got, tt.want)
			}
		})
	}
}

// ==================== TerminalStatuses Constant Tests ====================

func TestTerminalStatusesConstant(t *testing.T) {
	// Verify all expected terminal statuses are present
	expected := map[string]bool{
		"completed": true,
		"failed":    true,
		"canceled":  true,
		"cancelled": true,
		"stopped":   true,
	}

	for _, ts := range TerminalStatuses {
		if !expected[ts] {
			t.Errorf("TerminalStatuses contains unexpected status: %q", ts)
		}
		delete(expected, ts)
	}

	// Verify no expected statuses are missing
	if len(expected) > 0 {
		t.Errorf("TerminalStatuses missing statuses: %v", expected)
	}
}

// ==================== Integration Tests ====================

func TestStatusInterfaceImplementation(t *testing.T) {
	// Verify all status types implement the Status interface
	var _ Status = FightStatusScheduled
	var _ Status = EventStatusPending
	var _ Status = TrainingSessionPending
	var _ Status = WorldClockRunning

	t.Log("All status types implement Status interface")
}

func TestGenericStatusHandling(t *testing.T) {
	// Test that generic function can work with any Status type
	checkTerminal := func(s Status) string {
		if s.IsTerminal() {
			return "terminal"
		}
		return "active"
	}

	tests := []struct {
		name string
		s    Status
		want string
	}{
		{"fight completed", FightStatusCompleted, "terminal"},
		{"fight scheduled", FightStatusScheduled, "active"},
		{"event failed", EventStatusFailed, "terminal"},
		{"event pending", EventStatusPending, "active"},
		{"training cancelled", TrainingSessionCancelled, "terminal"},
		{"clock stopped", WorldClockStopped, "terminal"},
		{"clock running", WorldClockRunning, "active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkTerminal(tt.s)
			if got != tt.want {
				t.Errorf("checkTerminal(%v) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}
