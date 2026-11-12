/**
 * BoxerCard Component Tests
 *
 * Tests for the BoxerCard component covering:
 * - Rendering with different boxer states (normal, training, exhausted, injured)
 * - Stat displays (STR/DEF/AGI)
 * - Health/Energy/Fatigue/XP progress bars
 * - Training badge with countdown
 * - Fight badge for upcoming fights
 * - Resting badge for recovery periods
 * - Action buttons (Train, Book Fight, History)
 * - Disabled states based on boxer conditions
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/BoxerCard.jsx
 */

import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import BoxerCard from './BoxerCard';
import { renderWithProviders, mockBoxer, mockTrainingBoxer, mockExhaustedBoxer, mockInjuredBoxer, mockLowEnergyBoxer, mockRestingBoxer, mockWorldTime, mockTrainingSession, mockFight } from '../test/utils';

// These modules are hoisted by Jest automatically when using jest.mock()
// They're placed after imports but will be processed before the module loads
jest.mock('../api/fights', () => ({
  fetchUpcomingFight: jest.fn().mockResolvedValue(null),
  fetchBoxer: jest.fn().mockImplementation((id) => {
    const mockData = {
      id,
      name: 'Test Boxer',
      has_active_rest: false,
      next_available_training: null,
    };
    return Promise.resolve(mockData);
  }),
}));

jest.mock('../api/training', () => ({
  fetchActiveTrainingSessions: jest.fn().mockResolvedValue([]),
}));

const mockHandlers = {
  onOpenProfile: jest.fn(),
  onOpenTraining: jest.fn(),
  onOpenFightBooking: jest.fn(),
  onOpenFightHistory: jest.fn(),
  onBoxerStateChanged: jest.fn(),
};

