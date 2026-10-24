import React, { useState, useEffect } from 'react';
import './TrainingHistory.css';
import { fetchTrainingHistory } from '../api/training';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const TrainingHistory = ({ boxerId, compact = false }) => {
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
      <div className={`training-history ${compact ? 'compact' : ''}`}>
        <div className="history-loading">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`training-history ${compact ? 'compact' : ''}`}>
        <ErrorDisplay error={error} variant="alert" title="Failed to Load History" />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className={`training-history ${compact ? 'compact' : ''}`}>
        <div className="history-empty-state">
          No training sessions yet.
        </div>
      </div>
    );
  }

  return (
    <div className={`training-history ${compact ? 'compact' : ''}`}>
      <div className="history-list">
        {history.map((session) => (
          <div key={session.id} className={`history-item ${getTrainingTypeColor(session.training_type)} ${compact ? 'compact' : ''}`}>
            <div className="history-left">
              <span className="history-icon">{getTrainingTypeIcon(session.training_type)}</span>
              <div className="history-content-compact">
                <div className="history-header-compact">
                  <strong>{session.training_type || 'Unknown'} Training</strong>
                  <span className="history-date">{formatDateTime(session.completed_at)}</span>
                </div>
                <div className="history-gains-compact">
                  {formatGain(session.strength_gain) && (
                    <span className="gain-tag strength">STR {formatGain(session.strength_gain)}</span>
                  )}
                  {formatGain(session.defense_gain) && (
                    <span className="gain-tag defense">DEF {formatGain(session.defense_gain)}</span>
                  )}
                  {formatGain(session.agility_gain) && (
                    <span className="gain-tag agility">AGI {formatGain(session.agility_gain)}</span>
                  )}
                  {session.xp_gained > 0 && (
                    <span className="gain-tag xp">+{session.xp_gained} XP</span>
                  )}
                  <span className="history-duration">{session.duration_hours}h • {session.energy_cost || 0}E</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default TrainingHistory;
