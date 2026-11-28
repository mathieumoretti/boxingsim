import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import TopBar from './TopBar';
import { SkipLink } from './SkipLink';
import OpponentList from './OpponentList.jsx';
import PageLayout from './PageLayout';
import { Badge } from './ui';
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

      // Success - back to the boxer's profile (shows the scheduled fight)
      navigate(id ? `/boxer/${id}` : '/boxers');
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
        <SkipLink targetId="booking-content" />
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
          backButtonLabel="← Back to Dashboard"
        >
          <main id="booking-content" className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8 focus:outline-none" tabIndex="-1">
            <div className="flex flex-col items-center justify-center py-20">
              <div className="w-12 h-12 border-4 border-primary/30 border-t-primary rounded-full animate-spin mb-4"></div>
              <p className="text-base text-white/60">Loading boxer details...</p>
            </div>
          </main>
        </PageLayout>
      </>
    );
  }

  // Error state for boxer data
  if (boxerError || !boxer) {
    return (
      <>
        <SkipLink targetId="booking-content" />
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
          backButtonLabel="← Back to Dashboard"
        >
          <main id="booking-content" className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8 focus:outline-none" tabIndex="-1">
            <ErrorDisplay
              error={boxerError || 'Boxer not found'}
              onRetry={loadBoxer}
              title="Failed to Load Boxer"
              icon={true}
            />
          </main>
        </PageLayout>
      </>
    );
  }

  return (
    <>
      <SkipLink targetId="booking-content" />
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        variant="fight-booking-page"
        onBack={() => navigate(id ? `/boxer/${id}` : '/boxers')}
        backButtonLabel="← Back to Profile"
      >
        {/* Main Content */}
        <main id="booking-content" className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8 focus:outline-none" tabIndex="-1">
          {/* Boxer Summary (Header Info) */}
          <header className="bg-gradient-to-br from-white/10 to-white/5 border border-white/10 rounded-xl p-6 mb-6 text-white" aria-label="Fighter information">
            <div className="mb-3">
              <span className="text-xs font-semibold uppercase tracking-wider text-white/70">BOOK A FIGHT FOR</span>
              <h1 className="text-3xl font-bold m-0">{boxer.name}</h1>
            </div>
            <div className="flex flex-wrap items-center gap-4">
              {/* Migrated to Badge component (MAT-154) */}
              <Badge variant="level" size="md">Level {boxer.level}</Badge>
              <div className="flex items-center gap-3 text-sm font-semibold">
                <span className="text-[#FF6B6B]" title="Strength">STR {boxer.stats?.strength || 0}</span>
                <span className="text-[#4ECDC4]" title="Defense">DEF {boxer.stats?.defense || 0}</span>
                <span className="text-[#95E1D3]" title="Agility">AGI {boxer.stats?.agility || 0}</span>
              </div>
            </div>
          </header>

        {/* Filter Controls */}
        <section aria-label="Opponent filters" className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-6 border border-white/10 mb-6">
          <fieldset className="mb-5 pb-5 border-b border-white/10">
            <legend className="text-sm font-semibold text-white/80 mb-3">Level Range: {getLevelRangeMin()}-{getLevelRangeMax()}</legend>
            <input
              type="range"
              min="0"
              max="10"
              value={filters.levelRange}
              onChange={(e) => handleFilterChange('levelRange', parseInt(e.target.value))}
              className="w-full h-2 bg-white/15 rounded-lg appearance-none cursor-pointer accent-primary-light"
              aria-label="Set level range for opponent matching"
            />
            <div className="flex justify-between mt-2 text-xs font-semibold text-white/50">
              <span>0</span>
              <span>5</span>
              <span>10+</span>
            </div>
          </fieldset>

          <div className="flex flex-wrap gap-4" role="group" aria-label="Filter options">
            <label className="inline-flex items-center gap-2 cursor-pointer group">
              <input
                type="checkbox"
                checked={filters.aiOnly}
                onChange={(e) => handleFilterChange('aiOnly', e.target.checked)}
                className="w-4 h-4 accent-primary-light cursor-pointer"
              />
              <span className="text-sm font-semibold text-white/80 group-hover:text-white transition-colors">AI Fighters Only</span>
            </label>

            <label className="inline-flex items-center gap-2 cursor-pointer group">
              <input
                type="checkbox"
                checked={filters.minHealth >= 80}
                onChange={(e) => handleFilterChange('minHealth', e.target.checked ? 80 : 30)}
                className="w-4 h-4 accent-primary-light cursor-pointer"
              />
              <span className="text-sm font-semibold text-white/80 group-hover:text-white transition-colors">Excellent Health Only</span>
            </label>
          </div>
        </section>

        {/* Opponent Selection */}
        <section aria-label="Opponent selection" className="mb-6">
          <OpponentList
            boxerId={boxer.id}
            boxer={boxer}
            filters={filters}
            onSelectOpponent={handleSelectOpponent}
          />
        </section>

        {/* Match Validation Warnings */}
        {hasWarnings && matchValidation && (
          <div role="alert" className="bg-amber-500/10 border border-amber-400/30 rounded-lg p-4 mb-6">
            <h4 className="m-0 mb-2 text-base font-semibold text-amber-200">⚠️ Match Concerns:</h4>
            <ul className="m-0 space-y-1 list-disc list-inside text-sm text-amber-200/90">
              {matchValidation.warnings.map((warning, idx) => (
                <li key={idx}>{warning}</li>
              ))}
            </ul>
          </div>
        )}

        {/* Booking Actions */}
        <footer className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-6 border border-white/10">
          {selectedOpponent && (
            <div className="mb-4 p-4 bg-white/5 rounded-lg border border-white/10">
              <p className="m-0 text-sm font-semibold text-white">
                Selected: <strong>{selectedOpponent.opponent.boxer.name}</strong>
              </p>
              <p className="m-0 mt-1 text-sm font-semibold text-white/70">
                Match Quality: <strong>{Math.round(selectedOpponent.score.overall_score * 100)}%</strong>
              </p>
            </div>
          )}

          {bookingError && (
            <div className="mb-4">
              <ErrorDisplay error={bookingError} variant="compact" icon={true} />
            </div>
          )}

          <button
            className={`w-full py-3 px-4 text-base font-semibold rounded-lg border-none transition-all duration-150 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 ${
              !selectedOpponent || isBooking
                ? 'bg-white/10 text-white/40 cursor-not-allowed'
                : 'bg-gradient-to-br from-success to-emerald-600 text-white hover:bg-gradient-to-br hover:from-emerald-400 hover:to-success hover:-translate-y-0.5 hover:shadow-[0_4px_12px_rgba(34,197,94,0.3)]'
            }`}
            onClick={handleBookFight}
            disabled={!selectedOpponent || isBooking}
          >
            {isBooking ? (
              <span className="flex items-center justify-center gap-2">
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                Booking...
              </span>
            ) : (
              `Book Fight vs ${selectedOpponent?.opponent?.boxer?.name || 'Select Opponent'}`
            )}
          </button>
        </footer>
      </main>
      </PageLayout>
    </>
  );
};

export default FightBookingPage;
