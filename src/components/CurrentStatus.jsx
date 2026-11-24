/**
 * CurrentStatus - Boxer current status display (MAT-148)
 *
 * Shows Health, Energy, and Fatigue with color-coded bars and recovery info.
 */

import React from 'react';

const CurrentStatus = ({ boxer, worldTime }) => {
  if (!boxer) return null;

  const health = Math.round(boxer.health || 0);
  const energy = Math.round(boxer.energy || 0);
  const fatigue = Math.round(boxer.fatigue_score || 0);

  // Get health color based on percentage
  const getHealthColorClass = (healthValue) => {
    if (healthValue > 50) return 'bg-gradient-to-r from-green-400 to-green-500';
    if (healthValue > 25) return 'bg-gradient-to-r from-yellow-400 to-orange-400';
    return 'bg-gradient-to-r from-red-400 to-red-600';
  };

  // Get fatigue color based on exhaustion threshold
  const getFatigueColorClass = (fatigueValue) => {
    if (fatigueValue >= 80) return 'bg-gradient-to-r from-red-500 to-red-600 animate-[healthCritical_2s_ease]';   // Red - exhausted
    if (fatigueValue >= 50) return 'bg-gradient-to-r from-orange-400 to-warning';    // Yellow - warning
    return 'bg-gradient-to-r from-green-400 to-emerald-500';                         // Green - healthy
  };

  // Calculate recovery time for energy (simple approximation)
  const calculateRecoveryTime = () => {
    if (!boxer.has_active_rest) return null;

    if (worldTime && boxer.next_available_training) {
      const now = new Date(worldTime.current_game_time);
      const then = new Date(boxer.next_available_training);
      const diffMs = then - now;

      if (diffMs <= 0) return null; // Already recovered

      const hours = Math.floor(diffMs / (1000 * 60 * 60));
      const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

      if (hours > 0) {
        return `${hours}h ${minutes}m`;
      } else if (minutes > 0) {
        return `${minutes}m`;
      }
    }
    return null;
  };

  const recoveryTime = calculateRecoveryTime();
  const isForcedRest = boxer.forced_rest_until !== null;

  const StatusRow = ({ label, value, max, colorClass, icon, tooltip }) => {
    const percent = (value / max) * 100;

    return (
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          <span className="text-sm flex-shrink-0" aria-hidden="true">{icon}</span>
          <span className="text-sm font-semibold text-gray-700">{label}</span>
          {tooltip && <span className="sr-only" title={tooltip}>{tooltip}</span>}
        </div>
        <div className="flex items-center gap-3">
          <div className="flex-1 h-3 bg-gray-200 rounded-full overflow-hidden border border-gray-300">
            <div className={`h-full ${colorClass} transition-all duration-500`}>
              <div style={{ width: `${percent}%` }}></div>
            </div>
          </div>
          <span className="text-sm font-bold text-gray-800 min-w-[3.5rem] text-right">{value}<small className="text-gray-500 font-normal">/ {max}</small></span>
        </div>
      </div>
    );
  };

  return (
    <div className="bg-gradient-to-br from-gray-50 to-slate-100 rounded-xl p-6 border border-gray-200 shadow-[0_2px_8px_rgba(0,0,0,0.04)] mb-6">
      <div className="flex items-center justify-between mb-5 pb-4 border-b border-gray-200">
        <h3 className="text-lg font-bold text-gray-800 m-0 flex items-center gap-2">
          <span className="text-xl" aria-hidden="true">📈</span>
          Current Status
        </h3>
        {boxer.has_active_rest && (
          <div className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-bold ${isForcedRest ? 'bg-red-100 text-red-800 border-2 border-red-400' : 'bg-blue-100 text-blue-800 border border-blue-300'}`}>
            <span>{isForcedRest ? '🚑' : '🏥'}</span>
            <span>In Recovery</span>
            {recoveryTime && <span className="text-gray-600">({recoveryTime} remaining)</span>}
          </div>
        )}
      </div>

      <div className="space-y-4">
        <StatusRow
          label="Health"
          value={health}
          max={100}
          colorClass={getHealthColorClass(health)}
          icon="❤️"
          tooltip={health < 50 ? "Boxer is injured and needs rest" : health < 75 ? "Boxer has minor injuries" : "Boxer is in good health"}
        />

        <StatusRow
          label="Energy"
          value={energy}
          max={100}
          colorClass="bg-gradient-to-r from-blue-400 to-indigo-500"
          icon="⚡"
          tooltip={energy < 15 ? "Insufficient energy for training (need 15+)" : energy < 30 ? "Low energy - rest recommended" : "Energy level is good"}
        />

        <StatusRow
          label="Fatigue"
          value={fatigue}
          max={100}
          colorClass={getFatigueColorClass(fatigue)}
          icon="😴"
          tooltip={fatigue >= 80 ? "Boxer is exhausted and cannot train" : fatigue >= 50 ? "Boxer is approaching exhaustion" : "Fatigue level is manageable"}
        />
      </div>

      {/* Training eligibility indicator */}
      <div className="mt-5 pt-4 border-t border-gray-200">
        {fatigue >= 80 ? (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-red-100 border-2 border-red-400 rounded-lg text-sm font-bold text-red-800">
            ❌ Exhausted - Rest Required
          </span>
        ) : energy < 15 ? (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-orange-100 border-2 border-orange-400 rounded-lg text-sm font-bold text-orange-800">
            ❌ Low Energy - Rest Needed
          </span>
        ) : health < 50 ? (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-yellow-100 border-2 border-yellow-400 rounded-lg text-sm font-bold text-yellow-800">
            ⚠️ Injured - Recovery Recommended
          </span>
        ) : (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-green-100 border-2 border-green-400 rounded-lg text-sm font-bold text-green-800">
            ✅ Ready for Training
          </span>
        )}
      </div>
    </div>
  );
};

export default CurrentStatus;
