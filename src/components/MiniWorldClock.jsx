import React, { useState, useEffect } from 'react';
import './MiniWorldClock.css';
import { API_BASE_URL, authenticatedFetch } from '../utils/auth';

const MiniWorldClock = () => {
  const [timeData, setTimeData] = useState(null);

  useEffect(() => {
    const loadTime = async () => {
      try {
        const response = await authenticatedFetch(`${API_BASE_URL}/world/time`, { method: 'GET' });
        if (response.ok) {
          const data = await response.json();
          setTimeData(data);

          // Refresh every second to show live countdown
          const interval = setInterval(async () => {
            try {
              const resp = await authenticatedFetch(`${API_BASE_URL}/world/time`, { method: 'GET' });
              if (resp.ok) {
                const data = await resp.json();
                setTimeData(data);
              }
            } catch (err) {
              // Silently fail - time display will just be stale
            }
          }, 1000);

          return () => clearInterval(interval);
        }
      } catch (err) {
        // Silently fail - component will show loading state
      }
    };

    loadTime();
  }, []);

  if (!timeData) {
    return <div className="mini-world-clock mini-world-clock-loading">Loading...</div>;
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
