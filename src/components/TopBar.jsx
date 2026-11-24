/**
 * TopBar Component
 *
 * Global navigation bar implementing "The Four Corners" structure:
 * - Left: Brand logo + app title
 * - Center: Primary 4-view navigation (Dashboard | Boxers | Rankings | World)
 *       + World Clock (game time anchor)
 * - Right: User actions (Add Fighter, user menu, logout)
 *
 * @param {Object} props
 * @param {Object} [props.currentUser] - Current user data
 * @param {function} props.onLogout - Logout handler callback
 */

import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import Navigation from './Navigation';
import MiniWorldClock from './MiniWorldClock';
import { MobileNavToggle } from './MobileNavToggle';
import { MobileNavigationDrawer } from './MobileNavigationDrawer';

const TopBar = ({ currentUser, onLogout }) => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const handleLogout = () => {
    onLogout();
  };

  const handleMobileNavToggle = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  return (
    <>
      <nav className="sticky top-0 z-[var(--z-sticky)] flex items-center justify-between gap-6 px-6 py-3 bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light shadow-topbar backdrop-blur-md">
        {/* Left: Brand */}
        <div className="flex items-center shrink-0">
          <Link
            to="/dashboard"
            className="flex items-center gap-2 text-decoration-none transition-opacity duration-150 hover:opacity-90 active:scale-95 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2 focus-visible:rounded-md"
            aria-label="Boxing Simulator Home"
          >
            <span className="text-xl leading-none" aria-hidden="true">🥊</span>
            <span className="text-lg font-bold text-white letter-spacing-[-0.01em] drop-shadow-md md:hidden lg:block xl:block">
              Boxing Sim
            </span>
          </Link>
        </div>

        {/* Center: Primary Navigation + World Clock (hidden on mobile) */}
        <div className="flex items-center justify-center gap-8 flex-1 min-w-0 hidden lg:flex">
          <Navigation activeView={getCurrentView()} />
          <div className="flex items-center pl-6 border-l border-white/10">
            <MiniWorldClock />
          </div>
        </div>

        {/* Right: Mobile Toggle + User Actions */}
        <div className="flex items-center justify-end shrink-0">
          {/* Mobile navigation toggle - visible on mobile only */}
          <MobileNavToggle
            isOpen={isMobileMenuOpen}
            onToggle={handleMobileNavToggle}
          />

          <div className="flex items-center gap-4">
            <Link to="/create-boxer" className="flex items-center gap-1 px-4 py-2 text-sm font-medium text-white bg-gradient-to-br from-primary-dark to-primary rounded-md cursor-pointer transition-all duration-150 shadow-[0_2px_8px_rgba(37,99,235,0.4)] hover:bg-gradient-to-br hover:from-primary hover:to-primary-light hover:shadow-[0_4px_12px_rgba(37,99,235,0.6)] hover:-translate-y-1 active:translate-y-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-white focus-visible:outline-offset-2 focus-visible:shadow-[0_0_0_4px_rgba(59,130,246,0.3)]">
              <span className="text-lg font-bold leading-none md:hidden lg:block xl:block">+</span>
              <span className="hidden md:block">Add Fighter</span>
            </Link>

            {currentUser && (
              <div className="flex items-center gap-3">
                <span className="text-sm text-text-muted font-medium truncate max-w-[120px] hidden md:block">
                  {currentUser.username}
                </span>
                <button
                  onClick={handleLogout}
                  className="px-3 py-2 text-sm font-medium text-text-muted bg-transparent border border-white/20 rounded-md cursor-pointer transition-all duration-150 hover:text-danger-bright hover:border-danger hover:bg-danger/10 active:scale-95 focus-visible:outline focus-visible:outline-2 focus-visible:outline-danger focus-visible:outline-offset-2"
                  aria-label="Logout"
                >
                  Logout
                </button>
              </div>
            )}
          </div>
        </div>
      </nav>

      {/* Mobile navigation drawer */}
      <MobileNavigationDrawer
        isOpen={isMobileMenuOpen}
        onClose={() => setIsMobileMenuOpen(false)}
      />
    </>
  );
};

/**
 * Determine current view based on pathname for active state highlighting
 */
function getCurrentView() {
  const path = window.location.pathname;

  if (path === '/dashboard') return 'dashboard';
  if (path.startsWith('/boxer/')) return 'boxers';
  if (path.startsWith('/rankings')) return 'rankings';
  if (path.startsWith('/world/')) return 'world';

  // Default to dashboard for unknown routes
  return 'dashboard';
}

export default TopBar;
