package seeding

import (
	"math"
)

// FighterArchetype defines a fighting style with stat distribution templates.
type FighterArchetype struct {
	Name             string       // Human-readable name (e.g., "Brawler", "Technician")
	Description      string       // Detailed description of the fighting style
	StatDistribution StatTemplate // Relative weights for stats
}

// StatTemplate defines the min/max multipliers for each stat based on archetype.
type StatTemplate struct {
	StrengthRange [2]float64 // Min/max multiplier (e.g., [0.8, 1.2])
	DefenseRange  [2]float64
	AgilityRange  [2]float64
	HealthRange   [2]float64
	EnergyRange   [2]float64
}

// Archetypes defines all fighter archetypes available in the system.
var Archetypes = []FighterArchetype{
	{
		Name:        "Brawler",
		Description: "Power-focused fighter with devastating strength but lower agility. Relies on knockout power.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.9, 1.3},
			DefenseRange:  [2]float64{0.7, 0.95},
			AgilityRange:  [2]float64{0.6, 0.85},
			HealthRange:   [2]float64{0.85, 1.1},
			EnergyRange:   [2]float64{0.8, 1.0},
		},
	},
	{
		Name:        "Technician",
		Description: "Highly skilled boxer with excellent footwork and precision. Balanced fighter with agility focus.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.75, 0.95},
			DefenseRange:  [2]float64{0.85, 1.05},
			AgilityRange:  [2]float64{1.0, 1.3},
			HealthRange:   [2]float64{0.8, 1.0},
			EnergyRange:   [2]float64{0.9, 1.15},
		},
	},
	{
		Name:        "Tank",
		Description: "Defensive specialist with high durability and stamina. Hard to knock down but slow.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.8, 1.0},
			DefenseRange:  [2]float64{1.0, 1.35},
			AgilityRange:  [2]float64{0.6, 0.8},
			HealthRange:   [2]float64{1.1, 1.3},
			EnergyRange:   [2]float64{1.05, 1.25},
		},
	},
	{
		Name:        "Glass Cannon",
		Description: "Explosive fighter with incredible power and speed but poor defense. High risk, high reward.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{1.0, 1.35},
			DefenseRange:  [2]float64{0.5, 0.75},
			AgilityRange:  [2]float64{0.95, 1.25},
			HealthRange:   [2]float64{0.7, 0.85},
			EnergyRange:   [2]float64{0.7, 0.9},
		},
	},
	{
		Name:        "All-Rounder",
		Description: "Balanced fighter with no major weaknesses. Reliable across all stats.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.85, 1.05},
			DefenseRange:  [2]float64{0.85, 1.05},
			AgilityRange:  [2]float64{0.85, 1.05},
			HealthRange:   [2]float64{0.9, 1.1},
			EnergyRange:   [2]float64{0.9, 1.1},
		},
	},
	{
		Name:        "Counterpuncher",
		Description: "Patient fighter who waits for openings. Excellent defense and timing with moderate power.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.7, 0.9},
			DefenseRange:  [2]float64{0.95, 1.2},
			AgilityRange:  [2]float64{0.85, 1.1},
			HealthRange:   [2]float64{0.8, 1.0},
			EnergyRange:   [2]float64{0.95, 1.2},
		},
	},
	{
		Name:        "Speedster",
		Description: "Lightning-fast fighter who overwhelms opponents with speed and combinations.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.65, 0.85},
			DefenseRange:  [2]float64{0.75, 0.95},
			AgilityRange:  [2]float64{1.1, 1.35},
			HealthRange:   [2]float64{0.75, 0.9},
			EnergyRange:   [2]float64{0.85, 1.05},
		},
	},
	{
		Name:        "Warrior",
		Description: "Balanced fighter with above-average strength and health. Fights with heart and determination.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.9, 1.15},
			DefenseRange:  [2]float64{0.85, 1.05},
			AgilityRange:  [2]float64{0.75, 0.95},
			HealthRange:   [2]float64{1.0, 1.2},
			EnergyRange:   [2]float64{0.9, 1.1},
		},
	},
	{
		Name:        "Defensive Specialist",
		Description: "Master of defense and movement. Tiring opponents while looking for openings.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.6, 0.8},
			DefenseRange:  [2]float64{1.1, 1.35},
			AgilityRange:  [2]float64{0.95, 1.15},
			HealthRange:   [2]float64{0.85, 1.0},
			EnergyRange:   [2]float64{1.0, 1.2},
		},
	},
	{
		Name:        "Pressure Fighter",
		Description: "Constant forward pressure with strong chin and good power. Relentless aggressor.",
		StatDistribution: StatTemplate{
			StrengthRange: [2]float64{0.85, 1.1},
			DefenseRange:  [2]float64{0.85, 1.05},
			AgilityRange:  [2]float64{0.7, 0.9},
			HealthRange:   [2]float64{0.95, 1.15},
			EnergyRange:   [2]float64{1.0, 1.2},
		},
	},
}

