/**
 * Alert Component
 *
 * Reusable alert/message display for success, warning, error, and info states.
 * Replaces inline error/success messages with consistent UI.
 *
 * @see MAT-154 - Additional Refactoring Opportunities
 * @see MAT-129 - Error Handling & User Feedback
 *
 * @example
 * // Success message after action
 * <Alert variant="success" title="Success!" message="Training scheduled successfully" />
 *
 * @example
 * // Warning with dismissible option
 * <Alert
 *   variant="warning"
 *   title="Low Energy"
 *   message="Boxer has insufficient energy for training"
 *   dismissible
 *   onDismiss={handleDismiss}
 />

 * @example
 * // Error display
 * <Alert variant="error" title="Error" message={errorMessage} />
 */

import React from 'react';
import './Alert.css';

/**
 * Alert component props interface
 *
 * @typedef {Object} AlertProps
 * @property {'info' | 'success' | 'warning' | 'error'} [variant='info'] - Alert type
 * @property {string} [title] - Alert title/heading
 * @property {string} [message] - Alert message/body text
 * @property {boolean} [dismissible=false] - Show dismiss button
 * @property {function} [onDismiss] - Dismiss callback (required if dismissible=true)
 * @property {string} [className] - Additional CSS classes
 */

export function Alert({
  variant = 'info',
  title,
  message,
  dismissible = false,
  onDismiss,
  className = '',
}) {
  const alertClass = `alert alert-${variant} ${dismissible ? 'alert-dismissible' : ''} ${className}`.trim();

  // Icons for each variant
  const icons = {
    info: 'ℹ️',
    success: '✅',
    warning: '⚠️',
    error: '❌',
  };

  return (
    <div className={alertClass} role="alert">
      <span className="alert-icon" aria-hidden="true">{icons[variant]}</span>
      <div className="alert-content">
        {title && <strong className="alert-title">{title}</strong>}
        {message && <p className="alert-message">{message}</p>}
      </div>
      {dismissible && onDismiss && (
        <button
          className="alert-dismiss"
          onClick={onDismiss}
          aria-label="Dismiss alert"
        >
          ✕
        </button>
      )}
    </div>
  );
}

export default Alert;
