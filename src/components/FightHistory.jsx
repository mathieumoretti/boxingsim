import React, { useState, useEffect } from 'react';
import './FightHistory.css';
import { fetchFightHistory, formatTimeAgo, getGroupLabel } from '../utils/fights';

/**
 * FightEntryCard - Single fight row component (MAT-103)
 * Displays compact fight info with expandable details
 */
const FightEntryCard = ({ fight, myBoxerId, currentGameTime }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  // Determine if this was a win, loss, or draw for my boxer
  const isWin = fight.winner_id === myBoxerId;
  const isLoss = fight.winner_id !== null && fight.winner_id !== myBoxerId;
  const isDraw = fight.winner_id === null && fight.status === 'completed';

  // Determine opponent boxer ID
  const opponentId = fight.boxer1_id === myBoxerId ? fight.boxer2_id : fight.boxer1_id;

  // Get result badge class and letter
  const getResultBadge = () => {
    if (isWin) return <span className="result-badge win">W</span>;
    if (isLoss) return <span className="result-badge loss">L</span>;
    if (isDraw) return <span className="result-badge draw">D</span>;
    return null;
  };

  // Check if fight ended by knockout
  const isKnockout = fight.data?.method === 'knockout' ||
                     (fight.data?.winner_health !== undefined && fight.data.winner_health > 50);

  // Calculate time ago using game world time instead of real time
  const getTimeAgoDisplay = () => {
    const fightDate = fight.end_time || fight.created_at;
    if (!fightDate) return '?';

    // If we have game time, use it; otherwise fall back to real time
    const now = currentGameTime ? new Date(currentGameTime) : new Date();
    const fightDateTime = new Date(fightDate);

    const diffMs = now - fightDateTime;
    const diffSeconds = Math.floor(diffMs / 1000);
    const diffMinutes = Math.floor(diffSeconds / 60);
    const diffHours = Math.floor(diffMinutes / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffSeconds < 60) return 'Just now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays <= 30) return `${diffDays}d ago`;
    return `${Math.floor(diffDays / 30)}mo ago`;
  };

  return (
    <div className={`fight-entry-card ${isExpanded ? 'expanded' : ''}`}>
      <div className="fight-entry-header" onClick={() => setIsExpanded(!isExpanded)}>
        {/* Result Badge */}
        {getResultBadge()}

        {/* Opponent Info */}
        <div className="fight-info">
          <div className="opponent-name">
            vs {fight.opponent_name || `Boxer #${opponentId}`}
          </div>
          <div className="fight-summary">
            Round {fight.round} • {isKnockout && '🔥 KO'}
          </div>
        </div>

        {/* Time Ago */}
        <div className="fight-time">
          {getTimeAgoDisplay()}
        </div>

        {/* Expand Icon */}
        <div className="expand-icon">
          {isExpanded ? '▲' : '▼'}
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && (
        <div className="fight-entry-details">
          <div className="detail-row">
            <span className="detail-label">Final Result:</span>
            <span className="detail-value">
              {isWin ? '🎉 Victory!' : isLoss ? '💔 Defeat' : '🤝 Draw'}
            </span>
          </div>

          {fight.data && (
            <>
              {(fight.data.my_health !== undefined || fight.data.opponent_health !== undefined) && (
                <div className="detail-row">
                  <span className="detail-label">Health at End:</span>
                  <span className="detail-value">
                    You: {Math.round(fight.data.my_health || 100)}% •
                    Opponent: {Math.round(fight.data.opponent_health || 100)}%
                  </span>
                </div>
              )}

              {fight.data.method && (
                <div className="detail-row">
                  <span className="detail-label">Method:</span>
                  <span className="detail-value capitalize">{fight.data.method}</span>
                </div>
              )}
            </>
          )}

          <div className="detail-row">
            <span className="detail-label">Rounds Fought:</span>
            <span className="detail-value">{fight.round} of {fight.round || 12}</span>
          </div>

          {fight.data?.xp_gained && (
            <div className="detail-row xp-gained">
              <span className="detail-label">XP Gained:</span>
              <span className="detail-value">+{Math.round(fight.data.xp_gained)}</span>
            </div>
          )}

          {fight.data?.reward && (
            <div className="detail-row reward">
              <span className="detail-label">Reward:</span>
              <span className="detail-value">${Math.round(fight.data.reward)}</span>
            </div>
          )}

          <div className="detail-row">
            <span className="detail-label">Date & Time:</span>
            <span className="detail-value">
              {new Date(fight.end_time || fight.created_at).toLocaleString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: 'numeric',
                minute: '2-digit',
                hour12: true
              })}
            </span>
          </div>
        </div>
      )}
    </div>
  );
};

