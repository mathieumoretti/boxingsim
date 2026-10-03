import '@testing-library/jest-dom';
import axios from 'axios';

// Mock axios globally for all tests
global.axios = axios;
axios.defaults.baseURL = '/api';

// Mock console.error to reduce noise in test output, but still show it
const originalConsoleError = console.error;
console.error = (...args) => {
  if (typeof args[0] === 'string' && args[0].includes('Warning: Each child in a list should have a unique "key"')) {
    return;
  }
  originalConsoleError.apply(console, args);
};

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // Deprecated
    removeListener: jest.fn(), // Deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock ResizeObserver
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserver;
