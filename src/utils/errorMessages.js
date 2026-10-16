/**
 * Error Message Utilities (MAT-129)
 *
 * Provides user-friendly error messages with specific, actionable feedback.
 * Converts ApiError objects and HTTP status codes into meaningful messages.
 */

import { ApiError } from '../api/client';

/**
 * Get a user-friendly error message from an error object.
 *
 * @param {Error|null} error - The error object (preferably ApiError)
 * @param {string} context - Operation context: 'training', 'fightBooking', 'auth', 'boxerCreation', 'ranking', 'general'
 * @param {Object} options - Additional context for message generation
 * @returns {string} User-friendly error message
 */
export function getUserErrorMessage(error, context = 'general', options = {}) {
  if (!error) {
    return 'An unknown error occurred';
  }

  // Handle ApiError with structured data (from MAT-120 client)
  if (error instanceof ApiError) {
    // Backend-provided message is most specific
    if (error.body?.message) {
      return error.body.message;
    }

    // Operation-specific messages based on status code
    return getOperationErrorMessage(error, context, options);
  }

  // Handle network errors (not ApiError)
  return getNetworkErrorMessage(error);
}

/**
 * Get operation-specific error messages.
 */
function getOperationErrorMessage(error, context, options) {
  const { boxerName = 'the boxer', opponentName = 'the opponent' } = options;

  switch (context) {
    case 'training':
      return getTrainingErrorMessage(error, options);

    case 'fightBooking':
      return getFightBookingErrorMessage(error, options);

    case 'auth':
      return getAuthErrorMessage(error);

    case 'boxerCreation':
      return getBoxerCreationErrorMessage(error);

    case 'ranking':
      return getRankingErrorMessage(error);

    case 'fightHistory':
      return getFightHistoryErrorMessage(error);

    case 'opponents':
      return getOpponentsErrorMessage(error, options);

    case 'boxerDetails':
      return getBoxerDetailsErrorMessage(error);

    default:
      return getGeneralErrorMessage(error);
  }
}

/**
 * Training operation error messages.
 */
function getTrainingErrorMessage(error, { boxerName = 'the boxer' } = {}) {
  switch (error.status) {
    case 409:
      // Conflict - already training or unavailable
      return `${boxerName || 'The boxer'} is currently in a training session. Wait for it to complete before scheduling another.`;

    case 400:
      // Bad request - injured, invalid stats, etc.
      if (error.body?.message?.toLowerCase().includes('injur')) {
        return `${boxerName || 'The boxer'} is injured and cannot train. Wait for recovery to complete.`;
      }
      if (error.body?.message?.toLowerCase().includes('rest')) {
        return `${boxerName || 'The boxer'} needs rest before training again.`;
      }
      return `Unable to schedule training: ${error.body?.message || 'Please check the boxers status and try again.'}`;

    case 404:
      return `${boxerName || 'The boxer'} was not found.`;

    default:
      return getGeneralErrorMessage(error, 'training');
  }
}

/**
 * Fight booking error messages.
 */
function getFightBookingErrorMessage(error, { boxerName = 'the boxer', opponentName = 'the opponent' } = {}) {
  switch (error.status) {
    case 409:
      // Conflict - scheduling conflict or availability issue
      if (error.body?.message?.toLowerCase().includes('schedul')) {
        return `Scheduling conflict detected. ${boxerName || 'The boxer'} has another event at this time.`;
      }
      if (error.body?.message?.toLowerCase().includes('opponent')) {
        return `${opponentName || 'The opponent'} is not available for this fight.`;
      }
      return `Cannot schedule this fight due to a conflict. Try a different date or opponent.`;

    case 400:
      // Bad request - validation errors
      if (error.body?.message?.toLowerCase().includes('avail')) {
        return `${opponentName || 'The opponent'} is not available for fights at this time.`;
      }
      if (error.body?.message?.toLowerCase().includes('health') ||
          error.body?.message?.toLowerCase().includes('injur')) {
        return `${boxerName || 'The boxer'} or ${opponentName || 'the opponent'} needs recovery before fighting.`;
      }
      return `Fight booking failed: ${error.body?.message || 'Please check both boxers are available and try again.'}`;

    case 404:
      return `One of the boxers was not found. They may have been deleted.`;

    case 422:
      // Unprocessable entity - validation failed
      return error.body?.message || 'Fight validation failed. Please check your selection and try again.';

    default:
      return getGeneralErrorMessage(error, 'fight booking');
  }
}

/**
 * Authentication error messages (login/register).
 */
function getAuthErrorMessage(error) {
  switch (error.status) {
    case 401:
      // Invalid credentials
      if (error.body?.message) {
        return error.body.message;
      }
      return 'Invalid username or password. Please check your credentials and try again.';

    case 403:
      // Account locked, disabled, etc.
      if (error.body?.message?.toLowerCase().includes('lock')) {
        return 'Your account is temporarily locked due to too many failed attempts. Please try again later.';
      }
      return 'Access denied. Your account may be restricted or disabled.';

    case 409:
      // Conflict - user exists (registration)
      if (error.body?.message?.toLowerCase().includes('exist') ||
          error.body?.message?.toLowerCase().includes('taken')) {
        return 'This username or email is already registered. Try logging in or use a different name.';
      }
      return 'Registration conflict: ' + (error.body?.message || 'Please try a different username.');

    case 422:
      // Validation error - weak password, etc.
      if (error.body?.message?.toLowerCase().includes('password')) {
        return error.body.message || 'Password does not meet requirements. Use at least 6 characters with letters and numbers.';
      }
      return `Invalid input: ${error.body?.message || 'Please check all fields are filled correctly.'}`;

    default:
      return getGeneralErrorMessage(error, 'authentication');
  }
}

