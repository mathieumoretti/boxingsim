/**
 * BoxerHeader - Avatar, name, level, and XP progress section (MAT-148)
 *
 * Displays the boxer's identity information with visual treatment.
 * Future enhancement: Replace emoji avatar with AI-generated boxer portrait.
 */

import React from 'react';
import { Badge } from './ui';
import './BoxerProfile.css';

const BoxerHeader = ({ boxer }) => {
  if (!boxer) return null;

  // Generate consistent avatar based on boxer name (hash to emoji)
  // TODO: Replace with AI-generated boxer portrait (Phase 3 enhancement)
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

  const avatar = getAvatarEmoji(boxer.name);
  const nicknameDisplay = boxer.nickname ? `"${boxer.nickname}"` : null;

  // Calculate XP percentage
  const level = boxer.level || 1;
  const experience = boxer.experience || 0;
  const xpRequired = level * 100;
  const xpPercent = Math.min(100, (experience / xpRequired) * 100);

  return (
    <div className="boxer-header">
      <div className="header-avatar-section">
        <div className="avatar-large">{avatar}</div>
        <div className="header-info">
          <h2 className="boxer-name-heading">{boxer.name}</h2>
          {nicknameDisplay && (
            <p className="boxer-nickname-display">{nicknameDisplay}</p>
          )}
          {/* Migrated to Badge component (MAT-154) */}
          <Badge variant="level" size="lg">Level {level}</Badge>
        </div>
      </div>

      <div className="xp-progress-section">
        <div className="xp-labels">
          <span className="xp-text">Experience Points</span>
          <span className="xp-values">{Math.floor(experience)} / {xpRequired} XP</span>
        </div>
        <div className="xp-bar-wrapper">
          <div
            className="xp-progress-bar"
            style={{ width: `${xpPercent}%` }}
          ></div>
        </div>
        <span className="xp-percent">{Math.round(xpPercent)}% to next level</span>
      </div>
    </div>
  );
};

export default BoxerHeader;
