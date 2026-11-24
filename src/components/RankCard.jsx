/**
 * RankCard Component (MAT-105)
 *
 * Displays a single boxer in the rankings list with tier badges, stats, and navigation.
 */

import React from 'react';
import { useNavigate } from 'react-router-dom';
import { StatDisplay, Badge } from './ui';

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
    const winRate = calculateWinRate();
    const record = `${boxer.wins}-${boxer.losses}`;
    return { primary: winRate, secondary: record };
  };

  const displayValue = getDisplayValue();
  const tierBadge = getTierBadge(rank);

  // Handle card click to navigate to boxer details
  const handleCardClick = () => {
    navigate(`/boxer/${boxer.id}`);
  };

  // Tier-based card styling with gradients and shadows
  const getCardStyles = () => {
    const base = 'flex items-center gap-4 p-5 rounded-xl border transition-all duration-200 cursor-pointer group hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2';

    switch (rank) {
      case 1:
        return `${base} bg-gradient-to-br from-yellow-50 to-orange-50 border-yellow-300 shadow-[0_4px_16px_rgba(234,179,8,0.2),inset_0_1px_0_rgba(255,255,255,0.5)] hover:shadow-[0_8px_24px_rgba(234,179,8,0.3)]`;
      case 2:
        return `${base} bg-gradient-to-br from-gray-50 to-gray-100 border-gray-300 shadow-[0_4px_16px_rgba(175,184,193,0.2),inset_0_1px_0_rgba(255,255,255,0.5)] hover:shadow-[0_8px_24px_rgba(175,184,193,0.3)]`;
      case 3:
        return `${base} bg-gradient-to-br from-orange-50 to-orange-100 border-orange-300 shadow-[0_4px_16px_rgba(213,132,84,0.2),inset_0_1px_0_rgba(255,255,255,0.5)] hover:shadow-[0_8px_24px_rgba(213,132,84,0.3)]`;
      default:
        return `${base} bg-white border-gray-200 shadow-[0_2px_12px_rgba(0,0,0,0.05),inset_0_1px_0_rgba(255,255,255,0.8)] hover:shadow-[0_4px_20px_rgba(0,0,0,0.1)]`;
    }
  };

  // Position badge styling for top 3
  const getPositionBadgeStyles = () => {
    const base = 'flex items-center justify-center rounded-full text-xl font-bold w-10 h-10';

    switch (rank) {
      case 1:
        return `${base} bg-gradient-to-br from-yellow-300 via-warning to-orange-500 text-text-on-gold shadow-[0_2px_8px_rgba(234,179,8,0.4)]`;
      case 2:
        return `${base} bg-gradient-to-br from-gray-300 to-gray-400 text-white shadow-[0_2px_8px_rgba(156,163,175,0.4)]`;
      case 3:
        return `${base} bg-gradient-to-br from-orange-300 to-orange-500 text-white shadow-[0_2px_8px_rgba(194,65,12,0.4)]`;
      default:
        return `${base} bg-gradient-to-br from-navy-medium to-navy-light text-white shadow-[0_2px_8px_rgba(79,70,229,0.3)]`;
    }
  };

  return (
    <div
      className={getCardStyles()}
      onClick={handleCardClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && handleCardClick()}
    >
      {/* Position badge */}
      <div className={getPositionBadgeStyles()}>
        {tierBadge}
      </div>

      {/* Boxer info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-3 mb-2">
          <span
            className="text-2xl flex-shrink-0"
            title={`Avatar for ${boxer.name}`}
            aria-hidden="true"
          >
            👩‍🦰
          </span>
          <div className="min-w-0">
            <span className="text-base font-bold text-gray-900 block truncate">{boxer.name}</span>
            {boxer.nickname && (
              <span className="text-sm text-gray-500 italic block truncate">"{boxer.nickname}"</span>
            )}
          </div>
        </div>

        {/* Stats row */}
        <div className="flex items-center gap-4 flex-wrap">
          <StatDisplay label="LVL" value={boxer.level} variant="mini" color="level" />

          <div className="flex items-baseline gap-1.5 text-sm">
            <span className="text-gray-500 font-medium">Record:</span>
            <span className="font-semibold text-gray-900">
              {boxer.wins}-{boxer.losses}{boxer.draws > 0 ? `-${boxer.draws}` : ''}
            </span>
          </div>

          <div className="flex items-baseline gap-1.5 text-sm">
            <span className="text-gray-500 font-medium">Win Rate:</span>
            <span className="font-bold text-primary">{displayValue.primary}</span>
          </div>
        </div>
      </div>

      {/* Arrow indicator for desktop (visible on hover) */}
      <div
        className="text-2xl text-gray-400 group-hover:text-primary transition-colors duration-200 hidden sm:block"
        onClick={(e) => { e.stopPropagation(); handleCardClick(); }}
        aria-hidden="true"
      >
        →
      </div>

      {/* View Profile button for mobile */}
      <button
        className="sm:hidden text-sm font-semibold text-primary cursor-pointer hover:underline focus-visible:outline-none"
        onClick={handleCardClick}
      >
        View Profile →
      </button>
    </div>
  );
};

export default RankCard;
