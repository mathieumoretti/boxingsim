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
        {/* Energy Status Section */}
          <section className="energy-status-section">
            <h2>Current Energy</h2>
            <div className="energy-status-card">
              <div className="energy-info">
                <span className="energy-label">Energy Available:</span>
                <span className={`energy-value ${boxer.energy <= 10 ? 'low' : boxer.energy <= 30 ? 'medium' : ''}`}>
                  {Math.round(boxer.energy || 0)} / 100
                </span>
              </div>
              <div className="energy-bar-wrapper">
                <div
                  className="energy-progress"
                  style={{ width: `${Math.min((boxer.energy || 0), 100)}%` }}
                ></div>
              </div>
              {boxer.has_active_rest && (
                <div className="rest-status">
                  <span className="rest-icon">{boxer.forced_rest_until ? '🚑' : '🏥'}</span>
                  <span>Currently resting — cannot schedule new training</span>
                </div>
              )}
            </div>
          </section>

          {/* Training Scheduler Section */}
          <section className="training-scheduler-section">
            <h2>Schedule New Training</h2>
            <TrainingScheduler
              boxerId={boxer.id}
              boxer={boxer}
              onClose={() => navigate('/dashboard')}
              onTrainingScheduled={handleTrainingScheduled}
            />
          </section>

          {/* Stat Progression Chart Section */}
          <section className="stat-progression-section">
            <h2>Stat Progression</h2>
            <p className="section-description">Track how {boxer.nickname || boxer.name}'s stats have evolved over time through training sessions.</p>
            <StatProgressionChart boxerId={boxer.id} />
          </section>

          {/* Training History Section */}
          <section className="training-history-section">
            <h2>Training History</h2>
            <p className="section-description">Past training sessions completed by {boxer.name}.</p>
            <TrainingHistory boxerId={boxer.id} />
          </section>
        </PageLayout>
      </div>
  );
};

export default TrainingPage;
