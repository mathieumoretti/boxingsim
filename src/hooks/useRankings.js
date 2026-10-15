/**
 * useRankings Hook - React hook for fetching and managing boxer rankings (MAT-120, MAT-105)
 */

import { useState, useEffect, useCallback } from 'react';
import { fetchRankings as apiFetchRankings, RANKING_CRITERIA } from '../api/rankings';

const DEFAULT_LIMIT = 25;

/**
 * Custom hook to fetch and manage boxer rankings with pagination.
 *
 * @param {Object} options - Hook options
 * @param {string} options.criteria - Sort criteria (default: 'win_rate')
 * @param {number} options.page - Page number (default: 1)
 * @returns {Object} Hook result with rankings data, loading state, error, and control functions
 */
export function useRankings({ criteria = 'win_rate', page = 1 } = {}) {
  const [data, setData] = useState([]);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: DEFAULT_LIMIT,
    total: 0,
    pages: 0
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchRankings = useCallback(async (requestPage = page, requestCriteria = criteria) => {
    setIsLoading(true);
    setError(null);

    try {
      const result = await apiFetchRankings({
        criteria: requestCriteria,
        page: requestPage,
        limit: DEFAULT_LIMIT
      });

      setData(result.boxers || []);
      setPagination({
        page: requestPage,
        limit: DEFAULT_LIMIT,
        total: result.totalCount || 0,
        pages: result.pages || 0
      });
    } catch (err) {
      setError(err.message || 'Failed to fetch rankings');
      setData([]);
      console.error('Error in useRankings:', err);
    } finally {
      setIsLoading(false);
    }
  }, [criteria, page]);

  // Initial fetch and when criteria changes
  useEffect(() => {
    fetchRankings(1, criteria);
  }, [criteria]);

  return {
    data,
    pagination,
    isLoading,
    error,
    criteria,
    RANKING_CRITERIA,
    fetchRankings,
    setPage: (newPage) => fetchRankings(newPage, criteria)
  };
}

export default useRankings;
