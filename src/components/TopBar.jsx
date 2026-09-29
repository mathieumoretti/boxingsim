import React from 'react';
import { Link } from 'react-router-dom';
import './TopBar.css';
import MiniWorldClock from './MiniWorldClock';

const TopBar = ({ currentUser, onLogout }) => {
  return (
    <nav className="top-bar">
      <div className="top-bar-left">
        <h1 className="app-title">🥊 Boxing Simulator</h1>
      </div>

      <div className="top-bar-center">
        <MiniWorldClock />
      </div>

      <div className="top-bar-right">
        {/* Navigation Links */}
        <div className="nav-links">
          <Link to="/dashboard" className="nav-link">
            📊 Dashboard
          </Link>
          <Link to="/rankings" className="nav-link">
            🏆 Rankings
          </Link>
          <Link to="/create-boxer" className="nav-link">
            ➕ Create Boxer
          </Link>
        </div>

        <span className="user-welcome">Welcome, {currentUser?.username || 'User'}</span>
        <button onClick={onLogout} className="logout-btn">Logout</button>
      </div>
    </nav>
  );
};

export default TopBar;
