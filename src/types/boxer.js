/**
 * boxer - Boxer data type definitions (MAT-153)
 *
 * Documentation for boxer-related data structures including recovery fields.
 * Actual validation handled by backend API responses.
 *
 * @module types/boxer
 */

/**
 * @typedef {Object} Boxer
 * Complete boxer data from API
 * @property {number} id - Unique identifier
 * @property {string} name - Fighter name
 * @property {string} [nickname] - Optional nickname/ring name
 * @property {number} level - Current experience level
 * @property {number} experience - Current XP points
 * @property {number} strength - Strength stat
 * @property {number} defense - Defense stat
 * @property {number} agility - Agility stat
 * @property {number} energy - Current energy (0-100)
 * @property {number} health - Current health percentage (0-100)
 * @property {number} fatigue_score - Current fatigue level (0-100)
 * @property {string} [forced_recovery_until] - ISO datetime when forced recovery ends
 * @property {string} [next_available_at] - When boxer can next fight/train
 * @property {Object} [career_record] - Fight record stats
 * @property {number} [wins] - Total wins
 * @property {number} [losses] - Total losses
 * @property {number} [draws] - Total draws
 */

/**
 * @typedef {Object} RecoveryStatus
 * Computed recovery status information
 * @property {boolean} is_on_recovery - Currently on forced recovery
 * @property {Date} [recovery_until] - When recovery ends (parsed Date)
 * @property {string} time_remaining - Human-readable countdown
 * @property {number} [percentage_healed] - Healing progress percentage
 */

/**
 * @typedef {Object} FightEligibility
 * Result of fight eligibility check
 * @property {boolean} can_fight - Whether boxer is eligible to fight
 * @property {string} [reason] - Explanation if not eligible
 */

// Export type names as comments for documentation purposes
export {};
