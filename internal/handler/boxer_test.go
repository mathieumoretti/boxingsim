package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxerHandler_CreateBoxer(t *testing.T) {
	// For now we just test that it doesn't panic
	handler := NewBoxerHandler()

	req := httptest.NewRequest("POST", "/boxers", nil)
	w := httptest.NewRecorder()

	handler.CreateBoxer(w, req)

	// We expect bad request since we don't have a real implementation yet but the handler
	// is doing some validation that causes a 400 error instead of 501
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBoxerHandler_GetBoxer(t *testing.T) {
	// For now we just test that it doesn't panic
	handler := NewBoxerHandler()

	req := httptest.NewRequest("GET", "/boxers/1", nil)
	w := httptest.NewRecorder()

	handler.GetBoxer(w, req)

	// We expect not implemented since we don't have a real implementation yet
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestBoxerHandler_UpdateBoxer(t *testing.T) {
	// For now we just test that it doesn't panic
	handler := NewBoxerHandler()

	req := httptest.NewRequest("PUT", "/boxers/1", nil)
	w := httptest.NewRecorder()

	handler.UpdateBoxer(w, req)

	// We expect bad request since we don't have a real implementation yet but the handler
	// is doing some validation that causes a 400 error instead of 501
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
