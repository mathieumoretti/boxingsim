package domain

import (
	"fmt"
	"net/url"
	"strconv"
)

// Pagination represents standard pagination parameters for data retrieval.
// It provides a consistent way to page through large datasets across the application.
//
// Usage:
//
//	p := Pagination{}
//	if err := p.FromQuery(r.URL.Query()); err != nil {
//	    return errors.BadRequest("Invalid pagination parameters")
//	}
//	offset := p.Offset()
//	limit := p.Limit
//	results, err := store.GetBoxers(ctx, offset, limit)
type Pagination struct {
	Page  int // 1-indexed page number (default: 1)
	Limit int // Items per page (default: 20, max: 100)
}

// Default pagination constants.
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MinLimit     = 1
	MaxLimit     = 100
	MinPage      = 1
)

// DefaultPagination returns a Pagination struct with default values.
func DefaultPagination() Pagination {
	return Pagination{Page: DefaultPage, Limit: DefaultLimit}
}

// FromQuery extracts pagination parameters from HTTP query string.
// Recognizes "page" (1-indexed) and "limit" query parameters.
//
// Example: ?page=2&limit=50 -> Page: 2, Limit: 50
// Example: ?page=-1       -> Error (negative page number)
// Example: ?limit=200     -> Limit capped to MaxLimit (100)
func (p *Pagination) FromQuery(values url.Values) error {
	// Set defaults first
	p.Page = DefaultPage
	p.Limit = DefaultLimit

	// Parse page parameter
	if pageStr := values.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return fmt.Errorf("invalid page parameter: %w", err)
		}
		if page < MinPage {
			return fmt.Errorf("page must be at least %d, got %d", MinPage, page)
		}
		p.Page = page
	}

	// Parse limit parameter
	if limitStr := values.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return fmt.Errorf("invalid limit parameter: %w", err)
		}
		if limit < MinLimit {
			return fmt.Errorf("limit must be at least %d, got %d", MinLimit, limit)
		}
		if limit > MaxLimit {
			// Cap to maximum for performance protection
			limit = MaxLimit
		}
		p.Limit = limit
	}

	return nil
}

// Offset returns the SQL OFFSET value for the current page.
// This is used in database queries to skip records before the current page.
//
// Formula: offset = (page - 1) * limit
// Example: Page=2, Limit=20 -> Offset=20 (skip first 20 records)
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

// IsValid returns true if the pagination parameters are valid.
func (p Pagination) IsValid() bool {
	return p.Page >= MinPage && p.Limit >= MinLimit && p.Limit <= MaxLimit
}

// HasNext returns true if there is a next page given the total count.
func (p Pagination) HasNext(totalCount int) bool {
	return p.Offset()+p.Limit < totalCount
}

// HasPrev returns true if there is a previous page.
func (p Pagination) HasPrev() bool {
	return p.Page > MinPage
}

// TotalPages calculates the total number of pages for a given count.
// Uses ceiling division to ensure partial pages are counted.
//
// Example: totalCount=45, limit=20 -> totalPages=3
func (p Pagination) TotalPages(totalCount int) int {
	if p.Limit <= 0 {
		return 0
	}
	return (totalCount + p.Limit - 1) / p.Limit
}

// NextPage returns the next page number if available, or 0 if at end.
func (p Pagination) NextPage(totalCount int) int {
	if p.HasNext(totalCount) {
		return p.Page + 1
	}
	return 0
}

// PrevPage returns the previous page number if available, or 0 if at start.
func (p Pagination) PrevPage() int {
	if p.HasPrev() {
		return p.Page - 1
	}
	return 0
}

// PaginatedResponse wraps paginated data with metadata for API responses.
// This provides clients with information about pagination state and navigation.
//
// Example response:
//
//	{
//	  "data": [...],
//	  "pagination": {
//	    "page": 2,
//	    "limit": 20,
//	    "total_count": 45,
//	    "total_pages": 3,
//	    "has_next": true,
//	    "has_prev": true
//	  }
//	}
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo contains metadata about the current pagination state.
type PaginationInfo struct {
	Page       int  `json:"page"`        // Current page number (1-indexed)
	Limit      int  `json:"limit"`       // Items per page
	TotalCount int  `json:"total_count"` // Total number of items across all pages
	TotalPages int  `json:"total_pages"` // Total number of pages
	HasNext    bool `json:"has_next"`    // True if there is a next page
	HasPrev    bool `json:"has_prev"`    // True if there is a previous page
}

// NewPaginatedResponse creates a PaginatedResponse with calculated metadata.
func NewPaginatedResponse[T any](data []T, pagination Pagination, totalCount int) PaginatedResponse[T] {
	totalPages := pagination.TotalPages(totalCount)

	return PaginatedResponse[T]{
		Data: data,
		Pagination: PaginationInfo{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			TotalCount: totalCount,
			TotalPages: totalPages,
			HasNext:    pagination.HasNext(totalCount),
			HasPrev:    pagination.HasPrev(),
		},
	}
}

// BuildSQLParams returns offset and limit values for use in SQL queries.
// This helper provides a consistent way to extract parameters for database queries.
//
// Usage:
//
//	offset, limit := p.BuildSQLParams()
//	query := "SELECT * FROM boxers ORDER BY id LIMIT $1 OFFSET $2"
//	rows, err := db.Query(query, limit, offset)
func (p Pagination) BuildSQLParams() (offset, limit int) {
	return p.Offset(), p.Limit
}

// String returns a human-readable representation for debugging.
func (p Pagination) String() string {
	return fmt.Sprintf("Pagination{Page: %d, Limit: %d, Offset: %d}", p.Page, p.Limit, p.Offset())
}
