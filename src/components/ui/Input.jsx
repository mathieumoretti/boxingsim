/**
 * Input Component
 *
 * Reusable form input with support for text, email, number, and select types.
 * Includes label, error state, and helper text styling.
 *
 * @see MAT-150 - Shared UI Component Library
 * @see MAT-146 - Design Tokens Foundation
 *
 * @example
 * // Text input with label and helper text
 * <Input
 *   type="text"
 *   label="Boxer Name"
 *   helperText="Enter a unique name for your boxer"
 *   value={name}
 *   onChange={handleChange}
 * />
 *
 * @example
 * // Input with error state
 * <Input
 *   type="email"
 *   label="Email"
 *   error={errors.email}
 *   value={email}
 *   onChange={handleChange}
 * />
 *
 * @example
 * // Select dropdown
 * <Input
 *   type="select"
 *   label="Weight Class"
 *   options={[
 *     { value: "lightweight", label: "Lightweight (135 lbs)" },
 *     { value: "welterweight", label: "Welterweight (147 lbs)" }
 *   ]}
 *   value={weightClass}
 *   onChange={handleChange}
 * />
 */

import React from 'react';
import './Input.css';

/**
 * Input component props interface
 *
 * @typedef {Object} InputProps
 * @property {'text' | 'email' | 'number' | 'select'} [type='text'] - Input type
 * @property {string} [label] - Label text displayed above input
 * @property {string} [error] - Error message displayed below input
 * @property {string} [helperText] - Helper text displayed below input
 * @property {Array<{value: string, label: string}>} [options] - Options for select type
 * @property {string|number} [value] - Input value (controlled component)
 * @property {function} [onChange] - Change event handler
 * @property {boolean} [disabled=false] - Disabled state
 * @property {string} [placeholder] - Placeholder text
 * @property {string} [name] - Input name attribute
 * @property {string} [className] - Additional CSS classes
 */

export function Input({
  type = 'text',
  label,
  error,
  helperText,
  options = [],
  value,
  onChange,
  disabled = false,
  placeholder,
  name,
  className = '',
}) {
  const inputClass = `input-field input-${type} ${className}`.trim();

  // Determine if we have any validation/message content to show
  const hasHelperContent = helperText || error;

  return (
    <div className="input-wrapper">
      {label && (
        <label className="input-label" htmlFor={name}>
          {label}
        </label>
      )}

      {type === 'select' ? (
        <select
          className={inputClass}
          value={value || ''}
          onChange={onChange}
          disabled={disabled}
          name={name}
          id={name}
        >
          <option value="">Select an option</option>
          {options.map((option, index) => (
            <option key={index} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      ) : (
        <input
          type={type}
          className={inputClass}
          value={value}
          onChange={onChange}
          disabled={disabled}
          placeholder={placeholder}
          name={name}
          id={name}
        />
      )}

      {hasHelperContent && (
        <div className={`input-helper ${error ? 'input-helper-error' : ''}`}>
          {error || helperText}
        </div>
      )}
    </div>
  );
}

export default Input;
