package handler

import (
	"net/http"

	"github.com/mormm/boxing/internal/platform/logger"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

// GetDashboard renders the dashboard page
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	logger := logger.New("dashboard")
	logger.Info("GetDashboard endpoint called - Method: %s, URL: %s", r.Method, r.URL.Path)

	// In a real implementation, this would render the dashboard page
	// For now we'll return a simple JSON response for testing purposes

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message": "Dashboard loaded successfully"}`))
	logger.Info("Dashboard endpoint completed successfully")
}