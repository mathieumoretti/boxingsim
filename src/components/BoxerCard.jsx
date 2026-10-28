import React, { useState, useEffect, useCallback } from 'react';
import { ProgressBar, StatDisplay } from './ui';
import './BoxerCard.css';
import { fetchUpcomingFight, fetchBoxer } from '../api/fights';
import { fetchActiveTrainingSessions } from '../api/training';
import RestingBadge from './RestingBadge';
import FightBadge from './FightBadge';

// Training type mapping with enhanced icons
const trainingTypeConfig = {
  strength_training: { icon: '💪', label: 'Strength' },
  speed_training: { icon: '⚡', label: 'Speed' },
  defense_training: { icon: '🛡️', label: 'Defense' },
  technique_training: { icon: '🥊', label: 'Technique' },
  endurance_training: { icon: '🏃', label: 'Endurance' },
};

const trainingTypeIds = {
  1: 'strength_training',
  2: 'speed_training',
  3: 'defense_training',
  4: 'technique_training',
  5: 'endurance_training',
};

const BoxerCard = ({ boxer, worldTime, upcomingFight, onOpenProfile, onOpenTraining, onBoxerStateChanged, onOpenFightBooking, onOpenFightHistory }) => {
  const [activeSession, setActiveSession] = useState(null);
  const [countdown, setCountdown] = useState('');
  const [fetchedFight, setFetchedFight] = useState(upcomingFight || null);

  // Get training type name from ID
  const getTrainingTypeName = (typeId) => trainingTypeIds[typeId] || 'strength_training';

  // Format countdown display
  const formatCountdown = (remainingSeconds) => {
    if (remainingSeconds <= 0) return '';

    const hours = Math.floor(remainingSeconds / 3600);
    const minutes = Math.floor((remainingSeconds % 3600) / 60);
    const seconds = remainingSeconds % 60;

    if (hours > 0) return `${hours}h ${minutes}m`;
    if (minutes > 0) return `${minutes}m ${seconds}s`;
    return `${seconds}s`;
  };

  // Fetch active training session for this boxer
  const loadActiveTraining = useCallback(async () => {
    try {
      const sessions = await fetchActiveTrainingSessions(boxer.id);

      // Find the most recent pending session
      const pendingSession = sessions.find(s => s.status === 'pending');
      setActiveSession(pendingSession);
    } catch (err) {
      // Log for debugging - boxer card still works without training status
      console.debug(`BoxerCard (${boxer.name}): Failed to load active training`, err);
    }
  }, [boxer.id, boxer.name]);

  useEffect(() => {
    if (boxer.id) {
      loadActiveTraining();
    }
  }, [boxer.id, loadActiveTraining]);

  // Fetch upcoming fight for this boxer (MAT-106)
  useEffect(() => {
    const fetchUpcomingFightData = async () => {
      // If fight data is already provided via prop, use it
      if (upcomingFight) {
        setFetchedFight(upcomingFight);
        return;
      }

      try {
        const fightData = await fetchUpcomingFight(boxer.id);
        setFetchedFight(fightData);
      } catch (err) {
        // Log for debugging - boxer card still works without fight data
        console.debug(`BoxerCard (${boxer.name}): Failed to load upcoming fight`, err);
      }
    };

    if (boxer.id) {
      fetchUpcomingFightData();
    }
  }, [boxer.id, upcomingFight, boxer.name]);

  // Poll for training session status when there's active training
  useEffect(() => {
    if (!activeSession) return;

    const pollInterval = setInterval(async () => {
      try {
        const sessions = await fetchActiveTrainingSessions(boxer.id);

        const pendingSession = sessions.find(s => s.status === 'pending');

        // If training completed (no longer pending), notify parent to refresh boxer data
        if (!pendingSession && activeSession) {
          setActiveSession(null);
          if (onBoxerStateChanged) {
            onBoxerStateChanged(boxer.id);
          }
        } else {
          setActiveSession(pendingSession);
        }
      } catch (err) {
        // Log for debugging - polling will retry
        console.debug(`BoxerCard (${boxer.name}): Training poll failed`, err.message);
      }
    }, 5000); // Check every 5 seconds

    return () => clearInterval(pollInterval);
  }, [activeSession, boxer.id, onBoxerStateChanged, boxer.name]);

  // Poll for rest status changes when boxer has active rest (MAT-96)
  useEffect(() => {
    if (!boxer.has_active_rest) return;

    const pollInterval = setInterval(async () => {
      try {
        const updatedBoxer = await fetchBoxer(boxer.id);

        // Check if rest status changed
        if (!updatedBoxer.has_active_rest && boxer.has_active_rest) {
          // Rest ended - notify parent to refresh all boxer data
          if (onBoxerStateChanged) {
            onBoxerStateChanged(boxer.id);
          }
        } else if (updatedBoxer.has_active_rest) {
          // Still resting - update local boxer state with new has_active_rest and next_available_training values
          const parentElement = document.getElementById(`boxer-card-${boxer.id}`);
          // Parent will refresh via Dashboard's world time polling for countdown updates
        }
      } catch (err) {
        // Log for debugging - polling will retry
        console.debug(`BoxerCard (${boxer.name}): Rest status poll failed`, err.message);
      }
    }, 5000); // Check every 5 seconds

    return () => clearInterval(pollInterval);
  }, [boxer.has_active_rest, boxer.id, onBoxerStateChanged, boxer.name]);

  // Calculate countdown based on scheduled_completion_time from API and world time
  useEffect(() => {
    if (!activeSession || !worldTime || !activeSession.scheduled_completion_time) {
      setCountdown('');
      return;
    }

    const calculateRemainingTime = () => {
      // Use the session's scheduled_completion_time (which is in game world time)
      const scheduledTime = new Date(activeSession.scheduled_completion_time);
      // Use current game world time, not real-time clock
      const nowGameTime = new Date(worldTime.current_game_time);

      // Calculate remaining seconds based on game time difference
      const remainingMs = scheduledTime - nowGameTime;
      const remainingSeconds = Math.max(0, Math.ceil(remainingMs / 1000));

      return formatCountdown(remainingSeconds);
    };

    setCountdown(calculateRemainingTime());

    // Update countdown when worldTime updates (which refreshes every 5 seconds)
    const countdownInterval = setInterval(() => {
      setCountdown(calculateRemainingTime());
    }, 1000);

    return () => clearInterval(countdownInterval);
  }, [activeSession, worldTime]);

  // Generate a consistent avatar based on boxer name (hash to emoji)
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

  // Get health color based on percentage
  const getHealthColorClass = (health, maxHealth = 100) => {
    const pct = (health / maxHealth) * 100;
    if (pct > 50) return 'bg-health-high';
    if (pct > 25) return 'bg-health-medium';
    return 'bg-health-low';
  };

  // Get fatigue color based on exhaustion threshold
  const getFatigueColorClass = (fatigueScore) => {
    if (fatigueScore >= 80) return 'bg-fatigue-critical';   // Red - exhausted, cannot train
    if (fatigueScore >= 50) return 'bg-fatigue-warning';    // Yellow - approaching exhaustion
    return 'bg-fatigue-ok';                                   // Green - healthy
  };

  // Get level badge styling
  const getLevelBadgeClass = (level) => {
    if (!level || level < 1) return '';
    if (level <= 10) return 'badge-level-1';
    if (level <= 25) return 'badge-level-2';
    if (level <= 50) return 'badge-level-3';
    if (level <= 75) return 'badge-level-4';
    return 'badge-level-5';
  };

  const avatar = getAvatarEmoji(boxer.name);
  const nicknameDisplay = boxer.nickname ? `"${boxer.nickname}"` : null;
  const healthBarClass = getHealthColorClass(boxer.health, 100);
  const levelBadgeClass = getLevelBadgeClass(boxer.level || 1);
  const maxHealth = 100;

  // Get training icon based on session type
  const trainingIcon = activeSession ? (trainingTypeConfig[getTrainingTypeName(activeSession.training_type_id)]?.icon || '🏋️') : null;

  // Button disabled state: health < 50, energy < 15, fatigue >= 80 (exhaustion), or already has pending training
  // Use !activeSession to handle both null and undefined (falsy values mean no active session)
  const isButtonDisabled = boxer.health < 50 || boxer.energy < 15 || boxer.fatigue_score >= 80 || !!activeSession;

  return (
    <div className="boxer-card" id={`boxer-card-${boxer.id}`}>
      {/* Level Badge */}
      <div className={`level-badge ${levelBadgeClass}`}>LVL {boxer.level || 1}</div>

      {/* Fight Badge - Shows next scheduled fight (MAT-106) */}
      {fetchedFight && (
        <FightBadge
          fight={fetchedFight}
          currentGameTime={worldTime?.current_game_time}
        />
      )}

      {/* Rest Status Badge - Shows when boxer has active rest period */}
      <RestingBadge
        hasActiveRest={boxer.has_active_rest}
        restEndsAt={boxer.next_available_training}
        forcedRestUntil={boxer.forced_rest_until}
        currentGameTime={worldTime?.current_game_time}
      />

      {/* Training Status Badge - Shows when boxer has pending training */}
      {activeSession && (
        <div className="training-badge">
          <span className="training-icon">{trainingIcon}</span>
          <span className="training-status">Training</span>
          {countdown && (
            <span className="training-timer" title="Time remaining">
              ⏱️ {countdown}
            </span>
          )}
        </div>
      )}

      {/* Avatar & Name Section - Clickable to view profile (MAT-148) */}
      <div className="card-header" onClick={onOpenProfile} title="Click to view profile">
        <div className="avatar-circle">{avatar}</div>
        <div className="name-section">
          <h3 className="boxer-name">{boxer.name}</h3>
          {nicknameDisplay && (
            <p className="boxer-nickname">{nicknameDisplay}</p>
          )}
        </div>
      </div>

      {/* Health Bar */}
      <ProgressBar
        value={boxer.health}
        max={maxHealth}
        variant="health"
        label="Health"
      />

      {/* Energy Bar */}
      <ProgressBar
        value={boxer.energy}
        max={100}
        variant="energy"
        label="Energy"
      />

      {/* Fatigue Bar */}
      <ProgressBar
        value={boxer.fatigue_score}
        max={100}
        variant="fatigue"
        label="Fatigue"
      />

      {/* XP Progress Bar */}
      <ProgressBar
        value={boxer.experience || 0}
        max={(boxer.level || 1) * 100}
        variant="xp"
        label="XP"
        showPercentage={false}
      />

      {/* Combat Stats */}
      <div className="combat-stats">
        <StatDisplay label="STR" value={boxer.strength} icon="💪" color="strength" />
        <StatDisplay label="AGI" value={boxer.agility} icon="⚡" color="agility" />
        <StatDisplay label="DEF" value={boxer.defense} icon="🛡️" color="defense" />
      </div>

      {/* Action Buttons */}
      <div className="card-actions">
        <button
          className="train-btn"
          onClick={onOpenTraining}
          disabled={isButtonDisabled}
          title={activeSession ? 'Boxer is currently training' : boxer.fatigue_score >= 80 ? `Boxer is exhausted (fatigue ${Math.round(boxer.fatigue_score)}/100). Rest required.` : boxer.health < 50 ? 'Boxer needs recovery (health < 50%)' : boxer.energy < 15 ? 'Insufficient energy (need 15+)' : 'Schedule Training'}
        >
          {activeSession ? 'Training In Progress...' : 'Schedule Training'}
        </button>

        <button
          className="fight-btn"
          onClick={onOpenFightBooking}
          disabled={boxer.health < 30 || boxer.has_active_rest}
          title={boxer.health < 30 ? 'Boxer too injured to fight (need 30+ health)' : boxer.has_active_rest ? 'Boxer is resting' : 'Book a Fight'}
        >
          🥊 Book Fight
        </button>

        <button
          className="history-btn"
          onClick={onOpenFightHistory}
          title="View Fight History"
        >
          📜 History
        </button>
      </div>

      {/* Position coordinates (optional - can be removed) */}
      <div className="card-footer">
        <small className="position-text">
          Pos: {boxer.position_x?.toFixed(0)}, {boxer.position_y?.toFixed(0)}
        </small>
      </div>
    </div>
  );
};

export default BoxerCard;
