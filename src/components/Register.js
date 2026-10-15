import React, { useState } from 'react';
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

    // Only clear error if we're not already loading and have valid input
    if (!isLoading && username.trim() && email.trim() && password.trim()) {
      setError(null);
      setIsLoading(true);
    }

    try {
      await registerUser({ username, email, password });
      // Registration successful - redirect to login
      window.location.href = '/login';
    } catch (err) {
      // Use specific error message from utility
      setError(getUserErrorMessage(err, 'auth'));
      setIsLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-form">
        <h2>Register</h2>
        {error && <ErrorDisplay error={error} variant="alert" icon={true} title="Registration Failed" />}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <input
              type="text"
              id="reg-username"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <input
              type="email"
              id="reg-email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <input
              type="password"
              id="reg-password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit" disabled={isLoading}>
            {isLoading ? 'Registering...' : 'Register'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default Register;