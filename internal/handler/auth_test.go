package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mormm/boxing/internal/model"
)

func TestAuthHandler_RegisterUser(t *testing.T) {
	// Create a mock auth service
	handler := NewAuthHandler()

	// Prepare test data
	registerReq := model.UserRegister{
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	// Create request body
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler - we're just testing that it doesn't panic
	handler.RegisterUser(w, req)

	// We expect OK status now since we have an implementation
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_RegisterUser_PasswordMismatch(t *testing.T) {
	handler := NewAuthHandler()

	// Prepare test data with mismatched passwords
	registerReq := model.UserRegister{
		Username:        "testuser",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "differentpassword",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.RegisterUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RegisterUser_InvalidJSON(t *testing.T) {
	handler := NewAuthHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.RegisterUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_LoginUser(t *testing.T) {
	// Create a mock auth service
	handler := NewAuthHandler()

	// Prepare test data
	loginReq := model.UserLogin{
		Username: "testuser",
		Password: "password123",
	}

	// Create request body
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler - we're just testing that it doesn't panic
	handler.LoginUser(w, req)

	// We expect OK status now since we have an implementation
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthHandler_LoginUser_InvalidJSON(t *testing.T) {
	handler := NewAuthHandler()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.LoginUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
