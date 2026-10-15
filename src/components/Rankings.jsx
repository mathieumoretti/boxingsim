import React, { useState, useEffect, useCallback } from 'react';
import RankCard from './RankCard';
import './Rankings.css';
import { fetchRankings, RANKING_CRITERIA, getCriteriaDisplayValue } from '../api/rankings';
import { ErrorDisplay } from './ErrorDisplay';
import { getUserErrorMessage } from '../utils/errorMessages';

const DEFAULT_LIMIT = 25;

/**
 * Rankings Component (MAT-105)
 *
 * Displays boxer rankings sorted by various criteria with search, filtering,
 * and pagination. Uses the MAT-100 backend API for data retrieval.
 */
function Rankings() {
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

  return (
    <div className="rankings-page">
      {/* Header Section */}
      <div className="rankings-header">
        <div className="header-title">
          <span className="rankings-icon">🏆</span>
          <h1>{criteriaLabel} Rankings</h1>
        </div>

        {/* Search Box */}
        <div className="search-box">
          <input
            type="text"
            placeholder="Search boxers..."
            value={searchQuery}
            onChange={handleSearchChange}
            className="search-input"
          />
          <span className="search-icon">🔍</span>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="rankings-filters">
        {/* Criteria Selector */}
        <div className="filter-group">
          <label htmlFor="criteria-select">Sort by:</label>
          <select
            id="criteria-select"
            value={criteria}
            onChange={handleCriteriaChange}
            className="criteria-select"
          >
            {RANKING_CRITERIA.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {/* Level Range Slider */}
        <div className="filter-group filter-group-large">
          <label htmlFor="level-range">Level:</label>
          <div className="level-slider-container">
            <input
              type="range"
              id="level-range"
              min="0"
              max="100"
              value={filters.maxLevel}
              onChange={handleMaxLevelChange}
              className="level-slider"
            />
            <div className="level-indicators">
              <span className="level-min">{filters.minLevel}</span>
              <span className="level-max">{filters.maxLevel}</span>
            </div>
          </div>
        </div>

        {/* AI Only Toggle */}
        <div className="filter-group">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={filters.aiOnly}
              onChange={handleAiOnlyChange}
            />
            <span>AI Only</span>
          </label>
        </div>
      </div>

      {/* Results Info */}
      <div className="rankings-info">
        <span className="results-count">
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

      {/* Rankings List - Table View */}
      <div className="rankings-list">
        {isLoading && rankingsData.length === 0 ? (
          /* Loading Skeletons */
          <>
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="rank-card loading">
                <div className="rank-skeleton">
                  <div className="skeleton-badge"></div>
                </div>
                <div className="rank-skeleton-info">
                  <div className="skeleton-name"></div>
                  <div className="skeleton-stats"></div>
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
          <div className="rankings-empty">
            <span className="empty-icon">📭</span>
            <h2>No boxers found</h2>
            <p>Try adjusting your search or filters</p>
          </div>
        )}
      </div>

      {/* Pagination Controls */}
      {pagination.pages > 1 && !isLoading && (
        <div className="rankings-pagination">
          <button
            className="pagination-btn"
            onClick={() => handlePageChange(page - 1)}
            disabled={page <= 1}
          >
            ← Prev
          </button>

          <div className="pagination-numbers">
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
                  className={`pagination-number ${page === pageNum ? 'active' : ''}`}
                  onClick={() => handlePageChange(pageNum)}
                >
                  {pageNum}
                </button>
              );
            })}
          </div>

          <button
            className="pagination-btn"
            onClick={() => handlePageChange(page + 1)}
            disabled={page >= pagination.pages}
          >
            Next →
          </button>
        </div>
      )}

      {/* Page Indicator */}
      {pagination.pages > 1 && (
        <div className="rankings-page-info">
          Page {page} of {pagination.pages}
        </div>
      )}
    </div>
  );
}

export default Rankings;
