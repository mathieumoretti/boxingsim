/**
 * BoxerProfile - Comprehensive boxer profile page (MAT-148)
 *
 * Aggregates all boxer-related views: stats, fight history, training history,
 * and recent activity in a single dedicated page at /boxer/:id
 */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import TopBar from './TopBar';
import PageLayout from './PageLayout';
import BoxerHeader from './BoxerHeader';
import CareerRecord from './CareerRecord';
import StatsOverview from './StatsOverview';
import CurrentStatus from './CurrentStatus';
import RecentActivity from './RecentActivity';
import { ErrorDisplay } from './ErrorDisplay';
import { fetchBoxer, fetchUpcomingFight } from '../api/fights';
import { fetchWorldTime } from '../api/world';
import { getUserErrorMessage } from '../utils/errorMessages';
import NextFight from './NextFight';
import './BoxerProfile.css';

const BoxerProfile = ({ user, onLogout }) => {
  const { id: boxerIdParam } = useParams();
  const navigate = useNavigate();

  const [boxer, setBoxer] = useState(null);
  const [upcomingFight, setUpcomingFight] = useState(null);
  const [worldTime, setWorldTime] = useState(null);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadProfileData();

    // Refresh boxer data periodically to show updates
    const refreshInterval = setInterval(() => {
      if (boxerIdParam) {
        loadBoxerData(boxerIdParam);
        loadUpcomingFight(boxerIdParam);
        loadWorldTime();
      }
    }, 10000); // Refresh every 10 seconds

    return () => clearInterval(refreshInterval);
  }, [boxerIdParam]);

  const loadProfileData = async () => {
    if (!boxerIdParam) {
      setError('No boxer selected');
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Load boxer, upcoming fight, and world time in parallel
      await Promise.all([
        loadBoxerData(boxerIdParam),
        loadUpcomingFight(boxerIdParam),
        loadWorldTime()
      ]);
    } catch (err) {
      setError(getUserErrorMessage(err, 'general'));
    } finally {
      setIsLoading(false);
    }
  };

  const loadBoxerData = async (id) => {
    try {
      const data = await fetchBoxer(id);
      if (data && data.id) {
        setBoxer(data);
      } else {
        setError('Boxer not found');
        setBoxer(null);
      }
    } catch (err) {
      // Don't set error here if we're just refreshing - let the main loader handle it
      if (!boxer) {
        console.error('Failed to load boxer:', err);
      }
    }
  };

  const loadWorldTime = async () => {
    try {
      const timeData = await fetchWorldTime();
      setWorldTime(timeData);
    } catch (err) {
      // World time is optional - log but don't fail
      console.debug('Failed to load world time:', err.message);
    }
  };

  const loadUpcomingFight = async (id) => {
    try {
      const fightData = await fetchUpcomingFight(id);
      setUpcomingFight(fightData);
    } catch (err) {
      // 404 means no upcoming fight - that's fine, just log
      if (err.status !== 404) {
        console.debug('Failed to load upcoming fight:', err.message);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="boxer-profile">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout>
          <div className="boxer-profile-content">
            <div className="profile-loading">
              <div className="loading-spinner"></div>
              <p>Loading boxer profile...</p>
            </div>
          </div>
        </PageLayout>
      </div>
    );
  }

  if (error && !boxer) {
    return (
      <div className="boxer-profile">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <ErrorDisplay
            error={error}
            variant="alert"
            title="Failed to Load Profile"
            icon={true}
          />
        </PageLayout>
      </div>
    );
  }

  if (!boxer) {
    return (
      <div className="boxer-profile">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <div className="boxer-profile-content">
            <div className="profile-not-found">
              <h2>Boxer Not Found</h2>
              <p>The boxer you're looking for doesn't exist.</p>
            </div>
          </div>
        </PageLayout>
      </div>
    );
  }

  return (
    <div className="boxer-profile">
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
      >
        <div className="boxer-profile-content">
          {/* Boxer Header - Avatar, Name, Level, XP */}
          <BoxerHeader boxer={boxer} />

          {/* Career Record - W/L/D/KO */}
          <CareerRecord boxer={boxer} />

          {/* Next Fight Card (only shows when there's an upcoming fight) */}
          <NextFight upcomingFight={upcomingFight} boxerId={boxer.id} gameTime={worldTime?.current_game_time} />

          {/* Stats Overview - STR/DEF/AGI */}
          <StatsOverview boxer={boxer} />

          {/* Current Status - Health/Energy/Fatigue */}
          <CurrentStatus boxer={boxer} worldTime={worldTime} />

          {/* Recent Activity Timeline */}
          <RecentActivity boxerId={boxer.id} gameTime={worldTime?.current_game_time} />

          {/* Action Buttons */}
          <div className="profile-actions">
            <Link to={`/boxer/${boxer.id}/fights`} className="action-btn action-primary">
              📜 View Fight History
            </Link>
            <Link to={`/boxer/${boxer.id}/train`} className="action-btn action-secondary">
              🏋️ Train Boxer
            </Link>
            <Link to={`/boxer/${boxer.id}/book-fight`} className="action-btn action-accent">
              🥊 Book Fight
            </Link>
          </div>
        </div>
      </PageLayout>
    </div>
  );
};

export default BoxerProfile;
