/**
 * Form Component Suite
 *
 * Reusable form components with consistent styling and structure.
 * Includes Form wrapper, FormGroup for field grouping, and FormRow for side-by-side layouts.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Complete form with all components
 * <Form onSubmit={handleSubmit}>
 *   <FormGroup label="Personal Information">
 *     <FormRow>
 *       <Input name="firstName" label="First Name" />
 *       <Input name="lastName" label="Last Name" />
 *     </FormRow>
 *   </FormGroup>
 *   <Button type="submit" variant="primary">Submit</Button>
 * </Form>
 */

import React from 'react';

// ===== FORM WRAPPER =====

/**
 * Form wrapper component props
 *
 * @typedef {Object} FormProps
 * @property {function} onSubmit - Form submit handler
 * @property {'horizontal' | 'vertical'} [layout='vertical'] - Form layout direction
 * @property {'default' | 'elevated'} [variant='default'] - Form visual style
 * @property {React.ReactNode} children - Form content
 * @property {string} [className] - Additional CSS classes
 * @property {string} [id] - Form ID
 * @property {boolean} [noValidate=false] - Disable HTML5 validation
 */

const formVariantStyles = {
  default: '',
  elevated: 'bg-white p-6 rounded-lg shadow-card border border-navy-border',
};

export function Form({
  onSubmit,
  layout = 'vertical',
  variant = 'default',
  children,
  className = '',
  id,
  noValidate = false,
  ...rest
}) {
  return (
    <form
      id={id}
      onSubmit={onSubmit}
      noValidate={noValidate}
      className={`
        ${layout === 'horizontal' ? 'space-y-6' : 'space-y-4'}
        ${formVariantStyles[variant]}
        ${className}
      `.trim()}
      {...rest}
    >
      {children}
    </form>
  );
}

// ===== FORM GROUP =====

/**
 * FormGroup component for grouping related form fields with optional section header
 *
 * @typedef {Object} FormGroupProps
 * @property {string} [title] - Group title/header text
 * @property {'default' | 'elevated'} [variant='default'] - Group visual style
 * @property {React.ReactNode} [icon] - Optional icon before title
 * @property {React.ReactNode} children - Form fields in group
 * @property {string} [className] - Additional CSS classes
 */

const groupVariantStyles = {
  default: '',
  elevated: 'bg-gray-50 p-5 rounded-lg border border-gray-200',
};

export function FormGroup({
  title,
  variant = 'default',
  icon,
  children,
  className = ''
}) {
  if (!title) {
    return (
      <div className={`space-y-3 ${className}`}>
        {children}
      </div>
    );
  }

  return (
    <fieldset className={`rounded-lg overflow-hidden ${groupVariantStyles[variant]} ${className}`}>
      {/* Header */}
      <div className="flex items-center gap-2 pb-3 mb-3 border-b border-gray-200">
        {icon && (
          <span className="text-xl select-none">{icon}</span>
        )}
        <legend className="font-semibold text-gray-800 text-base">
          {title}
        </legend>
      </div>

      {/* Fields */}
      <div className="space-y-3 pt-1">
        {children}
      </div>
    </fieldset>
  );
}

// ===== FORM ROW =====

/**
 * FormRow component for side-by-side field layout
 *
 * @typedef {Object} FormRowProps
 * @property {number} [gap=4] - Gap between columns (Tailwind spacing scale)
 * @property {'sm' | 'md' | 'lg'} [columns='md'] - Number of columns to split
 * @property {React.ReactNode} children - Form fields in row
 * @property {string} [className] - Additional CSS classes
 */

const columnStyles = {
  sm: 'grid-cols-1 sm:grid-cols-2',
  md: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
  lg: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4',
};

export function FormRow({
  gap = 4,
  columns = 'md',
  children,
  className = ''
}) {
  const gapClass = `gap-${gap}`;

  return (
    <div
      className={`grid ${columnStyles[columns]} ${gapClass} ${className}`}
    >
      {React.Children.map(children, (child) => (
        <div className="w-full">
          {child}
        </div>
      ))}
    </div>
  );
}

// ===== FORM ACTIONS =====

/**
 * FormActions component for form footer buttons
 *
 * @typedef {Object} FormActionsProps
 * @property {'default' | 'right' | 'center'} [align='right'] - Button alignment
 * @property {React.ReactNode} children - Action buttons
 * @property {string} [className] - Additional CSS classes
 */

const alignStyles = {
  default: 'flex gap-3',
  right: 'flex justify-end gap-3',
  center: 'flex justify-center gap-3',
};

export function FormActions({
  align = 'right',
  children,
  className = ''
}) {
  return (
    <div
      className={`mt-6 pt-4 border-t border-gray-200 ${alignStyles[align]} ${className}`}
    >
      {children}
    </div>
  );
}

export default Form;
