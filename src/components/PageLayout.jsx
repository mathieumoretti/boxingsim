import React from 'react';
import './PageLayout.css';

/**
 * PageLayout - Standardized page wrapper component
 *
 * Provides consistent layout structure for dedicated pages with:
 * - Back button (optional)
 * - Cross-navigation button (optional - e.g., to related pages)
 * - Content wrapper with responsive breakpoints
 *
 * Note: TopBar is rendered globally in App.js, so this component
 * only handles the back button and content wrapper.
 *
 * @param {Object} props - Component props
 * @param {function} [props.onBack] - Callback for back button (usually navigate)
 * @param {React.ReactNode} props.children - Page content
 * @param {string} [props.className] - Additional classes for content wrapper
 * @param {boolean} [props.showBackButton=true] - Toggle back button visibility
 * @param {string} [props.backButtonLabel='← Back'] - Custom back button text
 * @param {string} [props.crossNavTo] - URL for cross-navigation button
 * @param {string} [props.crossNavLabel] - Label for cross-navigation button
 * @param {boolean} [props.showCrossNav=false] - Toggle cross-nav visibility
 */
const PageLayout = ({
  children,
  className = '',
  pageClassName = '', // Additional classes for the .page-layout container itself
  onBack,
  showBackButton = true,
  backButtonLabel = '← Back',
  crossNavTo,
  crossNavLabel,
  showCrossNav = true
}) => {
  const hasCrossNav = showCrossNav && crossNavTo && crossNavLabel;

  return (
    <div className={`page-layout ${pageClassName}`}>
      {/* Navigation Bar */}
      <div className="page-nav">
        {/* Back Button */}
        {showBackButton && onBack && (
          <button
            onClick={onBack}
            className="back-btn"
            aria-label="Go back"
          >
            {backButtonLabel}
          </button>
        )}

        {/* Cross-Navigation Button */}
        {hasCrossNav && (
          <a
            href={crossNavTo}
            className="cross-nav-btn"
            aria-label={crossNavLabel}
          >
            {crossNavLabel}
          </a>
        )}
      </div>

      {/* Content Wrapper */}
      <main className={`page-content ${className}`}>
        {children}
      </main>
    </div>
  );
};

export default PageLayout;
