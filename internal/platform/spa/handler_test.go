package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandler_SPA_Routing(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create index.html
	indexHTML := `<!DOCTYPE html><html><head><title>Test</title></head><body><div id="root"></div></body></html>`
	indexPath := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(indexHTML), 0644); err != nil {
		t.Fatalf("Failed to create test index.html: %v", err)
	}

	// Create a test JS bundle
	bundleJS := `console.log('test');`
	bundlePath := filepath.Join(tmpDir, "bundle.js")
	if err := os.WriteFile(bundlePath, []byte(bundleJS), 0644); err != nil {
		t.Fatalf("Failed to create test bundle.js: %v", err)
	}

	handler := New(tmpDir)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectContent  string // substring that should appear in response
	}{
		{
			name:           "Root path returns index.html",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectContent:  "<div id=\"root\"></div>",
		},
		{
			name:           "SPA route /dashboard returns index.html",
			path:           "/dashboard",
			expectedStatus: http.StatusOK,
			expectContent:  "<div id=\"root\"></div>",
		},
		{
			name:           "SPA route with ID /boxer/123 returns index.html",
			path:           "/boxer/123",
			expectedStatus: http.StatusOK,
			expectContent:  "<div id=\"root\"></div>",
		},
		{
			name:           "Nested SPA route /boxer/123/train returns index.html",
			path:           "/boxer/123/train",
			expectedStatus: http.StatusOK,
			expectContent:  "<div id=\"root\"></div>",
		},
		{
			name:           "Fight route /fight/456 returns index.html",
			path:           "/fight/456",
			expectedStatus: http.StatusOK,
			expectContent:  "<div id=\"root\"></div>",
		},
		{
			name:           "Static JS file returns actual content",
			path:           "/bundle.js",
			expectedStatus: http.StatusOK,
			expectContent:  "console.log('test')",
		},
		{
			name:           "Non-existent static file returns 404 not index.html",
			path:           "/nonexistent.js",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "API path returns 404 (not handled by SPA)",
			path:           "/api/boxers",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Rankings route returns index.html",
			path:           "/rankings",
			expectedStatus: http.StatusOK,
			expectContent:  "<div id=\"root\"></div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d for path %s, got %d", tt.expectedStatus, tt.path, w.Code)
			}

			if tt.expectContent != "" && !contains(w.Body.String(), tt.expectContent) {
				t.Errorf("Expected response to contain '%s' for path %s, got: %s",
					tt.expectContent, tt.path, w.Body.String())
			}
		})
	}
}

func TestHandler_MethodRestrictions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create index.html
	indexPath := filepath.Join(tmpDir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("Failed to create test index.html: %v", err)
	}

	handler := New(tmpDir)

	tests := []struct {
		name       string
		method     string
		path       string
		expectCode int
	}{
		{
			name:       "POST requests return 404",
			method:     http.MethodPost,
			path:       "/boxer/123",
			expectCode: http.StatusNotFound,
		},
		{
			name:       "PUT requests return 404",
			method:     http.MethodPut,
			path:       "/boxer/123",
			expectCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("Expected status %d for %s %s, got %d",
					tt.expectCode, tt.method, tt.path, w.Code)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
