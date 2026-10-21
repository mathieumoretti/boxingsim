import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import TopBar from './TopBar';
import FightHistory from './FightHistory';
import PageLayout from './PageLayout';
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

  return (
    <>
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
        backButtonLabel="← Back to Dashboard"
      >
        {/* Wrapper div for CSS selectors */}
        <div className="fight-history-page">
          <FightHistory boxerId={id} currentGameTime={null} />
        </div>
      </PageLayout>
    </>
  );
};

export default FightHistoryPage;
