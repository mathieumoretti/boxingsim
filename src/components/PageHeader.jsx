/**
 * PageHeader Component
 *
 * Standardized header section for all pages providing:
 * - Consistent title placement (H1)
 * - Action buttons area
 * - Optional subtitle/description
 * - Breadcrumb integration
 *
 * Usage:
 * <PageHeader
 *   title="My Boxers"
 *   subtitle="Manage your fighter roster"
 *   actions={<Button onClick={handleCreate}>+ Add Fighter</Button>}
 * />
 *
 * @param {Object} props
 * @param {string} props.title - Page title (H1)
 * @param {string} [props.subtitle] - Optional subtitle/description
 * @param {React.ReactNode} [props.actions] - Action buttons/slots
 * @param {Array} [props.breadcrumbs] - Breadcrumb items array
 */

import React from 'react';
import Breadcrumbs from './Breadcrumbs';

const PageHeader = ({
  title,
  subtitle,
  actions,
  breadcrumbs
}) => {
  return (
    <header className="space-y-4 mb-6">
      {/* Optional Breadcrumbs */}
      {breadcrumbs && breadcrumbs.length > 0 && (
        <div>
          <Breadcrumbs items={breadcrumbs} />
        </div>
      )}

      {/* Title Section */}
      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
        <div className="space-y-1">
          <h1 className="text-3xl font-bold text-white m-0">{title}</h1>

          {subtitle && (
            <p className="text-base text-white/60 m-0">{subtitle}</p>
          )}
        </div>

        {/* Actions Slot */}
        {actions && (
          <div className="flex items-center gap-3">
            {actions}
          </div>
        )}
      </div>
    </header>
  );
};

export default PageHeader;
