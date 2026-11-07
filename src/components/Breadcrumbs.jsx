/**
 * Breadcrumbs Component
 *
 * Provides page-level context navigation showing the user's location
 * within the application hierarchy.
 *
 * Usage patterns:
 * - Dashboard / Boxers / [Fighter Name]
 * - Rankings / Weight Class / Fighter Details
 *
 * @param {Object} props
 * @param {Array} props.items - Array of breadcrumb items { label, href }
 * @param {string} [props.separator=' / '] - Separator between items
 */

import React from 'react';
import { Link } from 'react-router-dom';
import './Breadcrumbs.css';

const Breadcrumbs = ({ items = [], separator = ' / ' }) => {
  if (!items.length) return null;

  return (
    <nav className="breadcrumbs" aria-label="Breadcrumb navigation">
      <ol className="breadcrumb-list">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;

          return (
            <li key={index} className="breadcrumb-item">
              {isLast ? (
                // Last item - current page (non-clickable)
                <span className="breadcrumb-link breadcrumb-link--current" aria-current="page">
                  {item.label}
                </span>
              ) : (
                // Previous items - clickable links
                <Link
                  to={item.href || '#'}
                  className="breadcrumb-link"
                  onClick={(e) => {
                    if (!item.href) e.preventDefault();
                  }}
                >
                  {item.label}
                </Link>
              )}

              {/* Separator (not shown on last item) */}
              {!isLast && (
                <span className="breadcrumb-separator" aria-hidden="true">
                  {separator}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
};

export default Breadcrumbs;