/**
 * Boxer creation error messages.
 */
function getBoxerCreationErrorMessage(error) {
  switch (error.status) {
    case 409:
      // Conflict - name already exists
      if (error.body?.message?.toLowerCase().includes('name') ||
          error.body?.message?.toLowerCase().includes('exist')) {
        return 'A boxer with this name already exists. Choose a different name.';
      }
      return 'Boxer creation conflict: ' + (error.body?.message || 'Please try again.');

    case 400:
      // Bad request - invalid stats, etc.
      if (error.body?.message?.toLowerCase().includes('stat') ||
          error.body?.message?.toLowerCase().includes('invalid')) {
        return 'Invalid boxer stats. All values should be between 1 and 100.';
      }
      return error.body?.message || 'Unable to create boxer. Please check all fields are valid.';

    case 422:
      // Validation failed
      return error.body?.message || 'Boxer creation validation failed. Please check your input.';

    default:
      return getGeneralErrorMessage(error, 'boxer creation');
  }
}

/**
 * Rankings error messages.
 */
function getRankingErrorMessage(error) {
  switch (error.status) {
    case 404:
      return 'No rankings data found. There may not be any boxers in the system yet.';

    case 500:
      return 'Failed to calculate rankings. Please try refreshing the page.';

    default:
      return getGeneralErrorMessage(error, 'loading rankings');
  }
}

/**
 * Fight history error messages.
 */
function getFightHistoryErrorMessage(error) {
  switch (error.status) {
    case 404:
      return 'No fight history found for this boxer. They may not have fought yet.';

    default:
      return getGeneralErrorMessage(error, 'loading fight history');
  }
}

/**
 * Opponents list error messages.
 */
function getOpponentsErrorMessage(error, { boxerName = 'the boxer' } = {}) {
  switch (error.status) {
    case 404:
      return `No suitable opponents found for ${boxerName || 'this boxer'}. The boxer may be too new or have unique characteristics.`;

    default:
      return getGeneralErrorMessage(error, 'finding opponents');
  }
}

/**
 * Boxer details error messages.
 */
function getBoxerDetailsErrorMessage(error) {
  switch (error.status) {
    case 404:
      return 'Boxer not found. They may have been deleted or the ID is invalid.';

    default:
      return getGeneralErrorMessage(error, 'loading boxer details');
  }
}

/**
 * General error messages for any operation.
 */
function getGeneralErrorMessage(error, operation = '') {
  const action = operation ? ` ${operation}` : '';

  switch (error.status) {
    case 400:
      return error.body?.message || `Invalid request${action}. Please check your input and try again.`;

    case 401:
      // Should redirect automatically, but if shown:
      return 'Your session has expired. Please log in again.';

    case 403:
      return 'You do not have permission to perform this action.';

    case 404:
      return error.body?.message || `Resource not found${action}.`;

    case 409:
      return error.body?.message || `Conflict${action}. Please try again.`;

    case 422:
      return error.body?.message || `Validation failed${action}. Please check your input.`;

    case 500:
      return 'An internal server error occurred. Please try again later.';

    case 502:
    case 503:
    case 504:
      return 'The server is temporarily unavailable. Please check your connection and try again.';

    default:
      // Status unknown or network error
      if (error.status === 0) {
        return 'Network connection lost. Check your internet connection and try again.';
      }
      return `HTTP error ${error.status}${operation ? ` while ${operation}` : ''}. Please try again.`;
  }
}

/**
 * Get network-specific error messages.
 */
function getNetworkErrorMessage(error) {
  const message = error.message?.toLowerCase() || '';

  if (message.includes('fetch') || message.includes('network')) {
    return 'Network connection lost. Check your internet connection and try again.';
  }

  if (message.includes('abort')) {
    return 'Request was cancelled.';
  }

  if (message.includes('timeout')) {
    return 'Request timed out. The server is taking too long to respond. Try again.';
  }

  // Generic network error
  return `Network error: ${error.message || 'Unable to connect. Please check your connection and try again.'}`;
}

/**
 * Check if an error indicates a network issue (not an HTTP error).
 */
export function isNetworkError(error) {
  if (!(error instanceof ApiError)) {
    return true; // Not an ApiError means it's likely a network error
  }

  return error.status === 0 ||
         message.includes('fetch') ||
         message.includes('network');
}

/**
 * Check if an error is a 404 (resource not found).
 */
export function isNotFoundError(error) {
  return error instanceof ApiError && error.status === 404;
}

/**
 * Check if an error indicates server issues (5xx).
 */
export function isServerError(error) {
  return error instanceof ApiError &&
         error.status >= 500 &&
         error.status < 600;
}

/**
 * Check if an error is a conflict (409).
 */
export function isConflictError(error) {
  return error instanceof ApiError && error.status === 409;
}

/**
 * Get a short error summary for compact UI spaces (icons, badges).
 */
export function getErrorSummary(error) {
  if (isNotFoundError(error)) {
    return 'Not Found';
  }

  if (isConflictError(error)) {
    return 'Conflict';
  }

  if (isServerError(error)) {
    return 'Server Error';
  }

  if (error instanceof ApiError) {
    return error.statusText || `Error ${error.status}`;
  }

  return 'Error';
}
