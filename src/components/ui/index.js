/**
 * UI Components Barrel Export
 *
 * Centralized exports for all shared UI components.
 * Import from here instead of individual files for cleaner code.
 *
 * @see MAT-150 - Shared UI Component Library
 *
 * @example
 * // Import specific components
 * import { Button, Input, ProgressBar, LoadingSpinner } from '../components/ui';
 *
 * @example
 * // Import all as namespace
 * import * as UI from '../components/ui';
 * UI.Button({ variant: 'primary' })
 */

export { Button } from './Button.jsx';
export { Input } from './Input.jsx';
export { ProgressBar } from './ProgressBar.jsx';
export { LoadingSpinner } from './LoadingSpinner.jsx';
