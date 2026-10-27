/**
 * NextFight - Upcoming fight display card (MAT-123)
 *
 * Shows scheduled fight information with countdown and action links.
 * Only renders when boxer has an upcoming scheduled or in-progress fight.
 */

import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { formatFightDate, cancelFight } from '../api/fights';
import './BoxerProfile.css';
import './NextFight.css';

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

  return (
    <div className={`next-fight-section ${isImminent ? 'next-fight-imminent' : ''}`}>
      {/* Section Title */}
      <h3 className="section-title">
        <span className="section-icon">🥊</span>
        Next Fight
        {status === 'in_progress' && (
          <span className="live-badge" style={{ marginLeft: 'auto' }}>● LIVE</span>
        )}
      </h3>

      {/* Fight Info */}
      <div className="next-fight-content">
        <div className="opponent-display">
          <span className="opponent-name">{displayOpponentName}</span>
        </div>

        <div className="fight-details">
          <span className="rounds-badge">{rounds} Rounds</span>
          {countdown && (
            <span className="schedule-countdown">
              ⏰ {countdown}
            </span>
          )}
          {!countdown && scheduledTime && (
            <span className="schedule-date">
              {formatFightDate(scheduledTime)}
            </span>
          )}

          {/* Cancel Button - Only for scheduled fights */}
          {canCancel && (
            <button
              onClick={() => setShowConfirmModal(true)}
              className="cancel-fight-btn"
              disabled={isCancelling}
              aria-label={`Cancel fight against ${opponentName}`}
            >
              {isCancelling ? '⏳ Cancelling...' : '🗑️ Cancel Fight'}
            </button>
          )}

          {/* Action Link */}
          <Link
            to={`/fights/${fightId}`}
            className="view-fight-link"
            aria-label={`View details for fight against ${opponentName}`}
          >
            View Fight Details →
          </Link>
        </div>
      </div>

      {/* Error Message */}
      {cancelError && (
        <div className="cancel-error">
          ⚠️ {cancelError}
        </div>
      )}

      {/* Confirmation Modal */}
      {showConfirmModal && (
        <div className="modal-overlay" onClick={() => setShowConfirmModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>🥊 Cancel Scheduled Fight?</h3>
            <p>Are you sure you want to cancel the scheduled fight against {displayOpponentName}?</p>
            <p className="modal-warning">
              This action cannot be undone. Both boxers will become available for new bookings immediately.
            </p>
            <div className="modal-actions">
              <button
                className="modal-btn modal-btn-secondary"
                onClick={() => setShowConfirmModal(false)}
                disabled={isCancelling}
              >
                Keep Fight
              </button>
              <button
                className="modal-btn modal-btn-danger"
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
