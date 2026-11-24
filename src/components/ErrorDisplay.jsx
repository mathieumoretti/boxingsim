/**
 * ErrorDisplay Component (MAT-129)
 *
 * Reusable error display component with consistent styling and retry functionality.
 * Provides user-friendly error messages with appropriate visual feedback.
 */

import React from 'react';

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
    <div className="flex items-start gap-4 p-5 bg-gradient-to-br from-red-50 to-orange-50 border border-danger/30 rounded-lg shadow-[0_2px_16px_rgba(239,68,68,0.1),inset_0_1px_0_rgba(255,255,255,0.5)]">
      {icon && (
        <svg
          className="flex-shrink-0 w-6 h-6 text-danger"
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
      <div className="flex-1 min-w-0">
        <h4 className="text-base font-bold text-gray-900 m-0 mb-1">{title}</h4>
        <p className="text-sm text-gray-700 m-0 break-words">{error}</p>
        {onRetry && (
          <button
            className="mt-3 inline-flex items-center gap-2 px-5 py-2.5 bg-gradient-to-br from-danger to-danger-darker text-white border-none rounded-md text-sm font-semibold cursor-pointer transition-all duration-150 hover:bg-gradient-to-br hover:from-danger-brighter hover:to-danger hover:-translate-y-0.5 hover:shadow-[0_4px_12px_rgba(239,68,68,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-danger focus-visible:outline-offset-2"
            onClick={onRetry}
            type="button"
          >
            ⟳ Try Again
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
    <div className="flex items-center gap-3 p-3 bg-gradient-to-r from-red-50 to-orange-50 border border-danger/20 rounded-md">
      {icon && (
        <svg
          className="flex-shrink-0 w-4 h-4 text-danger"
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
      <span className="text-sm text-gray-700 truncate flex-1 min-w-0">{error}</span>
      {onRetry && (
        <button
          className="text-xs font-semibold text-danger cursor-pointer hover:underline focus-visible:outline-none"
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
    <div className="flex items-center gap-2 text-sm text-danger">
      {icon && (
        <svg
          className="flex-shrink-0 w-3.5 h-3.5"
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
      <span>{error}</span>
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
    <div className="flex items-center gap-1.5 text-xs text-danger mt-1.5">
      <svg
        className="flex-shrink-0 w-3 h-3"
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
