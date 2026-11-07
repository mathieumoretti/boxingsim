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
import '../styles/ring-canvas.css';

const BoxersList = ({ user }) => {
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
    window.location.replace('/login');
  }, []);

  // Calculate quick stats
  const stats = {
    totalBoxers: boxers.length,
    inTraining: boxers.filter(b => b.has_active_training).length,
    resting: boxers.filter(b => b.has_active_rest).length,
    fightsScheduled: Object.values(fightData).filter(f => f?.status === 'scheduled').length,
  };

  // Filter boxers based on search and status (MAT-122)
  const filterBoxers = useCallback((boxerList) => {
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
    <div className="boxers-list-page">
      <TopBar currentUser={currentUser} onLogout={handleLogout} />

      <div className="ring-canvas">
        <div className="container">
          {/* Page Header */}
          <PageHeader
            title="My Boxers"
            subtitle="Manage your fighter roster, track training progress, and book fights."
            actions={
              <button
                onClick={() => navigate('/create-boxer')}
                className="btn-create-boxer"
              >
                <span aria-hidden="true">+</span>
                Create Fighter
              </button>
            }
            breadcrumbs={[
              { label: 'Dashboard', href: '/dashboard' },
              { label: 'Boxers' }
            ]}
          />

          {/* Quick Stats Summary */}
          {boxers.length > 0 && (
            <section className="boxers-stats" aria-label="Roster summary">
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
            <div className="boxers-filter-bar">
              {/* Search Input */}
              <div className="filter-search">
                <input
                  type="text"
                  placeholder="Search boxers by name..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="filter-search-input"
                  aria-label="Search boxers"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="filter-search-clear"
                    aria-label="Clear search"
                  >
                    ×
                  </button>
                )}
              </div>

              {/* Status Filter */}
              <div className="filter-status" role="group" aria-label="Filter by status">
                <button
                  className={`filter-status-btn ${statusFilter === 'all' ? 'active' : ''}`}
                  onClick={() => setStatusFilter('all')}
                  aria-pressed={statusFilter === 'all'}
                >
                  All ({boxers.length})
                </button>
                <button
                  className={`filter-status-btn ${statusFilter === 'available' ? 'active' : ''}`}
                  onClick={() => setStatusFilter('available')}
                  aria-pressed={statusFilter === 'available'}
                >
                  Available ({boxers.filter(b => !b.has_active_training && !b.has_active_rest).length})
                </button>
                <button
                  className={`filter-status-btn ${statusFilter === 'training' ? 'active' : ''}`}
                  onClick={() => setStatusFilter('training')}
                  aria-pressed={statusFilter === 'training'}
                >
                  Training ({stats.inTraining})
                </button>
                <button
                  className={`filter-status-btn ${statusFilter === 'resting' ? 'active' : ''}`}
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
            <div className="boxers-results-count">
              Showing {filteredBoxers.length} of {boxers.length} boxers
              {(searchQuery || statusFilter !== 'all') && (
                <button
                  onClick={() => { setSearchQuery(''); setStatusFilter('all'); }}
                  className="boxers-clear-filters"
                >
                  Clear Filters
                </button>
              )}
            </div>
          )}

          {/* Main Content Area */}
          {isLoading ? (
            <div className="boxers-loading">
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
            <div className="boxers-empty-state">
              <div className="empty-state-icon" aria-hidden="true">🥊</div>
              <h2 className="empty-state-title">Your roster is empty</h2>
              <p className="empty-state-description">
                Create your first boxer and begin their journey to championship glory.
                Train, fight, and build a legendary career in the ring.
              </p>
              <button
                onClick={() => navigate('/create-boxer')}
                className="btn-empty-cta"
              >
                <span aria-hidden="true">➕</span>
                Create Your First Boxer
              </button>
            </div>
          ) : (
            <section className="boxers-section" aria-label="Boxer roster">
              {filteredBoxers.length === 0 ? (
                <div className="boxers-no-results">
                  <div className="no-results-icon" aria-hidden="true">🔍</div>
                  <h3 className="no-results-title">No boxers found</h3>
                  <p className="no-results-message">
                    {searchQuery
                      ? `No boxers match "${searchQuery}". Try a different search term.`
                      : 'No boxers match the selected filters.'}
                  </p>
                  <button
                    onClick={() => { setSearchQuery(''); setStatusFilter('all'); }}
                    className="btn-clear-all"
                  >
                    Clear All Filters
                  </button>
                </div>
              ) : (
                <div className="boxers-grid" role="list">
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
        </div>
      </div>
    </div>
  );
};

// Simple inline stat card component
const StatSummaryCard = ({ icon, value, label }) => (
  <div className="stat-summary-card">
    <span className="stat-summary-icon" aria-hidden="true">{icon}</span>
    <span className="stat-summary-value">{value}</span>
    <span className="stat-summary-label">{label}</span>
  </div>
);

export default BoxersList;
