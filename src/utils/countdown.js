/**
 * Formats a remaining time in seconds into a human-readable countdown string.
 * @param {number} remainingSeconds - The remaining time in seconds
 * @returns {string} Formatted countdown (e.g., "2h 30m", "45m 30s", "30s")
 */
export const formatCountdown = (remainingSeconds) => {
  // Don't show anything if time has elapsed
  if (remainingSeconds <= 0) return '';

  const hours = Math.floor(remainingSeconds / 3600);
  const minutes = Math.floor((remainingSeconds % 3600) / 60);
  const seconds = remainingSeconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds}s`;
  } else {
    return `${seconds}s`;
  }
};
