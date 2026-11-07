/**
 * Dashboard Component (Refactored for MAT-122)
 *
 * Quick-view dashboard showing:
 * - Notifications (boxers needing attention)
 * - Upcoming fights (today/this week)
 * - Recent activity feed
 * - Quick stats summary
 *
 * NOT a roster management page - that's the /boxers route.
 */

import React, { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import TopBar from './TopBar';
import MiniWorldClock from './MiniWorldClock';
import LoadingSpinner from './ui/LoadingSpinner';
import { ErrorDisplay } from './ErrorDisplay';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';
import './Dashboard.css';

const Dashboard = ({ user }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [notifications, setNotifications] = useState([]);
  const [upcomingFights, setUpcomingFights] = useState([]);
  const [recentActivity, setRecentActivity] = useState([]);
  const [quickStats, setQuickStats] = useState({});
  const [worldTime, setWorldTime] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const currentUser = user || getUser();
    setCurrentUser(currentUser);

    if (currentUser) {
      loadData(currentUser.id);
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

  const loadData = useCallback(async (userId) => {
    setIsLoading(true);
    setError('');

    try {
      // Fetch all dashboard data in parallel
      const [boxersResponse, fightsResponse] = await Promise.all([
        authenticatedFetch(`${API_BASE_URL}/users/${userId}/boxers`, { method: 'GET' }),
        authenticatedFetch(`${API_BASE_URL}/fights/active`, { method: 'GET' })
      ]);

      const boxersData = await boxersResponse.json();
      const fightsData = await fightsResponse.json();

      const boxers = Array.isArray(boxersData) ? boxersData : [];
      const fights = Array.isArray(fightsData) ? fightsData : [];

      // Calculate notifications (boxers needing attention)
      const now = new Date(worldTime?.game_time || new Date());
      const oneHourFromNow = new Date(now.getTime() + 60 * 60 * 1000);
      const sevenDaysFromNow = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);

      const dashboardNotifications = [];

      // Check for upcoming fights (within 24 hours)
      fights.forEach(fight => {
        const fighterBoxerId = fight.boxer1_id === getCurrentBoxerId(userId, fight)
          ? fight.boxer1_id
          : (fight.boxer2_id === getCurrentBoxerId(userId, fight) ? fight.boxer2_id : null);

        if (!fighterBoxerId) return;

        const scheduledTime = new Date(fight.scheduled_time);

        if (scheduledTime <= oneHourFromNow && scheduledTime >= now) {
          dashboardNotifications.push({
            type: 'fight_urgent',
            boxer_id: fighterBoxerId,
            message: `Fight in less than 1 hour!`,
            fight_id: fight.id,
            scheduled_time: fight.scheduled_time,
            priority: 'high'
          });
        } else if (scheduledTime <= sevenDaysFromNow && scheduledTime >= now) {
          dashboardNotifications.push({
            type: 'fight_upcoming',
            boxer_id: fighterBoxerId,
            message: `Fight scheduled for ${formatDate(scheduledTime)}`,
            fight_id: fight.id,
            scheduled_time: fight.scheduled_time,
            priority: 'medium'
          });
        }
      });

      // Check for boxers with active training/rest
      boxers.forEach(boxer => {
        if (boxer.has_active_training && boxer.training_type === 'rest') {
          // Resting boxer notification could be added here
        }
      });

      setNotifications(dashboardNotifications);

      // Filter upcoming fights (next 7 days) for dashboard
      const userFightIds = new Set();
      fights.forEach(fight => {
        if (fight.boxer1_id && boxers.some(b => b.id === fight.boxer1_id)) {
          userFightIds.add(fight.boxer1_id);
        }
        if (fight.boxer2_id && boxers.some(b => b.id === fight.boxer2_id)) {
          userFightIds.add(fight.boxer2_id);
        }
      });

      const upcoming = fights
        .filter(fight => {
          const scheduledTime = new Date(fight.scheduled_time);
          return scheduledTime >= now && scheduledTime <= sevenDaysFromNow;
        })
        .slice(0, 5); // Show max 5 upcoming fights

      setUpcomingFights(upcoming);

      // Recent activity (last 5 fights completed)
      const recent = fights
        .filter(fight => fight.status === 'completed')
        .sort((a, b) => new Date(b.scheduled_time) - new Date(a.scheduled_time))
        .slice(0, 5);

      setRecentActivity(recent);

      // Quick stats summary
      const stats = {
        totalBoxers: boxers.length,
        inTraining: boxers.filter(b => b.has_active_training && b.training_type !== 'rest').length,
        resting: boxers.filter(b => b.has_active_rest).length,
        fightsScheduled: upcomingFights.length || 0,
      };

      setQuickStats(stats);

    } catch (error) {
      if (!error.message.includes('Unauthorized')) {
        setError('Failed to load dashboard data');
      }
    } finally {
      setIsLoading(false);
    }
  }, [worldTime]);

  const handleLogout = () => {
    window.location.replace('/login');
  };

  if (isLoading) {
    return (
      <div className="dashboard-page">
        <TopBar currentUser={currentUser} onLogout={handleLogout} />
        <div className="dashboard-loading">
          <LoadingSpinner size="lg" aria-label="Loading dashboard" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard-page">
        <TopBar currentUser={currentUser} onLogout={handleLogout} />
        <ErrorDisplay
          error={error}
          onRetry={() => loadData(currentUser?.id)}
          variant="alert"
          title="Failed to Load Dashboard"
          icon={true}
        />
      </div>
    );
  }

  return (
    <div className="dashboard-page">
      <TopBar currentUser={currentUser} onLogout={handleLogout} />

      <main className="dashboard-main">
        {/* Quick Stats Row */}
        <section className="dashboard-stats" aria-label="Quick statistics">
          <StatCard
            icon="📊"
            value={quickStats.totalBoxers || 0}
            label="Total Boxers"
            linkTo="/boxers"
            href="/boxers"
          />
          <StatCard
            icon="💪"
            value={quickStats.inTraining || 0}
            label="In Training"
          />
          <StatCard
            icon="🛌"
            value={quickStats.resting || 0}
            label="Resting"
          />
          <StatCard
            icon="📅"
            value={quickStats.fightsScheduled || 0}
            label="Fights This Week"
          />
        </section>

        {/* Two Column Layout */}
        <div className="dashboard-content-grid">
          {/* Left Column: Notifications & Upcoming Fights */}
          <div className="dashboard-left-column">
            {/* Notifications Panel */}
            <section className="dashboard-panel notifications-panel" aria-label="Notifications">
              <header className="panel-header">
                <h2 className="panel-title">
                  <span aria-hidden="true">🔔</span>
                  Notifications
                </h2>
                {notifications.length > 0 && (
                  <span className="notification-count">{notifications.length}</span>
                )}
              </header>

              <div className="panel-content">
                {notifications.length === 0 ? (
                  <EmptyState
                    icon="🎉"
                    title="All Clear!"
                    message="No notifications at the moment."
                  />
                ) : (
                  <ul className="notification-list">
                    {notifications.map((notif, index) => (
                      <NotificationItem
                        key={index}
                        notification={notif}
                        worldTime={worldTime}
                      />
                    ))}
                  </ul>
                )}
              </div>
            </section>

            {/* Upcoming Fights Panel */}
            <section className="dashboard-panel fights-panel" aria-label="Upcoming fights">
              <header className="panel-header">
                <h2 className="panel-title">
                  <span aria-hidden="true">🥊</span>
                  Upcoming Fights
                </h2>
                {upcomingFights.length > 0 && (
                  <Link to="/boxers" className="panel-link">View All →</Link>
                )}
              </header>

              <div className="panel-content">
                {upcomingFights.length === 0 ? (
                  <EmptyState
                    icon="📅"
                    title="No Fights Scheduled"
                    message="No fights scheduled for the next 7 days."
                  />
                ) : (
                  <ul className="fights-list">
                    {upcomingFights.map(fight => (
                      <FightItem
                        key={fight.id}
                        fight={fight}
                        worldTime={worldTime}
                      />
                    ))}
                  </ul>
                )}
              </div>
            </section>
          </div>

          {/* Right Column: Recent Activity */}
          <div className="dashboard-right-column">
            <section className="dashboard-panel activity-panel" aria-label="Recent activity">
              <header className="panel-header">
                <h2 className="panel-title">
                  <span aria-hidden="true">📜</span>
                  Recent Activity
                </h2>
              </header>

              <div className="panel-content">
                {recentActivity.length === 0 ? (
                  <EmptyState
                    icon="📭"
                    title="No Recent Activity"
                    message="Fight results will appear here."
                  />
                ) : (
                  <ul className="activity-list">
                    {recentActivity.map(fight => (
                      <ActivityItem
                        key={fight.id}
                        fight={fight}
                      />
                    ))}
                  </ul>
                )}
              </div>
            </section>

            {/* Quick Actions Panel */}
            <section className="dashboard-panel actions-panel" aria-label="Quick actions">
              <header className="panel-header">
                <h2 className="panel-title">Quick Actions</h2>
              </header>

              <div className="panel-content actions-content">
                <Link to="/create-boxer" className="action-btn action-btn--primary">
                  <span aria-hidden="true">➕</span>
                  Create Fighter
                </Link>

                <Link to="/boxers" className="action-btn action-btn--secondary">
                  <span aria-hidden="true">🥊</span>
                  View All Boxers
                </Link>

                <Link to="/rankings" className="action-btn action-btn--secondary">
                  <span aria-hidden="true">🏆</span>
                  View Rankings
                </Link>
              </div>
            </section>
          </div>
        </div>
      </main>
    </div>
  );
};

// Helper Components

const StatCard = ({ icon, value, label, href }) => (
  <div className={`stat-card ${href ? 'stat-card--link' : ''}`}>
    {href && !window.location.href.includes(href) ? (
      <a href={href} className="stat-card-link" onClick={(e) => e.preventDefault()}>
        <span className="stat-icon" aria-hidden="true">{icon}</span>
        <span className="stat-value">{value}</span>
        <span className="stat-label">{label}</span>
      </a>
    ) : (
      <>
        <span className="stat-icon" aria-hidden="true">{icon}</span>
        <span className="stat-value">{value}</span>
        <span className="stat-label">{label}</span>
      </>
    )}
  </div>
);

const EmptyState = ({ icon, title, message }) => (
  <div className="empty-state">
    <div className="empty-icon" aria-hidden="true">{icon}</div>
    <h3 className="empty-title">{title}</h3>
    <p className="empty-message">{message}</p>
  </div>
);

const NotificationItem = ({ notification, worldTime }) => {
  const isUrgent = notification.priority === 'high';

  return (
    <li className={`notification-item ${isUrgent ? 'notification-item--urgent' : ''}`}>
      <div className="notification-content">
        <span className="notification-icon" aria-hidden="true">
          {isUrgent ? '⚠️' : 'ℹ️'}
        </span>
        <span className="notification-message">{notification.message}</span>
      </div>
      {notification.scheduled_time && (
        <TimeUntil datetime={notification.scheduled_time} worldTime={worldTime} />
      )}
    </li>
  );
};

const FightItem = ({ fight, worldTime }) => (
  <li className="fight-item">
    <div className="fight-info">
      <span className="fight-boxers">Fight #{fight.id}</span>
      <TimeUntil datetime={fight.scheduled_time} worldTime={worldTime} />
    </div>
    {fight.round && (
      <span className="fight-rounds">{fight.round} rounds</span>
    )}
  </li>
);

const ActivityItem = ({ fight }) => (
  <li className="activity-item">
    <span className="activity-icon" aria-hidden="true">✅</span>
    <span className="activity-text">Fight #{fight.id} completed</span>
    <time className="activity-time" dateTime={fight.scheduled_time}>
      {formatDate(fight.scheduled_time)}
    </time>
  </li>
);

const TimeUntil = ({ datetime, worldTime }) => {
  if (!worldTime) return null;

  const target = new Date(datetime);
  const now = new Date(worldTime.game_time);
  const diff = target - now;

  if (diff <= 0) return <span className="time-countdown time-countdown--now">Now</span>;

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return <span className="time-countdown">In {days}d</span>;
  }

  if (hours > 0) {
    return <span className="time-countdown">In {hours}h</span>;
  }

  return <span className="time-countdown time-countdown--urgent">In {minutes}m</span>;
};

// Utility functions

function getCurrentBoxerId(userId, fight) {
  // Simplified - in reality would check which boxer belongs to user
  return null;
}

function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

export default Dashboard;
