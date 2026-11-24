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
import { SkipLink } from './SkipLink';
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

  // Section card background style
  const sectionCardBase = 'bg-gradient-to-br from-white/20 to-white/5 backdrop-blur-sm border border-white/10 rounded-xl p-6 mb-6';

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
        <SkipLink targetId="profile-content" />
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout>
          <main id="profile-content" className="max-w-[1000px] mx-auto px-5 py-6 pb-[12rem] min-h-screen" tabIndex="-1">
            <div className="flex flex-col items-center justify-center min-h-[400px] gap-5">
              <div className="w-[50px] h-[50px] border-4 border-white/10 border-t-victory-gold rounded-full animate-[spin_1s_linear_infinite]" />
              <p className="text-white/60 text-sm">Loading boxer profile...</p>
            </div>
          </main>
        </PageLayout>
      </div>
    );
  }

  if (error && !boxer) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
        <SkipLink targetId="profile-content" />
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <main id="profile-content" className="max-w-[1000px] mx-auto px-5 py-6 pb-[12rem]" tabIndex="-1">
            <ErrorDisplay
              error={error}
              variant="alert"
              title="Failed to Load Profile"
              icon={true}
            />
          </main>
        </PageLayout>
      </div>
    );
  }

  if (!boxer) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
        <SkipLink targetId="profile-content" />
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <main id="profile-content" className="max-w-[1000px] mx-auto px-5 py-6 pb-[12rem]" tabIndex="-1">
            <div className="text-center py-12 px-5">
              <h2 className="text-white text-3xl m-0 mb-2">Boxer Not Found</h2>
              <p className="text-gray-500 text-lg m-0">The boxer you're looking for doesn't exist.</p>
            </div>
          </main>
        </PageLayout>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
      <SkipLink targetId="profile-content" />
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
      >
        <main id="profile-content" className="max-w-[1000px] mx-auto px-5 py-6 pb-[12rem]" tabIndex="-1">
          {/* Boxer Header - Avatar, Name, Level, XP */}
          <article aria-label="Boxer header" className={sectionCardBase}>
            <BoxerHeader boxer={boxer} />
          </article>

          {/* Career Record - W/L/D/KO */}
          <section aria-label="Career record" className={sectionCardBase}>
            <CareerRecord boxer={boxer} />
          </section>

          {/* Next Fight Card (only shows when there's an upcoming fight) */}
          {upcomingFight && (
            <article aria-label="Next scheduled fight" className={sectionCardBase}>
              <NextFight upcomingFight={upcomingFight} boxerId={boxer.id} gameTime={worldTime?.current_game_time} onRefresh={loadProfileData} />
            </article>
          )}

          {/* Stats Overview - STR/DEF/AGI */}
          <section aria-label="Stats overview" className={sectionCardBase}>
            <StatsOverview boxer={boxer} />
          </section>

          {/* Current Status - Health/Energy/Fatigue */}
          <section aria-label="Current status" className={sectionCardBase}>
            <CurrentStatus boxer={boxer} worldTime={worldTime} />
          </section>

          {/* Recent Activity Timeline */}
          <section aria-label="Recent activity timeline" className={sectionCardBase}>
            <RecentActivity boxerId={boxer.id} gameTime={worldTime?.current_game_time} />
          </section>

          {/* Action Buttons */}
          <nav className="flex flex-wrap items-center justify-center gap-3 mt-8 pt-6 border-t border-white/10" aria-label="Boxer actions">
            <Link to={`/boxer/${boxer.id}/fights`} className="inline-flex items-center gap-2 px-6 py-3 rounded-md text-sm font-semibold bg-gradient-to-br from-primary to-primary-dark text-white border-none transition-all duration-150 hover:bg-gradient-to-br hover:from-blue-400 hover:to-primary hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(0,0,0,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2 cursor-pointer">
              📜 View Fight History
            </Link>
            <Link to={`/boxer/${boxer.id}/train`} className="inline-flex items-center gap-2 px-6 py-3 rounded-md text-sm font-semibold bg-gradient-to-br from-success to-emerald-600 text-white border-none transition-all duration-150 hover:bg-gradient-to-br hover:from-emerald-400 hover:to-success hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(0,0,0,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-success focus-visible:outline-offset-2 cursor-pointer">
              🏋️ Train Boxer
            </Link>
            <Link to={`/boxer/${boxer.id}/book-fight`} className="inline-flex items-center gap-2 px-6 py-3 rounded-md text-sm font-semibold bg-gradient-to-br from-warning to-orange-600 text-white border-none transition-all duration-150 hover:bg-gradient-to-br hover:from-victory-gold hover:to-warning hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(0,0,0,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-warning focus-visible:outline-offset-2 cursor-pointer">
              🥊 Book Fight
            </Link>
          </nav>
        </main>
      </PageLayout>
    </div>
  );
};

export default BoxerProfile;
