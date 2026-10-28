/**
 * StatDisplay Component
 *
 * Reusable stat display with support for compact, bar, and mini variants.
 * Used for combat stats (STR/DEF/AGI), levels, and other numeric displays.
 *
 * @see MAT-154 - Additional Refactoring Opportunities
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Compact stat display (BoxerCard style)
 * <StatDisplay label="STR" value={85} icon="💪" color="strength" />
 *
 * @example
 * // Bar variant with percentage (StatsOverview style)
 * <StatDisplay
 *   label="Strength"
 *   value={strength}
 *   icon="💪"
 *   variant="bar"
 *   color="strength"
 *   maxValue={200}
 />

 * @example
 * // Mini variant (RankCard style)
 * <StatDisplay label="LVL" value={boxer.level} variant="mini" />
 */

import React from 'react';
import './StatDisplay.css';

/**
 * StatDisplay component props interface
 *
 * @typedef {Object} StatDisplayProps
 * @property {string} label - Display label (e.g., "STR", "Strength", "LVL")
 * @property {number} value - Numeric value to display
 * @property {'compact' | 'bar' | 'mini'} [variant='compact'] - Display variant
 * @property {'strength' | 'defense' | 'agility' | 'level' | 'neutral'} [color='neutral'] - Color scheme
 * @property {string} [icon] - Emoji or icon character (optional)
 * @property {number} [maxValue] - Maximum value for bar variant percentage calculation
 * @property {string} [className] - Additional CSS classes
 */

export function StatDisplay({
  label,
  value,
  variant = 'compact',
  color = 'neutral',
  icon,
  maxValue,
  className = '',
}) {
  const statClass = `stat-display stat-${variant} stat-${color} ${className}`.trim();

  // Round value for display
  const displayValue = Math.round(value);

  // Calculate percentage for bar variant
  const percentage = maxValue ? Math.min(100, Math.max(0, (value / maxValue) * 100)) : 0;

  if (variant === 'bar') {
    return (
      <div className={statClass}>
        {icon && <span className="stat-icon">{icon}</span>}
        <span className="stat-label">{label}</span>
        <div className="stat-bar-container">
          <div className={`stat-bar-fill stat-bar-${color}`}>
            <div style={{ width: `${percentage}%` }}></div>
          </div>
        </div>
        <span className="stat-value">{displayValue}</span>
      </div>
    );
  }

  if (variant === 'mini') {
    return (
      <div className={statClass}>
        <span className="stat-label">{label}</span>
        <span className="stat-value">{displayValue}</span>
      </div>
    );
  }

  // Compact variant (default)
  return (
    <div className={statClass}>
      {icon && <span className="stat-icon" title={label}>{icon}</span>}
      <div className="stat-info">
        <span className="stat-name">{label}</span>
        <span className="stat-value-sm">{displayValue}</span>
      </div>
    </div>
  );
}

export default StatDisplay;
