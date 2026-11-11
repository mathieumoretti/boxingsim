/**
 * Alert Component Tests
 *
 * Tests for the Alert component covering:
 * - All variants (info, success, warning, error)
 * - Title and message rendering
 * - Dismissible functionality
 * - Icon display for each variant
 * - Accessibility attributes
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/Alert.jsx
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { Alert } from './Alert';

describe('Alert', () => {
  describe('basic rendering', () => {
    test('renders children text correctly', () => {
      render(<Alert title="Test" message="This is a test alert" />);
      expect(screen.getByText('Test')).toBeInTheDocument();
      expect(screen.getByText('This is a test alert')).toBeInTheDocument();
    });

    test('uses div element with alert class', () => {
      const { container } = render(<Alert title="Test" message="Message" />);
      expect(container.querySelector('.alert')).toBeInTheDocument();
    });

    test('applies role="alert" for accessibility', () => {
      const { container } = render(<Alert title="Test" message="Message" />);
      expect(container.querySelector('.alert')).toHaveAttribute('role', 'alert');
    });
  });

  describe('variants', () => {
    test('renders with info variant by default', () => {
      const { container } = render(<Alert title="Test" message="Message" />);
      expect(container.querySelector('.alert')).toHaveClass('alert-info');
    });

    test('renders info variant with correct icon', () => {
      render(<Alert variant="info" title="Info" message="Information message" />);
      expect(screen.getByText('ℹ️')).toBeInTheDocument();
    });

    test('renders success variant', () => {
      const { container } = render(
        <Alert variant="success" title="Success" message="Action completed" />
      );
      expect(container.querySelector('.alert')).toHaveClass('alert-success');
      expect(screen.getByText('✅')).toBeInTheDocument();
    });

    test('renders warning variant', () => {
      const { container } = render(
        <Alert variant="warning" title="Warning" message="Check this out" />
      );
      expect(container.querySelector('.alert')).toHaveClass('alert-warning');
      expect(screen.getByText('⚠️')).toBeInTheDocument();
    });

    test('renders error variant', () => {
      const { container } = render(
        <Alert variant="error" title="Error" message="Something went wrong" />
      );
      expect(container.querySelector('.alert')).toHaveClass('alert-error');
      expect(screen.getByText('❌')).toBeInTheDocument();
    });

    test('all variants have appropriate icons', () => {
      render(<Alert variant="info" title="Test" message="Message" />);
      expect(screen.getByText('ℹ️')).toBeInTheDocument();

      render(<Alert variant="success" title="Test" message="Message" />);
      expect(screen.getByText('✅')).toBeInTheDocument();

      render(<Alert variant="warning" title="Test" message="Message" />);
      expect(screen.getByText('⚠️')).toBeInTheDocument();

      render(<Alert variant="error" title="Test" message="Message" />);
      expect(screen.getByText('❌')).toBeInTheDocument();
    });
  });

  describe('title rendering', () => {
    test('renders title when provided', () => {
      render(<Alert title="Alert Title" message="Message content" />);
      expect(screen.getByText('Alert Title')).toBeInTheDocument();
    });

    test('wraps title in strong element with alert-title class', () => {
      const { container } = render(
        <Alert title="Title" message="Message" />
      );
      expect(container.querySelector('.alert-title')).toHaveTextContent('Title');
    });

    test('does not render title when not provided', () => {
      const { container } = render(<Alert message="Only message" />);
      expect(container.querySelector('.alert-title')).not.toBeInTheDocument();
    });
  });

  describe('message rendering', () => {
    test('renders message when provided', () => {
      render(<Alert title="Title" message="This is the message body" />);
      expect(screen.getByText('This is the message body')).toBeInTheDocument();
    });

    test('wraps message in p element with alert-message class', () => {
      const { container } = render(
        <Alert title="Title" message="Message text" />
      );
      expect(container.querySelector('.alert-message')).toHaveTextContent('Message text');
    });

    test('does not render message when not provided', () => {
      const { container } = render(<Alert title="Only title" />);
      expect(container.querySelector('.alert-message')).not.toBeInTheDocument();
    });

    test('can render with only title (no message)', () => {
      render(<Alert title="Brief alert" />);
      expect(screen.getByText('Brief alert')).toBeInTheDocument();
    });

    test('can render with only message (no title)', () => {
      render(<Alert message="Message without title" />);
      expect(screen.getByText('Message without title')).toBeInTheDocument();
    });
  });

  describe('dismissible functionality', () => {
    test('does not show dismiss button by default', () => {
      const { container } = render(<Alert title="Test" message="Message" />);
      expect(container.querySelector('.alert-dismiss')).not.toBeInTheDocument();
    });

    test('shows dismiss button when dismissible is true and onDismiss provided', () => {
      render(
        <Alert
          title="Test"
          message="Message"
          dismissible
          onDismiss={jest.fn()}
        />
      );

      expect(screen.getByRole('button', { name: /dismiss alert/i })).toBeInTheDocument();
    });

    test('calls onDismiss when dismiss button is clicked', () => {
      const handleDismiss = jest.fn();
      render(
        <Alert
          title="Test"
          message="Message"
          dismissible
          onDismiss={handleDismiss}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /dismiss alert/i }));
      expect(handleDismiss).toHaveBeenCalledTimes(1);
    });

    test('dismissive button shows × symbol', () => {
      render(
        <Alert
          title="Test"
          message="Message"
          dismissible
          onDismiss={jest.fn()}
        />
      );

      const dismissBtn = screen.getByRole('button', { name: /dismiss alert/i });
      expect(dismissBtn).toHaveTextContent('✕');
    });

    test('dismissible adds alert-dismissible class', () => {
      const { container } = render(
        <Alert
          title="Test"
          message="Message"
          dismissible
          onDismiss={jest.fn()}
        />
      );

      expect(container.querySelector('.alert')).toHaveClass('alert-dismissible');
    });

    test('does not show dismiss when dismissible is true but no onDismiss', () => {
      const { container } = render(
        <Alert
          title="Test"
          message="Message"
          dismissible={true}
        />
      );

      // No onDismiss means no button rendered
      expect(container.querySelector('.alert-dismiss')).not.toBeInTheDocument();
    });
  });

  describe('icon rendering', () => {
    test('renders icon in alert-icon span', () => {
      const { container } = render(<Alert variant="success" title="Test" message="Message" />);
      expect(container.querySelector('.alert-icon')).toBeInTheDocument();
      expect(container.querySelector('.alert-icon')).toHaveTextContent('✅');
    });

    test('icon has aria-hidden attribute', () => {
      const { container } = render(<Alert variant="info" title="Test" message="Message" />);
      const icon = container.querySelector('.alert-icon');
      expect(icon).toHaveAttribute('aria-hidden', 'true');
    });

    test('icon appears before alert-content', () => {
      const { container } = render(<Alert variant="warning" title="Test" message="Message" />);
      const icon = container.querySelector('.alert-icon');
      const content = container.querySelector('.alert-content');

      // Both should exist
      expect(icon).toBeInTheDocument();
      expect(content).toBeInTheDocument();
    });
  });

  describe('alert content structure', () => {
    test('wraps title and message in alert-content div', () => {
      const { container } = render(<Alert title="Title" message="Message" />);
      const content = container.querySelector('.alert-content');

      expect(content).toBeInTheDocument();
      expect(content).toHaveTextContent('Title');
      expect(content).toHaveTextContent('Message');
    });

    test('renders alert-content even without title', () => {
      const { container } = render(<Alert message="Only message" />);
      expect(container.querySelector('.alert-content')).toBeInTheDocument();
    });

    test('renders alert-content even without message', () => {
      const { container } = render(<Alert title="Only title" />);
      expect(container.querySelector('.alert-content')).toBeInTheDocument();
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes', () => {
      const { container } = render(
        <Alert
          title="Test"
          message="Message"
          className="custom-class another-class"
        />
      );
      expect(container.querySelector('.alert')).toHaveClass('custom-class');
      expect(container.querySelector('.alert')).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(
        <Alert title="Test" message="Message" className="" />
      );
      const alert = container.querySelector('.alert');
      expect(alert).toHaveClass('alert');
      expect(alert).toHaveClass('alert-info');
    });
  });

  describe('typical use cases', () => {
    test('renders like a training success confirmation', () => {
      render(
        <Alert
          variant="success"
          title="Success!"
          message="Training scheduled successfully"
        />
      );

      expect(screen.getByText('Success!')).toBeInTheDocument();
      expect(screen.getByText('Training scheduled successfully')).toBeInTheDocument();
      expect(screen.getByText('✅')).toBeInTheDocument();
    });

    test('renders like a low energy warning with dismiss', () => {
      const handleDismiss = jest.fn();
      render(
        <Alert
          variant="warning"
          title="Low Energy"
          message="Boxer has insufficient energy for training"
          dismissible
          onDismiss={handleDismiss}
        />
      );

      expect(screen.getByText('Low Energy')).toBeInTheDocument();
      expect(screen.getByText('Boxer has insufficient energy for training')).toBeInTheDocument();
      expect(screen.getByText('⚠️')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /dismiss alert/i })).toBeInTheDocument();
    });

    test('renders like a fight booking error', () => {
      render(
        <Alert
          variant="error"
          title="Error"
          message="Failed to book fight. Please try again."
        />
      );

      expect(screen.getByText('Error')).toBeInTheDocument();
      expect(screen.getByText('Failed to book fight. Please try again.')).toBeInTheDocument();
      expect(screen.getByText('❌')).toBeInTheDocument();
    });

    test('renders like an info notification', () => {
      render(
        <Alert
          variant="info"
          title="Notice"
          message="Training session will start in 1 hour"
        />
      );

      expect(screen.getByText('Notice')).toBeInTheDocument();
      expect(screen.getByText('Training session will start in 1 hour')).toBeInTheDocument();
      expect(screen.getByText('ℹ️')).toBeInTheDocument();
    });

    test('renders all variant icons correctly', () => {
      const variants = [
        { variant: 'info', icon: 'ℹ️' },
        { variant: 'success', icon: '✅' },
        { variant: 'warning', icon: '⚠️' },
        { variant: 'error', icon: '❌' },
      ];

      variants.forEach(({ variant, icon }) => {
        render(<Alert variant={variant} title="Test" message="Message" />);
        expect(screen.getByText(icon)).toBeInTheDocument();
      });
    });
  });

  describe('combining props', () => {
    test('combines all props correctly', () => {
      const handleDismiss = jest.fn();
      const { container } = render(
        <Alert
          variant="success"
          title="Completed"
          message="Action was successful"
          dismissible
          onDismiss={handleDismiss}
          className="custom-alert"
        />
      );

      const alert = container.querySelector('.alert');
      expect(alert).toHaveClass('alert-success');
      expect(alert).toHaveClass('alert-dismissible');
      expect(alert).toHaveClass('custom-alert');

      expect(screen.getByText('Completed')).toBeInTheDocument();
      expect(screen.getByText('Action was successful')).toBeInTheDocument();
      expect(screen.getByText('✅')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /dismiss alert/i })).toBeInTheDocument();
    });

    test('all classes are properly combined without duplicates', () => {
      const handleDismiss = jest.fn();
      const { container } = render(
        <Alert
          variant="warning"
          title="Test"
          message="Message"
          dismissible
          onDismiss={handleDismiss}
        />
      );

      const alert = container.querySelector('.alert');
      const className = alert.className;

      // Check no duplicate classes
      const classes = className.split(' ');
      const uniqueClasses = new Set(classes);
      expect(classes.length).toBe(uniqueClasses.size);
    });
  });
});
