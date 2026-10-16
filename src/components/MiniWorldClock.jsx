import React, { useState, useEffect } from 'react';
import './MiniWorldClock.css';
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
    return <div className="mini-world-clock mini-world-clock-loading">Loading...</div>;
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

  const statusClass = timeData.status === 'running' ? 'status-running' :
                      timeData.status === 'paused' ? 'status-paused' : 'status-stopped';

  return (
    <div className="mini-world-clock">
      <div className={`clock-status ${statusClass}`}>
        {timeData.clock_running && <span className="status-dot"></span>}
        <span className="status-text">
          {timeData.status === 'running' ? 'Running' : timeData.status === 'paused' ? 'Paused' : 'Stopped'}
        </span>
      </div>
      <div className="game-time-display">
        <span className="time-label">GAME TIME</span>
        <span className="time-value">{timeData.formatted_time}</span>
      </div>
    </div>
  );
};

export default MiniWorldClock;
