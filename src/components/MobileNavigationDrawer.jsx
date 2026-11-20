/**
 * MobileNavigationDrawer Component (MAT-131 Phase 1)
 *
 * Slide-out navigation drawer for mobile devices.
 * Implements focus trap and keyboard accessibility.
 *
 * @param {boolean} isOpen - Whether the drawer is currently open
 * @param {Function} onClose - Callback to close the drawer
 */

import React, { useEffect, useRef } from 'react';
import { NavLink } from 'react-router-dom';
import MiniWorldClock from './MiniWorldClock';
import './MobileNavigationDrawer.css';

const navLinks = [
  { to: '/dashboard', icon: '📊', label: 'Dashboard' },
  { to: '/boxers', icon: '🥊', label: 'Boxers' },
  { to: '/rankings', icon: '🏆', label: 'Rankings' },
  { to: '/world/all-fighters', icon: '🌍', label: 'Fighter World' },
];

export function MobileNavigationDrawer({ isOpen, onClose }) {
  const drawerRef = useRef(null);
  const focusableElementsRef = useRef([]);

  // Close on Escape key
  useEffect(() => {
    const handleEscape = (e) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      // Focus first element when drawer opens
      setTimeout(() => {
        if (focusableElementsRef.current[0]) {
          focusableElementsRef.current[0].focus();
        }
      }, 100); // Allow animation to start
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
        className={`mobile-drawer-backdrop ${isOpen ? 'open' : ''}`}
        onClick={onClose}
        aria-hidden="true"
      ></div>

      {/* Drawer */}
      <nav
        id="mobile-navigation-drawer"
        ref={drawerRef}
        className={`mobile-navigation-drawer ${isOpen ? 'open' : ''}`}
        role="navigation"
        aria-label="Mobile navigation menu"
        aria-hidden={!isOpen}
      >
        <div className="drawer-header">
          <span className="drawer-title">Boxing Sim</span>
          <button
            onClick={onClose}
            className="drawer-close-btn"
            aria-label="Close menu"
          >
            <span aria-hidden="true">×</span>
          </button>
        </div>

        <div className="drawer-content">
          <ul className="drawer-nav-list">
            {navLinks.map((link) => (
              <li key={link.to}>
                <NavLink
                  to={link.to}
                  className="drawer-nav-link"
                  onClick={handleLinkClick}
                  end={link.to === '/dashboard'}
                >
                  <span className="drawer-nav-icon" aria-hidden="true">
                    {link.icon}
                  </span>
                  <span className="drawer-nav-label">{link.label}</span>
                </NavLink>
              </li>
            ))}
          </ul>

          <div className="drawer-world-clock">
            <MiniWorldClock />
          </div>
        </div>

        <div className="drawer-footer">
          <NavLink
            to="/create-boxer"
            className="drawer-create-btn"
            onClick={handleLinkClick}
          >
            <span aria-hidden="true">+</span>
            <span>Add Fighter</span>
          </NavLink>
        </div>
      </nav>
    </>
  );
}
