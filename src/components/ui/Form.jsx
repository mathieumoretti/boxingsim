/**
 * Form Component System (MAT-154 Phase 2)
 *
 * Shared form structure with FormGroup and FormRow components.
 * Replaces duplicated form implementations in TrainingScheduler.
 *
 * Usage:
 *   <Form onSubmit={handleSubmit} error={formError}>
 *     <FormRow>
 *       <FormGroup label="Select Training Type" error={errors.type}>
 *         <select name="type">...</select>
 *       </FormGroup>
 *       <FormGroup label="Duration (hours)">
 *         <input type="number" name="duration" min="1" max="8" />
 *       </FormGroup>
 *     </FormRow>
 *   </Form>
 *
 * @param {Object} children - Form content
 * @param {Function} onSubmit - Form submit handler
 * @param {string|null} error - Form-level error message
 * @param {string} className - Additional CSS classes
 */

import './Form.css';

export function Form({ children, onSubmit, error, className = '' }) {
  return (
    <form
      className={`form-container ${className}`}
      onSubmit={onSubmit}
    >
      {error && (
        <div className="form-error" style={{
          background: 'rgba(239, 68, 68, 0.1)',
          border: '1px solid var(--color-danger)',
          borderRadius: 'var(--radius-md)',
          padding: 'var(--space-3)',
          marginBottom: 'var(--space-4)',
          color: 'var(--color-danger-bright)',
          fontSize: 'var(--font-size-sm)'
        }}>
          {error}
        </div>
      )}
      {children}
    </form>
  );
}

/**
 * FormGroup - Individual form field with label and optional error
 *
 * @param {Object} children - Input/select/textarea element
 * @param {string} label - Field label text
 * @param {string|null} error - Field error message
 * @param {string} className - Additional CSS classes
 */
export function FormGroup({ children, label, error, className = '' }) {
  return (
    <div className={`form-group ${error ? 'error' : ''} ${className}`}>
      {label && (
        <label htmlFor={children?.props?.name}>
          {label}
          {children?.props?.required && <span style={{ color: 'var(--color-danger)' }}> *</span>}
        </label>
      )}
      {React.cloneElement(children, {
        ...children.props,
        name: children.props.name || '',
        id: children.props.id || children.props.name
      })}
      {error && (
        <span className="error-message">
          {error}
        </span>
      )}
    </div>
  );
}

/**
 * FormRow - Horizontal row container for multiple form groups
 *
 * @param {Object} children - FormGroup components
 * @param {string} className - Additional CSS classes
 */
export function FormRow({ children, className = '' }) {
  return (
    <div className={`form-row ${className}`}>
      {children}
    </div>
  );
}
