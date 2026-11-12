/**
 * FightBadge Component Tests
 *
 * Tests for the FightBadge component covering:
 * - Rendering with fight data
 * - Countdown timer calculation and display
 * - Urgency indicator styling (normal, urgent, critical, immediate)
 * - Pulse animation for near-term fights
 * - Conditional rendering (null when no fight or past fight)
 * - Link navigation to fight details page
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/FightBadge.jsx
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import FightBadge from './FightBadge';

// Mock react-router-dom Link component (must be before imports that use it)
jest.mock('react-router-dom', () => {
  const R = require('react');
  return {
    __esModule: true,
    Link: function({ to, children, className }) {
      return R.createElement('a', { href: to, className: className }, children);
    },
    MemoryRouter: function({ children }) {
      return R.createElement('div', null, children);
    },
  };
});

const { MemoryRouter } = require('react-router-dom');

const mockCurrentGameTime = new Date('2026-11-15T10:00:00Z').toISOString();

const mockFight = {
  fight_id: 'fight-123',
  opponent_name: 'Johnny Bango',
  scheduled_time: new Date('2026-11-15T14:00:00Z').toISOString(),
  status: 'scheduled',
  rounds: 10,
};

const renderWithRouter = (component) => {
  return render(<MemoryRouter>{component}</MemoryRouter>);
};

describe('FightBadge', () => {
  describe('conditional rendering', () => {
    test('returns null when fight is undefined', () => {
      const { container } = renderWithRouter(<FightBadge fight={undefined} currentGameTime={mockCurrentGameTime} />);
      expect(container.querySelector('.fight-badge')).not.toBeInTheDocument();
    });

    test('returns null when fight is null', () => {
      const { container } = renderWithRouter(<FightBadge fight={null} currentGameTime={mockCurrentGameTime} />);
      expect(container.querySelector('.fight-badge')).not.toBeInTheDocument();
    });

    test('returns null when scheduled_time is missing', () => {
      const { container } = renderWithRouter(
        <FightBadge fight={{ ...mockFight, scheduled_time: null }} currentGameTime={mockCurrentGameTime} />
      );
      expect(container.querySelector('.fight-badge')).not.toBeInTheDocument();
    });

    test('returns null when currentGameTime is missing', () => {
      const { container } = renderWithRouter(
        <FightBadge fight={mockFight} currentGameTime={null} />
      );
      expect(container.querySelector('.fight-badge')).not.toBeInTheDocument();
    });

    test('returns null when fight is in the past relative to game time', () => {
      const pastFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-14T10:00:00Z').toISOString(), // Before current game time
      };
      const { container } = renderWithRouter(
        <FightBadge fight={pastFight} currentGameTime={mockCurrentGameTime} />
      );
      expect(container.querySelector('.fight-badge')).not.toBeInTheDocument();
    });

    test('renders badge when fight is scheduled in the future', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText('vs Johnny Bango')).toBeInTheDocument();
    });
  });

  describe('basic rendering', () => {
    test('renders opponent name correctly', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText('vs Johnny Bango')).toBeInTheDocument();
    });

    test('renders fight icon (boxing glove emoji)', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText('🥊')).toBeInTheDocument();
    });

    test('renders timer with stop watch emoji', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      const timer = screen.getByText(/⏱️/);
      expect(timer).toBeInTheDocument();
    });

    test('wraps content in Link to fight details page', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      const link = screen.getByText('vs Johnny Bango').closest('a');
      expect(link).toHaveAttribute('href', '/fight/fight-123');
    });

    test('applies fight-badge base class', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      const badge = screen.getByText('vs Johnny Bango').closest('.fight-badge');
      expect(badge).toBeInTheDocument();
    });
  });

  describe('countdown timer calculation', () => {
    test('displays hours and minutes when fight is more than an hour away', () => {
      const futureFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T16:30:00Z').toISOString(), // 6.5 hours from game time
      };
      renderWithRouter(<FightBadge fight={futureFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText(/⏱️ 6h 30m/)).toBeInTheDocument();
    });

    test('displays minutes and seconds when fight is less than an hour away', () => {
      const nearFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T10:45:30Z').toISOString(), // 45 min 30 sec from game time
      };
      renderWithRouter(<FightBadge fight={nearFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText(/⏱️ 45m 30s/)).toBeInTheDocument();
    });

    test('displays seconds only when fight is less than a minute away', () => {
      const imminentFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T10:00:30Z').toISOString(), // 30 seconds from game time
      };
      renderWithRouter(<FightBadge fight={imminentFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText(/⏱️ 30s/)).toBeInTheDocument();
    });

    test('handles fight exactly at game time boundary', () => {
      const sameTimeFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T10:00:00Z').toISOString(), // Same as game time
      };
      renderWithRouter(<FightBadge fight={sameTimeFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText(/Starting!/)).toBeInTheDocument();
    });
  });

  describe('countdown urgency indicator', () => {
    test('applies countdown-normal when fight is more than 2 hours away', () => {
      const distantFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T14:00:00Z').toISOString(), // 4 hours from game time
      };
      renderWithRouter(<FightBadge fight={distantFight} currentGameTime={mockCurrentGameTime} />);
      const timer = screen.getByText(/⏱️ 4h 0m/).closest('.fight-timer');
      expect(timer).toHaveClass('countdown-normal');
    });

    test('applies countdown-urgent when fight is between 30 minutes and 2 hours away', () => {
      const urgentFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T11:30:00Z').toISOString(), // 1.5 hours from game time
      };
      renderWithRouter(<FightBadge fight={urgentFight} currentGameTime={mockCurrentGameTime} />);
      const timer = screen.getByText(/⏱️ 1h 30m/).closest('.fight-timer');
      expect(timer).toHaveClass('countdown-urgent');
    });

    test('applies countdown-critical when fight is less than 30 minutes away', () => {
      const criticalFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T10:15:00Z').toISOString(), // 15 minutes from game time
      };
      renderWithRouter(<FightBadge fight={criticalFight} currentGameTime={mockCurrentGameTime} />);
      const timer = screen.getByText(/⏱️ 15m 0s/).closest('.fight-timer');
      expect(timer).toHaveClass('countdown-critical');
    });

    test('applies countdown-immediate when fight time has passed or equals game time', () => {
      const immediateFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T10:00:00Z').toISOString(), // Same as game time
      };
      renderWithRouter(<FightBadge fight={immediateFight} currentGameTime={mockCurrentGameTime} />);
      const timer = screen.getByText(/Starting!/).closest('.fight-timer');
      expect(timer).toHaveClass('countdown-immediate');
    });
  });

  describe('pulse animation for near-term fights', () => {
    test('does not apply pulse-animation when fight is more than 2 hours away', () => {
      const distantFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T14:00:00Z').toISOString(), // 4 hours from game time
      };
      renderWithRouter(<FightBadge fight={distantFight} currentGameTime={mockCurrentGameTime} />);
      const badge = screen.getByText('vs Johnny Bango').closest('.fight-badge');
      expect(badge).not.toHaveClass('pulse-animation');
    });

    test('applies pulse-animation when fight is within 2 hours', () => {
      const nearFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T11:00:00Z').toISOString(), // 1 hour from game time
      };
      renderWithRouter(<FightBadge fight={nearFight} currentGameTime={mockCurrentGameTime} />);
      const badge = screen.getByText('vs Johnny Bango').closest('.fight-badge');
      expect(badge).toHaveClass('pulse-animation');
    });

    test('does not apply pulse-animation when fight is exactly 2 hours away (boundary case)', () => {
      const exactTwoHoursFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T12:00:00Z').toISOString(), // Exactly 2 hours
      };
      renderWithRouter(<FightBadge fight={exactTwoHoursFight} currentGameTime={mockCurrentGameTime} />);
      const badge = screen.getByText('vs Johnny Bango').closest('.fight-badge');
      expect(badge).not.toHaveClass('pulse-animation');
    });

    test('does not apply pulse-animation when fight is more than 2 hours away by one minute', () => {
      const slightlyMoreFight = {
        ...mockFight,
        scheduled_time: new Date('2026-11-15T12:01:00Z').toISOString(), // 2h 1m from game time
      };
      renderWithRouter(<FightBadge fight={slightlyMoreFight} currentGameTime={mockCurrentGameTime} />);
      const badge = screen.getByText('vs Johnny Bango').closest('.fight-badge');
      expect(badge).not.toHaveClass('pulse-animation');
    });
  });

  describe('tooltip and accessibility', () => {
    test('displays formatted date in timer tooltip', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      const timer = screen.getByText(/⏱️/).closest('.fight-timer');
      // The title should contain month, day, year, and time (e.g., "Nov 15, 2026, 9 AM")
      const title = timer.getAttribute('title');
      expect(title).toContain(',');
      expect(title).toMatch(/\d{4}/); // Contains year
    });

    test('displays rounds count in fight icon tooltip', () => {
      renderWithRouter(<FightBadge fight={mockFight} currentGameTime={mockCurrentGameTime} />);
      const icon = screen.getByText('🥊').closest('.fight-icon');
      expect(icon.getAttribute('title')).toContain('10 rounds');
    });

    test('handles different round counts in tooltip', () => {
      const titleFight = {
        ...mockFight,
        rounds: 12,
      };
      renderWithRouter(<FightBadge fight={titleFight} currentGameTime={mockCurrentGameTime} />);
      const icon = screen.getByText('🥊').closest('.fight-icon');
      expect(icon.getAttribute('title')).toContain('12 rounds');
    });
  });

  describe('typical use cases', () => {
    test('renders like a scheduled fight for tomorrow (3 PM)', () => {
      const today = new Date('2026-11-15T10:00:00Z').toISOString();
      const tomorrowFight = {
        fight_id: 'title-fight-456',
        opponent_name: 'Champion Mike Tyson',
        scheduled_time: new Date('2026-11-16T15:00:00Z').toISOString(), // Tomorrow 3 PM
        status: 'scheduled',
        rounds: 12,
      };
      renderWithRouter(<FightBadge fight={tomorrowFight} currentGameTime={today} />);

      expect(screen.getByText('vs Champion Mike Tyson')).toBeInTheDocument();
      expect(screen.getByText(/⏱️ 29h 0m/)).toBeInTheDocument();
      const badge = screen.getByText('vs Champion Mike Tyson').closest('.fight-badge');
      expect(badge).not.toHaveClass('pulse-animation');
    });

    test('renders like an upcoming fight today (2 hours from now)', () => {
      const gameTime = new Date('2026-11-15T14:00:00Z').toISOString();
      const todaysFight = {
        fight_id: 'undercard-789',
        opponent_name: 'Fast Freddie',
        scheduled_time: new Date('2026-11-15T16:00:00Z').toISOString(), // 2 hours from game time
        status: 'scheduled',
        rounds: 8,
      };
      renderWithRouter(<FightBadge fight={todaysFight} currentGameTime={gameTime} />);

      expect(screen.getByText('vs Fast Freddie')).toBeInTheDocument();
      expect(screen.getByText(/⏱️ 2h 0m/)).toBeInTheDocument();
    });

    test('renders like a critical upcoming fight (15 minutes)', () => {
      const gameTime = new Date('2026-11-15T20:45:00Z').toISOString();
      const imminentFight = {
        fight_id: 'rematch-321',
        opponent_name: 'Rival Ray',
        scheduled_time: new Date('2026-11-15T21:00:00Z').toISOString(), // 15 minutes from game time
        status: 'scheduled',
        rounds: 10,
      };
      renderWithRouter(<FightBadge fight={imminentFight} currentGameTime={gameTime} />);

      expect(screen.getByText('vs Rival Ray')).toBeInTheDocument();
      expect(screen.getByText(/⏱️ 15m 0s/)).toBeInTheDocument();
      const badge = screen.getByText('vs Rival Ray').closest('.fight-badge');
      expect(badge).toHaveClass('pulse-animation');
    });

    test('hides badge when fight time has passed', () => {
      const gameTime = new Date('2026-11-15T15:00:00Z').toISOString();
      const pastFight = {
        fight_id: 'missed-999',
        opponent_name: 'Missed Opponent',
        scheduled_time: new Date('2026-11-15T14:00:00Z').toISOString(), // 1 hour before game time
        status: 'scheduled',
        rounds: 10,
      };
      const { container } = renderWithRouter(<FightBadge fight={pastFight} currentGameTime={gameTime} />);
      // MemoryRouter wraps in div, so check that the only child is the wrapper with null inner content
      expect(container.firstChild).not.toHaveTextContent('Missed Opponent');
    });
  });

  describe('edge cases', () => {
    test('handles very long opponent names without breaking layout', () => {
      const longNameFight = {
        ...mockFight,
        opponent_name: 'The Invincible Unbeatable Champion World Famous Heavyweight King Kong Destroyer',
      };
      renderWithRouter(<FightBadge fight={longNameFight} currentGameTime={mockCurrentGameTime} />);
      expect(
        screen.getByText('vs The Invincible Unbeatable Champion World Famous Heavyweight King Kong Destroyer')
      ).toBeInTheDocument();
    });

    test('handles opponent name with special characters', () => {
      const specialCharFight = {
        ...mockFight,
        opponent_name: 'El "Campeón" García Jr.',
      };
      renderWithRouter(<FightBadge fight={specialCharFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText('vs El "Campeón" García Jr.')).toBeInTheDocument();
    });

    test('handles opponent name with emoji', () => {
      const emojiNameFight = {
        ...mockFight,
        opponent_name: 'Iron Fist 🥊 Williams',
      };
      renderWithRouter(<FightBadge fight={emojiNameFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText('vs Iron Fist 🥊 Williams')).toBeInTheDocument();
    });

    test('handles zero rounds gracefully', () => {
      const zeroRoundsFight = {
        ...mockFight,
        rounds: 0,
      };
      renderWithRouter(<FightBadge fight={zeroRoundsFight} currentGameTime={mockCurrentGameTime} />);
      const icon = screen.getByText('🥊').closest('.fight-icon');
      expect(icon.getAttribute('title')).toContain('0 rounds');
    });

    test('handles undefined status gracefully', () => {
      const noStatusFight = {
        fight_id: 'no-status',
        opponent_name: 'Test Opponent',
        scheduled_time: new Date('2026-11-15T12:00:00Z').toISOString(),
        rounds: 10,
      };
      renderWithRouter(<FightBadge fight={noStatusFight} currentGameTime={mockCurrentGameTime} />);
      expect(screen.getByText('vs Test Opponent')).toBeInTheDocument();
    });
  });
});
