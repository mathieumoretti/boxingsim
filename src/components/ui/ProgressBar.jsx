/**
 * ProgressBar Component
 *
 * Reusable progress bar with support for multiple stat variants.
 * Used for health, energy, fatigue, experience, strength, defense, and agility bars.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Health bar with current/max values
 * <ProgressBar variant="health" value={75} max={100} showValue />
 *
 * @example
 * // Experience bar without label
 * <ProgressBar variant="xp" value={80} max={100} compact />
 *
 * @example
 * // Stat training bar
 * <ProgressBar variant="strength" value={45} max={100} />
 */

import React from 'react';

/**
 * ProgressBar component props interface
 *
 * @typedef {Object} ProgressBarProps
 * @property {'health' | 'energy' | 'fatigue' | 'xp' | 'strength' | 'defense' | 'agility'} variant - Progress bar type
 * @property {number} value - Current value
 * @property {number} max - Maximum value
 * @property {string} [label] - Optional label text
 * @property {boolean} [showValue=false] - Show numeric value on bar
 * @property {boolean} [compact=false] - Compact mode (no labels, thinner)
 * @property {boolean} [animated=true] - Enable animations
 * @property {string} [className] - Additional CSS classes
 */

// Variant color styles with gradients and text colors
const variantStyles = {
  health: {
    bar: 'from-success to-emerald-600',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_10px_rgba(16,185,129,0.4)]',
    text: 'text-success-light',
  },
  energy: {
    bar: 'from-warning-bright to-amber-600',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_10px_rgba(251,191,36,0.4)]',
    text: 'text-warning-bright',
  },
  fatigue: {
    bar: 'from-danger-bright to-rose-600',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_10px_rgba(239,68,68,0.4)]',
    text: 'text-danger-bright',
  },
  xp: {
    bar: 'from-primary to-blue-600',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_8px_rgba(59,130,246,0.3)]',
    text: 'text-primary-light',
  },
  strength: {
    bar: 'from-strength to-red-700',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_10px_rgba(239,68,68,0.4)]',
    text: 'text-strength',
  },
  defense: {
    bar: 'from-defense to-blue-700',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_10px_rgba(59,130,246,0.4)]',
    text: 'text-defense',
  },
  agility: {
    bar: 'from-agility to-amber-700',
    bg: 'bg-white/15',
    shadow: 'shadow-[inset_0_1px_3px_rgba(0,0,0,0.15)]',
    glow: 'shadow-[0_0_10px_rgba(234,179,8,0.4)]',
    text: 'text-amber-400',
  },
};

// Height styles for different modes
const heightStyles = {
  compact: 'h-2',
  medium: 'h-5',
  large: 'h-8',
};

export function ProgressBar({
  variant = 'health',
  value,
  max = 100,
  label,
  showValue = false,
  compact = false,
  animated = true,
  className = '',
}) {
  // Calculate percentage safely
  const percentage = max > 0 ? Math.min(100, Math.max(0, (value / max) * 100)) : 0;

  // Determine height based on compact mode and whether showing value
  let heightClass = compact ? heightStyles.compact : (showValue ? heightStyles.large : heightStyles.medium);

  // Get variant styles
  const styles = variantStyles[variant] || variantStyles.health;

  // Animation transition for smooth filling
  const transitionClass = animated
    ? 'transition-all duration-300 ease-out'
    : '';

  // Glowing effect based on percentage (high values glow)
  const glowClass = percentage > 80 && animated
    ? styles.glow
    : '';

  return (
    <div className={`flex flex-col gap-1 ${className}`}>
      {/* Label */}
      {label && !compact && (
        <div className="flex items-center justify-between mb-1">
          <span className={`text-sm font-medium ${styles.text}`}>{label}</span>
          {!showValue && (
            <span className="text-xs text-white/50 font-mono">
              {value} / {max}
            </span>
          )}
        </div>
      )}

      {/* Compact label */}
      {label && compact && (
        <div className="flex items-center gap-2 mb-1">
          <span className="text-xs text-white/60 font-medium">{label}</span>
          <span className="text-xs text-white/50 font-mono">
            {Math.round(percentage)}%
          </span>
        </div>
      )}

      {/* Progress bar container */}
      <div
        className={`relative rounded-full overflow-hidden ${styles.bg} ${styles.shadow} ${transitionClass}`}
        style={{ height: compact ? '8px' : (showValue ? '32px' : '20px') }}
      >
        {/* Fill bar with gradient */}
        <div
          className={`absolute top-0 left-0 h-full rounded-full bg-gradient-to-r ${styles.bar} ${glowClass} transition-all duration-300 ease-out`}
          style={{ width: `${percentage}%` }}
        >
          {/* Shimmer effect overlay */}
          {animated && (
            <div
              className="absolute inset-0 w-full h-full bg-gradient-to-r from-transparent via-white/20 to-transparent"
              style={{
                backgroundSize: '200% 100%',
                animation: 'barShimmer 2s infinite linear',
                opacity: animated ? 0.5 : 0,
              }}
            />
          )}
        </div>

        {/* Value text overlay (only for large bars) */}
        {showValue && (
          <div className="absolute inset-0 flex items-center justify-center">
            <span className="text-white font-bold text-sm tabular-nums drop-shadow-md z-10">
              {value} / {max}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}

export default ProgressBar;
