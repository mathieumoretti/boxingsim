/**
 * StatDisplay Component
 *
 * Flexible stat display component for showing boxing stats in various formats.
 * Supports compact, bar, and mini variants with multiple color themes.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Compact stat display
 * <StatDisplay variant="compact" label="Strength" value={75} />
 *
 * @example
 * // Bar variant with progress
 * <StatDisplay variant="bar" label="Defense" value={42} max={100} color="defense" />
 *
 * @example
 * // Mini badge style
 * <StatDisplay variant="mini" label="AGI" value={58} color="agility" />
 */

import React from 'react';

/**
 * StatDisplay component props interface
 *
 * @typedef {Object} StatDisplayProps
 * @property {'compact' | 'bar' | 'mini'} variant - Display style variant
 * @property {string} label - Label for the stat
 * @property {number} value - Current stat value
 * @property {number} [max=100] - Maximum value (for bar variant)
 * @property {'default' | 'strength' | 'defense' | 'agility'} [color='default'] - Color theme
 * @property {boolean} [showPercentage=false] - Show percentage instead of raw value
 * @property {string} [className] - Additional CSS classes
 */

// Color theme styles with gradients and text colors
const colorStyles = {
  default: {
    gradient: 'from-gray-400 to-gray-600',
    bgGradient: 'bg-gradient-to-br from-gray-100 to-gray-200',
    text: 'text-gray-700',
    label: 'text-gray-500',
    barBg: 'bg-gray-200',
    shadow: '',
  },
  strength: {
    gradient: 'from-strength to-red-600',
    bgGradient: 'bg-gradient-to-br from-red-50 to-red-100',
    text: 'text-strength',
    label: 'text-gray-600',
    barBg: 'bg-gray-200',
    shadow: 'shadow-[0_2px_8px_rgba(239,68,68,0.15)]',
  },
  defense: {
    gradient: 'from-defense to-blue-600',
    bgGradient: 'bg-gradient-to-br from-blue-50 to-blue-100',
    text: 'text-defense',
    label: 'text-gray-600',
    barBg: 'bg-gray-200',
    shadow: 'shadow-[0_2px_8px_rgba(59,130,246,0.15)]',
  },
  agility: {
    gradient: 'from-agility to-amber-600',
    bgGradient: 'bg-gradient-to-br from-amber-50 to-amber-100',
    text: 'text-agility-dark',
    label: 'text-gray-600',
    barBg: 'bg-gray-200',
    shadow: 'shadow-[0_2px_8px_rgba(234,179,8,0.15)]',
  },
};

export function StatDisplay({
  variant = 'compact',
  label,
  value,
  max = 100,
  color = 'default',
  showPercentage = false,
  className = '',
}) {
  // Get theme styles
  const theme = colorStyles[color] || colorStyles.default;

  // Calculate percentage
  const percentage = max > 0 ? Math.round((value / max) * 100) : 0;

  // Format value display
  const displayValue = showPercentage
    ? `${percentage}%`
    : value;

  // ===== COMPACT VARIANT =====
  if (variant === 'compact') {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <span className={`text-sm font-medium ${theme.label}`}>{label}</span>
        <span className={`font-semibold text-base ${theme.text} tabular-nums`}>
          {displayValue}
        </span>
      </div>
    );
  }

  // ===== BAR VARIANT =====
  if (variant === 'bar') {
    return (
      <div className={`w-full ${className}`}>
        <div className="flex items-center justify-between mb-1.5">
          <span className={`text-sm font-medium ${theme.label}`}>{label}</span>
          <span className={`text-xs font-semibold ${theme.text} tabular-nums`}>
            {displayValue}
            {max > 0 && !showPercentage && (
              <span className="text-gray-400 ml-1">/ {max}</span>
            )}
          </span>
        </div>

        <div className={`h-3 rounded-full overflow-hidden ${theme.barBg}`}>
          <div
            className={`h-full rounded-full bg-gradient-to-r ${theme.gradient} transition-all duration-300 ease-out`}
            style={{ width: `${percentage}%` }}
          />
        </div>
      </div>
    );
  }

  // ===== MINI VARIANT =====
  if (variant === 'mini') {
    return (
      <div className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md ${theme.bgGradient} ${theme.shadow}`}>
        <span className="text-xs font-semibold text-gray-600 uppercase tracking-wide">
          {label}
        </span>
        <span className={`text-sm font-bold ${theme.text} tabular-nums`}>
          {displayValue}
        </span>
      </div>
    );
  }

  return null;
}

export default StatDisplay;
