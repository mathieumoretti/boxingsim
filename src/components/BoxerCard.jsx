/**
 * BoxerCard Component
 *
 * Comprehensive boxer display card with training session monitoring,
 * fight countdowns, rest status tracking, and action buttons.
 * Used in Dashboard, Boxers list, and Rankings views.
 */

import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import ProgressBar from './ui/ProgressBar';
import StatDisplay from './ui/StatDisplay';
import FightBadge from './FightBadge';
import RestingBadge from './RestingBadge';

/**
 * Generate consistent avatar emoji based on name hash
 */
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

/**
 * Calculate level tier for badge styling (1-5, champion at 5)
 */
const getLevelTier = (level) => {
  if (!level || level < 1) return 1;
  if (level <= 10) return 1;
  if (level <= 25) return 2;
  if (level <= 50) return 3;
  if (level <= 100) return 4;
  return 5; // Champion tier
};

/**
 * Format time remaining for training countdown
 */
const formatTimeRemaining = (seconds) => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
};

export function BoxerCard({ boxer, onTrainClick, onFightClick }) {
  // Training session state
  const [activeSession, setActiveSession] = useState(boxer.active_session);
  const [countdown, setCountdown] = useState(null);

  // Rest status polling state (MAT-96)
  const [isResting, setIsResting] = useState(boxer.is_resting);
  const [restEndTime, setRestEndTime] = useState(boxer.rest_end_time);

  // Upcoming fight state
  const [upcomingFight, setUpcomingFight] = useState(null);

  // Fetch upcoming fight for this boxer
  useEffect(() => {
    const fetchUpcomingFight = async () => {
      try {
        const response = await fetch(`/api/fights/upcoming/${boxer.id}`);
        if (response.ok) {
          const fight = await response.json();
          setUpcomingFight(fight);
        }
      } catch (err) {
        // 404 is OK - no upcoming fight
        if (err.status !== 404) {
          console.debug('[BoxerCard] Failed to fetch upcoming fight:', err.message);
        }
      }
    };

    fetchUpcomingFight();
  }, [boxer.id]);

  // Poll training session end time every second when active
  useEffect(() => {
    if (!activeSession || !activeSession.end_time) return;

    const updateCountdown = () => {
      const now = new Date().getTime();
      const endTime = new Date(activeSession.end_time).getTime();
      const remaining = Math.max(0, Math.floor((endTime - now) / 1000));

      setCountdown(remaining);

      // Training complete
      if (remaining <= 0) {
        setActiveSession(null);
        setCountdown(null);
      }
    };

    updateCountdown(); // Initial update
    const interval = setInterval(updateCountdown, 1000);
    return () => clearInterval(interval);
  }, [activeSession]);

  // Poll rest status every 5 seconds (MAT-96)
  useEffect(() => {
    if (!boxer.is_resting || !boxer.rest_end_time) {
      setIsResting(false);
      setRestEndTime(null);
      return;
    }

    const updateRestStatus = async () => {
      try {
        const response = await fetch(`/api/boxers/${boxer.id}`);
        if (response.ok) {
          const updatedBoxer = await response.json();
          setIsResting(updatedBoxer.is_resting);
          setRestEndTime(updatedBoxer.rest_end_time);

          // Stop polling if rest complete
          if (!updatedBoxer.is_resting) {
            setRestEndTime(null);
          }
        }
      } catch (err) {
        console.debug('[BoxerCard] Failed to refresh boxer data:', err.message);
      }
    };

    const interval = setInterval(updateRestStatus, 5000);
    return () => clearInterval(interval);
  }, [boxer.is_resting, boxer.id]);

  // Health/Energy/Fatigue color thresholds
  const getHealthColorClass = (value) => {
    if (value >= 80) return 'bg-health-high';
    if (value >= 50) return 'bg-health-medium';
    return 'bg-health-low';
  };

  const getFatigueColorClass = (value) => {
    if (value < 30) return 'bg-fatigue-ok';
    if (value < 70) return 'bg-fatigue-warning';
    return 'bg-fatigue-critical';
  };

  // Level badge styles by tier
  const levelBadgeStyles = {
    1: 'bg-gradient-to-br from-slate-400 to-slate-500 text-white',
    2: 'bg-gradient-to-br from-blue-400 to-blue-600 text-white shadow-[0_2px_8px_rgba(37,99,235,0.4)]',
    3: 'bg-gradient-to-br from-primary to-primary-dark text-white shadow-[0_2px_12px_rgba(37,99,235,0.5)]',
    4: 'bg-gradient-to-br from-victory-gold to-warning text-white shadow-[0_2px_16px_rgba(214,158,46,0.6)]',
    5: 'bg-gradient-to-br from-yellow-300 via-warning to-orange-500 text-white animate-championGlow shadow-[0_4px_24px_rgba(214,158,46,0.8),0_0_40px_rgba(253,224,71,0.4)]',
  };

  const levelTier = getLevelTier(boxer.level);
  const avatarEmoji = getAvatarEmoji(boxer.name);

  // Training badge styles with pulse animation
  const trainingBadgeBase = 'absolute -top-2 left-1/2 -translate-x-1/2 px-4 py-1 rounded-full text-xs font-bold shadow-lg flex items-center gap-1.5 z-20';
  const trainingBadgeStyles = `${trainingBadgeBase} bg-gradient-to-r from-primary to-primary-dark text-white animate-trainingPulse`;

  // Rest badge styles (teal for recovery, distinct from training)
  const restBadgeBase = 'absolute -bottom-2 right-1/2 translate-x-1/2 px-3 py-1 rounded-full text-xs font-semibold shadow-lg z-20';
  const restBadgeStyles = `${restBadgeBase} bg-gradient-to-r from-teal-500 to-teal-700 text-white ${isResting && boxer.forced_rest ? 'ring-2 ring-danger' : ''}`;

  // Button disabled states based on combat readiness
  const canFight = boxer.health >= 50 && boxer.energy >= 30 && boxer.fatigue <= 70;
  const canTrain = boxer.energy >= 20 && !activeSession && !isResting;

  const trainButtonDisabledReason = !canTrain
    ? (activeSession ? 'Training in progress' : boxer.energy < 20 ? 'Not enough energy' : 'Currently resting')
    : null;

  return (
    <div className="boxer-card relative bg-gradient-to-br from-navy-medium to-navy-dark rounded-2xl p-5 shadow-[0_8px_40px_rgba(0,0,0,0.6),0_2px_12px_rgba(0,0,0,0.4)] border border-navy-border transition-all duration-250 hover:shadow-[0_12px_48px_rgba(0,0,0,0.7),0_4px_16px_rgba(0,0,0,0.5)] hover:-translate-y-1 group overflow-hidden">
      {/* Card glow effect (::before in CSS) */}
      <div className="absolute inset-0 rounded-2xl bg-gradient-to-br from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />

      {/* Level Badge (absolute positioning) */}
      <div
        className={`absolute top-4 right-4 px-3 py-1.5 rounded-lg text-sm font-bold shadow-lg z-10 ${levelBadgeStyles[levelTier]}`}
      >
        LVL {boxer.level || 1}
      </div>

      {/* Training in Progress Badge */}
      {activeSession && countdown !== null && (
        <div className={trainingBadgeStyles} role="status" aria-live="polite">
          <span className="text-xs animate-pulse">⏳</span>
          <span>Training - {formatTimeRemaining(countdown)}</span>
        </div>
      )}

      {/* Resting Badge (MAT-96) */}
      {isResting && (
        <RestingBadge
          isForced={boxer.forced_rest}
          endTime={restEndTime}
          className={restBadgeStyles.replace('bg-gradient-to-r from-indigo-500 to-indigo-700 text-white', '')}
        />
      )}

      {/* Upcoming Fight Badge */}
      {upcomingFight && !activeSession && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2 z-30">
          <FightBadge fight={upcomingFight} />
        </div>
      )}

      {/* Card Content */}
      <div className="relative z-10">
        {/* Header: Avatar + Name */}
        <div className="flex items-start gap-4 mb-4">
          {/* Avatar */}
          <div className="w-16 h-16 rounded-full bg-gradient-to-br from-victory-gold to-warning flex items-center justify-center text-2xl shadow-[0_4px_12px_rgba(214,158,46,0.3)] flex-shrink-0">
            {avatarEmoji}
          </div>

          {/* Name + Nickname */}
          <div className="flex-1 min-w-0">
            <h3 className="text-lg font-bold text-white truncate">
              {boxer.name}
            </h3>
            {boxer.nickname && (
              <p className="text-sm text-gray-400 italic truncate">
                "{boxer.nickname}"
              </p>
            )}
          </div>
        </div>

        {/* Status Bars Section */}
        <div className="space-y-2.5 mb-4">
          {/* Health */}
          <div className="flex items-center gap-3">
            <span className="text-xs font-semibold text-gray-300 w-16 flex-shrink-0">Health</span>
            <ProgressBar
              value={boxer.health || 0}
              max={100}
              variant="health"
              animated={!activeSession}
              className="flex-1"
            />
            <span className="text-xs font-bold text-white w-12 text-right flex-shrink-0">
              {Math.round(boxer.health || 0)}%
            </span>
          </div>

          {/* Energy */}
          <div className="flex items-center gap-3">
            <span className="text-xs font-semibold text-gray-300 w-16 flex-shrink-0">Energy</span>
            <ProgressBar
              value={boxer.energy || 0}
              max={100}
              variant="energy"
              animated={!activeSession}
              className="flex-1"
            />
            <span className="text-xs font-bold text-white w-12 text-right flex-shrink-0">
              {Math.round(boxer.energy || 0)}%
            </span>
          </div>

          {/* Fatigue */}
          <div className="flex items-center gap-3">
            <span className="text-xs font-semibold text-gray-300 w-16 flex-shrink-0">Fatigue</span>
            <ProgressBar
              value={boxer.fatigue || 0}
              max={100}
              variant="fatigue"
              animated={!activeSession}
              className="flex-1"
            />
            <span className="text-xs font-bold text-white w-12 text-right flex-shrink-0">
              {Math.round(boxer.fatigue || 0)}%
            </span>
          </div>
        </div>

        {/* Combat Stats Grid */}
        <div className="grid grid-cols-3 gap-2 mb-4">
          <StatDisplay
            label="STR"
            value={boxer.strength || 0}
            icon="💪"
            color="strength"
            variant="compact"
          />
          <StatDisplay
            label="DEF"
            value={boxer.defense || 0}
            icon="🛡️"
            color="defense"
            variant="compact"
          />
          <StatDisplay
            label="AGI"
            value={boxer.agility || 0}
            icon="⚡"
            color="agility"
            variant="compact"
          />
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2.5 pt-3 border-t border-white/10">
          {/* Train Button */}
          <Link
            to={`/boxer/${boxer.id}/train`}
            onClick={(e) => {
              if (trainButtonDisabledReason) {
                e.preventDefault();
              } else if (onTrainClick) {
                onTrainClick(boxer.id);
              }
            }}
            title={trainButtonDisabledReason || 'Schedule Training'}
            className={`
              flex-1 px-4 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200 cursor-pointer
              flex items-center justify-center gap-2
              ${canTrain
                ? 'bg-gradient-to-br from-success to-emerald-600 text-white shadow-[0_2px_8px_rgba(16,185,129,0.4)] hover:bg-gradient-to-br hover:from-emerald-400 hover:to-success hover:shadow-[0_4px_12px_rgba(16,185,129,0.6)] hover:-translate-y-0.5 active:translate-y-0 active:shadow-subtle focus-visible:outline focus-visible:outline-2 focus-visible:outline-success focus-visible:outline-offset-2'
                : 'bg-gray-700/50 text-gray-400 cursor-not-allowed shadow-none'}
            `.trim()}
            aria-disabled={!canTrain}
            tabIndex={canTrain ? 0 : -1}
          >
            <span className="text-base">🏋️</span>
            <span className="hidden sm:inline">Train</span>
            <span className="sm:hidden">T</span>
          </Link>

          {/* Fight Button */}
          <button
            onClick={() => onFightClick && onFightClick(boxer.id)}
            disabled={!canFight || !!upcomingFight}
            title={!canFight ? `Not ready: ${canFight ? '' : 'Health/Energy/Fatigue requirements not met'}` : upcomingFight ? 'Fight already scheduled' : 'Book a Fight'}
            className={`
              px-4 py-2.5 rounded-lg text-sm font-semibold transition-all duration-200 cursor-pointer
              flex items-center justify-center gap-2 flex-shrink-0
              ${(!canFight || upcomingFight)
                ? 'bg-gray-700/50 text-gray-400 cursor-not-allowed shadow-none'
                : 'bg-gradient-to-br from-warning to-orange-600 text-white shadow-[0_2px_8px_rgba(245,158,11,0.4)] hover:bg-gradient-to-br hover:from-amber-400 hover:to-warning hover:shadow-[0_4px_12px_rgba(245,158,11,0.6)] hover:-translate-y-0.5 active:translate-y-0 active:shadow-subtle focus-visible:outline focus-visible:outline-2 focus-visible:outline-warning focus-visible:outline-offset-2'}
            `.trim()}
            aria-disabled={!canFight || !!upcomingFight}
          >
            <span className="text-base">🥊</span>
            <span className="hidden sm:inline">Fight</span>
            <span className="sm:hidden">F</span>
          </button>

          {/* History Button */}
          <Link
            to={`/boxer/${boxer.id}/fights`}
            className="px-4 py-2.5 rounded-lg text-sm font-semibold bg-gradient-to-br from-navy-light to-navy-medium text-white border border-white/20 shadow-[0_2px_8px_rgba(0,0,0,0.3)] hover:bg-gradient-to-br hover:from-navy-dark hover:to-navy-light hover:border-white/30 hover:shadow-[0_4px_12px_rgba(0,0,0,0.4)] hover:-translate-y-0.5 active:translate-y-0 active:shadow-subtle transition-all duration-200 cursor-pointer flex items-center justify-center gap-2 flex-shrink-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2"
          >
            <span className="text-base">📜</span>
            <span className="hidden sm:inline">History</span>
            <span className="sm:hidden">H</span>
          </Link>
        </div>
      </div>
    </div>
  );
}

export default BoxerCard;
