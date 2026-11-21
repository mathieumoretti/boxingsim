/**
 * RecoveryBadge - Displays recovery status for boxers on forced recovery (MAT-153)
 *
 * Two variants: compact inline badge or full banner with progress bar.
 * Only renders when boxer is actually on forced recovery.
 */

import React from 'react';
import {
  isOnForcedRecovery,
  getRecoveryStatus,
} from '../utils/recoveryUtils';

/**
 * RecoveryBadge Component
 * @param {Object} props
 * @param {Object} props.boxer - Boxer data object with forced_recovery_until field
 * @param {boolean} [props.compact=false] - If true, show as small badge instead of full banner
 */
const RecoveryBadge = ({ boxer, compact = false }) => {
  if (!isOnForcedRecovery(boxer)) {
    return null; // Don't render if not on recovery
  }

  const status = getRecoveryStatus(boxer);

  if (compact) {
    // Compact badge version - shows inline with boxer name/info
    return (
      <span
        className="recovery-badge recovery-badge-compact"
        title={`On forced recovery until ${status.recovery_until?.toLocaleString()}`}
      >
        <span className="recovery-badge-icon">🩹</span>
        <span className="recovery-badge-text">Recovering: {status.time_remaining}</span>
      </span>
    );
  }

  // Full banner version - for prominent display (e.g., on boxer profile)
  return (
    <div className="recovery-badge recovery-banner" role="alert">
      <div className="recovery-banner-content">
        <div className="recovery-banner-icon">⚠️</div>
        <div className="recovery-banner-details">
          <h4 className="recovery-banner-title">Forced Recovery Active</h4>
          <p className="recovery-banner-description">
            This boxer is on mandatory recovery from fight injuries.
            They cannot fight until their health has been restored.
          </p>
          <div className="recovery-banner-stats">
            <span className="recovery-banner-time">
              <strong>Time Remaining:</strong> {status.time_remaining}
            </span>
            {status.percentage_healed !== undefined && (
              <span className="recovery-banner-progress">
                <strong>Recovery Progress:</strong> {Math.round(status.percentage_healed)}%
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Optional progress bar */}
      {status.percentage_healed !== undefined && (
        <div className="recovery-banner-progress-bar">
          <div
            className="recovery-banner-progress-fill"
            style={{ width: `${Math.min(status.percentage_healed, 100)}%` }}
          />
        </div>
      )}
    </div>
  );
};

export default RecoveryBadge;
