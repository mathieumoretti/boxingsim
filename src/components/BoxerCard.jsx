import React, { useState, useEffect } from 'react';
import './BoxerCard.css';
import { API_BASE_URL, authenticatedFetch } from '../utils/auth';

const trainingTypeIcons = {
  strength_training: '💪',
  speed_training: '⚡',
  defense_training: '🛡️',
  technique_training: '🥊',
  endurance_training: '🏃',
};

const getTrainingTypeName = (typeId) => {
  const mapping = {
    1: 'strength_training',
    2: 'speed_training',
    3: 'defense_training',
    4: 'technique_training',
    5: 'endurance_training',
  };
  return mapping[typeId] || 'training';
};

const formatCountdown = (remainingSeconds) => {
  if (remainingSeconds <= 0) return 'Ready!';

  const hours = Math.floor(remainingSeconds / 3600);
  const minutes = Math.floor((remainingSeconds % 3600) / 60);
  const seconds = remainingSeconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds}s`;
  } else {
    return `${seconds}s`;
  }
};

const BoxerCard = ({ boxer, onOpenTraining }) => {
  const [activeSession, setActiveSession] = useState(null);
  const [countdown, setCountdown] = useState('');
  const [worldTime, setWorldTime] = useState(null);

  // Fetch active training session for this boxer
  useEffect(() => {
    const loadActiveTraining = async () => {
      try {
        const response = await authenticatedFetch(`${API_BASE_URL}/boxers/${boxer.id}/training-sessions`, {
          method: 'GET',
        });

        console.log('[BoxerCard] API response status:', response.status);

        if (response.ok) {
          const sessions = await response.json();
          console.log('[BoxerCard] Training sessions for boxer', boxer.name, ':', sessions);

          // Handle null/undefined - treat as empty array
          const sessionsArray = Array.isArray(sessions) ? sessions : [];
          console.log('[BoxerCard] Sessions array length:', sessionsArray.length);

          // Find the most recent pending session
          const pendingSession = sessionsArray.find(s => s.status === 'pending');

          console.log('[BoxerCard] Pending session found:', pendingSession);
          setActiveSession(pendingSession);
        } else {
          console.warn('[BoxerCard] Failed to load training sessions, status:', response.status);
        }
      } catch (err) {
        // Silently fail - boxer card still works without training status
        console.warn('[BoxerCard] Failed to load training session:', err);
      }
    };

    if (boxer.id) {
      loadActiveTraining();
    }
  }, [boxer.id, boxer.name]);

  // Fetch world time for countdown calculation
  useEffect(() => {
    const loadWorldTime = async () => {
      try {
        const response = await authenticatedFetch(`${API_BASE_URL}/world/time`, {
          method: 'GET',
        });

        if (response.ok) {
          const data = await response.json();
          setWorldTime(data);

          // Set up periodic refresh for countdown updates
          const interval = setInterval(async () => {
            try {
              const resp = await authenticatedFetch(`${API_BASE_URL}/world/time`, { method: 'GET' });
              if (resp.ok) {
                const data = await resp.json();
                setWorldTime(data);
              }
            } catch (err) {
              // Silently fail - countdown will just be stale
            }
          }, 5000);

          return () => clearInterval(interval);
        }
      } catch (err) {
        console.warn('Failed to load world time:', err);
      }
    };

    if (activeSession) {
      loadWorldTime();
    }
  }, [activeSession]);

  // Calculate countdown based on world clock time formula
  useEffect(() => {
    if (!activeSession || !worldTime) {
      setCountdown('');
      return;
    }

    const calculateRemainingTime = () => {
      // Use the same formula as WorldClock component
      // Game time advances based on seconds_per_game_hour
      const now = new Date();
      const realAnchor = new Date(worldTime.real_anchor);
      const currentGameTime = new Date(worldTime.current_game_time);

      // Calculate elapsed real time since world clock started
      const elapsedRealSeconds = (now - realAnchor) / 1000;

      // The session has a duration in game hours
      // We need to estimate when the session will complete based on world clock speed

      // For now, use a simpler approach: show "In Progress" since we don't have exact completion timestamp
      // In production, this would be calculated from scheduled_event tables or session start time

      const secondsPerGameHour = worldTime.seconds_per_game_hour || 60;
      const remainingSeconds = Math.max(0, Math.ceil(activeSession.duration_hours * secondsPerGameHour - elapsedRealSeconds));

      return formatCountdown(remainingSeconds);
    };

    setCountdown(calculateRemainingTime());

    // Update countdown every second while session is active
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
  const trainingIcon = activeSession ? (trainingTypeIcons[getTrainingTypeName(activeSession.training_type_id)] || '🏋️') : null;

  // Debug: Log boxer stats for disabled button check
  console.log('[BoxerCard] Button disabled check for', boxer.name, '- health:', boxer.health, 'energy:', boxer.energy, 'activeSession:', activeSession);

  // Button disabled state: health < 50, energy < 15, or already has pending training
  // Use !activeSession to handle both null and undefined (falsy values mean no active session)
  const isButtonDisabled = boxer.health < 50 || boxer.energy < 15 || !!activeSession;

  console.log('[BoxerCard] isButtonDisabled:', isButtonDisabled, '(hasActiveSession:', !!activeSession, ')');

  return (
    <div className="boxer-card">
      {/* Level Badge */}
      <div className={`level-badge ${levelBadgeClass}`}>LVL {boxer.level || 1}</div>

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

      {/* Avatar & Name Section */}
      <div className="card-header">
        <div className="avatar-circle">{avatar}</div>
        <div className="name-section">
          <h3 className="boxer-name">{boxer.name}</h3>
          {nicknameDisplay && (
            <p className="boxer-nickname">{nicknameDisplay}</p>
          )}
        </div>
      </div>

      {/* Health Bar */}
      <div className="stat-bar-container">
        <span className="stat-label">Health</span>
        <div className="bar-wrapper">
          <div
            className={`progress-bar ${healthBarClass}`}
            style={{ width: `${(boxer.health / maxHealth) * 100}%` }}
          ></div>
        </div>
        <span className="stat-value">{Math.floor(boxer.health)}<small>/ {maxHealth}</small></span>
      </div>

      {/* Energy Bar */}
      <div className="stat-bar-container">
        <span className="stat-label">Energy</span>
        <div className="bar-wrapper bar-energy">
          <div
            className={`progress-bar bg-energy`}
            style={{ width: `${(boxer.energy / 100) * 100}%` }}
          ></div>
        </div>
        <span className="stat-value">{Math.floor(boxer.energy)}<small>/ 100</small></span>
      </div>

      {/* XP Progress Bar */}
      <div className="stat-bar-container stat-bar-xp">
        <span className="stat-label">XP</span>
        <div className="bar-wrapper bar-xp">
          <div
            className={`progress-bar bg-xp`}
            style={{ width: `${((boxer.experience || 0) / ((boxer.level || 1) * 100)) * 100}%` }}
          ></div>
        </div>
        <span className="stat-value">{Math.floor(boxer.experience || 0)}<small>/ {(boxer.level || 1) * 100}</small></span>
      </div>

      {/* Combat Stats */}
      <div className="combat-stats">
        <div className="stat-item stat-strength">
          <span className="stat-icon" title="Strength">💪</span>
          <div className="stat-info">
            <span className="stat-name">STR</span>
            <span className="stat-value-sm">{Math.round(boxer.strength)}</span>
          </div>
        </div>

        <div className="stat-item stat-agility">
          <span className="stat-icon" title="Agility">⚡</span>
          <div className="stat-info">
            <span className="stat-name">AGI</span>
            <span className="stat-value-sm">{Math.round(boxer.agility)}</span>
          </div>
        </div>

        <div className="stat-item stat-defense">
          <span className="stat-icon" title="Defense">🛡️</span>
          <div className="stat-info">
            <span className="stat-name">DEF</span>
            <span className="stat-value-sm">{Math.round(boxer.defense)}</span>
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="card-actions">
        <button
          className="train-btn"
          onClick={onOpenTraining}
          disabled={isButtonDisabled}
          title={activeSession ? 'Boxer is currently training' : boxer.health < 50 ? 'Boxer needs recovery (health < 50%)' : boxer.energy < 15 ? 'Insufficient energy (need 15+)' : 'Schedule Training'}
        >
          {activeSession ? 'Training In Progress...' : 'Schedule Training'}
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
