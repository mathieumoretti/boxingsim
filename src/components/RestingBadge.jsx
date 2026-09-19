import React, { useState, useEffect } from 'react';
import './RestingBadge.css';

/**
 * RestingBadge component displays a boxer's rest status with countdown timer.
 * Shows different styling for voluntary rest (yellow/orange) vs forced rest (red).
 *
 * @param {boolean} hasActiveRest - Whether the boxer has an active rest period
 * @param {string|null} restEndsAt - ISO timestamp when rest ends (for voluntary rest)
 * @param {string|null} forcedRestUntil - ISO timestamp when forced rest ends
 * @param {string|null} currentGameTime - Current game world time in ISO format
 */
const RestingBadge = ({ hasActiveRest, restEndsAt, forcedRestUntil, currentGameTime }) => {
  const [remainingSeconds, setRemainingSeconds] = useState(null);

  if (!hasActiveRest) return null;

  // Determine if this is forced rest or voluntary rest (MAT-96)
  const isForcedRest = !!forcedRestUntil;

  // Calculate remaining rest time
  // Forced rest uses real-time (stored as time.Now()), voluntary rest uses game time (stored as gameTime.Add())
  const calculateRemainingSeconds = () => {
    if (isForcedRest && forcedRestUntil) {
      // Forced rest: use real-time for countdown
      const endTime = new Date(forcedRestUntil);
      const now = new Date();
      const remainingMs = endTime - now;
      return Math.max(0, Math.ceil(remainingMs / 1000));
    } else if (restEndsAt && currentGameTime) {
      // Voluntary rest: use game time for countdown
      const endTime = new Date(restEndsAt);
      const gameTime = new Date(currentGameTime);
      const remainingMs = endTime - gameTime;
      return Math.max(0, Math.ceil(remainingMs / 1000));
    }
    return null;
  };

  // Initialize countdown
  useEffect(() => {
    setRemainingSeconds(calculateRemainingSeconds());
  }, [restEndsAt, forcedRestUntil, currentGameTime, isForcedRest]);

  // Update countdown every second for forced rest (real-time) or when game time updates
  useEffect(() => {
    if (!hasActiveRest) return;

    const interval = setInterval(() => {
      setRemainingSeconds(calculateRemainingSeconds());
    }, 1000);

    return () => clearInterval(interval);
  }, [hasActiveRest, restEndsAt, forcedRestUntil, currentGameTime, isForcedRest]);

  // Format countdown display
  const formatRemainingTime = (seconds) => {
    if (seconds === null || seconds <= 0) return '';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}h ${minutes}m remaining`;
    } else {
      return `${minutes}m remaining`;
    }
  };

  const countdownDisplay = formatRemainingTime(remainingSeconds);

  return (
    <div className={`resting-badge ${isForcedRest ? 'resting-forced' : 'resting-voluntary'}`}>
      <span className="resting-icon">{isForcedRest ? '🚑' : '🏥'}</span>
      <span className="resting-status">{isForcedRest ? 'Forced Rest' : 'Resting'}</span>
      {countdownDisplay && (
        <span className="resting-timer" title="Time remaining until available">
          {countdownDisplay}
        </span>
      )}
    </div>
  );
};

export default RestingBadge;
