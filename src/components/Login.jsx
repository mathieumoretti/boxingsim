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
 */

import React, { useState } from 'react';
import { Link } from 'react-router-dom';
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
    <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light relative overflow-hidden flex items-center justify-center p-6">
      {/* Decorative ring ropes - subtle gradient bars */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-[15%] left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>
        <div className="absolute top-[50%] left-0 right-0 h-[3px] bg-gradient-to-r from-transparent via-white/15 to-transparent"></div>
        <div className="absolute top-[85%] left-0 right-0 h-[2px] bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>
      </div>

      {/* Main auth container */}
      <div className="w-full max-w-md relative z-10">
        {/* Auth card with glass morphism */}
        <div className="bg-white/10 backdrop-blur-lg rounded-2xl p-8 shadow-[0_8px_32px_rgba(0,0,0,0.4),inset_0_1px_0_rgba(255,255,255,0.1)] border border-white/20">
          {/* Logo/Branding */}
          <div className="text-center mb-8">
            <div className="text-6xl mb-4 drop-shadow-[0_4px_8px_rgba(0,0,0,0.5)]" aria-hidden="true">🥊</div>
            <h1 className="text-3xl font-extrabold text-white m-0 mb-2 tracking-tight">Boxing Simulator</h1>
            <p className="text-sm text-white/70 m-0 italic">Train. Schedule. Conquer. Own the Ring.</p>
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
          <form onSubmit={handleSubmit} noValidate className="space-y-5">
            <div>
              <label htmlFor="username" className="block text-sm font-semibold text-white mb-2">
                Username
                <span className="text-warning">*</span>
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
                className={`w-full px-4 py-3 rounded-lg text-sm bg-white/10 border-2 transition-all duration-150 placeholder:text-white/40 ${
                  error && !username.trim()
                    ? 'border-danger focus:border-danger focus:shadow-[0_0_0_3px_rgba(239,68,68,0.2)]'
                    : 'border-white/20 text-white focus:border-primary-light focus:shadow-[0_0_0_3px_rgba(124,58,237,0.2)]'
                } focus:outline-none`}
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-semibold text-white mb-2">
                Password
                <span className="text-warning">*</span>
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
                className={`w-full px-4 py-3 rounded-lg text-sm bg-white/10 border-2 transition-all duration-150 placeholder:text-white/40 ${
                  error && !password.trim()
                    ? 'border-danger focus:border-danger focus:shadow-[0_0_0_3px_rgba(239,68,68,0.2)]'
                    : 'border-white/20 text-white focus:border-primary-light focus:shadow-[0_0_0_3px_rgba(124,58,237,0.2)]'
                } focus:outline-none`}
              />
            </div>

            <button
              type="submit"
              disabled={isLoading || !username.trim() || !password.trim()}
              className={`w-full py-3 px-4 rounded-lg text-sm font-semibold border-none transition-all duration-150 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary-light focus-visible:outline-offset-2 ${
                isLoading || !username.trim() || !password.trim()
                  ? 'bg-white/20 text-white/50 cursor-not-allowed'
                  : 'bg-gradient-to-br from-primary to-primary-dark text-white hover:bg-gradient-to-br hover:from-blue-400 hover:to-primary hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(124,58,237,0.4)]'
              }`}
            >
              {isLoading ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                  Signing in...
                </span>
              ) : (
                'Sign In'
              )}
            </button>

            {/* Divider */}
            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-white/20"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-3 bg-transparent text-white/60">New to the Ring?</span>
              </div>
            </div>

            {/* Register Link */}
            <p className="text-center text-sm text-white/70 m-0 mb-3">
              Create your account to manage your boxing career
            </p>
            <Link
              to="/register"
              onClick={(e) => {
                if (isLoading) e.preventDefault();
              }}
              className="block w-full py-3 px-4 rounded-lg text-sm font-semibold bg-white/10 text-white border-2 border-white/20 transition-all duration-150 hover:bg-white/15 hover:border-primary-light hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(124,58,237,0.2)] text-center focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            >
              Create Account →
            </Link>
          </form>
        </div>

        {/* Decorative corner accents */}
        <div className="absolute -top-1 -left-1 w-4 h-4 border-t-2 border-l-2 border-warning/60 rounded-tl-lg"></div>
        <div className="absolute -bottom-1 -right-1 w-4 h-4 border-b-2 border-r-2 border-warning/60 rounded-br-lg"></div>
      </div>
    </div>
  );
};

export default Login;
