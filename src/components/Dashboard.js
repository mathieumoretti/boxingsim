import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import './Dashboard.css';
import BoxerCard from './BoxerCard.jsx';
import TrainingScheduler from './TrainingScheduler.jsx';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';

const Dashboard = ({ user, onLogout }) => {
  const [boxers, setBoxers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentUser, setCurrentUser] = useState(null);
  const [selectedBoxerForTraining, setSelectedBoxerForTraining] = useState(null);

  useEffect(() => {
    // Get user from props or localStorage
    const currentUser = user || getUser();
    setCurrentUser(currentUser);

    if (currentUser) {
      loadUserBoxers(currentUser.id);
    } else {
      setIsLoading(false);
    }
  }, [user]);

  const loadUserBoxers = async (userId) => {
    setIsLoading(true);
    setError('');

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/users/${userId}/boxers`, {
        method: 'GET',
      });

      const data = await response.json();

      if (response.ok) {
        // Ensure we always have an array, even if API returns null for empty results
        setBoxers(Array.isArray(data) ? data : []);
      } else {
        setError(data.error || 'Failed to load boxers');
      }
    } catch (error) {
      // Don't show error if redirected to login (401 handled by authenticatedFetch)
      if (!error.message.includes('Unauthorized')) {
        setError('Network error: ' + error.message);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogout = () => {
    onLogout();
  };

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <h1>Boxing Simulator</h1>
        <div className="user-info">
          <span>Welcome, {currentUser?.username || 'User'}!</span>
          <button onClick={handleLogout} className="logout-btn">Logout</button>
        </div>
      </header>

      <main className="dashboard-content">
        <section className="boxer-section">
          <div className="section-header">
            <h2>Your Boxers</h2>
            <Link to="/create-boxer">
              <button className="create-boxer-btn">Create New Boxer</button>
            </Link>
          </div>

          {isLoading ? (
            <div className="loading">Loading boxers...</div>
          ) : error ? (
            <div className="error-message">{error}</div>
          ) : boxers.length === 0 ? (
            <p>No boxers created yet.</p>
          ) : (
            <div className="boxers-grid">
              {boxers.map((boxer) => (
                <BoxerCard
                  key={boxer.id}
                  boxer={boxer}
                  onOpenTraining={() => setSelectedBoxerForTraining(boxer)}
                />
              ))}
            </div>
          )}
        </section>

        {/* Centralized Training Modal */}
        {selectedBoxerForTraining && (
          <div className="training-modal-overlay" onClick={() => setSelectedBoxerForTraining(null)}>
            <div className="training-modal" onClick={(e) => e.stopPropagation()}>
              <button
                className="modal-close"
                onClick={() => setSelectedBoxerForTraining(null)}
                aria-label="Close modal"
              >×</button>
              <TrainingScheduler
                boxerId={selectedBoxerForTraining.id}
                boxer={selectedBoxerForTraining}
                onClose={() => setSelectedBoxerForTraining(null)}
              />
            </div>
          </div>
        )}

        <section className="fight-section">
          <h2>Fight Arena</h2>
          <div className="fight-arena">
            <div className="boxer-display">
              <h4>Boxer 1</h4>
              <div className="boxer-stats">
                <p>Health: <span id="boxer1-health">0</span></p>
                <p>Energy: <span id="boxer1-energy">0</span></p>
                <p>Strength: <span id="boxer1-strength">0</span></p>
              </div>
            </div>
            <div className="fight-controls">
              <button className="fight-btn">Start Fight</button>
              <button className="reset-btn">Reset</button>
            </div>
            <div className="boxer-display">
              <h4>Boxer 2</h4>
              <div className="boxer-stats">
                <p>Health: <span id="boxer2-health">0</span></p>
                <p>Energy: <span id="boxer2-energy">0</span></p>
                <p>Strength: <span id="boxer2-strength">0</span></p>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Dashboard;
