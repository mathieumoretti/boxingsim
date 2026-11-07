/**
 * Navigation Component
 *
 * Primary navigation bar for the Boxing Simulator application.
 * Implements the 4-view structure: Dashboard | Boxers | Rankings | World
 *
 * @param {Object} props
 * @param {string} props.activeView - Current active view for highlighting
 */

import React from 'react';
import { NavLink } from 'react-router-dom';
import './Navigation.css';

const Navigation = ({ activeView }) => {
  return (
    <nav className="navigation" aria-label="Primary navigation">
      <ul className="nav-list">
        <li className="nav-item">
          <NavLink
            to="/dashboard"
            className={`nav-link ${activeView === 'dashboard' ? 'nav-link--active' : ''}`}
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          >
            <span className="nav-link-icon" aria-hidden="true">📊</span>
            <span className="nav-link-text">Dashboard</span>
          </NavLink>
        </li>

        <li className="nav-item">
          <NavLink
            to="/boxers"
            className={`nav-link ${activeView === 'boxers' ? 'nav-link--active' : ''}`}
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          >
            <span className="nav-link-icon" aria-hidden="true">🥊</span>
            <span className="nav-link-text">Boxers</span>
          </NavLink>
        </li>

        <li className="nav-item">
          <NavLink
            to="/rankings"
            className={`nav-link ${activeView === 'rankings' ? 'nav-link--active' : ''}`}
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          >
            <span className="nav-link-icon" aria-hidden="true">🏆</span>
            <span className="nav-link-text">Rankings</span>
          </NavLink>
        </li>

        <li className="nav-item">
          <NavLink
            to="/world/events"
            className={`nav-link ${activeView === 'world' ? 'nav-link--active' : ''}`}
            onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
          >
            <span className="nav-link-icon" aria-hidden="true">🌍</span>
            <span className="nav-link-text">World</span>
          </NavLink>
        </li>
      </ul>
    </nav>
  );
};

export default Navigation;
