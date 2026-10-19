/**
 * Training API - Data access layer for training operations (MAT-120)
 */

import { apiGetJson, apiPostJson } from './client';

/**
 * Fetch all available training types.
 *
 * @returns {Promise<Array>} Array of training type objects
 */
export async function fetchTrainingTypes() {
  const data = await apiGetJson(`/api/training-types`);
  return Array.isArray(data) ? data : [];
}

/**
 * Fetch active training sessions for a boxer.
 *
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Array>} Array of active training session objects
 */
export async function fetchActiveTrainingSessions(boxerId) {
  const data = await apiGetJson(`/api/boxers/${boxerId}/training-sessions`);
  return Array.isArray(data) ? data : [];
}

/**
 * Start a new training session for a boxer.
 *
 * @param {number} boxerId - The boxer's ID
 * @param {Object} trainingData - Training session data
 * @param {number} trainingData.trainingTypeId - Type of training (1=Strength, 2=Defense, 3=Agility)
 * @returns {Promise<Object>} Created training session object
 */
export async function startTrainingSession(boxerId, trainingData) {
  const data = await apiPostJson(`/api/boxers/${boxerId}/train`, trainingData);
  return data;
}

/**
 * Complete multiple active training sessions for a boxer.
 *
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Object>} Result with completed session count and updated boxer stats
 */
export async function completeActiveTrainingSessions(boxerId) {
  const data = await apiPostJson(`/api/training/bulk-complete`, { boxer_id: boxerId });
  return data;
}

/**
 * Get training progress information for a specific session.
 *
 * @param {Object} session - Training session object with start_time and duration_hours
 * @returns {Object} Progress info with isComplete, elapsedHours, remainingHours, percentComplete
 */
export function calculateTrainingProgress(session) {
  if (!session || !session.start_time || !session.duration_hours) {
    return {
      isComplete: false,
      elapsedHours: 0,
      remainingHours: 0,
      percentComplete: 0
    };
  }

  const startTime = new Date(session.start_time);
  const now = new Date();

  // Use game time if available (passed in via context)
  const elapsedMs = now - startTime;
  const elapsedHours = elapsedMs / (1000 * 60 * 60);

  const duration = parseFloat(session.duration_hours);
  const remaining = Math.max(0, duration - elapsedHours);
  const percent = Math.min(100, (elapsedHours / duration) * 100);

  return {
    isComplete: elapsedHours >= duration,
    elapsedHours: Math.round(elapsedHours * 100) / 100,
    remainingHours: Math.round(remaining * 100) / 100,
    percentComplete: Math.round(percent * 100) / 100
  };
}

/**
 * Fetch training history (completed sessions) for a boxer.
 *
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Array>} Array of completed training session objects with gains
 */
export async function fetchTrainingHistory(boxerId) {
  const data = await apiGetJson(`/api/boxers/${boxerId}/training-history`);
  return Array.isArray(data) ? data : [];
}

/**
 * Fetch stat progression data for a boxer (time-series of STR/DEF/AGI values).
 *
 * @param {number} boxerId - The boxer's ID
 * @returns {Promise<Object>} Object with points array containing date and stat values
 */
export async function fetchStatProgression(boxerId) {
  const data = await apiGetJson(`/api/boxers/${boxerId}/stat-progression`);
  return data || { points: [] };
}
