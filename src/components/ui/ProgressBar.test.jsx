/**
 * ProgressBar Component Tests
 *
 * Tests for the ProgressBar component covering:
 * - All variants (health, energy, fatigue, xp, strength, defense, agility, neutral)
 * - Percentage calculations
 * - Label and value display
 * - Accessibility attributes
 * - Edge cases (clamping values)
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/ProgressBar.jsx
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProgressBar } from './ProgressBar';

describe('ProgressBar', () => {
  describe('rendering', () => {
    test('renders progress container with required class', () => {
      render(<ProgressBar value={50} label="Test" />);
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    test('applies progress-bar base class', () => {
      const { container } = render(<ProgressBar value={50} />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toBeInTheDocument();
    });
  });

  describe('variants', () => {
    test('renders with neutral variant by default', () => {
      const { container } = render(<ProgressBar value={50} />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-neutral');
    });

    test('renders health variant', () => {
      const { container } = render(<ProgressBar value={85} variant="health" label="Health" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-health');
    });

    test('renders energy variant', () => {
      const { container } = render(<ProgressBar value={70} variant="energy" label="Energy" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-energy');
    });

    test('renders fatigue variant', () => {
      const { container } = render(<ProgressBar value={45} variant="fatigue" label="Fatigue" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-fatigue');
    });

    test('renders xp variant', () => {
      const { container } = render(<ProgressBar value={145} max={200} variant="xp" label="XP" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-xp');
    });

    test('renders strength variant', () => {
      const { container } = render(<ProgressBar value={85} variant="strength" label="Strength" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-strength');
    });

    test('renders defense variant', () => {
      const { container } = render(<ProgressBar value={70} variant="defense" label="Defense" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-defense');
    });

    test('renders agility variant', () => {
      const { container } = render(<ProgressBar value={90} variant="agility" label="Agility" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveClass('progress-agility');
    });
  });

  describe('percentage calculation', () => {
    test('calculates percentage correctly with max=100', () => {
      render(<ProgressBar value={75} max={100} label="Test" />);
      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    test('calculates percentage correctly with custom max', () => {
      render(<ProgressBar value={50} max={200} label="Test" />);
      // 50/200 = 25%, but displays as raw value when max !== 100
      expect(screen.getByText('50/200')).toBeInTheDocument();
    });

    test('shows raw value format when max is not 100', () => {
      render(<ProgressBar value={145} max={500} variant="xp" label="XP" />);
      expect(screen.getByText('145/500')).toBeInTheDocument();
    });

    test('clamps percentage to 100% when value exceeds max', () => {
      const { container } = render(<ProgressBar value={150} max={100} />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveStyle('width: 100%');
    });

    test('clamps percentage to 0% when value is negative', () => {
      const { container } = render(<ProgressBar value={-50} max={100} />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveStyle('width: 0%');
    });

    test('handles zero value correctly', () => {
      const { container } = render(<ProgressBar value={0} max={100} />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveStyle('width: 0%');
    });

    test('handles full value correctly', () => {
      const { container } = render(<ProgressBar value={100} max={100} />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveStyle('width: 100%');
    });
  });

  describe('label display', () => {
    test('renders label when provided', () => {
      render(<ProgressBar value={50} label="Health" />);
      expect(screen.getByText('Health')).toBeInTheDocument();
    });

    test('does not render label container when label is not provided', () => {
      const { container } = render(<ProgressBar value={50} />);
      expect(container.querySelector('.progress-label')).not.toBeInTheDocument();
    });

    test('renders label text in correct element', () => {
      const { container } = render(<ProgressBar value={50} label="Energy" />);
      expect(screen.getByText('Energy')).toHaveClass('progress-label-text');
    });
  });

  describe('percentage display toggle', () => {
    test('shows percentage by default', () => {
      render(<ProgressBar value={75} max={100} label="Test" />);
      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    test('hides percentage when showPercentage is false', () => {
      render(<ProgressBar value={75} max={100} label="Test" showPercentage={false} />);
      expect(screen.queryByText('75%')).not.toBeInTheDocument();
      expect(screen.getByText('Test')).toBeInTheDocument(); // Label still shows
    });

    test('shows raw value when showPercentage is true and max !== 100', () => {
      render(<ProgressBar value={145} max={200} label="XP" showPercentage={true} />);
      expect(screen.getByText('145/200')).toBeInTheDocument();
    });

    test('hides raw value when showPercentage is false and max !== 100', () => {
      render(<ProgressBar value={145} max={200} label="XP" showPercentage={false} />);
      expect(screen.queryByText('145/200')).not.toBeInTheDocument();
    });
  });

  describe('custom height', () => {
    test('applies custom height when provided', () => {
      const { container } = render(<ProgressBar value={50} height={30} />);
      const track = container.querySelector('.progress-track');
      expect(track).toHaveStyle('height: 30px');
    });

    test('does not apply height style when not provided', () => {
      const { container } = render(<ProgressBar value={50} />);
      const track = container.querySelector('.progress-track');
      expect(track.style.height).toBe('');
    });
  });

  describe('additional className', () => {
    test('applies additional CSS classes to container', () => {
      const { container } = render(<ProgressBar value={50} label="Test" className="custom-class another-class" />);
      const containerEl = container.querySelector('.progress-container');
      expect(containerEl).toHaveClass('custom-class');
      expect(containerEl).toHaveClass('another-class');
    });
  });

  describe('accessibility', () => {
    test('has role="progressbar"', () => {
      render(<ProgressBar value={50} />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toBeInTheDocument();
    });

    test('has aria-valuenow attribute', () => {
      render(<ProgressBar value={75} />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-valuenow', '75');
    });

    test('has aria-valuemin attribute set to 0', () => {
      render(<ProgressBar value={50} />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-valuemin', '0');
    });

    test('has aria-valuemax attribute set to max value', () => {
      render(<ProgressBar value={50} max={200} label="Test" />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-valuemax', '200');
    });

    test('has aria-label with label and display value', () => {
      render(<ProgressBar value={75} max={100} label="Health" />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-label', 'Health: 75%');
    });

    test('has aria-label with raw values when max !== 100', () => {
      render(<ProgressBar value={145} max={200} label="XP" />);
      const bar = screen.getByRole('progressbar');
      expect(bar).toHaveAttribute('aria-label', 'XP: 145/200');
    });
  });

  describe('edge cases', () => {
    test('handles decimal values by rounding percentage', () => {
      render(<ProgressBar value={66.666} max={100} label="Test" />);
      expect(screen.getByText('67%')).toBeInTheDocument();
    });

    test('handles very small max values', () => {
      const { container } = render(<ProgressBar value={1} max={2} label="Half" />);
      const bar = container.querySelector('.progress-bar');
      expect(bar).toHaveStyle('width: 50%');
    });

    test('handles large values proportionally', () => {
      render(<ProgressBar value={9999} max={10000} label="Almost Full" />);
      expect(screen.getByText('9999/10000')).toBeInTheDocument();
    });
  });
});
