import React, { useState, useEffect } from 'react';
import { fetchOpponents, getMatchQuality } from '../api/opponents';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const OpponentList = ({ boxerId, boxer, onSelectOpponent, filters = {} }) => {
  const [opponents, setOpponents] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadOpponents();
  }, [boxerId, filters]);

  const loadOpponents = async (isRetry = false) => {
    if (!boxerId) return;

    setIsLoading(true);
    if (isRetry) {
      setError(null);
    }

    try {
      // Build query parameters from filters
      // Backend uses level_range as ± value (e.g., "3" means boxer's level ±3)

      const filterParams = {
        availableOnly: true,
        ...(filters.levelRange !== undefined ? { minLevel: filters.levelRange } : {}),
        ...(filters.minHealth ? { minHealth: filters.minHealth } : {}),
        limit: filters.limit || 50
      };

      console.log('[OpponentList] Fetching opponents for boxer', boxerId, 'with filters:', filterParams);
      const data = await fetchOpponents(boxerId, filterParams);
      console.log('[OpponentList] Received', data.length, 'opponents');
      setOpponents(data || []);
    } catch (err) {
      const errorMsg = getUserErrorMessage(err, 'opponents', { boxerName: boxer?.name });
      setError(errorMsg);
      console.error('[OpponentList] Failed to load opponents:', err);
    } finally {
      setIsLoading(false);
    }
  };

  // Retry handler
  const handleRetry = () => {
    loadOpponents(true);
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
    const quality = getMatchQuality(score * 100);
    return quality.label;
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

  // Tailwind class builders for match quality colors
  const getCardBorderClass = (quality) => {
    switch (quality) {
      case 'excellent': return 'border-success';
      case 'good': return 'border-green-400';
      case 'fair': return 'border-warning';
      default: return 'border-white/30';
    }
  };

  const getBarColorClass = (quality) => {
    switch (quality) {
      case 'excellent': return 'bg-gradient-to-r from-success to-emerald-500';
      case 'good': return 'bg-gradient-to-r from-green-400 to-emerald-400';
      case 'fair': return 'bg-gradient-to-r from-warning to-yellow-400';
      default: return 'bg-gradient-to-r from-gray-400 to-gray-500';
    }
  };

  const getButtonClass = (quality) => {
    const base = 'w-full py-2.5 px-4 text-sm font-semibold rounded-md border-none transition-all duration-150 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2';
    switch (quality) {
      case 'excellent': return `${base} bg-gradient-to-br from-success to-emerald-600 text-white hover:bg-gradient-to-br hover:from-success hover:to-emerald-500 hover:-translate-y-0.5 hover:shadow-[0_4px_8px_rgba(34,197,94,0.3)] focus-visible:outline-success`;
      case 'good': return `${base} bg-gradient-to-br from-green-400 to-emerald-500 text-white hover:bg-gradient-to-br hover:from-green-300 hover:to-emerald-400 hover:-translate-y-0.5 focus-visible:outline-green-400`;
      case 'fair': return `${base} bg-gradient-to-br from-warning to-yellow-500 text-black hover:bg-gradient-to-br hover:from-warning hover:to-yellow-400 hover:-translate-y-0.5 focus-visible:outline-warning`;
      default: return `${base} bg-gradient-to-br from-gray-400 to-gray-500 text-white hover:bg-gradient-to-br hover:from-gray-300 hover:to-gray-400 hover:-translate-y-0.5 focus-visible:outline-gray-400`;
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-3">
        {[1, 2, 3].map(i => (
          <div key={i} className="bg-gradient-to-br from-white/10 to-white/5 rounded-lg border border-white/10 p-4 animate-pulse">
            <div className="h-5 bg-white/15 rounded w-3/4 mb-2"></div>
            <div className="h-4 bg-white/15 rounded w-1/2 mb-3"></div>
            <div className="h-8 bg-white/15 rounded w-full"></div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <ErrorDisplay
        error={error}
        onRetry={handleRetry}
        variant="compact"
        title="No Opponents"
      />
    );
  }

  if (opponents.length === 0) {
    return (
      <div className="text-center py-12 bg-gradient-to-br from-white/10 to-white/5 rounded-lg border border-white/10">
        <p className="m-0 text-base text-white/60 mb-2">No opponents found matching your criteria.</p>
        <p className="m-0 text-sm text-white/50 italic">Try adjusting the level range or health filters.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h4 className="m-0 text-lg font-semibold text-white">Available Opponents ({opponents.length})</h4>
      </div>

      <div className="space-y-3">
        {opponents.map((opportunity, index) => {
          const boxerData = opportunity.opponent?.boxer || {};
          const scoreData = opportunity.score || {};
          const overallScore = scoreData.overall_score || 0;
          const quality = getMatchQualityColor(overallScore);

          return (
            <div
              key={boxerData.id || index}
              className={`bg-gradient-to-br from-white/10 to-white/5 rounded-lg border ${getCardBorderClass(quality)} p-5 transition-all duration-150 hover:bg-white/15`}
              data-match-quality={overallScore}
              title={opportunity.reasoning || ''}
            >
              {isBestMatch(opportunity) && (
                <div className="absolute -top-2 -right-2 inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-gradient-to-r from-victory-gold to-warning text-black shadow-[0_2px_8px_rgba(214,158,46,0.3)]">
                  <span>⭐</span> Best Match
                </div>
              )}

              <div className="flex items-start justify-between mb-3">
                <div>
                  <h5 className="m-0 text-base font-bold text-white">{boxerData.name || 'Unknown'}</h5>
                  <span className="text-sm text-white/60">Level {boxerData.level || '?'}</span>
                </div>
              </div>

              <div className="flex flex-wrap gap-3 mb-4 text-sm">
                <div className="flex items-center gap-1.5">
                  <span className="text-white/50 font-semibold">Rank:</span>
                  <span className="font-bold text-white">
                    {opportunity.opponent?.rank || opportunity.opponent?.ranking_position
                      ? `#${opportunity.opponent.rank || opportunity.opponent.ranking_position}`
                      : 'Unranked'}
                  </span>
                </div>

                <div className="flex items-center gap-1.5">
                  <span className="text-white/50 font-semibold">Level Diff:</span>
                  <span className={`font-bold ${levelDiff(opportunity) === 0 ? 'text-success' : 'text-white'}`}>
                    {levelDiff(opportunity)}
                  </span>
                </div>

                {opportunity.opponent?.is_ai && (
                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold bg-cyan-500/10 border border-cyan-400/30 text-cyan-200">
                    AI
                  </span>
                )}
              </div>

              {/* Match Quality Score */}
              <div className="mb-4">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-xs font-semibold text-white/60">Match Quality</span>
                  <span className="text-sm font-bold text-white">{Math.round(overallScore * 100)}%</span>
                </div>
                <div className="h-2 bg-white/15 rounded-full overflow-hidden">
                  <div
                    className={`h-full ${getBarColorClass(quality)} transition-all duration-300`}
                    style={{ width: `${overallScore * 100}%` }}
                  ></div>
                </div>
              </div>

              {/* Score Breakdown (optional detail) */}
              <div className="flex flex-wrap gap-2 mb-4 text-xs">
                <span className="px-2 py-1 bg-white/10 rounded text-white/70 font-semibold">Level: {Math.round((scoreData.level_proximity || scoreData.level_score || 0) * 100)}%</span>
                <span className="px-2 py-1 bg-white/10 rounded text-white/70 font-semibold">Rank: {Math.round((scoreData.ranking_diff || scoreData.ranking_score || 0) * 100)}%</span>
                <span className="px-2 py-1 bg-white/10 rounded text-white/70 font-semibold">Power: {Math.round((scoreData.win_rate_match || scoreData.power_score || 0) * 100)}%</span>
              </div>

              {/* Select Button */}
              <button
                className={getButtonClass(quality)}
                onClick={() => onSelectOpponent(opportunity)}
                aria-label={`Book fight against ${boxerData.name}`}
              >
                Book Fight
              </button>

              {/* Reasoning Tooltip (visible on card) */}
              {opportunity.reasoning && (
                <div className="mt-3 pt-3 border-t border-white/10 text-xs italic text-white/50">
                  {opportunity.reasoning}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default OpponentList;
