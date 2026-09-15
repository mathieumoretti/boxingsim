# Game Mechanics Documentation

This document describes the core game mechanics for the Boxing Simulator, including fatigue management, training systems, and rest period mechanics.

---

## Table of Contents

1. [Fatigue System](#fatigue-system)
2. [Rest Period System](#rest-period-system)
3. [Training System](#training-system)
4. [Progression System](#progression-system)

---

## Fatigue System

The fatigue system models boxer exhaustion from training activities. Boxers accumulate fatigue points during training and recover during rest periods.

### Core Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `FatiguePerHour` | 15.0 | Fatigue points gained per hour of training |
| `DailyFatigueDecay` | 5.0 | Natural fatigue reduction per game day (no training) |
| `ExhaustionThreshold` | 80.0 | Threshold above which boxer cannot train |
| `MaxFatigueScore` | 100.0 | Maximum possible fatigue score |
| `MinFatigueScore` | 0.0 | Minimum possible fatigue score |

### Fatigue Accumulation Formula

```
fatigue_increase = duration_hours × 15.0
```

**Example:** A 3-hour training session increases fatigue by:
```
3.0 × 15.0 = 45.0 fatigue points
```

### Training Constraints

A boxer can train if and only if:
1. `FatigueScore < 80.0` (below exhaustion threshold)
2. No active forced rest period (`ForcedRestUntil` is nil or in the past)

If either constraint is violated, training is blocked with an error message.

### Forced Rest

When a boxer's fatigue reaches or exceeds the exhaustion threshold (80.0), they may be placed on mandatory forced rest:
- Training is completely blocked until `ForcedRestUntil` passes
- The forced rest period ends when `current_time >= ForcedRestUntil`
- After recovery, if fatigue drops below 80.0, the forced rest is automatically cleared

---

## Rest Period System

The rest period system recovers boxer energy and reduces fatigue after training sessions. Rest periods are automatically scheduled after training completion based on session duration and current fatigue levels.

### Rest Duration Formula

Rest duration is calculated using this formula:

```
base_rest_hours = ceil(training_duration_hours / 2)
fatigue_multiplier = max(1.0, fatigue_score / 60.0)
final_rest_hours = ceil(base_rest_hours × fatigue_multiplier)
```

**Where:**
- `training_duration_hours`: Duration of the completed training session (in hours)
- `fatigue_score`: Boxer's fatigue score immediately after training completion
- `ceil()`: Ceiling function rounding up to nearest integer
- `max(1.0, ...)`: Ensures minimum multiplier of 1.0 (no reduction for low fatigue)

### Rest Duration Examples

| Training Duration | Fatigue Score | Base Hours | Multiplier | Final Rest Hours |
|-------------------|---------------|------------|------------|------------------|
| 2 hours | 45.0 | ceil(2/2) = 1 | max(1.0, 45/60) = 1.0 | ceil(1 × 1.0) = **1 hour** |
| 3 hours | 45.0 | ceil(3/2) = 2 | max(1.0, 45/60) = 1.0 | ceil(2 × 1.0) = **2 hours** |
| 4 hours | 75.0 | ceil(4/2) = 2 | max(1.0, 75/60) = 1.25 | ceil(2 × 1.25) = **3 hours** |
| 6 hours | 90.0 | ceil(6/2) = 3 | max(1.0, 90/60) = 1.5 | ceil(3 × 1.5) = **5 hours** |
| 8 hours | 100.0 | ceil(8/2) = 4 | max(1.0, 100/60) = 1.67 | ceil(4 × 1.67) = **7 hours** |

### Recovery Benefits by Duration

The recovery benefits are tiered based on rest duration (in days):

| Rest Duration (Days) | Energy Recovery | Fatigue Reduction | Stat Decay Risk | Notes |
|----------------------|-----------------|-------------------|-----------------|-------|
| 1 day | 50% | 20.0 points | 0% | Light recovery |
| 2 days | 80% | 40.0 points | 0% | Moderate recovery |
| 3-6 days | 100% | 60.0 points | 0% | Full recovery with optimal fatigue reduction |
| 7+ days | 100% | 100.0 points (full reset) | **5%** | Extended rest causes minor stat decay |

**Notes on Stat Decay:**
- Stat decay applies to Strength, Defense, and Agility equally
- A 5% decay means: `new_stat = old_stat × 0.95`
- Example: Strength of 80.0 becomes 76.0 after a week-long rest

### Automatic Rest Scheduling

Rest periods are automatically scheduled when a training session completes:

1. Calculate rest duration using the formula above
2. Determine if forced rest is needed (`fatigue_score >= 80.0`)
3. Create a scheduled event with `event_type = "rest"` at `current_time + final_rest_hours`
4. Event data includes:
   - `rest_hours`: Calculated duration in hours
   - `forced_rest`: Boolean flag if exhaustion threshold exceeded
   - `training_id`: Reference to the completed training session

### Forced Rest During Exhaustion

If a boxer's fatigue score reaches or exceeds 80.0 after training, they are placed on forced rest:
- The `ForcedRestUntil` timestamp is set for the duration of the rest period
- Training remains blocked until both conditions are met:
  - `current_time >= ForcedRestUntil`
  - `FatigueScore < 80.0` (after recovery)

---

## Training System

### Overview

Training sessions allow boxers to improve their Strength, Defense, and Agility stats at the cost of energy points and fatigue accumulation.

### Energy Cost Calculation

```
energy_cost = training_type.energy_cost_per_hour × duration_hours
```

**Validation:** Boxer must have `Energy >= energy_cost` to schedule a training session.

### Stat Gains with Diminishing Returns

Effective stat gains are calculated using the progression service, which applies:
1. **Diminishing returns** based on current stat levels
2. **Fatigue modifier** that reduces efficiency as fatigue increases

The formula (simplified):
```
effective_gain = planned_gain × diminishing_return_multiplier × fatigue_modifier
```

### Experience and Leveling

Training sessions award experience points (XP). When accumulated XP reaches the level threshold, the boxer levels up, which may provide stat bonuses or other benefits.

---

## API Response Fields

The following fields are included in `BoxerResponse` to expose rest period status:

### has_active_rest (boolean)

Indicates whether the boxer currently has an active rest period.

```json
{
  "has_active_rest": true
}
```

**Logic:** This field is true when:
- `forced_rest_until != null` AND `current_time < forced_rest_until`

### next_available_training (timestamp, nullable)

The earliest time at which the boxer can begin a new training session.

```json
{
  "next_available_training": "2026-09-15T18:30:00Z"
}
```

**Value Logic:**
- If `has_active_rest = true`: Set to `forced_rest_until` timestamp
- If `has_active_rest = false` AND `fatigue_score < 80.0`: Set to `null` (available now)
- If `has_active_rest = false` AND `fatigue_score >= 80.0`: Set to `null` but training will be rejected

### Example Boxer Response with Rest

```json
{
  "id": 42,
  "name": "Mike Tyson",
  "energy": 65.0,
  "fatigue_score": 85.5,
  "forced_rest_until": "2026-09-15T18:30:00Z",
  "has_active_rest": true,
  "next_available_training": "2026-09-15T18:30:00Z"
}
```

### Example Boxer Response Available for Training

```json
{
  "id": 42,
  "name": "Mike Tyson",
  "energy": 100.0,
  "fatigue_score": 35.0,
  "forced_rest_until": null,
  "has_active_rest": false,
  "next_available_training": null
}
```

---

## Related Documentation

- [Game Design Document](game-design-document.md) - High-level game overview and architecture
- [System Design](system-design.md) - Technical system architecture
- [Database Design](database-design.md) - Schema and data model

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | September 2026 | Initial | Created rest period mechanics documentation (MAT-94) |
