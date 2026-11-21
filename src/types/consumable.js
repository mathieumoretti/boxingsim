/**
 * consumable - Consumable item type definitions (MAT-153)
 *
 * Documentation for consumable-related data structures.
 * Actual validation handled by backend API responses.
 *
 * @module types/consumable
 */

/**
 * @typedef {'energy' | 'health' | 'recovery'} ConsumableType
 * The category of effect this consumable provides
 */

/**
 * @typedef {Object} Consumable
 * Reference data for a consumable type (from GET /consumables)
 * @property {number} id - Unique identifier
 * @property {string} name - Display name
 * @property {string} [description] - Optional description of effects
 * @property {ConsumableType} consumable_type - Category of this item
 * @property {number} energy_restore - Amount of energy restored (0 if none)
 * @property {number} health_restore - Amount of health restored (0 if none)
 * @property {number} fatigue_reduction - Amount of fatigue reduced (0 if none)
 * @property {number} cooldown_hours - Cooldown duration in hours (0 if unlimited)
 * @property {boolean} is_unlimited - If true, no quantity/cooldown tracking
 */

/**
 * @typedef {Object} ConsumableInventoryItem
 * Boxer's inventory item with current state
 * @property {number} id - Consumable ID (reference to catalog)
 * @property {string} name - Display name
 * @property {string} [description] - Optional description
 * @property {ConsumableType} consumable_type - Category
 * @property {number} energy_restore - Energy restoration amount
 * @property {number} health_restore - Health restoration amount
 * @property {number} fatigue_reduction - Fatigue reduction amount
 * @property {number} cooldown_hours - Cooldown duration in hours
 * @property {boolean} is_unlimited - Unlimited use flag
 * @property {number} quantity - Remaining uses (0 if unlimited)
 * @property {boolean} on_cooldown - Currently on cooldown
 * @property {string} [available_at] - ISO datetime when cooldown expires
 */

/**
 * @typedef {Object} UseConsumableRequest
 * Request body for using a consumable
 * @property {number} consumable_id - ID of consumable to use
 */

/**
 * @typedef {Object} UsedItem
 * Information about the item that was consumed
 * @property {number} id - Consumable ID
 * @property {string} name - Display name
 * @property {ConsumableType} type - Category
 */

/**
 * @typedef {Object} CooldownInfo
 * Information about an active cooldown
 * @property {boolean} on_cooldown - Whether this item is on cooldown
 * @property {string} available_at - ISO datetime when ready to use again
 * @property {string} message - Human-readable countdown message
 */

/**
 * @typedef {Object} UseConsumableResponse
 * Response from POST /boxers/{id}/use-consumable
 * @property {boolean} success - Whether the operation succeeded
 * @property {string} [message] - Success or error message
 * @property {UsedItem} [consumed_item] - Item that was used (on success)
 * @property {CooldownInfo} [cooldown_info] - Cooldown details (if applicable)
 */

// Export type names as comments for documentation purposes
// Actual values are dynamic from API responses
export {};
