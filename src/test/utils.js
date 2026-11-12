/**
 * Test Utilities Index
 *
 * Central export point for all test utilities:
 * - Render functions with providers
 * - Mock data objects
 * - Mock API functions
 * - Custom testing helpers
 *
 * @see MAT-169 - Frontend Component Tests
 */

export { default as renderWithProviders, mockBoxer, mockTrainingBoxer, mockExhaustedBoxer, mockInjuredBoxer, mockLowEnergyBoxer, mockRestingBoxer, mockFight, mockCompletedFight, mockCancelledFight, mockWorldTime, mockTrainingSession, mockCompletedTrainingSession } from './render';

export { default as mockApi, resetMocks } from './mockApi';

/**
 * Wait for async operations to complete
 * @param {number} ms - Milliseconds to wait
 */
export const waitFor = (ms = 100) => new Promise(resolve => setTimeout(resolve, ms));

/**
 * Create a mock event object
 * @param {string} type - Event type
 * @param {Object} properties - Event properties
 */
export function createMockEvent(type = 'click', properties = {}) {
  return {
    type,
    preventDefault: jest.fn(),
    stopPropagation: jest.fn(),
    ...properties,
  };
}

/**
 * Get all elements by data-testid attribute
 * @param {HTMLElement} container - DOM container
 * @param {string} testId - Test ID to search for
 */
export function getAllByTestId(container, testId) {
  return container.querySelectorAll(`[data-testid="${testId}"]`);
}
