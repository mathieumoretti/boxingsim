import React, { useState, useEffect } from 'react';
//import './App.css';
import Login from './components/Login';
import Register from './components/Register';
import Dashboard from './components/Dashboard';
import CreateBoxer from './components/CreateBoxer';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { getToken, setToken, clearToken, getUser } from './utils/auth';

function App() {
  const [currentUser, setCurrentUser] = useState(null);
  const [token, setTokenState] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in
    const storedToken = getToken();
    const storedUser = getUser();
    if (storedToken && storedUser) {
      setTokenState(storedToken);
      setCurrentUser(storedUser);
    }
    setIsLoading(false);
  }, []);

  const handleLogin = (userData, tokenData) => {
    setCurrentUser(userData);
    setTokenState(tokenData);
    // Store token and user using the centralized helper
    setToken(tokenData, userData);
    // Redirect to dashboard after successful login
    window.location.replace('/dashboard');
  };

  const handleLogout = () => {
    setCurrentUser(null);
    setTokenState(null);
    // Clear token and user using the centralized helper
    clearToken();
    window.location.replace('/login');
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
