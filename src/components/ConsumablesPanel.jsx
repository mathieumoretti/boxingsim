/**
 * ConsumablesPanel - Display boxer's consumable inventory with use buttons (MAT-153)
 *
 * Shows items organized by type: Energy, Health, and Recovery.
 * Includes cooldown tracking and quantity management.
 */

import React, { useState } from 'react';
import { getConsumableTypeDisplayName, getConsumableIconClass } from '../utils/recoveryUtils';

/**
 * Individual consumable card component
 * @param {Object} props
 * @param {Object} props.item - ConsumableInventoryItem data
 * @param {Function} props.onUse - Callback when use button clicked
 * @param {boolean} [props.disabled] - Disable interaction
 */
const ConsumableCard = ({ item, onUse, disabled }) => {
  const isUnlimited = item.is_unlimited;
  const quantity = isUnlimited ? Infinity : item.quantity;
  const isAvailable = !disabled && quantity > 0;

  // Tailwind style objects for type-specific colors
  const typeColors = {
    energy: {
      card: 'bg-gradient-to-br from-green-50 to-emerald-100 border-green-400',
      badgeUnlimited: 'bg-cyan-100 text-sky-700',
      badgeQuantity: 'bg-gray-100 text-gray-700',
      badgeOutOfStock: 'bg-red-100 text-red-900',
      stat: 'bg-green-100 text-green-900',
      button: 'bg-gradient-to-br from-green-500 to-emerald-600 hover:from-green-400 hover:to-emerald-500',
    },
    health: {
      card: 'bg-gradient-to-br from-red-50 to-orange-100 border-red-400',
      badgeUnlimited: 'bg-cyan-100 text-sky-700',
      badgeQuantity: 'bg-gray-100 text-gray-700',
      badgeOutOfStock: 'bg-red-100 text-red-900',
      stat: 'bg-red-100 text-red-900',
      button: 'bg-gradient-to-br from-red-500 to-rose-600 hover:from-red-400 hover:to-rose-500',
    },
    recovery: {
      card: 'bg-gradient-to-br from-cyan-50 to-teal-100 border-cyan-400',
      badgeUnlimited: 'bg-cyan-100 text-sky-700',
      badgeQuantity: 'bg-gray-100 text-gray-700',
      badgeOutOfStock: 'bg-red-100 text-red-900',
      stat: 'bg-cyan-100 text-cyan-900',
      button: 'bg-gradient-to-br from-cyan-500 to-teal-600 hover:from-cyan-400 hover:to-teal-500',
    },
  };

  const colors = typeColors[item.consumable_type] || typeColors.energy;

  return (
    <div className={`flex flex-col p-4 bg-white border-2 rounded-lg transition-all duration-200 ${colors.card} ${!isAvailable ? 'opacity-60 cursor-not-allowed' : 'hover:border-gray-300 hover:shadow-[0_4px_8px_rgba(0,0,0,0.05)]'}`}>
      <div className="flex justify-between items-center mb-2">
        <h3 className="m-0 text-base font-semibold text-gray-900">{item.name}</h3>
        {isUnlimited ? (
          <span className={`inline-flex items-center justify-center px-2 py-0.5 text-sm font-semibold rounded ${colors.badgeUnlimited}`}>∞</span>
        ) : (
          <span className={`inline-flex items-center justify-center px-2 py-0.5 text-sm font-semibold rounded ${colors.badgeQuantity} ${quantity === 0 ? colors.badgeOutOfStock : ''}`}>
            ×{quantity}
          </span>
        )}
      </div>

      {item.description && (
        <p className="m-0 mb-3 text-sm leading-relaxed text-gray-600">{item.description}</p>
      )}

      {/* Stats display */}
      <div className="flex flex-wrap gap-1 mb-3">
        {item.energy_restore > 0 && (
          <span className={`px-2 py-1 text-xs font-semibold rounded ${colors.stat}`}>
            ⚡ +{Math.round(item.energy_restore)} Energy
          </span>
        )}
        {item.health_restore > 0 && (
          <span className={`px-2 py-1 text-xs font-semibold rounded ${colors.stat}`}>
            ❤️ +{Math.round(item.health_restore)} Health
          </span>
        )}
        {item.fatigue_reduction > 0 && (
          <span className={`px-2 py-1 text-xs font-semibold rounded ${colors.stat}`}>
            🌿 −{Math.round(item.fatigue_reduction)} Fatigue
          </span>
        )}
      </div>

      {/* Cooldown indicator for limited items */}
      {!isUnlimited && item.cooldown_hours > 0 && (
        <div className="mb-3">
          <span className="inline-block px-2 py-1 text-xs font-semibold rounded bg-amber-100 text-amber-900">
            ⏱️ {item.cooldown_hours}h cooldown
          </span>
        </div>
      )}

      {/* Use button */}
      <button
        className={`mt-auto py-2.5 px-4 text-base font-semibold border-none rounded-md cursor-pointer transition-all duration-200 text-white ${colors.button} focus-visible:outline focus-visible:outline-2 focus-visible:outline-gray-400 focus-visible:outline-offset-2`}
        onClick={() => onUse(item.id)}
        disabled={!isAvailable || item.on_cooldown}
        aria-label={`Use ${item.name}${item.on_cooldown ? ' (on cooldown)' : ''}`}
      >
        {!isUnlimited && quantity === 0
          ? 'Out of Stock'
          : item.on_cooldown
            ? 'On Cooldown'
            : 'Use Item'}
      </button>
    </div>
  );
};

