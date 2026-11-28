/**
 * RoundTimeline - Horizontal scrollable timeline of fight rounds (MAT-104)
 *
 * Features:
 * - Clickable round markers showing round numbers
 * - Color coding for neutral rounds and victory/defeat outcomes
 * - Final round highlighted with victory indicator
 */

import React, { useState } from 'react';

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

  // Get Tailwind classes for round marker based on status
  const getRoundClass = (roundNum) => {
    const status = getRoundStatus(roundNum);

    // Base marker styles
    let classes = 'inline-flex flex-col items-center justify-center rounded-lg transition-all duration-200 cursor-pointer select-none';

    if (status === 'completed') {
      classes += ' bg-gradient-to-br from-navy-medium to-navy-light border-2 border-white/15 text-white hover:bg-gradient-to-br hover:from-navy-light hover:to-navy-dark hover:border-primary-light hover:shadow-[0_4px_12px_rgba(124,58,237,0.3)]';
    } else if (status === 'final') {
      if (isKO) {
        classes += ' bg-gradient-to-br from-danger-bright to-danger border-2 border-warning/60 text-white shadow-[0_4px_16px_rgba(255,107,107,0.5)] scale-110';
      } else {
        classes += ' bg-gradient-to-br from-victory-gold to-warning border-2 border-yellow-300/60 text-black shadow-[0_4px_16px_rgba(214,158,46,0.5)] scale-110';
      }
    }

    if (selectedRound === roundNum) {
      classes += ' ring-2 ring-white ring-offset-2 ring-offset-navy-deep';
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
    <div className="w-full">
      {/* Timeline */}
      <div className="flex gap-3 overflow-x-auto pb-4 px-1 scrollbar-thin scrollbar-thumb-navy-medium scrollbar-track-transparent">
        {rounds.map((roundNum) => {
          const status = getRoundStatus(roundNum);

          // Don't render rounds that weren't reached
          if (status === 'not-reached') return null;

          return (
            <div
              key={roundNum}
              className={`${getRoundClass(roundNum)} w-[56px] h-[72px]`}
              onClick={() => handleRoundClick(roundNum)}
              title={`Round ${roundNum}${status === 'final' ? ` - ${isKO ? 'Knockout!' : 'Final Round!'}` : ''}`}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => e.key === 'Enter' && handleRoundClick(roundNum)}
            >
              {getRoundIcon(roundNum) && (
                <span className="text-[14px] mb-1">{getRoundIcon(roundNum)}</span>
              )}
              <span className="text-lg font-extrabold leading-none">{roundNum}</span>
              {status === 'final' && (
                <span className={`text-[10px] font-bold mt-1 ${isKO ? 'text-danger-bright' : 'text-white/80'}`}>
                  {isKO ? 'KO' : 'END'}
                </span>
              )}
            </div>
          );
        })}
      </div>

      {/* Round legend */}
      <div className="flex flex-wrap items-center gap-4 mt-3 px-1">
        <div className="flex items-center gap-2">
          <span className="w-4 h-4 rounded-md bg-gradient-to-br from-navy-medium to-navy-light border border-white/15"></span>
          <span className="text-sm text-white/70">Rounds fought</span>
        </div>
        {isKO && (
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-md bg-gradient-to-br from-danger-bright to-danger border border-warning/60"></span>
            <span className="text-sm text-white/70">Knockout</span>
          </div>
        )}
        {!isKO && (
          <div className="flex items-center gap-2">
            <span className="w-4 h-4 rounded-md bg-gradient-to-br from-victory-gold to-warning border border-yellow-300/60"></span>
            <span className="text-sm text-white/70">Decision</span>
          </div>
        )}
      </div>

      {/* Selected round info popup */}
      {selectedRound && selectedRound <= totalRounds && (
        <div className="mt-4 p-5 bg-gradient-to-br from-navy-medium to-navy-light border border-white/20 rounded-lg text-center shadow-[0_8px_24px_rgba(0,0,0,0.3)] animate-[slideUp_0.3s_ease]">
          <h4 className="text-white font-bold text-lg m-0 mb-2">Round {selectedRound}</h4>
          {selectedRound === totalRounds ? (
            <p className={`m-0 font-semibold ${isKO ? 'text-danger-bright' : 'text-victory-gold'}`}>
              {isKO ? '🔥 Knockout victory! The fight ended here.' : '🏆 Final round - decision made.'}
            </p>
          ) : (
            <p className="m-0 text-white/80">
              Round {selectedRound} of {totalRounds} completed.
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default RoundTimeline;
