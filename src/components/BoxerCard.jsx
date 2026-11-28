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
import { fetchActiveTrainingSessions } from '../api/training';

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

export function BoxerCard({ boxer, onTrainClick, onFightClick, onOpenFightBooking, upcomingFight, worldTime }) {
  // Training session — provided by the parent list when available (BoxersList
  // attaches it as `training_session`), otherwise fetched for this boxer.
  const [fetchedSession, setFetchedSession] = useState(null);

  useEffect(() => {
    // Only fetch when the parent did not already provide training status
    if (boxer.has_active_training !== undefined || boxer.training_session) return;

    let cancelled = false;
    fetchActiveTrainingSessions(boxer.id)
      .then((sessions) => {
        if (!cancelled && Array.isArray(sessions) && sessions.length > 0) {
          setFetchedSession(sessions[0]);
        }
      })
      .catch((err) => {
        console.debug('[BoxerCard] Failed to fetch training sessions:', err.message);
      });

    return () => { cancelled = true; };
  }, [boxer.id, boxer.has_active_training, boxer.training_session]);

  const trainingSession = boxer.training_session || fetchedSession || null;
  const inTraining = !!(boxer.has_active_training || trainingSession);

  // Rest status (from the API: has_active_rest + next_available_training)
  const isResting = !!boxer.has_active_rest;

  // Training countdown (seconds remaining, null when not determinable)
  const getTrainingRemaining = () => {
    if (!trainingSession) return null;
    let end = null;
    if (trainingSession.scheduled_completion_time) {
      end = new Date(trainingSession.scheduled_completion_time).getTime();
    } else if (trainingSession.start_time) {
      end = new Date(trainingSession.start_time).getTime() + (trainingSession.duration_hours || 0) * 3600000;
    }
    if (!end) return null;
    const now = worldTime?.current_game_time
      ? new Date(worldTime.current_game_time).getTime()
      : Date.now();
    return Math.max(0, Math.floor((end - now) / 1000));
  };

  const trainingRemaining = getTrainingRemaining();

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
  // API field is `fatigue_score` (not `fatigue`) — treat missing as 0
  const fatigueScore = boxer.fatigue_score ?? 0;
  const canFight = boxer.health >= 50 && boxer.energy >= 30 && fatigueScore <= 70;
  const canTrain = boxer.energy >= 20 && !inTraining && !isResting;

  const trainButtonDisabledReason = !canTrain
    ? (inTraining ? 'Training in progress' : isResting ? 'Currently resting' : 'Not enough energy')
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
      {inTraining && (
        <div className={trainingBadgeStyles} role="status" aria-live="polite">
          <span className="text-xs animate-pulse" aria-hidden="true">⏳</span>
          <span className="training-status">
            {trainingRemaining !== null
              ? `Training - ${formatTimeRemaining(trainingRemaining)}`
              : 'Training'}
          </span>
        </div>
      )}

      {/* Resting Badge (MAT-96) */}
      {isResting && (
        <RestingBadge
          hasActiveRest={isResting}
          restEndsAt={boxer.next_available_training}
          forcedRestUntil={boxer.forced_rest_until}
          currentGameTime={worldTime?.current_game_time}
        />
      )}

      {/* Upcoming Fight Badge */}
      {upcomingFight && upcomingFight.status === 'scheduled' && !inTraining && (
        <div className="fight-badge absolute -top-3 left-1/2 -translate-x-1/2 z-30">
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
            <h3 className="text-lg font-bold m-0">
              <Link
                to={`/boxer/${boxer.id}`}
                className="text-white truncate hover:text-primary-light hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 rounded"
                aria-label={`View profile of ${boxer.name}`}
              >
                {boxer.name}
              </Link>
            </h3>
            {boxer.nickname && (
              <p className="boxer-nickname text-sm text-gray-400 italic truncate">
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
              animated={false}
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
              animated={false}
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
              value={boxer.fatigue_score || 0}
              max={100}
              variant="fatigue"
              animated={false}
              className="flex-1"
            />
            <span className="text-xs font-bold text-white w-12 text-right flex-shrink-0">
              {Math.round(boxer.fatigue_score || 0)}%
            </span>
          </div>
        </div>

        {/* Combat Stats Grid */}
        <div className="grid grid-cols-3 gap-2 mb-4">
          <StatDisplay
            label="STR"
            value={Math.round(boxer.strength || 0)}
            icon="💪"
            color="strength"
            variant="compact"
          />
          <StatDisplay
            label="DEF"
            value={Math.round(boxer.defense || 0)}
            icon="🛡️"
            color="defense"
            variant="compact"
          />
          <StatDisplay
            label="AGI"
            value={Math.round(boxer.agility || 0)}
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
            onClick={() => {
              if (onOpenFightBooking) {
                onOpenFightBooking();
              } else if (onFightClick) {
                onFightClick(boxer.id);
              }
            }}
            disabled={!canFight || !!upcomingFight}
            title={!canFight ? 'Health/Energy/Fatigue requirements not met' : upcomingFight ? 'Fight already scheduled' : 'Book a Fight'}
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
