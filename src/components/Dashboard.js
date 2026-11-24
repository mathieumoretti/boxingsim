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
import { SkipLink } from './SkipLink';
import { MobileNavigationDrawer } from './MobileNavigationDrawer';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';

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
      <div className="min-h-screen bg-gray-50">
        <SkipLink targetId="main-content" />
        <TopBar currentUser={currentUser} onLogout={handleLogout} />
        <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
          <LoadingSpinner size="lg" aria-label="Loading dashboard" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50">
        <SkipLink targetId="main-content" />
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
    <div className="min-h-screen bg-gray-50">
      <SkipLink targetId="main-content" />
      <TopBar currentUser={currentUser} onLogout={handleLogout} />

      <main id="main-content" className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 focus:outline-none" tabIndex="-1">
        {/* Quick Stats Row */}
        <section className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8" aria-label="Quick statistics">
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
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left Column: Notifications & Upcoming Fights */}
          <div className="lg:col-span-2 space-y-6">
            {/* Notifications Panel */}
            <section className="bg-white rounded-xl shadow-[0_2px_8px_rgba(0,0,0,0.04)] overflow-hidden" aria-label="Notifications">
              <header className="flex items-center justify-between p-5 bg-gradient-to-br from-gray-50 to-slate-100 border-b border-gray-200">
                <h2 className="text-lg font-bold text-gray-800 m-0 flex items-center gap-2">
                  <span aria-hidden="true">🔔</span>
                  Notifications
                </h2>
                {notifications.length > 0 && (
                  <span className="inline-flex items-center justify-center w-8 h-8 rounded-full text-xs font-bold bg-danger text-white">
                    {notifications.length}
                  </span>
                )}
              </header>

              <div className="p-5">
                {notifications.length === 0 ? (
                  <EmptyState
                    icon="🎉"
                    title="All Clear!"
                    message="No notifications at the moment."
                  />
                ) : (
                  <ul className="space-y-3">
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
            <section className="bg-white rounded-xl shadow-[0_2px_8px_rgba(0,0,0,0.04)] overflow-hidden" aria-label="Upcoming fights">
              <header className="flex items-center justify-between p-5 bg-gradient-to-br from-gray-50 to-slate-100 border-b border-gray-200">
                <h2 className="text-lg font-bold text-gray-800 m-0 flex items-center gap-2">
                  <span aria-hidden="true">🥊</span>
                  Upcoming Fights
                </h2>
                {upcomingFights.length > 0 && (
                  <Link to="/boxers" className="text-sm font-semibold text-primary hover:text-primary-light transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2">View All →</Link>
                )}
              </header>

              <div className="p-5">
                {upcomingFights.length === 0 ? (
                  <EmptyState
                    icon="📅"
                    title="No Fights Scheduled"
                    message="No fights scheduled for the next 7 days."
                  />
                ) : (
                  <ul className="space-y-3">
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
          <div className="space-y-6">
            <section className="bg-white rounded-xl shadow-[0_2px_8px_rgba(0,0,0,0.04)] overflow-hidden" aria-label="Recent activity">
              <header className="p-5 bg-gradient-to-br from-gray-50 to-slate-100 border-b border-gray-200">
                <h2 className="text-lg font-bold text-gray-800 m-0 flex items-center gap-2">
                  <span aria-hidden="true">📜</span>
                  Recent Activity
                </h2>
              </header>

              <div className="p-5">
                {recentActivity.length === 0 ? (
                  <EmptyState
                    icon="📭"
                    title="No Recent Activity"
                    message="Fight results will appear here."
                  />
                ) : (
                  <ul className="space-y-3">
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
            <section className="bg-white rounded-xl shadow-[0_2px_8px_rgba(0,0,0,0.04)] overflow-hidden" aria-label="Quick actions">
              <header className="p-5 bg-gradient-to-br from-gray-50 to-slate-100 border-b border-gray-200">
                <h2 className="text-lg font-bold text-gray-800 m-0">Quick Actions</h2>
              </header>

              <div className="p-5 space-y-3">
                <Link to="/create-boxer" className="flex items-center gap-2 w-full py-2.5 px-4 bg-gradient-to-br from-primary to-primary-dark text-white rounded-lg font-semibold transition-all duration-150 hover:bg-gradient-to-br hover:from-primary-light hover:to-primary hover:-translate-y-0.5 hover:shadow-[0_4px_8px_rgba(124,58,237,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2">
                  <span aria-hidden="true">➕</span>
                  Create Fighter
                </Link>

                <Link to="/boxers" className="flex items-center gap-2 w-full py-2.5 px-4 bg-gray-100 text-gray-700 rounded-lg font-semibold transition-all duration-150 hover:bg-gray-200 hover:-translate-y-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-gray-400 focus-visible:outline-offset-2">
                  <span aria-hidden="true">🥊</span>
                  View All Boxers
                </Link>

                <Link to="/rankings" className="flex items-center gap-2 w-full py-2.5 px-4 bg-gray-100 text-gray-700 rounded-lg font-semibold transition-all duration-150 hover:bg-gray-200 hover:-translate-y-0.5 focus-visible:outline focus-visible:outline-2 focus-visible:outline-gray-400 focus-visible:outline-offset-2">
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

const StatCard = ({ icon, value, label, href }) => {
  const isCurrentPage = href && window.location.href.includes(href);

  return (
    <div className={`bg-white rounded-xl p-4 shadow-[0_2px_8px_rgba(0,0,0,0.04)] border border-gray-100 ${href && !isCurrentPage ? 'cursor-pointer hover:shadow-[0_4px_12px_rgba(0,0,0,0.08)] hover:-translate-y-0.5 transition-all duration-150' : ''}`}>
      {href && !isCurrentPage ? (
        <a href={href} className="block" onClick={(e) => e.preventDefault()}>
          <span className="text-2xl mb-1 block" aria-hidden="true">{icon}</span>
          <span className="text-2xl font-bold text-gray-900 block mb-0.5">{value}</span>
          <span className="text-sm text-gray-600">{label}</span>
        </a>
      ) : (
        <>
          <span className="text-2xl mb-1 block" aria-hidden="true">{icon}</span>
          <span className="text-2xl font-bold text-gray-900 block mb-0.5">{value}</span>
          <span className="text-sm text-gray-600">{label}</span>
        </>
      )}
    </div>
  );
};

const EmptyState = ({ icon, title, message }) => (
  <div className="text-center py-8">
    <div className="text-4xl mb-2" aria-hidden="true">{icon}</div>
    <h3 className="m-0 mb-1 text-sm font-semibold text-gray-700">{title}</h3>
    <p className="m-0 text-xs text-gray-500">{message}</p>
  </div>
);

const NotificationItem = ({ notification, worldTime }) => {
  const isUrgent = notification.priority === 'high';

  return (
    <li className={`p-3 rounded-lg border ${isUrgent ? 'bg-red-50 border-red-200' : 'bg-blue-50 border-blue-200'}`}>
      <div className="flex items-start gap-2">
        <span className="text-sm flex-shrink-0" aria-hidden="true">
          {isUrgent ? '⚠️' : 'ℹ️'}
        </span>
        <span className={`flex-1 text-sm ${isUrgent ? 'text-red-900' : 'text-blue-900'}`}>{notification.message}</span>
        {notification.scheduled_time && (
          <TimeUntil datetime={notification.scheduled_time} worldTime={worldTime} />
        )}
      </div>
    </li>
  );
};

const FightItem = ({ fight, worldTime }) => (
  <li className="p-3 rounded-lg bg-gray-50 border border-gray-200">
    <div className="flex items-center justify-between">
      <span className="text-sm font-semibold text-gray-700">Fight #{fight.id}</span>
      <TimeUntil datetime={fight.scheduled_time} worldTime={worldTime} />
    </div>
    {fight.round && (
      <span className="mt-1 inline-block text-xs text-gray-500">{fight.round} rounds</span>
    )}
  </li>
);

const ActivityItem = ({ fight }) => (
  <li className="p-3 rounded-lg bg-green-50 border border-green-200">
    <div className="flex items-center gap-2">
      <span className="text-sm flex-shrink-0" aria-hidden="true">✅</span>
      <span className="flex-1 text-sm font-semibold text-green-900">Fight #{fight.id} completed</span>
      <time className="text-xs text-green-700" dateTime={fight.scheduled_time}>
        {formatDate(fight.scheduled_time)}
      </time>
    </div>
  </li>
);

const TimeUntil = ({ datetime, worldTime }) => {
  if (!worldTime) return null;

  const target = new Date(datetime);
  const now = new Date(worldTime.game_time);
  const diff = target - now;

  if (diff <= 0) return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-danger text-white">Now</span>;

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 24) {
    const days = Math.floor(hours / 24);
    return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-gray-200 text-gray-700">In {days}d</span>;
  }

  if (hours > 0) {
    return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-orange-100 text-orange-800">In {hours}h</span>;
  }

  return <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-red-100 text-red-800">In {minutes}m</span>;
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
