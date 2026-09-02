// API base URL - will be proxied by webpack dev server
export const API_BASE_URL = '/api';

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
 * Make an authenticated API request
 * Automatically handles 401 responses by clearing token and redirecting to login
 * @param {string} url - The URL to fetch (relative path or full URL)
 * @param {Object} options - Fetch options (method, body, etc.)
 * @returns {Promise<Response>} The fetch response
 */
export const authenticatedFetch = async (url, options = {}) => {
  const token = getToken();

  // Merge headers, ensuring Authorization is included
  const headers = {
    ...options.headers,
    ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
  };

  const response = await fetch(url, {
    ...options,
    headers,
  });

  // Handle 401 Unauthorized - clear token and redirect to login
  if (response.status === 401) {
    clearToken();
    window.location.href = '/login';
    throw new Error('Unauthorized: Please login again');
  }

  return response;
};
