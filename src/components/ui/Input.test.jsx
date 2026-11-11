/**
 * Input Component Tests
 *
 * Tests for the Input component covering:
 * - All input types (text, email, number, select)
 * - Label rendering
 * - Error and helper text states
 * - Disabled state
 * - Placeholder display
 * - Change event handling
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/Input.jsx
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { Input } from './Input';

describe('Input', () => {
  const defaultProps = {
    type: 'text',
    label: 'Test Input',
    name: 'test-input',
    value: '',
    onChange: jest.fn(),
  };

  describe('basic rendering', () => {
    test('renders input field with correct type', () => {
      render(<Input {...defaultProps} />);
      expect(screen.getByLabelText('Test Input')).toBeInTheDocument();
      expect(screen.getByLabelText('Test Input')).toHaveAttribute('type', 'text');
    });

    test('renders label when provided', () => {
      render(<Input {...defaultProps} />);
      expect(screen.getByText('Test Input')).toBeInTheDocument();
    });

    test('does not render label when not provided', () => {
      render(<Input type="text" value="" onChange={jest.fn()} />);
      expect(screen.queryByText('Label')).not.toBeInTheDocument();
    });

    test('wraps input in input-wrapper div', () => {
      const { container } = render(<Input {...defaultProps} />);
      expect(container.querySelector('.input-wrapper')).toBeInTheDocument();
    });
  });

  describe('input types', () => {
    test('renders text input by default', () => {
      const { container } = render(<Input {...defaultProps} />);
      const input = container.querySelector('.input-field');
      expect(input).toHaveClass('input-text');
      expect(input.tagName.toLowerCase()).toBe('input');
    });

    test('renders email input type', () => {
      const { container } = render(<Input {...defaultProps} type="email" />);
      const input = container.querySelector('.input-field');
      expect(input).toHaveClass('input-email');
      expect(input).toHaveAttribute('type', 'email');
    });

    test('renders number input type', () => {
      const { container } = render(<Input {...defaultProps} type="number" />);
      const input = container.querySelector('.input-field');
      expect(input).toHaveClass('input-number');
      expect(input).toHaveAttribute('type', 'number');
    });

    test('renders select dropdown for select type', () => {
      render(
        <Input
          {...defaultProps}
          type="select"
          options={[
            { value: 'option1', label: 'Option 1' },
            { value: 'option2', label: 'Option 2' },
          ]}
        />
      );

      const select = screen.getByLabelText('Test Input');
      expect(select.tagName.toLowerCase()).toBe('select');
      expect(select).toHaveClass('input-select');
    });
  });

  describe('select type options', () => {
    test('renders default option in select', () => {
      render(
        <Input
          {...defaultProps}
          type="select"
          options={[{ value: 'test', label: 'Test' }]}
        />
      );

      expect(screen.getByText('Select an option')).toBeInTheDocument();
    });

    test('renders all provided options in select', () => {
      render(
        <Input
          {...defaultProps}
          type="select"
          options={[
            { value: 'lightweight', label: 'Lightweight (135 lbs)' },
            { value: 'welterweight', label: 'Welterweight (147 lbs)' },
            { value: 'middleweight', label: 'Middleweight (154 lbs)' },
          ]}
        />
      );

      expect(screen.getByText('Lightweight (135 lbs)')).toBeInTheDocument();
      expect(screen.getByText('Welterweight (147 lbs)')).toBeInTheDocument();
      expect(screen.getByText('Middleweight (154 lbs)')).toBeInTheDocument();
    });

    test('does not render options list when type is not select', () => {
      const { container } = render(
        <Input
          {...defaultProps}
          type="text"
          options={[{ value: 'test', label: 'Test' }]}
        />
      );

      expect(container.querySelectorAll('option')).toHaveLength(0);
    });
  });

  describe('value and onChange', () => {
    test('sets input value correctly', () => {
      render(<Input {...defaultProps} value="test value" />);
      expect(screen.getByLabelText('Test Input')).toHaveValue('test value');
    });

    test('calls onChange when input value changes', () => {
      const handleChange = jest.fn();
      render(<Input {...defaultProps} value="" onChange={handleChange} />);

      fireEvent.change(screen.getByLabelText('Test Input'), {
        target: { value: 'new value' },
      });

      expect(handleChange).toHaveBeenCalledTimes(1);
    });

    test('calls onChange when select changes', () => {
      const handleChange = jest.fn();
      render(
        <Input
          {...defaultProps}
          type="select"
          value=""
          onChange={handleChange}
          options={[{ value: 'option1', label: 'Option 1' }]}
        />
      );

      fireEvent.change(screen.getByLabelText('Test Input'), {
        target: { value: 'option1' },
      });

      expect(handleChange).toHaveBeenCalledTimes(1);
    });
  });

  describe('disabled state', () => {
    test('does not apply disabled attribute by default', () => {
      const { container } = render(<Input {...defaultProps} />);
      expect(container.querySelector('.input-field')).not.toBeDisabled();
    });

    test('applies disabled attribute when disabled is true', () => {
      render(<Input {...defaultProps} disabled />);
      expect(screen.getByLabelText('Test Input')).toBeDisabled();
    });

    test('disables select when disabled is true', () => {
      render(
        <Input
          {...defaultProps}
          type="select"
          disabled
          options={[{ value: 'test', label: 'Test' }]}
        />
      );

      expect(screen.getByLabelText('Test Input')).toBeDisabled();
    });
  });

  describe('placeholder', () => {
    test('renders placeholder text when provided', () => {
      render(
        <Input
          {...defaultProps}
          placeholder="Enter your name"
        />
      );

      expect(screen.getByLabelText('Test Input')).toHaveAttribute(
        'placeholder',
        'Enter your name'
      );
    });

    test('does not render placeholder when not provided', () => {
      const { container } = render(<Input {...defaultProps} />);
      expect(container.querySelector('.input-field')).not.toHaveAttribute(
        'placeholder'
      );
    });
  });

  describe('error state', () => {
    test('renders error message when error prop is provided', () => {
      render(<Input {...defaultProps} error="This field is required" />);
      expect(screen.getByText('This field is required')).toBeInTheDocument();
    });

    test('applies input-helper-error class when error is present', () => {
      const { container } = render(
        <Input {...defaultProps} error="Error message" />
      );

      expect(container.querySelector('.input-helper-error')).toBeInTheDocument();
    });

    test('does not show error helper when no error', () => {
      const { container } = render(<Input {...defaultProps} />);
      expect(container.querySelector('.input-helper-error')).not.toBeInTheDocument();
    });
  });

  describe('helper text', () => {
    test('renders helper text when provided', () => {
      render(
        <Input
          {...defaultProps}
          helperText="Enter a valid email address"
        />
      );

      expect(screen.getByText('Enter a valid email address')).toBeInTheDocument();
    });

    test('applies input-helper class for helper text', () => {
      const { container } = render(
        <Input {...defaultProps} helperText="Helper message" />
      );

      expect(container.querySelector('.input-helper')).toBeInTheDocument();
      expect(container.querySelector('.input-helper-error')).not.toBeInTheDocument();
    });

    test('does not show helper when neither error nor helperText', () => {
      const { container } = render(<Input {...defaultProps} />);
      expect(container.querySelector('.input-helper')).not.toBeInTheDocument();
    });

    test('error takes precedence over helperText', () => {
      render(
        <Input
          {...defaultProps}
          error="Invalid value"
          helperText="Helper message"
        />
      );

      expect(screen.getByText('Invalid value')).toBeInTheDocument();
      expect(screen.queryByText('Helper message')).not.toBeInTheDocument();
    });
  });

  describe('name and id attributes', () => {
    test('sets name attribute when provided', () => {
      render(<Input {...defaultProps} name="boxerName" />);
      expect(screen.getByLabelText('Test Input')).toHaveAttribute('name', 'boxerName');
    });

    test('sets id attribute matching name', () => {
      render(<Input {...defaultProps} name="boxerName" />);
      expect(screen.getByLabelText('Test Input')).toHaveAttribute('id', 'boxerName');
    });

    test('labels are associated with inputs via htmlFor', () => {
      const { container } = render(<Input {...defaultProps} name="testField" />);
      const label = container.querySelector('.input-label');
      expect(label).toHaveAttribute('for', 'testField');
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes', () => {
      const { container } = render(
        <Input {...defaultProps} className="custom-class another-class" />
      );
      const input = container.querySelector('.input-field');
      expect(input).toHaveClass('custom-class');
      expect(input).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(<Input {...defaultProps} className="" />);
      const input = container.querySelector('.input-field');
      expect(input).toHaveClass('input-field');
      expect(input).toHaveClass('input-text');
    });
  });

  describe('typical use cases', () => {
    test('renders like a name input form field', () => {
      render(
        <Input
          type="text"
          label="Boxer Name"
          name="boxerName"
          value=""
          onChange={jest.fn()}
          placeholder="Enter boxer name"
          helperText="This will be displayed on cards and rankings"
        />
      );

      expect(screen.getByText('Boxer Name')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Enter boxer name')).toBeInTheDocument();
      expect(screen.getByText('This will be displayed on cards and rankings')).toBeInTheDocument();
    });

    test('renders like an email input with error', () => {
      render(
        <Input
          type="email"
          label="Email Address"
          name="email"
          value="invalid-email"
          onChange={jest.fn()}
          error="Please enter a valid email address"
        />
      );

      expect(screen.getByText('Email Address')).toBeInTheDocument();
      expect(screen.getByLabelText('Email Address')).toHaveValue('invalid-email');
      expect(screen.getByText('Please enter a valid email address')).toBeInTheDocument();
    });

    test('renders like a weight class select', () => {
      render(
        <Input
          type="select"
          label="Weight Class"
          name="weightClass"
          value=""
          onChange={jest.fn()}
          options={[
            { value: 'lightweight', label: 'Lightweight (135 lbs)' },
            { value: 'welterweight', label: 'Welterweight (147 lbs)' },
            { value: 'middleweight', label: 'Middleweight (154 lbs)' },
            { value: 'heavyweight', label: 'Heavyweight (200+ lbs)' },
          ]}
        />
      );

      expect(screen.getByText('Weight Class')).toBeInTheDocument();
      const select = screen.getByLabelText('Weight Class');
      expect(select.querySelector('option[value="lightweight"]')).toHaveTextContent('Lightweight (135 lbs)');
    });

    test('renders like a number input for training duration', () => {
      render(
        <Input
          type="number"
          label="Duration (hours)"
          name="duration"
          value={2}
          onChange={jest.fn()}
          disabled
        />
      );

      expect(screen.getByText('Duration (hours)')).toBeInTheDocument();
      const input = screen.getByLabelText('Duration (hours)');
      expect(input).toHaveValue(2);
      expect(input).toBeDisabled();
    });
  });
});
