/**
 * StatsOverview - Combat stats display (MAT-148)
 *
 * Shows Strength, Defense, and Agility with visual bars and comparisons.
 * Migrated to StatDisplay component (MAT-154).
 */

import React from 'react';
import { StatDisplay } from './ui';
import './BoxerProfile.css';

const StatsOverview = ({ boxer }) => {
  if (!boxer) return null;

  const strength = Math.round(boxer.strength || 0);
  const defense = Math.round(boxer.defense || 0);
  const agility = Math.round(boxer.agility || 0);

  // Calculate stat averages for comparison (simple heuristic)
  const averageStatForLevel = (boxer.level || 1) * 10;
  const maxPossibleStat = averageStatForLevel * 2; // Estimated max

  return (
    <div className="stats-overview-section">
      <h3 className="section-title">Combat Stats</h3>

      <div className="stats-grid">
        <StatDisplay
          label="Strength"
          value={strength}
          icon="💪"
          variant="bar"
          color="strength"
          maxValue={maxPossibleStat}
        />

        <StatDisplay
          label="Defense"
          value={defense}
          icon="🛡️"
          variant="bar"
          color="defense"
          maxValue={maxPossibleStat}
        />

        <StatDisplay
          label="Agility"
          value={agility}
          icon="⚡"
          variant="bar"
          color="agility"
          maxValue={maxPossibleStat}
        />
      </div>

      <div className="stats-description">
        <p>Stats determine combat performance. Higher values mean better in-ring capabilities.</p>
      </div>
    </div>
  );
};

export default StatsOverview;
