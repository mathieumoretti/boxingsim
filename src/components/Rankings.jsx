/**
 * Rankings Component (MAT-105)
 *
 * Displays boxer rankings sorted by various criteria with search, filtering,
 * and pagination. Uses the MAT-100 backend API for data retrieval.
 */

import React, { useState, useEffect, useCallback } from 'react';
import TopBar from './TopBar';
import { SkipLink } from './SkipLink';
import RankCard from './RankCard';
import { fetchRankings, RANKING_CRITERIA, getCriteriaDisplayValue } from '../api/rankings';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const DEFAULT_LIMIT = 25;

function Rankings({ user, onLogout }) {
  const handleLogout = () => {
    if (onLogout) {
      onLogout();
    } else {
      window.location.replace('/login');
    }
  };

  // State management as specified in MAT-105
  const [criteria, setCriteria] = useState('win_rate');
  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState({
    aiOnly: false,
    minLevel: 0,
    maxLevel: 100
  });
  const [page, setPage] = useState(1);

  // Data and loading states
  const [rankingsData, setRankingsData] = useState([]);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: DEFAULT_LIMIT,
    total: 0,
    pages: 0
  });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Fetch rankings from backend (MAT-100 API)
  const loadRankings = useCallback(async (currentPage = 1, currentCriteria = criteria, isRetry = false) => {
    setIsLoading(true);
    if (isRetry) {
      setError(null);
    }

    try {
      const result = await fetchRankings({
        criteria: currentCriteria,
        page: currentPage,
        limit: DEFAULT_LIMIT
      });

      setRankingsData(result.boxers || []);
      setPagination({
        page: currentPage,
        limit: DEFAULT_LIMIT,
        total: result.totalCount || 0,
        pages: result.pages || 0
      });

    } catch (err) {
      const errorMsg = getUserErrorMessage(err, 'ranking');
      setError(errorMsg);
      setRankingsData([]);
      console.error('[Rankings] Failed to load rankings:', err);
    } finally {
      setIsLoading(false);
    }
  }, [criteria]);

  // Retry handler
  const handleRetry = useCallback(() => {
    loadRankings(1, criteria, true);
  }, [loadRankings, criteria]);

  // Initial load and when criteria changes
  useEffect(() => {
    loadRankings(1, criteria);
  }, [loadRankings]);

  // Client-side search/filter (applied on top of fetched data)
  const filteredData = React.useMemo(() => {
    return rankingsData.filter(boxer => {
      // Search by name or nickname
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const nameMatch = boxer.name?.toLowerCase().includes(query);
        const nicknameMatch = boxer.nickname?.toLowerCase().includes(query);
        if (!nameMatch && !nicknameMatch) {
          return false;
        }
      }

      // Level range filter
      if (boxer.level < filters.minLevel || boxer.level > filters.maxLevel) {
        return false;
      }

      // AI-only filter (future: when we have owner_id field)
      // For now, this would require backend support
      // if (filters.aiOnly && boxer.owner_id !== null) {
      //   return false;
      // }

      return true;
    });
  }, [rankingsData, searchQuery, filters]);

  // Client-side pagination for filtered results
  const paginatedData = React.useMemo(() => {
    const start = (page - 1) * DEFAULT_LIMIT;
    return filteredData.slice(start, start + DEFAULT_LIMIT);
  }, [filteredData, page]);

  // Reset to page 1 when search or filters change
  const handleSearchChange = (e) => {
    setSearchQuery(e.target.value);
    setPage(1);
  };

  const handleCriteriaChange = (e) => {
    setCriteria(e.target.value);
    setPage(1);
  };

  const handleAiOnlyChange = (e) => {
    setFilters(prev => ({ ...prev, aiOnly: e.target.checked }));
    setPage(1);
  };

  const handleMinLevelChange = (e) => {
    setFilters(prev => ({ ...prev, minLevel: parseInt(e.target.value) || 0 }));
    setPage(1);
  };

  const handleMaxLevelChange = (e) => {
    setFilters(prev => ({ ...prev, maxLevel: parseInt(e.target.value) || 100 }));
    setPage(1);
  };

  const handlePageChange = (newPage) => {
    setPage(newPage);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Criteria label for header
  const criteriaLabel = RANKING_CRITERIA.find(c => c.value === criteria)?.label || 'Rankings';

  // Filter group styles
  const filterGroupBase = 'flex items-center gap-2';
  const filterGroupLarge = `${filterGroupBase} flex-1 min-w-[200px] max-w-[350px]`;

  // Label styles for filters
  const filterLabel = 'text-xs font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap';

  return (
    <div className="min-h-screen bg-gradient-to-br from-navy-dark via-navy-medium to-navy-light">
      <SkipLink targetId="main-content" />
      <TopBar currentUser={user} onLogout={handleLogout} />

      <main id="main-content" className="max-w-[1400px] mx-auto px-6 py-6" tabIndex="-1">
      {/* Header Section */}
      <div className="flex flex-wrap items-center justify-between gap-6 mb-6">
        <div className="flex items-center gap-3">
          <span className="text-[36px] leading-none drop-shadow-[0_2px_4px_rgba(214,158,46,0.4)]" aria-hidden="true">🏆</span>
          <h1 className="text-xl font-extrabold text-white tracking-tight m-0">{criteriaLabel} Rankings</h1>
        </div>

        {/* Search Box */}
        <div className="relative w-[300px] flex-shrink-0">
          <input
            type="text"
            placeholder="Search boxers..."
            value={searchQuery}
            onChange={handleSearchChange}
            className="w-full px-5 py-3 border-2 border-navy-border rounded-md text-sm transition-all duration-150 bg-navy-deep text-white placeholder:text-text-muted focus:outline-none focus:border-primary-light focus:shadow-[0_0_0_3px_rgba(124,58,237,0.2)]"
            aria-label="Search boxers by name or nickname"
          />
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-base opacity-60" aria-hidden="true">🔍</span>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="flex flex-wrap items-center gap-6 px-5 py-4 bg-navy-deep rounded-lg shadow-[0_2px_16px_rgba(0,0,0,0.3),inset_0_1px_0_rgba(255,255,255,0.05)] border border-navy-border mb-4">
        {/* Criteria Selector */}
        <div className={filterGroupBase}>
          <label htmlFor="criteria-select" className={filterLabel}>Sort by:</label>
          <select
            id="criteria-select"
            value={criteria}
            onChange={handleCriteriaChange}
            className="px-6 py-2 border-2 border-navy-border rounded-md text-sm font-medium text-white bg-navy-dark cursor-pointer transition-all duration-150 min-w-[140px] focus:outline-none focus:border-primary-light focus:shadow-[0_0_0_3px_rgba(124,58,237,0.2)] hover:border-navy-border hover:bg-navy-medium"
            aria-label="Select ranking criteria"
          >
            {RANKING_CRITERIA.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Level Range Slider */}
        <div className={filterGroupLarge}>
          <label htmlFor="level-range" className={filterLabel}>Level:</label>
          <div className="flex items-center gap-3 flex-1">
            <input
              type="range"
              id="level-range"
              min="0"
              max="100"
              value={filters.maxLevel}
              onChange={handleMaxLevelChange}
              className="flex-1 h-[6px] rounded-sm bg-gradient-to-r from-navy-border to-navy-deep appearance-none cursor-pointer focus:outline-none"
              aria-label="Filter by maximum level"
            />
            <div className="flex justify-between w-[80px] text-xs text-text-muted font-semibold">
              <span>{filters.minLevel}</span>
              <span>{filters.maxLevel}</span>
            </div>
          </div>
        </div>

        {/* AI Only Toggle */}
        <div className={filterGroupBase}>
          <label className="flex items-center gap-2 cursor-pointer select-none text-sm text-text-bright font-medium">
            <input
              type="checkbox"
              checked={filters.aiOnly}
              onChange={handleAiOnlyChange}
              className="w-[18px] h-[18px] border-2 border-navy-border rounded-sm accent-primary bg-navy-deep cursor-pointer"
              aria-label="Show AI boxers only"
            />
            <span>AI Only</span>
          </label>
        </div>
      </div>

      {/* Results Info */}
      <div className="mb-4">
        <span className="text-sm text-text-muted font-medium">
          {isLoading ? 'Loading...' : `${filteredData.length} boxer${filteredData.length !== 1 ? 's' : ''} found`}
        </span>
      </div>

      {/* Error Display */}
      {error && (
        <ErrorDisplay
          error={error}
          onRetry={handleRetry}
          variant="alert"
          title="Failed to Load Rankings"
        />
      )}

      {/* Rankings List */}
      <div className="flex flex-col gap-3">
        {isLoading && rankingsData.length === 0 ? (
          /* Loading Skeletons */
          <>
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="rank-card loading p-4 rounded-lg bg-navy-deep/50 border border-navy-border animate-pulse">
                <div className="flex gap-4">
                  <div className="w-[36px] h-[36px] rounded-md bg-gradient-to-r from-navy-border via-navy-deep to-navy-border bg-[length:200%_100%]" />
                  <div className="flex-1 flex flex-col gap-2">
                    <div className="h-[20px] w-[60%] rounded-sm bg-gradient-to-r from-navy-border via-navy-deep to-navy-border bg-[length:200%_100%]" />
                    <div className="h-[14px] w-[40%] rounded-sm bg-gradient-to-r from-navy-border via-navy-deep to-navy-border bg-[length:200%_100%]" />
                  </div>
                </div>
              </div>
            ))}
          </>
        ) : paginatedData.length > 0 ? (
          /* Rank Cards */
          paginatedData.map((boxer, index) => {
            // Calculate actual rank based on filtered/sorted data
            const actualRank = ((page - 1) * DEFAULT_LIMIT) + index + 1;

            return (
              <RankCard
                key={boxer.id}
                boxer={boxer}
                rank={actualRank}
              />
            );
          })
        ) : (
          /* Empty State */
          <div className="text-center py-[10rem] px-5 bg-navy-deep rounded-lg shadow-[0_4px_20px_rgba(0,0,0,0.3),inset_0_1px_0_rgba(255,255,255,0.05)] border border-navy-border">
            <span className="text-5xl block mb-4 opacity-80" aria-hidden="true">📭</span>
            <h2 className="text-lg font-bold text-text-bright m-0 mb-2">No boxers found</h2>
            <p className="text-sm text-text-muted m-0">Try adjusting your search or filters</p>
          </div>
        )}
      </div>

      {/* Pagination Controls */}
      {pagination.pages > 1 && !isLoading && (
        <div className="flex flex-wrap items-center justify-center gap-4 mt-8 py-5">
          <button
            className="px-5 py-2 bg-gradient-to-br from-navy-medium to-navy-deep border border-navy-border rounded-md text-text-bright font-semibold cursor-pointer transition-all duration-150 text-sm hover:bg-gradient-to-br hover:from-navy-medium hover:to-navy-dark hover:border-primary-light hover:-translate-y-0.5 hover:shadow-[0_2px_8px_rgba(124,58,237,0.2)] disabled:opacity-40 disabled:cursor-not-allowed focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            onClick={() => handlePageChange(page - 1)}
            disabled={page <= 1}
            aria-label="Go to previous page"
          >
            ← Prev
          </button>

          <div className="flex gap-2">
            {Array.from({ length: Math.min(5, pagination.pages) }, (_, i) => {
              let pageNum;
              if (pagination.pages <= 5) {
                pageNum = i + 1;
              } else if (page <= 3) {
                pageNum = i + 1;
              } else if (page >= pagination.pages - 2) {
                pageNum = pagination.pages - 4 + i;
              } else {
                pageNum = page - 2 + i;
              }

              return (
                <button
                  key={pageNum}
                  className={`w-9 h-9 border rounded-md font-semibold cursor-pointer transition-all duration-150 flex items-center justify-center ${
                    page === pageNum
                      ? 'bg-gradient-to-br from-primary to-primary-dark border-primary text-white shadow-[0_2px_12px_rgba(124,58,237,0.4),inset_0_1px_0_rgba(255,255,255,0.2)]'
                      : 'border-navy-border bg-navy-deep text-text-bright hover:bg-navy-medium hover:border-primary-light hover:shadow-[0_0_12px_rgba(124,58,237,0.2)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2'
                  }`}
                  onClick={() => handlePageChange(pageNum)}
                  aria-label={`Go to page ${pageNum}`}
                  aria-current={page === pageNum ? 'page' : undefined}
                >
                  {pageNum}
                </button>
              );
            })}
          </div>

          <button
            className="px-5 py-2 bg-gradient-to-br from-navy-medium to-navy-deep border border-navy-border rounded-md text-text-bright font-semibold cursor-pointer transition-all duration-150 text-sm hover:bg-gradient-to-br hover:from-navy-medium hover:to-navy-dark hover:border-primary-light hover:-translate-y-0.5 hover:shadow-[0_2px_8px_rgba(124,58,237,0.2)] disabled:opacity-40 disabled:cursor-not-allowed focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2"
            onClick={() => handlePageChange(page + 1)}
            disabled={page >= pagination.pages}
            aria-label="Go to next page"
          >
            Next →
          </button>
        </div>
      )}

      {/* Page Indicator */}
      {pagination.pages > 1 && (
        <div className="text-center mt-4 text-text-muted text-xs font-medium">
          Page {page} of {pagination.pages}
        </div>
      )}
      </main>
    </div>
  );
}

export default Rankings;