/**
 * ConsumablesPanel Component
 * @param {Object} props
 * @param {number} props.boxerId - ID of the boxer whose inventory to display
 * @param {Array} props.inventory - Array of ConsumableInventoryItem objects
 * @param {Function} props.onLoad - Callback to refresh data after changes
 */
const ConsumablesPanel = ({ boxerId, inventory, onLoad }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [toast, setToast] = useState(null);

  // Group consumables by type for organized display
  const energyItems = inventory.filter(i => i.consumable_type === 'energy');
  const healthItems = inventory.filter(i => i.consumable_type === 'health');
  const recoveryItems = inventory.filter(i => i.consumable_type === 'recovery');

  const handleUseConsumable = async (consumableId) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(`/boxers/${boxerId}/use-consumable`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
        },
        body: JSON.stringify({ consumable_id: consumableId }),
      });

      const result = await response.json();

      if (result.success) {
        setToast({
          message: `Used ${result.consumed_item?.name || 'consumable'} successfully!`,
          type: 'success'
        });

        // Refresh inventory to update quantities and cooldowns
        onLoad();

        // Clear toast after 3 seconds
        setTimeout(() => setToast(null), 3000);
      } else {
        // Handle cooldown or other error responses
        const message = result.cooldown_info?.message || result.message || 'Failed to use consumable';
        setToast({ message, type: 'error' });
        setTimeout(() => setToast(null), 4000);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error occurred');
      setToast({ message: 'Failed to use consumable', type: 'error' });
      setTimeout(() => setToast(null), 3000);
    } finally {
      setLoading(false);
    }
  };

  /**
   * Section component for each consumable type
   */
  const Section = ({ title, items, icon, emptyMessage }) => (
    <div className="p-6 border-b border-gray-200 last:border-b-0">
      <h3 className="flex items-center gap-2 m-0 mb-4 text-lg font-semibold text-gray-800">
        <span className="text-xl">{icon}</span>
        {title}
      </h3>
      {items.length > 0 ? (
        <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-4">
          {items.map(item => (
            <ConsumableCard
              key={item.id}
              item={item}
              onUse={handleUseConsumable}
              disabled={loading}
            />
          ))}
        </div>
      ) : (
        <p className="text-center py-8 text-gray-400 italic">{emptyMessage}</p>
      )}
    </div>
  );

  // Toast notification component
  if (toast) {
    return (
      <div className="relative">
        {/* Fixed toast */}
        <div className={`fixed top-4 right-4 z-[9999] p-4 rounded-lg shadow-[0_10px_24px_rgba(0,0,0,0.1)] font-semibold animate-[slideIn_0.3s_ease-out] ${
          toast.type === 'success'
            ? 'bg-gradient-to-br from-green-100 to-emerald-100 border-2 border-green-500 text-green-900'
            : 'bg-gradient-to-br from-red-100 to-orange-100 border-2 border-red-500 text-red-900'
        }`}>
          {toast.type === 'success' ? '✅' : '⚠️'}
          {' '}{toast.message}
        </div>

        {/* Main panel behind toast */}
        <div className="bg-white rounded-xl shadow-[0_4px_8px_rgba(0,0,0,0.06)] overflow-hidden">
          <header className="p-6 bg-gradient-to-br from-gray-50 to-slate-100 border-b border-gray-200">
            <h2 className="m-0 mb-1 text-2xl font-bold text-gray-900">🎒 Consumables</h2>
            <p className="m-0 text-sm text-gray-600">Use items to restore energy, health, and reduce fatigue</p>
          </header>

          <Section
            title="Energy Restorers"
            items={energyItems}
            icon="⚡"
            emptyMessage="No energy items available"
          />

          <Section
            title="Health Restorers"
            items={healthItems}
            icon="❤️"
            emptyMessage="No health items available"
          />

          <Section
            title="Fatigue Reduction"
            items={recoveryItems}
            icon="🌿"
            emptyMessage="No recovery items available"
          />
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl shadow-[0_4px_8px_rgba(0,0,0,0.06)] overflow-hidden">
      <header className="p-6 bg-gradient-to-br from-gray-50 to-slate-100 border-b border-gray-200">
        <h2 className="m-0 mb-1 text-2xl font-bold text-gray-900">🎒 Consumables</h2>
        <p className="m-0 text-sm text-gray-600">Use items to restore energy, health, and reduce fatigue</p>
      </header>

      {error && (
        <div className="p-4 m-4 bg-red-50 border border-red-300 rounded-md text-red-900" role="alert">
          ⚠️ {error}
        </div>
      )}

      <Section
        title="Energy Restorers"
        items={energyItems}
        icon="⚡"
        emptyMessage="No energy items available"
      />

      <Section
        title="Health Restorers"
        items={healthItems}
        icon="❤️"
        emptyMessage="No health items available"
      />

      <Section
        title="Fatigue Reduction"
        items={recoveryItems}
        icon="🌿"
        emptyMessage="No recovery items available"
      />

      {inventory.length === 0 && !loading && (
        <div className="p-12 text-center text-gray-600">
          <p className="m-0">📦 Your inventory is empty. Consumables may be awarded through gameplay events.</p>
        </div>
      )}
    </div>
  );
};

export default ConsumablesPanel;
