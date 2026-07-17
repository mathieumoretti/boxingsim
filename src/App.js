import React, { useState, useEffect } from 'react';
//import './App.css';
import Login from './components/Login';
import Register from './components/Register';
import Dashboard from './components/Dashboard';
import CreateBoxer from './components/CreateBoxer';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';

// API base URL - adjust this to match your backend server
const API_BASE_URL = 'http://localhost:8080';

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
      const response = await fetch(`${API_BASE_URL}/users/1`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const userData = await response.json();
        setCurrentUser(userData);
        setIsLoading(false);
      } else {
        localStorage.removeItem('token');
        setToken(null);
        setIsLoading(false);
      }
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