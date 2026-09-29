import React from 'react';
import { useNavigate } from 'react-router-dom';
import './RankCard.css';

const RankCard = ({ boxer, rank }) => {
  const navigate = useNavigate();

  // Get tier badge for top 3 positions
  const getTierBadge = (position) => {
    if (position === 1) return '🥇';
    if (position === 2) return '🥈';
    if (position === 3) return '🥉';
    return `#${position}`;
  };

  // Calculate win rate for display
  const calculateWinRate = () => {
    const totalFights = boxer.wins + boxer.losses;
    if (totalFights === 0) return 'N/A';
    const winRate = (boxer.wins / totalFights) * 100;
    return `${winRate.toFixed(1)}%`;
  };

  // Format the display value based on ranking criteria
  const getDisplayValue = () => {
    // Default to showing win rate and record
    const winRate = calculateWinRate();
    const record = `${boxer.wins}-${boxer.losses}`;
    return { primary: winRate, secondary: record };
  };

  const displayValue = getDisplayValue();
  const tierBadge = getTierBadge(rank);

  // Handle card click to navigate to boxer details
  const handleCardClick = () => {
    // Navigate to a boxer detail page (route to be implemented)
    navigate(`/boxer/${boxer.id}`);
  };

  return (
    <div
      className={`rank-card rank-${rank === 1 ? 'gold' : rank === 2 ? 'silver' : rank === 3 ? 'bronze' : 'regular'}`}
      onClick={handleCardClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && handleCardClick()}
    >
      {/* Position badge */}
      <div className="rank-position">
        <span className="rank-badge">{tierBadge}</span>
      </div>

      {/* Boxer info */}
      <div className="rank-boxer-info">
        <div className="boxer-name-section">
          <span className="boxer-avatar" title={`Avatar for ${boxer.name}`}>
            {String.fromCharCode(0x1F452)}
          </span>
          <div className="boxer-text">
            <span className="boxer-name">{boxer.name}</span>
            {boxer.nickname && (
              <span className="boxer-nickname">"{boxer.nickname}"</span>
            )}
          </div>
        </div>

        {/* Stats */}
        <div className="rank-stats">
          <div className="stat-group stat-level">
            <span className="stat-label">LVL</span>
            <span className="stat-value">{boxer.level}</span>
          </div>

          <div className="stat-group stat-record">
            <span className="stat-label">Record</span>
            <span className="stat-value">{boxer.wins}-{boxer.losses}{boxer.draws > 0 ? `-${boxer.draws}` : ''}</span>
          </div>

          <div className="stat-group stat-winrate">
            <span className="stat-label">Win Rate</span>
            <span className="stat-value stat-value-highlight">{displayValue.primary}</span>
          </div>
        </div>
      </div>

      {/* Arrow indicator for navigation (desktop) */}
      <div className="rank-nav-arrow" onClick={(e) => { e.stopPropagation(); handleCardClick(); }}>→</div>

      {/* View Profile button for mobile */}
      <button className="rank-mobile-btn" onClick={handleCardClick}>
        View Profile →
      </button>
    </div>
  );
};

export default RankCard;
