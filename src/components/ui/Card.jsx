/**
 * Card Component
 *
 * Reusable card container with header, body, and footer sections.
 * Supports health rows, combat stats patterns, and interactive states.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Full card with all sections
 * <Card>
 *   <CardHeader title="Boxer Profile" icon="🥊">
 *     <Badge variant="level">LVL 25</Badge>
 *   </CardHeader>
 *   <CardBody>...</CardBody>
 *   <CardFooter>
 *     <Button variant="primary">Edit</Button>
 *   </CardFooter>
 * </Card>
 */

import React from 'react';

// ===== CARD CONTAINER =====

/**
 * Card container component props
 *
 * @typedef {Object} CardProps
 * @property {React.ReactNode} children - Card content
 * @property {'elevated' | 'minimal'} [variant='elevated'] - Card visual style
 * @property {boolean} [interactive=false] - Enable hover interaction effects
 * @property {string} [className] - Additional CSS classes
 */

const variantStyles = {
  elevated: 'bg-gradient-to-br from-white to-gray-50 shadow-card border border-navy-border hover:-translate-y-0.5 hover:shadow-elevated',
  minimal: 'bg-white shadow-subtle border border-gray-200',
};

export function Card({ variant = 'elevated', interactive = false, children, className = '' }) {
  const interactiveStyles = interactive
    ? 'transition-all duration-250 cursor-pointer'
    : '';

  return (
    <div
      className={`
        rounded-lg overflow-hidden
        ${variantStyles[variant]}
        ${interactiveStyles}
        ${className}
      `.trim()}
    >
      {children}
    </div>
  );
}

// ===== CARD HEADER =====

/**
 * CardHeader component props
 *
 * @typedef {Object} CardHeaderProps
 * @property {string} title - Header title text
 * @property {React.ReactNode} [icon] - Optional icon/emoji before title
 * @property {React.ReactNode} [actions] - Optional action buttons/badges in header
 * @property {'default' | 'strength' | 'defense' | 'agility'} [statColor='default'] - Stat-specific color theming
 * @property {'centered' | 'left-aligned'} [align='centered'] - Text alignment mode
 * @property {React.ReactNode} children - Additional content (badges, etc.)
 * @property {string} [className] - Additional CSS classes
 */

const statColorStyles = {
  default: 'from-navy-dark to-navy-medium text-white',
  strength: 'from-strength to-strength-dark text-white shadow-[0_4px_12px_rgba(239,68,68,0.15)] border-b-4 border-double border-opacity-40',
  defense: 'from-defense to-defense-dark text-white shadow-[0_4px_12px_rgba(59,130,246,0.15)] border-b-4 border-double border-opacity-40',
  agility: 'from-agility to-agility-dark text-white shadow-[0_4px_12px_rgba(234,179,8,0.15)] border-b-4 border-double border-opacity-40',
};

export function CardHeader({
  title,
  icon,
  actions,
  statColor = 'default',
  align = 'centered',
  children,
  className = ''
}) {
  const alignmentStyles = align === 'centered' ? 'text-center' : 'text-left';

  return (
    <div
      className={`
        px-6 py-5 mb-4 rounded-t-lg
        bg-gradient-to-br ${statColorStyles[statColor]}
        flex items-center gap-3
        ${alignmentStyles} justify-center
        ${className}
      `.trim()}
    >
      {icon && (
        <span className="text-xl sm:text-2xl select-none shrink-0">
          {icon}
        </span>
      )}

      {title && (
        <h3 className="font-bold text-base sm:text-lg leading-tight whitespace-nowrap">
          {title}
        </h3>
      )}

      {actions && (
        <div className="ml-auto flex items-center gap-2 shrink-0">
          {actions}
        </div>
      )}

      {children && (
        <div className={`${align === 'centered' ? '' : 'ml-auto'} flex items-center gap-2`}>
          {children}
        </div>
      )}
    </div>
  );
}

// ===== CARD BODY =====

/**
 * CardBody component props
 *
 * @typedef {Object} CardBodyProps
 * @property {boolean} [padded=true] - Enable padding
 * @property {'default' | 'minimal'} [variant='default'] - Content styling
 * @property {React.ReactNode} children - Body content
 * @property {string} [className] - Additional CSS classes
 */

const bodyVariantStyles = {
  default: 'bg-white text-gray-900',
  minimal: 'bg-navy-medium text-gray-100',
};

export function CardBody({ padded = true, variant = 'default', children, className = '' }) {
  const paddingStyles = padded ? 'p-5' : '';

  return (
    <div
      className={`
        transition-colors duration-200
        ${bodyVariantStyles[variant]}
        ${paddingStyles}
        ${className}
      `.trim()}
    >
      {children}
    </div>
  );
}

