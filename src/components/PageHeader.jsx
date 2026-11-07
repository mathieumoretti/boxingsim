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
 />
 *
 * @param {Object} props
 * @param {string} props.title - Page title (H1)
 * @param {string} [props.subtitle] - Optional subtitle/description
 * @param {React.ReactNode} [props.actions] - Action buttons/slots
 * @param {Array} [props.breadcrumbs] - Breadcrumb items array
 */

import React from 'react';
import Breadcrumbs from './Breadcrumbs';
import './PageHeader.css';

const PageHeader = ({
  title,
  subtitle,
  actions,
  breadcrumbs
}) => {
  return (
    <header className="page-header">
      {/* Optional Breadcrumbs */}
      {breadcrumbs && breadcrumbs.length > 0 && (
        <div className="page-header-breadcrumbs">
          <Breadcrumbs items={breadcrumbs} />
        </div>
      )}

      {/* Title Section */}
      <div className="page-header-content">
        <div className="page-header-text">
          <h1 className="page-title">{title}</h1>

          {subtitle && (
            <p className="page-subtitle">{subtitle}</p>
          )}
        </div>

        {/* Actions Slot */}
        {actions && (
          <div className="page-header-actions">
            {actions}
          </div>
        )}
      </div>
    </header>
  );
};

export default PageHeader;
