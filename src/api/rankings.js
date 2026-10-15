/**
 * Rankings API - Data access layer for boxer rankings (MAT-100, MAT-120)
 */

import { apiGetJson } from './client';

const DEFAULT_LIMIT = 25;

/**
 * Ranking criteria options matching backend implementation
 */
export const RANKING_CRITERIA = [
  { value: 'win_rate', label: 'Win Rate' },
  { value: 'total_fights', label: 'Total Fights' },
  { value: 'level', label: 'Level' },
  { value: 'strength', label: 'Strength' },
  { value: 'power_score', label: 'Power Score' },
];

/**
 * Fetch boxer rankings sorted by criteria with pagination.
 *
 * @param {Object} params - Query parameters
 * @param {string} params.criteria - Sort criteria (default: 'win_rate')
 * @param {number} params.page - Page number (default: 1)
 * @param {number} params.limit - Results per page (default: 25)
 * @returns {Promise<Object>} Response with boxers array and pagination info
 */
export async function fetchRankings({
  criteria = 'win_rate',
  page = 1,
  limit = DEFAULT_LIMIT,
} = {}) {
  const params = new URLSearchParams({
    criteria,
    page: page.toString(),
    limit: limit.toString(),
  });

  const data = await apiGetJson(`/api/rankings?${params}`);

  // Normalize response structure
  return {
    boxers: data.boxers || [],
    totalCount: data.total_count || 0,
    page: data.page || page,
    limit: data.limit || limit,
    pages: Math.ceil((data.total_count || 0) / (limit || DEFAULT_LIMIT)),
  };
}

/**
 * Calculate display value for a boxer based on ranking criteria.
 *
 * @param {Object} boxer - Boxer object with stats
 * @param {string} criteria - The ranking criteria
 * @returns {string|number} Formatted display value
 */
export function getCriteriaDisplayValue(boxer, criteria) {
  switch (criteria) {
    case 'win_rate':
      const totalFights = boxer.wins + boxer.losses;
      return totalFights > 0
        ? `${((boxer.wins / totalFights) * 100).toFixed(1)}%`
        : 'N/A';

    case 'total_fights':
      return (boxer.wins || 0) + (boxer.losses || 0) + (boxer.draws || 0);

    case 'level':
      return boxer.level || 1;

    case 'strength':
      return Math.round(boxer.strength || 0);

    case 'power_score':
      // Power score formula: (wins * 2) + knockouts + (level * 3)
      const powerScore =
        ((boxer.wins || 0) * 2) +
        (boxer.knockouts || 0) +
        ((boxer.level || 1) * 3);
      return powerScore;

    default:
      return boxer.ranking_score || 0;
  }
}
