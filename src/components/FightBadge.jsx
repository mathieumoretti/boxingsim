import React from 'react';
import { Link } from 'react-router-dom';
import './FightBadge.css';

const FightBadge = ({ fight, currentGameTime }) => {
  if (!fight) return null;

  const { opponent_name: opponentName, scheduled_time: scheduledTime, status, rounds } = fight;

  // Calculate remaining time until fight
  const calculateRemainingTime = () => {
    if (!scheduledTime || !currentGameTime) return '';

    const fightTime = new Date(scheduledTime);
    const gameTime = new Date(currentGameTime);
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

  // Determine countdown color based on time remaining
  const getCountdownColorClass = () => {
    if (!scheduledTime || !currentGameTime) return '';

    const fightTime = new Date(scheduledTime);
    const gameTime = new Date(currentGameTime);
    const diffMs = fightTime - gameTime;
    const diffHours = diffMs / (1000 * 60 * 60);

    if (diffHours <= 0) return 'countdown-immediate';
    if (diffHours < 1) return 'countdown-critical';
    if (diffHours < 24) return 'countdown-urgent';
    return 'countdown-normal';
  };

  // Determine if fight is within 24 hours for pulse animation
  const isWithin24Hours = () => {
    if (!scheduledTime || !currentGameTime) return false;

    const fightTime = new Date(scheduledTime);
    const gameTime = new Date(currentGameTime);
    const diffHours = (fightTime - gameTime) / (1000 * 60 * 60);

    return diffHours > 0 && diffHours < 24;
  };

  const countdownText = calculateRemainingTime();
  const countdownColorClass = getCountdownColorClass();
  const pulseClass = isWithin24Hours() ? 'pulse-animation' : '';

  // Format date for tooltip
  const formatDateTooltip = () => {
    if (!scheduledTime) return '';
    const d = new Date(scheduledTime);
    const options = { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', hour12: true };
    return d.toLocaleDateString('en-US', options);
  };

  return (
    <Link to={`/fights/${fight.fight_id}`} className={`fight-badge ${pulseClass}`}>
      <span className="fight-icon" title={`Upcoming fight: ${rounds} rounds`}>🥊</span>
      <span className="fight-info">vs {opponentName}</span>
      <span className={`fight-timer ${countdownColorClass}`} title={formatDateTooltip()}>
        ⏱️ {countdownText}
      </span>
    </Link>
  );
};

export default FightBadge;
