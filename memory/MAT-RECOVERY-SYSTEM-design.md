# MAT-RECOVERY-SYSTEM: Scheduled Rest & Healing Design Document

## Problem Statement

Boxers accumulate fatigue from training and lose health/energy from fights, but there is **no systematic healing mechanism** to restore these critical stats. This creates a "death spiral" where boxers become progressively weaker over time with no way to recover.

### Current Issues

1. **Health Never Recovers**: Boxers take damage in fights (can lose 30-60+ HP per fight) but have zero mechanism to heal
2. **No Post-Fight Recovery**: Unlike training which auto-schedules rest, fights don't trigger recovery periods
3. **Energy Management Gaps**: Rest events only recover 50-100% of energy; no long-term restoration system
4. **Forced Rest Inconsistent**: Training triggers forced rest at exhaustion (fatigue >= 80), but fights can critically damage boxers without triggering similar protection

## Terminology

| Term | Definition |
|------|------------|
| **Rest Event** | Scheduled event (`EventTypeRest`) that restores health, energy, and reduces fatigue when processed by worker |
| **Healing Period** | Duration of time a boxer spends recovering stats (1-7+ hours in game time) |
| **Forced Recovery** | Mandatory rest period preventing training/fight booking until completion (similar to current forced rest) |
| **Recovery Benefits** | Stat restoration amounts: health recovery %, energy recovery %, fatigue reduction points |

## Solution Architecture

### Overview

Extend the existing scheduled events system with comprehensive fight-induced recovery scheduling:

```
Fight Completed
       ↓
Calculate Recovery Needs (health damage, fatigue from fight)
       ↓
Schedule Rest Event (game_time + rest_duration)
       ↓
Worker Processes Rest Event
       ↓
Apply Recovery Benefits (health ↑, energy ↑, fatigue ↓)
       ↓
Clear Forced Recovery if no longer needed
```

### Component Changes

#### 1. Fight Service - Schedule Recovery After Fights

**File**: `internal/service/fight.go` (or `internal/fight/fight.go` for SimulateFight)

Add recovery scheduling after fight simulation completes:

```go
// Post-fight: schedule mandatory rest if boxer took significant damage or fatigue
func (s *FightService) SchedulePostFightRecovery(
    ctx context.Context,
    boxerID int,
    healthDamage float64,  // HP lost during fight
    fatigueIncrease float64,  // Fatigue accumulated during fight
) error {
    // Calculate required rest duration based on damage taken
    restHours := calculateRestDuration(healthDamage, fatigueIncrease)
    
    // Schedule forced recovery if damage is significant
    needsForcedRecovery := healthDamage > 20 || fatigueIncrease > 60
    
    // Create scheduled rest event with fight-specific data
    eventData := map[string]interface{}{
        "rest_hours":      restHours,
        "forced_recovery": needsForcedRecovery,
        "fight_id":        fightID,
        "health_damaged":  healthDamage,
        "fatigue_gained":  fatigueIncrease,
    }
    
    // ... schedule event using existing eventStore.Create()
}
```

**Rest Duration Calculation Logic**:
- Health damage ≤ 10 + fatigue ≤ 30 → **1 hour rest** (light recovery)
- Health damage 11-25 OR fatigue 31-60 → **2-4 hours rest** (moderate recovery)
- Health damage 26-50 OR fatigue 61-80 → **4-7 hours rest** (heavy recovery)
- Health damage > 50 OR fatigue > 80 → **7-14 hours rest** (critical recovery, forced)

#### 2. Fatigue Service - Enhanced Recovery Benefits

**File**: `internal/service/fatigue_service.go`

Current `ApplyRecovery()` already handles:
- ✅ Energy recovery (50-100% of max based on duration)
- ✅ Fatigue reduction (20-100 points based on duration)
- ⚠️ **Missing**: Health recovery!

**Add Health Recovery to Benefits Structure**:

