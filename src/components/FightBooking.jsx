import React, { useState, useEffect } from 'react';
import './FightBooking.css';
import OpponentList from './OpponentList.jsx';
import { fetchBestMatch, validateFight } from '../api/opponents';
import { bookFight } from '../api/fights';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

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
      const data = await fetchBestMatch(boxer.id);
      // Pre-select the best match if available
      if (data) {
        setSelectedOpponent({
          opponent: data.opponent,
          score: data.score,
          reasoning: data.score?.reasoning || ''
        });
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
      const validation = await validateFight({
        boxerId: boxer.id,
        opponentId: opponentId,
        scheduledTime: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString() // Tomorrow
      });
      setMatchValidation(validation);
    } catch (err) {
      // Validation is optional - don't block booking
    }
  };

  const handleBookFight = async () => {
    if (!selectedOpponent) return;

    setIsBooking(true);
    setBookingError('');

    try {
      const data = await bookFight(boxer.id, {
        opponent_id: selectedOpponent.opponent.boxer.id
      });

      // Success - close modal and refresh parent
      onFightBooked(data);
      onClose();
    } catch (err) {
      // Use specific error message from utility with boxer/opponent context
      const errorMessage = getUserErrorMessage(err, 'fightBooking', {
        boxerName: boxer.name,
        opponentName: selectedOpponent.opponent.boxer.name
      });
      console.error('[FightBooking] Error booking fight:', err);
      setBookingError(errorMessage);
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
            <ErrorDisplay error={bookingError} variant="compact" icon={true} />
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
