import React, { useState, useEffect } from 'react';
import './FightBooking.css';
import OpponentList from './OpponentList.jsx';
import { API_BASE_URL, authenticatedFetch } from '../utils/auth';

const FightBooking = ({ boxer, onClose, onFightBooked }) => {
  const [selectedOpponent, setSelectedOpponent] = useState(null);
  const [filters, setFilters] = useState({
    levelRange: 3, // ±3 levels by default
    aiOnly: false,
    minHealth: 30,
    limit: 50
  });
  const [isBooking, setIsBooking] = useState(false);
  const [bookingError, setBookingError] = useState('');
  const [matchValidation, setMatchValidation] = useState(null);

  // Load best match recommendation on mount
  useEffect(() => {
    if (boxer?.id) {
      loadBestMatch();
    }
  }, [boxer]);

  const loadBestMatch = async () => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/opponents/best_match?boxer_id=${boxer.id}`, {
        method: 'GET'
      });
      if (response.ok) {
        const data = await response.json();
        // Pre-select the best match if available
        if (data.opponent) {
          setSelectedOpponent({
            opponent: data.opponent,
            score: data.score,
            reasoning: data.score?.reasoning || ''
          });
        }
      }
    } catch (err) {
      // Silently fail - best match is optional enhancement
    }
  };

  const handleSelectOpponent = (opportunity) => {
    setSelectedOpponent(opportunity);
    validateMatch(opportunity.opponent.boxer.id);
  };

  const validateMatch = async (opponentId) => {
    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/opponents/validate`, {
        method: 'POST',
        body: JSON.stringify({
          boxer1_id: boxer.id,
          boxer2_id: opponentId
        })
      });

      if (response.ok) {
        const validation = await response.json();
        setMatchValidation(validation);
      }
    } catch (err) {
      // Validation is optional - don't block booking
    }
  };

  const handleBookFight = async () => {
    if (!selectedOpponent) return;

    setIsBooking(true);
    setBookingError('');

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/boxers/${boxer.id}/fights`, {
        method: 'POST',
        body: JSON.stringify({
          opponent_id: selectedOpponent.opponent.boxer.id
        })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to book fight');
      }

      // Success - close modal and refresh parent
      onFightBooked(data);
      onClose();
    } catch (err) {
      setBookingError(err.message);
    } finally {
      setIsBooking(false);
    }
  };

  const handleFilterChange = (filterKey, value) => {
    setFilters(prev => ({ ...prev, [filterKey]: value }));
    // Reset selection when filters change
    setSelectedOpponent(null);
    setMatchValidation(null);
  };

  const getLevelRangeMin = () => Math.max(1, boxer.level - filters.levelRange);
  const getLevelRangeMax = () => boxer.level + filters.levelRange;

  const hasWarnings = matchValidation && (
    matchValidation.warnings && matchValidation.warnings.length > 0
  );

  return (
    <div className="fight-booking-modal">
      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <h2>Book a Fight for {boxer.name}</h2>
          <button className="modal-close" onClick={onClose} aria-label="Close">×</button>
        </div>

        {/* Boxer Summary */}
        <div className="boxer-summary">
          <div className="boxer-info">
            <span className="label">Booking for:</span>
            <span className="value">{boxer.name} (Lvl {boxer.level})</span>
          </div>
          <div className="boxer-stats-preview">
            <span>STR: {boxer.stats?.strength || 0}</span>
            <span>DEF: {boxer.stats?.defense || 0}</span>
            <span>AGI: {boxer.stats?.agility || 0}</span>
          </div>
        </div>

        {/* Filter Controls */}
        <div className="filter-controls">
          <div className="filter-group">
            <label htmlFor="level-range-slider">
              Level Range: ±{filters.levelRange} (Lvl {getLevelRangeMin()}-{getLevelRangeMax()})
            </label>
            <input
              id="level-range-slider"
              type="range"
              min="0"
              max="10"
              value={filters.levelRange}
              onChange={(e) => handleFilterChange('levelRange', parseInt(e.target.value))}
              className="level-range-slider"
            />
            <div className="slider-markers">
              <span>0</span>
              <span>5</span>
              <span>10+</span>
            </div>
          </div>

          <div className="filter-toggles">
            <label className="toggle-checkbox">
              <input
                type="checkbox"
                checked={filters.aiOnly}
                onChange={(e) => handleFilterChange('aiOnly', e.target.checked)}
              />
              <span className="toggle-label">AI Fighters Only</span>
            </label>

            <label className="toggle-checkbox">
              <input
                type="checkbox"
                checked={filters.minHealth >= 80}
                onChange={(e) => handleFilterChange('minHealth', e.target.checked ? 80 : 30)}
              />
              <span className="toggle-label">Excellent Health Only</span>
            </label>
          </div>
        </div>

        {/* Opponent Selection */}
        <div className="opponent-section">
          <OpponentList
            boxerId={boxer.id}
            boxer={boxer}
            filters={filters}
            onSelectOpponent={handleSelectOpponent}
          />
        </div>

        {/* Match Validation Warnings */}
        {hasWarnings && matchValidation && (
          <div className="match-validation-warnings">
            <h4>⚠️ Match Concerns:</h4>
            <ul>
              {matchValidation.warnings.map((warning, idx) => (
                <li key={idx}>{warning}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Booking Actions */}
        <div className="modal-actions">
          {selectedOpponent && (
            <div className="selection-summary">
              <span>Selected: <strong>{selectedOpponent.opponent.boxer.name}</strong></span>
              <span>Match Quality: <strong>{Math.round(selectedOpponent.score.overall_score * 100)}%</strong></span>
            </div>
          )}

          {bookingError && (
            <div className="booking-error">{bookingError}</div>
          )}

          <button
            className="book-fight-btn"
            onClick={handleBookFight}
            disabled={!selectedOpponent || isBooking}
          >
            {isBooking ? 'Booking...' : `Book Fight vs ${selectedOpponent?.opponent?.boxer?.name || 'Select Opponent'}`}
          </button>
        </div>
      </div>
    </div>
  );
};

export default FightBooking;
