package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxerHandler_CreateBoxer(t *testing.T) {
	// Test with a valid boxer creation request
	handler := NewBoxerHandler(nil) // Using nil store for this test - we're just testing request parsing

	boxerCreate := map[string]interface{}{
		"name":       "Test Boxer",
		"nickname":   "TB",
		"position_x": 0.0,
		"position_y": 0.0,
		"strength":   10.0,
		"defense":    8.0,
		"agility":    12.0,
	}

	jsonData, _ := json.Marshal(boxerCreate)
	req := httptest.NewRequest("POST", "/boxers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", "1")

	w := httptest.NewRecorder()

	handler.CreateBoxer(w, req)

	// We expect a 503 error since we don't have a real database connection (Service Unavailable)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestBoxerHandler_GetBoxer(t *testing.T) {
	// Test with valid ID
	handler := NewBoxerHandler(nil) // Using nil store for this test - we're just testing request parsing

	req := httptest.NewRequest("GET", "/boxers/1", nil)
	w := httptest.NewRecorder()

	handler.GetBoxer(w, req)

	// We expect a 503 error since we don't have a real database connection (Service Unavailable)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestBoxerHandler_UpdateBoxer(t *testing.T) {
	// Test with a valid boxer update request
	handler := NewBoxerHandler(nil) // Using nil store for this test - we're just testing request parsing

	boxerUpdate := map[string]interface{}{
		"name":     "Updated Boxer",
		"strength": 15.0,
		"defense":  12.0,
		"agility":  18.0,
	}

	jsonData, _ := json.Marshal(boxerUpdate)
	req := httptest.NewRequest("PUT", "/boxers/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.UpdateBoxer(w, req)

	// We expect a 503 error since we don't have a real database connection (Service Unavailable)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestBoxerHandler_GetBoxersByUserID(t *testing.T) {
	// Test with valid user ID
	handler := NewBoxerHandler(nil) // Using nil store for this test

	req := httptest.NewRequest("GET", "/users/1/boxers", nil)
	w := httptest.NewRecorder()

	handler.GetBoxersByUserID(w, req)

	// We expect OK status since we have an implementation now (stub response)
	assert.Equal(t, http.StatusOK, w.Code)
}
