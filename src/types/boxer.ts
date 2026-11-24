// Boxer types (extending existing or creating if needed)

/**
 * Boxer model with recovery system fields (MAT-153)
 */
export interface Boxer {
  id: number;
  user_id: number;
  name: string;
  weight_class: string;
  strength: number;
  defense: number;
  agility: number;
  health: number;           // Current health (0-100)
  energy: number;          // Current energy (0-100)
  fatigue_score: number;   // Current fatigue level (0-100)

  // MAT-153: Recovery system fields
  forced_recovery_until?: string | null;  // ISO timestamp when recovery ends
  next_available_at?: string | null;      // ISO timestamp when boxer can act again

  // Career stats
  wins: number;
  losses: number;
  draws: number;

  created_at: string;
  updated_at: string;
}

/**
 * Recovery status for a boxer
 */
export interface RecoveryStatus {
  is_on_recovery: boolean;
  recovery_until?: Date | null;
  time_remaining?: string;    // Human-readable remaining time (e.g., "2 days 3 hours")
  percentage_healed?: number; // Progress percentage (0-100)
}

/**
 * Check if boxer can fight based on health and recovery
 */
export interface FightEligibility {
  can_fight: boolean;
  reason?: string;  // If not eligible, explains why
}
