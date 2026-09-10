import React from 'react';
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
        <span className="user-welcome">Welcome, {currentUser?.username || 'User'}</span>
        <button onClick={onLogout} className="logout-btn">Logout</button>
      </div>
    </nav>
  );
};

export default TopBar;
