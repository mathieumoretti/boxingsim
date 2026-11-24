import React, { useState, useEffect } from 'react';
import { fetchStatProgression } from '../api/training';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const StatProgressionChart = ({ boxerId, compact = false }) => {
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
  const chartWidth = compact ? 500 : 600;
  const chartHeight = compact ? 200 : 300;
  const padding = compact
    ? { top: 10, right: 0, bottom: 0, left: 50 }
    : { top: 20, right: 20, bottom: 40, left: 60 };
  const chartInnerWidth = chartWidth - padding.left - padding.right;
  const chartInnerHeight = chartHeight - padding.top - padding.bottom;

  // Stat configuration with Tailwind colors
  const statConfig = {
    strength: {
      color: '#dc2626', // red-600
      label: 'STR',
      fullLabel: 'Strength'
    },
    defense: {
      color: '#2563eb', // blue-600
      label: 'DEF',
      fullLabel: 'Defense'
    },
    agility: {
      color: '#ca8a04', // yellow-600
      label: 'AGI',
      fullLabel: 'Agility'
    }
  };

  const handleMouseLeave = () => {
    setHoveredPoint(null);
  };

  if (isLoading) {
    return (
      <div className="bg-navy-darkest rounded-xl p-5 md:p-4 shadow-[0_8px_40px_rgba(0,0,0,0.6)] border border-white/10 mb-5">
        <div className="flex items-center justify-center h-[200px] text-gray-500 text-sm font-medium">
          Loading...
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-navy-darkest rounded-xl p-5 md:p-4 shadow-[0_8px_40px_rgba(0,0,0,0.6)] border border-white/10 mb-5">
        <ErrorDisplay error={error} variant="alert" title="Failed to Load Chart" />
      </div>
    );
  }

  if (!data || !data.points || data.points.length === 0) {
    return (
      <div className="bg-navy-darkest rounded-xl p-5 md:p-4 shadow-[0_8px_40px_rgba(0,0,0,0.6)] border border-white/10 mb-5">
        <div className="flex items-center justify-center h-[200px]">
          <p className="text-sm font-medium text-gray-600 m-0">No stat data yet.</p>
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
    <div className="bg-navy-darkest rounded-xl p-5 md:p-4 shadow-[0_8px_40px_rgba(0,0,0,0.6)] border border-white/10 mb-5">
      <svg
        width={chartWidth}
        height={chartHeight}
        className="w-full"
        onMouseLeave={handleMouseLeave}
      >
        <defs>
          {/* Gradients for areas under lines */}
          <linearGradient id="strengthGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#dc2626" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#dc2626" stopOpacity="0.0" />
          </linearGradient>
          <linearGradient id="defenseGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#2563eb" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#2563eb" stopOpacity="0.0" />
          </linearGradient>
          <linearGradient id="agilityGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="#ca8a04" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#ca8a04" stopOpacity="0.0" />
          </linearGradient>

          {/* Drop shadows for lines */}
          <filter id="strengthShadow" x="-50%" y="-50%" width="200%" height="200%">
            <feDropShadow dx="0" dy="0" stdDeviation="2" floodColor="#dc2626" floodOpacity="0.5"/>
          </filter>
          <filter id="defenseShadow" x="-50%" y="-50%" width="200%" height="200%">
            <feDropShadow dx="0" dy="0" stdDeviation="2" floodColor="#2563eb" floodOpacity="0.5"/>
          </filter>
          <filter id="agilityShadow" x="-50%" y="-50%" width="200%" height="200%">
            <feDropShadow dx="0" dy="0" stdDeviation="2" floodColor="#ca8a04" floodOpacity="0.5"/>
          </filter>
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
                fill="#6b7280"
                fontSize={compact ? 9 : 10}
                fontFamily="inherit"
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
          filter="url(#strengthShadow)"
          className="transition-all duration-150 hover:stroke-[4] cursor-crosshair"
        />
        <path
          d={generatePath('defense')}
          fill="none"
          stroke={statConfig.defense.color}
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          filter="url(#defenseShadow)"
          className="transition-all duration-150 hover:stroke-[4] cursor-crosshair"
        />
        <path
          d={generatePath('agility')}
          fill="none"
          stroke={statConfig.agility.color}
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          filter="url(#agilityShadow)"
          className="transition-all duration-150 hover:stroke-[4] cursor-crosshair"
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
                filter="url(#strengthShadow)"
                className="transition-all duration-150 cursor-pointer hover:r-[7] hover:fill-white hover:stroke-[3]"
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
                filter="url(#defenseShadow)"
                className="transition-all duration-150 cursor-pointer hover:r-[7] hover:fill-white hover:stroke-[3]"
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
                filter="url(#agilityShadow)"
                className="transition-all duration-150 cursor-pointer hover:r-[7] hover:fill-white hover:stroke-[3]"
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
              fill="#6b7280"
              fontSize={compact ? 9 : 10}
              fontFamily="inherit"
            >
              {new Date(data.points[0].date).toLocaleDateString()}
            </text>
            <text
              x={getX(data.points.length - 1)}
              y={chartHeight - padding.bottom + 20}
              textAnchor="middle"
              fill="#6b7280"
              fontSize={compact ? 9 : 10}
              fontFamily="inherit"
            >
              {new Date(data.points[data.points.length - 1].date).toLocaleDateString()}
            </text>
          </>
        )}

        {/* Tooltip ring (shown on hover) */}
        {hoveredPoint && (
          <g>
            <circle
              cx={hoveredPoint.x}
              cy={hoveredPoint.y}
              r="8"
              fill="none"
              stroke="rgba(255, 255, 255, 0.5)"
              strokeWidth="2"
            />
          </g>
        )}
      </svg>

      {/* Tooltip overlay (HTML for better formatting) */}
      {hoveredPoint && (
        <div
          className="absolute bg-navy-medium border border-white/10 rounded-md p-3 shadow-[0_8px_24px_rgba(0,0,0,0.6)] pointer-events-none z-[50]"
          style={{
            left: hoveredPoint.x,
            top: hoveredPoint.y - 50,
            transform: 'translateX(-50%)'
          }}
        >
          <div className="flex justify-between items-center mb-2 pb-2 border-b border-white/10">
            <span className="text-xs text-gray-400 font-medium">{new Date(hoveredPoint.date).toLocaleDateString()}</span>
            <strong className="text-xs text-white font-bold uppercase tracking-wider">{statConfig[hoveredPoint.stat].fullLabel}</strong>
          </div>
          <div className="text-sm text-white font-bold">
            {statConfig[hoveredPoint.stat].label}: {Math.round(hoveredPoint.value)}
          </div>
        </div>
      )}

      {/* Legend */}
      <div className={`flex gap-4 md:gap-3 ${compact ? 'text-[0.65rem]' : ''}`}>
        <div className="inline-flex items-center gap-1 text-xs text-gray-400 font-medium">
          <span
            className="w-[8px] h-[8px] rounded-full shadow-[0_0_8px_rgba(220,38,38,0.5)]"
            style={{ backgroundColor: statConfig.strength.color }}
          ></span>
          <span>STR</span>
        </div>
        <div className="inline-flex items-center gap-1 text-xs text-gray-400 font-medium">
          <span
            className="w-[8px] h-[8px] rounded-full shadow-[0_0_8px_rgba(37,99,235,0.5)]"
            style={{ backgroundColor: statConfig.defense.color }}
          ></span>
          <span>DEF</span>
        </div>
        <div className="inline-flex items-center gap-1 text-xs text-gray-400 font-medium">
          <span
            className="w-[8px] h-[8px] rounded-full shadow-[0_0_8px_rgba(202,138,4,0.5)]"
            style={{ backgroundColor: statConfig.agility.color }}
          ></span>
          <span>AGI</span>
        </div>
      </div>
    </div>
  );
};

export default StatProgressionChart;
