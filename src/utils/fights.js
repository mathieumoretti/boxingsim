import { API_BASE_URL, authenticatedFetch } from './auth';

/**
 * Fetches the upcoming fight for a specific boxer.
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Object>} The fight data including opponent details
 */
export const fetchUpcomingFight = async (boxerId) => {
  const response = await authenticatedFetch(`/api/boxers/${boxerId}/upcoming-fight`, {
    method: 'GET',
  });

  if (!response.ok) {
    if (response.status === 404) {
      return null; // No upcoming fight found
    }
    throw new Error(`Failed to fetch upcoming fight: ${response.statusText}`);
  }

  return response.json();
};

/**
 * Fetches all active fights and filters for the specified boxer IDs.
 * More efficient than individual calls when checking multiple boxers.
 * @param {Array<number>} boxerIds - Array of boxer IDs to filter by
 * @returns {Promise<Array<Object>>} Array of fight objects matching the boxer IDs
 */
export const fetchNextFightForAllBoxers = async (boxerIds) => {
  if (!boxerIds || boxerIds.length === 0) {
    return [];
  }

  const response = await authenticatedFetch(`${API_BASE_URL}/fights/active`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch active fights: ${response.statusText}`);
  }

  const fights = await response.json();
  const fightsArray = Array.isArray(fights) ? fights : [];

  // Filter fights to only include those involving the specified boxers
  return fightsArray.filter(f =>
    boxerIds.includes(f.boxer1_id) || boxerIds.includes(f.boxer2_id)
  );
};

/**
 * Creates a map of boxerId -> fight data for efficient lookup.
 * @param {Array<Object>} boxers - Array of boxer objects
 * @param {Array<Object>} fights - Array of active fight objects
 * @returns {Object} Map of boxer ID to fight information
 */
export const createFightMap = (boxers, fights) => {
  const boxerIds = boxers.map(b => b.id);

  // Filter fights for these boxers
  const relevantFights = fights.filter(f =>
    boxerIds.includes(f.boxer1_id) || boxerIds.includes(f.boxer2_id)
  );

  // Create a map of boxerId -> fight info
  const fightMap = {};

  relevantFights.forEach(fight => {
    // For boxer1
    if (boxerIds.includes(fight.boxer1_id)) {
      fightMap[fight.boxer1_id] = {
        fight_id: fight.id,
        opponent_id: fight.boxer2_id,
        opponent_name: null, // We'll need to fetch or include opponent names separately
        scheduled_time: fight.scheduled_time,
        status: fight.status,
        rounds: fight.round || 12
      };
    }

    // For boxer2
    if (boxerIds.includes(fight.boxer2_id) && boxerIds.includes(fight.boxer1_id)) {
      fightMap[fight.boxer2_id] = {
        fight_id: fight.id,
        opponent_id: fight.boxer1_id,
        opponent_name: null,
        scheduled_time: fight.scheduled_time,
        status: fight.status,
        rounds: fight.round || 12
      };
    } else if (boxerIds.includes(fight.boxer2_id)) {
      fightMap[fight.boxer2_id] = {
        fight_id: fight.id,
        opponent_id: fight.boxer1_id,
        opponent_name: null,
        scheduled_time: fight.scheduled_time,
        status: fight.status,
        rounds: fight.round || 12
      };
    }
  });

  return fightMap;
};

/**
 * Formats a date for display in the UI.
 * @param {Date|string} date - Date object or ISO string
 * @returns {string} Formatted date string (e.g., "Sep 20, 6:00 PM")
 */
export const formatFightDate = (date) => {
  if (!date) return '';

  const d = new Date(date);
  const options = { month: 'short', day: 'numeric', hour: 'numeric', hour12: true };
  return d.toLocaleDateString('en-US', options);
};
