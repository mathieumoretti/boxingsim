/**
 * LoadingSpinner Component
 *
 * Reusable loading spinner/loader indicator.
 * Supports multiple sizes and colors for various loading states.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Default loading spinner
 * <LoadingSpinner />
 *
 * @example
 * // Large spinner with custom color
 * <LoadingSpinner size="lg" color="danger" />
 *
 * @example
 * // Spinner with label
 * <LoadingSpinner label="Saving..." size="md" />
 */

import React from 'react';

/**
 * LoadingSpinner component props interface
 *
 * @typedef {Object} LoadingSpinnerProps
 * @property {'sm' | 'md' | 'lg' | 'xl'} [size='md'] - Spinner size
 * @property {'primary' | 'secondary' | 'danger' | 'warning'} [color='primary'] - Spinner color theme
 * @property {string} [label] - Optional label text to display below spinner
 * @property {boolean} [fullScreen=false] - Full screen overlay mode
 * @property {string} [className] - Additional CSS classes
 */

// Size styles with dimensions and border widths
const sizeStyles = {
  sm: {
    width: 'w-6',
    height: 'h-6',
    borderWidth: 'border-2',
    label: 'text-sm',
  },
  md: {
    width: 'w-10',
    height: 'h-10',
    borderWidth: 'border-3',
    label: 'text-base',
  },
  lg: {
    width: 'w-14',
    height: 'h-14',
    borderWidth: 'border-4',
    label: 'text-lg',
  },
  xl: {
    width: 'w-20',
    height: 'h-20',
    borderWidth: 'border-5',
    label: 'text-xl',
  },
};

// Color theme styles with ring colors and opacity
const colorStyles = {
  primary: {
    ring: 'border-primary border-t-transparent',
    text: 'text-primary',
  },
  secondary: {
    ring: 'border-gray-400 border-t-transparent',
    text: 'text-gray-500',
  },
  danger: {
    ring: 'border-danger-bright border-t-transparent',
    text: 'text-danger',
  },
  warning: {
    ring: 'border-warning-bright border-t-transparent',
    text: 'text-warning',
  },
};

export function LoadingSpinner({
  size = 'md',
  color = 'primary',
  label,
  fullScreen = false,
  className = '',
}) {
  // Get size and color styles
  const sizes = sizeStyles[size] || sizeStyles.md;
  const colors = colorStyles[color] || colorStyles.primary;

  if (fullScreen) {
    return (
      <div className="fixed inset-0 bg-navy-dark/90 backdrop-blur-sm flex items-center justify-center z-[var(--z-modal)]">
        <div className={`flex flex-col items-center gap-4 ${className}`}>
          {/* Spinner */}
          <div
            className={`${sizes.width} ${sizes.height} rounded-full animate-spin ${colors.ring}`}
            role="status"
            aria-label="Loading"
          />

          {/* Label */}
          {label && (
            <span className={`${sizes.label} font-medium text-white`}>
              {label}
            </span>
          )}
        </div>
      </div>
    );
  }

  return (
    <div
      className={`flex flex-col items-center gap-3 ${className}`}
      role="status"
      aria-label={label || 'Loading'}
    >
      {/* Spinner */}
      <div
        className={`${sizes.width} ${sizes.height} rounded-full animate-spin ${colors.ring}`}
      />

      {/* Label */}
      {label && (
        <span className={`${sizes.label} font-medium ${colors.text}`}>
          {label}
        </span>
      )}
    </div>
  );
}

export default LoadingSpinner;
