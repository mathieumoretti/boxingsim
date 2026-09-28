import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import './FightDetails.css';
import { fetchFightDetails, fetchBoxer, formatTimeAgo } from '../utils/fights';
import RoundTimeline from './RoundTimeline';

/**
 * VictoryBanner - Displays fight result prominently (MAT-104)
 */
const VictoryBanner = ({ winner, loser, method, round }) => {
  if (!winner) return null;

  const isDraw = !winner && loser === null;

  if (isDraw) {
    return (
      <div className="victory-banner draw">
        <h1>🤝 IT'S A DRAW! 🤝</h1>
        <p>Both boxers fought valiantly to a standstill</p>
      </div>
    );
  }

  const isKO = method === 'knockout';
  const victoryEmoji = isKO ? '🔥' : '🏆';

  return (
    <div className={`victory-banner ${isKO ? 'ko' : ''}`}>
      <h1>
        {victoryEmoji} {winner.name.toUpperCase()} WINS {victoryEmoji}
      </h1>
      <p>
        by {method === 'knockout' ? 'Knockout' : method === 'decision' ? 'Decision' : method?.charAt(0).toUpperCase() + method.slice(1) || 'Victory'} - Round {round}
      </p>
    </div>
  );
};

/**
 * BoxerComparisonCard - Side-by-side boxer stats (MAT-104)
 */
const BoxerComparisonCard = ({ boxer, role, fightData }) => {
  if (!boxer) return null;

  const isWinner = role === 'winner';
  const healthAtEnd = role === 'boxer1' && fightData?.my_health !== undefined
    ? fightData.my_health
    : role === 'boxer2' && fightData?.opponent_health !== undefined
      ? fightData.opponent_health
      : null;

  // Generate avatar based on name
  const getAvatarEmoji = (name) => {
    const avatars = ['🥊', '💪', '🏆', '⚡', '🔥', '🛡️', '🎯', '😤'];
    if (!name) return avatars[0];
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = ((hash << 5) - hash) + name.charCodeAt(i);
      hash |= 0;
    }
    return avatars[Math.abs(hash) % avatars.length];
  };

  const getHealthColorClass = (health) => {
    if (!health) return 'bg-health-medium';
    const pct = health;
    if (pct > 50) return 'bg-health-high';
    if (pct > 25) return 'bg-health-medium';
    return 'bg-health-low';
  };

  const avatar = getAvatarEmoji(boxer.name);
  const healthClass = getHealthColorClass(healthAtEnd);

  return (
    <div className={`boxer-comparison-card ${isWinner ? 'winner' : ''} ${role}`}>
      <div className="card-avatar">{avatar}</div>

      <div className="card-header">
        <h3>{boxer.name}</h3>
        {boxer.nickname && <p className="nickname">"{boxer.nickname}"</p>}
        <span className={`level-badge ${role === 'winner' ? 'winner-bg' : ''}`}>LVL {boxer.level || 1}</span>
      </div>

      {/* Health at end of fight */}
      {healthAtEnd !== undefined && (
        <div className="fight-stat-row">
          <span className="stat-label">Health at End:</span>
          <div className="stat-bar-wrapper">
            <div
              className={`stat-bar ${healthClass}`}
              style={{ width: `${Math.min(100, Math.max(0, healthAtEnd))}%` }}
            ></div>
          </div>
          <span className="stat-value">{Math.round(healthAtEnd)}%</span>
        </div>
      )}

      {/* Combat Stats */}
      <div className="combat-stats-grid">
        <div className="stat-item">
          <span className="stat-icon">💪</span>
          <div className="stat-info">
            <span className="stat-name">STR</span>
            <span className="stat-value">{Math.round(boxer.strength)}</span>
          </div>
        </div>

        <div className="stat-item">
          <span className="stat-icon">🛡️</span>
          <div className="stat-info">
            <span className="stat-name">DEF</span>
            <span className="stat-value">{Math.round(boxer.defense)}</span>
          </div>
        </div>

        <div className="stat-item">
          <span className="stat-icon">⚡</span>
          <div className="stat-info">
            <span className="stat-name">AGI</span>
            <span className="stat-value">{Math.round(boxer.agility)}</span>
          </div>
        </div>
      </div>

      {/* Additional Info */}
      <div className="card-footer">
        <span className="info-item">
          <span className="label">Record:</span>
          <span className="value">
            {boxer.wins || 0}W - {boxer.losses || 0}L - {boxer.draws || 0}D
          </span>
        </span>
      </div>
    </div>
  );
};

