/**
 * BoxersList Page Component (MAT-122)
 *
 * Dedicated page for viewing and managing all user boxers.
 * Separated from Dashboard to establish scalable navigation structure.
 *
 * Features:
 * - Full roster grid with boxer cards
 * - Search/filter capability (future enhancement)
 * - Quick stats summary at top
 * - Empty state and error handling
 */

import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import TopBar from '../components/TopBar';
import BoxerCard from '../components/BoxerCard';
import PageHeader from '../components/PageHeader';
import { ErrorDisplay } from '../components/ErrorDisplay';
import LoadingSpinner from '../components/ui/LoadingSpinner';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';
import { SkipLink } from '../components/SkipLink';

const BoxersList = ({ user, onLogout }) => {
  const navigate = useNavigate();
  const [boxers, setBoxers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentUser, setCurrentUser] = useState(null);
  const [worldTime, setWorldTime] = useState(null);
  const [fightData, setFightData] = useState({});

  // Search and filter state (MAT-122)
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all'); // all, available, training, resting

  useEffect(() => {
    const currentUser = user || getUser();
    setCurrentUser(currentUser);

    if (currentUser) {
      loadUserBoxers(currentUser.id);
    } else {
      setIsLoading(false);
    }
  }, [user]);

  // Fetch world time for countdown displays
  useEffect(() => {
    const fetchWorldTime = async () => {
      try {
        const response = await authenticatedFetch(`${API_BASE_URL}/world/time`, { method: 'GET' });
        if (response.ok) {
          const data = await response.json();
          setWorldTime(data);
        }
      } catch (err) {
        // Silently fail
      }
    };

    fetchWorldTime();
    const interval = setInterval(fetchWorldTime, 5000);
    return () => clearInterval(interval);
  }, []);

  // Refresh the roster periodically so training/rest badges stay current
  useEffect(() => {
    if (!currentUser) return undefined;
    const interval = setInterval(() => loadUserBoxers(currentUser.id), 15000);
    return () => clearInterval(interval);
  }, [currentUser, loadUserBoxers]);

  const handleBoxerStateChanged = useCallback((boxerId) => {
    if (currentUser) {
      loadUserBoxers(currentUser.id);
    }
  }, [currentUser]);

  const loadUserBoxers = useCallback(async (userId) => {
    setIsLoading(true);
    setError('');

    try {
      const [boxersResponse, fightsResponse] = await Promise.all([
        authenticatedFetch(`${API_BASE_URL}/users/${userId}/boxers`, { method: 'GET' }),
        authenticatedFetch(`${API_BASE_URL}/fights/active`, { method: 'GET' })
      ]);

      const boxersData = await boxersResponse.json();
      let boxersArray = [];

      if (boxersResponse.ok) {
        boxersArray = Array.isArray(boxersData) ? boxersData : [];

        // Attach each boxer's active training session so cards can show
        // a training badge without extra per-card requests
        boxersArray = await Promise.all(boxersArray.map(async (boxer) => {
          try {
            const sessionResponse = await authenticatedFetch(
              `${API_BASE_URL}/boxers/${boxer.id}/training-sessions`,
              { method: 'GET' }
            );
            if (sessionResponse.ok) {
              const sessions = await sessionResponse.json();
              if (Array.isArray(sessions) && sessions.length > 0) {
                return { ...boxer, has_active_training: true, training_session: sessions[0] };
              }
            }
          } catch (err) {
            // Non-fatal — badge simply won't show
          }
          return { ...boxer, has_active_training: false };
        }));

        setBoxers(boxersArray);
      } else {
        setError(boxersData.error || 'Failed to load boxers');
      }

      // Build fight mapping
      if (fightsResponse.ok) {
        const fightsData = await fightsResponse.json();
        const activeFights = Array.isArray(fightsData) ? fightsData : [];

        if (boxersArray.length > 0) {
          const boxerIds = boxersArray.map(b => b.id);
          const fightMapping = {};

          activeFights.forEach(fight => {
            if (boxerIds.includes(fight.boxer1_id)) {
              fightMapping[fight.boxer1_id] = {
                fight_id: fight.id,
                opponent_id: fight.boxer2_id,
                scheduled_time: fight.scheduled_time,
                status: fight.status,
                rounds: fight.round || 12
              };
            } else if (boxerIds.includes(fight.boxer2_id)) {
              fightMapping[fight.boxer2_id] = {
                fight_id: fight.id,
                opponent_id: fight.boxer1_id,
                scheduled_time: fight.scheduled_time,
                status: fight.status,
                rounds: fight.round || 12
              };
            }
          });

          setFightData(fightMapping);
        }
      }

    } catch (error) {
      if (!error.message.includes('Unauthorized')) {
        setError('Network error: ' + error.message);
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleLogout = useCallback(() => {
    if (onLogout) {
      onLogout();
    } else {
      window.location.replace('/login');
    }
  }, [onLogout]);

  // Calculate quick stats (memoized to prevent recalculation on every render - MAT-132)
  const stats = React.useMemo(() => ({
    totalBoxers: boxers.length,
    inTraining: boxers.filter(b => b.has_active_training).length,
    resting: boxers.filter(b => b.has_active_rest).length,
    fightsScheduled: Object.values(fightData).filter(f => f?.status === 'scheduled').length,
  }), [boxers, fightData]);

  // Filter boxers based on search and status (memoized callback - MAT-132)
  const filterBoxers = React.useCallback((boxerList) => {
    return boxerList.filter(boxer => {
      // Search filter: name or nickname
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesName = boxer.name?.toLowerCase().includes(query);
        const matchesNickname = boxer.nickname?.toLowerCase().includes(query);
        if (!matchesName && !matchesNickname) {
          return false;
        }
      }

      // Status filter
      if (statusFilter === 'available') {
        return !boxer.has_active_training && !boxer.has_active_rest;
      } else if (statusFilter === 'training') {
        return boxer.has_active_training && !boxer.has_active_rest;
      } else if (statusFilter === 'resting') {
        return boxer.has_active_rest;
      }

      return true;
    });
  }, [searchQuery, statusFilter]);

  const filteredBoxers = filterBoxers(boxers);

  return (
    <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
      <SkipLink targetId="main-content" />
      <TopBar currentUser={currentUser} onLogout={handleLogout} />

      <main id="main-content" className="max-w-[1400px] mx-auto px-6 py-6" tabIndex="-1">
          {/* Page Header */}
          <PageHeader
            title="My Boxers"
            subtitle="Manage your fighter roster, track training progress, and book fights."
            actions={
              <button
                onClick={() => navigate('/create-boxer')}
                className="inline-flex items-center gap-2 px-4 py-2.5 bg-gradient-to-br from-primary to-primary-dark text-white rounded-lg text-sm font-semibold cursor-pointer transition-all duration-150 hover:from-primary-light hover:to-primary hover:-translate-y-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
              >
                <span aria-hidden="true">+</span>
                Create Fighter
              </button>
            }
          />

          {/* Quick Stats Summary */}
          {boxers.length > 0 && (
            <section className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8" aria-label="Roster summary">
              <StatSummaryCard
                icon="🥊"
                value={stats.totalBoxers}
                label="Total Boxers"
              />
              <StatSummaryCard
                icon="💪"
                value={stats.inTraining}
                label="In Training"
              />
              <StatSummaryCard
                icon="🛌"
                value={stats.resting}
                label="Resting"
              />
              <StatSummaryCard
                icon="📅"
                value={stats.fightsScheduled}
                label="Fights Scheduled"
              />
            </section>
          )}

          {/* Search and Filter Bar (MAT-122) */}
          {boxers.length > 0 && (
            <div className="flex flex-col md:flex-row md:items-center gap-4 mb-4">
              {/* Search Input */}
              <div className="relative flex-1 min-w-[200px] max-w-[350px]">
                <input
                  type="text"
                  placeholder="Search boxers by name..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full px-4 py-2.5 bg-navy-dark border border-white/20 rounded-lg text-sm text-white placeholder:text-white/40 focus:outline-none focus:border-primary-light focus:ring-2 focus:ring-primary/30"
                  aria-label="Search boxers"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-3 top-1/2 -translate-y-1/2 bg-transparent border-none text-white/60 hover:text-white cursor-pointer text-base"
                    aria-label="Clear search"
                  >
                    ×
                  </button>
                )}
              </div>

              {/* Status Filter */}
              <div className="flex flex-wrap gap-2" role="group" aria-label="Filter by status">
                <button
                  className={`px-4 py-2 rounded-lg text-sm font-semibold border transition-all duration-150 cursor-pointer ${statusFilter === 'all' ? 'bg-gradient-to-br from-primary to-primary-dark text-white border-primary/50' : 'bg-white/10 text-white/80 border-white/20 hover:bg-white/15 hover:border-white/30'}`}
                  onClick={() => setStatusFilter('all')}
                  aria-pressed={statusFilter === 'all'}
                >
                  All ({boxers.length})
                </button>
                <button
                  className={`px-4 py-2 rounded-lg text-sm font-semibold border transition-all duration-150 cursor-pointer ${statusFilter === 'available' ? 'bg-gradient-to-br from-primary to-primary-dark text-white border-primary/50' : 'bg-white/10 text-white/80 border-white/20 hover:bg-white/15 hover:border-white/30'}`}
                  onClick={() => setStatusFilter('available')}
                  aria-pressed={statusFilter === 'available'}
                >
                  Available ({boxers.filter(b => !b.has_active_training && !b.has_active_rest).length})
                </button>
                <button
                  className={`px-4 py-2 rounded-lg text-sm font-semibold border transition-all duration-150 cursor-pointer ${statusFilter === 'training' ? 'bg-gradient-to-br from-primary to-primary-dark text-white border-primary/50' : 'bg-white/10 text-white/80 border-white/20 hover:bg-white/15 hover:border-white/30'}`}
                  onClick={() => setStatusFilter('training')}
                  aria-pressed={statusFilter === 'training'}
                >
                  Training ({stats.inTraining})
                </button>
                <button
                  className={`px-4 py-2 rounded-lg text-sm font-semibold border transition-all duration-150 cursor-pointer ${statusFilter === 'resting' ? 'bg-gradient-to-br from-primary to-primary-dark text-white border-primary/50' : 'bg-white/10 text-white/80 border-white/20 hover:bg-white/15 hover:border-white/30'}`}
                  onClick={() => setStatusFilter('resting')}
                  aria-pressed={statusFilter === 'resting'}
                >
                  Resting ({stats.resting})
                </button>
              </div>
            </div>
          )}

          {/* Results Count */}
          {boxers.length > 0 && (
            <div className="text-sm text-white/60 mb-6 flex flex-wrap items-center gap-4">
              Showing {filteredBoxers.length} of {boxers.length} boxers
              {(searchQuery || statusFilter !== 'all') && (
                <button
                  onClick={() => { setSearchQuery(''); setStatusFilter('all'); }}
                  className="bg-transparent border-none text-primary-light text-sm font-semibold cursor-pointer hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
                >
                  Clear Filters
                </button>
              )}
            </div>
          )}

          {/* Main Content Area */}
          {isLoading ? (
            <div className="py-16 flex items-center justify-center">
              <LoadingSpinner size="lg" aria-label="Loading boxers" />
            </div>
          ) : error ? (
            <ErrorDisplay
              error={error}
              onRetry={() => loadUserBoxers(currentUser?.id)}
              variant="alert"
              title="Failed to Load Boxers"
              icon={true}
            />
          ) : boxers.length === 0 ? (
            <div className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-12 border border-white/10 text-center">
              <div className="text-5xl mb-4" aria-hidden="true">🥊</div>
              <h2 className="text-lg font-bold text-white m-0 mb-2">Your roster is empty</h2>
              <p className="text-white/60 m-0 mb-6">
                Create your first boxer and begin their journey to championship glory.
                Train, fight, and build a legendary career in the ring.
              </p>
              <button
                onClick={() => navigate('/create-boxer')}
                className="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-br from-primary to-primary-dark text-white rounded-lg font-semibold cursor-pointer transition-all duration-150 hover:from-primary-light hover:to-primary hover:-translate-y-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
              >
                <span aria-hidden="true">➕</span>
                Create Your First Boxer
              </button>
            </div>
          ) : (
            <section aria-label="Boxer roster">
              {filteredBoxers.length === 0 ? (
                <div className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-10 border border-white/10 text-center">
                  <div className="text-4xl mb-3" aria-hidden="true">🔍</div>
                  <h3 className="text-base font-bold text-white/80 m-0 mb-2">No boxers found</h3>
                  <p className="text-sm text-white/60 m-0 mb-5">
                    {searchQuery
                      ? `No boxers match "${searchQuery}". Try a different search term.`
                      : 'No boxers match the selected filters.'}
                  </p>
                  <button
                    onClick={() => { setSearchQuery(''); setStatusFilter('all'); }}
                    className="px-4 py-2 bg-white/10 text-white border border-white/20 rounded-lg text-sm font-semibold cursor-pointer transition-all duration-150 hover:bg-white/15 hover:border-white/30 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
                  >
                    Clear All Filters
                  </button>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6" role="list">
                  {filteredBoxers.map((boxer) => (
                  <BoxerCard
                    key={boxer.id}
                    boxer={boxer}
                    worldTime={worldTime}
                    upcomingFight={fightData[boxer.id]}
                    onOpenProfile={() => navigate(`/boxer/${boxer.id}`)}
                    onOpenTraining={() => navigate(`/boxer/${boxer.id}/train`)}
                    onOpenFightBooking={() => navigate(`/boxer/${boxer.id}/book-fight`)}
                    onOpenFightHistory={() => navigate(`/boxer/${boxer.id}/fights`)}
                    onBoxerStateChanged={handleBoxerStateChanged}
                  />
                ))}
                </div>
              )}
            </section>
          )}
      </main>
    </div>
  );
};

// Simple inline stat card component
const StatSummaryCard = ({ icon, value, label }) => (
  <div className="bg-gradient-to-br from-white/10 to-white/5 rounded-xl p-4 border border-white/10">
    <span className="text-2xl mb-1 block" aria-hidden="true">{icon}</span>
    <span className="text-2xl font-bold text-white block mb-0.5">{value}</span>
    <span className="text-sm text-white/70">{label}</span>
  </div>
);

export default BoxersList;
