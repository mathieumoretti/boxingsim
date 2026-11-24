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
    <div className="bg-gradient-to-br from-gray-50 to-slate-100 rounded-xl p-6 border border-gray-200 shadow-[0_2px_8px_rgba(0,0,0,0.04)] mb-6">
      <h3 className="text-lg font-bold text-gray-800 m-0 mb-5 flex items-center gap-2">
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

      <div className="mt-5 pt-4 border-t border-gray-200">
        <p className="text-sm text-gray-600 m-0 italic">Stats determine combat performance. Higher values mean better in-ring capabilities.</p>
      </div>
    </div>
  );
};

export default StatsOverview;
