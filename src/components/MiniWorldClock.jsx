import React, { useState, useEffect } from 'react';
import { fetchWorldTime } from '../api/world';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const MiniWorldClock = () => {
  const [timeData, setTimeData] = useState(null);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  const loadTime = async (isRetry = false) => {
    if (isRetry) {
      setError(null);
    }

    try {
      const data = await fetchWorldTime();
      setTimeData(data);
      setIsLoading(false);
    } catch (err) {
      console.warn('MiniWorldClock: Failed to load world time:', err);
      setError(getUserErrorMessage(err, 'general'));
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadTime();

    // Refresh every second to show live countdown
    const interval = setInterval(async () => {
      try {
        const data = await fetchWorldTime();
        setTimeData(data);
      } catch (err) {
        // Silently fail on refresh - time display will be stale but don't spam user with errors
        console.debug('MiniWorldClock: Poll update failed:', err.message);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return <div className="inline-flex items-center gap-2 bg-white/5 px-4 py-2 rounded-md border border-white/10 text-sm text-gray-500">Loading...</div>;
  }

  if (error) {
    return (
      <ErrorDisplay
        error={error}
        onRetry={() => loadTime(true)}
        variant="compact"
        title="Clock Unavailable"
      />
    );
  }

  // Dynamic status badge classes based on clock state
  const getStatusBadgeClass = (status) => {
    switch(status) {
      case 'running': return 'bg-green-100/80 text-green-700 border-green-300';
      case 'paused': return 'bg-yellow-100/80 text-yellow-700 border-yellow-300';
      default: return 'bg-gray-400/20 text-gray-500 border-gray-400/30';
    }
  };

  const statusText = timeData.status === 'running' ? 'Running' : timeData.status === 'paused' ? 'Paused' : 'Stopped';

  return (
    <div className="inline-flex items-center gap-2 sm:gap-5 bg-white/5 px-4 py-2 rounded-md border border-white/10">
      {/* Status Badge */}
      <div className={`inline-flex items-center gap-1 px-3 py-1.5 rounded-md text-xs font-semibold uppercase tracking-wide border ${getStatusBadgeClass(timeData.status)}`}>
        {timeData.clock_running && (
          <span className="w-[6px] h-[6px] rounded-full bg-green-500 animate-pulse"></span>
        )}
        <span>{statusText}</span>
      </div>

      {/* Time Display */}
      <div className="flex flex-col items-center">
        <span className="text-xs text-gray-400 uppercase tracking-wider mb-1 hidden sm:block">GAME TIME</span>
        <span className="text-sm font-bold text-white tabular-nums">{timeData.formatted_time}</span>
      </div>
    </div>
  );
};

export default MiniWorldClock;
