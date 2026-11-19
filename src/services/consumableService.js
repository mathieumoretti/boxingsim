/**
 * Consumable Service (MAT-153 Phase 8)
 * Handles all API interactions for the consumables system
 */

import { fetchAPI } from './api';

/**
 * Fetch all available consumables (public endpoint - no auth required)
 * @returns {Promise<Array>} Array of consumable objects
 */
export async function getAllConsumables() {
  return fetchAPI('consumables');
}

/**
 * Fetch a boxer's consumable inventory with quantities and cooldown status
 * @param {number} boxerId - The boxer ID to fetch inventory for
 * @returns {Promise<Array>} Array of ConsumableInventoryItem objects
 */
export async function getBoxerInventory(boxerId) {
  return fetchAPI(`boxers/${boxerId}/consumables`);
}

/**
 * Use a consumable on a boxer
 * @param {number} boxerId - The boxer ID
 * @param {number} consumableId - The consumable ID to use
 * @returns {Promise<Object>} UseConsumableResponse object
 */
export async function useConsumable(boxerId, consumableId) {
  const request = {
    consumable_id: consumableId,
  };

  return fetchAPI(`boxers/${boxerId}/use-consumable`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Check if a consumable is currently on cooldown
 * @param {Object} item - ConsumableInventoryItem object
 * @returns {boolean} True if item is on cooldown
 */
export function isConsumableOnCooldown(item) {
  if (item.is_unlimited) {
    return false; // Unlimited items don't have cooldowns
  }

  if (!item.on_cooldown) {
    return false;
  }

  // Check if available_at has passed
  if (!item.available_at) {
    return false;
  }

  const now = new Date();
  const availableAt = new Date(item.available_at);
  return now < availableAt;
}

/**
 * Format cooldown time to human-readable string
 * @param {string|null|undefined} availableAtStr - ISO datetime string when item becomes available
 * @returns {string} Formatted time remaining (e.g., "2d 5h", "Ready")
 */
export function formatCooldownTime(availableAtStr) {
  if (!availableAtStr) {
    return 'Ready';
  }

  const now = new Date();
  const availableAt = new Date(availableAtStr);
  const diffMs = availableAt.getTime() - now.getTime();

  if (diffMs <= 0) {
    return 'Ready';
  }

  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffDays > 0) {
    return `${diffDays}d ${diffHours % 24}h`;
  } else if (diffHours > 0) {
    return `${diffHours}h ${diffMinutes % 60}m`;
  } else if (diffMinutes > 0) {
    return `${diffMinutes}m`;
  } else {
    return 'Ready';
  }
}

/**
 * Get cooldown CSS class based on status
 * @param {Object} item - ConsumableInventoryItem object
 * @returns {string} CSS class name ('on-cooldown', 'out-of-stock', or '')
 */
export function getCooldownClass(item) {
  if (isConsumableOnCooldown(item)) {
    return 'on-cooldown';
  }
  if (item.quantity === 0 && !item.is_unlimited) {
    return 'out-of-stock';
  }
  return '';
}
