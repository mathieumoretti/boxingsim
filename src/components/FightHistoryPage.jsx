import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import TopBar from './TopBar';
import FightHistory from './FightHistory';
import PageLayout from './PageLayout';
import { fetchWorldTime } from '../api/world';

/**
 * FightHistoryPage - Dedicated page wrapper for FightHistory component (MAT-121)
 * Redesigned with dark arena theme matching boxer profile pages.
 *
 * Migrates FightHistory from modal to dedicated page at /boxer/:id/fights
 * Provides bookmarkable URL, full page real estate, and natural navigation flow.
 *
 * Uses PageLayout component for standardized page structure (MAT-152).
 */
const FightHistoryPage = ({ user, onLogout }) => {
  const navigate = useNavigate();
  const { id } = useParams(); // boxerId from /boxer/:id/fights
  const [currentGameTime, setCurrentGameTime] = useState(null);

  useEffect(() => {
    const loadWorldTime = async () => {
      try {
        const timeData = await fetchWorldTime();
        setCurrentGameTime(timeData?.current_game_time || null);
      } catch (err) {
        // World time is optional - log but don't fail
        console.debug('Failed to load world time:', err.message);
      }
    };

    loadWorldTime();

    // Refresh game time periodically
    const refreshInterval = setInterval(loadWorldTime, 10000);
    return () => clearInterval(refreshInterval);
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate(id ? `/boxer/${id}` : '/boxers')}
        backButtonLabel="← Back to Profile"
      >
        <FightHistory boxerId={id} currentGameTime={currentGameTime} />
      </PageLayout>
    </div>
  );
};

export default FightHistoryPage;
