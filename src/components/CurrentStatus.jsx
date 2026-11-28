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
    const percent = max > 0 ? Math.min(100, Math.max(0, (value / max) * 100)) : 0;

    return (
      <div className="space-y-1.5">
        <div className="flex items-center gap-2">
          <span className="text-sm flex-shrink-0" aria-hidden="true">{icon}</span>
          <span className="text-sm font-semibold text-white/80">{label}</span>
          {tooltip && <span className="sr-only" title={tooltip}>{tooltip}</span>}
        </div>
        <div className="flex items-center gap-3">
          <div className="flex-1 h-3 bg-white/15 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full ${colorClass} transition-all duration-500`}
              style={{ width: `${percent}%` }}
            />
          </div>
          <span className="text-sm font-bold text-white min-w-[3.5rem] text-right">{value}<small className="text-white/50 font-normal">/ {max}</small></span>
        </div>
      </div>
    );
  };

  return (
    <div className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-6 border border-white/10 mb-6">
      <div className="flex items-center justify-between mb-5 pb-4 border-b border-white/10">
        <h3 className="text-lg font-bold text-white m-0 flex items-center gap-2">
          <span className="text-xl" aria-hidden="true">📈</span>
          Current Status
        </h3>
        {boxer.has_active_rest && (
          <div className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-bold ${isForcedRest ? 'bg-red-500/15 text-red-200 border border-red-400/50' : 'bg-blue-500/15 text-blue-200 border border-blue-400/50'}`}>
            <span>{isForcedRest ? '🚑' : '🏥'}</span>
            <span>In Recovery</span>
            {recoveryTime && <span className="text-white/60">({recoveryTime} remaining)</span>}
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
      <div className="mt-5 pt-4 border-t border-white/10">
        {fatigue >= 80 ? (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-red-500/10 border border-red-400/40 rounded-lg text-sm font-bold text-red-200">
            ❌ Exhausted - Rest Required
          </span>
        ) : energy < 15 ? (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-orange-500/10 border border-orange-400/40 rounded-lg text-sm font-bold text-orange-200">
            ❌ Low Energy - Rest Needed
          </span>
        ) : health < 50 ? (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-yellow-500/10 border border-yellow-400/40 rounded-lg text-sm font-bold text-yellow-200">
            ⚠️ Injured - Recovery Recommended
          </span>
        ) : (
          <span className="inline-flex items-center gap-2 px-4 py-2 bg-green-500/10 border border-green-400/40 rounded-lg text-sm font-bold text-green-200">
            ✅ Ready for Training
          </span>
        )}
      </div>
    </div>
  );
};

export default CurrentStatus;
