package handler

import (
	"encoding/json"
	"net/http"
)

// BindJSON decodes JSON from the request body into the provided struct.
// Returns an error if the JSON is invalid or cannot be decoded.
func BindJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
