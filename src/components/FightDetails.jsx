/**
 * FightDetails - Theater mode arena page (MAT-104)
 *
 * Displays comprehensive fight results with:
 * - Victory banner for completed fights
 * - Side-by-side boxer comparison
 * - Round timeline visualization
 * - Play-by-play fight narrative
 * - Fight summary statistics
 */

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import PageLayout from './PageLayout';
import { LoadingSpinner, StatDisplay, Badge } from './ui';
import RoundTimeline from './RoundTimeline';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage, isNotFoundError } from '../utils/errorMessages';
import { fetchFightDetails, fetchBoxer } from '../api/fights';

/**
 * VictoryBanner - Displays fight result prominently (MAT-104)
 */
const VictoryBanner = ({ winner, loser, method, round }) => {
  if (!winner && loser === null) {
    return (
      <div className="py-6 px-8 rounded-xl text-center mb-8 bg-gradient-to-br from-gray-600 to-gray-700 text-white shadow-[0_8px_24px_rgba(75,85,99,0.5)] animate-[bannerSlideIn_0.5s_ease]">
        <h1 className="m-0 mb-2 text-2xl font-extrabold tracking-wide uppercase">🤝 IT'S A DRAW! 🤝</h1>
        <p className="m-0 text-base font-medium opacity-95">Both boxers fought valiantly to a standstill</p>
      </div>
    );
  }

  const isKO = method === 'knockout';
  const victoryEmoji = isKO ? '🔥' : '🏆';
  const bannerStyle = isKO
    ? 'bg-gradient-to-br from-red-400 to-danger text-white shadow-[0_8px_24px_rgba(255,107,107,0.5)]'
    : 'bg-gradient-to-br from-victory-gold to-warning text-text-on-gold shadow-[0_8px_24px_rgba(214,158,46,0.4)]';

  return (
    <div className={`py-6 px-8 rounded-xl text-center mb-8 ${bannerStyle} animate-[bannerSlideIn_0.5s_ease]`}>
      <h1 className="m-0 mb-2 text-2xl font-extrabold tracking-wide uppercase">
        {victoryEmoji} {winner.name.toUpperCase()} WINS {victoryEmoji}
      </h1>
      <p className="m-0 text-base font-medium opacity-95">
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
    if (!health) return 'bg-gradient-to-r from-warning to-warning-dark';
    const pct = health;
    if (pct > 50) return 'bg-gradient-to-r from-success to-success-darker';
    if (pct > 25) return 'bg-gradient-to-r from-warning to-warning-dark';
    return 'bg-gradient-to-r from-health-low-start to-health-low-end';
  };

  const avatar = getAvatarEmoji(boxer.name);
  const healthClass = getHealthColorClass(healthAtEnd);

  // Card base styles with winner glow effect
  const cardBase = 'flex-1 max-w-[320px] bg-white/5 rounded-xl p-6 border-2 border-white/15 transition-all duration-200 relative hover:shadow-[0_8px_24px_rgba(0,0,0,0.3)] hover:-translate-y-2';
  const cardStyles = isWinner
    ? `${cardBase} border-warning/80 bg-gradient-to-br from-yellow-300/10 to-amber-100/5 shadow-[0_8px_24px_rgba(214,158,46,0.3)]`
    : cardBase;

  return (
    <div className={cardStyles}>
      {/* Avatar */}
      <div className="absolute -top-7 left-1/2 -translate-x-1/2 w-[56px] h-[56px] bg-gradient-to-br from-navy-medium to-navy-light rounded-full flex items-center justify-center text-xl shadow-[0_4px_12px_rgba(0,0,0,0.4)] border-3 border-navy-dark">
        {avatar}
      </div>

      {/* Header */}
      <div className="text-center pt-9 mb-4">
        <h4 className="m-0 mb-1 text-lg font-bold text-white">{boxer.name}</h4>
        {boxer.nickname && <p className="m-0 mb-3 text-xs text-white/60 italic">"{boxer.nickname}"</p>}
        <Badge variant={isWinner ? 'winner' : 'level'} size="sm">LVL {boxer.level || 1}</Badge>
      </div>

      {/* Health row */}
      {healthAtEnd !== undefined && (
        <div className="flex items-center gap-3 mb-4 p-3 bg-white/5 rounded-md">
          <span className="text-xs font-semibold text-white/70">Health:</span>
          <div className="flex-1 h-2 bg-white/10 rounded-xs overflow-hidden">
            <div
              className={`h-full ${healthClass}`}
              style={{ width: `${Math.min(100, Math.max(0, healthAtEnd))}%` }}
            />
          </div>
          <span className="text-sm font-semibold text-white min-w-[40px] text-right">{Math.round(healthAtEnd)}%</span>
        </div>
      )}

      {/* Combat Stats */}
      <div className="grid grid-cols-3 gap-2 mb-4">
        <StatDisplay label="STR" value={boxer.strength} icon="💪" color="strength" variant="compact" />
        <StatDisplay label="DEF" value={boxer.defense} icon="🛡️" color="defense" variant="compact" />
        <StatDisplay label="AGI" value={boxer.agility} icon="⚡" color="agility" variant="compact" />
      </div>

      {/* Footer */}
      <div className="border-t border-white/10 pt-3 mt-4 text-center">
        <span className="text-xs text-white/50 font-semibold block mb-1">Record:</span>
        <span className="text-sm text-white font-semibold">
          {boxer.wins || 0}W - {boxer.losses || 0}L - {boxer.draws || 0}D
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

  const generatePlayByPlay = () => {
    if (!fightData || !fightData.log) return null;

    const logLines = fightData.log.split('\n').filter(line => line.trim());

    const getLogClass = (line) => {
      let cls = 'p-2 mb-1 text-sm leading-relaxed rounded-xs';
      if (line.includes('WINS') || line.includes('Victory')) cls += ' bg-gradient-to-r from-warning/25 to-warning/5 text-yellow-300';
      else if (line.includes('knockout') || line.includes('KO')) cls += ' bg-gradient-to-r from-red-400/25 to-red-400/5 text-red-400';
      else if (line.includes('damage') || line.includes('-')) cls += ' bg-white/8';
      else if (line.includes('MISS') || line.includes('dodge')) cls += ' bg-gradient-to-r from-teal-300/20 to-teal-300/5 text-teal-200';
      else if (line.includes('Exchange')) cls += ' bg-gradient-to-r from-red-300/20 to-red-300/5 text-red-200';
      return cls;
    };

    return (
      <div className="max-h-[300px] overflow-y-auto">
        <h4 className="text-center text-white text-sm mb-3 pb-2 border-b border-white/10">🥊 Round {selectedRound} Play-by-Play 🥊</h4>

        {logLines.length > 0 ? (
          <div className="flex flex-col gap-1 p-4 bg-black/30 rounded-lg font-mono">
            {logLines.map((line, index) => (
              <div key={index} className={getLogClass(line)}>
                {line}
              </div>
            ))}
          </div>
        ) : (
          <p className="text-center text-white/50 italic p-5">No detailed log available for this round.</p>
        )}

        {/* Navigation */}
        <div className="flex flex-wrap justify-between items-center gap-3 mt-4">
          <button
            className="bg-gradient-to-br from-navy-medium to-navy-light text-white border border-white/20 px-4 py-2 rounded-md text-sm font-semibold cursor-pointer transition-all duration-200 flex-1 hover:bg-gradient-to-br hover:from-navy-light hover:to-navy-dark hover:border-white/30 hover:-translate-y-1 hover:shadow-[0_4px_8px_rgba(0,0,0,0.4)] disabled:bg-white/10 disabled:text-white/40 disabled:cursor-not-allowed focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            onClick={() => setSelectedRound(Math.max(1, selectedRound - 1))}
            disabled={selectedRound === 1}
          >
            ← Previous Round
          </button>
          <span className="text-sm font-semibold text-white/80 whitespace-nowrap">Round {selectedRound} of {rounds}</span>
          <button
            className="bg-gradient-to-br from-navy-medium to-navy-light text-white border border-white/20 px-4 py-2 rounded-md text-sm font-semibold cursor-pointer transition-all duration-200 flex-1 hover:bg-gradient-to-br hover:from-navy-light hover:to-navy-dark hover:border-white/30 hover:-translate-y-1 hover:shadow-[0_4px_8px_rgba(0,0,0,0.4)] disabled:bg-white/10 disabled:text-white/40 disabled:cursor-not-allowed focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
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
 *
 * Theater mode page wrapper for fight details view.
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
        const fightData = await fetchFightDetails(id);
        setFight(fightData);

        const [boxer1Data, boxer2Data] = await Promise.all([
          fetchBoxer(fightData.boxer1_id),
          fetchBoxer(fightData.boxer2_id)
        ]);

        setBoxer1(boxer1Data);
        setBoxer2(boxer2Data);

        if (fightData.winner_id) {
          const winnerBoxer = fightData.winner_id === boxer1Data.id ? boxer1Data : boxer2Data;
          setWinner(winnerBoxer);
        } else if (fightData.status === 'completed') {
          setWinner(null);
        }
      } catch (err) {
        const errorMsg = getUserErrorMessage(err, 'general');
        setError(errorMsg);
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
  const fightData = fight?.data || {};
  const isFightCompleted = fight?.status === 'completed';
  const roundsFought = isFightCompleted ? (fight?.round || 12) : 0;
  const method = isFightCompleted ? (fightData?.method || 'decision') : null;

  // Get cross-nav target
  const getFightHistoryUrl = () => {
    if (!boxer1?.id && !fight?.boxer1_id) return null;
    const boxerId = boxer1?.id || fight?.boxer1_id;
    return `/boxer/${boxerId}/fights`;
  };

  const fightHistoryUrl = getFightHistoryUrl();

  // Section card base style
  const sectionCardBase = 'bg-white/8 backdrop-blur-sm border border-white/20 rounded-xl p-6 mb-6';

  if (isLoading) {
    return (
      <PageLayout
        pageClassName="fight-details-page"
        onBack={() => navigate('/dashboard')}
        backButtonLabel="← Back to Dashboard"
        showCrossNav={false}
      >
        <LoadingSpinner size="lg" label="Loading fight details..." />
      </PageLayout>
    );
  }

  if (error) {
    return (
      <PageLayout
        pageClassName="fight-details-page"
        onBack={() => navigate('/dashboard')}
        backButtonLabel="← Back to Dashboard"
      >
        <ErrorDisplay
          error={error}
          variant="alert"
          title={error?.includes('not found') ? 'Fight Not Found' : 'Error Loading Fight'}
          icon={true}
        />
      </PageLayout>
    );
  }

  return (
    <PageLayout
      pageClassName="fight-details-page"
      onBack={() => navigate('/dashboard')}
      backButtonLabel="← Back to Dashboard"
      crossNavTo={fightHistoryUrl}
      crossNavLabel="📜 Fight History ▶"
    >
      {/* Fight Info Header */}
      <div className={`${sectionCardBase} text-center`}>
        <h2 className="text-white/80 text-sm font-bold uppercase tracking-widest m-0 mb-4">
          {boxer1?.name || 'Boxer 1'} vs {boxer2?.name || 'Boxer 2'}
        </h2>

        {/* Status Badges */}
        {fight?.status === 'scheduled' && <Badge variant="fight-status-scheduled" icon="📅">Scheduled</Badge>}
        {fight?.status === 'in_progress' && <Badge variant="winner" icon="🥊">In Progress</Badge>}
        {fight?.status === 'completed' && <Badge variant="fight-status-completed" icon="✅">Completed</Badge>}
        {(fight?.status === 'canceled' || fight?.status === 'cancelled') && <Badge variant="fight-status-cancelled" icon="❌">Canceled</Badge>}

        {/* Scheduled Time */}
        {fight?.status === 'scheduled' && fight?.scheduled_time && (
          <>
            <p className="text-sm text-white/70 mt-2 m-0">
              Scheduled for Game Time:{' '}
              {new Date(fight.scheduled_time).toLocaleDateString('en-US', {
                year: 'numeric', month: 'long', day: 'numeric'
              })} at{' '}
              {new Date(fight.scheduled_time).toLocaleTimeString('en-US', {
                hour: 'numeric', minute: '2-digit', hour12: true
              })}
            </p>
            <p className="text-sm text-white/60 italic mt-2 p-3 bg-indigo-500/15 rounded-md m-0">
              This fight is scheduled for a future game time. Results will appear here once the worker processes the fight simulation.
            </p>
          </>
        )}

        {/* End Time for completed fights */}
        {fight?.status === 'completed' && (
          <p className="text-sm text-white/70 mt-2 m-0">
            📅 Completed at Game Time:{' '}
            {new Date(fight.end_time || fight.created_at).toLocaleDateString('en-US', {
              year: 'numeric', month: 'long', day: 'numeric'
            })} at{' '}
            {new Date(fight.end_time || fight.created_at).toLocaleTimeString('en-US', {
              hour: 'numeric', minute: '2-digit', hour12: true
            })}
          </p>
        )}
      </div>

      {/* Victory Banner */}
      {isFightCompleted && (
        <VictoryBanner
          winner={winner}
          loser={fightData.winner_id ? opponent : null}
          method={method}
          round={roundsFought}
        />
      )}

      {/* Boxer Comparison */}
      <div className={sectionCardBase}>
        <h3 className="text-white/80 text-sm font-bold uppercase tracking-widest m-0 mb-5 text-center">FIGHTER COMPARISON</h3>
        <div className="flex items-stretch justify-center gap-6 relative flex-wrap">
          <BoxerComparisonCard boxer={boxer1} role="boxer1" fightData={isFightCompleted ? fightData : undefined} />

          <div className="text-lg font-extrabold text-white p-3 bg-gradient-to-br from-danger-brighter to-danger rounded-full w-[56px] h-[56px] flex items-center justify-center shadow-[0_4px_12px_rgba(239,68,68,0.6)] flex-shrink-0">VS</div>

          <BoxerComparisonCard boxer={boxer2} role="boxer2" fightData={isFightCompleted ? fightData : undefined} />
        </div>
      </div>

      {/* Round Timeline */}
      {isFightCompleted && roundsFought > 0 && (
        <div className={sectionCardBase}>
          <h3 className="text-white/80 text-sm font-bold uppercase tracking-widest m-0 mb-5 text-center">ROUND TIMELINE</h3>
          <RoundTimeline totalRounds={roundsFought} method={method} />
        </div>
      )}

      {/* Play-by-Play Log */}
      {isFightCompleted && fightData?.log && (
        <div className={sectionCardBase}>
          <h3 className="text-white/80 text-sm font-bold uppercase tracking-widest m-0 mb-5 text-center">FIGHT NARRATIVE</h3>
          <PlayByPlayLog fightData={fightData} rounds={roundsFought} />
        </div>
      )}

      {/* Fight Summary Stats */}
      <div className={sectionCardBase}>
        <h3 className="text-white/80 text-sm font-bold uppercase tracking-widest m-0 mb-5 text-center">FIGHT SUMMARY</h3>
        <div className="grid grid-cols-auto-fit gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))' }}>
          <div className="bg-white/5 rounded-lg p-5 text-center transition-all duration-200 hover:-translate-y-2 hover:shadow-[0_8px_16px_rgba(0,0,0,0.2)]">
            <span className="text-xs text-white/50 font-semibold uppercase tracking-wider block mb-2">Status</span>
            <span className={`text-lg font-bold ${fight?.status === 'completed' ? 'text-success' : 'text-white'}`}>
              {fight?.status || 'Unknown'}
            </span>
          </div>

          {isFightCompleted && (
            <>
              <div className="bg-white/5 rounded-lg p-5 text-center transition-all duration-200 hover:-translate-y-2 hover:shadow-[0_8px_16px_rgba(0,0,0,0.2)]">
                <span className="text-xs text-white/50 font-semibold uppercase tracking-wider block mb-2">Rounds Fought</span>
                <span className="text-lg font-bold text-white">{roundsFought}</span>
              </div>

              {method && (
                <div className="bg-white/5 rounded-lg p-5 text-center transition-all duration-200 hover:-translate-y-2 hover:shadow-[0_8px_16px_rgba(0,0,0,0.2)]">
                  <span className="text-xs text-white/50 font-semibold uppercase tracking-wider block mb-2">Method of Victory</span>
                  <span className="text-lg font-bold text-white capitalize">{method}</span>
                </div>
              )}

              {fightData?.xp_gained && (
                <div className="bg-white/5 rounded-lg p-5 text-center transition-all duration-200 hover:-translate-y-2 hover:shadow-[0_8px_16px_rgba(0,0,0,0.2)]">
                  <span className="text-xs text-white/50 font-semibold uppercase tracking-wider block mb-2">XP Gained</span>
                  <span className="text-lg font-bold text-[#d69e2e]">+{Math.round(fightData.xp_gained)}</span>
                </div>
              )}

              {fightData?.reward && (
                <div className="bg-white/5 rounded-lg p-5 text-center transition-all duration-200 hover:-translate-y-2 hover:shadow-[0_8px_16px_rgba(0,0,0,0.2)]">
                  <span className="text-xs text-white/50 font-semibold uppercase tracking-wider block mb-2">Purse Earned</span>
                  <span className="text-lg font-bold text-primary">${Math.round(fightData.reward).toLocaleString()}</span>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* Boxer Links */}
      {boxer1 && boxer2 && (
        <div className={sectionCardBase}>
          <h3 className="text-white/80 text-sm font-bold uppercase tracking-widest m-0 mb-5 text-center">FIGHTERS</h3>
          <div className="flex gap-4 justify-center flex-wrap">
            <button
              className="bg-white/10 text-white border border-white/30 px-7 py-3 rounded-md text-sm font-semibold cursor-pointer transition-all duration-200 flex items-center gap-2 hover:bg-white/15 hover:border-white/50 hover:-translate-y-1 hover:shadow-[0_4px_16px_rgba(255,255,255,0.2)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
              onClick={() => navigate(`/boxer/${boxer1.id}`)}
            >
              📄 {boxer1.name} Profile
              <span className="text-xs bg-gradient-to-br from-navy-medium to-navy-light px-2 py-1 rounded-sm ml-1">📊 History</span>
            </button>
            <button
              className="bg-white/10 text-white border border-white/30 px-7 py-3 rounded-md text-sm font-semibold cursor-pointer transition-all duration-200 flex items-center gap-2 hover:bg-white/15 hover:border-white/50 hover:-translate-y-1 hover:shadow-[0_4px_16px_rgba(255,255,255,0.2)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
              onClick={() => navigate(`/boxer/${boxer2.id}`)}
            >
              📄 {boxer2.name} Profile
              <span className="text-xs bg-gradient-to-br from-navy-medium to-navy-light px-2 py-1 rounded-sm ml-1">📊 History</span>
            </button>
          </div>
        </div>
      )}
    </PageLayout>
  );
};

export default FightDetails;