/**
 * PlayByPlayLog - Text-based fight narrative (MAT-104)
 */
const PlayByPlayLog = ({ fightData, rounds }) => {
  const [selectedRound, setSelectedRound] = useState(1);

  // Generate play-by-play log from fight data
  const generatePlayByPlay = () => {
    if (!fightData || !fightData.log) return null;

    const logLines = fightData.log.split('\n').filter(line => line.trim());

    return (
      <div className="play-by-play-log">
        <h3>🥊 Round {selectedRound} Play-by-Play 🥊</h3>

        {logLines.length > 0 ? (
          <div className="log-content">
            {logLines.map((line, index) => {
              let logClass = 'log-line';
              if (line.includes('WINS') || line.includes('Victory')) logClass += ' victory';
              else if (line.includes('knockout') || line.includes('KO')) logClass += ' knockout';
              else if (line.includes('damage') || line.includes('-')) logClass += ' damage';
              else if (line.includes('MISS') || line.includes('dodge')) logClass += ' miss';
              else if (line.includes('Exchange')) logClass += ' exchange';

              return (
                <div key={index} className={logClass}>
                  {line}
                </div>
              );
            })}
          </div>
        ) : (
          <p className="no-log">No detailed log available for this round.</p>
        )}

        {/* Navigation */}
        <div className="log-nav">
          <button
            className="nav-btn prev"
            onClick={() => setSelectedRound(Math.max(1, selectedRound - 1))}
            disabled={selectedRound === 1}
          >
            ← Previous Round
          </button>
          <span className="round-indicator">Round {selectedRound} of {rounds}</span>
          <button
            className="nav-btn next"
            onClick={() => setSelectedRound(Math.min(rounds, selectedRound + 1))}
            disabled={selectedRound === rounds}
          >
            Next Round →
          </button>
        </div>
      </div>
    );
  };

  return generatePlayByPlay();
};

/**
 * FightDetails - Main component (MAT-104)
 */
