import React from 'react';
import './PageLayout.css';

/**
 * PageLayout - Standardized page wrapper component
 *
 * Provides consistent layout structure for dedicated pages with:
 * - Back button (optional)
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
 *
 * @example
 * // Basic usage with back button
 * <PageLayout onBack={() => navigate('/dashboard')}>
 *   <div className="content">...</div>
 * </PageLayout>
 *
 * @example
 * // Custom back button label
 * <PageLayout
 *   onBack={() => navigate('/dashboard')}
 *   backButtonLabel="← Back to Dashboard"
 * >
 *   {...children}
 * </PageLayout>
 *
 * @example
 * // Without back button
 * <PageLayout showBackButton={false}>
 *   {...children}
 * </PageLayout>
 */
const PageLayout = ({
  children,
  className = '',
  onBack,
  showBackButton = true,
  backButtonLabel = '← Back'
}) => {
  return (
    <div className="page-layout">
      {/* Back Button (optional) */}
      {showBackButton && onBack && (
        <div className="page-nav">
          <button
            onClick={onBack}
            className="back-btn"
            aria-label="Go back"
          >
            {backButtonLabel}
          </button>
        </div>
      )}

      {/* Content Wrapper */}
      <main className={`page-content ${className}`}>
        {children}
      </main>
    </div>
  );
};

export default PageLayout;
