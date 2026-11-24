import React, { useState, useEffect } from 'react';

/**
 * RestingBadge component displays a boxer's rest status with countdown timer.
 * Shows different styling for voluntary rest (teal) vs forced rest (red/orange).
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
      return `${hours}h ${minutes}m`;
    } else {
      return `${minutes}m`;
    }
  };

  const countdownDisplay = formatRemainingTime(remainingSeconds);

  // Badge variants - use dynamic classes for status
  const getBadgeClasses = () => {
    if (isForcedRest) {
      return 'bg-gradient-to-r from-red-500 to-red-600 shadow-[0_2px_8px_rgba(239,68,68,0.4)] animate-pulse hover:from-red-400 hover:to-red-500';
    } else {
      return 'bg-gradient-to-r from-teal-500 to-teal-700 shadow-[0_2px_6px_rgba(13,148,136,0.35)] hover:from-teal-400 hover:to-teal-600';
    }
  };

  return (
    <div
      className={`absolute top-2 right-2 px-3 py-1 rounded-full text-[9px] font-bold uppercase tracking-wide text-white backdrop-blur inline-flex items-center gap-1 transition-all duration-150 hover:-translate-y-0.5 ${getBadgeClasses()}`}
      title={isForcedRest ? 'Forced rest due to injury/exhaustion' : 'Voluntary rest for recovery'}
    >
      <span className="text-xs" aria-hidden="true">{isForcedRest ? '🚑' : '🏥'}</span>
      <span className="hidden sm:inline">{isForcedRest ? 'Forced Rest' : 'Resting'}</span>
      <span className="sm:hidden">{isForcedRest ? 'REST' : 'RST'}</span>
      {countdownDisplay && (
        <span
          className="inline-flex items-center px-2 py-0.5 rounded text-[8px] font-semibold text-white/95 bg-black/15"
          title="Time remaining until available"
        >
          {countdownDisplay}
        </span>
      )}
    </div>
  );
};

export default RestingBadge;
