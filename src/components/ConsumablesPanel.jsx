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

  // Determine badge color based on type
  const typeColors = {
    energy: 'badge-energy',
    health: 'badge-health',
    recovery: 'badge-recovery',
  };

  return (
    <div className={`consumable-card ${typeColors[item.consumable_type]} ${!isAvailable ? 'consumable-card-disabled' : ''}`}>
      <div className="consumable-card-header">
        <h3 className="consumable-card-name">{item.name}</h3>
        {isUnlimited ? (
          <span className="consumable-badge consumable-badge-unlimited">∞</span>
        ) : (
          <span className={`consumable-badge consumable-badge-quantity ${quantity === 0 ? 'badge-out-of-stock' : ''}`}>
            ×{quantity}
          </span>
        )}
      </div>

      {item.description && (
        <p className="consumable-card-description">{item.description}</p>
      )}

      {/* Stats display */}
      <div className="consumable-card-stats">
        {item.energy_restore > 0 && (
          <span className="consumable-stat consumable-stat-energy">
            ⚡ +{Math.round(item.energy_restore)} Energy
          </span>
        )}
        {item.health_restore > 0 && (
          <span className="consumable-stat consumable-stat-health">
            ❤️ +{Math.round(item.health_restore)} Health
          </span>
        )}
        {item.fatigue_reduction > 0 && (
          <span className="consumable-stat consumable-stat-recovery">
            🌿 −{Math.round(item.fatigue_reduction)} Fatigue
          </span>
        )}
      </div>

      {/* Cooldown indicator for limited items */}
      {!isUnlimited && item.cooldown_hours > 0 && (
        <div className="consumable-card-cooldown">
          <span className="cooldown-label">⏱️ {item.cooldown_hours}h cooldown</span>
        </div>
      )}

      {/* Use button */}
      <button
        className={`consumable-card-button ${typeColors[item.consumable_type]}`}
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
    <div className="consumables-section">
      <h3 className="consumables-section-header">
        <span className="section-icon">{icon}</span>
        {title}
      </h3>
      {items.length > 0 ? (
        <div className="consumables-grid">
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
        <p className="consumables-empty">{emptyMessage}</p>
      )}
    </div>
  );

  // Toast notification component
  if (toast) {
    return (
      <div className="consumables-panel-overlay">
        <div className={`consumables-toast ${toast.type}`}>
          {toast.type === 'success' ? '✅' : '⚠️'}
          {toast.message}
        </div>

        {/* Re-render main panel behind toast */}
        <div className="consumables-panel">
          <header className="consumables-panel-header">
            <h2>🎒 Consumables</h2>
            <p className="consumables-panel-subtitle">Use items to restore energy, health, and reduce fatigue</p>
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
    <div className="consumables-panel">
      <header className="consumables-panel-header">
        <h2>🎒 Consumables</h2>
        <p className="consumables-panel-subtitle">Use items to restore energy, health, and reduce fatigue</p>
      </header>

      {error && (
        <div className="consumables-error" role="alert">
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
        <div className="consumables-empty-state">
          <p>📦 Your inventory is empty. Consumables may be awarded through gameplay events.</p>
        </div>
      )}
    </div>
  );
};

export default ConsumablesPanel;
