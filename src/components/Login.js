import React, { useState } from 'react';
import './Auth.css';
import { login } from '../api/users';
import { ErrorDisplay, FieldError } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const Login = ({ onLogin }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();

    // Only clear error if we're not already loading and have valid input
    if (!isLoading && username.trim() && password.trim()) {
      setError(null);
      setIsLoading(true);
    }

    try {
      const data = await login({ username, password });
      // Pass the token and user data to the parent component
      // The parent will handle storing it in localStorage
      onLogin(data.user, data.accessToken || data.token);
    } catch (err) {
      // Use specific error message from utility
      setError(getUserErrorMessage(err, 'auth'));
      setIsLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-form">
        <h2>Login</h2>
        {error && <ErrorDisplay error={error} variant="alert" icon={true} title="Login Failed" />}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <input
              type="text"
              id="username"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <input
              type="password"
              id="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit" disabled={isLoading}>
            {isLoading ? 'Logging in...' : 'Login'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default Login;
