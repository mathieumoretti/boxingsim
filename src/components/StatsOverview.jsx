/**
 * StatsOverview - Combat stats display (MAT-148)
 *
 * Shows Strength, Defense, and Agility with visual bars and comparisons.
 * Migrated to StatDisplay component (MAT-154).
 */

import React from 'react';
import { StatDisplay } from './ui';

const StatsOverview = ({ boxer }) => {
  if (!boxer) return null;

  const strength = Math.round(boxer.strength || 0);
  const defense = Math.round(boxer.defense || 0);
  const agility = Math.round(boxer.agility || 0);

  // Calculate stat averages for comparison (simple heuristic)
  const averageStatForLevel = (boxer.level || 1) * 10;
  const maxPossibleStat = averageStatForLevel * 2; // Estimated max

  return (
    <div className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-6 border border-white/10 mb-6">
      <h3 className="text-lg font-bold text-white m-0 mb-5 flex items-center gap-2">
        <span className="text-xl" aria-hidden="true">💪</span>
        Combat Stats
      </h3>

      <div className="space-y-4">
        <StatDisplay
          label="Strength"
          value={strength}
          icon="💪"
          variant="bar"
          color="strength"
          max={Math.max(100, maxPossibleStat)}
        />

        <StatDisplay
          label="Defense"
          value={defense}
          icon="🛡️"
          variant="bar"
          color="defense"
          max={Math.max(100, maxPossibleStat)}
        />

        <StatDisplay
          label="Agility"
          value={agility}
          icon="⚡"
          variant="bar"
          color="agility"
          max={Math.max(100, maxPossibleStat)}
        />
      </div>

      <div className="mt-5 pt-4 border-t border-white/10">
        <p className="text-sm text-white/60 m-0 italic">Stats determine combat performance. Higher values mean better in-ring capabilities.</p>
      </div>
    </div>
  );
};

export default StatsOverview;
