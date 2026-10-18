import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import FightHistory from './FightHistory';
import './FightHistoryPage.css';

/**
 * FightHistoryPage - Dedicated page wrapper for FightHistory component (MAT-121)
 *
 * Migrates FightHistory from modal to dedicated page at /boxer/:id/fights
 * Provides bookmarkable URL, full page real estate, and natural navigation flow.
 */
const FightHistoryPage = () => {
  const { id } = useParams(); // boxerId from URL
  const navigate = useNavigate();

  const handleBack = () => {
    navigate('/dashboard');
  };

  return (
    <div className="fight-history-page">
      {/* Navigation Header */}
      <div className="page-nav">
        <button className="back-btn" onClick={handleBack}>
          ← Back to Dashboard
        </button>
      </div>

      {/* FightHistory Component */}
      <FightHistory boxerId={id} currentGameTime={null} />
    </div>
  );
};

export default FightHistoryPage;
