import React, { useState } from 'react';
import './TrainingControlPanel.css';
import { API_BASE_URL, authenticatedFetch } from '../utils/auth';

const TrainingControlPanel = ({ boxers, onClose, onRefreshBoxers }) => {
  const [isCompletingAll, setIsCompletingAll] = useState(false);
  const [isCompletingSpecific, setIsCompletingSpecific] = useState({});
  const [result, setResult] = useState(null);

  const completeAllTraining = async () => {
    if (!confirm('This will complete ALL pending training sessions across all boxers. Continue?')) {
      return;
    }

    setIsCompletingAll(true);
    setResult(null);

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/training/bulk-complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({}), // Empty body means complete all
      });

      const data = await response.json();

      if (response.ok) {
        setResult({ success: true, message: data.message, completed: data.completed, failed: data.failed });
        onRefreshBoxers();
      } else {
        setResult({ success: false, message: data.error || 'Failed to complete training sessions' });
      }
    } catch (error) {
      setResult({ success: false, message: `Network error: ${error.message}` });
    } finally {
      setIsCompletingAll(false);
    }
  };

  const completeBoxerTraining = async (boxerId, boxerName) => {
    if (!confirm(`This will complete all pending training sessions for ${boxerName}. Continue?`)) {
      return;
    }

    setIsCompletingSpecific(prev => ({ ...prev, [boxerId]: true }));
    setResult(null);

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/training/bulk-complete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ boxer_id: boxerId }),
      });

      const data = await response.json();

      if (response.ok) {
        setResult({ success: true, message: data.message, completed: data.completed, failed: data.failed });
        onRefreshBoxers();
      } else {
        setResult({ success: false, message: data.error || `Failed to complete training for ${boxerName}` });
      }
    } catch (error) {
      setResult({ success: false, message: `Network error: ${error.message}` });
    } finally {
      setIsCompletingSpecific(prev => ({ ...prev, [boxerId]: false }));
    }
  };

  return (
    <div className="training-control-panel-overlay" onClick={onClose}>
      <div className="training-control-panel" onClick={e => e.stopPropagation()}>
        <button className="panel-close" onClick={onClose} aria-label="Close panel">
          ×
        </button>

        <h2>Training Control Panel</h2>
        <p className="panel-description">
          Development-only feature to manually complete training sessions.
          This bypasses the normal world clock waiting period for testing purposes.
        </p>

        {result && (
          <div className={`result-message ${result.success ? 'success' : 'error'}`}>
            {result.message}
            {result.completed !== undefined && (
              <span className="result-details">
                ({result.completed} completed{result.failed > 0 && `, ${result.failed} failed`})
              </span>
            )}
          </div>
        )}

        <div className="control-section">
          <h3>Global Control</h3>
          <button
            className="complete-all-btn"
            onClick={completeAllTraining}
            disabled={isCompletingAll}
          >
            {isCompletingAll ? 'Completing...' : 'Complete All Training Sessions'}
          </button>
        </div>

        <div className="control-section">
          <h3>Per-Boxer Control</h3>
          <p className="section-description">Complete training for a specific boxer:</p>

          <div className="boxer-controls-list">
            {boxers.length === 0 ? (
              <p className="no-boxers">No boxers found</p>
            ) : (
              boxers.map(boxer => {
                const hasPendingTraining = boxer.pending_training_sessions && boxer.pending_training_sessions.length > 0;
                return (
                  <div key={boxer.id} className="boxer-control-item">
                    <span className="boxer-name">{boxer.name}</span>
                    <button
                      className={`complete-boxer-btn ${hasPendingTraining ? '' : 'disabled'}`}
                      onClick={() => completeBoxerTraining(boxer.id, boxer.name)}
                      disabled={!hasPendingTraining || isCompletingSpecific[boxer.id]}
                      title={!hasPendingTraining ? 'No pending training sessions' : undefined}
                    >
                      {isCompletingSpecific[boxer.id]
                        ? 'Completing...'
                        : hasPendingTraining
                          ? `Complete ${boxer.pending_training_sessions.length} Training Session(s)`
                          : 'No Pending Training'
                      }
                    </button>
                  </div>
                );
              })
            )}
          </div>
        </div>

        <button className="close-panel-btn" onClick={onClose}>
          Close Panel
        </button>
      </div>
    </div>
  );
};

export default TrainingControlPanel;
