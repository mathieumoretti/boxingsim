import React, { useState } from 'react';
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

  const getResultClasses = (success) => {
    if (success) {
      return 'bg-green-500/20 border-green-500/50 text-green-300';
    } else {
      return 'bg-red-500/20 border-red-500/50 text-red-300';
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black/70 flex items-center justify-center z-[2000] p-8 animate-fadeIn"
      onClick={onClose}
    >
      <div
        className="bg-gradient-to-br from-navy-medium to-navy-light rounded-xl p-6 max-w-[600px] w-full max-h-[80vh] overflow-y-auto shadow-[0_10px_40px_rgba(0,0,0,0.5)] relative text-white"
        onClick={e => e.stopPropagation()}
      >
        {/* Close Button */}
        <button
          className="absolute top-3 right-3 bg-white/10 border-none text-white text-xl w-9 h-9 rounded-md cursor-pointer flex items-center justify-center transition-all duration-150 hover:bg-white/20 focus-visible:outline-2 focus-visible:outline-yellow-400 focus-visible:outline-offset-2"
          onClick={onClose}
          aria-label="Close panel"
        >
          ×
        </button>

        <h2 className="m-0 mb-2 text-xl font-bold text-white">Training Control Panel</h2>
        <p className="m-0 mb-5 text-sm text-white/80 leading-relaxed">
          Development-only feature to manually complete training sessions.
          This bypasses the normal world clock waiting period for testing purposes.
        </p>

        {/* Result Messages */}
        {result && (
          <div
            className={`p-3 rounded-md mb-5 text-sm font-medium border ${getResultClasses(result.success)}`}
          >
            {result.message}
            {result.completed !== undefined && (
              <span className="block text-xs opacity-90 mt-1">
                ({result.completed} completed{result.failed > 0 && `, ${result.failed} failed`})
              </span>
            )}
          </div>
        )}

        {/* Global Control Section */}
        <div className="bg-white/5 rounded-md p-4 mb-4">
          <h3 className="m-0 mb-3 text-base font-semibold text-white">Global Control</h3>
          <button
            className="w-full bg-yellow-400 text-navy-dark border-none p-3 rounded-md cursor-pointer text-sm font-semibold transition-all duration-150 hover:bg-yellow-500 hover:-translate-y-0.5 shadow-[0_4px_12px_rgba(250,204,21,0.3)] disabled:opacity-60 disabled:cursor-not-allowed focus-visible:outline-2 focus-visible:outline-yellow-400 focus-visible:outline-offset-2"
            onClick={completeAllTraining}
            disabled={isCompletingAll}
          >
            {isCompletingAll ? 'Completing...' : 'Complete All Training Sessions'}
          </button>
        </div>

        {/* Per-Boxer Control Section */}
        <div className="bg-white/5 rounded-md p-4 mb-2">
          <h3 className="m-0 mb-3 text-base font-semibold text-white">Per-Boxer Control</h3>
          <p className="m-0 mb-3 text-xs text-white/70">Complete training for a specific boxer:</p>

          <div className="flex flex-col gap-2 max-h-[300px] overflow-y-auto">
            {boxers.length === 0 ? (
              <p className="text-sm text-white/50 text-center p-5">No boxers found</p>
            ) : (
              boxers.map(boxer => {
                const hasPendingTraining = boxer.pending_training_sessions && boxer.pending_training_sessions.length > 0;
                return (
                  <div
                    key={boxer.id}
                    className="flex items-center justify-between p-3 bg-white/8 rounded-md gap-3"
                  >
                    <span className="text-sm font-medium text-white flex-1 overflow-hidden text-ellipsis whitespace-nowrap">
                      {boxer.name}
                    </span>
                    <button
                      className={`border-none p-2 rounded-md cursor-pointer text-xs font-semibold transition-all duration-150 hover:-translate-y-0.5 whitespace-nowrap flex-shrink-0 focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 ${
                        hasPendingTraining
                          ? 'bg-primary text-white hover:bg-primary-dark'
                          : 'bg-white/20 text-white/50 cursor-not-allowed'
                      }`}
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

        {/* Close Panel Button */}
        <button
          className="w-full bg-transparent text-white/70 border border-white/30 p-3 rounded-md cursor-pointer text-sm font-medium transition-all duration-150 mt-2 hover:bg-white/10 hover:border-white/50 hover:text-white focus-visible:outline-2 focus-visible:outline-yellow-400 focus-visible:outline-offset-2"
          onClick={onClose}
        >
          Close Panel
        </button>
      </div>
    </div>
  );
};

export default TrainingControlPanel;
