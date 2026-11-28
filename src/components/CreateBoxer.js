import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';
import TopBar from './TopBar';
import { SkipLink } from './SkipLink';

const CreateBoxer = ({ user, onLogout }) => {
  const handleLogout = () => {
    if (onLogout) {
      onLogout();
    } else {
      window.location.replace('/login');
    }
  };

  const [name, setName] = useState('');
  const [boxerClass, setBoxerClass] = useState('heavyweight');
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);
    setSuccess('');

    // Ensure user is loaded from storage if not passed via props
    const currentUser = user || getUser();
    if (!currentUser) {
      setError('Please login first');
      setIsLoading(false);
      return;
    }

    try {
      const response = await authenticatedFetch(`${API_BASE_URL}/boxers`, {
        method: 'POST',
        body: JSON.stringify({
          name: name,
          nickname: '',
          position_x: 0.0,
          position_y: 0.0,
          strength: 10.0,
          defense: 10.0,
          agility: 10.0,
        })
      });

      const data = await response.json();

      if (response.ok) {
        // Successfully created boxer - show inline success message
        setSuccess('Boxer created successfully! Redirecting...');
        setTimeout(() => {
          navigate('/dashboard');
        }, 1000);
      } else {
        // Use specific error message from utility
        setError(getUserErrorMessage({ message: data.error }, 'boxerCreation'));
      }
    } catch (error) {
      // Use specific error message from utility
      const errorMsg = getUserErrorMessage(error, 'boxerCreation');
      setError(errorMsg);

      // Don't show error if redirected to login (401 handled by authenticatedFetch)
      if (!errorMsg.includes('Unauthorized')) {
        console.error('[CreateBoxer] Error creating boxer:', error);
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
      <SkipLink targetId="main-content" />
      <TopBar currentUser={user} onLogout={handleLogout} />

      <main id="main-content" className="max-w-[500px] mx-auto p-8" tabIndex="-1">
      <h2 className="text-xl font-bold text-white text-center mb-6 m-0">Create New Boxer</h2>

      {error && (
        <ErrorDisplay error={error} variant="alert" icon={true} title="Creation Failed" />
      )}

      {success && (
        <div className="p-4 bg-green-500/10 border border-green-400/30 rounded-lg text-center font-medium mb-4 text-green-200">
          {success}
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="form-group">
          <input
            type="text"
            id="boxer-name"
            placeholder="Boxer Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            className="w-full p-4 border border-white/20 rounded-lg text-base text-white bg-navy-dark transition-all duration-150 focus:outline-none focus:border-primary-light focus:ring-2 focus:ring-primary/30 placeholder:text-white/40"
          />
        </div>

        <div className="form-group">
          <select
            id="boxer-class"
            value={boxerClass}
            onChange={(e) => setBoxerClass(e.target.value)}
            className="w-full p-4 border border-white/20 rounded-lg text-base text-white bg-navy-dark transition-all duration-150 focus:outline-none focus:border-primary-light focus:ring-2 focus:ring-primary/30"
          >
            <option value="heavyweight">Heavyweight</option>
            <option value="middleweight">Middleweight</option>
            <option value="lightweight">Lightweight</option>
            <option value="flyweight">Flyweight</option>
          </select>
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className={`bg-gradient-to-r from-primary to-primary-dark text-white border-none p-4 rounded-lg cursor-pointer text-sm font-semibold transition-all duration-200 mt-2 shadow-[0_2px_8px_rgba(0,0,0,0.04)] hover:from-primary-dark hover:to-primary hover:-translate-y-0.5 active:translate-y-0 disabled:from-gray-300 disabled:to-gray-400 disabled:cursor-not-allowed disabled:opacity-70 focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2`}
        >
          {isLoading ? 'Creating...' : 'Create Boxer'}
        </button>
      </form>
      </main>
    </div>
  );
};

export default CreateBoxer;
