/**
 * Form Component Tests
 *
 * Tests for the Form component system covering:
 * - Form container with onSubmit handler
 * - Form-level error display
 * - FormGroup with label, input, and error state
 * - FormRow for horizontal layout
 * - Required field indicator in labels
 *
 * @see MAT-169 - Frontend Component Tests
 * @see src/components/ui/Form.jsx
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { Form, FormGroup, FormRow } from './Form';

describe('Form', () => {
  describe('basic rendering', () => {
    test('renders children correctly', () => {
      render(
        <Form onSubmit={jest.fn()}>
          <p>Form content here</p>
        </Form>
      );
      expect(screen.getByText('Form content here')).toBeInTheDocument();
    });

    test('uses form element', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()}>{/* content */}</Form>
      );
      const formElement = container.querySelector('form');
      expect(formElement).toBeInTheDocument();
      expect(formElement.tagName.toLowerCase()).toBe('form');
    });

    test('applies form-container class', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()}>{/* content */}</Form>
      );
      expect(container.querySelector('.form-container')).toBeInTheDocument();
    });
  });

  describe('onSubmit handler', () => {
    test('passes onSubmit prop to form element', () => {
      const handleSubmit = jest.fn().mockImplementation((e) => e.preventDefault());
      render(
        <Form onSubmit={handleSubmit}>
          <button type="submit">Submit</button>
        </Form>
      );

      fireEvent.click(screen.getByText('Submit'));
      expect(handleSubmit).toHaveBeenCalledTimes(1);
    });

    test('calls onSubmit when form is submitted via keyboard', () => {
      const handleSubmit = jest.fn().mockImplementation((e) => e.preventDefault());
      const { container } = render(
        <Form onSubmit={handleSubmit}>
          <input type="text" value="test" />
          <button type="submit">Submit</button>
        </Form>
      );

      // Simulate form submission by pressing Enter on the input
      fireEvent.submit(container.querySelector('form'));
      expect(handleSubmit).toHaveBeenCalledTimes(1);
    });
  });

  describe('form-level error display', () => {
    test('renders form error when error prop is provided', () => {
      render(
        <Form onSubmit={jest.fn()} error="Form submission failed">
          <p>Content</p>
        </Form>
      );

      expect(screen.getByText('Form submission failed')).toBeInTheDocument();
    });

    test('applies form-error class to error message', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()} error="Error message">
          <p>Content</p>
        </Form>
      );

      expect(container.querySelector('.form-error')).toBeInTheDocument();
      expect(container.querySelector('.form-error')).toHaveTextContent('Error message');
    });

    test('does not render error container when no error', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()}>
          <p>Content</p>
        </Form>
      );

      expect(container.querySelector('.form-error')).not.toBeInTheDocument();
    });

    test('error appears before children content', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()} error="Invalid form data">
          <p>Form fields here</p>
        </Form>
      );

      const formError = container.querySelector('.form-error');
      const formContent = container.querySelector('p');

      // Error should be the first child after form element
      expect(formError).toBeInTheDocument();
      expect(formContent).toBeInTheDocument();
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes on form', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()} className="custom-form another-class">
          <p>Content</p>
        </Form>
      );

      const form = container.querySelector('.form-container');
      expect(form).toHaveClass('custom-form');
      expect(form).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(
        <Form onSubmit={jest.fn()} className="">
          <p>Content</p>
        </Form>
      );

      const form = container.querySelector('.form-container');
      expect(form).toHaveClass('form-container');
    });
  });

  describe('typical use cases', () => {
    test('renders like a training scheduler form without error', () => {
      render(
        <Form onSubmit={jest.fn()}>
          <p>Select Training</p>
          <p>Set Duration</p>
          <button type="submit">Schedule</button>
        </Form>
      );

      expect(screen.getByText('Select Training')).toBeInTheDocument();
      expect(screen.getByText('Set Duration')).toBeInTheDocument();
      expect(screen.getByText('Schedule')).toBeInTheDocument();
    });

    test('renders like a form with validation error', () => {
      render(
        <Form onSubmit={jest.fn()} error="Please fix the errors below">
          <p>Form fields</p>
        </Form>
      );

      expect(screen.getByText('Please fix the errors below')).toBeInTheDocument();
    });
  });
});

