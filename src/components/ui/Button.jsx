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
import './Button.css';

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

export function Button({
  variant = 'secondary',
  size = 'md',
  disabled = false,
  onClick,
  children,
  className = '',
  type = 'button',
}) {
  const buttonClass = `btn btn-${variant} btn-${size} ${className}`.trim();

  return (
    <button
      type={type}
      className={buttonClass}
      onClick={onClick}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

export default Button;
