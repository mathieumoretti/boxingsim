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
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      aria-label={isOpen ? 'Close navigation menu' : 'Open navigation menu'}
      aria-expanded={isOpen}
      aria-controls="mobile-navigation-drawer"
      className={`
        inline-flex items-center justify-center w-[44px] h-[44px] p-0 rounded-md cursor-pointer transition-colors duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2
        ${isOpen ? 'bg-primary/10' : 'hover:bg-white/10'}
      `.trim()}
    >
      <span className="flex flex-col justify-between w-6 h-[18px] relative">
        {/* Top bar */}
        <span
          className={`block w-full h-[2px] bg-text-bright rounded-[1px] transition-all duration-150 ease ${isOpen ? 'rotate-45 translate-x-[5px] translate-y-[5px]' : ''}`}
        />

        {/* Middle bar */}
        <span
          className={`block w-full h-[2px] bg-text-bright rounded-[1px] transition-all duration-150 ease ${isOpen ? 'opacity-0 -translate-x-[10px]' : ''}`}
        />

        {/* Bottom bar */}
        <span
          className={`block w-full h-[2px] bg-text-bright rounded-[1px] transition-all duration-150 ease ${isOpen ? '-rotate-45 translate-x-[5px] -translate-y-[5px]' : ''}`}
        />
      </span>
    </button>
  );
}
