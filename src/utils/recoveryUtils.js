/**
 * recoveryUtils - Utility functions for recovery system (MAT-153)
 *
 * Provides helper functions for checking recovery status, formatting times,
 * and determining fight eligibility based on health and recovery state.
 */

import { MinHealthToFight } from './constants';

/**
 * Check if boxer is currently on forced recovery
 * @param {Object} boxer - Boxer object with forced_recovery_until field
 * @returns {boolean} True if boxer has an active forced recovery period
 */
export function isOnForcedRecovery(boxer) {
  if (!boxer || !boxer.forced_recovery_until) {
    return false;
  }

  const now = new Date();
  const recoveryUntil = new Date(boxer.forced_recovery_until);
  return now < recoveryUntil;
}

/**
 * Calculate recovery status with timing information
 * @param {Object} boxer - Boxer object with recovery and health fields
 * @returns {Object} Recovery status object with computed fields
 */
export function getRecoveryStatus(boxer) {
  const now = new Date();
  const recoveryUntil = boxer.forced_recovery_until ? new Date(boxer.forced_recovery_until) : null;

  const result = {
    is_on_recovery: isOnForcedRecovery(boxer),
    recovery_until: recoveryUntil,
    time_remaining: null,
    percentage_healed: undefined,
  };

  if (recoveryUntil && result.is_on_recovery) {
    const msRemaining = recoveryUntil.getTime() - now.getTime();
    result.time_remaining = formatTimeRemaining(msRemaining);

    // Calculate healing progress (health restoration during forced recovery)
    if (boxer.health !== undefined && boxer.max_health === 100) {
      // Assume full health is 100, calculate how far along they are in healing
      const healthGap = 100 - boxer.health;
      const minHealthThreshold = MinHealthToFight || 50;
      const maxHealthNeeded = 100 - minHealthThreshold;

      if (maxHealthNeeded > 0) {
        result.percentage_healed = ((maxHealthNeeded - healthGap) / maxHealthNeeded) * 100;
        result.percentage_healed = Math.max(0, Math.min(100, result.percentage_healed));
      }
    }
  }

  return result;
}

/**
 * Format milliseconds into human-readable time string
 * @param {number} ms - Milliseconds to format
 * @returns {string} Formatted time string (e.g., "2d 5h", "3h 15m")
 */
export function formatTimeRemaining(ms) {
  if (ms <= 0) return 'Ready';

  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    const remainingHours = hours % 24;
    return `${days}d ${remainingHours}h`;
  }
  if (hours > 0) {
    const remainingMinutes = minutes % 60;
    return `${hours}h ${remainingMinutes}m`;
  }
  if (minutes > 0) {
    return `${minutes}m`;
  }
  return 'Less than a minute';
}

/**
 * Check if boxer can fight based on health threshold
 * @param {Object} boxer - Boxer object with health field
 * @returns {boolean} True if boxer meets minimum health requirement
 */
export function canFight(boxer) {
  if (!boxer || !boxer.health !== undefined) {
    return false;
  }

  const minHealth = MinHealthToFight || 50;
  return boxer.health >= minHealth && !isOnForcedRecovery(boxer);
}

/**
 * Get human-readable reason why boxer cannot fight
 * @param {Object} boxer - Boxer object with health and recovery fields
 * @returns {string|null} Reason for ineligibility or null if eligible
 */
export function getFightEligibilityReason(boxer) {
  if (!boxer || typeof boxer.health === 'undefined') {
    return 'Invalid boxer data';
  }

  const minHealth = MinHealthToFight || 50;

  if (isOnForcedRecovery(boxer)) {
    const status = getRecoveryStatus(boxer);
    return `Boxer is on forced recovery. Available in ${status.time_remaining}.`;
  }

  if (boxer.health < minHealth) {
    return `Health too low (${Math.round(boxer.health)}%). Minimum required: ${minHealth}%.`;
  }

  return null; // Can fight
}

/**
 * Get display name for consumable type
 * @param {string} type - Consumable type ('energy', 'health', or 'recovery')
 * @returns {string} Human-readable type name
 */
export function getConsumableTypeDisplayName(type) {
  const names = {
    energy: 'Energy Restorers',
    health: 'Health Restorers',
    recovery: 'Fatigue Reduction',
  };
  return names[type] || type;
}

/**
 * Get icon class for consumable type
 * @param {string} type - Consumable type ('energy', 'health', or 'recovery')
 * @returns {string} Icon character for the type
 */
export function getConsumableIconClass(type) {
  const icons = {
    energy: '⚡',
    health: '❤️',
    recovery: '🌿',
  };
  return icons[type] || '📦';
}