```go
type RecoveryBenefits struct {
    HealthRecoveryPercent float64 // NEW: % of max health restored (50-100)
    EnergyRecoveryPercent float64
    FatigueReduction      float64
    StatDecayRisk         float64
}

func (s *FatigueService) GetRecoveryBenefits(restHours int, isPostFight bool) RecoveryBenefits {
    // Post-fight recovery restores MORE health than light rest
    if isPostFight {
        switch {
        case restHours == 1:
            return RecoveryBenefits{
                HealthRecoveryPercent: 50.0,  // NEW
                EnergyRecoveryPercent: 70.0,
                FatigueReduction:      30.0,
                StatDecayRisk:         0.0,
            }
        case restHours == 2:
            return RecoveryBenefits{
                HealthRecoveryPercent: 75.0,  // NEW
                EnergyRecoveryPercent: 90.0,
                FatigueReduction:      50.0,
                StatDecayRisk:         0.0,
            }
        case restHours >= 3 && restHours < 7:
            return RecoveryBenefits{
                HealthRecoveryPercent: 100.0, // NEW - FULL HEAL
                EnergyRecoveryPercent: 100.0,
                FatigueReduction:      80.0,
                StatDecayRisk:         0.0,
            }
        case restHours >= 7:
            return RecoveryBenefits{
                HealthRecoveryPercent: 100.0, // NEW
                EnergyRecoveryPercent: 100.0,
                FatigueReduction:      100.0,
                StatDecayRisk:         5.0,
            }
        }
    }
    // ... existing logic for training-based rest (unchanged)
}

func (s *FatigueService) ApplyRecovery(ctx context.Context, boxerID int, restHours int, isPostFight bool) error {
    // ... existing code ...
    
    benefits := s.GetRecoveryBenefits(restHours, isPostFight)
    
    // NEW: Apply health recovery (cap at 100)
    oldHealth := boxer.Health
    newHealth := math.Min(boxer.Health + benefits.HealthRecoveryPercent, 100.0)
    
    // Existing energy/fatigue recovery...
    
    boxer.Health = newHealth
    boxer.Energy = newEnergy
    boxer.FatigueScore = newFatigue
    
    s.logger.Info("Recovery applied for boxer ID=%d: health %.1f→%.1f (+%.1f), energy %.1f→%.1f, fatigue %.2f→%.2f",
        boxerID, oldHealth, newHealth, benefits.HealthRecoveryPercent, oldEnergy, newEnergy, oldFatigue, newFatigue)
    
    // ... rest of existing code
}
```

#### 3. Event Processor - Pass Fight Context to Recovery

**File**: `internal/service/event_processor.go`

Update `processRecovery()` to extract and forward fight context:

```go
func (p *EventProcessor) processRecovery(ctx context.Context, event *model.ScheduledEvent) error {
    var data map[string]any
    if len(event.EventData) > 0 {
        json.Unmarshal(event.EventData, &data)
    }
    
    restHours := getIntField(data, "rest_hours", 1)
    isPostFight := getBoolField(data, "is_fight_recovery", false) // NEW
    
    // Pass fight context to fatigue service for enhanced recovery rates
    if err := p.fatigueService.ApplyRecovery(ctx, event.BoxerID, restHours, isPostFight); err != nil {
        return fmt.Errorf("failed to apply recovery: %w", err)
    }
    
    return nil
}
```

#### 4. Boxer Model - Add Recovery Status Fields

**File**: `internal/model/boxer.go`

Track forced recovery state similar to existing `ForcedRestUntil`:

```go
type Boxer struct {
    // ... existing fields ...
    
    ForcedRestUntil     *time.Time `json:"forced_rest_until"`       // Existing: training-induced rest
    ForcedRecoveryUntil *time.Time `json:"forced_recovery_until"`   // NEW: fight-induced recovery
    
    // Optional: next available action time (min of forced_rest, forced_recovery)
    NextAvailableAt     *time.Time `json:"next_available_at,omitempty"`
}

// BoxerResponse additions for UI:
type BoxerResponse struct {
    // ... existing fields ...
    
    IsOnForcedRecovery  bool       `json:"is_on_forced_recovery"`   // NEW: combat rest active
    RecoveryEndTime     *time.Time `json:"recovery_end_time"`       // NEW: when recovery completes
    NeedsMedicalAttention bool      `json:"needs_medical_attention,omitempty"` // HEALTH < 30%
}
```

#### 5. Validation - Prevent Actions During Forced Recovery

**Files**: 
- `internal/service/fight.go` (BookFight)
- `internal/handler/training_handler.go` (CreateTrainingSession)

Add checks similar to existing `ForcedRestUntil`:

```go
// In BookFight():
if boxer.ForcedRecoveryUntil != nil && time.Now().Before(*boxer.ForcedRecoveryUntil) {
    return fmt.Errorf("boxer is on mandatory recovery from fight injuries (until %s)", 
        boxer.ForcedRecoveryUntil.Format(time.RFC3339))
}

// In CreateTrainingSession():
if boxer.ForcedRecoveryUntil != nil && time.Now().Before(*boxer.ForcedRecoveryUntil) {
    return fmt.Errorf("boxer is recovering from injuries - cannot train")
}
```

