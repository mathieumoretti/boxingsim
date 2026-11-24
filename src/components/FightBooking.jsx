import React, { useState, useEffect } from 'react';
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
        scheduledTime: new Date(Date.now() + 12 * 60 * 60 * 1000).toISOString() // 12 hours ahead
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

  // Dynamic status badge classes based on clock state
  const getStatusBadgeClass = (status) => {
    switch(status) {
      case 'running': return 'bg-green-100 text-green-700 border-green-300';
      case 'paused': return 'bg-yellow-100 text-yellow-700 border-yellow-300';
      default: return 'bg-gray-100 text-gray-600 border-gray-300';
    }
  };

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-[40] p-8">
      <div className="bg-white rounded-xl max-w-[900px] w-full max-h-[90vh] overflow-y-auto shadow-[0_10px_40px_rgba(0,0,0,0.3)]">
        {/* Header */}
        <div className="flex items-center justify-between p-6 bg-gradient-to-r from-navy-dark to-navy-medium text-white rounded-t-xl sticky top-0 z-10">
          <h2 className="text-xl font-semibold m-0">Book a Fight for {boxer.name}</h2>
          <button
            className="bg-transparent border-none text-white text-2xl leading-none cursor-pointer p-0 w-10 h-10 flex items-center justify-center rounded-full transition-bg hover:bg-white/10 focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            onClick={onClose}
            aria-label="Close"
          >
            ×
          </button>
        </div>

        {/* Boxer Summary */}
        <div className="flex justify-between items-center p-4 bg-gray-50 border-b border-gray-200">
          <div className="flex flex-col gap-1">
            <span className="text-sm text-gray-500">Booking for:</span>
            <span className="text-lg font-semibold text-gray-800">{boxer.name} (Lvl {boxer.level})</span>
          </div>
          <div className="flex gap-4 text-sm text-gray-600">
            <span className="bg-white px-3 py-1 rounded-sm shadow-[0_1px_3px_rgba(0,0,0,0.1)]">STR: {boxer.stats?.strength || 0}</span>
            <span className="bg-white px-3 py-1 rounded-sm shadow-[0_1px_3px_rgba(0,0,0,0.1)]">DEF: {boxer.stats?.defense || 0}</span>
            <span className="bg-white px-3 py-1 rounded-sm shadow-[0_1px_3px_rgba(0,0,0,0.1)]">AGI: {boxer.stats?.agility || 0}</span>
          </div>
        </div>

        {/* Filter Controls */}
        <div className="p-5 bg-gray-50 border-b border-gray-200">
          <div className="mb-4">
            <label htmlFor="level-range-slider" className="block text-sm font-medium text-gray-700 mb-2">
              Level Range: ±{filters.levelRange} (Lvl {getLevelRangeMin()}-{getLevelRangeMax()})
            </label>
            <input
              id="level-range-slider"
              type="range"
              min="0"
              max="10"
              value={filters.levelRange}
              onChange={(e) => handleFilterChange('levelRange', parseInt(e.target.value))}
              className="w-full h-[6px] rounded-sm bg-gray-300 outline-none appearance-none cursor-pointer accent-primary"
              style={{
                '--range-thumb-size': '20px',
                '--range-track-color': '#d1d5db'
              }}
            />
            <div className="flex justify-between text-xs text-gray-500 mt-1">
              <span>0</span>
              <span>5</span>
              <span>10+</span>
            </div>
          </div>

          <div className="flex gap-5 flex-wrap">
            <label className="flex items-center gap-2 cursor-pointer text-sm text-gray-700">
              <input
                type="checkbox"
                checked={filters.aiOnly}
                onChange={(e) => handleFilterChange('aiOnly', e.target.checked)}
                className="w-[18px] h-[18px] border-2 border-gray-300 rounded-sm accent-primary cursor-pointer focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
              />
              <span className="font-medium">AI Fighters Only</span>
            </label>

            <label className="flex items-center gap-2 cursor-pointer text-sm text-gray-700">
              <input
                type="checkbox"
                checked={filters.minHealth >= 80}
                onChange={(e) => handleFilterChange('minHealth', e.target.checked ? 80 : 30)}
                className="w-[18px] h-[18px] border-2 border-gray-300 rounded-sm accent-primary cursor-pointer focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
              />
              <span className="font-medium">Excellent Health Only</span>
            </label>
          </div>
        </div>

        {/* Opponent Selection */}
        <div className="p-0 px-6 pb-5">
          <OpponentList
            boxerId={boxer.id}
            boxer={boxer}
            filters={filters}
            onSelectOpponent={handleSelectOpponent}
          />
        </div>

        {/* Match Validation Warnings */}
        {hasWarnings && matchValidation && (
          <div className="p-4 bg-yellow-50 border border-yellow-400 rounded-md mx-6 mb-4">
            <h4 className="text-sm font-semibold text-yellow-700 m-0 mb-2">⚠️ Match Concerns:</h4>
            <ul className="list-disc list-inside text-xs text-yellow-800 m-0">
              {matchValidation.warnings.map((warning, idx) => (
                <li key={idx} className="mb-1">{warning}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Booking Actions */}
        <div className="p-5 bg-gray-50 border-t border-gray-200 rounded-b-xl">
          {selectedOpponent && (
            <div className="flex justify-between mb-4 text-sm text-gray-600">
              <span>Selected: <strong className="text-gray-800">{selectedOpponent.opponent.boxer.name}</strong></span>
              <span>Match Quality: <strong className="text-gray-800">{Math.round(selectedOpponent.score.overall_score * 100)}%</strong></span>
            </div>
          )}

          {bookingError && (
            <ErrorDisplay error={bookingError} variant="compact" icon={true} />
          )}

          <button
            className={`w-full p-4 bg-gradient-to-r from-red-500 to-red-600 text-white border-none rounded-md text-lg font-semibold cursor-pointer transition-all duration-150 hover:opacity-90 hover:-translate-y-0.5 disabled:bg-gray-400 disabled:cursor-not-allowed disabled:opacity-60 focus-visible:outline-2 focus-visible:outline-red-600 focus-visible:outline-offset-2 ${''}`}
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
