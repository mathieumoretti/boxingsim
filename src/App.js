import React, { useState, useEffect } from 'react';
//import './App.css';
import Login from './components/Login';
import Register from './components/Register';
import Dashboard from './components/Dashboard';
import CreateBoxer from './components/CreateBoxer';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';

// API base URL - will be proxied by webpack dev server
const API_BASE_URL = '/api';

function App() {
  const [currentUser, setCurrentUser] = useState(null);
  const [token, setToken] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
      setToken(storedToken);
      // Validate token by fetching user info
      fetchUser();
    } else {
      setIsLoading(false);
    }
  }, []);

  const fetchUser = async () => {
    try {
      // For now, we'll just use mock user data since the real endpoint isn't implemented
      // In a real implementation, this would be a call to fetch user details from backend
      const userData = {
        id: 1,
        username: "testuser",
        email: "user@example.com"
      };

      setCurrentUser(userData);
      setIsLoading(false);
    } catch (error) {
      localStorage.removeItem('token');
      setToken(null);
      setIsLoading(false);
    }
  };

  const handleLogin = (userData, tokenData) => {
    setCurrentUser(userData);
    setToken(tokenData);
    localStorage.setItem('token', tokenData);
    // Redirect to dashboard after successful login
    // Using window.location.replace to avoid going back in browser history
    window.location.replace('/dashboard');
  };

  const handleLogout = () => {
    setCurrentUser(null);
    setToken(null);
    localStorage.removeItem('token');
  };

  const ProtectedRoute = ({ children }) => {
    return currentUser ? children : <Navigate to="/login" replace />;
  };

  if (isLoading) {
    return <div className="loading">Loading...</div>;
  }

  return (
    <div className="app">
      <Routes>
        <Route path="/login" element={<Login onLogin={handleLogin} />} />
        <Route path="/register" element={<Register />} />

        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <Dashboard user={currentUser} onLogout={handleLogout} />
            </ProtectedRoute>
          }
        />

        <Route
          path="/create-boxer"
          element={
            <ProtectedRoute>
              <CreateBoxer user={currentUser} />
            </ProtectedRoute>
          }
        />

        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </div>
  );
}

export default App;