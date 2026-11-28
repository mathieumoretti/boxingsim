/**
 * FightBadge Component
 *
 * Displays upcoming fight countdown on boxer cards with urgency-based styling.
 * Shows time remaining until scheduled fight with color-coded states.
 */

import React, { useState, useEffect } from 'react';

export function FightBadge({ fight }) {
  const [timeLeft, setTimeLeft] = useState(null);

  useEffect(() => {
    if (!fight?.scheduled_time) return;

    const updateTimeLeft = () => {
      const now = new Date().getTime();
      const scheduled = new Date(fight.scheduled_time).getTime();
      const remaining = scheduled - now;

      if (remaining <= 0) {
        setTimeLeft(0);
      } else {
        setTimeLeft(remaining);
      }
    };

    updateTimeLeft();
    const interval = setInterval(updateTimeLeft, 60000); // Update every minute
    return () => clearInterval(interval);
  }, [fight?.scheduled_time]);

  if (timeLeft === null) return null;

  // Format time display based on remaining duration
  const formatCountdown = (ms) => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 24) {
      const days = Math.floor(hours / 24);
      return `${days}d`;
    } else if (hours > 0) {
      return `${hours}h`;
    } else if (minutes > 0) {
      return `${minutes}m`;
    } else {
      return 'NOW';
    }
  };

  // Urgency-based styling
  const getCountdownClass = () => {
    if (timeLeft === 0) return 'immediate';

    const hours = timeLeft / (1000 * 60 * 60);

    if (hours <= 1) return 'immediate';
    if (hours <= 2) return 'critical';
    if (hours <= 6) return 'urgent';
    return 'normal';
  };

  const urgencyClass = getCountdownClass();

  // Base badge styles with gradient background and shadow
  const badgeBase = 'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-bold shadow-lg backdrop-blur-sm border transform transition-all duration-200';

  // Urgency variants (combat-appropriate colors: amber→orange→red)
  const urgencyStyles = {
    normal: 'bg-gradient-to-r from-amber-600 to-orange-600 text-white border-amber-400/30',
    urgent: 'bg-gradient-to-r from-orange-500 to-orange-600 text-white border-orange-400/40 shadow-[0_2px_12px_rgba(249,115,22,0.5)]',
    critical: 'bg-gradient-to-r from-red-500 to-red-600 text-white border-red-400/50 shadow-[0_2px_16px_rgba(239,68,68,0.6)] animate-pulse',
    immediate: 'bg-gradient-to-r from-danger-bright to-danger text-white border-danger/60 shadow-[0_2px_20px_rgba(239,68,68,0.7)] animate-pulse scale-105',
  };

  return (
    <div
      className={`${badgeBase} ${urgencyStyles[urgencyClass]}`}
      role="status"
      aria-label={`Fight scheduled in ${formatCountdown(timeLeft)}`}
    >
      <span className="text-xs animate-[bounce_1s_infinite]">🥊</span>
      <span>Fight in {formatCountdown(timeLeft)}</span>
    </div>
  );
}

export default FightBadge;
