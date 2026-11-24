/**
 * Input Component
 *
 * Reusable input field with validation states and helpful feedback.
 * Supports text, password, email, number, and other HTML5 input types.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Basic text input
 * <Input label="Name" name="name" placeholder="Enter your name" />
 *
 * @example
 * // Input with error state
 * <Input label="Email" type="email" error="Invalid email address" />
 *
 * @example
 * // Password input with help text
 * <Input type="password" label="Password" help="Must be at least 8 characters" />
 */

import React from 'react';

/**
 * Input component props interface
 *
 * @typedef {Object} InputProps
 * @property {string} [type='text'] - HTML input type
 * @property {string} name - Input field name
 * @property {string} label - Field label
 * @property {string} [value] - Current value (controlled)
 * @property {string} [defaultValue] - Initial value (uncontrolled)
 * @property {string} [placeholder] - Placeholder text
 * @property {boolean} [disabled=false] - Disabled state
 * @property {boolean} [required=false] - Required field indicator
 * @property {string} [error] - Error message to display
 * @property {string} [help] - Help text to display
 * @property {'sm' | 'md'} [size='md'] - Input size
 * @property {function} [onChange] - Change event handler
 * @property {function} [onBlur] - Blur event handler
 * @property {function} [onClick] - Click event handler
 * @property {string} [className] - Additional CSS classes
 */

// Size styles
const sizeStyles = {
  sm: 'text-sm py-2 px-3',
  md: 'text-base py-3 px-4',
};

// State-based border and focus styles
const stateStyles = {
  default: {
    border: 'border-gray-300',
    focus: 'focus:border-primary focus:ring-2 focus:ring-primary/20 focus:ring-offset-0',
    icon: '',
  },
  error: {
    border: 'border-danger-bright',
    focus: 'focus:border-danger focus:ring-2 focus:ring-danger/20 focus:ring-offset-0',
    icon: '⚠️',
  },
  disabled: {
    border: 'border-gray-200',
    focus: '',
    base: 'bg-gray-100 text-gray-400 cursor-not-allowed opacity-60',
  },
};

export function Input({
  type = 'text',
  name,
  label,
  value,
  defaultValue,
  placeholder,
  disabled = false,
  required = false,
  error,
  help,
  size = 'md',
  onChange,
  onBlur,
  onClick,
  className = '',
  ...rest
}) {
  // Determine state styles
  const isDisabled = disabled || rest.readOnly;
  const hasError = !!error && !isDisabled;

  let styles = isDisabled ? stateStyles.disabled : (hasError ? stateStyles.error : stateStyles.default);

  // Construct input wrapper classes
  return (
    <div className={`flex flex-col gap-1.5 ${className}`}>
      {/* Label */}
      {label && (
        <label
          htmlFor={name}
          className="text-sm font-medium text-gray-700"
        >
          {label}
          {required && (
            <span className="text-danger-bright ml-0.5" aria-hidden="true">*</span>
          )}
        </label>
      )}

      {/* Input field wrapper */}
      <div className="relative">
        {/* Error icon (visible when has error) */}
        {hasError && type !== 'password' && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-xs pointer-events-none z-10">
            ⚠️
          </div>
        )}

        {/* Input element */}
        <input
          id={name}
          type={type}
          name={name}
          value={value}
          defaultValue={defaultValue}
          placeholder={placeholder}
          disabled={disabled}
          required={required}
          onChange={onChange}
          onBlur={onBlur}
          onClick={onClick}
          className={`
            w-full rounded-md border transition-all duration-200
            bg-white text-gray-900 shadow-subtle
            placeholder:text-gray-400
            ${sizeStyles[size]}
            ${styles.border}
            ${styles.focus}
            ${isDisabled ? styles.base : ''}
            ${hasError && type !== 'password' ? 'pl-8' : 'pl-4'}
          `.trim()}
          aria-invalid={hasError ? 'true' : undefined}
          aria-describedby={error || help ? `${name}-description` : undefined}
          {...rest}
        />

        {/* Password toggle could go here */}
      </div>

      {/* Error message */}
      {error && (
        <p
          id={`${name}-description`}
          className="text-sm text-danger-bright flex items-center gap-1"
          role="alert"
        >
          <span aria-hidden="true">⚠️</span>
          {error}
        </p>
      )}

      {/* Help text (only when no error) */}
      {help && !error && (
        <p
          id={`${name}-description`}
          className="text-sm text-gray-500"
        >
          {help}
        </p>
      )}
    </div>
  );
}

export default Input;
