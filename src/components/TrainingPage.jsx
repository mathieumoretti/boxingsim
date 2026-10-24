import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './TrainingPage.css';
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
      <div className="training-page">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout>
          <div className="training-page-content">
            <div className="loading-state">Loading boxer data...</div>
          </div>
        </PageLayout>
      </div>
    );
  }

  if (error && !boxer) {
    return (
      <div className="training-page">
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
      <div className="training-page">
        <TopBar currentUser={user} onLogout={onLogout} />
        <PageLayout
          onBack={() => navigate('/dashboard')}
        >
          <div className="training-page-content">
            <div className="error-state">Boxer not found</div>
          </div>
        </PageLayout>
      </div>
    );
  }

  return (
    <div className="training-page">
      <TopBar currentUser={user} onLogout={onLogout} />
      <PageLayout
        onBack={() => navigate('/dashboard')}
      >
        {/* Training Scheduler Section */}
        <section className="training-scheduler-section">
          <h3>Schedule Training</h3>
          <TrainingScheduler
            boxerId={boxer.id}
            boxer={boxer}
            onClose={() => navigate('/dashboard')}
            onTrainingScheduled={handleTrainingScheduled}
          />
        </section>

        {/* Stat Progression Chart Section */}
        <section className="stat-progression-section">
          <h3>Stat Progression</h3>
          <StatProgressionChart boxerId={boxer.id} compact />
        </section>

        {/* Training History Section */}
        <section className="training-history-section">
          <h3>Training History</h3>
          <TrainingHistory boxerId={boxer.id} compact />
        </section>
      </PageLayout>
    </div>
  );
};

export default TrainingPage;
