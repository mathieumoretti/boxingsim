/**
 * Jest Test Setup File
 *
 * Runs before each test file. Configures:
 * - @testing-library/jest-dom custom matchers
 * - Global mocks for fetch, localStorage, etc.
 * - Console filtering during tests
 *
 * @see MAT-169 - Frontend Component Tests
 */

import '@testing-library/jest-dom';

// Suppress act warnings during tests (common with async useEffect hooks in components)
const originalError = console.error;
const originalWarn = console.warn;

beforeAll(() => {
  // Silence "not wrapped in act" and "not configured to support act" warnings
  console.error = (...args) => {
    if (args[0]?.includes('An update to') && args[0]?.includes('was not wrapped in act')) {
      return;
    }
    if (args[0]?.includes('not configured to support act')) {
      return;
    }
    originalError.apply(console, args);
  };

  console.warn = (...args) => {
    // Suppress React Router deprecation warnings
    if (args[0]?.includes('React Router Future Flag Warning')) {
      return;
    }
    originalWarn.apply(console, args);
  };
});

afterAll(() => {
  console.error = originalError;
  console.warn = originalWarn;
});

// Mock fetch globally if needed
global.fetch = jest.fn().mockImplementation(() => Promise.resolve({
  json: () => Promise.resolve([]),
}));
