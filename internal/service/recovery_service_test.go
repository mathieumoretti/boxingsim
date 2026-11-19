package service

import (
	"testing"
	"time"

	"github.com/mormm/boxing/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestRecoveryService_CalculateRecoveryDuration validates recovery duration tiers
func TestRecoveryService_CalculateRecoveryDuration(t *testing.T) {
	service := &RecoveryService{} // Pure calculation, no dependencies needed

	tests := []struct {
		name       string
		healthLost float64
		expected   int
		tier       string
	}{
		{
			name:       "Minor damage 5 HP",
			healthLost: 5,
			expected:   MinorInjuryRecoveryHours, // 24 hours
			tier:       "minor",
		},
		{
			name:       "Minor damage 20 HP (at threshold)",
			healthLost: 20,
			expected:   MinorInjuryRecoveryHours, // 24 hours
			tier:       "minor",
		},
		{
			name:       "Moderate damage 21 HP",
			healthLost: 21,
			expected:   ModerateInjuryRecoveryHours, // 48 hours
			tier:       "moderate",
		},
		{
			name:       "Moderate damage 40 HP (at threshold)",
			healthLost: 40,
			expected:   ModerateInjuryRecoveryHours, // 48 hours
			tier:       "moderate",
		},
		{
			name:       "Major damage 41 HP",
			healthLost: 41,
			expected:   MajorInjuryRecoveryHours, // 72 hours
			tier:       "major",
		},
		{
			name:       "Major damage 80 HP (severe)",
			healthLost: 80,
			expected:   MajorInjuryRecoveryHours, // 72 hours
			tier:       "major",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateRecoveryDuration(tt.healthLost)
			assert.Equal(t, tt.expected, result, "Duration should match %s tier (%d hours)", tt.tier, tt.expected)
		})
	}
}

// TestRecoveryService_IsOnForcedRecovery validates recovery status checks
func TestRecoveryService_IsOnForcedRecovery(t *testing.T) {
	service := &RecoveryService{} // Pure calculation, no dependencies needed

	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		boxer    *model.Boxer
		expected bool
	}{
		{
			name: "No forced recovery (nil)",
			boxer: &model.Boxer{
				ID:                  1,
				ForcedRecoveryUntil: nil,
			},
			expected: false,
		},
		{
			name: "Recovery ended in past",
			boxer: &model.Boxer{
				ID:                  2,
				ForcedRecoveryUntil: &past,
			},
			expected: false,
		},
		{
			name: "Recovery ongoing (future timestamp)",
			boxer: &model.Boxer{
				ID:                  3,
				ForcedRecoveryUntil: &future,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsOnForcedRecovery(tt.boxer, now)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRecoveryService_CheckCanFight validates fight eligibility checks
func TestRecoveryService_CheckCanFight(t *testing.T) {
	service := &RecoveryService{} // Pure calculation, no dependencies needed

	now := time.Now()
	futureRecovery := now.Add(24 * time.Hour)

	tests := []struct {
		name         string
		boxer        *model.Boxer
		expectedCan  bool
		expectReason bool
	}{
		{
			name: "Healthy boxer can fight",
			boxer: &model.Boxer{
				ID:     1,
				Health: 80.0,
			},
			expectedCan:  true,
			expectReason: false,
		},
		{
			name: "Low health boxer cannot fight (below 50%)",
			boxer: &model.Boxer{
				ID:     2,
				Health: 40.0,
			},
			expectedCan:  false,
			expectReason: true,
		},
		{
			name: "Boxer on forced recovery cannot fight",
			boxer: &model.Boxer{
				ID:                  3,
				Health:              80.0,
				ForcedRecoveryUntil: &futureRecovery,
			},
			expectedCan:  false,
			expectReason: true,
		},
		{
			name: "Borderline health (exactly minimum) can fight",
			boxer: &model.Boxer{
				ID:     4,
				Health: MinHealthToFight,
			},
			expectedCan:  true,
			expectReason: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canFight, reason := service.CheckCanFight(tt.boxer, now)
			assert.Equal(t, tt.expectedCan, canFight)
			if tt.expectReason {
				assert.NotEmpty(t, reason, "Should provide reason when cannot fight")
			} else {
				assert.Empty(t, reason)
			}
		})
	}
}
