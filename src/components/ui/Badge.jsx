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
import './Badge.css';

/**
 * Badge component props interface
 *
 * @typedef {Object} BadgeProps
 * @property {React.ReactNode} children - Badge content
 * @property {'default' | 'level' | 'level-high' | 'training' | 'fight-status-scheduled' | 'fight-status-completed' | 'fight-status-cancelled' | 'winner'} [variant='default'] - Badge style variant
 * @property {'sm' | 'md' | 'lg'} [size='md'] - Badge size
 * @property {'static' | 'absolute-tr'} [position='static'] - Positioning mode
 * @property {string} [icon] - Optional leading icon/emoji
 * @property {string} [className] - Additional CSS classes
 */

export function Badge({
  children,
  variant = 'default',
  size = 'md',
  position = 'static',
  icon,
  className = '',
}) {
  const badgeClass = `badge badge-${variant} badge-${size} ${position === 'absolute-tr' ? 'badge-absolute' : ''} ${className}`.trim();

  return (
    <span className={badgeClass}>
      {icon && <span className="badge-icon">{icon}</span>}
      {children}
    </span>
  );
}

export default Badge;
