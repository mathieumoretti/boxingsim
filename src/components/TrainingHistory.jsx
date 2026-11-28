import React, { useState, useEffect } from 'react';
import { fetchTrainingHistory } from '../api/training';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

// Friendly display names for the internal training type keys (DB snake_case)
const TRAINING_TYPE_LABELS = {
  normal_training: 'General Training',
  strength_training: 'Strength Training',
  defense_drill: 'Defense Drills',
  agility_workout: 'Agility Workout',
  sparring: 'Sparring',
  cardio: 'Cardio',
};

const getTrainingTypeLabel = (type) => {
  if (!type) return '';
  if (TRAINING_TYPE_LABELS[type]) return TRAINING_TYPE_LABELS[type];
  // Fallback: pretty-print any unknown key (e.g. "speed_work" -> "Speed Work")
  return String(type)
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
};

// Map internal type names (e.g. "strength_training") to the base style key
const normalizeTrainingType = (type) => {
  const key = String(type || '').toLowerCase();
  if (key.startsWith('strength')) return 'strength';
  if (key.startsWith('defense')) return 'defense';
  if (key.startsWith('agility')) return 'agility';
  if (key.startsWith('normal')) return 'general';
  return key;
};

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
    switch (normalizeTrainingType(type)) {
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
    switch (normalizeTrainingType(type)) {
      case 'strength':
        return 'border-red-400/40';
      case 'defense':
        return 'border-blue-400/40';
      case 'agility':
        return 'border-amber-400/40';
      case 'recovery':
        return 'border-teal-400/40';
      default:
        return 'border-white/15';
    }
  };

  const getTypeBadgeClasses = (type) => {
    switch (normalizeTrainingType(type)) {
      case 'strength':
        return 'bg-red-500/10 border border-red-400/30 text-red-200';
      case 'defense':
        return 'bg-blue-500/10 border border-blue-400/30 text-blue-200';
      case 'agility':
        return 'bg-amber-500/10 border border-amber-400/30 text-amber-200';
      case 'recovery':
        return 'bg-teal-500/10 border border-teal-400/30 text-teal-200';
      default:
        return 'bg-white/10 border border-white/15 text-white/80';
    }
  };

  const getGainTagClasses = (type) => {
    switch (type) {
      case 'xp':
        return 'bg-victory-gold/10 border border-victory-gold/40 text-victory-gold';
      case 'strength':
        return 'bg-red-500/10 border border-red-400/30 text-red-200';
      case 'defense':
        return 'bg-blue-500/10 border border-blue-400/30 text-blue-200';
      case 'agility':
        return 'bg-amber-500/10 border border-amber-400/30 text-amber-200';
      case 'recovery':
        return 'bg-teal-500/10 border border-teal-400/30 text-teal-200';
      default:
        return 'bg-white/10 border border-white/15 text-white/70';
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
      <div className="bg-white/5 border border-white/10 rounded-xl p-8 text-center">
        <div className="text-white/60 text-base font-medium">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white/5 border border-white/10 rounded-xl p-8">
        <ErrorDisplay error={error} variant="alert" title="Failed to Load History" />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="bg-white/5 border border-white/10 rounded-xl p-8 text-center">
        <div className="text-white/60 text-base font-medium">No training sessions yet.</div>
      </div>
    );
  }

  return (
    <div className="max-w-[700px] mx-auto my-5 mb-10">
      <div className="flex flex-col gap-5">
        {history.map((session) => (
          <div
            key={session.id}
            className={`bg-white/5 border rounded-xl overflow-hidden transition-all duration-150 hover:bg-white/10 hover:-translate-y-0.5 ${getTrainingBorderClass(session.training_type)}`}
          >
            <div className="flex items-center justify-between p-4 gap-4">
              {/* Type Badge */}
              <div
                className={`inline-flex items-center gap-1 px-3 py-1 rounded-md text-xs font-bold uppercase tracking-wide ${getTypeBadgeClasses(session.training_type)}`}
              >
                <span className="text-base">{getTrainingTypeIcon(session.training_type)}</span>
                <span>{getTrainingTypeLabel(session.training_type) || 'Unknown'}</span>
              </div>

              {/* Session Info */}
              <div className="flex-1">
                <div className="text-sm font-bold text-white mb-1">{getTrainingTypeLabel(session.training_type) || 'Unknown'}</div>
                <div className="text-xs text-white/60">{formatDateTime(session.completed_at)}</div>
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
                <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-white/10 border border-white/15 rounded text-xs font-semibold text-white/70">
                  {session.duration_hours}h • {session.energy_cost || 0}E
                </span>
              </div>
            </div>

            {/* Summary Footer */}
            <div className="p-3 bg-white/5 border-t border-white/10 flex justify-between items-center">
              <span className="text-xs text-white/50 font-mono">ID: </span>
              <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-green-500/10 border border-green-400/30 rounded text-xs font-bold uppercase tracking-wide text-green-200">
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
