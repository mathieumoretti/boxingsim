/**
 * Badge Component
 *
 * Reusable status badge with support for multiple variants and positions.
 * Used for levels, training status, fight status, and other UI indicators.
 *
 * @see MAT-154 - Additional Refactoring Opportunities
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Level badge (compact)
 * <Badge variant="level" size="sm">LVL 15</Badge>
 *
 * @example
 * // Training status with icon
 * <Badge variant="training" icon="🏋️">Training</Badge>
 *
 * @example
 * // Fight status badge
 * <Badge variant="fight-status-scheduled" icon="📅">Scheduled</Badge>
 * <Badge variant="winner" icon="🏆">Victory</Badge>
 */

import React from 'react';

/**
 * Badge component props interface
 *
 * @typedef {Object} BadgeProps
 * @property {React.ReactNode} children - Badge content
 * @property {'default' | 'level' | 'level-high' | 'training' | 'fight-status-scheduled' | 'fight-status-completed' | 'fight-status-cancelled' | 'winner'} [variant='default'] - Badge style variant
 * @property {'sm' | 'md' | 'lg'} [size='md'] - Badge size
 * @property {'static' | 'absolute-tr' | 'absolute-tl'} [position='static'] - Positioning mode
 * @property {string} [icon] - Optional leading icon/emoji
 * @property {string} [className] - Additional CSS classes
 */

// Variant styles with Tailwind utilities
const variantStyles = {
  default: 'bg-gradient-to-br from-gray-400 to-gray-500 text-white',
  level: 'bg-gradient-to-br from-warning-bright to-orange-500 text-white',
  'level-high': 'bg-gradient-to-br from-warning-bright to-warning text-white shadow-[0_2px_8px_rgba(251,191,36,0.4)]',
  training: 'bg-gradient-to-br from-success to-emerald-600 text-white shadow-[0_2px_4px_rgba(16,185,129,0.3)]',
  'fight-status-scheduled': 'bg-gradient-to-br from-accent-purple to-accent-purple-dark text-white',
  'fight-status-completed': 'bg-gradient-to-br from-success to-emerald-600 text-white',
  'fight-status-cancelled': 'bg-gradient-to-br from-danger to-danger-dark text-white',
  winner: 'bg-gradient-to-br from-victory-light to-warning text-slate-800 shadow-[0_2px_8px_rgba(252,211,74,0.4)]',
};

// Size styles with Tailwind utilities
const sizeStyles = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2 py-1 text-xs',
  lg: 'px-3 py-2 text-sm font-bold',
};

// Position styles with Tailwind utilities
const positionStyles = {
  static: '',
  'absolute-tr': 'absolute top-2 right-2 z-[var(--z-dropdown)]',
  'absolute-tl': 'absolute top-2 left-2 z-[var(--z-dropdown)]',
};

export function Badge({
  children,
  variant = 'default',
  size = 'md',
  position = 'static',
  icon,
  className = '',
}) {
  return (
    <span
      className={`
        inline-flex items-center justify-center gap-1
        rounded-sm leading-none whitespace-nowrap
        ${variantStyles[variant]}
        ${sizeStyles[size]}
        ${positionStyles[position]}
        ${className}
      `.trim()}
    >
      {icon && <span className="text-sm leading-none">{icon}</span>}
      {children}
    </span>
  );
}

export default Badge;
