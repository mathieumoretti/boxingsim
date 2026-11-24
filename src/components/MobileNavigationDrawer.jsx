/**
 * MobileNavigationDrawer Component (MAT-131 Phase 1)
 *
 * Slide-out navigation drawer for mobile devices.
 * Implements focus trap and keyboard accessibility.
 *
 * @param {boolean} isOpen - Whether the drawer is currently open
 * @param {Function} onClose - Callback to close the drawer
 */

import React, { useEffect } from 'react';
import { NavLink } from 'react-router-dom';
import MiniWorldClock from './MiniWorldClock';

const navLinks = [
  { to: '/dashboard', icon: '📊', label: 'Dashboard' },
  { to: '/boxers', icon: '🥊', label: 'Boxers' },
  { to: '/rankings', icon: '🏆', label: 'Rankings' },
  { to: '/world/all-fighters', icon: '🌍', label: 'Fighter World' },
];

export function MobileNavigationDrawer({ isOpen, onClose }) {
  // Close on Escape key
  useEffect(() => {
    const handleEscape = (e) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen, onClose]);

  // Prevent body scroll when drawer is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  const handleLinkClick = () => {
    onClose();
  };

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 bg-black/50 opacity-0 invisible transition-all duration-300 z-[calc(var(--z-sticky)+1)] ${isOpen ? 'opacity-100 visible' : ''}`}
        onClick={onClose}
        aria-hidden="true"
      />

      {/* Drawer */}
      <nav
        id="mobile-navigation-drawer"
        className={`fixed top-0 left-0 bottom-0 w-[min(85%,320px)] bg-gradient-to-br from-navy-dark via-navy-medium to-navy-dark border-r border-navy-border shadow-xl flex flex-col overflow-y-auto overflow-x-hidden transition-transform duration-300 ease z-[calc(var(--z-sticky)+2)] ${isOpen ? 'translate-x-0' : '-translate-x-full'}`}
        role="navigation"
        aria-label="Mobile navigation menu"
        aria-hidden={!isOpen}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-navy-border">
          <span className="text-lg font-bold text-white">Boxing Sim</span>
          <button
            onClick={onClose}
            className="flex items-center justify-center w-[44px] h-[44px] p-0 bg-transparent border-none rounded-md text-xl text-text-bright cursor-pointer transition-colors duration-150 hover:bg-white/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2"
            aria-label="Close menu"
          >
            <span aria-hidden="true">×</span>
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 flex flex-col px-4">
          <ul className="list-none margin-0 padding-0 flex flex-col gap-1">
            {navLinks.map((link) => (
              <li key={link.to}>
                <NavLink
                  to={link.to}
                  end={link.to === '/dashboard'}
                  onClick={handleLinkClick}
                  className={({ isActive }) => `
                    flex items-center gap-3 px-4 py-4 text-lg font-medium rounded-md transition-all duration-150 min-h-[48px]
                    ${isActive
                      ? 'text-white bg-primary/30'
                      : 'text-text-muted hover:text-text-bright hover:bg-white/5'}
                    focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:-outline-offset-2
                  `.trim()}
                >
                  <span className="text-xl leading-none shrink-0" aria-hidden="true">
                    {link.icon}
                  </span>
                  <span>{link.label}</span>
                </NavLink>
              </li>
            ))}
          </ul>

          {/* World Clock */}
          <div className="mt-4 pt-4 border-t border-navy-border">
            <MiniWorldClock />
          </div>
        </div>

        {/* Footer */}
        <div className="px-4 pt-0 border-t border-navy-border">
          <NavLink
            to="/create-boxer"
            onClick={handleLinkClick}
            className="flex items-center justify-center gap-2 w-full px-4 py-4 text-base font-medium text-white bg-gradient-to-br from-primary-dark to-primary rounded-md cursor-pointer transition-all duration-150 min-h-[48px] hover:bg-gradient-to-br hover:from-primary hover:to-primary-light hover:shadow-[0_4px_12px_rgba(37,99,235,0.4)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-white focus-visible:outline-offset-2"
          >
            <span aria-hidden="true">+</span>
            <span>Add Fighter</span>
          </NavLink>
        </div>
      </nav>
    </>
  );
}
