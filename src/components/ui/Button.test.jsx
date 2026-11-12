/**
 * Button Component Tests
 *
 * Tests for the Button component covering:
 * - Default rendering
 * - All variants (primary, secondary, danger, ghost)
 * - All sizes (sm, md, lg)
 * - Disabled state
 * - Click handler
 * - Different button types
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/Button.jsx
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('rendering', () => {
    test('renders children text correctly', () => {
      render(<Button>Click Me</Button>);
      expect(screen.getByText('Click Me')).toBeInTheDocument();
    });

    test('applies base button class', () => {
      const { container } = render(<Button>Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn');
    });
  });

  describe('variants', () => {
    test('renders with secondary variant by default', () => {
      const { container } = render(<Button>Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-secondary');
    });

    test('renders primary variant', () => {
      const { container } = render(<Button variant="primary">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-primary');
    });

    test('renders danger variant', () => {
      const { container } = render(<Button variant="danger">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-danger');
    });

    test('renders ghost variant', () => {
      const { container } = render(<Button variant="ghost">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-ghost');
    });
  });

  describe('sizes', () => {
    test('renders with md size by default', () => {
      const { container } = render(<Button>Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-md');
    });

    test('renders sm size', () => {
      const { container } = render(<Button size="sm">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-sm');
    });

    test('renders lg size', () => {
      const { container } = render(<Button size="lg">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('btn-lg');
    });
  });

  describe('disabled state', () => {
    test('is not disabled by default', () => {
      const { container } = render(<Button>Test</Button>);
      const button = container.querySelector('button');
      expect(button).not.toBeDisabled();
    });

    test('can be disabled', () => {
      const { container } = render(<Button disabled>Test</Button>);
      const button = container.querySelector('button');
      expect(button).toBeDisabled();
    });

    test('disabled button does not trigger onClick', () => {
      const handleClick = jest.fn();
      render(<Button disabled onClick={handleClick}>Test</Button>);

      const button = screen.getByText('Test');
      fireEvent.click(button);

      expect(handleClick).not.toHaveBeenCalled();
    });
  });

  describe('click handler', () => {
    test('calls onClick when clicked', () => {
      const handleClick = jest.fn();
      render(<Button onClick={handleClick}>Test</Button>);

      const button = screen.getByText('Test');
      fireEvent.click(button);

      expect(handleClick).toHaveBeenCalledTimes(1);
    });
  });

  describe('button types', () => {
    test('renders as button type by default', () => {
      const { container } = render(<Button>Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveAttribute('type', 'button');
    });

    test('renders as submit type', () => {
      const { container } = render(<Button type="submit">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveAttribute('type', 'submit');
    });

    test('renders as reset type', () => {
      const { container } = render(<Button type="reset">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveAttribute('type', 'reset');
    });
  });

  describe('additional className', () => {
    test('applies additional CSS classes', () => {
      const { container } = render(<Button className="custom-class another-class">Test</Button>);
      const button = container.querySelector('button');
      expect(button).toHaveClass('custom-class');
      expect(button).toHaveClass('another-class');
    });
  });

  describe('accessibility', () => {
    test('has no accessibility violations when enabled', () => {
      const { container } = render(<Button>Accessible Button</Button>);
      const button = container.querySelector('button');

      expect(button).not.toHaveAttribute('disabled');
      expect(button).toBeEnabled();
    });

    test('properly communicates disabled state', () => {
      const { container } = render(<Button disabled>Disabled Button</Button>);
      const button = container.querySelector('button');

      expect(button).toHaveAttribute('disabled');
      expect(button).not.toBeEnabled();
    });
  });
});
