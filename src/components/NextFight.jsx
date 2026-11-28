/**
 * NextFight - Upcoming fight display card (MAT-123)
 *
 * Shows scheduled fight information with countdown and action links.
 * Only renders when boxer has an upcoming scheduled or in-progress fight.
 */

import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { formatFightDate, cancelFight } from '../api/fights';

const NextFight = ({ upcomingFight, boxerId, gameTime, onRefresh }) => {
  const [isCancelling, setIsCancelling] = useState(false);
  const [cancelError, setCancelError] = useState(null);
  const [showConfirmModal, setShowConfirmModal] = useState(false);

  if (!upcomingFight) {
    return null;
  }

  const { opponent_name: opponentName, opponent_nickname: opponentNickname, scheduled_time: scheduledTime, status, rounds, fight_id: fightId } = upcomingFight;

  // Format the display name as "Name 'Nickname'" if nickname exists
  const displayOpponentName = opponentNickname
    ? `${opponentName} '${opponentNickname}'`
    : opponentName;

  // Use game time for countdown calculation
  const referenceTime = gameTime ? new Date(gameTime) : new Date();
  const fightTime = scheduledTime ? new Date(scheduledTime) : null;

  // Determine urgency based on status and game time
  const isImminent = status === 'in_progress' || (fightTime && fightTime <= new Date(referenceTime.getTime() + 24 * 60 * 60 * 1000));

  // Only show cancel button for scheduled fights that haven't started yet
  const canCancel = status === 'scheduled';

  // Calculate countdown if scheduled in the future
  const countdown = fightTime && fightTime > referenceTime ? formatCountdown(fightTime - referenceTime) : null;

  const handleCancelFight = async () => {
    if (!canCancel) return;

    setIsCancelling(true);
    setCancelError(null);

    try {
      await cancelFight(fightId, { reason: 'User requested cancellation' });
      setShowConfirmModal(false);
      // Refresh data after successful cancellation
      if (onRefresh) {
        onRefresh();
      }
    } catch (error) {
      console.error('Failed to cancel fight:', error);
      setCancelError(error.message || 'Failed to cancel fight');
    } finally {
      setIsCancelling(false);
    }
  };

  // Card base styles with imminent state variation
  const cardBase = 'bg-gradient-to-br from-white/20 to-white/5 backdrop-blur-sm border rounded-xl p-6 mb-6 transition-all duration-200';
  const cardStyles = isImminent
    ? `${cardBase} border-orange-400/60 shadow-[0_4px_16px_rgba(249,115,22,0.3)]`
    : `${cardBase} border-white/10`;

  return (
    <div className={cardStyles}>
      {/* Section Title */}
      <h3 className="flex items-center gap-2 m-0 mb-4 text-sm font-semibold uppercase tracking-wider text-victory-gold">
        <span className="text-lg" aria-hidden="true">🥊</span>
        Next Fight
        {status === 'in_progress' && (
          <span className="ml-auto inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-gradient-to-r from-danger to-danger-darker text-white animate-pulse" style={{ marginLeft: 'auto' }}>● LIVE</span>
        )}
      </h3>

      {/* Fight Info */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 min-w-0">
          <span className="text-lg font-bold text-white truncate block">{displayOpponentName}</span>
        </div>

        <div className="flex flex-col items-end gap-2 flex-shrink-0">
          <div className="flex items-center gap-3">
            <span className="inline-flex items-center px-3 py-1 rounded-md text-xs font-semibold bg-navy-medium border border-white/15 text-white">
              {rounds} Rounds
            </span>
            {countdown && (
              <span className="text-sm text-orange-400 font-medium">⏰ {countdown}</span>
            )}
            {!countdown && scheduledTime && (
              <span className="text-sm text-white/70">
                {formatFightDate(scheduledTime)}
              </span>
            )}
          </div>

          {/* Action Link */}
          <Link
            to={`/fights/${fightId}`}
            className="text-sm font-semibold text-primary hover:text-primary-light transition-colors duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            aria-label={`View details for fight against ${opponentName}`}
          >
            View Fight Details →
          </Link>
        </div>
      </div>

      {/* Cancel Button - Only for scheduled fights */}
      {canCancel && (
        <button
          onClick={() => setShowConfirmModal(true)}
          className={`mt-4 inline-flex items-center gap-2 px-4 py-2 rounded-md text-sm font-semibold border-none cursor-pointer transition-all duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-danger focus-visible:outline-offset-2 ${
            isCancelling
              ? 'bg-white/10 text-white/50 cursor-not-allowed'
              : 'bg-gradient-to-br from-danger to-danger-darker text-white hover:bg-gradient-to-br hover:from-danger-bright hover:to-danger hover:-translate-y-0.5 hover:shadow-[0_4px_8px_rgba(239,68,68,0.3)]'
          }`}
          disabled={isCancelling}
          aria-label={`Cancel fight against ${opponentName}`}
        >
          {isCancelling ? '⏳ Cancelling...' : '🗑️ Cancel Fight'}
        </button>
      )}

      {/* Error Message */}
      {cancelError && (
        <div className="mt-4 p-3 bg-red-500/10 border border-red-400/30 rounded-md text-sm text-red-200" role="alert">
          ⚠️ {cancelError}
        </div>
      )}

      {/* Confirmation Modal */}
      {showConfirmModal && (
        <div
          className="fixed inset-0 z-[9999] flex items-center justify-center p-6 bg-black/70 backdrop-blur-sm animate-[fadeIn_0.2s_ease]"
          onClick={() => setShowConfirmModal(false)}
          role="dialog"
          aria-modal="true"
          aria-labelledby="modal-title"
        >
          <div
            className="bg-white rounded-xl shadow-[0_12px_48px_rgba(0,0,0,0.4)] max-w-md w-full p-6 animate-[slideUp_0.3s_ease]"
            onClick={(e) => e.stopPropagation()}
            role="document"
          >
            <h3 id="modal-title" className="text-xl font-bold text-gray-900 m-0 mb-4">🥊 Cancel Scheduled Fight?</h3>
            <p className="text-gray-700 m-0 mb-2">Are you sure you want to cancel the scheduled fight against {displayOpponentName}?</p>
            <p className="text-sm text-orange-600 font-medium m-0 mb-6 bg-orange-50 p-3 rounded-md border border-orange-200">
              This action cannot be undone. Both boxers will become available for new bookings immediately.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                className="px-5 py-2.5 bg-gray-100 text-gray-700 border border-gray-300 rounded-md text-sm font-semibold cursor-pointer transition-all duration-150 hover:bg-gray-200 hover:-translate-y-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-gray-400 focus-visible:outline-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={() => setShowConfirmModal(false)}
                disabled={isCancelling}
              >
                Keep Fight
              </button>
              <button
                className="px-5 py-2.5 bg-gradient-to-br from-danger to-danger-darker text-white border-none rounded-md text-sm font-semibold cursor-pointer transition-all duration-150 hover:bg-gradient-to-br hover:from-danger-bright hover:to-danger hover:-translate-y-0.5 hover:shadow-[0_4px_12px_rgba(239,68,68,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-danger focus-visible:outline-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={handleCancelFight}
                disabled={isCancelling}
              >
                {isCancelling ? 'Cancelling...' : 'Yes, Cancel Fight'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Format countdown in days/hours/minutes
const formatCountdown = (ms) => {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  return `${minutes}m`;
};

export default NextFight;