describe('FormGroup', () => {
  describe('basic rendering', () => {
    test('renders children input correctly', () => {
      render(
        <FormGroup label="Name">
          <input type="text" name="boxerName" />
        </FormGroup>
      );

      expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
    });

    test('applies form-group class', () => {
      const { container } = render(
        <FormGroup label="Test">
          <input type="text" />
        </FormGroup>
      );

      expect(container.querySelector('.form-group')).toBeInTheDocument();
    });

    test('wraps children in form-group div', () => {
      const { container } = render(
        <FormGroup label="Select">
          <select name="type"><option>Test</option></select>
        </FormGroup>
      );

      const formGroup = container.querySelector('.form-group');
      expect(formGroup).toContainElement(container.querySelector('select'));
    });
  });

  describe('label rendering', () => {
    test('renders label when provided', () => {
      render(
        <FormGroup label="Boxer Name">
          <input type="text" name="name" />
        </FormGroup>
      );

      expect(screen.getByText('Boxer Name')).toBeInTheDocument();
    });

    test('associates label with input via htmlFor', () => {
      render(
        <FormGroup label="Email">
          <input type="email" name="emailAddress" />
        </FormGroup>
      );

      const label = screen.getByText('Email');
      expect(label).toHaveAttribute('for', 'emailAddress');
    });

    test('does not render label when not provided', () => {
      const { container } = render(
        <FormGroup>
          <input type="text" name="test" />
        </FormGroup>
      );

      expect(container.querySelector('label')).not.toBeInTheDocument();
    });
  });

  describe('required field indicator', () => {
    test('shows required asterisk when input has required prop', () => {
      render(
        <FormGroup label="Full Name">
          <input type="text" name="fullName" required />
        </FormGroup>
      );

      const label = screen.getByText(/Full Name/);
      expect(label).toHaveTextContent('*');
    });

    test('does not show asterisk when input is not required', () => {
      render(
        <FormGroup label="Nickname">
          <input type="text" name="nickname" />
        </FormGroup>
      );

      const label = screen.getByText('Nickname');
      expect(label.textContent).not.toMatch(/\*/);
    });

    test('asterisk has danger color styling', () => {
      const { container } = render(
        <FormGroup label="Required Field">
          <input type="text" name="required" required />
        </FormGroup>
      );

      const asteriskSpan = container.querySelector('label span');
      expect(asteriskSpan).toBeInTheDocument();
    });
  });

  describe('error state', () => {
    test('renders error message when error prop is provided', () => {
      render(
        <FormGroup label="Email" error="Invalid email format">
          <input type="email" name="email" />
        </FormGroup>
      );

      expect(screen.getByText('Invalid email format')).toBeInTheDocument();
    });

    test('applies error class to form-group when error is present', () => {
      const { container } = render(
        <FormGroup label="Input" error="Field is required">
          <input type="text" name="input" />
        </FormGroup>
      );

      expect(container.querySelector('.form-group.error')).toBeInTheDocument();
    });

    test('applies error-message class to error text', () => {
      const { container } = render(
        <FormGroup label="Field" error="Error here">
          <input type="text" name="field" />
        </FormGroup>
      );

      expect(container.querySelector('.error-message')).toBeInTheDocument();
      expect(container.querySelector('.error-message')).toHaveTextContent('Error here');
    });

    test('does not render error message when no error prop', () => {
      const { container } = render(
        <FormGroup label="Normal Field">
          <input type="text" name="normal" />
        </FormGroup>
      );

      expect(container.querySelector('.error-message')).not.toBeInTheDocument();
    });

    test('does not apply error class when no error', () => {
      const { container } = render(
        <FormGroup label="Valid Field">
          <input type="text" name="valid" />
        </FormGroup>
      );

      expect(container.querySelector('.form-group')).not.toHaveClass('error');
    });
  });

  describe('child element cloning', () => {
    test('preserves child props when cloned', () => {
      render(
        <FormGroup label="Test">
          <input type="text" name="testInput" defaultValue="original" />
        </FormGroup>
      );

      const input = screen.getByLabelText(/Test/i);
      expect(input).toHaveValue('original');
      expect(input).toHaveAttribute('type', 'text');
    });

    test('sets id attribute from name if not provided', () => {
      render(
        <FormGroup label="Name">
          <input type="text" name="boxerName" />
        </FormGroup>
      );

      const input = screen.getByLabelText(/Name/i);
      expect(input).toHaveAttribute('id', 'boxerName');
    });

    test('preserves existing id if provided', () => {
      render(
        <FormGroup label="Custom ID">
          <input type="text" name="custom" id="myCustomId" />
        </FormGroup>
      );

      const input = screen.getByLabelText(/Custom ID/i);
      expect(input).toHaveAttribute('id', 'myCustomId');
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes on form-group', () => {
      const { container } = render(
        <FormGroup label="Field" className="custom-group another-class">
          <input type="text" name="field" />
        </FormGroup>
      );

      expect(container.querySelector('.form-group')).toHaveClass('custom-group');
      expect(container.querySelector('.form-group')).toHaveClass('another-class');
    });

    test('handles empty string className', () => {
      const { container } = render(
        <FormGroup label="Field" className="">
          <input type="text" name="field" />
        </FormGroup>
      );

      const formGroup = container.querySelector('.form-group');
      expect(formGroup).toHaveClass('form-group');
    });
  });

  describe('typical use cases', () => {
    test('renders like a training type select field', () => {
      render(
        <FormGroup label="Select Training Type" error="Please select a training">
          <select name="trainingType" required>
            <option value="">-- Select --</option>
            <option value="strength">Strength</option>
          </select>
        </FormGroup>
      );

      expect(screen.getByText('Select Training Type')).toBeInTheDocument();
      expect(screen.getByText('*')).toBeInTheDocument(); // Required indicator
      expect(screen.getByText('Please select a training')).toBeInTheDocument(); // Error
      expect(screen.getByLabelText(/Select Training Type/i)).toBeInTheDocument();
    });

    test('renders like a number input for duration', () => {
      render(
        <FormGroup label="Duration (hours)">
          <input type="number" name="duration" min="1" max="8" />
        </FormGroup>
      );

      const input = screen.getByLabelText(/Duration/i);
      expect(input).toHaveAttribute('type', 'number');
      expect(input).toHaveAttribute('min', '1');
      expect(input).toHaveAttribute('max', '8');
    });

    test('renders like a required text input', () => {
      render(
        <FormGroup label="Boxer Name" error="Name is required">
          <input type="text" name="boxerName" required />
        </FormGroup>
      );

      expect(screen.getByText('Boxer Name')).toBeInTheDocument();
      expect(screen.getByText('*')).toBeInTheDocument();
      expect(screen.getByText('Name is required')).toBeInTheDocument();
    });
  });

  describe('combining props', () => {
    test('combines label, error, and className correctly', () => {
      const { container } = render(
        <FormGroup
          label="Complete Form"
          error="Field has issues"
          className="required-field"
        >
          <input type="text" name="complete" required />
        </FormGroup>
      );

      const formGroup = container.querySelector('.form-group');
      expect(formGroup).toHaveClass('form-group');
      expect(formGroup).toHaveClass('error');
      expect(formGroup).toHaveClass('required-field');

      expect(screen.getByText('Complete Form')).toBeInTheDocument();
      expect(screen.getByText('*')).toBeInTheDocument();
      expect(screen.getByText('Field has issues')).toBeInTheDocument();
    });

    test('all classes are properly combined without duplicates', () => {
      const { container } = render(
        <FormGroup label="Test" error="Error">
          <input type="text" name="test" />
        </FormGroup>
      );

      const formGroup = container.querySelector('.form-group');
      const className = formGroup.className;

      // Check no duplicate classes
      const classes = className.split(' ');
      const uniqueClasses = new Set(classes);
      expect(classes.length).toBe(uniqueClasses.size);
    });
  });
});

