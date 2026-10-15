/**
 * useBoxer Hook - React hook for fetching and managing boxer data (MAT-120)
 */

import { useState, useEffect, useCallback } from 'react';
import { fetchBoxer as apiFetchBoxer, fetchUpcomingFight as apiFetchUpcomingFight } from '../api/fights';

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

  const refresh = useCallback(async () => {
    if (!boxerId) return;

    setIsLoading(true);
    setError(null);

    try {
      // Fetch boxer data and upcoming fight in parallel
      const [boxerData, fightData] = await Promise.all([
        apiFetchBoxer(boxerId),
        apiFetchUpcomingFight(boxerId)
      ]);

      setBoxer(boxerData);
      setUpcomingFight(fightData);
    } catch (err) {
      setError(err.message || 'Failed to fetch boxer data');
      console.error('Error in useBoxer:', err);
    } finally {
      setIsLoading(false);
    }
  }, [boxerId]);

  // Initial fetch and when boxerId changes
  useEffect(() => {
    refresh();
  }, [refresh]);

  return {
    boxer,
    upcomingFight,
    isLoading,
    error,
    refresh
  };
}

export default useBoxer;
