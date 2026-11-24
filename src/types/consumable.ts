// Consumable types for MAT-153 Recovery System

export type ConsumableType = 'energy' | 'health' | 'recovery';

/**
 * Consumable definition from the reference table (GET /consumables)
 */
export interface Consumable {
  id: number;
  name: string;
  description?: string | null;
  consumable_type: ConsumableType;
  energy_restore: number;
  health_restore: number;
  fatigue_reduction: number;
  cooldown_hours: number;
  is_unlimited: boolean;
  created_at?: string;
  updated_at?: string;
}

/**
 * Boxer's consumable inventory item (GET /boxers/{id}/consumables)
 */
export interface ConsumableInventoryItem extends Consumable {
  quantity: number;
  last_used_at?: string | null;
  on_cooldown?: boolean;
  available_at?: string | null;
}

/**
 * Request body for using a consumable (POST /boxers/{id}/use-consumable)
 */
export interface UseConsumableRequest {
  consumable_id: number;
}

/**
 * Cooldown information returned after using a consumable
 */
export interface CooldownInfo {
  on_cooldown: boolean;
  available_at?: string;
  message: string;
}

/**
 * Consumed item details in response
 */
export interface ConsumedItem {
  id: number;
  name: string;
  description?: string | null;
  consumable_type: ConsumableType;
  energy_restore: number;
  health_restore: number;
  fatigue_reduction: number;
  cooldown_hours: number;
  is_unlimited: boolean;
}

/**
 * Response from using a consumable
 */
export interface UseConsumableResponse {
  success: boolean;
  message: string;
  consumed_item?: ConsumedItem | null;
  cooldown_info?: CooldownInfo | null;
}
