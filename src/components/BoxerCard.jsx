import React from 'react';

import './BoxerCard.css';

const BoxerCard = ({ boxer }) => {
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
  const nicknameDisplay = boxer.nickname ? `"\${boxer.nickname}"` : null;
  const healthBarClass = getHealthColorClass(boxer.health, 100);
  const levelBadgeClass = getLevelBadgeClass(boxer.level || 1);
  const maxHealth = 100;

  return (
    <div className="boxer-card">
      {/* Level Badge */}
      <div className={`level-badge ${levelBadgeClass}`}>LVL {boxer.level || 1}</div>

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
