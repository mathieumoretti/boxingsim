package domain

import (
	"net/url"
	"testing"
)

func TestPagination_FromQuery(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		wantPage    int
		wantLimit   int
		wantErr     bool
		errContains string
	}{
		{
			name:        "no parameters uses defaults",
			queryParams: "",
			wantPage:    DefaultPage,
			wantLimit:   DefaultLimit,
			wantErr:     false,
		},
		{
			name:        "valid page parameter",
			queryParams: "page=3",
			wantPage:    3,
			wantLimit:   DefaultLimit,
			wantErr:     false,
		},
		{
			name:        "valid limit parameter",
			queryParams: "limit=50",
			wantPage:    DefaultPage,
			wantLimit:   50,
			wantErr:     false,
		},
		{
			name:        "both page and limit",
			queryParams: "page=2&limit=25",
			wantPage:    2,
			wantLimit:   25,
			wantErr:     false,
		},
		{
			name:        "limit capped to max",
			queryParams: "page=1&limit=500",
			wantPage:    1,
			wantLimit:   MaxLimit,
			wantErr:     false,
		},
		{
			name:        "negative page number",
			queryParams: "page=-1",
			wantPage:    0,
			wantLimit:   DefaultLimit,
			wantErr:     true,
			errContains: "at least",
		},
		{
			name:        "zero page number",
			queryParams: "page=0",
			wantPage:    0,
			wantLimit:   DefaultLimit,
			wantErr:     true,
			errContains: "at least",
		},
		{
			name:        "invalid page parameter (non-numeric)",
			queryParams: "page=abc",
			wantPage:    0,
			wantLimit:   DefaultLimit,
			wantErr:     true,
			errContains: "invalid page",
		},
		{
			name:        "negative limit",
			queryParams: "limit=-10",
			wantPage:    DefaultPage,
			wantLimit:   0,
			wantErr:     true,
			errContains: "at least",
		},
		{
			name:        "zero limit",
			queryParams: "limit=0",
			wantPage:    DefaultPage,
			wantLimit:   0,
			wantErr:     true,
			errContains: "at least",
		},
		{
			name:        "invalid limit parameter (non-numeric)",
			queryParams: "limit=xyz",
			wantPage:    DefaultPage,
			wantLimit:   0,
			wantErr:     true,
			errContains: "invalid limit",
		},
		{
			name:        "minimum valid values",
			queryParams: "page=1&limit=1",
			wantPage:    1,
			wantLimit:   1,
			wantErr:     false,
		},
		{
			name:        "maximum valid limit",
			queryParams: "page=1&limit=100",
			wantPage:    1,
			wantLimit:   100,
			wantErr:     false,
		},
		{
			name:        "large page number",
			queryParams: "page=9999&limit=10",
			wantPage:    9999,
			wantLimit:   10,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsedURL *url.URL
			var err error
			if tt.queryParams == "" {
				parsedURL, _ = url.Parse("http://example.com")
			} else {
				parsedURL, err = url.Parse("http://example.com?" + tt.queryParams)
				if err != nil {
					t.Fatalf("Failed to parse URL: %v", err)
				}
			}

			p := Pagination{}
			err = p.FromQuery(parsedURL.Query())

			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("FromQuery() error = %v, wantErr %v, err message = %v", err, tt.wantErr, err)
			}

			if tt.errContains != "" && err != nil {
				if errStr := err.Error(); !contains(errStr, tt.errContains) {
					t.Errorf("FromQuery() error = %q, want error containing %q", errStr, tt.errContains)
				}
			}

			if !tt.wantErr {
				if p.Page != tt.wantPage {
					t.Errorf("Page = %d, want %d", p.Page, tt.wantPage)
				}
				if p.Limit != tt.wantLimit {
					t.Errorf("Limit = %d, want %d", p.Limit, tt.wantLimit)
				}
			}
		})
	}
}

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		limit int
		want  int
	}{
		{"page 1, limit 20", 1, 20, 0},
		{"page 2, limit 20", 2, 20, 20},
		{"page 3, limit 20", 3, 20, 40},
		{"page 1, limit 50", 1, 50, 0},
		{"page 10, limit 10", 10, 10, 90},
		{"page 100, limit 20", 100, 20, 1980},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: tt.limit}
			if got := p.Offset(); got != tt.want {
				t.Errorf("Offset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPagination_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		limit int
		want  bool
	}{
		{"valid default", 1, 20, true},
		{"valid large page", 9999, 50, true},
		{"valid min values", 1, 1, true},
		{"valid max limit", 1, 100, true},
		{"invalid zero page", 0, 20, false},
		{"invalid negative page", -1, 20, false},
		{"invalid zero limit", 1, 0, false},
		{"invalid negative limit", 1, -10, false},
		{"invalid over max limit", 1, 101, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: tt.limit}
			if got := p.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPagination_HasNext(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		totalCount int
		want       bool
	}{
		{"has next page", 1, 20, 45, true},
		{"no next page - exact fit", 1, 20, 20, false},
		{"no next page - under limit", 1, 20, 15, false},
		{"has next on page 2", 2, 20, 45, true},
		{"no next on last page", 3, 20, 45, false},
		{"empty results", 1, 20, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: tt.limit}
			if got := p.HasNext(tt.totalCount); got != tt.want {
				t.Errorf("HasNext(%d) = %v, want %v", tt.totalCount, got, tt.want)
			}
		})
	}
}

