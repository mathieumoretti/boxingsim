import React, { useState, useEffect } from 'react';
import './TrainingHistory.css';
import { fetchTrainingHistory } from '../api/training';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const TrainingHistory = ({ boxerId }) => {
  const [history, setHistory] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadTrainingHistory();
  }, [boxerId]);

  const loadTrainingHistory = async () => {
    if (!boxerId) return;

    setIsLoading(true);
    setError(null);

    try {
      const data = await fetchTrainingHistory(boxerId);
      setHistory(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(getUserErrorMessage(err, 'training', { context: 'history' }));
    } finally {
      setIsLoading(false);
    }
  };

  const getTrainingTypeIcon = (type) => {
    switch (type?.toLowerCase()) {
      case 'strength':
        return '💪';
      case 'defense':
        return '🛡️';
      case 'agility':
        return '⚡';
      case 'recovery':
        return '🏥';
      default:
        return '📚';
    }
  };

  const getTrainingTypeColor = (type) => {
    switch (type?.toLowerCase()) {
      case 'strength':
        return 'training-strength';
      case 'defense':
        return 'training-defense';
      case 'agility':
        return 'training-agility';
      case 'recovery':
        return 'training-recovery';
      default:
        return 'training-default';
    }
  };

  const formatDateTime = (isoString) => {
    if (!isoString) return 'Unknown date';

    const date = new Date(isoString);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return `Today at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
    } else if (diffDays === 1) {
      return `Yesterday at ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
    } else if (diffDays < 7) {
      return `${diffDays} days ago`;
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' });
    }
  };

  const formatGain = (value) => {
    if (value === null || value === undefined || value === 0) return null;
    return `+${value > 0.01 ? value.toFixed(2) : '<0.01'}`;
  };

  if (isLoading) {
    return (
      <div className="training-history">
        <div className="history-loading">Loading training history...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="training-history">
        <ErrorDisplay error={error} variant="alert" title="Failed to Load History" />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="training-history">
        <div className="history-empty-state">
          <div className="empty-icon">📝</div>
          <h4>No Training History Yet</h4>
          <p>This boxer hasn't completed any training sessions yet. Schedule their first session above!</p>
        </div>
      </div>
    );
  }

  return (
    <div className="training-history">
      <div className="history-list">
        {history.map((session) => (
          <div key={session.id} className={`history-item ${getTrainingTypeColor(session.training_type)}`}>
            <div className="history-icon">
              {getTrainingTypeIcon(session.training_type)}
            </div>

            <div className="history-content">
              <div className="history-header">
                <h4>{session.training_type || 'Unknown'} Training</h4>
                <span className="history-date">{formatDateTime(session.completed_at)}</span>
              </div>

              <div className="history-details">
                <span className="detail-item">
                  <strong>Duration:</strong> {session.duration_hours} hour{session.duration_hours !== 1 ? 's' : ''}
                </span>
                <span className="detail-item">
                  <strong>Energy Cost:</strong> {session.energy_cost || 0}
                </span>
              </div>

              {/* Stat Gains */}
              <div className="history-gains">
                {formatGain(session.strength_gain) && (
                  <span className="gain-item strength">
                    💪 STR {formatGain(session.strength_gain)}
                  </span>
                )}
                {formatGain(session.defense_gain) && (
                  <span className="gain-item defense">
                    🛡️ DEF {formatGain(session.defense_gain)}
                  </span>
                )}
                {formatGain(session.agility_gain) && (
                  <span className="gain-item agility">
                    ⚡ AGI {formatGain(session.agility_gain)}
                  </span>
                )}
                {session.xp_gained > 0 && (
                  <span className="gain-item xp">
                    📊 +{session.xp_gained} XP
                  </span>
                )}
              </div>
            </div>

            <div className="history-chevron">
              →
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default TrainingHistory;
