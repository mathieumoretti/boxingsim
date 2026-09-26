import React, { useState, useEffect } from 'react';
import './OpponentList.css';
import { API_BASE_URL, authenticatedFetch } from '../utils/auth';

const OpponentList = ({ boxerId, boxer, onSelectOpponent, filters = {} }) => {
  const [opponents, setOpponents] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadOpponents();
  }, [boxerId, filters]);

  const loadOpponents = async () => {
    if (!boxerId) return;

    setIsLoading(true);
    setError('');

    try {
      // Build query parameters from filters
      const params = new URLSearchParams();

      if (filters.levelRange) {
        params.append('level_range', filters.levelRange);
      }
      if (filters.aiOnly !== undefined) {
        params.append('ai_only', filters.aiOnly.toString());
      }
      if (filters.minHealth) {
        params.append('min_health', filters.minHealth.toString());
      }
      if (filters.limit) {
        params.append('limit', filters.limit.toString());
      }

      const response = await authenticatedFetch(`${API_BASE_URL}/boxers/${boxerId}/opponents?${params.toString()}`, {
        method: 'GET'
      });

      if (!response.ok) {
        throw new Error(`Failed to load opponents: ${response.statusText}`);
      }

      const data = await response.json();
      setOpponents(data.opportunities || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const getMatchQualityColor = (score) => {
    // Score is now passed as 0-1 float from API, convert to 0-100 for comparison
    const scorePercent = score * 100;
    if (scorePercent >= 70) return 'excellent';
    if (scorePercent >= 50) return 'good';
    if (scorePercent >= 30) return 'fair';
    return 'poor';
  };

  const getMatchQualityLabel = (score) => {
    const scorePercent = score * 100;
    if (scorePercent >= 70) return 'Excellent Match';
    if (scorePercent >= 50) return 'Good Match';
    if (scorePercent >= 30) return 'Fair Match';
    return 'Risky Match';
  };

  const levelDiff = (opportunity) => {
    if (!boxer || !opportunity?.opponent?.boxer) return 0;
    const diff = opportunity.opponent.boxer.level - boxer.level;
    return diff > 0 ? `+${diff}` : diff;
  };

  const isBestMatch = (opportunity) => {
    // Find the highest scored opponent
    if (opponents.length === 0) return false;
    const bestScore = Math.max(...opponents.map(o => o.score?.overall_score || 0));
    return (opportunity.score?.overall_score || 0) === bestScore;
  };

  if (isLoading) {
    return <div className="opponent-list-loading">Loading opponents...</div>;
  }

  if (error) {
    return <div className="opponent-list-error">Error: {error}</div>;
  }

  if (opponents.length === 0) {
    return (
      <div className="opponent-list-empty">
        <p>No opponents found matching your criteria.</p>
      </div>
    );
  }

  return (
    <div className="opponent-list">
      <div className="opponent-list-header">
        <h4>Available Opponents ({opponents.length})</h4>
      </div>

      <div className="opponent-list-content">
        {opponents.map((opportunity, index) => {
          const boxerData = opportunity.opponent?.boxer || {};
          const scoreData = opportunity.score || {};
          const overallScore = scoreData.overall_score || 0;

          return (
            <div
              key={boxerData.id || index}
              className={`opponent-card ${getMatchQualityColor(overallScore)}`}
              data-match-quality={overallScore}
              title={opportunity.reasoning || ''}
            >
              {isBestMatch(opportunity) && (
                <div className="best-match-badge">⭐ Best Match</div>
              )}

              <div className="opponent-card-header">
                <div className="opponent-name">{boxerData.name || 'Unknown'}</div>
                <div className="opponent-level">Level {boxerData.level || '?'}</div>
              </div>

              <div className="opponent-stats">
                <div className="stat-item">
                  <span className="stat-label">Rank:</span>
                  <span className="stat-value">
                    {opportunity.opponent?.rank || opportunity.opponent?.ranking_position
                      ? `#${opportunity.opponent.rank || opportunity.opponent.ranking_position}`
                      : 'Unranked'}
                  </span>
                </div>

                <div className="stat-item">
                  <span className="stat-label">Level Diff:</span>
                  <span className={`stat-value ${levelDiff(opportunity) === 0 ? 'same-level' : ''}`}>
                    {levelDiff(opportunity)}
                  </span>
                </div>

                {opportunity.opponent?.is_ai && (
                  <div className="stat-item ai-badge">
                    <span className="ai-indicator">AI</span>
                  </div>
                )}
              </div>

              {/* Match Quality Score */}
              <div className="match-quality-section">
                <div className="match-quality-label">Match Quality</div>
                <div className="match-quality-bar-container">
                  <div
                    className={`match-quality-bar ${getMatchQualityColor(overallScore)}`}
                    style={{ width: `${overallScore * 100}%` }}
                  ></div>
                </div>
                <div className="match-quality-score">{Math.round(overallScore * 100)}%</div>
              </div>

              {/* Score Breakdown (optional detail) */}
              <div className="score-breakdown">
                <span className="breakdown-item">Level: {Math.round((scoreData.level_proximity || scoreData.level_score || 0) * 100)}%</span>
                <span className="breakdown-item">Rank: {Math.round((scoreData.ranking_diff || scoreData.ranking_score || 0) * 100)}%</span>
                <span className="breakdown-item">Power: {Math.round((scoreData.win_rate_match || scoreData.power_score || 0) * 100)}%</span>
              </div>

              {/* Select Button */}
              <button
                className={`select-opponent-btn ${getMatchQualityColor(overallScore)}`}
                onClick={() => onSelectOpponent(opportunity)}
              >
                Book Fight
              </button>

              {/* Reasoning Tooltip (visible on card) */}
              {opportunity.reasoning && (
                <div className="match-reasoning">{opportunity.reasoning}</div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default OpponentList;
