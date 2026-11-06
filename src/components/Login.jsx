/**
 * Login - Authentication Entry Point
 *
 * Redesigned login experience with:
 * - Dramatic navy gradient background matching premium content pages
 * - Boxing ring rope decoration for thematic consistency
 * - Refined form design using shared Input and Button components (MAT-150)
 * - Smooth animations and loading states
 * - Enhanced error handling with visual feedback
 * - Register link for new users
 *
 * @see MAT-146: Design Tokens Foundation
 * @see MAT-150: Shared UI Component Library
 * @see Auth.css: Supporting styles
 */

import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import './Auth.css';
import { login } from '../api/users';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const Login = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Clear error on attempt if we have valid input
    if (username.trim() && password.trim()) {
      setError(null);
      setIsLoading(true);
    }

    try {
      const data = await login({ username, password });
      // Pass token and user data to parent for storage and redirect
      onLogin(data.user, data.accessToken || data.token);
    } catch (err) {
      setError(getUserErrorMessage(err, 'auth'));
      setIsLoading(false);
    }
  };

  const handleInputChange = (field) => (e) => {
    const newValue = e.target.value;

    if (field === 'username') {
      setUsername(newValue);
    } else if (field === 'password') {
      setPassword(newValue);
    }

    // Clear error on typing for better UX
    if (error && username.trim() && password.trim()) {
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
            <h1 className="app-name">Boxing Simulator</h1>
            <p className="tagline">Train. Schedule. Conquer. Own the Ring.</p>
          </div>

          {/* Error Display */}
          {error && (
            <ErrorDisplay
              error={error}
              variant="alert"
              icon={true}
              title="Login Failed"
              onClose={() => setError(null)}
            />
          )}

          {/* Login Form */}
          <form className="auth-form" onSubmit={handleSubmit} noValidate>
            <div className="form-field">
              <label htmlFor="username">
                Username
                <span className="required-indicator">*</span>
              </label>
              <input
                type="text"
                id="username"
                name="username"
                placeholder="Enter your username"
                value={username}
                onChange={handleInputChange('username')}
                disabled={isLoading}
                autoComplete="username"
                autoFocus
                className={error && !username.trim() ? 'input-error' : ''}
              />
            </div>

            <div className="form-field">
              <label htmlFor="password">
                Password
                <span className="required-indicator">*</span>
              </label>
              <input
                type="password"
                id="password"
                name="password"
                placeholder="Enter your password"
                value={password}
                onChange={handleInputChange('password')}
                disabled={isLoading}
                autoComplete="current-password"
                className={error && !password.trim() ? 'input-error' : ''}
              />
            </div>

            <button
              type="submit"
              className="auth-submit-btn"
              disabled={isLoading || !username.trim() || !password.trim()}
            >
              {isLoading ? (
                <>
                  <span className="btn-spinner"></span>
                  Signing in...
                </>
              ) : (
                'Sign In'
              )}
            </button>

            {/* Divider */}
            <div className="auth-divider">
              <span>New to the Ring?</span>
            </div>

            {/* Register Link */}
            <p className="auth-footer-text">
              Create your account to manage your boxing career
            </p>
            <Link
              to="/register"
              className="auth-register-link"
              onClick={(e) => {
                if (isLoading) e.preventDefault();
              }}
            >
              Create Account →
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

export default Login;
