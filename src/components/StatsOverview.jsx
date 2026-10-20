/**
 * StatsOverview - Combat stats display (MAT-148)
 *
 * Shows Strength, Defense, and Agility with visual bars and comparisons.
 */

import React from 'react';
import './BoxerProfile.css';

const StatsOverview = ({ boxer }) => {
  if (!boxer) return null;

  const strength = Math.round(boxer.strength || 0);
  const defense = Math.round(boxer.defense || 0);
  const agility = Math.round(boxer.agility || 0);

  // Calculate stat averages for comparison (simple heuristic)
  const averageStatForLevel = (boxer.level || 1) * 10;
  const maxPossibleStat = averageStatForLevel * 2; // Estimated max

  const StatBar = ({ label, value, icon, colorClass }) => {
    const percent = Math.min(100, (value / maxPossibleStat) * 100);
    const isAboveAverage = value >= averageStatForLevel;

    return (
      <div className="stat-row">
        <div className="stat-label-container">
          <span className="stat-icon">{icon}</span>
          <span className="stat-name">{label}</span>
        </div>
        <div className="stat-bar-container-full">
          <div className={`stat-bar-fill ${colorClass}`}>
            <div style={{ width: `${percent}%` }}></div>
          </div>
        </div>
        <div className="stat-value-container">
          <span className="stat-number">{value}</span>
          {isAboveAverage && (
            <span className="stat-badge positive">Excellent</span>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="stats-overview-section">
      <h3 className="section-title">Combat Stats</h3>

      <div className="stats-grid">
        <StatBar
          label="Strength"
          value={strength}
          icon="💪"
          colorClass="stat-strength"
        />

        <StatBar
          label="Defense"
          value={defense}
          icon="🛡️"
          colorClass="stat-defense"
        />

        <StatBar
          label="Agility"
          value={agility}
          icon="⚡"
          colorClass="stat-agility"
        />
      </div>

      <div className="stats-description">
        <p>Stats determine combat performance. Higher values mean better in-ring capabilities.</p>
      </div>
    </div>
  );
};

export default StatsOverview;
