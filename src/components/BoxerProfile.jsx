/**
 * BoxerProfile - Comprehensive boxer profile page (MAT-148)
 *
 * Aggregates all boxer-related views: stats, fight history, training history,
 * and recent activity in a single dedicated page at /boxer/:id
 */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import TopBar from './TopBar';
import BoxerHeader from './BoxerHeader';
import CareerRecord from './CareerRecord';
import StatsOverview from './StatsOverview';
import CurrentStatus from './CurrentStatus';
import RecentActivity from './RecentActivity';
import { ErrorDisplay } from './ErrorDisplay';
import { fetchBoxer } from '../api/fights';
import { fetchWorldTime } from '../api/world';
import { getUserErrorMessage } from '../utils/errorMessages';
import './BoxerProfile.css';

const BoxerProfile = ({ user, onLogout }) => {
  const { id: boxerIdParam } = useParams();
  const navigate = useNavigate();

  const [boxer, setBoxer] = useState(null);
  const [worldTime, setWorldTime] = useState(null);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadProfileData();

    // Refresh boxer data periodically to show updates
    const refreshInterval = setInterval(() => {
      if (boxerIdParam) {
        loadBoxerData(boxerIdParam);
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
      // Load boxer and world time in parallel
      await Promise.all([
        loadBoxerData(boxerIdParam),
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

  const handleBack = () => {
    navigate('/dashboard');
  };

  if (isLoading) {
    return (
      <div className="boxer-profile">
        <TopBar currentUser={user} onLogout={onLogout} />
        <main className="boxer-profile-content">
          <div className="profile-loading">
            <div className="loading-spinner"></div>
            <p>Loading boxer profile...</p>
          </div>
        </main>
      </div>
    );
  }

  if (error && !boxer) {
    return (
      <div className="boxer-profile">
        <TopBar currentUser={user} onLogout={onLogout} />
        <main className="boxer-profile-content">
          <ErrorDisplay
            error={error}
            variant="alert"
            title="Failed to Load Profile"
            icon={true}
          />
          <div className="profile-actions-error">
            <button onClick={handleBack} className="back-btn">
              ← Back to Dashboard
            </button>
          </div>
        </main>
      </div>
    );
  }

  if (!boxer) {
    return (
      <div className="boxer-profile">
        <TopBar currentUser={user} onLogout={onLogout} />
        <main className="boxer-profile-content">
          <div className="profile-not-found">
            <h2>Boxer Not Found</h2>
            <p>The boxer you're looking for doesn't exist.</p>
          </div>
          <button onClick={handleBack} className="back-btn">
            ← Back to Dashboard
          </button>
        </main>
      </div>
    );
  }

  return (
    <div className="boxer-profile">
      <TopBar currentUser={user} onLogout={onLogout} />

      <main className="boxer-profile-content">
        {/* Navigation Header */}
        <div className="profile-nav">
          <button onClick={handleBack} className="back-btn" aria-label="Back to dashboard">
            ← Back to Dashboard
          </button>
        </div>

        {/* Boxer Header - Avatar, Name, Level, XP */}
        <BoxerHeader boxer={boxer} />

        {/* Career Record - W/L/D/KO */}
        <CareerRecord boxer={boxer} />

        {/* Stats Overview - STR/DEF/AGI */}
        <StatsOverview boxer={boxer} />

        {/* Current Status - Health/Energy/Fatigue */}
        <CurrentStatus boxer={boxer} worldTime={worldTime} />

        {/* Recent Activity Timeline */}
        <RecentActivity boxerId={boxer.id} />

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
      </main>
    </div>
  );
};

export default BoxerProfile;
