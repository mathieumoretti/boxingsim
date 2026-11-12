/**
 * StatDisplay Component Tests
 *
 * Tests for the StatDisplay component covering:
 * - All variants (compact, bar, mini)
 * - All colors (strength, defense, agility, level, neutral)
 * - Icon rendering
 * - Value display and rounding
 * - Bar variant with percentage calculation
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/StatDisplay.jsx
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { StatDisplay } from './StatDisplay';

describe('StatDisplay', () => {
  describe('rendering', () => {
    test('renders with compact variant by default', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      expect(container.querySelector('.stat-display')).toHaveClass('stat-compact');
    });

    test('renders label text correctly', () => {
      render(<StatDisplay label="Strength" value={85} />);
      expect(screen.getByText('Strength')).toBeInTheDocument();
    });

    test('renders value correctly', () => {
      render(<StatDisplay label="STR" value={85} />);
      expect(screen.getByText('85')).toBeInTheDocument();
    });

    test('rounds decimal values for display', () => {
      render(<StatDisplay label="STR" value={85.7} />);
      expect(screen.getByText('86')).toBeInTheDocument();
    });

    test('rounds down for values below 0.5', () => {
      render(<StatDisplay label="STR" value={85.2} />);
      expect(screen.getByText('85')).toBeInTheDocument();
    });
  });

  describe('variants', () => {
    test('renders compact variant (default)', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      const statEl = container.querySelector('.stat-display');
      expect(statEl).toHaveClass('stat-compact');

      // Compact has stat-info wrapper
      expect(container.querySelector('.stat-info')).toBeInTheDocument();
    });

    test('renders bar variant', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} variant="bar" maxValue={100} />
      );
      const statEl = container.querySelector('.stat-display');
      expect(statEl).toHaveClass('stat-bar');

      // Bar has stat-bar-container
      expect(container.querySelector('.stat-bar-container')).toBeInTheDocument();
    });

    test('renders mini variant', () => {
      const { container } = render(<StatDisplay label="LVL" value={15} variant="mini" />);
      const statEl = container.querySelector('.stat-display');
      expect(statEl).toHaveClass('stat-mini');

      // Mini does not have stat-info wrapper (only label and value)
      expect(container.querySelector('.stat-info')).not.toBeInTheDocument();
    });
  });

  describe('colors', () => {
    test('applies neutral color by default', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      expect(container.querySelector('.stat-display')).toHaveClass('stat-neutral');
    });

    test('applies strength color', () => {
      const { container } = render(<StatDisplay label="STR" value={85} color="strength" />);
      expect(container.querySelector('.stat-display')).toHaveClass('stat-strength');
    });

    test('applies defense color', () => {
      const { container } = render(<StatDisplay label="DEF" value={70} color="defense" />);
      expect(container.querySelector('.stat-display')).toHaveClass('stat-defense');
    });

    test('applies agility color', () => {
      const { container } = render(<StatDisplay label="AGI" value={90} color="agility" />);
      expect(container.querySelector('.stat-display')).toHaveClass('stat-agility');
    });

    test('applies level color', () => {
      const { container } = render(<StatDisplay label="LVL" value={50} color="level" />);
      expect(container.querySelector('.stat-display')).toHaveClass('stat-level');
    });
  });

  describe('icon rendering (compact variant)', () => {
    test('renders icon when provided', () => {
      const { container } = render(<StatDisplay label="STR" value={85} icon="💪" />);
      expect(container.querySelector('.stat-icon')).toBeInTheDocument();
      expect(screen.getByText('💪')).toBeInTheDocument();
    });

    test('does not render icon when not provided', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      expect(container.querySelector('.stat-icon')).not.toBeInTheDocument();
    });

    test('icon has title attribute matching label', () => {
      const { container } = render(<StatDisplay label="Strength" value={85} icon="💪" />);
      expect(container.querySelector('.stat-icon')).toHaveAttribute('title', 'Strength');
    });
  });

  describe('icon rendering (bar variant)', () => {
    test('renders icon when provided in bar variant', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} variant="bar" icon="💪" maxValue={100} color="strength" />
      );
      expect(container.querySelector('.stat-icon')).toBeInTheDocument();
    });

    test('does not render title on icon in bar variant', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} variant="bar" icon="💪" maxValue={100} />
      );
      expect(container.querySelector('.stat-icon')).not.toHaveAttribute('title');
    });
  });

  describe('icon rendering (mini variant)', () => {
    test('does not render icon in mini variant', () => {
      const { container } = render(<StatDisplay label="LVL" value={15} variant="mini" icon="⭐" />);
      expect(container.querySelector('.stat-icon')).not.toBeInTheDocument();
    });
  });

  describe('bar variant specific', () => {
    test('renders stat-bar-fill element', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} variant="bar" maxValue={100} />
      );
      expect(container.querySelector('.stat-bar-fill')).toBeInTheDocument();
    });

    test('calculates percentage correctly for bar', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={50} variant="bar" maxValue={100} />
      );
      const fillInner = container.querySelector('.stat-bar-fill div');
      expect(fillInner).toHaveStyle('width: 50%');
    });

    test('clamps percentage to 100% when value exceeds max', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={150} variant="bar" maxValue={100} />
      );
      const fillInner = container.querySelector('.stat-bar-fill div');
      expect(fillInner).toHaveStyle('width: 100%');
    });

    test('clamps percentage to 0% when value is negative', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={-10} variant="bar" maxValue={100} />
      );
      const fillInner = container.querySelector('.stat-bar-fill div');
      expect(fillInner).toHaveStyle('width: 0%');
    });

    test('applies color class to bar fill', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} variant="bar" color="strength" maxValue={100} />
      );
      expect(container.querySelector('.stat-bar-fill')).toHaveClass('stat-bar-strength');
    });

    test('shows rounded value in bar variant', () => {
      render(
        <StatDisplay label="Strength" value={85.7} variant="bar" maxValue={100} />
      );
      expect(screen.getByText('86')).toBeInTheDocument();
    });

    test('renders stat-label in bar variant', () => {
      render(
        <StatDisplay label="Strength" value={85} variant="bar" maxValue={100} />
      );
      expect(screen.getByText('Strength')).toHaveClass('stat-label');
    });

    test('renders stat-value in bar variant', () => {
      render(
        <StatDisplay label="Strength" value={85} variant="bar" maxValue={100} />
      );
      expect(screen.getByText('85')).toHaveClass('stat-value');
    });

    test('handles missing maxValue with 0% bar', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} variant="bar" />
      );
      const fillInner = container.querySelector('.stat-bar-fill div');
      expect(fillInner).toHaveStyle('width: 0%');
    });
  });

  describe('compact variant specific', () => {
    test('renders stat-name element', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      expect(container.querySelector('.stat-name')).toBeInTheDocument();
    });

    test('renders stat-value-sm element', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      expect(container.querySelector('.stat-value-sm')).toBeInTheDocument();
    });

    test('wraps name and value in stat-info', () => {
      const { container } = render(<StatDisplay label="STR" value={85} />);
      expect(container.querySelector('.stat-info')).toBeInTheDocument();
    });
  });

  describe('mini variant specific', () => {
    test('renders label with stat-label class', () => {
      const { container } = render(<StatDisplay label="LVL" value={15} variant="mini" />);
      expect(screen.getByText('LVL')).toHaveClass('stat-label');
    });

    test('renders value with stat-value class', () => {
      const { container } = render(<StatDisplay label="LVL" value={15} variant="mini" />);
      expect(screen.getByText('15')).toHaveClass('stat-value');
    });

    test('does not render icon in mini variant', () => {
      const { container } = render(<StatDisplay label="LVL" value={15} variant="mini" icon="⭐" />);
      expect(container.querySelector('.stat-icon')).not.toBeInTheDocument();
    });
  });

  describe('additional className', () => {
    test('applies additional CSS classes', () => {
      const { container } = render(
        <StatDisplay label="STR" value={85} className="custom-class another-class" />
      );
      expect(container.querySelector('.stat-display')).toHaveClass('custom-class');
      expect(container.querySelector('.stat-display')).toHaveClass('another-class');
    });
  });

  describe('typical use cases', () => {
    test('renders like BoxerCard combat stat (STR)', () => {
      const { container } = render(<StatDisplay label="STR" value={85} icon="💪" color="strength" />);

      expect(container.querySelector('.stat-display')).toHaveClass('stat-compact');
      expect(container.querySelector('.stat-display')).toHaveClass('stat-strength');
      expect(screen.getByText('💪')).toBeInTheDocument();
      expect(screen.getByText('STR')).toBeInTheDocument();
      expect(screen.getByText('85')).toBeInTheDocument();
    });

    test('renders like BoxerCard combat stat (AGI)', () => {
      const { container } = render(<StatDisplay label="AGI" value={90} icon="⚡" color="agility" />);

      expect(screen.getByText('⚡')).toBeInTheDocument();
      expect(screen.getByText('AGI')).toBeInTheDocument();
      expect(screen.getByText('90')).toBeInTheDocument();
    });

    test('renders like BoxerCard combat stat (DEF)', () => {
      const { container } = render(<StatDisplay label="DEF" value={70} icon="🛡️" color="defense" />);

      expect(screen.getByText('🛡️')).toBeInTheDocument();
      expect(screen.getByText('DEF')).toBeInTheDocument();
      expect(screen.getByText('70')).toBeInTheDocument();
    });

    test('renders like RankCard mini stat', () => {
      const { container } = render(<StatDisplay label="LVL" value={50} variant="mini" />);

      expect(container.querySelector('.stat-display')).toHaveClass('stat-mini');
      expect(screen.getByText('LVL')).toBeInTheDocument();
      expect(screen.getByText('50')).toBeInTheDocument();
    });

    test('renders like StatsOverview bar stat', () => {
      const { container } = render(
        <StatDisplay label="Strength" value={85} icon="💪" variant="bar" color="strength" maxValue={200} />
      );

      expect(container.querySelector('.stat-display')).toHaveClass('stat-bar');
      expect(screen.getByText('💪')).toBeInTheDocument();
      expect(screen.getByText('Strength')).toBeInTheDocument();
      expect(screen.getByText('85')).toBeInTheDocument();
    });
  });
});
