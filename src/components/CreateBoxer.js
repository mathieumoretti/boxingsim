import React, { useState } from 'react';
import './CreateBoxer.css';
import { useNavigate } from 'react-router-dom';

const CreateBoxer = ({ user }) => {
  const [name, setName] = useState('');
  const [boxerClass, setBoxerClass] = useState('heavyweight');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    if (!user) {
      setError('Please login first');
      setIsLoading(false);
      return;
    }

    try {
      const response = await fetch('http://localhost:8080/boxers', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          name: name,
          class: boxerClass,
          max_health: 100,
          max_energy: 100,
          strength: 10,
          defense: 10,
          agility: 10,
          skill_points: 0
        })
      });

      const data = await response.json();

      if (response.ok) {
        // Successfully created boxer
        alert('Boxer created successfully');
        navigate('/dashboard');
      } else {
        setError(data.error || 'Failed to create boxer');
      }
    } catch (error) {
      setError('Network error: ' + error.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="create-boxer">
      <h2>Create New Boxer</h2>

      {error && <div className="error-message">{error}</div>}

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