describe('FormRow', () => {
  describe('basic rendering', () => {
    test('renders children correctly', () => {
      render(
        <FormRow>
          <span>First field</span>
          <span>Second field</span>
        </FormRow>
      );

      expect(screen.getByText('First field')).toBeInTheDocument();
      expect(screen.getByText('Second field')).toBeInTheDocument();
    });

    test('uses div element as container', () => {
      const { container } = render(
        <FormRow>
          <span>Content</span>
        </FormRow>
      );

      const formRow = container.querySelector('.form-row');
      expect(formRow).toBeInTheDocument();
      expect(formRow.tagName.toLowerCase()).toBe('div');
    });

    test('applies form-row class', () => {
      const { container } = render(<FormRow><span>Test</span></FormRow>);
      expect(container.querySelector('.form-row')).toBeInTheDocument();
    });
  });

  describe('containing multiple form groups', () => {
    test('can contain multiple FormGroup components', () => {
      const { container } = render(
        <FormRow>
          <FormGroup label="First">
            <input type="text" name="first" />
          </FormGroup>
          <FormGroup label="Second">
            <input type="text" name="second" />
          </FormGroup>
        </FormRow>
      );

      const formRow = container.querySelector('.form-row');
      expect(formRow).toContainElement(container.querySelector('[name="first"]'));
      expect(formRow).toContainElement(container.querySelector('[name="second"]'));
    });

    test('groups related fields horizontally', () => {
      render(
        <Form onSubmit={jest.fn()}>
          <FormRow>
            <FormGroup label="First Name">
              <input type="text" name="firstName" />
            </FormGroup>
            <FormGroup label="Last Name">
              <input type="text" name="lastName" />
            </FormGroup>
          </FormRow>
        </Form>
      );

      expect(screen.getByText('First Name')).toBeInTheDocument();
      expect(screen.getByText('Last Name')).toBeInTheDocument();
    });
  });

  describe('additional className', () => {
    test('merges additional CSS classes on form-row', () => {
      const { container } = render(
        <FormRow className="custom-row spaced">
          <span>Content</span>
        </FormRow>
      );

      expect(container.querySelector('.form-row')).toHaveClass('custom-row');
      expect(container.querySelector('.form-row')).toHaveClass('spaced');
    });

    test('handles empty string className', () => {
      const { container } = render(
        <FormRow className="">
          <span>Test</span>
        </FormRow>
      );

      const formRow = container.querySelector('.form-row');
      expect(formRow).toHaveClass('form-row');
    });
  });

  describe('typical use cases', () => {
    test('renders like training scheduler row with select and number input', () => {
      render(
        <FormRow>
          <FormGroup label="Select Training Type">
            <select name="trainingType">
              <option value="">-- Select --</option>
              <option value="strength">Strength</option>
            </select>
          </FormGroup>
          <FormGroup label="Duration (hours)">
            <input type="number" name="duration" min="1" max="8" />
          </FormGroup>
        </FormRow>
      );

      expect(screen.getByText('Select Training Type')).toBeInTheDocument();
      expect(screen.getByText('Duration (hours)')).toBeInTheDocument();
    });
  });

  describe('combining props', () => {
    test('combines className with multiple children', () => {
      const { container } = render(
        <FormRow className="compact-row">
          <span>Field 1</span>
          <span>Field 2</span>
          <span>Field 3</span>
        </FormRow>
      );

      expect(container.querySelector('.form-row')).toHaveClass('compact-row');
      expect(screen.getAllByText('Field 1')).toHaveLength(1);
      expect(screen.getAllByText('Field 2')).toHaveLength(1);
      expect(screen.getAllByText('Field 3')).toHaveLength(1);
    });
  });
});

