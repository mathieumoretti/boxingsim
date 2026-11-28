import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { fetchFightHistory, formatTimeAgo, getGroupLabel } from '../api/fights';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage, isNotFoundError } from '../utils/errorMessages';

/**
 * FightEntryCard - Single fight row component (MAT-103)
 * Displays compact fight info with expandable details
 * Memoized to prevent unnecessary re-renders when currentGameTime updates (MAT-132)
 */
const FightEntryCard = React.memo(({ fight, myBoxerId, currentGameTime }) => {
  const [isExpanded, setIsExpanded] = useState(false);

  // Convert myBoxerId to integer for proper comparison (URL params are strings)
  const myBoxerIdInt = typeof myBoxerId === 'string' ? parseInt(myBoxerId, 10) : myBoxerId;

  // Determine if this was a win, loss, or draw for my boxer
  const isWin = fight.winner_id === myBoxerIdInt;
  const isLoss = fight.winner_id !== null && fight.winner_id !== myBoxerIdInt;
  const isDraw = fight.winner_id === null && fight.status === 'completed';

  // Determine opponent boxer ID
  const opponentId = fight.boxer1_id === myBoxerIdInt ? fight.boxer2_id : fight.boxer1_id;

  // Get result badge class and letter
  const getResultBadge = () => {
    if (isWin) return <span className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-green-500 to-emerald-600 text-white text-xs font-bold">W</span>;
    if (isLoss) return <span className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-red-500 to-red-600 text-white text-xs font-bold">L</span>;
    if (isDraw) return <span className="inline-flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-br from-gray-400 to-gray-600 text-white text-xs font-bold">D</span>;
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
    <div className={`bg-gradient-to-br from-white/10 to-white/5 rounded-xl border border-white/10 shadow-[0_2px_8px_rgba(0,0,0,0.2)] overflow-hidden transition-all duration-150 hover:bg-white/15 ${isExpanded ? 'ring-2 ring-primary/30' : ''}`}>
      {/* Header row - clickable to expand */}
      <div
        className="flex items-center gap-4 p-5 cursor-pointer select-none"
        onClick={(e) => {
          // Don't prevent default if clicking the View Details link
          if (e.target.closest('.view-details-link')) {
            return;
          }
          setIsExpanded(!isExpanded);
        }}
        role="button"
        tabIndex={0}
        aria-expanded={isExpanded}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            setIsExpanded(!isExpanded);
          }
        }}
      >
        {/* Result Badge */}
        {getResultBadge()}

        {/* Opponent Info */}
        <div className="flex-1 min-w-0">
          <div className="text-sm font-bold text-white truncate">
            vs {fight.opponent_name || `Boxer #${opponentId}`}
          </div>
          <div className="flex items-center gap-2 mt-1 text-xs text-white/60">
            <span>Round {fight.round}</span>
            {isKnockout && <span className="text-orange-300 font-semibold">🔥 KO</span>}
            <span>•</span>
            <Link
              to={`/fight/${fight.id}`}
              className="view-details-link inline-flex items-center gap-1 px-3 py-1.5 bg-white/10 border border-white/20 rounded-md text-white text-xs font-semibold transition-all duration-150 hover:bg-white/15 hover:border-white/30 hover:-translate-y-0.5 focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            >
              View Details →
            </Link>
          </div>
        </div>

        {/* Time Ago */}
        <div className="text-xs text-white/50 whitespace-nowrap flex-shrink-0">
          {getTimeAgoDisplay()}
        </div>

        {/* Expand Icon */}
        <div
          className={`text-base text-white/40 transition-all duration-150 flex-shrink-0 ${isExpanded || ''}`}
        >
          {isExpanded ? '▲' : '▼'}
        </div>
      </div>

      {/* Expanded Details */}
      {isExpanded && (
        <div className="px-5 pb-5 pt-0 space-y-3 bg-white/5">
          <div className="flex items-center justify-between py-2 border-b border-white/10">
            <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">Final Result:</span>
            <span className="text-sm font-bold text-white">
              {isWin ? '🎉 Victory!' : isLoss ? '💔 Defeat' : '🤝 Draw'}
            </span>
          </div>

          {fight.data && (
            <>
              {(fight.data.my_health !== undefined || fight.data.opponent_health !== undefined) && (
                <div className="flex items-center justify-between py-2 border-b border-white/10">
                  <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">Health at End:</span>
                  <span className="text-sm font-bold text-white">
                    You: {Math.round(fight.data.my_health || 100)}% •
                    Opponent: {Math.round(fight.data.opponent_health || 100)}%
                  </span>
                </div>
              )}

              {fight.data.method && (
                <div className="flex items-center justify-between py-2 border-b border-white/10">
                  <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">Method:</span>
                  <span className="text-sm font-bold text-white capitalize">{fight.data.method}</span>
                </div>
              )}
            </>
          )}

          <div className="flex items-center justify-between py-2 border-b border-white/10">
            <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">Rounds Fought:</span>
            <span className="text-sm font-bold text-white">{fight.round} of {fight.round || 12}</span>
          </div>

          {fight.data?.xp_gained && (
            <div className="flex items-center justify-between py-2 border-b border-white/10">
              <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">XP Gained:</span>
              <span className="text-sm font-bold text-warning-bright">+{Math.round(fight.data.xp_gained)}</span>
            </div>
          )}

          {fight.data?.reward && (
            <div className="flex items-center justify-between py-2 border-b border-white/10">
              <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">Reward:</span>
              <span className="text-sm font-medium text-white/70">${Math.round(fight.data.reward)}</span>
            </div>
          )}

          <div className="flex items-center justify-between py-2">
            <span className="text-xs text-white/60 font-semibold uppercase tracking-wide">Game Time:</span>
            <span className="text-sm font-bold text-white text-right">
              {new Date(fight.end_time || fight.created_at).toLocaleString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
                hour: 'numeric',
                minute: '2-digit',
                hour12: true
              })} (in-game)
            </span>
          </div>
        </div>
      )}
    </div>
  );
}, (prevProps, nextProps) => {
  // Only re-render if fight data changes or boxer ID changes
  // currentGameTime can change frequently but doesn't affect card content for completed fights
  return prevProps.fight.id === nextProps.fight.id &&
         prevProps.myBoxerId === nextProps.myBoxerId;
});

