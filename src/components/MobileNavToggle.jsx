/**
 * MobileNavToggle Component (MAT-131 Phase 1)
 *
 * Hamburger menu button for mobile navigation.
 * Accessible with ARIA attributes and keyboard support.
 *
 * @param {boolean} isOpen - Whether the mobile menu is currently open
 * @param {Function} onToggle - Callback to toggle menu open/close state
 */

import React from 'react';
import './MobileNavToggle.css';

export function MobileNavToggle({ isOpen, onToggle }) {
  const handleClick = () => {
    onToggle();
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onToggle();
    }
  };

  return (
    <button
      className="mobile-nav-toggle"
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      aria-label={isOpen ? 'Close navigation menu' : 'Open navigation menu'}
      aria-expanded={isOpen}
      aria-controls="mobile-navigation-drawer"
    >
      <span className="hamburger-icon">
        <span className={`bar bar-top`}></span>
        <span className={`bar bar-middle`}></span>
        <span className={`bar bar-bottom`}></span>
      </span>
    </button>
  );
}
