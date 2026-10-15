/**
 * Opponent Discovery API - Data access layer for opponent matching (MAT-102, MAT-120)
 */

import { apiGetJson } from './client';

/**
 * Fetch a list of opponents for a boxer with optional filters.
 *
 * @param {number} boxerId - The boxer ID to find opponents for
 * @param {Object} filters - Filter parameters
 * @param {boolean} filters.availableOnly - Only show available (not in fight) boxers
 * @param {number} filters.minLevel - Minimum level filter
 * @param {number} filters.maxLevel - Maximum level filter
 * @param {string} filters.sortBy - Sort field (score, level, name, etc.)
 * @returns {Promise<Array>} Array of opponent objects with scoring data
 */
export async function fetchOpponents(boxerId, filters = {}) {
  const params = new URLSearchParams();

  if (filters.availableOnly !== undefined) {
    params.append('available_only', filters.availableOnly.toString());
  }
  if (filters.minLevel !== undefined) {
    params.append('min_level', filters.minLevel.toString());
  }
  if (filters.maxLevel !== undefined) {
    params.append('max_level', filters.maxLevel.toString());
  }
  if (filters.sortBy) {
    params.append('sort_by', filters.sortBy);
  }

  const url = `/api/boxers/${boxerId}/opponents?${params}`;
  const data = await apiGetJson(url);

  return Array.isArray(data) ? data : [];
}

/**
 * Fetch the best match for a boxer.
 *
 * @param {number} boxerId - The boxer ID to find best match for
 * @returns {Promise<Object|null>} Best matching opponent or null if none found
 */
export async function fetchBestMatch(boxerId) {
  try {
    const data = await apiGetJson(`/api/opponents/best_match?boxer_id=${boxerId}`);
    return data;
  } catch (error) {
    // 404 means no match found
    if (error.status === 404) {
      return null;
    }
    throw error;
  }
}

/**
 * Validate that a fight can be booked between two boxers.
 *
 * @param {Object} validationData - Validation parameters
 * @param {number} validationData.boxerId - Primary boxer ID
 * @param {number} validationData.opponentId - Opponent boxer ID
 * @param {string} validationData.scheduledTime - ISO datetime string
 * @returns {Promise<Object>} Validation result with isValid and message
 */
export async function validateFight(validationData) {
  const params = new URLSearchParams({
    boxer_id: validationData.boxerId.toString(),
    opponent_id: validationData.opponentId.toString(),
    scheduled_time: validationData.scheduledTime,
  });

  const data = await apiGetJson(`/api/opponents/validate?${params}`);
  return data;
}

/**
 * Calculate match quality score for display.
 *
 * @param {number} score - Raw match score from backend
 * @returns {Object} Quality label, emoji, and color class
 */
export function getMatchQuality(score) {
  if (score >= 80) {
    return { label: 'Excellent', emoji: '⭐', className: 'quality-excellent' };
  } else if (score >= 60) {
    return { label: 'Great', emoji: '👍', className: 'quality-good' };
  } else if (score >= 40) {
    return { label: 'Good', emoji: '👌', className: 'quality-fair' };
  } else if (score >= 20) {
    return { label: 'Fair', emoji: '📊', className: 'quality-poor' };
  } else {
    return { label: 'Poor', emoji: '❌', className: 'quality-very-poor' };
  }
}
