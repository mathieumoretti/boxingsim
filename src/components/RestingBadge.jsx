import React from 'react';
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
  if (!hasActiveRest) return null;

  // Calculate remaining rest time
  const calculateRemainingSeconds = () => {
    // Use forcedRestUntil if available (forced rest takes priority), otherwise use restEndsAt
    const restEndTime = forcedRestUntil || restEndsAt;

    if (!restEndTime || !currentGameTime) return null;

    const endTime = new Date(restEndTime);
    const gameTime = new Date(currentGameTime);
    const remainingMs = endTime - gameTime;
    const remainingSeconds = Math.max(0, Math.ceil(remainingMs / 1000));

    return remainingSeconds;
  };

  const remainingSeconds = calculateRemainingSeconds();

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

  // Determine if this is forced rest or voluntary rest
  const isForcedRest = !!forcedRestUntil;

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