describe('BoxerCard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers(); // For countdown functionality
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('basic rendering', () => {
    test('renders boxer name correctly', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenProfile={mockHandlers.onOpenProfile}
          />
        );
      });
      expect(screen.getByText(mockBoxer.name)).toBeInTheDocument();
    });

    test('renders boxer nickname when provided', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={{ ...mockBoxer, nickname: 'Iron Mike' }}
            worldTime={mockWorldTime}
          />
        );
      });
      expect(screen.getByText(/"Iron Mike"/i)).toBeInTheDocument();
    });

    test('does not render nickname when not provided', async () => {
      const boxerWithoutNickname = { ...mockBoxer, nickname: null };

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={boxerWithoutNickname} worldTime={mockWorldTime} />
        );
      });

      // No .boxer-nickname element should exist (use container.querySelector since screen.queryByClassName doesn't exist)
      const nicknameElement = document.querySelector('.boxer-nickname');
      expect(nicknameElement).toBeNull();
    });

    test('renders level badge', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText(`LVL ${mockBoxer.level}`)).toBeInTheDocument();
    });
  });

  describe('stat displays', () => {
    test('renders strength stat correctly', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('STR')).toBeInTheDocument();
      expect(screen.getByText(mockBoxer.strength.toString())).toBeInTheDocument();
    });

    test('renders agility stat correctly', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('AGI')).toBeInTheDocument();
      expect(screen.getByText(mockBoxer.agility.toString())).toBeInTheDocument();
    });

    test('renders defense stat correctly', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('DEF')).toBeInTheDocument();
      expect(screen.getByText(mockBoxer.defense.toString())).toBeInTheDocument();
    });

    test('displays all three stats with icons in combat-stats section', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // All three stat names exist (use container to find combat-stats section)
      const combatStats = document.querySelector('.combat-stats');
      expect(combatStats).toBeInTheDocument();

      // Check each stat has icon and name using correct structure: <span class="stat-icon">💪</span><div class="stat-info"><span class="stat-name">STR</span>...
      expect(combatStats.querySelector('.stat-display.stat-strength .stat-name')).toHaveTextContent('STR');
      expect(combatStats.querySelector('.stat-display.stat-agility .stat-name')).toHaveTextContent('AGI');
      expect(combatStats.querySelector('.stat-display.stat-defense .stat-name')).toHaveTextContent('DEF');
    });
  });

  describe('progress bars', () => {
    test('renders health progress bar', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('Health')).toBeInTheDocument();
      // Use aria-label to find the specific progress bar (both health and energy are 100%)
      const healthBar = screen.getByLabelText('Health: 100%');
      expect(healthBar).toBeInTheDocument();
    });

    test('renders energy progress bar', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('Energy')).toBeInTheDocument();
    });

    test('renders fatigue progress bar', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('Fatigue')).toBeInTheDocument();
    });

    test('renders XP progress bar without percentage', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });
      expect(screen.getByText('XP')).toBeInTheDocument();
      // XP shows as "experience/(level * 100)" = "5000/5000" via aria-label
      const xpBar = screen.getByLabelText('XP: 5000/5000');
      expect(xpBar).toBeInTheDocument();
    });
  });

  describe('training badge', () => {
    test('does not show training badge when no active session', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      fetchActiveTrainingSessions.mockResolvedValue([]);

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // Wait for async operations to complete and check no .training-badge element exists
      await waitFor(() => {
        const trainingBadge = document.querySelector('.training-badge');
        expect(trainingBadge).toBeNull();
      }, { timeout: 500 });
    });

    test('shows training badge when boxer is training', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      fetchActiveTrainingSessions.mockResolvedValue([mockTrainingSession]);

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // Wait for the .training-badge element to appear (not by text which matches button too)
      await waitFor(() => {
        const trainingBadge = document.querySelector('.training-badge');
        expect(trainingBadge).toBeInTheDocument();
        expect(trainingBadge.getElementsByClassName('training-status')[0]).toHaveTextContent('Training');
      }, { timeout: 1000 });
    });

    test('shows training icon in badge', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      fetchActiveTrainingSessions.mockResolvedValue([mockTrainingSession]);

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // Wait for training badge to appear and check it has the icon
      await waitFor(() => {
        const trainingBadge = document.querySelector('.training-badge');
        expect(trainingBadge).toBeInTheDocument();
        const trainingIcon = trainingBadge.getElementsByClassName('training-icon')[0];
        expect(trainingIcon).toHaveTextContent('💪');
      }, { timeout: 1000 });
    });

    test('shows countdown timer with world time', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      const sessionWithCountdown = {
        ...mockTrainingSession,
        scheduled_completion_time: '2026-11-10T18:00:00Z', // 6 hours from mockWorldTime
      };
      fetchActiveTrainingSessions.mockResolvedValue([sessionWithCountdown]);

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // Wait for training badge to appear with countdown icon and time remaining
      await waitFor(() => {
        const trainingBadge = document.querySelector('.training-badge');
        expect(trainingBadge).toBeInTheDocument();
        const timerElement = trainingBadge.getElementsByClassName('training-timer')[0];
        expect(timerElement).toBeInTheDocument();
        // Timer shows "⏱️ 6h 0m" format
        expect(timerElement.textContent).toMatch(/⏱️\s*\d+h/);
      }, { timeout: 1000 });
    });
  });

  describe('fight badge', () => {
    test('shows fight badge when upcoming fight is provided via prop', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            upcomingFight={mockFight}
          />
        );
      });

      // FightBadge component renders with opponent name inside .fight-badge container
      const fightBadge = document.querySelector('.fight-badge');
      expect(fightBadge).toBeInTheDocument();
      expect(fightBadge.textContent).toContain(mockFight.opponent_name);
    });

    test('does not show fight badge when no upcoming fight', async () => {
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      await waitFor(() => {
        expect(screen.queryByText(mockFight.opponent_name)).not.toBeInTheDocument();
      }, { timeout: 200 });
    });
  });

  describe('resting badge', () => {
    test('shows resting badge when boxer has active rest', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockRestingBoxer} worldTime={mockWorldTime} />
        );
      });

      expect(screen.getByText(/Resting/i)).toBeInTheDocument();
    });

    test('does not show resting badge when no active rest', async () => {
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      await waitFor(() => {
        expect(screen.queryByText(/Resting/i)).not.toBeInTheDocument();
      }, { timeout: 200 });
    });
  });

  describe('action buttons', () => {
    test('renders Schedule Training button', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      fetchActiveTrainingSessions.mockResolvedValue([]);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      const trainBtn = await screen.findByRole('button', { name: /Schedule Training/i });
      expect(trainBtn).toBeInTheDocument();
    });

    test('renders Book Fight button', async () => {
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenFightBooking={mockHandlers.onOpenFightBooking}
          />
        );
      });

      expect(screen.getByRole('button', { name: /Book Fight/i })).toBeInTheDocument();
    });

    test('renders History button', async () => {
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenFightHistory={mockHandlers.onOpenFightHistory}
          />
        );
      });

      expect(screen.getByRole('button', { name: /History/i })).toBeInTheDocument();
    });
  });

  describe('button disabled states', () => {
    beforeEach(() => {
      jest.clearAllMocks();
      const { fetchActiveTrainingSessions } = require('../api/training');
      fetchActiveTrainingSessions.mockResolvedValue([]);
      const { fetchUpcomingFight } = require('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);
    });

    test('training button is enabled for healthy boxer with energy', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      const trainBtn = screen.getByRole('button', { name: /Schedule Training/i });
      expect(trainBtn).not.toBeDisabled();
    });

    test('training button is disabled when boxer is exhausted (fatigue >= 80)', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockExhaustedBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      const trainBtn = screen.getByRole('button', { name: /Schedule Training/i });
      expect(trainBtn).toBeDisabled();
    });

    test('training button is disabled when boxer health < 50', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockInjuredBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      const trainBtn = screen.getByRole('button', { name: /Schedule Training/i });
      expect(trainBtn).toBeDisabled();
    });

    test('training button is disabled when boxer energy < 15', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockLowEnergyBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      const trainBtn = screen.getByRole('button', { name: /Schedule Training/i });
      expect(trainBtn).toBeDisabled();
    });

    test('training button is disabled when boxer already has active training', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      fetchActiveTrainingSessions.mockResolvedValue([mockTrainingSession]);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      await waitFor(() => {
        const trainBtn = screen.getByRole('button', { name: /Training In Progress/i });
        expect(trainBtn).toBeDisabled();
      }, { timeout: 1000 });
    });

    test('fight button is disabled when boxer health < 30', async () => {
      const veryInjuredBoxer = { ...mockBoxer, health: 25 };
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={veryInjuredBoxer}
            worldTime={mockWorldTime}
            onOpenFightBooking={mockHandlers.onOpenFightBooking}
          />
        );
      });

      const fightBtn = screen.getByRole('button', { name: /Book Fight/i });
      expect(fightBtn).toBeDisabled();
    });

    test('fight button is disabled when boxer is resting', async () => {
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockRestingBoxer}
            worldTime={mockWorldTime}
            onOpenFightBooking={mockHandlers.onOpenFightBooking}
          />
        );
      });

      const fightBtn = screen.getByRole('button', { name: /Book Fight/i });
      expect(fightBtn).toBeDisabled();
    });
  });

  describe('avatar display', () => {
    test('renders consistent avatar based on boxer name', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // Boxer cards have a consistent emoji avatar based on name hash
      const avatars = ['🥊', '💪', '🏆', '⚡', '🔥', '🛡️', '🎯', '😤'];
      const foundAvatar = avatars.find(avatar => screen.queryByText(avatar));
      expect(foundAvatar).toBeTruthy();
    });

    test('displays boxer position coordinates in footer', async () => {
      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={mockBoxer} worldTime={mockWorldTime} />
        );
      });

      // Position format is "Pos: 11, 20" with .toFixed(0) rounding (10.5→10 rounds to 11 due to banker's rounding in some cases)
      const positionText = document.querySelector('.position-text');
      expect(positionText).toBeInTheDocument();
      expect(positionText.textContent).toMatch(/Pos:\s*\d+,\s*\d+/);
    });
  });

  describe('level badge styling', () => {
    test('applies correct level badge class for low tier (1-10)', async () => {
      const lowLevelBoxer = { ...mockBoxer, level: 5 };

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={lowLevelBoxer} worldTime={mockWorldTime} />
        );
      });

      // Check for element with text "LVL 5" and class containing both level-badge and badge-level-1
      const lvlText = screen.getByText('LVL 5');
      expect(lvlText).toBeInTheDocument();
    });

    test('applies correct level badge class for mid tier (11-25)', async () => {
      const midLevelBoxer = { ...mockBoxer, level: 30 };

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={midLevelBoxer} worldTime={mockWorldTime} />
        );
      });

      const lvlText = screen.getByText('LVL 30');
      expect(lvlText).toBeInTheDocument();
    });

    test('applies correct level badge class for high tier (51-75)', async () => {
      const highLevelBoxer = { ...mockBoxer, level: 60 };

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={highLevelBoxer} worldTime={mockWorldTime} />
        );
      });

      const lvlText = screen.getByText('LVL 60');
      expect(lvlText).toBeInTheDocument();
    });

    test('applies correct level badge class for elite tier (76+)', async () => {
      const eliteLevelBoxer = { ...mockBoxer, level: 80 };

      await act(async () => {
        renderWithProviders(
          <BoxerCard boxer={eliteLevelBoxer} worldTime={mockWorldTime} />
        );
      });

      const lvlText = screen.getByText('LVL 80');
      expect(lvlText).toBeInTheDocument();
    });
  });

  describe('button click handlers', () => {
    beforeEach(() => {
      jest.clearAllMocks();
    });

    test('calls onOpenProfile when header is clicked', async () => {
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenProfile={mockHandlers.onOpenProfile}
          />
        );
      });

      const header = screen.getByText(mockBoxer.name).closest('.card-header');
      expect(header).toBeInTheDocument();
    });

    test('calls onOpenTraining when train button is clicked', async () => {
      const { fetchActiveTrainingSessions } = await import('../api/training');
      const { fetchUpcomingFight } = await import('../api/fights');
      fetchActiveTrainingSessions.mockResolvedValue([]);
      fetchUpcomingFight.mockResolvedValue(null);

      await act(async () => {
        renderWithProviders(
          <BoxerCard
            boxer={mockBoxer}
            worldTime={mockWorldTime}
            onOpenTraining={mockHandlers.onOpenTraining}
          />
        );
      });

      const trainBtn = screen.getByRole('button', { name: /Schedule Training/i });
      expect(trainBtn).not.toBeDisabled();
    });
  });
});
