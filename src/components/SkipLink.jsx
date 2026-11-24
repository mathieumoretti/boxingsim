/**
 * SkipLink Component (MAT-131 Phase 1)
 *
 * Accessibility skip link that allows keyboard users to bypass
 * navigation and jump directly to main content.
 *
 * @param {string} targetId - ID of the element to skip to
 */

import React from 'react';

export function SkipLink({ targetId = 'main-content' }) {
  const handleSkip = (e) => {
    e.preventDefault();
    const target = document.getElementById(targetId);
    if (target) {
      target.focus();
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
  };

  return (
    <a
      href={`#${targetId}`}
      onClick={handleSkip}
      className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4 focus:z-[9999] focus:bg-primary focus:text-white focus:px-4 focus:py-2 focus:rounded-md focus:shadow-[0_4px_16px_rgba(0,0,0,0.3)] focus:outline-none transition-all duration-150"
    >
      Skip to main content
    </a>
  );
}
