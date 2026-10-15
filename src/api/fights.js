/**
 * Fights API - Data access layer for fight-related operations (MAT-120)
 */

import { apiGetJson, apiPostJson } from './client';

/**
 * Fetches the fight history for a specific boxer (MAT-103).
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Array>} Array of fight records with opponent details
 */
export async function fetchFightHistory(boxerId) {
  const data = await apiGetJson(`/api/boxers/${boxerId}/fights-history`);
  return Array.isArray(data) ? data : [];
}

/**
 * Fetches the upcoming fight for a specific boxer.
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Object|null>} The fight data or null if no upcoming fight
 */
export async function fetchUpcomingFight(boxerId) {
  try {
    const data = await apiGetJson(`/api/boxers/${boxerId}/upcoming-fight`);
    return data;
  } catch (error) {
    // 404 means no upcoming fight
    if (error.status === 404) {
      return null;
    }
    throw error;
  }
}

/**
 * Fetches all active fights and filters for the specified boxer IDs.
 * More efficient than individual calls when checking multiple boxers.
 * @param {Array<number>} boxerIds - Array of boxer IDs to filter by
 * @returns {Promise<Array<Object>>} Array of fight objects matching the boxer IDs
 */
export async function fetchActiveFights(boxerIds) {
  if (!boxerIds || boxerIds.length === 0) {
    return [];
  }

  const data = await apiGetJson(`/api/fights/active`);
  const fightsArray = Array.isArray(data) ? data : [];

  // Filter fights to only include those involving the specified boxers
  return fightsArray.filter(fight =>
    boxerIds.includes(fight.boxer1_id) || boxerIds.includes(fight.boxer2_id)
  );
}

/**
 * Creates a map of boxerId -> fight data for efficient lookup.
 * @param {Array<Object>} boxers - Array of boxer objects
 * @param {Array<Object>} fights - Array of active fight objects
 * @returns {Object} Map of boxer ID to fight information
 */
export function createFightMap(boxers, fights) {
  const boxerIds = boxers.map(b => b.id);

  // Filter fights for these boxers
  const relevantFights = fights.filter(fight =>
    boxerIds.includes(fight.boxer1_id) || boxerIds.includes(fight.boxer2_id)
  );

  // Create a map of boxerId -> fight info
  const fightMap = {};

  relevantFights.forEach(fight => {
    // For boxer1
    if (boxerIds.includes(fight.boxer1_id)) {
      fightMap[fight.boxer1_id] = {
        fightId: fight.id,
        opponentId: fight.boxer2_id,
        opponentName: null, // We'll need to fetch or include opponent names separately
        scheduledTime: fight.scheduled_time,
        status: fight.status,
        rounds: fight.round || 12
      };
    }

    // For boxer2
    if (boxerIds.includes(fight.boxer2_id)) {
      fightMap[fight.boxer2_id] = {
        fightId: fight.id,
        opponentId: fight.boxer1_id,
        opponentName: null,
        scheduledTime: fight.scheduled_time,
        status: fight.status,
        rounds: fight.round || 12
      };
    }
  });

  return fightMap;
}

/**
 * Formats a date for display in the UI.
 * @param {Date|string} date - Date object or ISO string
 * @returns {string} Formatted date string (e.g., "Sep 20, 6:00 PM")
 */
export function formatFightDate(date) {
  if (!date) return '';

  const d = new Date(date);
  const options = { month: 'short', day: 'numeric', hour: 'numeric', hour12: true };
  return d.toLocaleDateString('en-US', options);
}

/**
 * Formats a date as time ago (e.g., "2 hours ago", "Yesterday at 3:42 PM") (MAT-103).
 * @param {Date|string} date - Date object or ISO string
 * @returns {string} Formatted time ago string
 */
export function formatTimeAgo(date) {
  if (!date) return '';

  const now = new Date();
  const d = new Date(date);
  const diffMs = now - d;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) {
    return 'Just now';
  } else if (diffMins < 60) {
    return `${diffMins}m ago`;
  } else if (diffHours < 24) {
    return `${diffHours}h ago`;
  } else if (diffDays === 1) {
    return 'Yesterday';
  } else if (diffDays < 7) {
    return `${diffDays} days ago`;
  } else {
    const options = { month: 'short', day: 'numeric' };
    return d.toLocaleDateString('en-US', options);
  }
}

/**
 * Gets the display label for a date group (MAT-103).
 * @param {string} key - The group key (today, yesterday, lastWeek, older)
 * @returns {string} Formatted label
 */
export function getGroupLabel(key) {
  const labels = {
    today: 'TODAY',
    yesterday: 'YESTERDAY',
    lastWeek: 'LAST WEEK',
    older: 'OLDER'
  };
  return labels[key] || key;
}

/**
 * Fetches details for a specific fight by ID (MAT-104).
 * @param {number} fightId - The fight's ID
 * @returns {Promise<Object>} Fight details with boxer info and play-by-play
 */
export async function fetchFightDetails(fightId) {
  const data = await apiGetJson(`/api/fights/${fightId}`);
  return data;
}

/**
 * Fetches details for a specific boxer by ID (MAT-104).
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Object>} Boxer details with stats
 */
export async function fetchBoxer(boxerId) {
  const data = await apiGetJson(`/api/boxers/${boxerId}`);
  return data;
}

/**
 * Books a fight between two boxers.
 * @param {number} boxerId - The boxer ID to schedule the fight for
 * @param {Object} fightData - Fight scheduling data
 * @param {number} fightData.opponentId - Opponent boxer ID
 * @param {string} fightData.scheduledTime - ISO datetime string for scheduled time
 * @param {number} fightData.rounds - Number of rounds (default: 12)
 * @returns {Promise<Object>} Created fight object
 */
export async function bookFight(boxerId, fightData) {
  const data = await apiPostJson(`/api/boxers/${boxerId}/fights`, fightData);
  return data;
}
