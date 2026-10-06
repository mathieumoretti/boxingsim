package errors

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents the envelope for error responses.
type ErrorResponse struct {
	Error *APIError `json:"error"`
}

// WriteError writes a standardized error response to the HTTP writer.
func WriteError(w http.ResponseWriter, apiErr *APIError) {
	statusCode := GetHTTPStatus(apiErr.Code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: apiErr})
}

// WriteSuccess writes a standardized success response.
func WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]any{
		"success": true,
		"data":    data,
	}
	_ = json.NewEncoder(w).Encode(response)
}

// WriteJSON writes a raw JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteSimpleError writes a simple error message (for backward compatibility during migration).
// Use WriteError() for new code.
func WriteSimpleError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
