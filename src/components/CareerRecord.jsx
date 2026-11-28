/**
 * CareerRecord - Boxer career record display (MAT-148)
 *
 * Shows wins, losses, draws, and knockouts in a prominent format.
 */

import React from 'react';

const CareerRecord = ({ boxer }) => {
  if (!boxer) return null;

  const wins = boxer.wins || 0;
  const losses = boxer.losses || 0;
  const draws = boxer.draws || 0;
  const knockouts = boxer.knockouts || 0;

  const totalFights = wins + losses + draws;
  const winRate = totalFights > 0 ? Math.round((wins / totalFights) * 100) : 0;
  const koRate = wins > 0 ? Math.round((knockouts / wins) * 100) : 0;

  return (
    <div className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-6 border border-white/10 mb-6">
      <h3 className="text-lg font-bold text-white m-0 mb-4 flex items-center gap-2">
        <span className="text-xl" aria-hidden="true">📊</span>
        Career Record
      </h3>

      <div className="space-y-5">
        {/* Main Record: W-L-D */}
        <div className="flex items-center justify-center gap-3 py-4 bg-white/5 rounded-lg border border-white/10">
          <div className="text-center px-6 py-2">
            <span className="block text-4xl font-bold text-success-light mb-1">{wins}</span>
            <span className="text-xs font-semibold uppercase tracking-wider text-white/60">Wins</span>
          </div>
          <div className="text-2xl font-light text-white/30">-</div>
          <div className="text-center px-6 py-2">
            <span className="block text-4xl font-bold text-danger-bright mb-1">{losses}</span>
            <span className="text-xs font-semibold uppercase tracking-wider text-white/60">Losses</span>
          </div>
          <div className="text-2xl font-light text-white/30">-</div>
          <div className="text-center px-6 py-2">
            <span className="block text-4xl font-bold text-white/70 mb-1">{draws}</span>
            <span className="text-xs font-semibold uppercase tracking-wider text-white/60">Draws</span>
          </div>
        </div>

        {/* Additional Stats */}
        <div className="flex flex-wrap justify-center gap-3">
          <div className="inline-flex flex-col items-center px-5 py-3 bg-white/5 rounded-lg border border-white/10">
            <span className="text-2xl font-bold text-white">{totalFights}</span>
            <span className="text-xs font-semibold text-white/60 uppercase">Total Fights</span>
          </div>

          <div className="inline-flex flex-col items-center px-5 py-3 bg-success/10 rounded-lg border border-success/40">
            <span className="text-2xl font-bold text-success-light">{winRate}%</span>
            <span className="text-xs font-semibold text-white/60 uppercase">Win Rate</span>
          </div>

          {knockouts > 0 && (
            <>
              <div className="inline-flex flex-col items-center px-5 py-3 bg-orange-500/10 rounded-lg border border-orange-400/40">
                <span className="text-2xl font-bold text-orange-400">{koRate}%</span>
                <span className="text-xs font-semibold text-white/60 uppercase">KO Rate</span>
              </div>

              <div className="inline-flex flex-col items-center px-5 py-3 bg-red-500/10 rounded-lg border border-red-400/40">
                <span className="text-2xl font-bold text-red-400">{knockouts}</span>
                <span className="text-xs font-semibold text-white/60 uppercase">Knockouts</span>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default CareerRecord;
