import React, { useState, useEffect } from 'react';
import './StatProgressionChart.css';
import { fetchStatProgression } from '../api/training';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const StatProgressionChart = ({ boxerId }) => {
  const [data, setData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [hoveredPoint, setHoveredPoint] = useState(null);

  useEffect(() => {
    loadStatProgression();
  }, [boxerId]);

  const loadStatProgression = async () => {
    if (!boxerId) return;

    setIsLoading(true);
    setError(null);

    try {
      const progressionData = await fetchStatProgression(boxerId);
      setData(progressionData);
    } catch (err) {
      setError(getUserErrorMessage(err, 'training', { context: 'progression' }));
    } finally {
      setIsLoading(false);
    }
  };

  // Chart dimensions
  const chartWidth = 600;
  const chartHeight = 300;
  const padding = { top: 20, right: 20, bottom: 40, left: 60 };
  const chartInnerWidth = chartWidth - padding.left - padding.right;
  const chartInnerHeight = chartHeight - padding.top - padding.bottom;

  // Stat configuration
  const statConfig = {
    strength: {
      color: '#fc8181',
      label: 'STR',
      fullLabel: 'Strength'
    },
    defense: {
      color: '#63b3ed',
      label: 'DEF',
      fullLabel: 'Defense'
    },
    agility: {
      color: '#f6ad55',
      label: 'AGI',
      fullLabel: 'Agility'
    }
  };

  const handleMouseLeave = () => {
    setHoveredPoint(null);
  };

  if (isLoading) {
    return (
      <div className="stat-progression-chart">
        <div className="chart-loading">Loading stat progression...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="stat-progression-chart">
        <ErrorDisplay error={error} variant="alert" title="Failed to Load Chart" />
      </div>
    );
  }

  if (!data || !data.points || data.points.length === 0) {
    return (
      <div className="stat-progression-chart">
        <div className="chart-empty-state">
          <div className="empty-icon">📊</div>
          <h4>No Stat Data Available</h4>
          <p>Stat progression data will appear after your boxer completes training sessions.</p>
        </div>
      </div>
    );
  }

  // Calculate scales
  const allStats = data.points.flatMap(p => [
    p.strength || 0,
    p.defense || 0,
    p.agility || 0
  ]);

  const maxStat = Math.max(...allStats, 20); // Minimum scale of 20
  const minStat = 0;
  const statRange = maxStat - minStat;

  const getX = (index) => padding.left + (index / (data.points.length - 1 || 1)) * chartInnerWidth;
  const getY = (statValue) => padding.top + chartInnerHeight - ((statValue - minStat) / statRange) * chartInnerHeight;

  // Generate path data for each stat
  const generatePath = (statKey) => {
    return data.points.map((point, index) => {
      const value = point[statKey] || 0;
      const x = getX(index);
      const y = getY(value);
      return `${index === 0 ? 'M' : 'L'} ${x} ${y}`;
    }).join(' ');
  };

  // Generate area path (for gradient fill under lines)
  const generateAreaPath = (statKey) => {
    const linePath = generatePath(statKey);
    const firstX = getX(0);
    const lastX = getX(data.points.length - 1);
    const bottomY = padding.top + chartInnerHeight;

    return `${linePath} L ${lastX} ${bottomY} L ${firstX} ${bottomY} Z`;
  };

  return (
    <div className="stat-progression-chart">
      <svg
        width={chartWidth}
        height={chartHeight}
        className="progression-svg"
        onMouseLeave={handleMouseLeave}
      >
        <defs>
          {/* Gradients for areas under lines */}
          <linearGradient id="strengthGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#fc8181" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#fc8181" stopOpacity="0.0" />
          </linearGradient>
          <linearGradient id="defenseGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#63b3ed" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#63b3ed" stopOpacity="0.0" />
          </linearGradient>
          <linearGradient id="agilityGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#f6ad55" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#f6ad55" stopOpacity="0.0" />
          </linearGradient>

          {/* Tooltip circle marker */}
          <circle id="hoverMarker" r="6" fill="white" stroke="#000" strokeWidth="2" />
        </defs>

        {/* Grid lines (horizontal) */}
        {[0, 0.25, 0.5, 0.75, 1].map((ratio, i) => {
          const y = padding.top + chartInnerHeight - (ratio * chartInnerHeight);
          const statValue = minStat + ratio * statRange;
          return (
            <g key={i}>
              <line
                x1={padding.left}
                y1={y}
                x2={chartWidth - padding.right}
                y2={y}
                stroke="rgba(255, 255, 255, 0.1)"
                strokeWidth="1"
                strokeDasharray="4,4"
              />
              <text
                x={padding.left - 10}
                y={y + 4}
                textAnchor="end"
                className="grid-label"
              >
                {Math.round(statValue)}
              </text>
            </g>
          );
        })}

        {/* Area fills under lines */}
        <path d={generateAreaPath('strength')} fill="url(#strengthGradient)" />
        <path d={generateAreaPath('defense')} fill="url(#defenseGradient)" />
        <path d={generateAreaPath('agility')} fill="url(#agilityGradient)" />

        {/* Lines for each stat */}
        <path
          d={generatePath('strength')}
          fill="none"
          stroke={statConfig.strength.color}
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="stat-line"
        />
        <path
          d={generatePath('defense')}
          fill="none"
          stroke={statConfig.defense.color}
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="stat-line"
        />
        <path
          d={generatePath('agility')}
          fill="none"
          stroke={statConfig.agility.color}
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="stat-line"
        />

        {/* Data points (circles) */}
        {data.points.map((point, index) => {
          const strengthX = getX(index);
          const strengthY = getY(point.strength || 0);
          const defenseX = getX(index);
          const defenseY = getY(point.defense || 0);
          const agilityX = getX(index);
          const agilityY = getY(point.agility || 0);

          return (
            <g key={index}>
              {/* Strength point */}
              <circle
                cx={strengthX}
                cy={strengthY}
                r="4"
                fill={statConfig.strength.color}
                stroke="white"
                strokeWidth="2"
                className="data-point"
                onMouseEnter={(e) => setHoveredPoint({ index, stat: 'strength', value: point.strength, x: strengthX, y: strengthY, date: point.date })}
              />
              {/* Defense point */}
              <circle
                cx={defenseX}
                cy={defenseY}
                r="4"
                fill={statConfig.defense.color}
                stroke="white"
                strokeWidth="2"
                className="data-point"
                onMouseEnter={(e) => setHoveredPoint({ index, stat: 'defense', value: point.defense, x: defenseX, y: defenseY, date: point.date })}
              />
              {/* Agility point */}
              <circle
                cx={agilityX}
                cy={agilityY}
                r="4"
                fill={statConfig.agility.color}
                stroke="white"
                strokeWidth="2"
                className="data-point"
                onMouseEnter={(e) => setHoveredPoint({ index, stat: 'agility', value: point.agility, x: agilityX, y: agilityY, date: point.date })}
              />
            </g>
          );
        })}

        {/* X-axis labels (dates at start and end) */}
        {data.points.length > 0 && (
          <>
            <text
              x={getX(0)}
              y={chartHeight - padding.bottom + 20}
              textAnchor="middle"
              className="x-label"
            >
              {new Date(data.points[0].date).toLocaleDateString()}
            </text>
            <text
              x={getX(data.points.length - 1)}
              y={chartHeight - padding.bottom + 20}
              textAnchor="middle"
              className="x-label"
            >
              {new Date(data.points[data.points.length - 1].date).toLocaleDateString()}
            </text>
          </>
        )}

        {/* Tooltip (shown on hover) */}
        {hoveredPoint && (
          <g>
            <circle
              cx={hoveredPoint.x}
              cy={hoveredPoint.y}
              r="8"
              fill="none"
              stroke="rgba(255, 255, 255, 0.5)"
              strokeWidth="2"
              className="tooltip-target"
            />
          </g>
        )}
      </svg>

      {/* Tooltip overlay (HTML for better formatting) */}
      {hoveredPoint && (
        <div
          className="chart-tooltip"
          style={{ left: hoveredPoint.x, top: hoveredPoint.y - 50 }}
        >
          <div className="tooltip-header">
            <strong>{statConfig[hoveredPoint.stat].fullLabel}</strong>
            <span>{new Date(hoveredPoint.date).toLocaleDateString()}</span>
          </div>
          <div className="tooltip-value">
            {statConfig[hoveredPoint.stat].label}: {Math.round(hoveredPoint.value)}
          </div>
        </div>
      )}

      {/* Legend */}
      <div className="chart-legend">
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: statConfig.strength.color }}></span>
          <span>{statConfig.strength.fullLabel}</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: statConfig.defense.color }}></span>
          <span>{statConfig.defense.fullLabel}</span>
        </div>
        <div className="legend-item">
          <span className="legend-color" style={{ backgroundColor: statConfig.agility.color }}></span>
          <span>{statConfig.agility.fullLabel}</span>
        </div>
      </div>
    </div>
  );
};

export default StatProgressionChart;
