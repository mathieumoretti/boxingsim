import React from 'react';
import { Link } from 'react-router-dom';
import './FightBadge.css';

const FightBadge = ({ fight, currentGameTime }) => {
  if (!fight) return null;

  const { opponent_name: opponentName, scheduled_time: scheduledTime, status, rounds } = fight;

  // Don't show badge for fights scheduled in the past relative to game time
  // These should be either completed or cancelled (MAT-FIGHT-SCHEDULING)
  if (!scheduledTime || !currentGameTime) return null;

  const fightTime = new Date(scheduledTime);
  const gameTime = new Date(currentGameTime);

  // Hide badge for past-dated scheduled fights (data cleanup needed)
  if (fightTime < gameTime) {
    return null;
  }

  // Calculate remaining time until fight
  const calculateRemainingTime = () => {
    const diffMs = fightTime - gameTime;
    const diffSeconds = Math.max(0, Math.ceil(diffMs / 1000));

    if (diffSeconds <= 0) return 'Starting!';

    const hours = Math.floor(diffSeconds / 3600);
    const minutes = Math.floor((diffSeconds % 3600) / 60);
    const seconds = diffSeconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    } else {
      return `${seconds}s`;
    }
  };

  // Determine countdown color based on time remaining (adjusted for faster testing)
  const getCountdownColorClass = () => {
    const diffMs = fightTime - gameTime;
    const diffHours = diffMs / (1000 * 60 * 60);

    if (diffHours <= 0) return 'countdown-immediate';
    if (diffHours < 0.5) return 'countdown-critical'; // Under 30 mins
    if (diffHours < 2) return 'countdown-urgent'; // Under 2 hours
    return 'countdown-normal';
  };

  // Determine if fight is within pulse window for animation
  const isWithinPulseWindow = () => {
    const diffHours = (fightTime - gameTime) / (1000 * 60 * 60);

    return diffHours > 0 && diffHours < 2;
  };

  const countdownText = calculateRemainingTime();
  const countdownColorClass = getCountdownColorClass();
  const pulseClass = isWithinPulseWindow() ? 'pulse-animation' : '';

  // Format date for tooltip
  const formatDateTooltip = () => {
    const options = { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', hour12: true };
    return fightTime.toLocaleDateString('en-US', options);
  };

  return (
    <Link to={`/fight/${fight.fight_id}`} className={`fight-badge ${pulseClass}`}>
      <span className="fight-icon" title={`Upcoming fight: ${rounds} rounds`}>🥊</span>
      <span className="fight-info">vs {opponentName}</span>
      <span className={`fight-timer ${countdownColorClass}`} title={formatDateTooltip()}>
        ⏱️ {countdownText}
      </span>
    </Link>
  );
};

export default FightBadge;
