import React, { useState, useEffect } from 'react';
import './WorldClock.css';
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

  if (isLoading) {
    return (
      <div className="world-clock">
        <div className="world-clock-loading">Loading game time...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="world-clock">
        <div className="world-clock-error">{error}</div>
      </div>
    );
  }

  if (!clockData) {
    return null;
  }

  const statusColor = clockData.status === 'running' ? 'status-running' :
                      clockData.status === 'paused' ? 'status-paused' :
                      'status-stopped';

  const nextHourTimer = formatRemainingTime(clockData.seconds_per_game_hour);

  return (
    <div className="world-clock">
      <div className="world-clock-content">
        {/* Main time display */}
        <div className="time-display">
          <div className="time-label">Current Game Time</div>
          <div className="main-time">{clockData.formatted_time}</div>
        </div>

        {/* Status indicator */}
        <div className={`status-indicator ${statusColor}`}>
          {clockData.clock_running ? (
            <span className="status-dot"></span>
          ) : null}
          <span className="status-text">
            {clockData.status === 'running' ? 'Clock Running' :
             clockData.status === 'paused' ? 'Paused' : 'Stopped'}
          </span>
        </div>

        {/* Additional info */}
        <div className="clock-details">
          <div className="detail-item">
            <span className="detail-label">Elapsed Time</span>
            <span className="detail-value">{clockData.time_since_start}</span>
          </div>
          <div className="detail-item">
            <span className="detail-label">Speed</span>
            <span className="detail-value">{clockData.speed_factor}x</span>
          </div>
          {nextHourTimer && (
            <div className="detail-item">
              <span className="detail-label">Next Hour In</span>
              <span className="detail-value">{nextHourTimer}</span>
            </div>
          )}
        </div>

        {/* Refresh button */}
        <button onClick={handleRefresh} className="refresh-btn" aria-label="Refresh game time">
          ↻
        </button>
      </div>
    </div>
  );
};

export default WorldClock;
