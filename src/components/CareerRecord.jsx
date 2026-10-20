/**
 * CareerRecord - Boxer career record display (MAT-148)
 *
 * Shows wins, losses, draws, and knockouts in a prominent format.
 */

import React from 'react';
import './BoxerProfile.css';

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
    <div className="career-record-section">
      <h3 className="section-title">Career Record</h3>

      <div className="record-stats-container">
        {/* Main Record: W-L-D */}
        <div className="main-record">
          <div className="record-item wins">
            <span className="record-number">{wins}</span>
            <span className="record-label">Wins</span>
          </div>
          <div className="record-separator">-</div>
          <div className="record-item losses">
            <span className="record-number">{losses}</span>
            <span className="record-label">Losses</span>
          </div>
          <div className="record-separator">-</div>
          <div className="record-item draws">
            <span className="record-number">{draws}</span>
            <span className="record-label">Draws</span>
          </div>
        </div>

        {/* Additional Stats */}
        <div className="additional-record-stats">
          <div className="stat-pill total-fights">
            <span className="stat-pill-value">{totalFights}</span>
            <span className="stat-pill-label">Total Fights</span>
          </div>

          <div className="stat-pill win-rate">
            <span className="stat-pill-value">{winRate}%</span>
            <span className="stat-pill-label">Win Rate</span>
          </div>

          {knockouts > 0 && (
            <div className="stat-pill ko-rate">
              <span className="stat-pill-value">{koRate}%</span>
              <span className="stat-pill-label">KO Rate</span>
            </div>
          )}

          {knockouts > 0 && (
            <div className="stat-pill total-kos">
              <span className="stat-pill-value">{knockouts}</span>
              <span className="stat-pill-label">Knockouts</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default CareerRecord;
