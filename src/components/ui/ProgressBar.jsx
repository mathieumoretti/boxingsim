/**
 * ProgressBar Component
 *
 * Reusable progress bar with configurable value, max, and color variants.
 * Used for health, energy, fatigue, XP, and stat displays.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Health bar with percentage display
 * <ProgressBar value={85} max={100} variant="health" label="Health" />
 *
 * @example
 * // XP bar without percentage (absolute value)
 * <ProgressBar value={145} max={200} variant="xp" showPercentage={false} />
 *
 * @example
 * // Fatigue bar with custom max
 * <ProgressBar value={45} max={100} variant="fatigue" label="Fatigue" />
 */

import React from 'react';
import './ProgressBar.css';

/**
 * ProgressBar component props interface
 *
 * @typedef {Object} ProgressBarProps
 * @property {number} value - Current value (required)
 * @property {number} [max=100] - Maximum value for percentage calculation
 * @property {'health' | 'energy' | 'fatigue' | 'xp' | 'strength' | 'defense' | 'agility' | 'neutral'} [variant='neutral'] - Color variant
 * @property {string} [label] - Label text displayed with the bar
 * @property {boolean} [showPercentage=true] - Show percentage/value text next to bar
 * @property {string} [className] - Additional CSS classes
 * @property {number} [height] - Custom height in pixels (defaults to design token)
 */

export function ProgressBar({
  value,
  max = 100,
  variant = 'neutral',
  label,
  showPercentage = true,
  className = '',
  height,
}) {
  // Calculate percentage (clamped between 0 and 100)
  const percentage = Math.min(100, Math.max(0, (value / max) * 100));

  // Determine if we should show the percentage text or raw value
  const isPercentage = max === 100;
  const displayValue = isPercentage
    ? `${Math.round(percentage)}%`
    : `${value}/${max}`;

  const containerClass = `progress-container ${className}`.trim();
  const barClass = `progress-bar progress-${variant}`.trim();

  return (
    <div className={containerClass}>
      {label && (
        <div className="progress-label">
          <span className="progress-label-text">{label}</span>
          {showPercentage && (
            <span className="progress-label-value">{displayValue}</span>
          )}
        </div>
      )}

      <div
        className="progress-track"
        style={height ? { height: `${height}px` } : undefined}
      >
        <div
          className={barClass}
          style={{ width: `${percentage}%` }}
          role="progressbar"
          aria-valuenow={value}
          aria-valuemin={0}
          aria-valuemax={max}
          aria-label={`${label}: ${displayValue}`}
        />
      </div>
    </div>
  );
}

export default ProgressBar;
