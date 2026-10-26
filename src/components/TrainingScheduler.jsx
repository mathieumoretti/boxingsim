import React, { useState, useEffect } from 'react';
import './TrainingScheduler.css';
import { fetchTrainingTypes, startTrainingSession } from '../api/training';
import { fetchWorldTime } from '../api/world';
import { ErrorDisplay } from './ErrorDisplay';
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

  if (isLoading) {
    return (
      <div className="training-scheduler">
        <div className="loading">Loading training types...</div>
      </div>
    );
  }

  if (trainingTypes.length === 0) {
    return (
      <div className="training-scheduler">
        <div className="error-message">No training types available</div>
      </div>
    );
  }

  return (
    <div className={`training-scheduler ${compact ? 'compact' : ''} ${boxer?.has_active_rest ? 'has-rest' : ''}`}>
      {/* Energy Indicator - inline at top */}
      {selectedType && (
        <div className="energy-indicator">
          <span className="energy-label-text">Energy: </span>
          <span className={`energy-availability ${energyCost > boxer?.energy ? 'insufficient' : ''}`}>
            {boxer?.energy || 0} / 100 available
          </span>
          {energyCost <= boxer?.energy && (
            <span className="energy-after-text">({Math.round(boxer?.energy - energyCost)} after)</span>
          )}
        </div>
      )}

      {/* Rest Period Warning - compacted */}
      {boxer?.has_active_rest && (
        <div className={`rest-warning ${isForcedRest ? 'rest-forced' : 'rest-voluntary'}`}>
          <span className="warning-icon">{isForcedRest ? '🚑' : '🏥'}</span>
          <strong>{isForcedRest ? 'Forced Rest Active' : 'Currently Resting'}</strong>
          {restCountdownDisplay && (
            <span className="countdown-text">— {restCountdownDisplay}</span>
          )}
        </div>
      )}

      <form onSubmit={handleSubmit} className="training-form">
        {error && <ErrorDisplay error={error} variant="alert" title="Training Error" />}
        {success && <div className="success-message">{success}</div>}

        {/* Compact form layout */}
        <div className="form-row">
          {/* Training Type Selection */}
          <div className="form-group type-select">
            <label htmlFor="training-type">Type</label>
            <select
              id="training-type"
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value)}
              required
              disabled={isSubmitting || boxer?.has_active_rest}
              className="training-select"
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
          <div className="form-group duration-select">
            <label htmlFor="duration">Duration</label>
            <div className="duration-control">
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
                className="duration-slider"
              />
              <span className="duration-value">{durationHours}h</span>
            </div>
          </div>
        </div>

        {/* Training Type Description */}
        {trainingType && (
          <div className="training-description">
            <p className="description-text">{trainingType.description || 'No description available'}</p>
          </div>
        )}

        {/* Energy Cost Preview - moved up */}
        {selectedType && (
          <div className="energy-preview-compact">
            <span className="preview-label">Cost: </span>
            <span className={`preview-value ${energyCost > boxer?.energy ? 'over-budget' : ''}`}>
              {energyCost} energy
            </span>
          </div>
        )}

        {/* Planned Gains Preview */}
        {selectedType && (
          <>
            {/* Fatigue Effectiveness Warning */}
            {effectiveGains.fatigueMultiplier < 0.8 && (
              <div className="fatigue-warning">
                ⚠️ High fatigue reduces gains by {Math.round((1 - effectiveGains.fatigueMultiplier) * 100)}%
              </div>
            )}

            <div className="gains-preview-compact">
              <div className="gain-bars-compact">
                <div className={`gain-bar-mini strength ${effectiveGains.diminishingReturns.strength < 0.6 ? 'reduced' : ''}`}>
                  <span className="gain-label-mini">STR +{effectiveGains.strength}</span>
                  <div className="gain-progress-mini" style={{ width: `${Math.min((effectiveGains.strength / 5) * 100, 100)}%` }}></div>
                </div>
                <div className={`gain-bar-mini defense ${effectiveGains.diminishingReturns.defense < 0.6 ? 'reduced' : ''}`}>
                  <span className="gain-label-mini">DEF +{effectiveGains.defense}</span>
                  <div className="gain-progress-mini" style={{ width: `${Math.min((effectiveGains.defense / 5) * 100, 100)}%` }}></div>
                </div>
                <div className={`gain-bar-mini agility ${effectiveGains.diminishingReturns.agility < 0.6 ? 'reduced' : ''}`}>
                  <span className="gain-label-mini">AGI +{effectiveGains.agility}</span>
                  <div className="gain-progress-mini" style={{ width: `${Math.min((effectiveGains.agility / 5) * 100, 100)}%` }}></div>
                </div>
              </div>
              <div className="xp-preview-compact">
                +{effectiveGains.xp} XP toward Level {boxer?.level || 1}
              </div>
            </div>
          </>
        )}

        {/* Submit Button */}
        <button
          type="submit"
          className="schedule-btn"
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
