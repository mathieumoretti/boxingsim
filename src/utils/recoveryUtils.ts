/**
 * Recovery System Utilities (MAT-153)
 * Helper functions for recovery status calculations and formatting
 */

import { Boxer, RecoveryStatus } from '../types/boxer';

/**
 * Check if a boxer is currently on forced recovery
 */
export function isOnForcedRecovery(boxer: Boxer): boolean {
  if (!boxer.forced_recovery_until) {
    return false;
  }

  const now = new Date();
  const recoveryUntil = new Date(boxer.forced_recovery_until);
  return now < recoveryUntil;
}

/**
 * Get recovery status for a boxer
 */
export function getRecoveryStatus(boxer: Boxer): RecoveryStatus {
  if (!boxer.forced_recovery_until) {
    return {
      is_on_recovery: false,
    };
  }

  const now = new Date();
  const recoveryUntil = new Date(boxer.forced_recovery_until);

  if (now >= recoveryUntil) {
    // Recovery has ended but boxers data might not be refreshed yet
    return {
      is_on_recovery: false,
    };
  }

  const timeRemainingMs = recoveryUntil.getTime() - now.getTime();
  const status: RecoveryStatus = {
    is_on_recovery: true,
    recovery_until: recoveryUntil,
    time_remaining: formatTimeRemaining(timeRemainingMs),
  };

  // Calculate healing progress based on health restoration rate (5 HP/hour)
  const maxHealth = 100;
  const healthRestorePerHour = 5;

  // Estimate how much health could have been restored so far
  const totalRecoveryHours = timeRemainingMs / (1000 * 60 * 60);
  const potentialRestoration = Math.min(totalRecoveryHours * healthRestorePerHour, 30); // Max 30 HP per rest

  if (boxer.health < maxHealth) {
    status.percentage_healed = Math.min(100, (potentialRestoration / (maxHealth - boxer.health)) * 100);
  } else {
    status.percentage_healed = 100;
  }

  return status;
}

/**
 * Format milliseconds to human-readable time remaining string
 */
export function formatTimeRemaining(ms: number): string {
  if (ms <= 0) {
    return 'Ready';
  }

  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    const remainingHours = hours % 24;
    return `${days} day${days !== 1 ? 's' : ''}${remainingHours > 0 ? ` ${remainingHours}h` : ''}`;
  } else if (hours > 0) {
    const remainingMinutes = minutes % 60;
    return `${hours} hour${hours !== 1 ? 's' : ''}${remainingMinutes > 0 ? ` ${remainingMinutes}m` : ''}`;
  } else if (minutes > 0) {
    return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
  } else {
    return `${seconds} second${seconds !== 1 ? 's' : ''}`;
  }
}

/**
 * Check if boxer can fight based on health and recovery status
 */
export function canFight(boxer: Boxer): boolean {
  // Minimum health threshold (matches backend constant MinHealthToFight = 50.0)
  const minHealth = 50;

  if (boxer.health < minHealth) {
    return false;
  }

  if (isOnForcedRecovery(boxer)) {
    return false;
  }

  return true;
}

/**
 * Get reason why boxer cannot fight
 */
export function getFightEligibilityReason(boxer: Boxer): string | null {
  const minHealth = 50;

  if (boxer.health < minHealth) {
    return `Health too low (${Math.round(boxer.health)}%). Need at least ${minHealth}% to fight.`;
  }

  if (isOnForcedRecovery(boxer)) {
    const status = getRecoveryStatus(boxer);
    return `Boxer is on forced recovery. Available in ${status.time_remaining}.`;
  }

  return null;
}

/**
 * Get consumable type display name
 */
export function getConsumableTypeDisplayName(type: string): string {
  const typeMap: Record<string, string> = {
    energy: 'Energy Restoration',
    health: 'Health Restoration',
    recovery: 'Fatigue Reduction',
  };
  return typeMap[type] || type;
}

/**
 * Get consumable icon class based on type
 */
export function getConsumableIconClass(type: string): string {
  const iconMap: Record<string, string> = {
    energy: 'icon-energy',
    health: 'icon-health',
    recovery: 'icon-recovery',
  };
  return iconMap[type] || 'icon-consumable';
}
