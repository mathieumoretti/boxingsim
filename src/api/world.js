/**
 * World API - Data access layer for world clock and game state (MAT-120)
 */

import { apiGetJson } from './client';

/**
 * Fetch current game world time.
 *
 * @returns {Promise<Object>} World time object with current_time timestamp
 */
export async function fetchWorldTime() {
  const data = await apiGetJson(`/api/world/time`);
  return data;
}

/**
 * Get the current game time as an ISO string.
 * Convenience wrapper that extracts just the timestamp.
 *
 * @returns {Promise<string|null>} ISO datetime string or null if unavailable
 */
export async function getCurrentGameTime() {
  try {
    const data = await fetchWorldTime();
    return data.current_time || null;
  } catch (error) {
    console.error('Failed to fetch game time:', error);
    return null;
  }
}
