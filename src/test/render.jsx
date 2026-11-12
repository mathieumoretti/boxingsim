/**
 * Test Render Utilities
 *
 * Provides helper functions for rendering components with:
 * - React Router setup
 * - Mock providers (context)
 * - Custom wrapper components
 *
 * @see MAT-169 - Frontend Component Tests
 */

import React from 'react';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';

/**
 * Default wrapper with Router for navigation-dependent components
 */
function DefaultProviders({ children }) {
  return (
    <BrowserRouter>
      {children}
    </BrowserRouter>
  );
}

/**
 * Custom render function that includes default providers
 *
 * @param {React.ReactNode} ui - Component to render
 * @param {Object} options - Render options
 * @param {Object} [options.wrapper] - Custom wrapper component
 * @returns {Object} Render result with testing utilities
 *
 * @example
 * // Simple render
 * const { getByText } = renderWithProviders(<MyComponent />);
 *
 * @example
 * // With custom provider
 * const { getByText } = renderWithProviders(<MyComponent />, {
 *   wrapper: ({ children }) => (
 *     <MyContextProvider>
 *       {children}
 *     </MyContextProvider>
 *   )
 * });
 */
function renderWithProviders(ui, { wrapper: CustomProviders = DefaultProviders, ...options } = {}) {
  return render(ui, {
    wrapper: CustomProviders,
    ...options
  });
}

/**
 * Mock data for testing boxer-related components
 */
export const mockBoxer = {
  id: 1,
  name: 'Mike Tyson',
  nickname: 'Iron Mike',
  strength: 85,
  defense: 70,
  agility: 90,
  health: 100,
  energy: 100,
  fatigue_score: 20,
  experience: 5000,
  level: 50,
  wins: 50,
  losses: 5,
  draws: 0,
  has_active_rest: false,
  next_available_training: null,
  forced_rest_until: null,
  position_x: 10.5,
  position_y: 20.3,
};

export const mockTrainingBoxer = {
  ...mockBoxer,
  fatigue_score: 45,
};

export const mockExhaustedBoxer = {
  ...mockBoxer,
  fatigue_score: 85,
};

export const mockInjuredBoxer = {
  ...mockBoxer,
  health: 25,
};

export const mockLowEnergyBoxer = {
  ...mockBoxer,
  energy: 10,
};

export const mockRestingBoxer = {
  ...mockBoxer,
  has_active_rest: true,
  next_available_training: '2026-11-15T14:30:00Z',
};

/**
 * Mock data for fight-related components
 */
export const mockFight = {
  id: 1,
  boxer1_id: 1,
  boxer2_id: 2,
  opponent_name: 'Evander Holyfield',
  scheduled_time: '2026-11-20T19:00:00Z',
  status: 'scheduled',
  rounds: 12,
};

export const mockCompletedFight = {
  ...mockFight,
  status: 'completed',
  winner_id: 1,
  rounds_played: 8,
};

export const mockCancelledFight = {
  ...mockFight,
  status: 'cancelled',
};

/**
 * Mock data for world clock / game time
 */
export const mockWorldTime = {
  current_game_time: '2026-11-10T12:00:00Z',
  server_time_offset_seconds: 0,
};

/**
 * Mock data for training sessions
 */
export const mockTrainingSession = {
  id: 1,
  boxer_id: 1,
  training_type_id: 1,
  status: 'pending',
  scheduled_start_time: '2026-11-10T12:00:00Z',
  scheduled_completion_time: '2026-11-10T18:00:00Z',
  stat_boost_type: 'strength',
  stat_boost_value: 5,
};

export const mockCompletedTrainingSession = {
  ...mockTrainingSession,
  status: 'completed',
  completed_at: '2026-11-10T18:00:00Z',
};

export default renderWithProviders;
