import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './FightBookingPage.css';
import TopBar from './TopBar';
import OpponentList from './OpponentList.jsx';
import PageLayout from './PageLayout';
import { fetchBestMatch } from '../api/opponents';
import { bookFight } from '../api/fights';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';

/**
 * FightBookingPage - Dedicated page for booking fights (MAT-145)
 *
 * Migrates FightBooking from modal to dedicated page at /boxer/:id/book-fight
 * Provides bookmarkable URL, full page real estate, and natural navigation flow.
 * Follows the same pattern as FightHistoryPage (MAT-121).
 */
const FightBookingPage = ({ user, onLogout }) => {
  const { id } = useParams(); // boxerId from URL
  const navigate = useNavigate();

  const [boxer, setBoxer] = useState(null);
  const [isLoadingBoxer, setIsLoadingBoxer] = useState(true);
  const [boxerError, setBoxerError] = useState(null);

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
  const [currentUser, setCurrentUser] = useState(null);

  // Load boxer data
  useEffect(() => {
    if (id) {
      loadBoxer();
    }

    // Get user from props or localStorage
    const currentUserData = user || getUser();
    setCurrentUser(currentUserData);
  }, [id, user]);

  // Load best match recommendation on mount
  useEffect(() => {
    if (boxer?.id) {
      loadBestMatch();
    }
  }, [boxer]);

  const loadBoxer = async () => {
    setIsLoadingBoxer(true);
    setBoxerError(null);

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/boxers/${id}`, { method: 'GET' });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to load boxer');
      }

      const data = await response.json();
      setBoxer(data);
    } catch (err) {
      console.error('[FightBookingPage] Error loading boxer:', err);
      setBoxerError(err.message || 'Failed to load boxer details');
    } finally {
      setIsLoadingBoxer(false);
    }
  };

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
      // Validation is optional - don't block booking
      // Using bookFight API which does validation internally
    } catch (err) {
      // Validation is optional - don't block booking
    }
  };

  const handleBookFight = async () => {
    if (!selectedOpponent || !boxer) return;

    setIsBooking(true);
    setBookingError('');

    try {
      await bookFight(boxer.id, {
        opponent_id: selectedOpponent.opponent.boxer.id
      });

      // Success - redirect to dashboard
      navigate('/dashboard');
    } catch (err) {
      // Use specific error message from utility with boxer/opponent context
      const errorMessage = getUserErrorMessage(err, 'fightBooking', {
        boxerName: boxer.name,
        opponentName: selectedOpponent.opponent.boxer.name
      });
      console.error('[FightBookingPage] Error booking fight:', err);
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

  const getLevelRangeMin = () => boxer && Math.max(1, boxer.level - filters.levelRange);
  const getLevelRangeMax = () => boxer && (boxer.level + filters.levelRange);

  const hasWarnings = matchValidation && (
    matchValidation.warnings && matchValidation.warnings.length > 0
  );

  // Loading state for boxer data
  if (isLoadingBoxer) {
    return (
      <>
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
          backButtonLabel="← Back to Dashboard"
        >
          <div className="loading-boxer">
            <div className="loading-spinner"></div>
            <p>Loading boxer details...</p>
          </div>
        </PageLayout>
      </>
    );
  }

  // Error state for boxer data
  if (boxerError || !boxer) {
    return (
      <>
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
          backButtonLabel="← Back to Dashboard"
        >
          <ErrorDisplay
            error={boxerError || 'Boxer not found'}
            onRetry={loadBoxer}
            title="Failed to Load Boxer"
            icon={true}
          />
        </PageLayout>
      </>
    );
  }

  return (
    <>
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
        backButtonLabel="← Back to Dashboard"
      >
        {/* Main Content */}
        <div className="booking-container">
          {/* Header */}
          <div className="booking-header">
            <h1>Book a Fight for {boxer.name}</h1>
          </div>

        {/* Boxer Summary */}
        <div className="boxer-summary">
          <div className="boxer-info">
            <span className="label">Booking for:</span>
            <span className="value">{boxer.name} (Level {boxer.level})</span>
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
              Level Range: ±{filters.levelRange} (Level {getLevelRangeMin()}-{getLevelRangeMax()})
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
        <div className="booking-actions">
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
      </PageLayout>
    </>
  );
};

export default FightBookingPage;
