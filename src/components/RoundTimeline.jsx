import React, { useState } from 'react';
import './FightDetails.css';

/**
 * RoundTimeline - Horizontal scrollable timeline of fight rounds (MAT-104)
 *
 * Features:
 * - Clickable round markers showing round numbers
 * - Color coding for neutral rounds and victory/defeat outcomes
 * - Final round highlighted with victory indicator
 */
const RoundTimeline = ({ totalRounds, method }) => {
  const [selectedRound, setSelectedRound] = useState(null);

  // Determine if this was a knockout
  const isKO = method === 'knockout';

  // Get round status based on position relative to final round
  const getRoundStatus = (roundNum) => {
    if (roundNum > totalRounds) return 'not-reached';
    if (roundNum === totalRounds) return 'final';
    return 'completed';
  };

  // Get CSS class for round marker based on status
  const getRoundClass = (roundNum) => {
    const baseClass = 'round-marker';
    const status = getRoundStatus(roundNum);

    let classes = `${baseClass} ${status}`;

    if (selectedRound === roundNum) {
      classes += ' selected';
    }

    return classes;
  };

  // Get display icon for round marker
  const getRoundIcon = (roundNum) => {
    const status = getRoundStatus(roundNum);

    switch (status) {
      case 'final':
        if (isKO) return '🔥';
        return '🏆';
      case 'not-reached':
        return '';
      default:
        return '';
    }
  };

  // Handle round click
  const handleRoundClick = (roundNum) => {
    if (roundNum <= totalRounds) {
      setSelectedRound(selectedRound === roundNum ? null : roundNum);
    }
  };

  // Generate round markers
  const maxRounds = 12; // Standard fight rounds
  const rounds = Array.from({ length: maxRounds }, (_, i) => i + 1);

  return (
    <div className="round-timeline-container">
      <div className="round-timeline">
        {rounds.map((roundNum) => {
          const status = getRoundStatus(roundNum);

          // Don't render rounds that weren't reached
          if (status === 'not-reached') return null;

          return (
            <div
              key={roundNum}
              className={getRoundClass(roundNum)}
              onClick={() => handleRoundClick(roundNum)}
              title={`Round ${roundNum}${status === 'final' ? ` - ${isKO ? 'Knockout!' : 'Final Round!'}` : ''}`}
            >
              {getRoundIcon(roundNum) && (
                <span className="round-icon">{getRoundIcon(roundNum)}</span>
              )}
              <span className="round-number">{roundNum}</span>
              {status === 'final' && (
                <span className="final-indicator">
                  {isKO ? 'KO' : 'END'}
                </span>
              )}
            </div>
          );
        })}
      </div>

      {/* Round legend */}
      <div className="round-legend">
        <div className="legend-item">
          <span className="legend-marker completed"></span>
          <span>Rounds fought</span>
        </div>
        {isKO && (
          <div className="legend-item">
            <span className="legend-marker final ko"></span>
            <span>Knockout</span>
          </div>
        )}
        {!isKO && (
          <div className="legend-item">
            <span className="legend-marker final decision"></span>
            <span>Decision</span>
          </div>
        )}
      </div>

      {/* Selected round info */}
      {selectedRound && selectedRound <= totalRounds && (
        <div className="round-details-popup">
          <h4>Round {selectedRound}</h4>
          {selectedRound === totalRounds ? (
            <p className="final-round-text">
              {isKO ? '🔥 Knockout victory! The fight ended here.' : '🏆 Final round - decision made.'}
            </p>
          ) : (
            <p className="mid-round-text">
              Round {selectedRound} of {totalRounds} completed.
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default RoundTimeline;
