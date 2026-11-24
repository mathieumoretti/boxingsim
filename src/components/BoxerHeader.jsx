/**
 * BoxerHeader - Avatar, name, level, and XP progress section (MAT-148)
 *
 * Displays the boxer's identity information with visual treatment.
 * Future enhancement: Replace emoji avatar with AI-generated boxer portrait.
 */

import React from 'react';
import { Badge } from './ui';

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
    <div className="bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light rounded-xl p-6 shadow-[0_4px_16px_rgba(31,40,51,0.2)] text-white mb-6">
      <div className="flex flex-col sm:flex-row gap-6 items-start">
        {/* Avatar Section */}
        <div className="flex-shrink-0">
          <div className="w-20 h-20 rounded-full bg-gradient-to-br from-white/20 to-white/10 border-2 border-white/30 flex items-center justify-center text-4xl shadow-[inset_0_2px_8px_rgba(255,255,255,0.1)]">
            {avatar}
          </div>
        </div>

        {/* Info Section */}
        <div className="flex-1 min-w-0">
          <h2 className="text-2xl sm:text-3xl font-bold m-0 truncate">{boxer.name}</h2>
          {nicknameDisplay && (
            <p className="text-lg text-white/70 italic m-0 mb-2 truncate">{nicknameDisplay}</p>
          )}
          {/* Migrated to Badge component (MAT-154) */}
          <Badge variant="level" size="lg">Level {level}</Badge>
        </div>
      </div>

      {/* XP Progress Section */}
      <div className="mt-6 pt-6 border-t border-white/10">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-semibold text-white/90">Experience Points</span>
          <span className="text-sm font-bold text-victory-gold">{Math.floor(experience)} / {xpRequired} XP</span>
        </div>
        <div className="h-3 bg-navy-dark/50 rounded-full overflow-hidden border border-white/10">
          <div
            className="h-full bg-gradient-to-r from-victory-gold to-warning transition-all duration-500"
            style={{ width: `${xpPercent}%` }}
          ></div>
        </div>
        <span className="text-xs text-white/60 mt-1 block">{Math.round(xpPercent)}% to next level</span>
      </div>
    </div>
  );
};

export default BoxerHeader;
