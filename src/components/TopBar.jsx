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

import React from 'react';
import { Link } from 'react-router-dom';
import Navigation from './Navigation';
import MiniWorldClock from './MiniWorldClock';
import './TopBar.css';

const TopBar = ({ currentUser, onLogout }) => {
  const handleLogout = () => {
    onLogout();
  };

  return (
    <nav className="top-bar" role="navigation" aria-label="Global navigation">
      {/* Left: Brand */}
      <div className="top-bar-left">
        <Link to="/dashboard" className="brand-logo" aria-label="Boxing Simulator Home">
          <span className="brand-icon" aria-hidden="true">🥊</span>
          <span className="brand-title">Boxing Sim</span>
        </Link>
      </div>

      {/* Center: Primary Navigation + World Clock */}
      <div className="top-bar-center">
        <Navigation activeView={getCurrentView()} />
        <div className="world-clock-wrapper">
          <MiniWorldClock />
        </div>
      </div>

      {/* Right: User Actions */}
      <div className="top-bar-right">
        <div className="user-actions">
          <Link to="/create-boxer" className="btn-add-fighter">
            <span aria-hidden="true">+</span>
            <span>Add Fighter</span>
          </Link>

          {currentUser && (
            <div className="user-menu">
              <span className="user-welcome">{currentUser.username}</span>
              <button onClick={handleLogout} className="btn-logout" aria-label="Logout">
                Logout
              </button>
            </div>
          )}
        </div>
      </div>
    </nav>
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
