/**
 * Register - User Registration Page
 *
 * Redesigned registration experience matching the Login page aesthetic:
 * - Dramatic navy gradient background for thematic consistency
 * - Boxing ring rope decoration
 * - Refined form design with smooth animations
 * - Enhanced error handling and feedback
 * - Link back to Login for existing users
 *
 * @see MAT-146: Design Tokens Foundation
 * @see MAT-150: Shared UI Component Library
 * @see Auth.css: Supporting styles
 */

import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import './Auth.css';
import { registerUser } from '../api/users';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const Register = () => {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Basic validation
    if (!username.trim() || !email.trim() || !password.trim()) {
      return;
    }

    setError(null);
    setIsLoading(true);

    try {
      await registerUser({ username, email, password });
      // Registration successful - redirect to login with a flash message opportunity
      window.location.href = '/login';
    } catch (err) {
      setError(getUserErrorMessage(err, 'auth'));
      setIsLoading(false);
    }
  };

  const handleInputChange = (field) => (e) => {
    const newValue = e.target.value;

    if (field === 'username') {
      setUsername(newValue);
    } else if (field === 'email') {
      setEmail(newValue);
    } else if (field === 'password') {
      setPassword(newValue);
    }

    // Clear error on typing for better UX
    if (error) {
      setError(null);
    }
  };

  return (
    <div className="auth-page">
      {/* Decorative ring ropes */}
      <div className="ring-ropes">
        <div className="rope rope-top"></div>
        <div className="rope rope-middle"></div>
        <div className="rope rope-bottom"></div>
      </div>

      <div className="auth-container">
        <div className="auth-card">
          {/* Logo/Branding */}
          <div className="auth-branding">
            <div className="boxing-glove-icon">🥊</div>
            <h1 className="app-name">Join the Fight</h1>
            <p className="tagline">Start your boxing career today</p>
          </div>

          {/* Error Display */}
          {error && (
            <ErrorDisplay
              error={error}
              variant="alert"
              icon={true}
              title="Registration Failed"
              onClose={() => setError(null)}
            />
          )}

          {/* Registration Form */}
          <form className="auth-form" onSubmit={handleSubmit} noValidate>
            <div className="form-field">
              <label htmlFor="reg-username">
                Username
                <span className="required-indicator">*</span>
              </label>
              <input
                type="text"
                id="reg-username"
                name="username"
                placeholder="Choose a username"
                value={username}
                onChange={handleInputChange('username')}
                disabled={isLoading}
                autoComplete="username"
                autoFocus
                className={error && !username.trim() ? 'input-error' : ''}
              />
            </div>

            <div className="form-field">
              <label htmlFor="reg-email">
                Email Address
                <span className="required-indicator">*</span>
              </label>
              <input
                type="email"
                id="reg-email"
                name="email"
                placeholder="you@example.com"
                value={email}
                onChange={handleInputChange('email')}
                disabled={isLoading}
                autoComplete="email"
                className={error && !email.trim() ? 'input-error' : ''}
              />
            </div>

            <div className="form-field">
              <label htmlFor="reg-password">
                Password
                <span className="required-indicator">*</span>
              </label>
              <input
                type="password"
                id="reg-password"
                name="password"
                placeholder="Create a strong password"
                value={password}
                onChange={handleInputChange('password')}
                disabled={isLoading}
                autoComplete="new-password"
                className={error && !password.trim() ? 'input-error' : ''}
              />
            </div>

            <button
              type="submit"
              className="auth-submit-btn"
              disabled={isLoading || !username.trim() || !email.trim() || !password.trim()}
            >
              {isLoading ? (
                <>
                  <span className="btn-spinner"></span>
                  Creating Account...
                </>
              ) : (
                'Create Account'
              )}
            </button>

            {/* Divider */}
            <div className="auth-divider">
              <span>Already Have an Account?</span>
            </div>

            {/* Login Link */}
            <Link
              to="/login"
              className="auth-register-link"
              onClick={(e) => {
                if (isLoading) e.preventDefault();
              }}
            >
              ← Sign In Instead
            </Link>
          </form>
        </div>

        {/* Decorative elements */}
        <div className="auth-decoration">
          <div className="corner-accent corner-top-left"></div>
          <div className="corner-accent corner-bottom-right"></div>
        </div>
      </div>
    </div>
  );
};

export default Register;
