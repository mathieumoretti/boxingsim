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
        className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-semibold bg-gradient-to-r from-red-100 to-orange-100 border border-red-400 text-red-900 shadow-[0_2px_8px_rgba(239,68,68,0.15)]"
        title={`On forced recovery until ${status.recovery_until?.toLocaleString()}`}
      >
        <span className="text-base">🩹</span>
        <span>Recovering: {status.time_remaining}</span>
      </span>
    );
  }

  // Full banner version - for prominent display (e.g., on boxer profile)
  return (
    <div className="flex flex-col gap-3 p-4 bg-gradient-to-br from-amber-100 to-yellow-100 border-2 border-amber-500 rounded-lg mb-6 shadow-[0_4px_16px_rgba(245,158,11,0.15)]" role="alert">
      <div className="flex gap-4 items-start">
        <div className="text-4xl flex-shrink-0">⚠️</div>
        <div className="flex-1">
          <h4 className="m-0 mb-2 text-lg font-bold text-amber-900">Forced Recovery Active</h4>
          <p className="m-0 mb-3 text-base leading-relaxed text-amber-800">
            This boxer is on mandatory recovery from fight injuries.
            They cannot fight until their health has been restored.
          </p>
          <div className="flex flex-wrap gap-6 text-base text-amber-800">
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
        <div className="w-full h-2 bg-amber-100 rounded-full overflow-hidden">
          <div
            className="h-full bg-gradient-to-r from-amber-500 to-amber-600 transition-all duration-300"
            style={{ width: `${Math.min(status.percentage_healed, 100)}%` }}
          />
        </div>
      )}
    </div>
  );
};

export default RecoveryBadge;