### Database Schema Changes

**New Column Needed**: `forced_recovery_until` (TIMESTAMP NULL) in boxers table

**Migration SQL**:
```sql
-- Add forced recovery tracking column
ALTER TABLE boxers 
ADD COLUMN forced_recovery_until TIMESTAMPTZ NULL,
ADD COLUMN next_available_at TIMESTAMPTZ NULL;

-- Add index for querying boxers in recovery
CREATE INDEX idx_boxers_forced_recovery ON boxers(forced_recovery_until) 
WHERE forced_recovery_until IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN boxers.forced_recovery_until IS 
'Timestamp until which boxer is on mandatory recovery from fight injuries (cannot train/fight)';
```

## Linear Issue Breakdown

### MAT-XXX: Backend Recovery Foundation (Core Infrastructure)

**Scope**: Database + model + store changes

- [ ] Migration: Add `forced_recovery_until` column to boxers table
- [ ] Model: Add `ForcedRecoveryUntil` field to Boxer struct
- [ ] Store: Add query for boxers in forced recovery state
- [ ] Response: Add `IsOnForcedRecovery`, `RecoveryEndTime`, `NeedsMedicalAttention` to API responses

**Acceptance Criteria**:
- ✅ Database schema updated with forced_recovery_until column
- ✅ Boxer model has new field properly serialized
- ✅ Unit tests verify column exists and can be null

---

### MAT-XXX: Enhanced Recovery Service (Business Logic)

**Scope**: FatigueService enhancement with health recovery

- [ ] Add `HealthRecoveryPercent` to `RecoveryBenefits` struct
- [ ] Update `GetRecoveryBenefits(restHours, isPostFight)` with fight-specific rates
- [ ] Modify `ApplyRecovery()` to restore health in addition to energy/fatigue
- [ ] Add helper: `CalculatePostFightRestDuration(healthDamage, fatigueGain)`
- [ ] Unit tests for recovery benefits calculation (training vs post-fight)

**Acceptance Criteria**:
- ✅ 1-hour post-fight rest recovers 50% health
- ✅ 3-hour+ post-fight rest fully heals boxer (100% health)
- ✅ Long rest (7+ hours) still applies stat decay tradeoff
- ✅ All existing training-based recovery unchanged

---

### MAT-XXX: Post-Fight Recovery Scheduling (Integration)

**Scope**: Fight service integration with event system

- [ ] Add `SchedulePostFightRecovery(ctx, boxerID, healthDamage, fatigueGain)` to FightService
- [ ] Call from SimulateFight() after fight completes for each participant
- [ ] Schedule forced recovery if: health damage > 20 OR fatigue gain > 60
- [ ] Set boxer.ForcedRecoveryUntil when scheduling forced recovery
- [ ] Integration test: fight → event created → recovery scheduled

**Acceptance Criteria**:
- ✅ Heavy fights (>50 HP damage) schedule 7+ hour rest events
- ✅ Forced recovery timestamp set correctly for severe injuries
- ✅ Rest event contains is_fight_recovery=true flag
- ✅ Worker can process post-fight rest events successfully

---

### MAT-XXX: Validation & Constraints (Guardrails)

**Scope**: Prevent actions during forced recovery

- [ ] Add `ForcedRecoveryUntil` check to BookFight() handler
- [ ] Add `ForcedRecoveryUntil` check to CreateTrainingSession() handler  
- [ ] Clear ForcedRecoveryUntil when rest event completes (in FatigueService.ApplyRecovery)
- [ ] Return descriptive error messages: "recovering from injuries until {timestamp}"
- [ ] Unit tests for validation constraints

**Acceptance Criteria**:
- ✅ Cannot book fight while boxer is in forced recovery
- ✅ Cannot schedule training during forced recovery period
- ✅ Forced recovery cleared automatically when rest event processes
- ✅ Error messages clearly communicate recovery end time

---

### MAT-XXX: Frontend Recovery UI (User Interface)

**Scope**: React components to display recovery status

- [ ] BoxerCard.jsx: Add "Injured" badge when health < 50% or in forced recovery
- [ ] BoxerProfilePage: Show recovery timer/countdown if forced recovery active
- [ ] TrainingScheduler: Disable button + tooltip if boxer recovering
- [ ] FightBookingPage: Filter out boxers in forced recovery from opponent list
- [ ] Add `recoveryEndTime` display with countdown timer

