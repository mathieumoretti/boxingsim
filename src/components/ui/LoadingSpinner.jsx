/**
 * LoadingSpinner Component
 *
 * Reusable loading spinner with configurable size and optional label.
 * Based on FightDetails loading state animation.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 * @see MAT-104 - Fight Details Page Component (original animation source)
 *
 * @example
 * // Simple spinner
 * <LoadingSpinner size="md" />
 *
 * @example
 * // Spinner with loading text
 * <LoadingSpinner size="lg" label="Loading fight details..." />
 */

import React from 'react';
import './LoadingSpinner.css';

/**
 * LoadingSpinner component props interface
 *
 * @typedef {Object} LoadingSpinnerProps
 * @property {'sm' | 'md' | 'lg'} [size='md'] - Spinner size
 * @property {string} [label] - Optional loading text displayed below spinner
 * @property {string} [className] - Additional CSS classes
 */

export function LoadingSpinner({
  size = 'md',
  label,
  className = '',
}) {
  const containerClass = `loading-spinner loading-spinner-${size} ${className}`.trim();

  return (
    <div className={containerClass} role="status" aria-label={label || 'Loading'}>
      <div className="spinner" />
      {label && <p className="loading-label">{label}</p>}
    </div>
  );
}

export default LoadingSpinner;
