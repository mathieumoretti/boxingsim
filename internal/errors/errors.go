package errors

import (
	"net/http"
)

// ErrorCode represents a standardized error code for API responses.
type ErrorCode string

// Standardized error codes following HTTP semantics.
const (
	// General errors
	ErrorCodeInternal   ErrorCode = "INTERNAL_ERROR"
	ErrorCodeBadRequest ErrorCode = "BAD_REQUEST"
	ErrorCodeValidation ErrorCode = "VALIDATION_ERROR"

	// Authentication & Authorization
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"

	// Resource errors
	ErrorCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrorCodeConflict           ErrorCode = "CONFLICT"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Business logic errors
	ErrorCodeInvalidState  ErrorCode = "INVALID_STATE"
	ErrorCodeLimitExceeded ErrorCode = "LIMIT_EXCEEDED"

	// Training-specific errors
	ErrorCodeEnergyInsufficient ErrorCode = "ENERGY_INSUFFICIENT"
	ErrorCodeBoxerInRecovery    ErrorCode = "BOXER_IN_RECOVERY"
	ErrorCodeTrainingScheduled  ErrorCode = "TRAINING_SCHEDULED"
	ErrorCodeBoxerResting       ErrorCode = "BOXER_RESTING"

	// Fatigue errors
	ErrorCodeFatigueTooHigh ErrorCode = "FATIGUE_TOO_HIGH"
)

// APIError represents a standardized error response for the API.
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return string(e.Code) + ": " + e.Message
}

// WithDetails adds optional details to the error.
func (e *APIError) WithDetails(details any) *APIError {
	e.Details = details
	return e
}

// Helper functions for common error scenarios.

// Internal creates an internal server error.
func Internal(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeInternal,
		Message: message,
	}
}

// NotFound creates a not found error.
func NotFound(resource string) *APIError {
	return &APIError{
		Code:    ErrorCodeNotFound,
		Message: resource + " not found",
	}
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeUnauthorized,
		Message: message,
	}
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeForbidden,
		Message: message,
	}
}

// Validation creates a validation error.
func Validation(field, message string) *APIError {
	return &APIError{
		Code:    ErrorCodeValidation,
		Message: message,
		Details: map[string]string{
			"field": field,
		},
	}
}

// Conflict creates a conflict error.
func Conflict(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeConflict,
		Message: message,
	}
}

// BadRequest creates a bad request error.
func BadRequest(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeBadRequest,
		Message: message,
	}
}

// InvalidState creates an invalid state error.
func InvalidState(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeInvalidState,
		Message: message,
	}
}

// InvalidStateWithDetails creates an invalid state error with additional context.
func InvalidStateWithDetails(message string, details map[string]any) *APIError {
	return &APIError{
		Code:    ErrorCodeInvalidState,
		Message: message,
		Details: details,
	}
}

// EnergyInsufficient creates an energy insufficient error with required/available details.
func EnergyInsufficient(required, available int) *APIError {
	return &APIError{
		Code:    ErrorCodeEnergyInsufficient,
		Message: "Insufficient energy for training",
		Details: map[string]interface{}{
			"required":  required,
			"available": available,
		},
	}
}

// BoxerInRecovery creates a boxer in recovery error.
func BoxerInRecovery(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeBoxerInRecovery,
		Message: message,
	}
}

// TrainingScheduled creates a training already scheduled error.
func TrainingScheduled() *APIError {
	return &APIError{
		Code:    ErrorCodeTrainingScheduled,
		Message: "Boxer already has a pending training session",
	}
}

// BoxerResting creates a boxer resting error with rest end time.
func BoxerResting(restEndsAt string) *APIError {
	return &APIError{
		Code:    ErrorCodeBoxerResting,
		Message: "Boxer is currently resting",
		Details: map[string]string{
			"rest_ends_at": restEndsAt,
		},
	}
}

// FatigueTooHigh creates a fatigue too high error.
func FatigueTooHigh(message string) *APIError {
	return &APIError{
		Code:    ErrorCodeFatigueTooHigh,
		Message: message,
	}
}

// ServiceUnavailable creates a service unavailable error.
func ServiceUnavailable() *APIError {
	return &APIError{
		Code:    ErrorCodeServiceUnavailable,
		Message: "Service temporarily unavailable",
	}
}

// GetHTTPStatus returns the appropriate HTTP status code for an error code.
func GetHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrorCodeInternal:
		return http.StatusInternalServerError
	case ErrorCodeBadRequest, ErrorCodeValidation:
		return http.StatusBadRequest
	case ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeForbidden:
		return http.StatusForbidden
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeInvalidState, ErrorCodeLimitExceeded:
		return http.StatusUnprocessableEntity
	case ErrorCodeEnergyInsufficient, ErrorCodeBoxerInRecovery,
		ErrorCodeTrainingScheduled, ErrorCodeBoxerResting,
		ErrorCodeFatigueTooHigh:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// Wrap wraps a standard error into an APIError with additional context.
func Wrap(err error, code ErrorCode, message string) *APIError {
	apiErr := &APIError{
		Code:    code,
		Message: message,
	}
	if err != nil {
		apiErr.Details = map[string]string{
			"original_error": err.Error(),
		}
	}
	return apiErr
}