const FightDetails = () => {
  const { id } = useParams();
  const navigate = useNavigate();

  const [fight, setFight] = useState(null);
  const [boxer1, setBoxer1] = useState(null);
  const [boxer2, setBoxer2] = useState(null);
  const [winner, setWinner] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Fetch fight and boxer data
  useEffect(() => {
    const loadFightDetails = async () => {
      if (!id) return;

      setIsLoading(true);
      setError(null);

      try {
        // Fetch fight details
        const fightData = await fetchFightDetails(id);
        setFight(fightData);

        // Fetch boxer details in parallel
        const [boxer1Data, boxer2Data] = await Promise.all([
          fetchBoxer(fightData.boxer1_id),
          fetchBoxer(fightData.boxer2_id)
        ]);

        setBoxer1(boxer1Data);
        setBoxer2(boxer2Data);

        // Determine winner
        if (fightData.winner_id) {
          const winnerBoxer = fightData.winner_id === boxer1Data.id ? boxer1Data : boxer2Data;
          setWinner(winnerBoxer);
        } else if (fightData.status === 'completed') {
          // Draw - no winner
          setWinner(null);
        }
      } catch (err) {
        setError(`Failed to load fight details: ${err.message}`);
        console.error('Error loading fight details:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadFightDetails();
  }, [id]);

  // Get opponent based on winner
  const getOpponent = () => {
    if (!winner || !boxer1 || !boxer2) return null;
    return winner.id === boxer1.id ? boxer2 : boxer1;
  };

  const opponent = getOpponent();

  // Calculate fight data for display
  const fightData = fight?.data || {};
  const roundsFought = fight?.round || 12;
  const method = fightData?.method || 'decision';

  // Render loading state
  if (isLoading) {
    return (
      <div className="fight-details-page">
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Loading fight details...</p>
        </div>
      </div>
    );
  }

  // Render error state
  if (error) {
    return (
      <div className="fight-details-page">
        <div className="error-state">
          <span className="error-icon">⚠️</span>
          <h2>Error Loading Fight</h2>
          <p>{error}</p>
          <button
            className="back-btn"
            onClick={() => navigate('/dashboard')}
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  // Render fight details
  return (
    <div className="fight-details-page">
      {/* Navigation Header */}
      <div className="details-nav">
        <button
          className="back-btn"
          onClick={() => navigate('/dashboard')}
        >
          ← Back to Dashboard
        </button>
      </div>

      {/* Victory Banner */}
      <VictoryBanner
        winner={winner}
        loser={fightData.winner_id ? opponent : null}
        method={method}
        round={roundsFought}
      />

      {/* Fight Info Header */}
      <div className="fight-info-header">
        <h2>
          {boxer1?.name || 'Boxer 1'} vs {boxer2?.name || 'Boxer 2'}
        </h2>
        <p className="fight-date">
          📅 {new Date(fight?.end_time || fight?.created_at).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'long',
            day: 'numeric'
          })} at{' '}
          {new Date(fight?.end_time || fight?.created_at).toLocaleTimeString('en-US', {
            hour: 'numeric',
            minute: '2-digit',
            hour12: true
          })}
        </p>
        <p className="time-ago">
          ({formatTimeAgo(fight?.end_time || fight?.created_at)})
        </p>
      </div>

      {/* Boxer Comparison */}
      <div className="boxer-comparison-section">
        <h3>FIGHTER COMPARISON</h3>
        <div className="comparison-layout">
          <BoxerComparisonCard
            boxer={boxer1}
            role="boxer1"
            fightData={fightData}
          />

          <div className="vs-badge">VS</div>

          <BoxerComparisonCard
            boxer={boxer2}
            role="boxer2"
            fightData={fightData}
          />
        </div>
      </div>

      {/* Round Timeline */}
      {roundsFought && (
        <div className="timeline-section">
          <h3>ROUND TIMELINE</h3>
          <RoundTimeline
            totalRounds={roundsFought}
            method={method}
          />
        </div>
      )}

      {/* Play-by-Play Log */}
      {fightData?.log && (
        <div className="playbyplay-section">
          <h3>FIGHT NARRATIVE</h3>
          <PlayByPlayLog
            fightData={fightData}
            rounds={roundsFought}
          />
        </div>
      )}

      {/* Fight Summary Stats */}
      <div className="fight-summary-section">
        <h3>FIGHT SUMMARY</h3>
        <div className="summary-grid">
          <div className="summary-card">
            <span className="label">Status</span>
            <span className={`value ${fight?.status === 'completed' ? 'completed' : ''}`}>
              {fight?.status || 'Unknown'}
            </span>
          </div>

          <div className="summary-card">
            <span className="label">Rounds Fought</span>
            <span className="value">{roundsFought}</span>
          </div>

          {method && (
            <div className="summary-card">
              <span className="label">Method of Victory</span>
              <span className="value capitalize">{method}</span>
            </div>
          )}

          {fightData?.xp_gained && (
            <div className="summary-card xp">
              <span className="label">XP Gained</span>
              <span className="value">+{Math.round(fightData.xp_gained)}</span>
            </div>
          )}

          {fightData?.reward && (
            <div className="summary-card reward">
              <span className="label">Purse Earned</span>
              <span className="value">${Math.round(fightData.reward).toLocaleString()}</span>
            </div>
          )}
        </div>
      </div>

      {/* Boxer Links */}
      {boxer1 && boxer2 && (
        <div className="boxer-links-section">
          <h3>FIGHTERS</h3>
          <div className="link-buttons">
            <button
              className="boxer-link-btn"
              onClick={() => navigate(`/boxer/${boxer1.id}`)}
            >
              View {boxer1.name}'s Profile
            </button>
            <button
              className="boxer-link-btn"
              onClick={() => navigate(`/boxer/${boxer2.id}`)}
            >
              View {boxer2.name}'s Profile
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default FightDetails;
