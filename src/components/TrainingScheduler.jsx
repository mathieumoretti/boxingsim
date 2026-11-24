import React, { useState, useEffect } from 'react';
import { fetchTrainingTypes, startTrainingSession } from '../api/training';
import { fetchWorldTime } from '../api/world';
import { ErrorDisplay } from './ErrorDisplay';
import { Alert } from './ui';
import { getUserErrorMessage } from '../utils/errorMessages';

const TrainingScheduler = ({ boxerId, boxer, onClose, onTrainingScheduled, compact = false }) => {
  const [trainingTypes, setTrainingTypes] = useState([]);
  const [selectedType, setSelectedType] = useState('');
  const [durationHours, setDurationHours] = useState(2);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [worldTime, setWorldTime] = useState(null);

  useEffect(() => {
    loadTrainingTypes();
  }, []);

  // Fetch world time for rest countdown (only when boxer has active rest)
  useEffect(() => {
    const loadWorldTimeData = async () => {
      try {
        const data = await fetchWorldTime();
        setWorldTime(data);

        // Set up periodic refresh for countdown updates
        const interval = setInterval(async () => {
          try {
            const data = await fetchWorldTime();
            setWorldTime(data);
          } catch (err) {
            // Silently fail - countdown will just be stale
          }
        }, 5000);

        return () => clearInterval(interval);
      } catch (err) {
        console.warn('Failed to load world time:', err);
      }
    };

    // Load world time if boxer has active rest period
    if (boxer?.has_active_rest) {
      loadWorldTimeData();
    }
  }, [boxer?.has_active_rest]);

  const loadTrainingTypes = async () => {
    try {
      const data = await fetchTrainingTypes();
      setTrainingTypes(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(getUserErrorMessage(err, 'general'));
    } finally {
      setIsLoading(false);
    }
  };

  const calculateEnergyCost = () => {
    if (!selectedType) return 0;
    const type = trainingTypes.find(t => t.id === parseInt(selectedType));
    return type ? Math.round(type.energy_cost * durationHours) : 0;
  };

  const calculateEffectiveGains = () => {
    if (!selectedType || !boxer) return {
      strength: 0, defense: 0, agility: 0, xp: 0,
      fatigueMultiplier: 1,
      diminishingReturns: { strength: 1, defense: 1, agility: 1 }
    };

    const type = trainingTypes.find(t => t.id === parseInt(selectedType));
    if (!type) return {
      strength: 0, defense: 0, agility: 0, xp: 0,
      fatigueMultiplier: 1,
      diminishingReturns: { strength: 1, defense: 1, agility: 1 }
    };

    // Base gains
    const baseStrength = type.strength_gain_factor * durationHours;
    const baseDefense = type.defense_gain_factor * durationHours;
    const baseAgility = type.agility_gain_factor * durationHours;

    // Diminishing returns multiplier per stat: 1 / (1 + current_stat / 100)
    const diminishingSTR = 1 / (1 + (boxer.strength || 0) / 100);
    const diminishingDEF = 1 / (1 + (boxer.defense || 0) / 100);
    const diminishingAGI = 1 / (1 + (boxer.agility || 0) / 100);

    // Fatigue effectiveness multiplier: max(0.5, 1 - fatigue_score / 200)
    const fatigueScore = boxer.fatigue_score || 0;
    const fatigueMultiplier = Math.max(0.5, 1 - fatigueScore / 200);

    // Effective gains with all modifiers
    const effSTR = baseStrength * diminishingSTR * fatigueMultiplier;
    const effDEF = baseDefense * diminishingDEF * fatigueMultiplier;
    const effAGI = baseAgility * diminishingAGI * fatigueMultiplier;

    // XP calculation: (eff_str + eff_def + eff_agi) × duration_hours × 10
    const totalEffectiveGains = effSTR + effDEF + effAGI;
    const xpGained = totalEffectiveGains * durationHours * 10;

    return {
      strength: Math.round(effSTR * 100) / 100,
      defense: Math.round(effDEF * 100) / 100,
      agility: Math.round(effAGI * 100) / 100,
      xp: Math.round(xpGained),
      fatigueMultiplier: Math.round(fatigueMultiplier * 100) / 100,
      diminishingReturns: {
        strength: Math.round(diminishingSTR * 100) / 100,
        defense: Math.round(diminishingDEF * 100) / 100,
        agility: Math.round(diminishingAGI * 100) / 100,
      }
    };
  };

  const getSelectedTrainingType = () => {
    return trainingTypes.find(t => t.id === parseInt(selectedType));
  };

  // Calculate remaining seconds until rest ends (same logic as RestingBadge)
  const calculateRemainingRestSeconds = () => {
    if (!boxer?.has_active_rest || !worldTime?.current_game_time) return null;

    // Use forcedRestUntil if available (forced rest takes priority), otherwise use next_available_training
    const restEndTime = boxer.forced_rest_until || boxer.next_available_training;

    if (!restEndTime) return null;

    const endTime = new Date(restEndTime);
    const gameTime = new Date(worldTime.current_game_time);
    const remainingMs = endTime - gameTime;
    const remainingSeconds = Math.max(0, Math.ceil(remainingMs / 1000));

    return remainingSeconds;
  };

  // Format countdown display (same logic as RestingBadge)
  const formatRestCountdown = (seconds) => {
    if (seconds === null || seconds <= 0) return '';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}h ${minutes}m remaining`;
    } else {
      return `${minutes}m remaining`;
    }
  };

  // Determine rest type for display
  const isForcedRest = boxer?.forced_rest_until !== null && boxer?.forced_rest_until !== undefined;

  const remainingRestSeconds = calculateRemainingRestSeconds();
  const restCountdownDisplay = formatRestCountdown(remainingRestSeconds);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setIsSubmitting(true);

    try {
      console.log('TrainingScheduler: Submitting with boxerId:', boxerId, 'type of boxerId:', typeof boxerId);
      const data = await startTrainingSession(boxerId, {
        training_type_id: parseInt(selectedType),
        duration_hours: durationHours,
        scheduled_at: new Date().toISOString()
      });

      setSuccess('Training session scheduled successfully!');
      setError('');
      // Reset form after success
      setSelectedType('');
      setDurationHours(2);
      // Call the refresh callback to reload boxer list in parent component
      if (onTrainingScheduled) {
        onTrainingScheduled();
      }
      // Close modal after a brief delay to show success message
      setTimeout(() => {
        setSuccess('');
        if (onClose) onClose();
      }, 1500);
    } catch (err) {
      // Use specific error message from utility with boxer name context
      setError(getUserErrorMessage(err, 'training', { boxerName: boxer?.name }));
    } finally {
      setIsSubmitting(false);
    }
  };

  const energyCost = calculateEnergyCost();
  const effectiveGains = calculateEffectiveGains();
  const trainingType = getSelectedTrainingType();

  // Dynamic classes for rest warning
  const getRestWarningClasses = () => {
    if (isForcedRest) {
      return 'bg-gradient-to-r from-red-50 to-orange-100 border-b border-red-200';
    } else {
      return 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-b border-yellow-200';
    }
  };

  const getEnergyStatusClasses = () => {
    if (energyCost > boxer?.energy) {
      return 'text-red-600';
    } else if (boxer?.energy - energyCost < 30) {
      return 'text-yellow-600';
    } else {
      return 'text-green-600';
    }
  };

  const getGainBarClasses = (stat) => {
    const isReduced = effectiveGains.diminishingReturns[stat] < 0.6;
    const colors = {
      strength: isReduced ? 'bg-red-300' : 'bg-gradient-to-r from-red-400 to-red-500',
      defense: isReduced ? 'bg-blue-300' : 'bg-gradient-to-r from-blue-400 to-blue-500',
      agility: isReduced ? 'bg-yellow-300' : 'bg-gradient-to-r from-yellow-400 to-yellow-500',
    };
    return colors[stat] || 'bg-gray-300';
  };

  if (isLoading) {
    return (
      <div className="p-8">
        <div className="text-center text-gray-500 py-12">Loading training types...</div>
      </div>
    );
  }

  if (trainingTypes.length === 0) {
    return (
      <div className="p-8">
        <div className="text-center text-red-600 bg-red-50 py-12 rounded-lg">No training types available</div>
      </div>
    );
  }

  return (
    <div className={`p-6 ${boxer?.has_active_rest ? 'has-rest' : ''}`}>
      {/* Energy Indicator - inline at top */}
      {selectedType && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <span className="text-sm text-gray-700">Energy: </span>
          <span className={`font-semibold ${getEnergyStatusClasses()}`}>
            {boxer?.energy || 0} / 100 available
          </span>
          {energyCost <= boxer?.energy && (
            <span className="text-sm text-gray-600 ml-2">({Math.round(boxer?.energy - energyCost)} after)</span>
          )}
        </div>
      )}

      {/* Rest Period Warning */}
      {boxer?.has_active_rest && (
        <div className={`mb-4 p-3 rounded-lg ${getRestWarningClasses()}`}>
          <div className="flex items-center gap-2">
            <span className="text-xl">{isForcedRest ? '🚑' : '🏥'}</span>
            <strong className={isForcedRest ? 'text-red-800' : 'text-yellow-800'}>
              {isForcedRest ? 'Forced Rest Active' : 'Currently Resting'}
            </strong>
            {restCountdownDisplay && (
              <span className={`text-xs ${isForcedRest ? 'text-red-700' : 'text-yellow-700'}`}>
                — {restCountdownDisplay}
              </span>
            )}
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {error && <ErrorDisplay error={error} variant="alert" title="Training Error" />}
        {success && <Alert variant="success" title="Success!" message={success} />}

        {/* Form Row: Type and Duration */}
        <div className="flex gap-4 flex-wrap">
          {/* Training Type Selection */}
          <div className="flex-1 min-w-[200px]">
            <label htmlFor="training-type" className="block text-sm font-medium text-gray-700 mb-2">Type</label>
            <select
              id="training-type"
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
              required
              disabled={isSubmitting || boxer?.has_active_rest}
              className="w-full p-3 border border-gray-300 rounded-lg text-sm bg-white transition-all duration-150 focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 disabled:bg-gray-100 disabled:cursor-not-allowed"
            >
              <option value="">Select type...</option>
              {trainingTypes.map(type => (
                <option key={type.id} value={type.id}>
                  {type.name} ({type.energy_cost} E/h)
                </option>
              ))}
            </select>
          </div>

          {/* Duration Slider */}
          <div className="flex-1 min-w-[200px]">
            <label htmlFor="duration" className="block text-sm font-medium text-gray-700 mb-2">Duration</label>
            <div className="flex items-center gap-4">
              <input
                id="duration"
                type="range"
                min="1"
                max="8"
                step="0.5"
                value={durationHours}
                onChange={(e) => setDurationHours(parseFloat(e.target.value))}
                required
                disabled={isSubmitting || boxer?.has_active_rest}
                className="flex-1 h-[6px] rounded bg-gradient-to-r from-indigo-200 to-indigo-400 appearance-none cursor-pointer accent-indigo-600 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2"
              />
              <span className="inline-flex items-center gap-1 px-3 py-1 bg-gradient-to-r from-gray-100 to-gray-200 rounded text-sm font-bold text-gray-800 min-w-[60px] justify-center">
                {durationHours}h
              </span>
            </div>
            <div className="flex justify-between w-[60px] mt-2 text-xs text-gray-500 font-semibold">
              <span>1</span>
              <span>8</span>
            </div>
          </div>
        </div>

        {/* Training Type Description */}
        {trainingType && (
          <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
            <p className="text-sm text-blue-800 m-0">{trainingType.description || 'No description available'}</p>
          </div>
        )}

        {/* Energy Cost Preview */}
        {selectedType && (
          <div className="flex items-center gap-2 p-3 bg-gray-50 rounded-lg">
            <span className="text-sm text-gray-700 font-medium">Cost: </span>
            <span className={`text-sm font-bold ${energyCost > boxer?.energy ? 'text-red-600' : 'text-gray-800'}`}>
              {energyCost} energy
            </span>
          </div>
        )}

        {/* Planned Gains Preview */}
        {selectedType && (
          <>
            {/* Fatigue Effectiveness Warning */}
            {effectiveGains.fatigueMultiplier < 0.8 && (
              <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg text-sm text-yellow-800">
                ⚠️ High fatigue reduces gains by {Math.round((1 - effectiveGains.fatigueMultiplier) * 100)}%
              </div>
            )}

            <div className="p-4 bg-gradient-to-br from-slate-50 to-slate-100 rounded-lg">
              {/* Gain Bars */}
              <div className="space-y-2 mb-3">
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-600 font-semibold w-[80px]">STR</span>
                  <div className="flex-1 h-[6px] bg-gray-200 rounded overflow-hidden">
                    <div
                      className={`h-full rounded transition-all duration-200 ${getGainBarClasses('strength')}`}
                      style={{ width: `${Math.min((effectiveGains.strength / 5) * 100, 100)}%` }}
                    ></div>
                  </div>
                  <span className="text-sm text-gray-800 font-bold w-[60px] text-right">+{effectiveGains.strength}</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-600 font-semibold w-[80px]">DEF</span>
                  <div className="flex-1 h-[6px] bg-gray-200 rounded overflow-hidden">
                    <div
                      className={`h-full rounded transition-all duration-200 ${getGainBarClasses('defense')}`}
                      style={{ width: `${Math.min((effectiveGains.defense / 5) * 100, 100)}%` }}
                    ></div>
                  </div>
                  <span className="text-sm text-gray-800 font-bold w-[60px] text-right">+{effectiveGains.defense}</span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-600 font-semibold w-[80px]">AGI</span>
                  <div className="flex-1 h-[6px] bg-gray-200 rounded overflow-hidden">
                    <div
                      className={`h-full rounded transition-all duration-200 ${getGainBarClasses('agility')}`}
                      style={{ width: `${Math.min((effectiveGains.agility / 5) * 100, 100)}%` }}
                    ></div>
                  </div>
                  <span className="text-sm text-gray-800 font-bold w-[60px] text-right">+{effectiveGains.agility}</span>
                </div>
              </div>

              {/* XP Preview */}
              <div className="text-center pt-2 border-t border-slate-200 text-sm text-purple-700 font-medium">
                +{effectiveGains.xp} XP toward Level {boxer?.level || 1}
              </div>
            </div>
          </>
        )}

        {/* Submit Button */}
        <button
          type="submit"
          className={`w-full inline-flex items-center justify-center gap-2 p-3 bg-gradient-to-r from-purple-500 to-purple-600 text-white border-none rounded-lg text-base font-bold uppercase tracking-wide cursor-pointer transition-all duration-150 hover:from-purple-400 hover:to-purple-500 hover:-translate-y-0.5 shadow-[0_4px_12px_rgba(124,58,237,0.35)] disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 ${''}`}
          disabled={!selectedType || isSubmitting || boxer?.has_active_rest}
          title={boxer?.has_active_rest ? (isForcedRest ? 'Cannot schedule during forced rest' : 'Cannot schedule during rest') : 'Schedule Training'}
        >
          {isSubmitting ? 'Scheduling...' : 'Schedule Training'}
        </button>
      </form>
    </div>
  );
};

export default TrainingScheduler;
