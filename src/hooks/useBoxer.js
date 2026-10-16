/**
 * useBoxer Hook - React hook for fetching and managing boxer data (MAT-120, MAT-129)
 */

import { useState, useEffect, useCallback } from 'react';
import { fetchBoxer as apiFetchBoxer, fetchUpcomingFight as apiFetchUpcomingFight } from '../api/fights';
import { getUserErrorMessage } from '../utils/errorMessages';

/**
 * Custom hook to fetch and manage boxer data.
 *
 * @param {number|null} boxerId - The boxer ID to fetch (null/undefined disables the hook)
 * @returns {Object} Hook result with boxer data, loading state, error, and refresh function
 */
export function useBoxer(boxerId) {
  const [boxer, setBoxer] = useState(null);
  const [upcomingFight, setUpcomingFight] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  const refresh = useCallback(async (isRetry = false) => {
    if (!boxerId) return;

    setIsLoading(true);
    if (isRetry) {
      setError(null);
    }

    try {
      // Fetch boxer data and upcoming fight in parallel
      const [boxerData, fightData] = await Promise.all([
        apiFetchBoxer(boxerId),
        apiFetchUpcomingFight(boxerId)
      ]);

      setBoxer(boxerData);
      setUpcomingFight(fightData);
    } catch (err) {
      // Use structured error message from utility
      const errorMsg = getUserErrorMessage(err, 'boxerDetails');
      setError({
        message: errorMsg,
        raw: err, // Full ApiError for debugging
        code: err.status || 0
      });
      console.error('Error in useBoxer:', err);
    } finally {
      setIsLoading(false);
    }
  }, [boxerId]);

  // Retry handler exposed to components
  const retry = useCallback(() => {
    refresh(true);
  }, [refresh]);

  // Initial fetch and when boxerId changes
  useEffect(() => {
    refresh();
  }, [refresh]);

  return {
    boxer,
    upcomingFight,
    isLoading,
    error,
    refresh,
    retry
  };
}

export default useBoxer;
