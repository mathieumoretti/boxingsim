import { API_BASE_URL, authenticatedFetch } from './auth';

/**
 * Fetches the fight history for a specific boxer (MAT-103).
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Array>} Array of fight records with opponent details
 */
export const fetchFightHistory = async (boxerId) => {
  const response = await authenticatedFetch(`/api/boxers/${boxerId}/fights-history`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch fight history: ${response.statusText}`);
  }

  return response.json();
};

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

/**
 * Formats a date as time ago (e.g., "2 hours ago", "Yesterday at 3:42 PM") (MAT-103).
 * @param {Date|string} date - Date object or ISO string
 * @returns {string} Formatted time ago string
 */
export const formatTimeAgo = (date) => {
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
};

/**
 * Gets the display label for a date group (MAT-103).
 * @param {string} key - The group key (today, yesterday, lastWeek, older)
 * @returns {string} Formatted label
 */
export const getGroupLabel = (key) => {
  const labels = {
    today: 'TODAY',
    yesterday: 'YESTERDAY',
    lastWeek: 'LAST WEEK',
    older: 'OLDER'
  };
  return labels[key] || key;
};

/**
 * Fetches details for a specific fight by ID (MAT-104).
 * @param {number} fightId - The fight's ID
 * @returns {Promise<Object>} Fight details
 */
export const fetchFightDetails = async (fightId) => {
  const response = await authenticatedFetch(`/api/fights/${fightId}`, {
    method: 'GET',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch fight details: ${response.statusText}`);
  }

  return response.json();
};
