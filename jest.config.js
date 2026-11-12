/**
 * Jest Configuration
 *
 * Configured for React Testing Library with:
 * - Babel transpilation (reuses webpack presets)
 * - CSS module handling
 * - Node.js environment for API tests, jsdom for component tests
 *
 * @see MAT-169 - Frontend Component Tests
 */

module.exports = {
  testEnvironment: 'jsdom',

  testMatch: [
    '**/?(*.)+(test).[jt]s?(x)',
    '**/__tests__/*.[jt]s?(x)'
  ],

  moduleFileExtensions: [
    'js',
    'jsx',
    'json'
  ],

  moduleNameMapper: {
    '^@api/(.*)$': '<rootDir>/src/api/$1',
    '^@components/(.*)$': '<rootDir>/src/components/$1',
    '^@hooks/(.*)$': '<rootDir>/src/hooks/$1',
    '\\.(css|less|scss)$': 'identity-obj-proxy'
  },

  setupFilesAfterEnv: [
    '<rootDir>/src/test/setup.js'
  ],

  transform: {
    '^.+\\.[jt]sx?$': ['babel-jest', {
      presets: ['@babel/preset-env', '@babel/preset-react']
    }]
  },

  // Collect coverage for src files only
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/**/*.test.{js,jsx}',
    '!src/test/**'
  ],

  // Coverage thresholds (can be adjusted)
  coverageThreshold: {
    global: {
      branches: 50,
      functions: 50,
      lines: 50,
      statements: 50
    }
  }
};
