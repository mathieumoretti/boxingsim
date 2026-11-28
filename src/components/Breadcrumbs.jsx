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

const Breadcrumbs = ({ items = [], separator = ' / ' }) => {
  if (!items.length) return null;

  return (
    <nav className="flex items-center gap-2 text-sm" aria-label="Breadcrumb navigation">
      <ol className="flex items-center gap-2">
        {items.map((item, index) => {
          const isLast = index === items.length - 1;

          return (
            <li key={index} className="flex items-center gap-2">
              {isLast ? (
                // Last item - current page (non-clickable)
                <span className="font-semibold text-white" aria-current="page">
                  {item.label}
                </span>
              ) : (
                // Previous items - clickable links
                <Link
                  to={item.href || '#'}
                  className="font-medium text-primary-light hover:text-white transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2 rounded"
                  onClick={(e) => {
                    if (!item.href) e.preventDefault();
                  }}
                >
                  {item.label}
                </Link>
              )}

              {/* Separator (not shown on last item) */}
              {!isLast && (
                <span className="text-white/40" aria-hidden="true">
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
