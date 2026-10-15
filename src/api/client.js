/**
 * API Client - HTTP Transport Layer (MAT-120)
 *
 * Provides centralized configuration, authentication, and error handling
 * for all API requests in the application.
 */

/**
 * Base URL for API requests.
 * In development, webpack dev server proxies /api to backend.
 * In production, configure as needed.
 */
export const API_BASE_URL = '/api';

/**
 * Custom error class for API errors with structured data
 */
export class ApiError extends Error {
  constructor(message, status, statusText, body = null) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.statusText = statusText;
    this.body = body;
  }
}

/**
 * Get the stored JWT token from localStorage
 * @returns {string|null} The access token or null if not found
 */
export const getToken = () => {
  return localStorage.getItem('token');
};

/**
 * Store the JWT token and user data in localStorage
 * @param {string} token - The access token
 * @param {Object} user - User data
 */
export const setToken = (token, user) => {
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
};

/**
 * Clear the stored token and user data from localStorage
 */
export const clearToken = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
};

/**
 * Get the stored user from localStorage
 * @returns {Object|null} The user object or null if not found
 */
export const getUser = () => {
  const userData = localStorage.getItem('user');
  if (userData) {
    try {
      return JSON.parse(userData);
    } catch (e) {
      // Invalid JSON, clear it
      clearToken();
      return null;
    }
  }
  return null;
};

/**
 * Create fetch headers with authentication token
 * @param {Object} additionalHeaders - Additional headers to merge
 * @returns {HeadersInit} Headers object with Authorization header
 */
export const getAuthHeaders = (additionalHeaders = {}) => {
  const token = getToken();
  return {
    'Content-Type': 'application/json',
    ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    ...additionalHeaders,
  };
};

/**
 * Options for API client configuration
 */
export const defaultConfig = {
  // Whether to throw ApiError on HTTP errors (default: true)
  throwOnHttpError: true,

  // Maximum retry attempts for failed requests (default: 0, no retry)
  maxRetries: 0,

  // Initial delay between retries in milliseconds (default: 1000)
  retryDelayMs: 1000,

  // Whether to redirect to login on 401 (default: true)
  redirectOnUnauth: true,
};

/**
 * Make an API request with error handling and optional retries.
 *
 * @param {string} url - The URL path (e.g., '/boxers/123') or full URL
 * @param {RequestInit} options - Fetch options (method, body, etc.)
 * @param {Object} config - Optional configuration overrides
 * @returns {Promise<Response>} The fetch response (caller parses JSON)
 * @throws {ApiError} When HTTP error occurs and throwOnHttpError is true
 */
export async function apiRequest(url, options = {}, config = {}) {
  const {
    throwOnHttpError = defaultConfig.throwOnHttpError,
    maxRetries = defaultConfig.maxRetries,
    retryDelayMs = defaultConfig.retryDelayMs,
    redirectOnUnauth = defaultConfig.redirectOnUnauth,
  } = config;

  let lastError = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      // Add auth token to headers
      const token = getToken();
      const headers = {
        'Content-Type': 'application/json',
        ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
        ...options.headers, // Allow caller to override/add headers
      };

      const response = await fetch(url, {
        ...options,
        headers,
      });

      // Handle 401 Unauthorized - clear token and redirect
      if (response.status === 401 && redirectOnUnauth) {
        clearToken();
        window.location.href = '/login';
        throw new ApiError('Unauthorized: Please login again', 401, 'Unauthorized');
      }

      // Handle HTTP errors
      if (!response.ok) {
        // Try to parse error body for more details
        let body = null;
        try {
          body = await response.json();
        } catch (e) {
          // Response might not be JSON, use text
          body = { message: await response.text() || response.statusText };
        }

        if (throwOnHttpError) {
          throw new ApiError(
            body.message || `HTTP error ${response.status}`,
            response.status,
            response.statusText,
            body
          );
        }
      }

      return response;

    } catch (error) {
      // Don't retry on non-network errors (e.g., 401, 403, 404)
      if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
        throw error;
      }

      lastError = error;

      // Retry on network errors if configured
      if (attempt < maxRetries) {
        const delay = retryDelayMs * Math.pow(2, attempt); // Exponential backoff
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }

      // No more retries, throw last error
      throw error;
    }
  }

  // Should never reach here, but just in case
  throw lastError || new ApiError('Unknown error occurred', 0, 'Unknown');
}

/**
 * Make an API request and parse the response as JSON.
 * This is a convenience wrapper around apiRequest().
 *
 * @param {string} url - The URL path or full URL
 * @param {RequestInit} options - Fetch options
 * @param {Object} config - Optional configuration overrides
 * @returns {Promise<Object>} Parsed response body
 */
export async function apiGetJson(url, options = {}, config = {}) {
  const response = await apiRequest(url, { ...options, method: 'GET' }, config);

  if (!response.ok) {
    // This shouldn't happen if throwOnHttpError is true
    const body = await response.json().catch(() => ({}));
    throw new ApiError('Failed to parse response', response.status, response.statusText, body);
  }

  return response.json();
}

/**
 * Make a POST request with JSON body.
 *
 * @param {string} url - The URL path or full URL
 * @param {Object} body - Request body (will be JSON.stringify'd)
 * @param {RequestInit} options - Additional fetch options
 * @param {Object} config - Optional configuration overrides
 * @returns {Promise<Object>} Parsed response body
 */
export async function apiPostJson(url, body, options = {}, config = {}) {
  const response = await apiRequest(url, {
    method: 'POST',
    body: JSON.stringify(body),
    ...options,
  }, config);

  // Some POST endpoints return 204 No Content (no body)
  if (response.status === 204) {
    return null;
  }

  return response.json();
}

/**
 * Make a PUT request with JSON body.
 *
 * @param {string} url - The URL path or full URL
 * @param {Object} body - Request body (will be JSON.stringify'd)
 * @param {RequestInit} options - Additional fetch options
 * @param {Object} config - Optional configuration overrides
 * @returns {Promise<Object>} Parsed response body
 */
export async function apiPutJson(url, body, options = {}, config = {}) {
  const response = await apiRequest(url, {
    method: 'PUT',
    body: JSON.stringify(body),
    ...options,
  }, config);

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

/**
 * Make a DELETE request.
 *
 * @param {string} url - The URL path or full URL
 * @param {RequestInit} options - Additional fetch options
 * @param {Object} config - Optional configuration overrides
 * @returns {Promise<Object|null>} Parsed response body or null for 204
 */
export async function apiDelete(url, options = {}, config = {}) {
  const response = await apiRequest(url, {
    method: 'DELETE',
    ...options,
  }, config);

  if (response.status === 204) {
    return null;
  }

  return response.json();
}
