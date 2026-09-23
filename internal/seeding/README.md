# Seeding Package

The `seeding` package provides AI boxer population generation for the boxing simulation game. This is part of **MAT-101: Implement AI boxer population generator**.

## Features

### Name Generation (`names.go`)
Generates realistic boxer names with ethnic diversity:
- 200+ male first names (American, African American, Latino, European, Filipino)
- 50+ female first names for mixed roster support
- 100+ surnames with global representation
- 100+ fighting nicknames ("Iron", "The Beast", "Money", etc.)

```go
generator := seeding.NewNameGenerator()
name := generator.GenerateBoxerName("male", true)
// Output: Marcus 'Iron' Jackson
```

### Fighter Archetypes (`archetypes.go`)
Ten distinct fighting styles with unique stat distributions:

| Archetype | Description | Stat Profile |
|-----------|-------------|--------------|
| Brawler | Power-focused, devastating strength | High STR, Med DEF, Low AGI |
| Technician | Skilled boxer, excellent footwork | Balanced + High AGI |
| Tank | Defensive specialist, high durability | High DEF/HP, Low AGI |
| Glass Cannon | Explosive power, poor defense | Very High STR/AGI, Low DEF |
| All-Rounder | No major weaknesses | Slightly above average all-around |
| Counterpuncher | Patient, waits for openings | Med-High DEF/AGI, Low-Med STR |
| Speedster | Lightning-fast combinations | Very High AGI, Lower STR/DEF |
| Warrior | Heart and determination | Above avg STR/HP, Low-Med AGI |
| Defensive Specialist | Master of defense/movement | High DEF/AGI, Low STR |
| Pressure Fighter | Relentless forward pressure | Med-High STR/HP, Low AGI |

### Stat Generation (`stats.go`)
Realistic stat distributions based on level and archetype:
- Level 1 boxers: stats 20-40 range
- Level 50 boxers: stats 80-95 range (never perfect 100)
- Pyramid level distribution (more low-level, fewer high-level)
- Career records with wins/losses/draws/knockouts

### Population Generator (`population.go`)
Main orchestrator for bulk AI boxer generation:

```go
boxers, err := seeding.GenerateBoxersForSeeding(100, userID)
```

## Usage

### CLI Commands

Generate 100 AI boxers (default):
```bash
make seed-pop
# or
go run cmd/seed/main.go population
```

Generate custom count:
```bash
make seed-pop COUNT=50
# or
go run cmd/seed/main.go population --count=200
```

### API Usage

```go
package main

import (
    "github.com/mormm/boxing/internal/seeding"
)

func main() {
    // Generate a single AI boxer
    generator := seeding.NewPopulationGenerator()
    boxer := generator.GenerateAIBoxer(1) // userID = 1
    
    // Access generated data
    fmt.Println(boxer.Name.FullName)        // "Marcus 'Iron' Jackson"
    fmt.Println(boxer.Archetype.Name)       // "Brawler"
    fmt.Println(boxer.Stats.Strength)       // 45.2
    fmt.Println(boxer.Career.Wins)          // 12
    
    // Generate bulk population
    boxers := generator.GenerateAIPopulation(100, 1)
    
    // Convert to model for database insertion
    createReq := boxer.ToBoxerCreate()
}
```

## Distribution Statistics

### Level Distribution (Pyramid)
- Levels 1-10: 50% of population
- Levels 11-20: 22% of population
- Levels 21-30: 14% of population
- Levels 31-40: 9% of population
- Levels 41-50: 5% of population

### Gender Distribution
- Male: ~85%
- Female: ~15%

### Nickname Frequency
- With nickname: ~70%
- Without nickname: ~30%

## Testing

Run all seeding tests:
```bash
go test ./internal/seeding/... -v
```

Test coverage includes:
- Name generation and diversity
- Archetype stat distributions
- Level distribution (pyramid)
- Stat calculations by level/archetype
- Career record generation
- Map position bounds
- Population diversity metrics

## Data Quality Guidelines

### Name Diversity
- Various ethnic backgrounds (boxing is global)
- Mix of common and uncommon names
- Authentic-sounding nicknames

### Stat Distribution
- Normal distribution curves where possible
- High-level fighters have imperfections
- Level 40+ fighters have clear specialties

### Realism Checks
- No fighter with all stats > 90 before level 40
- Knockout artists have STR > AGI correlation
- Technical boxers show AGI/DEF balance
