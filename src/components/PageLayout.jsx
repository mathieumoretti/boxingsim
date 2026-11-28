/**
 * PageLayout - Standardized page wrapper component
 *
 * Provides consistent layout structure for dedicated pages with:
 * - Unified navy page background (full-bleed)
 * - Back button (optional)
 * - Cross-navigation button (optional - e.g., to related pages)
 * - Content wrapper with responsive breakpoints
 *
 * Note: TopBar is rendered by each page; the top padding here clears
 * the sticky 64px TopBar plus breathing room.
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

import React from 'react';
import { Button } from './ui';

const PageLayout = ({
  children,
  className = '',
  pageClassName = '', // Additional classes for the inner content container
  onBack,
  showBackButton = true,
  backButtonLabel = '← Back',
  crossNavTo,
  crossNavLabel,
  showCrossNav = true,
}) => {
  const hasCrossNav = showCrossNav && crossNavTo && crossNavLabel;

  return (
    <div className="min-h-screen w-full bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
      <div className={`pt-[calc(theme(spacing.6)+64px)] px-6 pb-10 max-w-[1200px] mx-auto ${pageClassName}`}>
        {/* Navigation Bar */}
        <div className="flex justify-between items-center gap-4 mb-6">
          {/* Back Button */}
          {showBackButton && onBack && (
            <Button
              variant="ghost"
              size="md"
              onClick={onBack}
              aria-label="Go back"
              className="bg-white/10 text-white border border-white/20 hover:bg-white/15 hover:border-white/30"
            >
              {backButtonLabel}
            </Button>
          )}

          {/* Cross-Navigation Button */}
          {hasCrossNav && (
            <a
              href={crossNavTo}
              className="inline-flex items-center gap-2 px-5 py-3 bg-gradient-to-br from-navy-medium to-navy-light text-white border border-white/20 rounded-md text-sm font-semibold cursor-pointer transition-all duration-150 shadow-[0_2px_8px_rgba(0,0,0,0.2)] hover:bg-gradient-to-br hover:from-navy-light hover:to-navy-medium hover:border-white/30 hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(0,0,0,0.3)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent-purple-light focus-visible:outline-offset-2"
              aria-label={crossNavLabel}
            >
              {crossNavLabel}
            </a>
          )}
        </div>

        {/* Content Wrapper */}
        <main className={className}>
          {children}
        </main>
      </div>
    </div>
  );
};

export default PageLayout;
