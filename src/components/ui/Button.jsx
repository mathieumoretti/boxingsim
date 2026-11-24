/**
 * Button Component
 *
 * Reusable button with consistent styling, variants, and sizes.
 * Enforces design system consistency across the application.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Primary action button
 * <Button variant="primary" size="md">Schedule Fight</Button>
 *
 * @example
 * // Danger button with onClick handler
 * <Button variant="danger" onClick={handleDelete}>Delete Boxer</Button>
 *
 * @example
 * // Disabled state
 * <Button disabled>Unavailable Action</Button>
 */

import React from 'react';

/**
 * Button component props interface
 *
 * @typedef {Object} ButtonProps
 * @property {'primary' | 'secondary' | 'danger' | 'ghost'} [variant='secondary'] - Button style variant
 * @property {'sm' | 'md' | 'lg'} [size='md'] - Button size
 * @property {boolean} [disabled=false] - Disabled state
 * @property {function} [onClick] - Click event handler
 * @property {React.ReactNode} children - Button content
 * @property {string} [className] - Additional CSS classes
 * @property {'button' | 'submit' | 'reset'} [type='button'] - HTML button type
 */

// Variant base styles with Tailwind utilities
const variantStyles = {
  primary: 'bg-gradient-to-br from-primary to-primary-dark text-white shadow-card hover:-translate-y-0.5 hover:shadow-elevated active:translate-y-0 active:shadow-subtle',
  secondary: 'bg-gradient-to-br from-gray-500 to-gray-600 text-white shadow-subtle hover:bg-gradient-to-br hover:from-gray-600 hover:to-gray-700 hover:-translate-y-0.5 hover:shadow-card active:translate-y-0 active:shadow-subtle',
  danger: 'bg-gradient-to-br from-danger-bright to-danger text-white shadow-card hover:-translate-y-0.5 hover:shadow-elevated active:translate-y-0 active:shadow-subtle',
  ghost: 'bg-transparent text-gray-700 border border-gray-300 shadow-none hover:bg-gray-100 hover:border-gray-600 hover:text-gray-800 hover:-translate-y-0.5 hover:shadow-subtle active:translate-y-0 active:bg-gray-200',
};

// Size padding styles
const sizeStyles = {
  sm: 'px-3 py-2 text-sm',
  md: 'px-5 py-3 text-base',
  lg: 'px-6 py-4 text-lg',
};

export function Button({
  variant = 'secondary',
  size = 'md',
  disabled = false,
  onClick,
  children,
  className = '',
  type = 'button',
}) {
  // Base button styles
  const baseStyles = `
    inline-flex items-center justify-center gap-2
    border-none rounded-md
    font-semibold cursor-pointer transition-all duration-150
    text-decoration-none whitespace-nowrap
    focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2
    disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none
  `;

  return (
    <button
      type={type}
      className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`.trim()}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

export default Button;
