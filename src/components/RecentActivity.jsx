/**
 * RecentActivity - Combined activity timeline (MAT-148)
 *
 * Merges fight history and training history into a single chronological view.
 * Uses Option A: Fetch separately and merge in frontend.
 */

import React, { useState, useEffect } from 'react';
import { fetchFightHistory } from '../api/fights';
import { fetchTrainingHistory } from '../api/training';
import './BoxerProfile.css';

const RecentActivity = ({ boxerId, gameTime }) => {
  const [activities, setActivities] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadRecentActivity();
  }, [boxerId, gameTime]);

  const loadRecentActivity = async () => {
    if (!boxerId) {
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Fetch fights and training in parallel (Option A)
      const [fightsData, trainingData] = await Promise.all([
        fetchFightHistory(boxerId).catch(err => {
          console.error('Failed to fetch fight history:', err);
          return [];
        }),
        fetchTrainingHistory(boxerId).catch(err => {
          console.error('Failed to fetch training history:', err);
          return [];
        })
      ]);

      // Transform fights into activity format
      const fightActivities = fightsData.map(fight => ({
        type: 'fight',
        id: fight.id,
        date: fight.end_time || fight.scheduled_time,
        title: getFightTitle(fight),
        description: getFightDescription(fight.data),
        icon: getFightIcon(fight.winner_id, fight.boxer1_id, fight.boxer2_id, fight.data),
        colorClass: getFightColorClass(fight.winner_id, fight.boxer1_id, fight.boxer2_id),
        outcome: deriveOutcome(fight.winner_id, fight.boxer1_id, fight.boxer2_id, fight.status),
        opponentName: fight.opponent_name,
        rounds: fight.round || '?',
        isScheduled: fight.status === 'scheduled' && !fight.end_time // Scheduled but not completed
      }));

      // Transform training into activity format
      const trainingActivities = trainingData.map(session => ({
        type: 'training',
        id: session.id,
        date: session.completed_at || session.created_at,
        title: getTrainingTitle(session),
        description: getTrainingDescription(session),
        icon: getTrainingIcon(session.training_type_id),
        colorClass: 'bg-activity-training',
        gains: session.gains || {}
      }));

      // Combine and sort by date (most recent first)
      const combined = [...fightActivities, ...trainingActivities]
        .sort((a, b) => new Date(b.date) - new Date(a.date))
        .slice(0, 10); // Limit to last 10 activities

      setActivities(combined);
    } catch (err) {
      console.error('Failed to load recent activity:', err);
      setError('Failed to load activity history');
    } finally {
      setIsLoading(false);
    }
  };

  // Derive outcome from winner_id and boxer IDs
  const deriveOutcome = (winnerId, boxer1Id, boxer2Id) => {
    if (!winnerId) return null;
    if (winnerId === boxer1Id) return 'win';
    if (winnerId === boxer2Id) return 'loss';
    return 'draw';
  };

  // Helper functions for fight activities
  const getFightIcon = (winnerId, boxer1Id, boxer2Id, data) => {
    const outcome = deriveOutcome(winnerId, boxer1Id, boxer2Id);
    switch (outcome) {
      case 'win': return '🥊';
      case 'loss': return '☠️';
      case 'draw': return '🤝';
      default: return '📅'; // Scheduled/upcoming fight
    }
  };

  const getFightColorClass = (winnerId, boxer1Id, boxer2Id) => {
    const outcome = deriveOutcome(winnerId, boxer1Id, boxer2Id);
    switch (outcome) {
      case 'win': return 'bg-activity-win';
      case 'loss': return 'bg-activity-loss';
      case 'draw': return 'bg-activity-draw';
      default: return 'bg-activity-upcoming';
    }
  };

  const getFightTitle = (fight) => {
    const outcome = deriveOutcome(fight.winner_id, fight.boxer1_id, fight.boxer2_id);
    const opponent = fight.opponent_name || 'Opponent';
    const rounds = fight.round || '?';

    if (outcome === null && fight.status === 'scheduled') {
      return `vs "${opponent}"`;
    }

    switch (outcome) {
      case 'win':
        return `Defeated "${opponent}" in ${rounds} round${rounds !== '1' ? 's' : ''}`;
      case 'loss':
        return `Lost to "${opponent}" in ${rounds} round${rounds !== '1' ? 's' : ''}`;
      case 'draw':
        return `Drew with "${opponent}" after ${rounds} round${rounds !== '1' ? 's' : ''}`;
      default:
        return `Fought "${opponent}"`;
    }
  };

  const getFightDescription = (data) => {
    // Check for KO in fight data (JSONB field from backend)
    if (data && data.knockout_round) {
      return `KO at Round ${data.knockout_round}`;
    }
    return null;
  };

  // Helper functions for training activities
  const getTrainingIcon = (typeId) => {
    const icons = {
      1: '💪', // strength
      2: '⚡', // speed
      3: '🛡️', // defense
      4: '🥊', // technique
      5: '🏃'  // endurance
    };
    return icons[typeId] || '🏋️';
  };

  const getTrainingTitle = (session) => {
    const typeNames = {
      1: 'Strength Training',
      2: 'Speed Training',
      3: 'Defense Training',
      4: 'Technique Training',
      5: 'Endurance Training'
    };
    return typeNames[session.training_type_id] || 'Training Session';
  };

  const getTrainingDescription = (session) => {
    const gains = session.gains || {};
    const parts = [];

    if (gains.strength) parts.push(`+${gains.strength} STR`);
    if (gains.defense) parts.push(`+${gains.defense} DEF`);
    if (gains.agility) parts.push(`+${gains.agility} AGI`);

    if (parts.length > 0) {
      return `${parts.join(', ')}`;
    }
    return null;
  };

  // Format relative time using game time for scheduled events
  const formatTimeAgo = (dateString, isFutureEvent = false) => {
    if (!dateString) return '';

    const date = new Date(dateString);

    // Use game time if available, otherwise fall back to real time
    const referenceTime = gameTime ? new Date(gameTime) : new Date();

    // For future events (scheduled fights), show "in X days/hours"
    if (isFutureEvent && date > referenceTime) {
      const diffMs = date - referenceTime;
      const diffHours = Math.floor(diffMs / 3600000);
      const diffDays = Math.floor(diffMs / 86400000);

      if (diffDays > 0) {
        return `in ${diffDays} day${diffDays !== 1 ? 's' : ''}`;
      } else if (diffHours > 0) {
        return `in ${diffHours} hour${diffHours !== 1 ? 's' : ''}`;
      } else {
        return 'Today';
      }
    }

    // For past events, show "X days/hours ago"
    const diffMs = referenceTime - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays}d ago`;

    // For older items, show date
    const options = { month: 'short', day: 'numeric' };
    return date.toLocaleDateString('en-US', options);
  };

  if (isLoading) {
    return (
      <div className="recent-activity-section">
        <h3 className="section-title">Recent Activity</h3>
        <div className="activity-loading">
          <div className="activity-skeleton"></div>
          <div className="activity-skeleton"></div>
          <div className="activity-skeleton"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="recent-activity-section">
        <h3 className="section-title">Recent Activity</h3>
        <div className="activity-error">{error}</div>
      </div>
    );
  }

  if (activities.length === 0) {
    return (
      <div className="recent-activity-section">
        <h3 className="section-title">Recent Activity</h3>
        <div className="activity-empty">
          <p>No activity yet. Start training or book a fight!</p>
        </div>
      </div>
    );
  }

  return (
    <div className="recent-activity-section">
      <h3 className="section-title">Recent Activity</h3>

      <div className="activity-timeline">
        {activities.map((activity, index) => (
          <div
            key={`${activity.type}-${activity.id}`}
            className={`activity-item ${activity.colorClass || ''} ${activity.isScheduled ? 'activity-upcoming' : ''}`}
          >
            <div className="activity-icon">{activity.icon}</div>
            <div className="activity-content">
              <div className="activity-header">
                <span className="activity-time">{formatTimeAgo(activity.date, activity.isScheduled)}</span>
                <span className="activity-title">{activity.title}</span>
              </div>
              {activity.description && (
                <div className="activity-description">{activity.description}</div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default RecentActivity;