// GetArchetypeByName returns an archetype by its name.
func GetArchetypeByName(name string) *FighterArchetype {
	for i := range Archetypes {
		if Archetypes[i].Name == name {
			return &Archetypes[i]
		}
	}
	return nil
}

// RandomArchetype returns a random archetype with weighted distribution.
// Common archetypes (All-Rounder, Brawler, Technician) are more frequent.
func RandomArchetype() FighterArchetype {
	// Weighted probabilities for realistic distribution
	weights := []float64{
		0.15, // Brawler - common
		0.15, // Technician - common
		0.10, // Tank - moderately common
		0.05, // Glass Cannon - rare
		0.20, // All-Rounder - very common
		0.10, // Counterpuncher - moderately common
		0.08, // Speedster - less common
		0.10, // Warrior - moderately common
		0.05, // Defensive Specialist - rare
		0.02, // Pressure Fighter - very rare
	}

	randVal := randFloat64()
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if randVal < cumulative {
			return Archetypes[i]
		}
	}

	return Archetypes[4] // Default to All-Rounder if something goes wrong
}

// CalculateStat calculates a stat value based on level, archetype, and base ranges.
// Level 1 boxers start at 20-40 range, max level (50) capped at 80-95.
func CalculateStat(level int, templateRange [2]float64, baseMin float64, baseMax float64) float64 {
	if level < 1 {
		level = 1
	}

	// Determine the stat range based on level
	// Level progression: linear scaling from L1 to L50
	levelProgress := math.Max(0, math.Min(float64(level-1), 49)) / 49.0 // 0.0 at L1, 1.0 at L50

	// Base stat range for this level
	levelMin := baseMin + (baseMax-baseMin)*levelProgress
	levelMax := baseMin*1.3 + (baseMax*1.3-baseMin*1.3)*levelProgress // Allow some variance

	// Apply archetype multiplier
	multiplierRange := templateRange[1] - templateRange[0]
	multiplier := templateRange[0] + randFloat64()*multiplierRange

	// Calculate final stat within the level range, modified by archetype
	statMin := levelMin * 0.85 // Allow slight deviation below expected min
	statMax := levelMax * 1.15 // Allow slight deviation above expected max

	// Add variance for more natural feel
	varianceRange := statMax - statMin
	baseStat := statMin + randFloat64()*varianceRange

	// Apply archetype bonus/penalty
	finalStat := baseStat * multiplier

	// Clamp to reasonable bounds (5-98 range, never 100 for AI)
	if finalStat < 5 {
		finalStat = 5 + randFloat64()*3 // Minimum viable stat
	}
	if finalStat > 98 {
		finalStat = 95 + randFloat64()*3 // Maximum AI stat (never perfect 100)
	}

	return roundToOneDecimal(finalStat)
}

// CalculateHealth calculates health based on level and archetype.
func CalculateHealth(level int, templateRange [2]float64) float64 {
	// Health scales differently - starts higher and has more variance
	baseHealth := 50.0 + float64(level)*1.5 // L1=51, L50=225 base
	maxHealth := baseHealth * 1.2

	multiplierRange := templateRange[1] - templateRange[0]
	multiplier := templateRange[0] + randFloat64()*multiplierRange

	health := baseHealth + randFloat64()*(maxHealth-baseHealth)
	health = health * multiplier

	// Clamp health (30-250 reasonable range)
	if health < 30 {
		health = 30 + randFloat64()*10
	}
	if health > 250 {
		health = 240 + randFloat64()*10
	}

	return roundToOneDecimal(health)
}

// CalculateEnergy calculates energy/stamina based on level and archetype.
func CalculateEnergy(level int, templateRange [2]float64) float64 {
	// Energy scales with level - endurance increases
	baseEnergy := 40.0 + float64(level)*1.2 // L1=41, L50=100 base
	maxEnergy := baseEnergy * 1.3

	multiplierRange := templateRange[1] - templateRange[0]
	multiplier := templateRange[0] + randFloat64()*multiplierRange

	energy := baseEnergy + randFloat64()*(maxEnergy-baseEnergy)
	energy = energy * multiplier

	// Clamp energy (25-150 reasonable range)
	if energy < 25 {
		energy = 25 + randFloat64()*8
	}
	if energy > 150 {
		energy = 140 + randFloat64()*10
	}

	return roundToOneDecimal(energy)
}

// roundToOneDecimal rounds a float to one decimal place.
func roundToOneDecimal(val float64) float64 {
	return math.Round(val*10) / 10
}
