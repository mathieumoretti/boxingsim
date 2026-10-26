package spa

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Handler serves a React SPA with proper routing support.
// It handles:
// - Serving actual static files (JS, CSS, images) from the build directory
// - Serving index.html for any unknown path (SPA fallback for React Router)
// - Returning 404 for truly missing resources (not treating .js/.css as SPA routes)
type Handler struct {
	// Dir is the directory containing built frontend files (e.g., "./dist")
	Dir string
}

// New creates a new SPA handler with the given static file directory.
func New(dir string) *Handler {
	return &Handler{
		Dir: dir,
	}
}

// ServeHTTP implements http.Handler.
// It serves index.html for SPA routes while properly handling static assets.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle GET requests for document navigation
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	path := r.URL.Path

	// Don't serve API endpoints - they should be handled by their own routes
	if strings.HasPrefix(path, "/api") {
		http.NotFound(w, r)
		return
	}

	// Try to find the file in the static directory
	fs := http.FileServer(http.Dir(h.Dir))

	// Check if this is a request for an actual static file
	// (not starting with / prevents path traversal, and we check file extensions)
	if h.isStaticFile(path) {
		// Serve the actual file - if it doesn't exist, return 404
		fs.ServeHTTP(w, r)
		return
	}

	// For everything else (SPA routes like /boxer/123), try to serve index.html
	// First check if index.html exists in the static directory
 indexPath := filepath.Join(h.Dir, "index.html")
	if h.fileExists(indexPath) {
		// Serve index.html for SPA routing
		http.ServeFile(w, r, indexPath)
	} else {
		// No index.html found - truly a 404
		http.NotFound(w, r)
	}
}

// isStaticFile checks if the path corresponds to a static asset.
// These should be served as actual files, not fallback to index.html.
func (h *Handler) isStaticFile(path string) bool {
	// Static file extensions that should NOT fall back to index.html
	staticExtensions := []string{
		".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg",
		".ico", ".woff", ".woff2", ".ttf", ".eot", ".json",
		".txt", ".xml", ".pdf", ".mp4", ".webm", ".ogg", ".mp3",
	}

	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(staticExtensions, ext)
}

// fileExists checks if a file exists at the given path.
func (h *Handler) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
