/**
 * Alert Component
 *
 * Reusable alert/notification component with multiple variants.
 * Supports info, success, warning, and error states.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Success alert with title
 * <Alert variant="success" title="Success!">
 *   Boxer created successfully
 * </Alert>
 *
 * @example
 * // Error alert without title
 * <Alert variant="error">
 *   Something went wrong
 * </Alert>
 *
 * @example
 * // Warning with dismiss button
 * <Alert variant="warning" onDismiss={handleDismiss}>
 *   Training session will end soon
 * </Alert>
 */

import React from 'react';

/**
 * Alert component props interface
 *
 * @typedef {Object} AlertProps
 * @property {'info' | 'success' | 'warning' | 'error'} [variant='info'] - Alert variant
 * @property {string} [title] - Optional alert title
 * @property {React.ReactNode} children - Alert content
 * @property {function} [onDismiss] - Dismiss callback
 * @property {boolean} [dismissible=false] - Show dismiss button
 * @property {boolean} [animated=true] - Enable entrance animation
 * @property {string} [className] - Additional CSS classes
 */

// Variant icon mapping
const variantIcons = {
  info: 'ℹ️',
  success: '✅',
  warning: '⚠️',
  error: '❌',
};

// Variant color styles with gradients and borders
const variantStyles = {
  info: {
    background: 'bg-gradient-to-br from-blue-50 to-blue-100 border-t-2 border-primary/60',
    titleText: 'text-blue-900',
    text: 'text-blue-800',
    iconBg: 'bg-blue-100/50',
    shadow: '',
  },
  success: {
    background: 'bg-gradient-to-br from-success-light to-emerald-50 border-t-2 border-success/60',
    titleText: 'text-success-darker',
    text: 'text-success-dark',
    iconBg: 'bg-success/10',
    shadow: '',
  },
  warning: {
    background: 'bg-gradient-to-br from-warning-bright/20 to-amber-50 border-t-2 border-warning/60',
    titleText: 'text-amber-900',
    text: 'text-amber-800',
    iconBg: 'bg-warning/10',
    shadow: '',
  },
  error: {
    background: 'bg-gradient-to-br from-danger-bright/20 to-red-50 border-t-2 border-danger/60 shadow-[0_2px_8px_rgba(239,68,68,0.15)]',
    titleText: 'text-danger-darker',
    text: 'text-danger-dark',
    iconBg: 'bg-danger/10',
    shadow: '',
  },
};

// Icon container styles based on variant
const iconContainerStyles = {
  info: 'bg-blue-100 text-blue-700',
  success: 'bg-success-light text-success-dark',
  warning: 'bg-warning-bright/20 text-warning-dark',
  error: 'bg-danger-bright/20 text-danger',
};

export function Alert({
  variant = 'info',
  title,
  children,
  onDismiss,
  dismissible = false,
  animated = true,
  className = '',
}) {
  // Get variant styles
  const styles = variantStyles[variant] || variantStyles.info;
  const icon = variantIcons[variant];

  // Entrance animation classes
  const animationClasses = animated
    ? 'animate-[fadeIn_0.3s_ease-out]'
    : '';

  return (
    <div
      className={`
        relative rounded-lg p-4 mb-4
        ${styles.background}
        ${styles.border}
        ${styles.shadow}
        ${animationClasses}
        ${className}
      `.trim()}
      role="alert"
    >
      {/* Icon */}
      <div className={`inline-flex items-center justify-center w-8 h-8 rounded-full mr-3 shrink-0 ${iconContainerStyles[variant]}`}>
        <span className="text-sm">{icon}</span>
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        {title && (
          <h4 className={`font-semibold mb-1 ${styles.titleText}`}>
            {title}
          </h4>
        )}

        <div className={`text-sm ${styles.text}`}>
          {children}
        </div>
      </div>

      {/* Dismiss button */}
      {dismissible && onDismiss && (
        <button
          onClick={onDismiss}
          type="button"
          className="absolute top-2 right-2 p-1 rounded-md hover:bg-black/10 transition-colors duration-150 text-gray-500 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-1"
          aria-label="Dismiss alert"
        >
          <span className="text-lg leading-none">✕</span>
        </button>
      )}
    </div>
  );
}

export default Alert;
