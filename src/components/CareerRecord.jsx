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
    <div className="bg-gradient-to-br from-gray-50 to-slate-100 rounded-xl p-6 border border-gray-200 shadow-[0_2px_8px_rgba(0,0,0,0.04)] mb-6">
      <h3 className="text-lg font-bold text-gray-800 m-0 mb-4 flex items-center gap-2">
        <span className="text-xl" aria-hidden="true">📊</span>
        Career Record
      </h3>

      <div className="space-y-5">
        {/* Main Record: W-L-D */}
        <div className="flex items-center justify-center gap-3 py-4 bg-white rounded-lg border border-gray-200 shadow-sm">
          <div className="text-center px-6 py-2">
            <span className="block text-4xl font-bold text-success mb-1">{wins}</span>
            <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">Wins</span>
          </div>
          <div className="text-2xl font-light text-gray-300">-</div>
          <div className="text-center px-6 py-2">
            <span className="block text-4xl font-bold text-danger mb-1">{losses}</span>
            <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">Losses</span>
          </div>
          <div className="text-2xl font-light text-gray-300">-</div>
          <div className="text-center px-6 py-2">
            <span className="block text-4xl font-bold text-gray-600 mb-1">{draws}</span>
            <span className="text-xs font-semibold uppercase tracking-wider text-gray-500">Draws</span>
          </div>
        </div>

        {/* Additional Stats */}
        <div className="flex flex-wrap justify-center gap-3">
          <div className="inline-flex flex-col items-center px-5 py-3 bg-white rounded-lg border border-gray-200 shadow-sm">
            <span className="text-2xl font-bold text-gray-900">{totalFights}</span>
            <span className="text-xs font-semibold text-gray-500 uppercase">Total Fights</span>
          </div>

          <div className="inline-flex flex-col items-center px-5 py-3 bg-gradient-to-br from-success/10 to-emerald-100 rounded-lg border border-success/30 shadow-sm">
            <span className="text-2xl font-bold text-success">{winRate}%</span>
            <span className="text-xs font-semibold text-gray-600 uppercase">Win Rate</span>
          </div>

          {knockouts > 0 && (
            <>
              <div className="inline-flex flex-col items-center px-5 py-3 bg-gradient-to-br from-orange-50 to-red-50 rounded-lg border border-orange-300 shadow-sm">
                <span className="text-2xl font-bold text-orange-600">{koRate}%</span>
                <span className="text-xs font-semibold text-gray-600 uppercase">KO Rate</span>
              </div>

              <div className="inline-flex flex-col items-center px-5 py-3 bg-gradient-to-br from-red-50 to-orange-50 rounded-lg border border-red-300 shadow-sm">
                <span className="text-2xl font-bold text-red-600">{knockouts}</span>
                <span className="text-xs font-semibold text-gray-600 uppercase">Knockouts</span>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default CareerRecord;
