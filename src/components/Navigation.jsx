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

const Navigation = ({ activeView }) => {
  const navItems = [
    { to: '/dashboard', label: 'Dashboard', icon: '📊', view: 'dashboard' },
    { to: '/boxers', label: 'Boxers', icon: '🥊', view: 'boxers' },
    { to: '/rankings', label: 'Rankings', icon: '🏆', view: 'rankings' },
    { to: '/world/events', label: 'World', icon: '🌍', view: 'world' },
  ];

  const handleClick = () => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <nav className="flex justify-center items-center py-2" aria-label="Primary navigation">
      <ul className="flex gap-1 list-none margin-0 padding-0">
        {navItems.map((item) => (
          <li key={item.to} className="flex">
            <NavLink
              to={item.to}
              onClick={handleClick}
              className={({ isActive }) => `
                flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md transition-all duration-150 relative letter-spacing-[0.01em]
                ${isActive
                  ? 'text-primary-light bg-primary/15'
                  : 'text-text-muted hover:text-text-bright hover:bg-primary/10'}
                focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2
                after:absolute after:content-[''] after:bottom-0 after:left-1/2 after:w-0 after:h-[2px] after:bg-primary after:-translate-x-1/2 after:transition-all after:duration-250 after:opacity-0
                hover:after:w-[calc(100%-1.5rem)] hover:after:opacity-100
                ${isActive ? 'after:w-[calc(100%-1.5rem)] after:opacity-100' : ''}
              `}
            >
              <span className="text-base leading-none md:hidden lg:block xl:block" aria-hidden="true">
                {item.icon}
              </span>
              <span className="hidden md:block whitespace-nowrap">{item.label}</span>

              {/* Mobile: icon only */}
              <span className="md:hidden text-xl leading-none" aria-hidden="true">
                {item.icon}
              </span>
            </NavLink>
          </li>
        ))}
      </ul>
    </nav>
  );
};

export default Navigation;
