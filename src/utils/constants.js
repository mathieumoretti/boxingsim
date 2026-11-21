/**
 * constants - Shared application constants (MAT-153)
 *
 * Centralized constants for game mechanics including health thresholds,
 * recovery calculations, and fatigue values.
 */

// Health system constants
export const MinHealthToFight = 50; // Minimum health percentage required to fight
export const MaxHealth = 100;       // Maximum health value

// Recovery duration thresholds (in hours)
export const MinorInjuryRecoveryHours = 24;   // ≤20 HP lost
export const ModerateInjuryRecoveryHours = 48; // 21-40 HP lost
export const MajorInjuryRecoveryHours = 72;    // >40 HP lost

// Health damage thresholds for recovery calculation
export const ModerateDamageThreshold = 20; // ≤20 HP = minor injury
export const MajorDamageThreshold = 40;    // >20 HP = moderate, >40 = major

// Recovery system constants
export const HealthRestorePerRecoveryHour = 5;  // HP restored per hour of rest
export const MaxHealthRestoreInOneRest = 30;    // Max HP restored in single rest event

// Fatigue system constants
export const ExhaustionThreshold = 80; // Fatigue level that blocks training
export const MinFatigueScore = 0;     // Minimum fatigue score (floor)
export const MaxFatigueScore = 100;   // Maximum fatigue score (cap)
export const DefaultForcedRestDays = 3; // Default forced rest duration when exhausted
