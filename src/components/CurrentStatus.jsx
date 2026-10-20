/**
 * CurrentStatus - Boxer current status display (MAT-148)
 *
 * Shows Health, Energy, and Fatigue with color-coded bars and recovery info.
 */

import React from 'react';
import './BoxerProfile.css';

const CurrentStatus = ({ boxer, worldTime }) => {
  if (!boxer) return null;

  const health = Math.round(boxer.health || 0);
  const energy = Math.round(boxer.energy || 0);
  const fatigue = Math.round(boxer.fatigue_score || 0);

  // Get health color based on percentage
  const getHealthColorClass = (healthValue) => {
    if (healthValue > 50) return 'bg-health-high';
    if (healthValue > 25) return 'bg-health-medium';
    return 'bg-health-low';
  };

  // Get fatigue color based on exhaustion threshold
  const getFatigueColorClass = (fatigueValue) => {
    if (fatigueValue >= 80) return 'bg-fatigue-critical';   // Red - exhausted
    if (fatigueValue >= 50) return 'bg-fatigue-warning';    // Yellow - warning
    return 'bg-fatigue-ok';                                   // Green - healthy
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
      <div className="status-row">
        <div className="status-label-container">
          <span className="status-icon">{icon}</span>
          <span className="status-name">{label}</span>
          {tooltip && <span className="status-tooltiptext" title={tooltip}></span>}
        </div>
        <div className="status-bar-container-full">
          <div className={`status-bar-fill ${colorClass}`}>
            <div style={{ width: `${percent}%` }}></div>
          </div>
        </div>
        <div className="status-value-container">
          <span className="status-number">{value}<small>/ {max}</small></span>
        </div>
      </div>
    );
  };

  return (
    <div className="current-status-section">
      <div className="section-header-row">
        <h3 className="section-title">Current Status</h3>
        {boxer.has_active_rest && (
          <div className={`rest-indicator ${isForcedRest ? 'forced' : ''}`}>
            {isForcedRest ? '🚑' : '🏥'}
            <span>In Recovery</span>
            {recoveryTime && <span className="recovery-countdown">({recoveryTime} remaining)</span>}
          </div>
        )}
      </div>

      <div className="status-grid">
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
          colorClass="bg-energy"
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
      <div className="training-eligibility">
        {fatigue >= 80 ? (
          <span className="eligibility-ineligible">❌ Exhausted - Rest Required</span>
        ) : energy < 15 ? (
          <span className="eligibility-ineligible">❌ Low Energy - Rest Needed</span>
        ) : health < 50 ? (
          <span className="eligibility-warning">⚠️ Injured - Recovery Recommended</span>
        ) : (
          <span className="eligibility-ready">✅ Ready for Training</span>
        )}
      </div>
    </div>
  );
};

export default CurrentStatus;