func TestPagination_HasPrev(t *testing.T) {
	tests := []struct {
		name string
		page int
		want bool
	}{
		{"no prev on page 1", 1, false},
		{"has prev on page 2", 2, true},
		{"has prev on page 10", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: 20}
			if got := p.HasPrev(); got != tt.want {
				t.Errorf("HasPrev() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPagination_TotalPages(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		totalCount int
		want       int
	}{
		{"exact fit", 20, 40, 2},
		{"partial last page", 20, 45, 3},
		{"single item", 20, 1, 1},
		{"empty results", 20, 0, 0},
		{"limit 1", 1, 5, 5},
		{"large count", 20, 250, 13},
		{"limit larger than count", 100, 25, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: 1, Limit: tt.limit}
			if got := p.TotalPages(tt.totalCount); got != tt.want {
				t.Errorf("TotalPages(%d) = %d, want %d", tt.totalCount, got, tt.want)
			}
		})
	}
}

func TestPagination_NextPage(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		totalCount int
		want       int
	}{
		{"next page exists", 1, 20, 45, 2},
		{"no next page - exact fit", 1, 20, 20, 0},
		{"no next page - last page", 3, 20, 45, 0},
		{"next from middle page", 2, 20, 100, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: tt.limit}
			if got := p.NextPage(tt.totalCount); got != tt.want {
				t.Errorf("NextPage(%d) = %d, want %d", tt.totalCount, got, tt.want)
			}
		})
	}
}

func TestPagination_PrevPage(t *testing.T) {
	tests := []struct {
		name string
		page int
		want int
	}{
		{"no prev on page 1", 1, 0},
		{"prev from page 2", 2, 1},
		{"prev from page 10", 10, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: 20}
			if got := p.PrevPage(); got != tt.want {
				t.Errorf("PrevPage() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPagination_BuildSQLParams(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		limit     int
		wantOff   int
		wantLimit int
	}{
		{"page 1", 1, 20, 0, 20},
		{"page 2", 2, 20, 20, 20},
		{"page 5 limit 10", 5, 10, 40, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pagination{Page: tt.page, Limit: tt.limit}
			offset, limit := p.BuildSQLParams()
			if offset != tt.wantOff {
				t.Errorf("Offset = %d, want %d", offset, tt.wantOff)
			}
			if limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", limit, tt.wantLimit)
			}
		})
	}
}

func TestPagination_String(t *testing.T) {
	p := Pagination{Page: 3, Limit: 25}
	got := p.String()
	want := "Pagination{Page: 3, Limit: 25, Offset: 50}"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestDefaultPagination(t *testing.T) {
	p := DefaultPagination()
	if p.Page != DefaultPage {
		t.Errorf("Page = %d, want %d", p.Page, DefaultPage)
	}
	if p.Limit != DefaultLimit {
		t.Errorf("Limit = %d, want %d", p.Limit, DefaultLimit)
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	data := []string{"item1", "item2", "item3"}
	pagination := Pagination{Page: 2, Limit: 10}
	totalCount := 25

	resp := NewPaginatedResponse(data, pagination, totalCount)

	if len(resp.Data) != 3 {
		t.Errorf("Data length = %d, want 3", len(resp.Data))
	}

	if resp.Pagination.Page != 2 {
		t.Errorf("Page = %d, want 2", resp.Pagination.Page)
	}
	if resp.Pagination.Limit != 10 {
		t.Errorf("Limit = %d, want 10", resp.Pagination.Limit)
	}
	if resp.Pagination.TotalCount != 25 {
		t.Errorf("TotalCount = %d, want 25", resp.Pagination.TotalCount)
	}
	if resp.Pagination.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", resp.Pagination.TotalPages)
	}
	if !resp.Pagination.HasNext {
		t.Error("HasNext should be true")
	}
	if !resp.Pagination.HasPrev {
		t.Error("HasPrev should be true")
	}
}

func TestNewPaginatedResponse_FirstPage(t *testing.T) {
	data := []int{1, 2, 3}
	pagination := Pagination{Page: 1, Limit: 10}
	totalCount := 15

	resp := NewPaginatedResponse(data, pagination, totalCount)

	if resp.Pagination.HasPrev {
		t.Error("HasPrev should be false on first page")
	}
	if !resp.Pagination.HasNext {
		t.Error("HasNext should be true")
	}
}

func TestNewPaginatedResponse_LastPage(t *testing.T) {
	data := []int{1, 2}
	pagination := Pagination{Page: 2, Limit: 10}
	totalCount := 15

	resp := NewPaginatedResponse(data, pagination, totalCount)

	if !resp.Pagination.HasPrev {
		t.Error("HasPrev should be true")
	}
	if resp.Pagination.HasNext {
		t.Error("HasNext should be false on last page")
	}
}

func TestNewPaginatedResponse_Empty(t *testing.T) {
	data := []string{}
	pagination := Pagination{Page: 1, Limit: 10}
	totalCount := 0

	resp := NewPaginatedResponse(data, pagination, totalCount)

	if len(resp.Data) != 0 {
		t.Errorf("Data length = %d, want 0", len(resp.Data))
	}
	if resp.Pagination.TotalPages != 0 {
		t.Errorf("TotalPages = %d, want 0", resp.Pagination.TotalPages)
	}
	if resp.Pagination.HasNext {
		t.Error("HasNext should be false for empty results")
	}
	if resp.Pagination.HasPrev {
		t.Error("HasPrev should be false for empty results")
	}
}

// Helper function for testing error messages
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
