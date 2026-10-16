import React, { useState } from 'react';
import './CreateBoxer.css';
import { useNavigate } from 'react-router-dom';
import { API_BASE_URL, authenticatedFetch, getUser } from '../utils/auth';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const CreateBoxer = ({ user }) => {
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
    <div className="create-boxer">
      <h2>Create New Boxer</h2>

      {error && (
        <ErrorDisplay error={error} variant="alert" icon={true} title="Creation Failed" />
      )}

      {success && <div className="success-message">{success}</div>}

      <form onSubmit={handleSubmit} className="boxer-form">
        <div className="form-group">
          <input
            type="text"
            id="boxer-name"
            placeholder="Boxer Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <select
            id="boxer-class"
            value={boxerClass}
            onChange={(e) => setBoxerClass(e.target.value)}
          >
            <option value="heavyweight">Heavyweight</option>
            <option value="middleweight">Middleweight</option>
            <option value="lightweight">Lightweight</option>
            <option value="flyweight">Flyweight</option>
          </select>
        </div>

        <button type="submit" disabled={isLoading}>
          {isLoading ? 'Creating...' : 'Create Boxer'}
        </button>
      </form>
    </div>
  );
};

export default CreateBoxer;
