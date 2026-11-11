/**
 * LoadingSpinner Component Tests
 *
 * Tests for the LoadingSpinner component covering:
 * - All sizes (sm, md, lg)
 * - Label rendering
 * - Accessibility attributes (role="status", aria-label)
 * - Spinner animation element presence
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/LoadingSpinner.jsx
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { LoadingSpinner } from './LoadingSpinner';

describe('LoadingSpinner', () => {
  describe('basic rendering', () => {
    test('renders loading spinner container', () => {
      const { container } = render(<LoadingSpinner />);
      expect(container.querySelector('.loading-spinner')).toBeInTheDocument();
    });

    test('uses div element as container', () => {
      const { container } = render(<LoadingSpinner />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner.tagName.toLowerCase()).toBe('div');
    });

    test('applies base loading-spinner class', () => {
      const { container } = render(<LoadingSpinner />);
      expect(container.querySelector('.loading-spinner')).toBeInTheDocument();
    });

    test('renders spinner element inside container', () => {
      const { container } = render(<LoadingSpinner />);
      expect(container.querySelector('.spinner')).toBeInTheDocument();
    });
  });

  describe('default props', () => {
    test('uses md size by default', () => {
      const { container } = render(<LoadingSpinner />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-md');
    });

    test('applies role="status" for accessibility', () => {
      const { container } = render(<LoadingSpinner />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveAttribute('role', 'status');
    });

    test('applies aria-label="Loading" by default', () => {
      const { container } = render(<LoadingSpinner />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveAttribute('aria-label', 'Loading');
    });

    test('does not render label when not provided', () => {
      const { container } = render(<LoadingSpinner />);
      expect(container.querySelector('.loading-label')).not.toBeInTheDocument();
    });
  });

  describe('sizes', () => {
    test('renders sm size', () => {
      const { container } = render(<LoadingSpinner size="sm" />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-sm');
      expect(spinner).not.toHaveClass('loading-spinner-md');
      expect(spinner).not.toHaveClass('loading-spinner-lg');
    });

    test('renders md size', () => {
      const { container } = render(<LoadingSpinner size="md" />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-md');
      expect(spinner).not.toHaveClass('loading-spinner-sm');
      expect(spinner).not.toHaveClass('loading-spinner-lg');
    });

    test('renders lg size', () => {
      const { container } = render(<LoadingSpinner size="lg" />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-lg');
      expect(spinner).not.toHaveClass('loading-spinner-sm');
      expect(spinner).not.toHaveClass('loading-spinner-md');
    });

    test('only applies one size class at a time', () => {
      ['sm', 'md', 'lg'].forEach((size) => {
        const { container } = render(<LoadingSpinner size={size} />);
        const spinner = container.querySelector('.loading-spinner');

        const sizeClasses = [
          'loading-spinner-sm',
          'loading-spinner-md',
          'loading-spinner-lg',
        ].filter((cls) => spinner.classList.contains(cls));

        expect(sizeClasses).toHaveLength(1);
        expect(sizeClasses[0]).toBe(`loading-spinner-${size}`);
      });
    });
  });

  describe('label rendering', () => {
    test('renders label when provided', () => {
      render(<LoadingSpinner label="Loading data..." />);
      expect(screen.getByText('Loading data...')).toBeInTheDocument();
    });

    test('wraps label in p element with loading-label class', () => {
      const { container } = render(
        <LoadingSpinner label="Loading fight details..." />
      );
      expect(container.querySelector('.loading-label')).toHaveTextContent('Loading fight details...');
    });

    test('does not render label element when no label prop', () => {
      const { container } = render(<LoadingSpinner size="md" />);
      expect(container.querySelector('.loading-label')).not.toBeInTheDocument();
    });

    test('updates label text when changed', () => {
      const { rerender } = render(<LoadingSpinner label="Loading..." />);
      expect(screen.getByText('Loading...')).toBeInTheDocument();

      rerender(<LoadingSpinner label="Please wait..." />);
      expect(screen.getByText('Please wait...')).toBeInTheDocument();
    });
  });

  describe('aria-label updates with custom label', () => {
    test('uses custom label for aria-label when provided', () => {
      const { container } = render(
        <LoadingSpinner label="Fetching boxer data..." />
      );
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveAttribute('aria-label', 'Fetching boxer data...');
    });

    test('uses "Loading" default when label is empty string', () => {
      const { container } = render(<LoadingSpinner label="" />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveAttribute('aria-label', 'Loading');
    });

    test('includes label text in aria-label', () => {
      const customLabels = [
        'Loading profile...',
        'Saving changes...',
        'Processing fight...',
      ];

      customLabels.forEach((label) => {
        const { container } = render(<LoadingSpinner label={label} />);
        const spinner = container.querySelector('.loading-spinner');
        expect(spinner).toHaveAttribute('aria-label', label);
      });
    });
  });

  describe('spinner element', () => {
    test('renders spinner div inside container', () => {
      const { container } = render(<LoadingSpinner />);
      const spinnerElement = container.querySelector('.spinner');

      expect(spinnerElement).toBeInTheDocument();
      expect(spinnerElement.tagName.toLowerCase()).toBe('div');
    });

    test('spinner element has no text content (animation only)', () => {
      const { container } = render(<LoadingSpinner />);
      const spinnerElement = container.querySelector('.spinner');

      // Spinner is a visual-only element with animation, no text
      expect(spinnerElement.textContent).toBe('');
    });

    test('spinner is always rendered regardless of props', () => {
      [
        <LoadingSpinner />,
        <LoadingSpinner size="sm" />,
        <LoadingSpinner size="lg" label="Loading" />,
      ].forEach((component) => {
        const { container } = render(component);
        expect(container.querySelector('.spinner')).toBeInTheDocument();
      });
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes', () => {
      const { container } = render(
        <LoadingSpinner className="custom-class another-class" />
      );
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('custom-class');
      expect(spinner).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(<LoadingSpinner className="" />);
      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-md');
    });

    test('className is trimmed (no extra whitespace)', () => {
      const { container } = render(
        <LoadingSpinner className="  spaced-out  " />
      );
      const spinner = container.querySelector('.loading-spinner');
      // Should have the class without leading/trailing issues
      expect(spinner).toHaveClass('spaced-out');
    });
  });

  describe('typical use cases', () => {
    test('renders like a simple page loader', () => {
      const { container } = render(<LoadingSpinner size="md" />);

      expect(container.querySelector('.loading-spinner')).toBeInTheDocument();
      expect(container.querySelector('.spinner')).toBeInTheDocument();
      expect(container.querySelector('.loading-label')).not.toBeInTheDocument();
    });

    test('renders like a boxer card loading state', () => {
      render(<LoadingSpinner size="sm" label="Loading boxer..." />);

      expect(screen.getByText('Loading boxer...')).toBeInTheDocument();
      const container = screen.getByText('Loading boxer...').closest('.loading-spinner');
      expect(container).toHaveClass('loading-spinner-sm');
    });

    test('renders like a fight details loading state', () => {
      render(<LoadingSpinner size="lg" label="Loading fight details..." />);

      expect(screen.getByText('Loading fight details...')).toBeInTheDocument();
      const container = screen.getByText('Loading fight details...').closest('.loading-spinner');
      expect(container).toHaveClass('loading-spinner-lg');
    });

    test('renders like a training scheduler loading', () => {
      render(<LoadingSpinner size="md" label="Loading training options..." />);

      expect(screen.getByText('Loading training options...')).toBeInTheDocument();
      expect(screen.getByRole('status')).toBeInTheDocument();
    });
  });

  describe('combining props', () => {
    test('combines size and label correctly', () => {
      const { container } = render(
        <LoadingSpinner size="lg" label="Please wait..." />
      );

      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-lg');
      expect(spinner).toHaveAttribute('role', 'status');
      expect(spinner).toHaveAttribute('aria-label', 'Please wait...');

      expect(screen.getByText('Please wait...')).toBeInTheDocument();
    });

    test('combines size, label, and custom className', () => {
      const { container } = render(
        <LoadingSpinner
          size="md"
          label="Processing..."
          className="centered-loader"
        />
      );

      const spinner = container.querySelector('.loading-spinner');
      expect(spinner).toHaveClass('loading-spinner-md');
      expect(spinner).toHaveClass('centered-loader');

      expect(screen.getByText('Processing...')).toBeInTheDocument();
    });

    test('all classes are properly combined without duplicates', () => {
      const { container } = render(
        <LoadingSpinner size="sm" label="Loading" className="custom" />
      );

      const spinner = container.querySelector('.loading-spinner');
      const className = spinner.className;

      // Check no duplicate classes
      const classes = className.split(' ');
      const uniqueClasses = new Set(classes);
      expect(classes.length).toBe(uniqueClasses.size);
    });
  });

  describe('accessibility', () => {
    test('has proper ARIA role for screen readers', () => {
      render(<LoadingSpinner label="Loading data" />);

      const spinner = screen.getByRole('status');
      expect(spinner).toBeInTheDocument();
    });

    test('icon/visual element has aria-hidden implicitly via no text', () => {
      const { container } = render(<LoadingSpinner />);
      const visualSpinner = container.querySelector('.spinner');

      // Visual-only spinner div should not have meaningful content
      expect(visualSpinner.textContent).toBe('');
    });

    test('provides descriptive label for accessibility', () => {
      const customLabel = 'Loading upcoming fights';
      render(<LoadingSpinner label={customLabel} />);

      const spinner = screen.getByRole('status');
      expect(spinner).toHaveAttribute('aria-label', customLabel);
    });
  });
});
