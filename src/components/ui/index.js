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
export { StatDisplay } from './StatDisplay.jsx';
export { Badge } from './Badge.jsx';
export { Alert } from './Alert.jsx';

// MAT-154 Phase 2 - Card & Form Components
export { Card, CardHeader, CardBody, CardFooter } from './Card.jsx';
export { Form, FormGroup, FormRow } from './Form.jsx';
