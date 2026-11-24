import React, { useState, useEffect } from 'react';
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

  const getTrainingBorderClass = (type) => {
    switch (type?.toLowerCase()) {
      case 'strength':
        return 'border-l-[4px] border-l-red-500';
      case 'defense':
        return 'border-l-[4px] border-l-blue-500';
      case 'agility':
        return 'border-l-[4px] border-l-yellow-500';
      case 'recovery':
        return 'border-l-[4px] border-l-teal-500';
      default:
        return 'border-l-[4px] border-l-gray-400';
    }
  };

  const getTypeBadgeClasses = (type) => {
    switch (type?.toLowerCase()) {
      case 'strength':
        return 'bg-gradient-to-r from-red-50 to-red-100 text-red-700';
      case 'defense':
        return 'bg-gradient-to-r from-blue-50 to-blue-100 text-blue-700';
      case 'agility':
        return 'bg-gradient-to-r from-yellow-50 to-yellow-200 text-yellow-800';
      case 'recovery':
        return 'bg-gradient-to-r from-teal-50 to-teal-200 text-teal-800';
      default:
        return 'bg-gray-100 text-gray-700';
    }
  };

  const getGainTagClasses = (type) => {
    switch (type) {
      case 'xp':
        return 'bg-gradient-to-r from-purple-50 to-purple-200 text-purple-800';
      case 'strength':
        return 'bg-gradient-to-r from-red-50 to-red-100 text-red-700';
      case 'defense':
        return 'bg-gradient-to-r from-blue-50 to-blue-100 text-blue-700';
      case 'agility':
        return 'bg-gradient-to-r from-yellow-50 to-yellow-200 text-yellow-800';
      case 'recovery':
        return 'bg-gradient-to-r from-teal-50 to-teal-200 text-teal-800';
      default:
        return 'bg-gray-100 text-gray-700';
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
      <div className="bg-white rounded-xl p-8 text-center shadow-[0_2px_8px_rgba(0,0,0,0.04)]">
        <div className="text-gray-500 text-base font-medium">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-xl p-8 shadow-[0_2px_8px_rgba(0,0,0,0.04)]">
        <ErrorDisplay error={error} variant="alert" title="Failed to Load History" />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="bg-white rounded-xl p-8 text-center shadow-[0_2px_8px_rgba(0,0,0,0.04)]">
        <div className="text-gray-500 text-base font-medium">No training sessions yet.</div>
      </div>
    );
  }

  return (
    <div className="max-w-[700px] mx-auto my-5 mb-10">
      <div className="flex flex-col gap-5">
        {history.map((session) => (
          <div
            key={session.id}
            className={`bg-gradient-to-br from-white to-gray-50 rounded-xl shadow-[0_2px_8px_rgba(0,0,0,0.04)] overflow-hidden transition-all duration-150 hover:shadow-[0_4px_16px_rgba(0,0,0,0.08)] hover:-translate-y-0.5 ${getTrainingBorderClass(session.training_type)}`}
          >
            <div className="flex items-center justify-between p-4 gap-4">
              {/* Type Badge */}
              <div
                className={`inline-flex items-center gap-1 px-3 py-1 rounded-md text-xs font-bold uppercase tracking-wide ${getTypeBadgeClasses(session.training_type)}`}
              >
                <span className="text-base">{getTrainingTypeIcon(session.training_type)}</span>
                <span>{session.training_type || 'Unknown'}</span>
              </div>

              {/* Session Info */}
              <div className="flex-1">
                <div className="text-sm font-bold text-gray-800 mb-1">{session.training_type || 'Unknown'} Training</div>
                <div className="text-xs text-gray-500">{formatDateTime(session.completed_at)}</div>
              </div>

              {/* Gains Display */}
              <div className="flex gap-2 flex-wrap">
                {formatGain(session.strength_gain) && (
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold uppercase tracking-wide ${getGainTagClasses('strength')}`}>
                    STR {formatGain(session.strength_gain)}
                  </span>
                )}
                {formatGain(session.defense_gain) && (
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold uppercase tracking-wide ${getGainTagClasses('defense')}`}>
                    DEF {formatGain(session.defense_gain)}
                  </span>
                )}
                {formatGain(session.agility_gain) && (
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold uppercase tracking-wide ${getGainTagClasses('agility')}`}>
                    AGI {formatGain(session.agility_gain)}
                  </span>
                )}
                {session.xp_gained > 0 && (
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold uppercase tracking-wide ${getGainTagClasses('xp')}`}>
                    +{session.xp_gained} XP
                  </span>
                )}
                <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-gradient-to-r from-gray-100 to-gray-200 rounded text-xs font-semibold text-gray-600">
                  {session.duration_hours}h • {session.energy_cost || 0}E
                </span>
              </div>
            </div>

            {/* Summary Footer */}
            <div className="p-3 bg-black/5 border-t border-gray-200 flex justify-between items-center">
              <span className="text-xs text-gray-500 font-mono">ID: {session.id.slice(-8)}</span>
              <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-gradient-to-r from-green-50 to-green-200 rounded text-xs font-bold uppercase tracking-wide text-green-800">
                ✓ Completed
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default TrainingHistory;
