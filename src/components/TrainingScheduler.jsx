import React, { useState, useEffect } from 'react';
import './TrainingScheduler.css';
import { API_BASE_URL } from '../utils/auth';

const TrainingScheduler = ({ boxerId, boxer, onClose }) => {
  const [trainingTypes, setTrainingTypes] = useState([]);
  const [selectedType, setSelectedType] = useState('');
  const [durationHours, setDurationHours] = useState(2);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    loadTrainingTypes();
  }, []);

  const loadTrainingTypes = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/training-types`, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      });

      if (response.ok) {
        const data = await response.json();
        setTrainingTypes(Array.isArray(data) ? data : []);
      } else {
        setError('Failed to load training types');
      }
    } catch (err) {
      setError('Network error: ' + err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const calculateEnergyCost = () => {
    if (!selectedType) return 0;
    const type = trainingTypes.find(t => t.id === parseInt(selectedType));
    return type ? Math.round(type.energy_cost * durationHours) : 0;
  };

  const calculateGains = () => {
    if (!selectedType) return { strength: 0, defense: 0, agility: 0 };
    const type = trainingTypes.find(t => t.id === parseInt(selectedType));
    if (!type) return { strength: 0, defense: 0, agility: 0 };

    return {
      strength: Math.round(type.strength_gain_factor * durationHours * 100) / 100,
      defense: Math.round(type.defense_gain_factor * durationHours * 100) / 100,
      agility: Math.round(type.agility_gain_factor * durationHours * 100) / 100
    };
  };

  const getSelectedTrainingType = () => {
    return trainingTypes.find(t => t.id === parseInt(selectedType));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setIsSubmitting(true);

    try {
      const response = await fetch(`${API_BASE_URL}/boxers/${boxerId}/train`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
          training_type_id: parseInt(selectedType),
          duration_hours: durationHours,
          scheduled_at: new Date().toISOString()
        })
      });

      const data = await response.json();

      if (response.ok) {
        setSuccess('Training session scheduled successfully!');
        setError('');
        // Reset form after success
        setSelectedType('');
        setDurationHours(2);
        // Close modal after a brief delay to show success message
        setTimeout(() => {
          setSuccess('');
          if (onClose) onClose();
        }, 1500);
      } else {
        setError(data.error || 'Failed to schedule training');
      }
    } catch (err) {
      if (err.message.includes('Unauthorized')) {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
      } else {
        setError('Network error: ' + err.message);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const energyCost = calculateEnergyCost();
  const gains = calculateGains();
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
    <div className="training-scheduler">
      <h2>Schedule Training</h2>

      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}

      <form onSubmit={handleSubmit} className="training-form">
        {/* Training Type Selection */}
        <div className="form-group">
          <label htmlFor="training-type">Training Type</label>
          <select
            id="training-type"
            value={selectedType}
            onChange={(e) => setSelectedType(e.target.value)}
            required
            disabled={isSubmitting}
            className="training-select"
          >
            <option value="">Select a training type...</option>
            {trainingTypes.map(type => (
              <option key={type.id} value={type.id}>
                {type.name} - {type.energy_cost} energy/hour
              </option>
            ))}
          </select>
        </div>

        {/* Training Type Description */}
        {trainingType && (
          <div className="training-description">
            <p><strong>{trainingType.name}</strong></p>
            <p className="description-text">{trainingType.description || 'No description available'}</p>
          </div>
        )}

        {/* Duration Slider */}
        <div className="form-group">
          <label htmlFor="duration">
            Duration: <span className="highlight">{durationHours} hour{durationHours !== 1 ? 's' : ''}</span>
          </label>
          <input
            id="duration"
            type="range"
            min="1"
            max="8"
            step="0.5"
            value={durationHours}
            onChange={(e) => setDurationHours(parseFloat(e.target.value))}
            required
            disabled={isSubmitting}
            className="duration-slider"
          />
          <div className="range-markers">
            <span>1h</span>
            <span>4h</span>
            <span>8h</span>
          </div>
        </div>

        {/* Energy Cost Preview */}
        {selectedType && (
          <div className="energy-preview">
            <div className="preview-item">
              <span className="preview-label">Energy Cost:</span>
              <span className={`preview-value ${energyCost > boxer?.energy ? 'over-budget' : ''}`}>
                {energyCost} / {boxer?.energy || 100} available
              </span>
            </div>
          </div>
        )}

        {/* Planned Gains Preview */}
        {selectedType && (
          <div className="gains-preview">
            <h4>Planned Stat Gains</h4>
            <div className="gain-bars">
              <div className="gain-bar strength">
                <span className="gain-label">💪 STR +{gains.strength}</span>
                <div
                  className="gain-progress"
                  style={{ width: `${(gains.strength / 10) * 100}%` }}
                ></div>
              </div>
              <div className="gain-bar defense">
                <span className="gain-label">🛡️ DEF +{gains.defense}</span>
                <div
                  className="gain-progress"
                  style={{ width: `${(gains.defense / 10) * 100}%` }}
                ></div>
              </div>
              <div className="gain-bar agility">
                <span className="gain-label">⚡ AGI +{gains.agility}</span>
                <div
                  className="gain-progress"
                  style={{ width: `${(gains.agility / 10) * 100}%` }}
                ></div>
              </div>
            </div>
          </div>
        )}

        {/* Submit Button */}
        <button
          type="submit"
          className="schedule-btn"
          disabled={!selectedType || isSubmitting}
        >
          {isSubmitting ? 'Scheduling...' : 'Schedule Training Session'}
        </button>
      </form>
    </div>
  );
};

export default TrainingScheduler;
