/**
 * TrainingPage - Training management page content
 *
 * Compact single-column layout inheriting navy gradient from PageLayout.
 * Minimal separation for clean information flow.
 */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import TopBar from './TopBar';
import PageLayout from './PageLayout';
import TrainingScheduler from './TrainingScheduler';
import TrainingHistory from './TrainingHistory';
import StatProgressionChart from './StatProgressionChart';
import { ErrorDisplay } from './ErrorDisplay';
import { fetchBoxer } from '../api/fights';
import { getUserErrorMessage } from '../utils/errorMessages';

const TrainingPage = ({ user, onLogout }) => {
  const { id: boxerIdParam } = useParams();
  const navigate = useNavigate();
  const [boxer, setBoxer] = useState(null);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadBoxerData();
  }, [boxerIdParam]);

  const loadBoxerData = async () => {
    if (!boxerIdParam) {
      setError('No boxer selected');
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const data = await fetchBoxer(boxerIdParam);
      if (data && data.id) {
        setBoxer(data);
      } else {
        setError('Boxer not found');
      }
    } catch (err) {
      setError(getUserErrorMessage(err, 'general'));
    } finally {
      setIsLoading(false);
    }
  };

  // Refresh boxer data after training scheduled
  const handleTrainingScheduled = () => {
    loadBoxerData();
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-navy-medium to-navy-light">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout>
          <div className="max-w-[900px] mx-auto px-6 py-6 pb-[10rem]">
            <div className="text-center py-8 px-6 text-white/60 text-sm">Loading boxer data...</div>
          </div>
        </PageLayout>
      </div>
    );
  }

  if (error && !boxer) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-navy-medium to-navy-light">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <ErrorDisplay error={error} variant="alert" title="Failed to Load Training Page" icon={true} />
        </PageLayout>
      </div>
    );
  }

  if (!boxer) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-navy-medium to-navy-light">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <div className="max-w-[900px] mx-auto px-6 py-6 pb-[10rem]">
            <div className="text-center py-8 px-6 text-white/60 text-sm">Boxer not found</div>
          </div>
        </PageLayout>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-navy-medium to-navy-light">
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
      >
        <div className="max-w-[900px] mx-auto px-6 py-6 pb-8 md:pb-10 lg:pb-8">
          {/* Training Scheduler Section */}
          <section className="mb-4 md:mb-4 lg:mb-4">
            <h3 className="text-sm font-semibold text-victory-gold uppercase tracking-wider m-0 mb-2">Schedule Training</h3>
            <TrainingScheduler
              boxerId={boxer.id}
              boxer={boxer}
              onClose={() => navigate('/dashboard')}
              onTrainingScheduled={handleTrainingScheduled}
            />
          </section>

          {/* Stat Progression Chart Section */}
          <section className="mb-4 md:mb-4 lg:mb-4">
            <h3 className="text-sm font-semibold text-victory-gold uppercase tracking-wider m-0 mb-2">Stat Progression</h3>
            <StatProgressionChart boxerId={boxer.id} compact />
          </section>

          {/* Training History Section */}
          <section className="mb-4 md:mb-4 lg:mb-4">
            <h3 className="text-sm font-semibold text-victory-gold uppercase tracking-wider m-0 mb-2">Training History</h3>
            <TrainingHistory boxerId={boxer.id} compact />
          </section>
        </div>
      </PageLayout>
    </div>
  );
};

export default TrainingPage;
