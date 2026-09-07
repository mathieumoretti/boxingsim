package service

import (
	"math"

	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/logger"
)

// ProgressionService handles stat progression calculations with diminishing returns
// and fatigue-based effectiveness modifiers (MAT-22).
const (
	// DiminishingReturnsDivider controls the curve of stat diminishing returns.
	// Formula: multiplier = 1 / (1 + current_stat / DiminishingReturnsDivider)
	// At stat=0: multiplier = 1.0 (100% effectiveness)
	// At stat=50: multiplier = 0.67 (33% reduction)
	// At stat=100: multiplier = 0.50 (50% reduction)
	// At stat=200: multiplier = 0.33 (67% reduction)
	DiminishingReturnsDivider = 100.0

	// FatigueEffectivenessDivider controls how fatigue reduces training effectiveness.
	// Formula: multiplier = max(MinFatigueMultiplier, 1.0 - (fatigue_score / FatigueEffectivenessDivider))
	// At fatigue=0: multiplier = 1.0 (100% effectiveness)
	// At fatigue=50: multiplier = 0.75 (25% reduction)
	// At fatigue=100: multiplier = 0.50 (50% reduction - floor)
	FatigueEffectivenessDivider = 200.0

	// MinFatigueMultiplier is the minimum effectiveness multiplier due to fatigue.
	// Even at maximum fatigue, training yields 50% of normal gains.
	MinFatigueMultiplier = 0.5

	// XPBaseMultiplier converts effective stat gains into experience points.
	// Formula: xp_gained = (effective_str + effective_def + effective_agi) * duration_hours * XPBaseMultiplier
	XPBaseMultiplier = 10.0

	// LevelXPBase is the base multiplier for level threshold calculation.
	// Formula: level_threshold(level) = LevelXPBase * level^1.5
	// Level 1 → 2: requires 100 XP
	// Level 2 → 3: requires 260 XP (cumulative: 360)
	// Level 5 → 6: requires 707 XP (cumulative: ~2,400)
	// Level 10 → 11: requires 3,162 XP (cumulative: ~18,000)
	LevelXPBase = 100.0

	// LevelXPExponent controls how quickly level requirements scale.
	// Using 1.5 provides a smooth accelerating curve.
	LevelXPExponent = 1.5
)

// EffectiveGains represents the calculated stat gains after applying all modifiers.
type EffectiveGains struct {
	Strength  float64 `json:"strength"`
	Defense   float64 `json:"defense"`
	Agility   float64 `json:"agility"`
	XP        float64 `json:"experience"`
	LevelGain int     `json:"level_gain"`

	// Multipliers used in calculation (for debugging/logging)
	DiminishingReturnsStrength float64 `json:"diminishing_strength"`
	DiminishingReturnsDefense  float64 `json:"diminishing_defense"`
	DiminishingReturnsAgility  float64 `json:"diminishing_agility"`
	FatigueMultiplier          float64 `json:"fatigue_multiplier"`
}

// ProgressionService handles stat progression calculations.
type ProgressionService struct {
	logger *logger.Logger
}

// NewProgressionService creates a new ProgressionService instance.
func NewProgressionService(lg *logger.Logger) *ProgressionService {
	return &ProgressionService{
		logger: lg,
	}
}

// CalculateDiminishingReturns returns the multiplier based on current stat value.
// Higher stats progress slower to prevent runaway power scaling.
func (s *ProgressionService) CalculateDiminishingReturns(currentStat float64) float64 {
	return 1.0 / (1.0 + currentStat/DiminishingReturnsDivider)
}

// CalculateFatigueEffectiveness returns the effectiveness multiplier based on fatigue score.
// Returns a value between MinFatigueMultiplier (0.5) and 1.0.
func (s *ProgressionService) CalculateFatigueEffectiveness(fatigueScore float64) float64 {
	multiplier := 1.0 - (fatigueScore / FatigueEffectivenessDivider)
	return math.Max(MinFatigueMultiplier, multiplier)
}

// CalculateEffectiveGains computes final gains with all modifiers applied.
// This is the core function that combines base gains, diminishing returns, and fatigue effects.
func (s *ProgressionService) CalculateEffectiveGains(
	boxer *model.Boxer,
	baseStrengthGain float64,
	baseDefenseGain float64,
	baseAgilityGain float64,
	durationHours float64,
) EffectiveGains {
	// Calculate diminishing returns multiplier per stat
	dimStr := s.CalculateDiminishingReturns(boxer.Strength)
	dimDef := s.CalculateDiminishingReturns(boxer.Defense)
	dimAgi := s.CalculateDiminishingReturns(boxer.Agility)

	// Calculate fatigue effectiveness multiplier (0.5 - 1.0)
	fatigueMult := s.CalculateFatigueEffectiveness(boxer.FatigueScore)

	// Apply all modifiers to get effective gains
	effectiveStr := baseStrengthGain * dimStr * fatigueMult
	effectiveDef := baseDefenseGain * dimDef * fatigueMult
	effectiveAgi := baseAgilityGain * dimAgi * fatigueMult

	// Calculate XP gained from effective stat improvements
	totalEffectiveGains := effectiveStr + effectiveDef + effectiveAgi
	xpGained := totalEffectiveGains * durationHours * XPBaseMultiplier

	return EffectiveGains{
		Strength:                   effectiveStr,
		Defense:                    effectiveDef,
		Agility:                    effectiveAgi,
		XP:                         xpGained,
		DiminishingReturnsStrength: dimStr,
		DiminishingReturnsDefense:  dimDef,
		DiminishingReturnsAgility:  dimAgi,
		FatigueMultiplier:          fatigueMult,
		LevelGain:                  0, // Will be updated by ApplyLevelUp
	}
}

// CalculateLevelThreshold returns the XP required to reach the next level from current level.
// Formula: threshold = LevelXPBase * level^1.5
func (s *ProgressionService) CalculateLevelThreshold(currentLevel int) float64 {
	level := float64(currentLevel)
	return LevelXPBase * math.Pow(level, LevelXPExponent)
}

// ApplyLevelUp increments the boxer's level if they've earned enough XP.
// Returns the number of levels gained (0 or more).
func (s *ProgressionService) ApplyLevelUp(boxer *model.Boxer) int {
	if boxer.Level == 0 {
		boxer.Level = 1
	}

	levelsGained := 0
	for {
		nextLevelThreshold := s.CalculateLevelThreshold(boxer.Level)
		if boxer.Experience >= nextLevelThreshold {
			boxer.Level++
			levelsGained++
			s.logger.Info("Boxer ID=%d leveled up to level %d (XP: %.0f / %.0f)",
				boxer.ID, boxer.Level, boxer.Experience, nextLevelThreshold)
		} else {
			break
		}
	}

	return levelsGained
}
