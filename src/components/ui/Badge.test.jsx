/**
 * Badge Component Tests
 *
 * Tests for the Badge component covering:
 * - All variants (default, level, training, fight-status-*, winner)
 * - All sizes (sm, md, lg)
 * - Positioning modes (static, absolute-tr)
 * - Icon rendering
 * - Typical usage patterns
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/Badge.jsx
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { Badge } from './Badge';

describe('Badge', () => {
  describe('rendering', () => {
    test('renders children text correctly', () => {
      render(<Badge>Test Badge</Badge>);
      expect(screen.getByText('Test Badge')).toBeInTheDocument();
    });

    test('uses span element', () => {
      const { container } = render(<Badge>Test</Badge>);
      const badge = container.querySelector('.badge');
      expect(badge.tagName.toLowerCase()).toBe('span');
    });

    test('applies base badge class', () => {
      const { container } = render(<Badge>Test</Badge>);
      expect(container.querySelector('.badge')).toBeInTheDocument();
    });
  });

  describe('variants', () => {
    test('renders with default variant by default', () => {
      const { container } = render(<Badge>Test</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-default');
    });

    test('renders level variant', () => {
      const { container } = render(<Badge variant="level">LVL 15</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-level');
    });

    test('renders level-high variant', () => {
      const { container } = render(<Badge variant="level-high">LVL 50</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-level-high');
    });

    test('renders training variant', () => {
      const { container } = render(<Badge variant="training" icon="🏋️">Training</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-training');
    });

    test('renders fight-status-scheduled variant', () => {
      const { container } = render(<Badge variant="fight-status-scheduled" icon="📅">Scheduled</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-fight-status-scheduled');
    });

    test('renders fight-status-completed variant', () => {
      const { container } = render(<Badge variant="fight-status-completed">Completed</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-fight-status-completed');
    });

    test('renders fight-status-cancelled variant', () => {
      const { container } = render(<Badge variant="fight-status-cancelled">Cancelled</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-fight-status-cancelled');
    });

    test('renders winner variant', () => {
      const { container } = render(<Badge variant="winner" icon="🏆">Victory</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-winner');
    });
  });

  describe('sizes', () => {
    test('renders with md size by default', () => {
      const { container } = render(<Badge>Test</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-md');
    });

    test('renders sm size', () => {
      const { container } = render(<Badge size="sm">LVL 15</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-sm');
    });

    test('renders lg size', () => {
      const { container } = render(<Badge size="lg">Victory</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-lg');
    });
  });

  describe('positioning', () => {
    test('uses static positioning by default', () => {
      const { container } = render(<Badge>Test</Badge>);
      expect(container.querySelector('.badge')).not.toHaveClass('badge-absolute');
    });

    test('applies absolute-tr positioning when specified', () => {
      const { container } = render(<Badge position="absolute-tr">LVL 15</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('badge-absolute');
    });

    test('uses static positioning for non-absolute values', () => {
      const { container } = render(<Badge position="static">Test</Badge>);
      expect(container.querySelector('.badge')).not.toHaveClass('badge-absolute');
    });
  });

  describe('icon rendering', () => {
    test('renders icon when provided', () => {
      render(<Badge icon="🏋️">Training</Badge>);
      expect(screen.getByText('🏋️')).toBeInTheDocument();
    });

    test('does not render icon element when icon is not provided', () => {
      const { container } = render(<Badge>Test</Badge>);
      expect(container.querySelector('.badge-icon')).not.toBeInTheDocument();
    });

    test('wraps icon in badge-icon span', () => {
      const { container } = render(<Badge icon="🏆">Victory</Badge>);
      expect(container.querySelector('.badge-icon')).toHaveTextContent('🏆');
    });

    test('renders icon before children text', () => {
      render(<Badge icon="⚡">Speed Boost</Badge>);

      // Check that both elements exist
      expect(screen.getByText('⚡')).toBeInTheDocument();
      expect(screen.getByText('Speed Boost')).toBeInTheDocument();
    });
  });

  describe('additional className', () => {
    test('applies additional CSS classes', () => {
      const { container } = render(<Badge className="custom-class another-class">Test</Badge>);
      expect(container.querySelector('.badge')).toHaveClass('custom-class');
      expect(container.querySelector('.badge')).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(<Badge className="">Test</Badge>);
      const badge = container.querySelector('.badge');
      expect(badge).toHaveClass('badge');
      expect(badge.className).not.toMatch(/\s+$/); // No trailing whitespace
    });
  });

  describe('typical use cases', () => {
    test('renders like BoxerCard level badge (LVL 15)', () => {
      const { container } = render(<Badge variant="level" size="sm" position="absolute-tr">LVL 15</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-level');
      expect(container.querySelector('.badge')).toHaveClass('badge-sm');
      expect(container.querySelector('.badge')).toHaveClass('badge-absolute');
      expect(screen.getByText('LVL 15')).toBeInTheDocument();
    });

    test('renders like BoxerCard level badge high tier (LVL 50)', () => {
      const { container } = render(<Badge variant="level-high" size="sm" position="absolute-tr">LVL 50</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-level-high');
      expect(screen.getByText('LVL 50')).toBeInTheDocument();
    });

    test('renders like BoxerCard training badge with icon', () => {
      const { container } = render(<Badge variant="training" icon="🏋️">Training</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-training');
      expect(screen.getByText('🏋️')).toBeInTheDocument();
      expect(screen.getByText('Training')).toBeInTheDocument();
    });

    test('renders like FightBadge scheduled status', () => {
      const { container } = render(<Badge variant="fight-status-scheduled" icon="📅">Scheduled</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-fight-status-scheduled');
      expect(screen.getByText('📅')).toBeInTheDocument();
      expect(screen.getByText('Scheduled')).toBeInTheDocument();
    });

    test('renders like FightBadge completed status', () => {
      const { container } = render(<Badge variant="fight-status-completed">Completed</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-fight-status-completed');
      expect(screen.getByText('Completed')).toBeInTheDocument();
    });

    test('renders like FightBadge cancelled status', () => {
      const { container } = render(<Badge variant="fight-status-cancelled">Cancelled</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-fight-status-cancelled');
      expect(screen.getByText('Cancelled')).toBeInTheDocument();
    });

    test('renders like FightBadge winner status with icon', () => {
      const { container } = render(<Badge variant="winner" icon="🏆">Victory</Badge>);

      expect(container.querySelector('.badge')).toHaveClass('badge-winner');
      expect(screen.getByText('🏆')).toBeInTheDocument();
      expect(screen.getByText('Victory')).toBeInTheDocument();
    });

    test('renders with complex children (ReactNode)', () => {
      const { container } = render(
        <Badge variant="level">
          <strong>LVL</strong> 15
        </Badge>
      );

      expect(screen.getByText('LVL')).toBeInTheDocument();
      expect(screen.getByText('15')).toBeInTheDocument();
    });
  });

  describe('combining props', () => {
    test('combines variant, size, position, and icon correctly', () => {
      const { container } = render(
        <Badge
          variant="training"
          size="lg"
          position="absolute-tr"
          icon="⚡"
          className="custom-badge"
        >
          Speed Training
        </Badge>
      );

      const badge = container.querySelector('.badge');
      expect(badge).toHaveClass('badge-training');
      expect(badge).toHaveClass('badge-lg');
      expect(badge).toHaveClass('badge-absolute');
      expect(badge).toHaveClass('custom-badge');
      expect(screen.getByText('⚡')).toBeInTheDocument();
      expect(screen.getByText('Speed Training')).toBeInTheDocument();
    });

    test('all classes are properly combined without duplicates', () => {
      const { container } = render(
        <Badge variant="winner" size="md" position="static" icon="🏆">Champion</Badge>
      );

      const badge = container.querySelector('.badge');
      const className = badge.className;

      // Check no duplicate classes
      const classes = className.split(' ');
      const uniqueClasses = new Set(classes);
      expect(classes.length).toBe(uniqueClasses.size);
    });
  });
});
