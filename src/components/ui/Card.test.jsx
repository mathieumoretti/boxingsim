/**
 * Card Component Tests
 *
 * Tests for the Card component system covering:
 * - Card container with variants
 * - CardHeader, CardBody, CardFooter sub-components
 * - Clickable state and onClick handlers
 * - Additional className merging
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/Card.jsx
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { Card, CardHeader, CardBody, CardFooter } from './Card';

describe('Card', () => {
  describe('container rendering', () => {
    test('renders children correctly', () => {
      render(<Card>Card Content</Card>);
      expect(screen.getByText('Card Content')).toBeInTheDocument();
    });

    test('applies base card class', () => {
      const { container } = render(<Card>Test</Card>);
      expect(container.querySelector('.card')).toBeInTheDocument();
    });
  });

  describe('variants', () => {
    test('renders with default variant by default', () => {
      const { container } = render(<Card>Test</Card>);
      const card = container.querySelector('.card');
      expect(card).toHaveClass('card-default');
    });

    test('renders light variant', () => {
      const { container } = render(<Card variant="light">Test</Card>);
      const card = container.querySelector('.card');
      expect(card).toHaveClass('card-light');
    });

    test('renders winner variant', () => {
      const { container } = render(<Card variant="winner">Test</Card>);
      const card = container.querySelector('.card');
      expect(card).toHaveClass('card-winner');
    });
  });

  describe('clickable state', () => {
    test('is not clickable by default', () => {
      const { container } = render(<Card>Test</Card>);
      const card = container.querySelector('.card');
      expect(card).not.toHaveClass('card-clickable');
      expect(card).not.toHaveAttribute('onClick');
    });

    test('adds clickable class when clickable prop is true', () => {
      const handleClick = jest.fn();
      const { container } = render(<Card clickable onClick={handleClick}>Test</Card>);
      const card = container.querySelector('.card');
      expect(card).toHaveClass('card-clickable');
    });

    test('calls onClick when clicked and clickable is true', () => {
      const handleClick = jest.fn();
      render(<Card clickable onClick={handleClick}>Test</Card>);

      const card = screen.getByText('Test').closest('.card');
      fireEvent.click(card);

      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    test('does not call onClick when clickable is false', () => {
      const handleClick = jest.fn();
      const { container } = render(<Card onClick={handleClick}>Test</Card>);
      const card = container.querySelector('.card');

      fireEvent.click(card);

      expect(handleClick).not.toHaveBeenCalled();
      expect(card).not.toHaveAttribute('onclick');
    });

    test('conditionally applies onClick handler only when clickable is true', () => {
      const handleClick = jest.fn();

      // Without clickable prop, onClick is not passed to the DOM element
      render(<Card onClick={handleClick}>Test</Card>);
      expect(screen.getByText('Test').closest('.card')).not.toHaveClass('card-clickable');

      // With clickable prop, handler should be applied
      render(<Card clickable onClick={handleClick}>Clickable Card</Card>);
      const clickableCard = screen.getByText('Clickable Card').closest('.card');
      expect(clickableCard).toHaveClass('card-clickable');
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes', () => {
      const { container } = render(<Card className="custom-class another-class">Test</Card>);
      const card = container.querySelector('.card');
      expect(card).toHaveClass('custom-class');
      expect(card).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(<Card className="">Test</Card>);
      const card = container.querySelector('.card');
      expect(card).toHaveClass('card');
      expect(card).toHaveClass('card-default');
    });
  });

  describe('with sub-components', () => {
    test('renders with CardHeader, CardBody, and CardFooter', () => {
      render(
        <Card>
          <CardHeader>Header Content</CardHeader>
          <CardBody>Body Content</CardBody>
          <CardFooter>Footer Content</CardFooter>
        </Card>
      );

      expect(screen.getByText('Header Content')).toBeInTheDocument();
      expect(screen.getByText('Body Content')).toBeInTheDocument();
      expect(screen.getByText('Footer Content')).toBeInTheDocument();
    });

    test('applies correct structure for typical RankCard use case', () => {
      const { container } = render(
        <Card variant="default" clickable>
          <CardHeader>
            <span>Rank #1</span>
          </CardHeader>
          <CardBody>
            <p>Champion details here</p>
          </CardBody>
          <CardFooter>
            <small>Record: 50-5-0</small>
          </CardFooter>
        </Card>
      );

      const card = container.querySelector('.card');
      expect(card).toHaveClass('card-default');
      expect(card).toHaveClass('card-clickable');

      expect(container.querySelector('.card-header')).toBeInTheDocument();
      expect(container.querySelector('.card-body')).toBeInTheDocument();
      expect(container.querySelector('.card-footer')).toBeInTheDocument();
    });
  });
});

describe('CardHeader', () => {
  test('renders children correctly', () => {
    render(<CardHeader>Header Content</CardHeader>);
    expect(screen.getByText('Header Content')).toBeInTheDocument();
  });

  test('applies card-header class', () => {
    const { container } = render(<CardHeader>Test</CardHeader>);
    expect(container.querySelector('.card-header')).toBeInTheDocument();
  });

  test('merges additional className', () => {
    const { container } = render(
      <CardHeader className="custom-header">Test</CardHeader>
    );
    expect(container.querySelector('.card-header')).toHaveClass('custom-header');
  });
});

describe('CardBody', () => {
  test('renders children correctly', () => {
    render(<CardBody>Body Content</CardBody>);
    expect(screen.getByText('Body Content')).toBeInTheDocument();
  });

  test('applies card-body class', () => {
    const { container } = render(<CardBody>Test</CardBody>);
    expect(container.querySelector('.card-body')).toBeInTheDocument();
  });

  test('merges additional className', () => {
    const { container } = render(
      <CardBody className="custom-body">Test</CardBody>
    );
    expect(container.querySelector('.card-body')).toHaveClass('custom-body');
  });
});

describe('CardFooter', () => {
  test('renders children correctly', () => {
    render(<CardFooter>Footer Content</CardFooter>);
    expect(screen.getByText('Footer Content')).toBeInTheDocument();
  });

  test('applies card-footer class', () => {
    const { container } = render(<CardFooter>Test</CardFooter>);
    expect(container.querySelector('.card-footer')).toBeInTheDocument();
  });

  test('merges additional className', () => {
    const { container } = render(
      <CardFooter className="custom-footer">Test</CardFooter>
    );
    expect(container.querySelector('.card-footer')).toHaveClass('custom-footer');
  });
});

describe('integration with typical usage patterns', () => {
  test('renders a complete fighter card structure', () => {
    render(
      <Card variant="default" clickable onClick={() => {}}>
        <CardHeader>
          <div className="avatar">🥊</div>
          <h3>Mike Tyson</h3>
        </CardHeader>
        <CardBody>
          <p>Current Champion</p>
          <div className="stats">STR: 85 | DEF: 70 | AGI: 90</div>
        </CardBody>
        <CardFooter>
          <span className="record">Record: 50-5-0</span>
          <span className="level">Level 50</span>
        </CardFooter>
      </Card>
    );

    expect(screen.getByText('Mike Tyson')).toBeInTheDocument();
    expect(screen.getByText('Current Champion')).toBeInTheDocument();
    expect(screen.getByText('Record: 50-5-0')).toBeInTheDocument();
    expect(screen.getByText('Level 50')).toBeInTheDocument();
  });
});
