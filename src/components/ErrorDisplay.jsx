/**
 * ErrorDisplay Component (MAT-129)
 *
 * Reusable error display component with consistent styling and retry functionality.
 * Provides user-friendly error messages with appropriate visual feedback.
 */

import React from 'react';
import './ErrorDisplay.css';

/**
 * Props:
 * - error: Error object or string message
 * - onRetry: Optional callback function to retry the operation
 * - variant: 'alert' (default) | 'compact' | 'inline'
 * - icon: boolean to show/hide error icon (default: true)
 * - title: Optional custom title for alert variant
 */
export function ErrorDisplay({
  error,
  onRetry,
  variant = 'alert',
  icon = true,
  title = 'Error'
}) {
  // Handle different error formats
  const errorMessage = getErrorMessage(error);

  if (!errorMessage) {
    return null;
  }

  switch (variant) {
    case 'compact':
      return <CompactErrorDisplay error={errorMessage} onRetry={onRetry} icon={icon} />;

    case 'inline':
      return <InlineErrorDisplay error={errorMessage} icon={icon} />;

    case 'alert':
    default:
      return <AlertErrorDisplay error={errorMessage} onRetry={onRetry} title={title} icon={icon} />;
  }
}

/**
 * Alert variant - Full error box with icon and optional retry button.
 */
function AlertErrorDisplay({ error, onRetry, title, icon }) {
  return (
    <div className="error-display error-alert">
      {icon && (
        <svg
          className="error-icon"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="15" y1="9" x2="9" y2="15" />
          <line x1="9" y1="9" x2="15" y2="15" />
        </svg>
      )}
      <div className="error-content">
        <h4 className="error-title">{title}</h4>
        <p className="error-message">{error}</p>
        {onRetry && (
          <button
            className="error-retry-btn"
            onClick={onRetry}
            type="button"
          >
            Try Again
          </button>
        )}
      </div>
    </div>
  );
}

/**
 * Compact variant - One-line error with icon, for cards or inline spaces.
 */
function CompactErrorDisplay({ error, onRetry, icon }) {
  return (
    <div className="error-display error-compact">
      {icon && (
        <svg
          className="error-icon"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="15" y1="9" x2="9" y2="15" />
          <line x1="9" y1="9" x2="15" y2="15" />
        </svg>
      )}
      <span className="error-message">{error}</span>
      {onRetry && (
        <button
          className="error-retry-btn error-retry-text"
          onClick={onRetry}
          type="button"
        >
          Retry
        </button>
      )}
    </div>
  );
}

/**
 * Inline variant - Simple text error, no box styling.
 */
function InlineErrorDisplay({ error, icon }) {
  return (
    <div className="error-display error-inline">
      {icon && (
        <svg
          className="error-icon"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <circle cx="12" cy="12" r="10" />
          <line x1="15" y1="9" x2="9" y2="15" />
          <line x1="9" y1="9" x2="15" y2="15" />
        </svg>
      )}
      <span className="error-message">{error}</span>
    </div>
  );
}

/**
 * Extract error message from various error formats.
 */
function getErrorMessage(error) {
  if (!error) {
    return null;
  }

  // String error
  if (typeof error === 'string') {
    return error;
  }

  // Object with message property
  if (typeof error === 'object' && error.message) {
    return error.message;
  }

  // Error object itself
  if (error instanceof Error) {
    return error.message;
  }

  return null;
}

/**
 * FieldError - Individual field validation error display.
 */
export function FieldError({ message }) {
  if (!message) {
    return null;
  }

  return (
    <div className="field-error">
      <svg
        className="field-error-icon"
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
      >
        <circle cx="12" cy="12" r="10" />
        <line x1="15" y1="9" x2="9" y2="15" />
        <line x1="9" y1="9" x2="15" y2="15" />
      </svg>
      <span>{message}</span>
    </div>
  );
}

export default ErrorDisplay;