/**
 * TimelineSection - Date-grouped section with header (MAT-103)
 */
const TimelineSection = ({ label, fights, myBoxerId, currentGameTime }) => {
  if (!fights || fights.length === 0) return null;

  return (
    <div className="timeline-section">
      <div className="timeline-header">
        <h3 className="timeline-label">{label}</h3>
        <span className="fight-count">{fights.length} fight{fights.length !== 1 ? 's' : ''}</span>
      </div>
      <div className="fights-list">
        {fights.map(fight => (
          <FightEntryCard key={fight.id} fight={fight} myBoxerId={myBoxerId} currentGameTime={currentGameTime} />
        ))}
      </div>
    </div>
  );
};

/**
 * FightHistory - Main component displaying boxer's fight history (MAT-103)
 */
const FightHistory = ({ boxerId, currentGameTime }) => {
  const [fights, setFights] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [groupedFights, setGroupedFights] = useState({});

  // Fetch fight history
  useEffect(() => {
    const loadFightHistory = async () => {
      if (!boxerId) return;

      setIsLoading(true);
      setError(null);

      try {
        const data = await fetchFightHistory(boxerId);

        // Ensure we always have an array
        const fightsArray = Array.isArray(data) ? data : [];
        setFights(fightsArray);

        // Group fights by date range (using game time)
        if (fightsArray.length > 0 && currentGameTime) {
          const groups = groupFightsByDate(fightsArray, currentGameTime);
          setGroupedFights(groups);
        }
      } catch (err) {
        setError('Failed to load fight history');
        console.error('Error loading fight history:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadFightHistory();
  }, [boxerId]);

  // Refresh function
  const handleRefresh = async () => {
    await new Promise(resolve => setTimeout(resolve, 300)); // Small delay for UX
    setIsLoading(true);

    try {
      const data = await fetchFightHistory(boxerId);
      const fightsArray = Array.isArray(data) ? data : [];
      setFights(fightsArray);

      if (fightsArray.length > 0 && currentGameTime) {
        const groups = groupFightsByDate(fightsArray, currentGameTime);
        setGroupedFights(groups);
      }
      setError(null);
    } catch (err) {
      setError('Failed to refresh fight history');
      console.error('Error refreshing fight history:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Group fights by date range for timeline view (using game time)
  const groupFightsByDate = (fightsList, gameTime) => {
    const now = new Date(gameTime);
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    const lastWeek = new Date(today);
    lastWeek.setDate(lastWeek.getDate() - 7);

    const groups = {
      today: [],
      yesterday: [],
      lastWeek: [],
      older: []
    };

    fightsList.forEach(fight => {
      const fightDate = new Date(fight.end_time || fight.created_at);
      const fightDateOnly = new Date(fightDate.getFullYear(), fightDate.getMonth(), fightDate.getDate());

      if (fightDateOnly.getTime() === today.getTime()) {
        groups.today.push(fight);
      } else if (fightDateOnly.getTime() === yesterday.getTime()) {
        groups.yesterday.push(fight);
      } else if (fightDate > lastWeek) {
        groups.lastWeek.push(fight);
      } else {
        groups.older.push(fight);
      }
    });

    return groups;
  };

  // Render content based on state
  const renderContent = () => {
    if (isLoading) {
      return <div className="loading">Loading fight history...</div>;
    }

    if (error) {
      return (
        <div className="error-state">
          <p>⚠️ {error}</p>
          <button onClick={handleRefresh} className="retry-btn">Try Again</button>
        </div>
      );
    }

    if (fights.length === 0) {
      return (
        <div className="empty-state">
          <span className="empty-icon">📜</span>
          <h3>No Fights Yet</h3>
          <p>This boxer hasn't fought any matches yet.</p>
          <p className="empty-hint">Book a fight to see the action!</p>
        </div>
      );
    }

    // Render timeline sections with fights
    return (
      <>
        {Object.entries(groupedFights).map(([key, groupFights]) =>
          groupFights.length > 0 && (
            <TimelineSection
              key={key}
              label={getGroupLabel(key)}
              fights={groupFights}
              myBoxerId={boxerId}
              currentGameTime={currentGameTime}
            />
          )
        )}
      </>
    );
  };

  return (
    <div className="fight-history-container">
      <div className="fight-history-header">
        <h2>Fight History</h2>
        <button
          onClick={handleRefresh}
          className="refresh-btn"
          disabled={isLoading}
          title="Refresh fight history"
        >
          {isLoading ? '⟳ Loading...' : '🔄 Refresh'}
        </button>
      </div>

      <div className="fight-history-content">
        {renderContent()}
      </div>
    </div>
  );
};

export default FightHistory;
