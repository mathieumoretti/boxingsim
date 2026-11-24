import React, { useState, useEffect } from 'react';
import { API_BASE_URL, authenticatedFetch } from '../utils/auth';

const WorldClock = () => {
  const [clockData, setClockData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [refreshTimer, setRefreshTimer] = useState(0);

  useEffect(() => {
    loadWorldTime();
  }, []);

  const loadWorldTime = async () => {
    setIsLoading(true);
    setError('');

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/world/time`, {
        method: 'GET',
      });

      if (response.ok) {
        const data = await response.json();
        setClockData(data);

        // Set up refresh timer based on seconds_per_game_hour
        // Refresh every 1 second to show live updates
        setRefreshTimer(60 - new Date().getSeconds());
      } else {
        setError(data.error || 'Failed to load world time');
      }
    } catch (err) {
      if (!err.message.includes('Unauthorized')) {
        setError('Network error: ' + err.message);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleRefresh = () => {
    loadWorldTime();
  };

  const formatRemainingTime = (secondsPerGameHour) => {
    if (!clockData || !clockData.clock_running) return '';

    // Calculate seconds until next game hour based on real time
    const now = new Date();
    const secondsIntoCurrentMinute = now.getSeconds();
    const remainingInMinute = 60 - secondsIntoCurrentMinute;

    if (secondsPerGameHour <= 60) {
      return `${remainingInMinute}s`;
    }

    const minutesUntilNextHour = Math.ceil(secondsPerGameHour / 60);
    return `${minutesUntilNextHour}m`;
  };

  // Dynamic status indicator classes
  const getStatusClasses = (status) => {
    switch(status) {
      case 'running':
        return 'bg-green-500/15 text-green-500 border-green-500/30';
      case 'paused':
        return 'bg-yellow-500/15 text-yellow-500 border-yellow-500/30';
      default:
        return 'bg-gray-600/15 text-gray-600 border-gray-600/30';
    }
  };

  const statusText = clockData?.status === 'running' ? 'Clock Running' :
                     clockData?.status === 'paused' ? 'Paused' : 'Stopped';

  if (isLoading) {
    return (
      <div className="bg-gradient-to-br from-navy-dark to-navy-medium rounded-2xl p-5 shadow-[0_4px_20px_rgba(0,0,0,0.3)] min-w-[320px] max-w-[400px]">
        <div className="text-gray-500 text-center p-5 text-sm">Loading game time...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-gradient-to-br from-navy-dark to-navy-medium rounded-2xl p-5 shadow-[0_4px_20px_rgba(0,0,0,0.3)] min-w-[320px] max-w-[400px]">
        <div className="text-red-400 bg-red-500/10 p-3 rounded-md text-center">
          {error}
        </div>
      </div>
    );
  }

  if (!clockData) {
    return null;
  }

  const nextHourTimer = formatRemainingTime(clockData.seconds_per_game_hour);

  return (
    <div className="bg-gradient-to-br from-navy-dark to-navy-medium rounded-2xl p-5 shadow-[0_4px_20px_rgba(0,0,0,0.3)] min-w-[320px] max-w-[400px]">
      <div className="flex flex-col gap-4 relative">
        {/* Refresh button */}
        <button
          onClick={handleRefresh}
          className="absolute top-3 right-3 w-8 h-8 rounded-full bg-white/10 border-none text-gray-500 text-base cursor-pointer transition-all duration-200 flex items-center justify-center hover:bg-white/20 hover:text-white hover:rotate-180 active:scale-95 active:rotate-180 focus-visible:outline-2 focus-visible:outline-yellow-400 focus-visible:outline-offset-2"
          aria-label="Refresh game time"
        >
          ↻
        </button>

        {/* Main time display */}
        <div className="text-center">
          <div className="text-xs text-gray-500 uppercase tracking-widest mb-2">Current Game Time</div>
          <div className="text-[1.4rem] font-semibold text-white leading-tight">{clockData.formatted_time}</div>
        </div>

        {/* Status indicator */}
        <div
          className={`inline-flex items-center justify-center gap-2 p-2 px-4 rounded-full text-sm font-medium w-fit mx-auto border ${getStatusClasses(clockData.status)}`}
        >
          {clockData.clock_running && (
            <span className="w-[8px] h-[8px] rounded-full bg-green-500 animate-pulse"></span>
          )}
          <span>{statusText}</span>
        </div>

        {/* Additional info */}
        <div className="flex justify-between pt-3 border-t border-white/10">
          <div className="flex flex-col gap-1">
            <span className="text-xs text-gray-600 uppercase tracking-wide">Elapsed Time</span>
            <span className="text-sm font-semibold text-gray-200">{clockData.time_since_start}</span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-gray-600 uppercase tracking-wide">Speed</span>
            <span className="text-sm font-semibold text-gray-200">{clockData.speed_factor}x</span>
          </div>
          {nextHourTimer && (
            <div className="flex flex-col gap-1">
              <span className="text-xs text-gray-600 uppercase tracking-wide">Next Hour</span>
              <span className="text-sm font-semibold text-gray-200">{nextHourTimer}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default WorldClock;
