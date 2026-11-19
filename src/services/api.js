/**
 * API client for Boxing Sim backend
 * Centralized HTTP client with error handling and auth support
 */

const API_BASE_URL = '/'; // Backend serves at root, React dev server proxies /api/*

/**
 * Get authentication token from local storage
 * @returns {Object} Headers object with Authorization if token exists
 */
function getAuthHeader() {
  const token = localStorage.getItem('auth_token');
  if (!token) {
    return {};
  }
  return { 'Authorization': `Bearer ${token}` };
}

/**
 * Generic fetch wrapper with error handling
 * @param {string} endpoint - The API endpoint (without leading /)
 * @param {RequestInit} options - Fetch options (method, body, headers, etc.)
 * @returns {Promise<Object>} Parsed JSON response
 */
async function fetchAPI(endpoint, options = {}) {
  const defaultHeaders = {
    'Content-Type': 'application/json',
    ...getAuthHeader(),
  };

  const mergedHeaders = {
    ...defaultHeaders,
    ...(options.headers || {}),
  };

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers: mergedHeaders,
    });

    if (!response.ok) {
      // Handle specific error responses
      const errorData = await response.json().catch(() => ({}));

      throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
    }

    // Handle empty responses
    if (response.status === 204 || response.headers.get('Content-Length') === '0') {
      return {};
    }

    return response.json();
  } catch (error) {
    if (error instanceof Error) {
      throw error;
    }
    throw new Error('An unexpected error occurred');
  }
}

export { fetchAPI, getAuthHeader };