/**
 * TimelineSection - Date-grouped section with header (MAT-103)
 */
const TimelineSection = ({ label, fights, myBoxerId, currentGameTime }) => {
  if (!fights || fights.length === 0) return null;

  return (
    <div className="space-y-4">
      <div className="relative mb-4">
        <h3 className="text-sm font-bold text-white/60 uppercase tracking-wide m-0">{label}</h3>
        <span className="inline-block px-2 py-1 bg-white/10 text-xs font-semibold text-white/60 rounded-md ml-2">
          {fights.length} fight{fights.length !== 1 ? 's' : ''}
        </span>
        <div className="h-1 mt-2 bg-gradient-to-r from-white/20 to-transparent"></div>
      </div>
      <div className="space-y-3">
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
    const loadFightHistory = async (isRetry = false) => {
      if (!boxerId) return;

      setIsLoading(true);
      if (isRetry) {
        setError(null);
      }

      try {
        const data = await fetchFightHistory(boxerId);

        // Ensure we always have an array
        const fightsArray = Array.isArray(data) ? data : [];
        setFights(fightsArray);

        // Group fights by date range (using game time or real time as fallback)
        if (fightsArray.length > 0) {
          const groups = currentGameTime
            ? groupFightsByDate(fightsArray, currentGameTime)
            : groupFightsByRealDate(fightsArray); // Fallback to real dates
          setGroupedFights(groups);
        }
      } catch (err) {
        // Use specific error message based on error type
        const errorMsg = getUserErrorMessage(err, 'fightHistory');
        setError(errorMsg);
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

      if (fightsArray.length > 0) {
        const groups = currentGameTime
          ? groupFightsByDate(fightsArray, currentGameTime)
          : groupFightsByRealDate(fightsArray); // Fallback to real dates
        setGroupedFights(groups);
      }
      setError(null);
    } catch (err) {
      // Use specific error message based on error type
      const errorMsg = getUserErrorMessage(err, 'fightHistory');
      setError(errorMsg);
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

  // Fallback: Group fights by real dates when game time is unavailable
  const groupFightsByRealDate = (fightsList) => {
    const now = new Date();
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
      return (
        <div className="p-8 text-center bg-gradient-to-br from-white/10 to-white/5 rounded-xl border border-white/10">
          <div className="animate-spin w-8 h-8 border-2 border-primary-light border-t-transparent rounded-full mx-auto mb-4"></div>
          <p className="text-white/60 font-medium">Loading fight history...</p>
        </div>
      );
    }

    if (error) {
      return (
        <ErrorDisplay
          error={error}
          onRetry={handleRefresh}
          variant="alert"
          title="Fight History Unavailable"
        />
      );
    }

    if (fights.length === 0) {
      return (
        <div className="p-12 text-center bg-gradient-to-br from-white/10 to-white/5 rounded-xl border border-white/10">
          <span className="text-5xl block mb-4" role="img" aria-label="empty scroll">📜</span>
          <h3 className="text-lg font-semibold text-white/80 m-0 mb-2">No Fights Yet</h3>
          <p className="text-white/60 m-0 mb-1">This boxer hasn't fought any matches yet.</p>
          <p className="text-sm text-white/50 italic m-0">Book a fight to see the action!</p>
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
    <div className="max-w-[900px] mx-auto my-8">
      <div className="flex items-center justify-between mb-6 pb-4 border-b border-white/10">
        <h2 className="text-2xl font-bold text-white m-0">Fight History</h2>
        <button
          onClick={handleRefresh}
          disabled={isLoading}
          className="inline-flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-primary to-primary-dark text-white rounded-lg font-semibold transition-all duration-150 hover:from-primary-dark hover:to-navy-medium hover:-translate-y-0.5 disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
          title="Refresh fight history"
        >
          {isLoading ? '⟳ Loading...' : '🔄 Refresh'}
        </button>
      </div>

      <div className="space-y-6">
        {renderContent()}
      </div>
    </div>
  );
};

export default FightHistory;
