package seeding

import (
	"fmt"

	"github.com/mormm/boxing/internal/model"
)

// PopulationGenerator orchestrates the generation of AI boxer populations.
type PopulationGenerator struct {
	nameGenerator  *NameGenerator
	totalGenerated int
}

// NewPopulationGenerator creates a new PopulationGenerator.
func NewPopulationGenerator() *PopulationGenerator {
	return &PopulationGenerator{
		nameGenerator:  NewNameGenerator(),
		totalGenerated: 0,
	}
}

// AIBoxer represents a complete AI boxer with all generated data.
type AIBoxer struct {
	Name      BoxerName
	Archetype FighterArchetype
	Stats     BoxerStats
	Career    CareerRecord
	Level     int
	PositionX float64
	PositionY float64
	UserID    int // The system/AI user ID this boxer belongs to
}

// GenerateAIBoxer creates a single AI boxer with all attributes.
func (g *PopulationGenerator) GenerateAIBoxer(userID int) AIBoxer {
	// Determine gender (mostly male for traditional boxing feel, but include females)
	var gender string
	if randFloat64() < 0.15 { // 15% female boxers
		gender = "female"
	} else {
		gender = "male"
	}

	// Generate name with nickname (70% include nickname)
	name := g.nameGenerator.GenerateBoxerName(gender, true)

	// Select archetype
	archetype := RandomArchetype()

	// Generate level (pyramid distribution)
	level := GenerateRandomLevel()

	// Generate stats based on level and archetype
	stats := GenerateBoxerStats(level, archetype)

	// Generate career record
	career := GenerateCareerRecord(level, archetype, stats)

	// Generate map position
	posX, posY := GenerateMapPosition()

	g.totalGenerated++

	return AIBoxer{
		Name:      name,
		Archetype: archetype,
		Stats:     stats,
		Career:    career,
		Level:     level,
		PositionX: posX,
		PositionY: posY,
		UserID:    userID,
	}
}

// GenerateAIPopulation creates a batch of AI boxers.
func (g *PopulationGenerator) GenerateAIPopulation(count int, userID int) []AIBoxer {
	boxers := make([]AIBoxer, count)
	for i := range boxers {
		boxers[i] = g.GenerateAIBoxer(userID)
	}
	return boxers
}

// ToBoxerCreate converts an AIBoxer to the model.BoxerCreate format.
func (b AIBoxer) ToBoxerCreate() *model.BoxerCreate {
	nickname := b.Name.Nickname

	return &model.BoxerCreate{
		Name:      b.Name.FullName,
		Nickname:  nickname,
		PositionX: b.PositionX,
		PositionY: b.PositionY,
		Strength:  b.Stats.Strength,
		Defense:   b.Stats.Defense,
		Agility:   b.Stats.Agility,
	}
}

// PrintSummary prints a human-readable summary of an AI boxer.
func (b AIBoxer) PrintSummary() {
	fmt.Printf("  %-35s (%s)\n", b.Name.DisplayName, b.Archetype.Name)
	fmt.Printf("    Level: %d | XP: %.0f\n", b.Level, GenerateExperience(b.Level))
	fmt.Printf("    Stats: STR %.1f | DEF %.1f | AGI %.1f | HP %.0f | EN %.0f\n",
		b.Stats.Strength, b.Stats.Defense, b.Stats.Agility, b.Stats.Health, b.Stats.Energy)
	fmt.Printf("    Record: %dW-%dL-%dD (%d KO)\n", b.Career.Wins, b.Career.Losses, b.Career.Draws, b.Career.KOs)
	fmt.Printf("    Position: (%.1f, %.1f)\n", b.PositionX, b.PositionY)
}

// GenerateBoxersForSeeding generates the complete set of boxer creation requests.
// This is the main entry point for the seeding system.
func GenerateBoxersForSeeding(count int, userID int) ([]*model.BoxerCreate, error) {
	generator := NewPopulationGenerator()
	boxers := generator.GenerateAIPopulation(count, userID)

	creationRequests := make([]*model.BoxerCreate, count)
	for i, boxer := range boxers {
		creationRequests[i] = boxer.ToBoxerCreate()
	}

	return creationRequests, nil
}