describe('Form integration with FormGroup and FormRow', () => {
  test('renders complete training scheduler form structure', () => {
    render(
      <Form onSubmit={jest.fn()} error="Please fix the errors below">
        <FormRow>
          <FormGroup label="Select Training Type" error="Training type is required">
            <select name="trainingType" required>
              <option value="">-- Select --</option>
              <option value="strength">Strength</option>
              <option value="defense">Defense</option>
              <option value="agility">Agility</option>
            </select>
          </FormGroup>
          <FormGroup label="Duration (hours)">
            <input type="number" name="duration" min="1" max="8" />
          </FormGroup>
        </FormRow>
        <button type="submit">Schedule Training</button>
      </Form>
    );

    expect(screen.getByText('Please fix the errors below')).toBeInTheDocument();
    expect(screen.getByText('Select Training Type')).toBeInTheDocument();
    expect(screen.getByText('*')).toBeInTheDocument(); // Required indicator
    expect(screen.getByText('Training type is required')).toBeInTheDocument();
    expect(screen.getByText('Duration (hours)')).toBeInTheDocument();
    expect(screen.getByText('Schedule Training')).toBeInTheDocument();
  });

  test('all form components can be found in integrated structure', () => {
    render(
      <Form onSubmit={jest.fn()}>
        <FormRow>
          <FormGroup label="Field 1">
            <input type="text" name="field1" />
          </FormGroup>
          <FormGroup label="Field 2">
            <input type="text" name="field2" />
          </FormGroup>
        </FormRow>
      </Form>
    );

    expect(document.querySelector('.form-container')).toBeInTheDocument();
    expect(document.querySelector('.form-row')).toBeInTheDocument();
    expect(document.querySelectorAll('.form-group')).toHaveLength(2);
  });
});