// ===== CARD FOOTER =====

/**
 * CardFooter component props
 *
 * @typedef {Object} CardFooterProps
 * @property {'default' | 'minimal'} [variant='default'] - Footer styling
 * @property {React.ReactNode} children - Footer content (buttons, links)
 * @property {string} [className] - Additional CSS classes
 */

const footerVariantStyles = {
  default: 'bg-gray-50 border-t border-gray-200 text-gray-800',
  minimal: 'bg-navy-dark border-t border-navy-border text-white',
};

export function CardFooter({ variant = 'default', children, className = '' }) {
  return (
    <div
      className={`
        px-6 py-4 flex items-center justify-between gap-3
        transition-colors duration-200
        ${footerVariantStyles[variant]}
        ${className}
      `.trim()}
    >
      {children}
    </div>
  );
}

// ===== HEALTH ROW PATTERN =====

/**
 * HealthRow component for displaying health status with visual indicators
 *
 * @typedef {Object} HealthRowProps
 * @property {string} label - Status label
 * @property {number} [value] - Current numeric value
 * @property {number} [max=100] - Maximum value
 * @property {'excellent' | 'good' | 'fair' | 'poor' | 'critical'} [status] - Health status level
 * @property {boolean} [animate=true] - Enable pulse animation
 * @property {React.ReactNode} [icon] - Optional leading icon
 */

const healthStatusStyles = {
  excellent: 'from-success to-emerald-600 text-white',
  good: 'from-success-dark to-success text-white',
  fair: 'from-warning-bright to-warning text-slate-800',
  poor: 'from-danger-bright to-danger text-white',
  critical: 'bg-gradient-to-r from-danger to-danger-dark animate-pulse text-white',
};

const healthStatusPulse = {
  excellent: 'shadow-[0_2px_8px_rgba(16,185,129,0.3)]',
  good: '',
  fair: 'shadow-[0_2px_8px_rgba(251,191,36,0.4)]',
  poor: 'shadow-[0_2px_8px_rgba(239,68,68,0.3)]',
  critical: 'shadow-[0_2px_12px_rgba(220,38,38,0.5)] animate-pulse',
};

export function HealthRow({
  label,
  value,
  max = 100,
  status = 'excellent',
  animate = true,
  icon
}) {
  const percentage = max > 0 ? (value / max) * 100 : 0;
  const pulseClass = animate ? healthStatusPulse[status] : '';

  return (
    <div className={`flex items-center gap-3 rounded-md p-2 mb-2 bg-gradient-to-r ${healthStatusStyles[status]} ${pulseClass} transition-all duration-300`}>
      {icon && (
        <span className="text-lg shrink-0">{icon}</span>
      )}

      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-1">
          <span className="font-semibold text-sm whitespace-nowrap">{label}</span>
          {max > 0 && (
            <span className="text-xs font-medium opacity-90">
              {value} / {max}
            </span>
          )}
        </div>

        {max > 0 && (
          <div className="w-full h-1.5 bg-black/20 rounded-full overflow-hidden">
            <div
              className="h-full bg-white/90 rounded-full transition-all duration-300"
              style={{ width: `${percentage}%` }}
            />
          </div>
        )}
      </div>
    </div>
  );
}

// ===== COMBAT STAT PATTERN =====

/**
 * CombatStatPattern component for displaying strength/defense/agility stats
 *
 * @typedef {Object} CombatStatProps
 * @property {'strength' | 'defense' | 'agility'} stat - Stat type
 * @property {number} value - Current stat value
 * @property {string} [label] - Custom label (defaults to stat name)
 */

const combatStatStyles = {
  strength: 'bg-gradient-to-r from-strength to-strength-dark text-white shadow-[0_2px_8px_rgba(239,68,68,0.4)]',
  defense: 'bg-gradient-to-r from-defense to-defense-dark text-white shadow-[0_2px_8px_rgba(59,130,246,0.4)]',
  agility: 'bg-gradient-to-r from-agility to-agility-dark text-slate-800 shadow-[0_2px_8px_rgba(234,179,8,0.4)]',
};

export function CombatStat({ stat, value, label }) {
  const statNames = { strength: 'Strength', defense: 'Defense', agility: 'Agility' };
  const displayLabel = label || statNames[stat];

  return (
    <div className={`p-4 rounded-lg flex items-center justify-between ${combatStatStyles[stat]}`}>
      <span className="font-semibold">{displayLabel}</span>
      <span className="text-xl font-bold tabular-nums">{value}</span>
    </div>
  );
}

export default Card;
