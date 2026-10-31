import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import TopBar from './TopBar';
import FightHistory from './FightHistory';
import PageLayout from './PageLayout';
import { fetchWorldTime } from '../api/world';
import './FightHistoryPage.css';

/**
 * FightHistoryPage - Dedicated page wrapper for FightHistory component (MAT-121)
 *
 * Migrates FightHistory from modal to dedicated page at /boxer/:id/fights
 * Provides bookmarkable URL, full page real estate, and natural navigation flow.
 *
 * Uses PageLayout component for standardized page structure (MAT-152).
 */
const FightHistoryPage = ({ user, onLogout }) => {
  const { id } = useParams(); // boxerId from URL
  const navigate = useNavigate();
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
    <>
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
        backButtonLabel="← Back to Dashboard"
      >
        {/* Wrapper div for CSS selectors */}
        <div className="fight-history-page">
          <FightHistory boxerId={id} currentGameTime={currentGameTime} />
        </div>
      </PageLayout>
    </>
  );
};

export default FightHistoryPage;
