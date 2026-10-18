import React, { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import './Dashboard.css';
import TopBar from './TopBar.jsx';
import BoxerCard from './BoxerCard.jsx';
import TrainingScheduler from './TrainingScheduler.jsx';
import TrainingControlPanel from './TrainingControlPanel.jsx';
import { ErrorDisplay } from './ErrorDisplay';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';
import { createFightMap } from '../utils/fights';

const Dashboard = ({ user, onLogout }) => {
  const navigate = useNavigate();
  const [boxers, setBoxers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentUser, setCurrentUser] = useState(null);
  const [selectedBoxerForTraining, setSelectedBoxerForTraining] = useState(null);
  const [showTrainingControlPanel, setShowTrainingControlPanel] = useState(false);
  const [worldTime, setWorldTime] = useState(null);
  const [fightData, setFightData] = useState({}); // boxerId -> fight mapping (MAT-106)
  const [errorRefresh, setErrorRefresh] = useState(null); // For error retry tracking

  useEffect(() => {
    // Get user from props or localStorage
    const currentUser = user || getUser();
    setCurrentUser(currentUser);

    if (currentUser) {
      loadUserBoxers(currentUser.id);
    } else {
      setIsLoading(false);
    }
  }, [user, loadUserBoxers]);

  // Fetch and refresh world time every 5 seconds for countdown displays
  useEffect(() => {
    const fetchWorldTime = async () => {
      try {
        const response = await authenticatedFetch(`${API_BASE_URL}/world/time`, { method: 'GET' });
        if (response.ok) {
          const data = await response.json();
          setWorldTime(data);
        }
      } catch (err) {
        // Silently fail - countdowns will just not update
      }
    };

    // Initial fetch
    fetchWorldTime();

    // Refresh every 5 seconds for real-time countdown updates
    const interval = setInterval(fetchWorldTime, 5000);

    return () => clearInterval(interval);
  }, []);

  // Handler called by BoxerCard when a boxer's state changes (training completed, rest ended)
  const handleBoxerStateChanged = useCallback((boxerId) => {
    // Refresh all boxer data to get updated has_active_rest and forced_rest_until values
    if (currentUser) {
      loadUserBoxers(currentUser.id);
    }
  }, [currentUser, loadUserBoxers]);

  const loadUserBoxers = useCallback(async (userId) => {
    setIsLoading(true);
    setError('');

    try {
      // Fetch boxers and active fights in parallel (MAT-106)
      const [boxersResponse, fightsResponse] = await Promise.all([
        authenticatedFetch(`${API_BASE_URL}/users/${userId}/boxers`, { method: 'GET' }),
        authenticatedFetch(`${API_BASE_URL}/fights/active`, { method: 'GET' })
      ]);

      const boxersData = await boxersResponse.json();
      let boxersArray = [];

      if (boxersResponse.ok) {
        // Ensure we always have an array, even if API returns null for empty results
        boxersArray = Array.isArray(boxersData) ? boxersData : [];
        setBoxers(boxersArray);
      } else {
        setError(boxersData.error || 'Failed to load boxers');
      }

      // Fetch active fights and create a mapping (MAT-106)
      let activeFights = [];
      if (fightsResponse.ok) {
        const fightsData = await fightsResponse.json();
        activeFights = Array.isArray(fightsData) ? fightsData : [];

        // Create fight map for boxers
        if (boxersArray.length > 0) {
          const boxerIds = boxersArray.map(b => b.id);

          // Create a mapping of boxerId -> next fight info
          const fightMapping = {};
          activeFights.forEach(fight => {
            // Only include fights involving this user's boxers
            if (boxerIds.includes(fight.boxer1_id)) {
              // This user's boxer is boxer1
              fightMapping[fight.boxer1_id] = {
                fight_id: fight.id,
                opponent_id: fight.boxer2_id,
                scheduled_time: fight.scheduled_time,
                status: fight.status,
                rounds: fight.round || 12
              };
            } else if (boxerIds.includes(fight.boxer2_id)) {
              // This user's boxer is boxer2
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
      // Don't show error if redirected to login (401 handled by authenticatedFetch)
      if (!error.message.includes('Unauthorized')) {
        setError('Network error: ' + error.message);
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleLogout = useCallback(() => {
    onLogout();
  }, [onLogout]);

  // Retry handler for error state (MAT-130)
  const handleRetry = useCallback(() => {
    if (currentUser) {
      setErrorRefresh(null); // Clear any previous retry tracking
      loadUserBoxers(currentUser.id);
    }
  }, [currentUser, loadUserBoxers]);

  // Render skeleton loader cards (MAT-130)
  const renderSkeletonCards = () => {
    return (
      <div className="boxers-grid">
        {[1, 2, 3].map((i) => (
          <div key={i} className="boxer-card-skeleton">
            <div className="skeleton-avatar"></div>
            <div className="skeleton-name-section">
              <div className="skeleton-name"></div>
              <div className="skeleton-nickname"></div>
            </div>
            <div className="skeleton-stat-bar" style={{ width: '80%' }}></div>
            <div className="skeleton-stat-bar" style={{ width: '75%' }}></div>
            <div className="skeleton-stat-bar" style={{ width: '70%' }}></div>
            <div className="skeleton-combat-stats">
              <div className="skeleton-stat-item"></div>
              <div className="skeleton-stat-item"></div>
              <div className="skeleton-stat-item"></div>
            </div>
            <div className="skeleton-card-actions">
              <div className="skeleton-btn"></div>
              <div className="skeleton-btn"></div>
              <div className="skeleton-btn"></div>
            </div>
          </div>
        ))}
      </div>
    );
  };

  // Render empty state (MAT-130)
  const renderEmptyState = () => {
    return (
      <div className="boxers-empty-state">
        <div className="empty-state-icon">🥊</div>
        <h3>No Boxers Yet</h3>
        <p>Create your first boxer and start their journey to the championship!</p>
        <Link to="/create-boxer" className="empty-state-cta">
          Create Your First Boxer
        </Link>
      </div>
    );
  };

  // Render error state (MAT-130)
  const renderErrorState = () => {
    return (
      <ErrorDisplay
        error={error}
        onRetry={handleRetry}
        variant="alert"
        title="Failed to Load Boxers"
        icon={true}
      />
    );
  };

  return (
    <div className="dashboard">
      <TopBar currentUser={currentUser} onLogout={handleLogout} />

      <main className="dashboard-content">
        <section className="boxer-section">
          <div className="section-header">
            <h2>Your Boxers</h2>
            <div className="header-actions">
              <Link to="/create-boxer">
                <button className="create-boxer-btn">Create New Boxer</button>
              </Link>
              {/* Development-only training control - hidden in production */}
              {process.env.NODE_ENV === 'development' && (
                <button
                  className="dev-control-btn"
                  onClick={() => setShowTrainingControlPanel(true)}
                  title="Development-only: Manually complete training sessions"
                >
                  ⚙️ Training Control
                </button>
              )}
            </div>
          </div>

          {isLoading ? (
            renderSkeletonCards()
          ) : error ? (
            renderErrorState()
          ) : boxers.length === 0 ? (
            renderEmptyState()
          ) : (
            <div className="boxers-grid">
              {boxers.map((boxer) => (
                <BoxerCard
                  key={boxer.id}
                  boxer={boxer}
                  worldTime={worldTime}
                  upcomingFight={fightData[boxer.id]}
                  onOpenTraining={() => setSelectedBoxerForTraining(boxer)}
                  onOpenFightBooking={() => navigate(`/boxer/${boxer.id}/book-fight`)}
                  onOpenFightHistory={() => navigate(`/boxer/${boxer.id}/fights`)}
                  onBoxerStateChanged={handleBoxerStateChanged}
                />
              ))}
            </div>
          )}
        </section>

        {/* Centralized Training Modal */}
        {selectedBoxerForTraining && (
          <div className="training-modal-overlay" onClick={() => setSelectedBoxerForTraining(null)}>
            <div className="training-modal" onClick={(e) => e.stopPropagation()}>
              <button
                className="modal-close"
                onClick={() => setSelectedBoxerForTraining(null)}
                aria-label="Close modal"
              >×</button>
              <TrainingScheduler
                boxerId={selectedBoxerForTraining.id}
                boxer={selectedBoxerForTraining}
                onClose={() => setSelectedBoxerForTraining(null)}
                onTrainingScheduled={() => {
                  loadUserBoxers(currentUser.id);
                  setSelectedBoxerForTraining(null);
                }}
              />
            </div>
          </div>
        )}

        {/* Training Control Panel - Development Only */}
        {showTrainingControlPanel && (
          <TrainingControlPanel
            boxers={boxers}
            onClose={() => setShowTrainingControlPanel(false)}
            onRefreshBoxers={() => loadUserBoxers(currentUser.id)}
          />
        )}

        <section className="fight-section">
          <h2>Fight Arena</h2>
          <div className="fight-arena">
            <div className="boxer-display">
              <h4>Boxer 1</h4>
              <div className="boxer-stats">
                <p>Health: <span id="boxer1-health">0</span></p>
                <p>Energy: <span id="boxer1-energy">0</span></p>
                <p>Strength: <span id="boxer1-strength">0</span></p>
              </div>
            </div>
            <div className="fight-controls">
              <button className="fight-btn">Start Fight</button>
              <button className="reset-btn">Reset</button>
            </div>
            <div className="boxer-display">
              <h4>Boxer 2</h4>
              <div className="boxer-stats">
                <p>Health: <span id="boxer2-health">0</span></p>
                <p>Energy: <span id="boxer2-energy">0</span></p>
                <p>Strength: <span id="boxer2-strength">0</span></p>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Dashboard;
