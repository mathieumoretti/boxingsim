/**
 * SkipLink Component (MAT-131 Phase 1)
 *
 * Accessibility skip link that allows keyboard users to bypass
 * navigation and jump directly to main content.
 *
 * @param {string} targetId - ID of the element to skip to
 */

import React from 'react';
import './SkipLink.css';

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
      className="skip-link"
    >
      Skip to main content
    </a>
  );
}
