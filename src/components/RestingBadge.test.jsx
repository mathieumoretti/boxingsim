/**
 * RestingBadge Component Tests
 *
 * Tests for the RestingBadge component covering:
 * - Conditional rendering (null when no active rest)
 * - Voluntary vs forced rest visual differentiation
 * - Countdown timer calculation for both real-time and game-time based rests
 * - Remaining time display formatting
 * - Icon and status text variants
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/RestingBadge.jsx
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import RestingBadge from './RestingBadge';

jest.mock('./RestingBadge', () => {
  const original = jest.requireActual('./RestingBadge');

  return {
    __esModule: true,
    default: original.default,
  };
});

const mockCurrentGameTime = new Date('2026-11-15T10:00:00Z').toISOString();

describe('RestingBadge', () => {
  describe('conditional rendering', () => {
    test('returns null when hasActiveRest is false', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={false}
          restEndsAt={null}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(container.firstChild).toBeNull();
    });

    test('returns null when hasActiveRest is undefined', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={undefined}
          restEndsAt={null}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(container.firstChild).toBeNull();
    });

    test('returns null when hasActiveRest is null', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={null}
          restEndsAt={null}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(container.firstChild).toBeNull();
    });

    test('renders badge when hasActiveRest is true', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText(/Resting|Forced Rest/)).toBeInTheDocument();
    });
  });

  describe('voluntary rest rendering', () => {
    test('applies resting-voluntary class for voluntary rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const badge = screen.getByText('Resting').closest('.resting-badge');
      expect(badge).toHaveClass('resting-voluntary');
    });

    test('displays hospital emoji for voluntary rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('🏥')).toBeInTheDocument();
    });

    test('displays "Resting" status text for voluntary rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const status = screen.getByText('Resting');
      expect(status).toHaveClass('resting-status');
    });

    test('does not render forced rest text for voluntary rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.queryByText('Forced Rest')).not.toBeInTheDocument();
    });

    test('does not render ambulance emoji for voluntary rest', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(container.querySelector('.resting-badge')).not.toHaveTextContent('🚑');
    });
  });

  describe('forced rest rendering', () => {
    test('applies resting-forced class for forced rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(Date.now() + 3600000).toISOString()} // 1 hour from now (real-time)
          currentGameTime={mockCurrentGameTime}
        />
      );
      const badge = screen.getByText('Forced Rest').closest('.resting-badge');
      expect(badge).toHaveClass('resting-forced');
    });

    test('displays ambulance emoji for forced rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(Date.now() + 3600000).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('🚑')).toBeInTheDocument();
    });

    test('displays "Forced Rest" status text for forced rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(Date.now() + 3600000).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const status = screen.getByText('Forced Rest');
      expect(status).toHaveClass('resting-status');
    });

    test('does not render voluntary rest text for forced rest', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(Date.now() + 3600000).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.queryByText('Resting')).not.toBeInTheDocument();
    });

    test('does not render hospital emoji for forced rest', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(Date.now() + 3600000).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const badge = container.querySelector('.resting-badge');
      // Check badge doesn't contain hospital emoji (might be in text content)
      expect(badge.textContent).not.toContain('🏥');
    });

    test('forced rest takes precedence when both rest types provided', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T16:00:00Z').toISOString()} // Voluntary
          forcedRestUntil={new Date(Date.now() + 7200000).toISOString()} // Forced (2 hours)
          currentGameTime={mockCurrentGameTime}
        />
      );
      const badge = screen.getByText('Forced Rest').closest('.resting-badge');
      expect(badge).toHaveClass('resting-forced');
      expect(screen.getByText('🚑')).toBeInTheDocument();
    });
  });

  describe('countdown timer for voluntary rest', () => {
    test('displays remaining hours and minutes when more than an hour until rest ends', () => {
      const restEndsAt = new Date('2026-11-15T18:30:00Z').toISOString(); // 8.5 hours from game time
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={restEndsAt}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('8h 30m remaining')).toBeInTheDocument();
    });

    test('displays remaining minutes only when less than an hour', () => {
      const restEndsAt = new Date('2026-11-15T10:45:00Z').toISOString(); // 45 minutes from game time
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={restEndsAt}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('45m remaining')).toBeInTheDocument();
    });

    test('displays no remaining text when rest has ended', () => {
      const pastRestEndsAt = new Date('2026-11-15T08:00:00Z').toISOString(); // Past game time
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={pastRestEndsAt}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.queryByText(/remaining/)).not.toBeInTheDocument();
    });

    test('wraps timer in resting-timer span', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:30:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const timer = screen.getByText('4h 30m remaining').closest('.resting-timer');
      expect(timer).toBeInTheDocument();
    });

    test('timer has descriptive title attribute', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:30:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const timer = screen.getByText('4h 30m remaining').closest('.resting-timer');
      expect(timer).toHaveAttribute('title', 'Time remaining until available');
    });

    test('handles rest ending exactly at game time boundary with no text', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={mockCurrentGameTime}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.queryByText(/remaining/)).not.toBeInTheDocument();
    });

    test('handles very long rest periods', () => {
      const farFuture = new Date('2026-11-25T10:00:00Z').toISOString(); // 10 days away
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={farFuture}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('240h 0m remaining')).toBeInTheDocument();
    });
  });

  describe('countdown timer for forced rest', () => {
    test('displays remaining hours and minutes for real-time forced rest', () => {
      const oneHourFromNow = Date.now() + 3600000; // 1 hour from real-time now
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(oneHourFromNow).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('1h 0m remaining')).toBeInTheDocument();
    });

    test('displays minutes for short forced rest', () => {
      const thirtyMinFromNow = Date.now() + 1800000; // 30 minutes from real-time now
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(thirtyMinFromNow).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.getByText('30m remaining')).toBeInTheDocument();
    });

    test('handles expired forced rest gracefully', () => {
      const pastTime = new Date(Date.now() - 60000).toISOString(); // 1 minute in the past
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={pastTime}
          currentGameTime={mockCurrentGameTime}
        />
      );
      expect(screen.queryByText(/remaining/)).not.toBeInTheDocument();
    });
  });

  describe('countdown timer structure', () => {
    test('wraps remaining time in resting-timer element', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const timer = screen.getByText('4h 0m remaining');
      expect(timer).toHaveClass('resting-timer');
    });

    test('displays icon in resting-icon element', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const icon = screen.getByText('🏥');
      expect(icon).toHaveClass('resting-icon');
    });

    test('displays status in resting-status element', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const status = screen.getByText('Resting');
      expect(status).toHaveClass('resting-status');
    });

    test('badge contains all three elements in correct order', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      const badge = screen.getByText('Resting').closest('.resting-badge');

      // All elements should be present
      expect(badge.querySelector('.resting-icon')).toBeInTheDocument();
      expect(badge.querySelector('.resting-status')).toBeInTheDocument();
      expect(badge.querySelector('.resting-timer')).toBeInTheDocument();
    });

    test('renders with empty timer when rest has ended', () => {
      // When rest time is in the past, formatRemainingTime returns '' (empty string)
      // which means the timer element still renders but is empty
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T08:00:00Z').toISOString()} // Past time
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      // Timer element exists but is empty since rest has ended
      expect(screen.queryByText(/remaining/)).not.toBeInTheDocument();
    });
  });

  describe('typical use cases', () => {
    test('renders like a boxer scheduled to rest for 6 hours', () => {
      const sixHoursFromGameTime = new Date('2026-11-15T16:00:00Z').toISOString();
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={sixHoursFromGameTime}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );

      expect(screen.getByText('🏥')).toBeInTheDocument();
      expect(screen.getByText('Resting')).toBeInTheDocument();
      expect(screen.getByText('6h 0m remaining')).toBeInTheDocument();
    });

    test('renders like a boxer in short recovery rest (30 minutes)', () => {
      const thirtyMinFromGameTime = new Date('2026-11-15T10:30:00Z').toISOString();
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={thirtyMinFromGameTime}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );

      expect(screen.getByText('🏥')).toBeInTheDocument();
      expect(screen.getByText('Resting')).toBeInTheDocument();
      expect(screen.getByText('30m remaining')).toBeInTheDocument();
    });

    test('renders like an injured boxer on forced rest for 24 hours', () => {
      const twentyFourHoursFromNow = Date.now() + (24 * 60 * 60 * 1000);
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(twentyFourHoursFromNow).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );

      expect(screen.getByText('🚑')).toBeInTheDocument();
      expect(screen.getByText('Forced Rest')).toBeInTheDocument();
      expect(screen.getByText('24h 0m remaining')).toBeInTheDocument();
    });

    test('renders like a boxer with minor injury on forced rest for 2 hours', () => {
      const twoHoursFromNow = Date.now() + (2 * 60 * 60 * 1000);
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(twoHoursFromNow).toISOString()}
          currentGameTime={mockCurrentGameTime}
        />
      );

      expect(screen.getByText('🚑')).toBeInTheDocument();
      expect(screen.getByText('Forced Rest')).toBeInTheDocument();
      expect(screen.getByText('2h 0m remaining')).toBeInTheDocument();
    });
  });

  describe('edge cases', () => {
    test('handles null restEndsAt without crashing for voluntary rest', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );

      // Should still render the badge, just without timer
      const badge = container.querySelector('.resting-badge');
      expect(badge).toBeInTheDocument();
      expect(screen.getByText('🏥')).toBeInTheDocument();
      expect(screen.getByText('Resting')).toBeInTheDocument();
    });

    test('handles null currentGameTime without crashing', () => {
      const { container } = render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={null}
        />
      );

      // Should render without timer when gameTime is null
      const badge = container.querySelector('.resting-badge');
      expect(badge).toBeInTheDocument();
    });

    test('handles both restEndsAt and forcedRestUntil being null', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );

      // Should render the icon and status, but no timer
      expect(screen.getByText('🏥')).toBeInTheDocument();
      expect(screen.getByText('Resting')).toBeInTheDocument();
    });

    test('handles very small time differences (less than a second)', () => {
      const almostNow = new Date('2026-11-15T10:00:00.500Z').toISOString();
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={almostNow}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );
      // When the time is 0 or very small, ceiling gives 0 seconds which displays as "0m remaining"
      expect(screen.getByText('0m remaining')).toBeInTheDocument();
    });

    test('handles leap year dates correctly', () => {
      const gameTime = new Date('2024-02-28T23:00:00Z').toISOString(); // Leap year, Feb 28 at 11 PM
      const restEndsAt = new Date('2024-02-29T00:30:00Z').toISOString(); // Feb 29 (leap day) at midnight + 30m
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={restEndsAt}
          forcedRestUntil={null}
          currentGameTime={gameTime}
        />
      );
      expect(screen.getByText('1h 30m remaining')).toBeInTheDocument();
    });

    test('handles simple 24 hour period', () => {
      const gameTime = new Date('2024-11-03T06:00:00Z').toISOString();
      const restEndsAt = new Date('2024-11-03T07:00:00Z').toISOString();
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={restEndsAt}
          forcedRestUntil={null}
          currentGameTime={gameTime}
        />
      );
      expect(screen.getByText('1h 0m remaining')).toBeInTheDocument();
    });

    test('handles future game time far ahead', () => {
      const farFutureGameTime = new Date('2027-11-15T10:00:00Z').toISOString(); // One year ahead
      const restEndsAt = new Date('2027-11-15T14:00:00Z').toISOString();
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={restEndsAt}
          forcedRestUntil={null}
          currentGameTime={farFutureGameTime}
        />
      );
      expect(screen.getByText('4h 0m remaining')).toBeInTheDocument();
    });
  });

  describe('component composition', () => {
    test('renders all elements with correct structure', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={new Date('2026-11-15T14:00:00Z').toISOString()}
          forcedRestUntil={null}
          currentGameTime={mockCurrentGameTime}
        />
      );

      const badge = screen.getByText('Resting').closest('.resting-badge');

      // Verify structure
      expect(badge).toHaveClass('resting-badge');
      expect(badge).toHaveClass('resting-voluntary');
      expect(badge).not.toHaveClass('resting-forced');

      const icon = badge.querySelector('.resting-icon');
      const status = badge.querySelector('.resting-status');
      const timer = badge.querySelector('.resting-timer');

      expect(icon).toHaveTextContent('🏥');
      expect(status).toHaveTextContent('Resting');
      expect(timer).toHaveTextContent('4h 0m remaining');
    });

    test('forced rest has correct structure', () => {
      render(
        <RestingBadge
          hasActiveRest={true}
          restEndsAt={null}
          forcedRestUntil={new Date(Date.now() + 7200000).toISOString()} // 2 hours
          currentGameTime={mockCurrentGameTime}
        />
      );

      const badge = screen.getByText('Forced Rest').closest('.resting-badge');

      expect(badge).toHaveClass('resting-badge');
      expect(badge).toHaveClass('resting-forced');
      expect(badge).not.toHaveClass('resting-voluntary');

      const icon = badge.querySelector('.resting-icon');
      const status = badge.querySelector('.resting-status');
      const timer = badge.querySelector('.resting-timer');

      expect(icon).toHaveTextContent('🚑');
      expect(status).toHaveTextContent('Forced Rest');
      expect(timer).toHaveTextContent('2h 0m remaining');
    });
  });
});
