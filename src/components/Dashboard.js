import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import './Dashboard.css';
import TopBar from './TopBar.jsx';
import BoxerCard from './BoxerCard.jsx';
import TrainingControlPanel from './TrainingControlPanel.jsx';
import { ErrorDisplay } from './ErrorDisplay';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';

const Dashboard = ({ user, onLogout }) => {
  const navigate = useNavigate();
  const [boxers, setBoxers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentUser, setCurrentUser] = useState(null);
  const [showTrainingControlPanel, setShowTrainingControlPanel] = useState(false);
  const [worldTime, setWorldTime] = useState(null);
  const [fightData, setFightData] = useState({}); // boxerId -> fight mapping (MAT-106)

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
        boxersArray = Array.isArray(boxersData) ? boxersData : [];
        setBoxers(boxersArray);
      } else {
        setError(boxersData.error || 'Failed to load boxers');
      }

      // Fetch active fights and create a mapping (MAT-106)
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
    onLogout();
  }, [onLogout]);

  const handleRetry = useCallback(() => {
    if (currentUser) {
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
        <div className="empty-state-illustration" aria-hidden="true">🥊</div>
        <h2 className="empty-state-title">Your roster is empty</h2>
        <p className="empty-state-description">
          Create your first boxer and begin their journey to championship glory.
          Train, fight, and build a legendary career in the ring.
        </p>
        <Link to="/create-boxer" className="empty-state-cta">
          <span className="cta-icon" aria-hidden="true">➕</span>
          <span>Create Your First Boxer</span>
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

  const boxerCount = boxers.length;

  // Calculate quick stats for roster summary
  const stats = {
    totalBoxers: boxerCount,
    inTraining: boxers.filter(b => b.has_active_training).length,
    resting: boxers.filter(b => b.has_active_rest).length,
    fightsScheduled: Object.keys(fightData).filter(id => fightData[id]?.status === 'scheduled').length,
    averageLevel: boxerCount > 0 ? Math.round(boxers.reduce((sum, b) => sum + (b.level || 1), 0) / boxerCount) : 0,
    totalWins: boxers.reduce((sum, b) => sum + (b.wins || 0), 0),
    totalLosses: boxers.reduce((sum, b) => sum + (b.losses || 0), 0),
    totalDraws: boxers.reduce((sum, b) => sum + (b.draws || 0), 0),
  };

  return (
    <div className="dashboard">
      <TopBar currentUser={currentUser} onLogout={handleLogout} />

      <main className="dashboard-content">
        <header className="dashboard-header">
          <div className="header-left">
            <h1 className="page-title">My Boxers</h1>
          </div>
        </header>

        {/* Quick Stats Summary */}
        {boxerCount > 0 && (
          <section className="quick-stats" aria-label="Roster summary">
            <div className="stat-summary-card">
              <span className="stat-summary-icon" aria-hidden="true">🥊</span>
              <span className="stat-summary-value">{stats.totalBoxers}</span>
              <span className="stat-summary-label">Total Boxers</span>
            </div>
            <div className="stat-summary-card">
              <span className="stat-summary-icon" aria-hidden="true">💪</span>
              <span className="stat-summary-value">{stats.inTraining}</span>
              <span className="stat-summary-label">In Training</span>
            </div>
            <div className="stat-summary-card">
              <span className="stat-summary-icon" aria-hidden="true">🛌</span>
              <span className="stat-summary-value">{stats.resting}</span>
              <span className="stat-summary-label">Resting</span>
            </div>
            <div className="stat-summary-card">
              <span className="stat-summary-icon" aria-hidden="true">📅</span>
              <span className="stat-summary-value">{stats.fightsScheduled}</span>
              <span className="stat-summary-label">Fights Scheduled</span>
            </div>
            <div className="stat-summary-card stat-summary-card-wide">
              <span className="stat-summary-subtitle">Career Record: {stats.totalWins}W - {stats.totalLosses}L - {stats.totalDraws}D | Avg Level: {stats.averageLevel}</span>
            </div>
          </section>
        )}

        <section className="boxers-section" aria-label="Boxer roster">
          {isLoading ? (
            renderSkeletonCards()
          ) : error ? (
            renderErrorState()
          ) : boxers.length === 0 ? (
            renderEmptyState()
          ) : (
            <div className="boxers-grid" role="list">
              {boxers.map((boxer) => (
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

        {showTrainingControlPanel && (
          <TrainingControlPanel
            boxers={boxers}
            onClose={() => setShowTrainingControlPanel(false)}
            onRefreshBoxers={() => loadUserBoxers(currentUser.id)}
          />
        )}
      </main>
    </div>
  );
};

export default Dashboard;
