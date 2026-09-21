package seeding

// BoxerStats represents the generated stats for a boxer.
type BoxerStats struct {
	Strength float64 // Physical power and knockout ability
	Defense  float64 // Ability to absorb and block punches
	Agility  float64 // Speed, footwork, and reaction time
	Health   float64 // Total hit points / durability
	Energy   float64 // Stamina pool for training and fights
}

// LevelDistribution determines the probability weight for each level tier.
// More low-level boxers, fewer high-level ones (pyramid distribution).
func getLevelTier() int {
	randVal := randFloat64()

	if randVal < 0.50 {
		return 1 // Levels 1-10: 50% of population
	} else if randVal < 0.72 {
		return 2 // Levels 11-20: 22% of population
	} else if randVal < 0.86 {
		return 3 // Levels 21-30: 14% of population
	} else if randVal < 0.95 {
		return 4 // Levels 31-40: 9% of population
	} else {
		return 5 // Levels 41-50: 5% of population
	}
}

// GenerateRandomLevel generates a level based on pyramid distribution.
func GenerateRandomLevel() int {
	tier := getLevelTier()

	switch tier {
	case 1:
		return 1 + randIntn(10) // Levels 1-10
	case 2:
		return 11 + randIntn(10) // Levels 11-20
	case 3:
		return 21 + randIntn(10) // Levels 21-30
	case 4:
		return 31 + randIntn(10) // Levels 31-40
	default:
		return 41 + randIntn(10) // Levels 41-50
	}
}

// GetLevelProbability returns the probability weight for a specific level (for testing).
func GetLevelProbability(level int) float64 {
	if level < 1 || level > 50 {
		return 0
	}

	switch {
	case level >= 1 && level <= 10:
		return 0.05 // 5% per level in tier 1 (50% total / 10 levels)
	case level >= 11 && level <= 20:
		return 0.022 // 2.2% per level in tier 2 (22% total / 10 levels)
	case level >= 21 && level <= 30:
		return 0.014 // 1.4% per level in tier 3 (14% total / 10 levels)
	case level >= 31 && level <= 40:
		return 0.009 // 0.9% per level in tier 4 (9% total / 10 levels)
	default:
		return 0.005 // 0.5% per level in tier 5 (5% total / 10 levels)
	}
}

// GenerateBoxerStats creates realistic stats based on level and archetype.
func GenerateBoxerStats(level int, archetype FighterArchetype) BoxerStats {
	template := archetype.StatDistribution

	// Stat base ranges for minimum level (L1): 20-40 range
	const (
		baseStrengthMin = 20.0
		baseStrengthMax = 40.0
		baseDefenseMin  = 20.0
		baseDefenseMax  = 40.0
		baseAgilityMin  = 20.0
		baseAgilityMax  = 40.0
	)

	strength := CalculateStat(level, template.StrengthRange, baseStrengthMin, baseStrengthMax)
	defense := CalculateStat(level, template.DefenseRange, baseDefenseMin, baseDefenseMax)
	agility := CalculateStat(level, template.AgilityRange, baseAgilityMin, baseAgilityMax)
	health := CalculateHealth(level, template.HealthRange)
	energy := CalculateEnergy(level, template.EnergyRange)

	return BoxerStats{
		Strength: strength,
		Defense:  defense,
		Agility:  agility,
		Health:   health,
		Energy:   energy,
	}
}

// GenerateExperience calculates experience points based on level.
// Level 1 = 0 XP, each level requires more XP.
func GenerateExperience(level int) float64 {
	if level <= 1 {
		return 0
	}

	// XP formula: sum of (base + multiplier * previous_level) for all levels
	// Simplified: quadratic growth
	baseXP := 100.0
	multiplier := 1.5

	xp := baseXP * float64(level-1) * multiplier
	// Add some variance (±5%)
	variance := randFloat64()*0.1 - 0.05
	xp = xp * (1 + variance)

	return roundToOneDecimal(xp)
}

// GenerateCareerRecord creates wins/losses/draws/knockouts based on level and archetype.
func GenerateCareerRecord(level int, archetype FighterArchetype, stats BoxerStats) CareerRecord {
	if level <= 2 {
		// New boxers have no record yet
		return CareerRecord{
			Wins:   0,
			Losses: 0,
			Draws:  0,
			KOs:    0,
		}
	}

	// Number of fights roughly correlates with level
	// L3 boxer might have 1-5 fights, L50 boxer might have 80-150+ fights
	baseFights := (level - 2) * randIntn(4+2) // Multiply by 2-4 for variance
	if baseFights < 1 {
		baseFights = 1
	}

	// Win rate based on archetype and stats
	var baseWinRate float64
	switch archetype.Name {
	case "Brawler":
		baseWinRate = 0.55 + (stats.Strength / 200) // High strength = higher win rate
	case "Tank":
		baseWinRate = 0.60 + (stats.Defense / 200)
	case "Technician":
		baseWinRate = 0.65 + (stats.Agility / 200)
	case "Glass Cannon":
		baseWinRate = 0.50 + (stats.Strength / 180) // Volatile - high power but risky
	case "Counterpuncher":
		baseWinRate = 0.62 + (stats.Defense / 180)
	default:
		baseWinRate = 0.55 // All-Rounder, etc.
	}

	// Clamp win rate
	if baseWinRate > 0.85 {
		baseWinRate = 0.85
	}
	if baseWinRate < 0.35 {
		baseWinRate = 0.35
	}

	// Calculate wins with variance
	wins := int(float64(baseFights) * baseWinRate)
	draws := randIntn(baseFights/10 + 1) // Rare draws
	losses := baseFights - wins - draws

	if losses < 0 {
		losses = 0
		wins = baseFights - draws
	}

	// Knockouts based on strength and archetype
	var koRate float64
	switch archetype.Name {
	case "Brawler", "Glass Cannon":
		koRate = 0.5 + (stats.Strength / 300) // High KO rate
	case "Warrior", "Pressure Fighter":
		koRate = 0.35 + (stats.Strength / 250)
	default:
		koRate = 0.2 + (stats.Strength / 400) // Lower KO rate for technical boxers
	}

	if koRate > 0.8 {
		koRate = 0.8
	}

	knockouts := int(float64(wins) * koRate)
	if knockouts > wins {
		knockouts = wins // Can't have more KOs than wins
	}

	return CareerRecord{
		Wins:   wins,
		Losses: losses,
		Draws:  draws,
		KOs:    knockouts,
	}
}

// CareerRecord represents a boxer's fight record.
type CareerRecord struct {
	Wins   int // Total wins
	Losses int // Total losses
	Draws  int // Total draws
	KOs    int // Knockout victories
}

// GenerateMapPosition creates a valid map position within bounds.
// Map bounds: X: 0-100, Y: 0-73.33 (based on boxing gym image dimensions)
func GenerateMapPosition() (float64, float64) {
	// Safe bounds with margin from edges
	const (
		minX = 2.0
		maxX = 98.0
		minY = 2.0
		maxY = 71.33
	)

	x := minX + randFloat64()*(maxX-minX)
	y := minY + randFloat64()*(maxY-minY)

	return roundToOneDecimal(x), roundToOneDecimal(y)
}
