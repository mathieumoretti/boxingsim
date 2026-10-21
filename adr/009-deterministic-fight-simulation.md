# 009 - Deterministic Fight Simulation with Versioning

**Date**: 2026-10-20  
**Status**: Accepted  
**Issue**: MAT-114 (Fight Simulation Audit)

## Context

The fight simulation system had several critical limitations:

1. **Non-deterministic results**: Used `math/rand` without seeding, making the same fight produce different results each time
2. **Hardcoded magic numbers**: Constants like damage multiplier (0.1), max rounds (12), evasion threshold (0.4) were scattered in code
3. **No version tracking**: No way to know which algorithm version produced a historical fight result
4. **Difficult debugging**: Couldn't reproduce specific fight outcomes for bug investigation

This made it impossible to:
- Replay fights identically
- Debug specific fight scenarios
- Tune balance parameters without code changes
- Track algorithm improvements over time

## Decision

Implement a deterministic, configurable, and versioned fight simulation system with:

### 1. Configuration Struct

```go
type FightSimulationConfig struct {
    Version          string  // Algorithm version (semver)
    MaxRounds        int     // Default: 12
    DamageMultiplier float64 // Default: 0.1
    EvasionThreshold float64 // Default: 0.4
    EnergyDrain      float64 // Default: 10.0
    ExperienceGain   float64 // Default: 50.0
}
```

All simulation parameters loaded from Viper config with environment variable overrides.

### 2. Seeded RNG for Determinism

```go
// Generate deterministic seed from fight data
seed := int64(fightID*1000 + boxer1ID + boxer2ID)
rng := rand.New(rand.NewSource(seed))

// Use seeded RNG in attack processing
damageVariance := rng.Float64() * 0.2
```

Same fight (same IDs) always produces same seed, ensuring identical results on replay.

### 3. Version Tracking

Database column `fights.simulation_version` records algorithm version per fight:

```sql
ALTER TABLE fights ADD COLUMN simulation_version VARCHAR(20) DEFAULT '1.0.0';
```

Enables:
- Knowing which algorithm produced each historical result
- Future algorithm changes without breaking replay of old fights
- A/B testing different simulation versions

### 4. Unit Test Coverage

Three test cases verify correctness:
- **Determinism**: Same seed → identical results (3 runs verified)
- **Different seeds**: Different fight IDs → different outcomes
- **Configurability**: Parameter changes affect results as expected

## Consequences

### Positive

1. **Replayability**: Any fight can be reproduced exactly for debugging or analysis
2. **Tunable balance**: Damage, rounds, evasion, etc. configurable without code changes
3. **Version control**: Algorithm evolution tracked per-fight
4. **Better testing**: Deterministic simulations are easier to test
5. **A/B testing ready**: Can run fights with different configs for balance experiments

### Negative

1. **Migration required**: Existing fights need backfill (handled in migration)
2. **Seed formula visibility**: The seed calculation is now part of the "contract" - changing it breaks replayability for old fights
3. **Slightly more complex code**: RNG management adds complexity vs simple `rand.Float64()`

### Neutral

1. **Backward compatible**: Default values match old hardcoded values
2. **Existing fights preserved**: All historical fights get version "1.0.0"
3. **No performance impact**: Seeded RNG has same performance as global rand

## Implementation Details

### Files Changed

- `internal/platform/config/config.go` - FightSimulationConfig struct and defaults
- `internal/model/fight.go` - SimulationVersion field in Fight and FightResponse
- `internal/store/fight_store.go` - SetSimulationVersionTx(), updated SQL queries
- `internal/fight/fight.go` - SimulateFight() uses config, processAttack() accepts RNG
- `db/migrations/00009_add_fight_simulation_version.up.sql` - New migration
- `internal/fight/simulation_test.go` - Determinism and configuration tests

### Seed Formula

```go
seed = fightID * 1000 + boxer1ID + boxer2ID
```

Chosen for simplicity and guarantees:
- Same fight always produces same seed
- Different fights produce different seeds (usually)
- No collisions within reasonable ID ranges

Could be improved later with hash-based seeding if needed.

### Damage Variance

Added ±10% random variance to damage to maintain some randomness while keeping determinism:

```go
damageVariance := rng.Float64() * 0.2 // Range [0, 0.2)
damageVariance = damageVariance - 0.1  // Center around 0 → Range [-0.1, 0.1)
damage := baseDamage * (1 + damageVariance)
```

## Alternatives Considered

### A. Global Config vs Per-Fight Config

**Chosen**: Global config loaded at startup

**Alternative**: Store simulation config per-fight in database
- **Pros**: Historical accuracy, can change configs retroactively
- **Cons**: Overkill for current needs, adds database overhead
- **Decision**: Global config is sufficient; versioning handles algorithm changes

### B. Hash-Based Seeding

**Chosen**: Simple arithmetic formula `fightID*1000 + boxer1ID + boxer2ID`

**Alternative**: Use cryptographic hash (SHA256) of fight data
- **Pros**: More robust against collisions, more "secure"
- **Cons**: Overkill for game simulation, harder to reason about
- **Decision**: Simple formula is transparent and sufficient

### C. Full Fight State Recording

**Chosen**: Record only simulation version

**Alternative**: Record full RNG seed state or random number sequence
- **Pros**: Perfect replay capability
- **Cons**: Database bloat, overkill for current needs
- **Decision**: Version tracking + seeded formula provides sufficient replayability

## Testing Verification

```bash
# All tests pass
go test ./internal/fight/... -v

=== RUN   TestDeterministicSimulation
✅ Determinism verified: Same seed produces identical results (version 1.0.0)
--- PASS: TestDeterministicSimulation (0.00s)

=== RUN   TestDifferentSeedsProduceDifferentResults  
✅ Different seeds verified: Fight 12345 (seed 12345003) vs Fight 67890 (seed 67890003)
--- PASS: TestDifferentSeedsProduceDifferentResults (0.00s)

=== RUN   TestConfigurableSimulationParameters
✅ Configurable parameters verified: Damage multiplier affects outcomes
--- PASS: TestConfigurableSimulationParameters (0.00s)
```

## Future Enhancements

1. **Fight Replay API**: Expose deterministic simulation via API for fight replays
2. **Algorithm Versioning**: Bump to v1.1.0 when making balance changes
3. **A/B Testing Framework**: Run parallel simulations with different configs
4. **Balance Dashboard**: UI for tweaking simulation parameters live

## References

- Issue: MAT-114 (Fight Simulation Audit)
- Migration: 00009_add_fight_simulation_version.up.sql
- Config defaults: `internal/platform/config/config.go` lines 68-77