**Acceptance Criteria**:
- ✅ BoxerCard shows red "Injured" badge when health < 50%
- ✅ Recovery countdown displays time remaining until available
- ✅ Training button disabled with tooltip during forced recovery
- ✅ Opponent discovery filters out injured boxers by default (with toggle)

---

### MAT-XXX: Testing & Validation (Quality Assurance)

**Scope**: Integration tests + edge cases

- [ ] Integration test: Fight simulation → rest event creation → worker processing → health restored
- [ ] Edge case: Boxer at 10 HP takes fight → recovers to 100% after 3-hour rest
- [ ] Edge case: Forced recovery overlap (training exhaustion + fight injury)
- [ ] Load test: Multiple fights scheduled back-to-back trigger cascading recovery periods
- [ ] Manual test guide for developer verification

**Acceptance Criteria**:
- ✅ All integration tests pass with deterministic results
- ✅ Edge cases handled gracefully (no crashes, proper error messages)
- ✅ Documentation includes manual testing steps
- ✅ Worker logs show expected recovery processing flow

---

## Configuration Constants

Add to `config/fight.yaml`:

```yaml
fight:
  # ... existing config ...
  
  recovery:
    light_rest_hours: 1        # Health damage ≤ 10, fatigue ≤ 30
    moderate_rest_min_hours: 2 # Health damage 11-25 OR fatigue 31-60
    moderate_rest_max_hours: 4
    heavy_rest_min_hours: 4    # Health damage 26-50 OR fatigue 61-80
    heavy_rest_max_hours: 7
    critical_rest_min_hours: 7 # Health damage > 50 OR fatigue > 80
    critical_rest_max_hours: 14
    
    forced_recovery_threshold_health: 20  # Damage taken > 20 triggers forced recovery
    forced_recovery_threshold_fatigue: 60 # Fatigue gain > 60 triggers forced recovery
    
    health_recovery_rates:
      one_hour: 50.0     # % health restored (1 hour rest)
      two_hours: 75.0    # % health restored (2 hour rest)
      three_to_six_hours: 100.0  # Full heal (3-6 hour rest)
```

## Testing Strategy

### Unit Tests

1. **Recovery Benefits Calculation**:
   - Test `GetRecoveryBenefits()` for training vs post-fight scenarios
   - Verify health recovery percentages match configuration
   
2. **Rest Duration Calculation**:
   - Test threshold logic for light/moderate/heavy/critical classifications
   - Verify forced recovery triggers at correct damage levels

### Integration Tests

1. **Fight → Recovery Flow**:
   ```go
   t.Run("Fight schedules recovery event", func(t *testing.T) {
       // Simulate fight with health damage
       err := fightService.SimulateFight(fightID)
       assert.NoError(t, err)
       
       // Verify rest event created for both boxers
       events, err := eventStore.GetByBoxerID(ctx, boxer1ID)
       assert.Len(t, events, 1)
       assert.Equal(t, model.EventTypeRest, events[0].EventType)
   })
   ```

2. **Worker Processing Recovery**:
   ```go
   t.Run("Worker processes recovery event", func(t *testing.T) {
       // Create pending rest event with fight context
       // Run worker loop
       // Verify boxer health restored, forced recovery cleared
   })
   ```

### Manual Testing Steps

1. Schedule heavyweight fight (high damage configuration)
2. Observe post-fight: boxer health reduced (e.g., 80 → 35 HP)
3. Check scheduled events: rest event appears for both boxers
4. Advance game time until rest event processes
5. Verify boxer stats: health restored, fatigue reduced
6. Attempt to schedule training: blocked if forced recovery active
7. After recovery completes: training available again

## Related Issues

- MAT-96: Rest badge implementation (existing UI component)
- MAT-143: Automatic cleanup for old failed events (handles zombie rest events)
- MAT-129: Error handling standardization (apply to new validation errors)
- MAT-145: Fight booking page (add recovery status filtering)

## Success Metrics

- ✅ Boxers can fully recover health after fights (death spiral eliminated)
- ✅ No manual intervention required to restore boxer stats
- ✅ Forced recovery prevents further damage during vulnerable state
- ✅ Game loop sustainable: train → fight → recover → repeat

---

**Status**: Design document ready for implementation  
**Estimated Effort**: 18-24 hours across 6 issues (~3-4h per issue)  
**Dependencies**: None (extends existing event system, no new dependencies)
