/**
 * User/Auth API - Data access layer for authentication and user operations (MAT-120)
 */

import { apiGetJson, apiPostJson } from './client';

/**
 * Register a new user.
 *
 * @param {Object} userData - Registration data
 * @param {string} userData.username - Desired username
 * @param {string} userData.password - Password
 * @param {string} userData.confirmPassword - Password confirmation
 * @returns {Promise<Object>} Response with token and user object on success
 */
export async function registerUser(userData) {
  const data = await apiPostJson(`/api/auth/register`, userData);
  return data;
}

/**
 * Login with username and password.
 *
 * @param {Object} credentials - Login credentials
 * @param {string} credentials.username - Username
 * @param {string} credentials.password - Password
 * @returns {Promise<Object>} Response with token and user object on success
 */
export async function login(credentials) {
  const data = await apiPostJson(`/api/auth/login`, credentials);
  return data;
}

/**
 * Fetch current authenticated user details.
 *
 * @returns {Promise<Object>} Current user object
 */
export async function getCurrentUser() {
  const data = await apiGetJson(`/api/auth/me`);
  return data;
}

/**
 * Fetch boxers owned by the current user.
 *
 * @param {number} userId - User ID
 * @returns {Promise<Array>} Array of boxer objects owned by the user
 */
export async function fetchUserBoxers(userId) {
  const data = await apiGetJson(`/api/users/${userId}/boxers`);
  return Array.isArray(data) ? data : [];
}
