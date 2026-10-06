# API Error Codes

This document describes the standardized error response format and all error codes used by the Boxing Simulation API.

## Error Response Format

All API errors follow this standardized structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {}  // Optional additional context
  }
}
```

### Example Responses

**Not Found (404):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Boxer not found"
  }
}
```

**Validation Error (422):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Name is required",
    "details": {
      "field": "name"
    }
  }
}
```

**Internal Server Error (500):**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to process request",
    "details": {
      "original_error": "database connection failed"
    }
  }
}
```

## Error Codes Reference

### Authentication & Authorization

| Code | HTTP Status | Description | Endpoints |
|------|-------------|-------------|-----------|
| `UNAUTHORIZED` | 401 | Missing or invalid authentication token | All protected endpoints |
| `FORBIDDEN` | 403 | User lacks permission for the requested action | Admin endpoints |

### Resource Errors

| Code | HTTP Status | Description | Endpoints |
|------|-------------|-------------|-----------|
| `NOT_FOUND` | 404 | Requested resource does not exist | All GET, PUT, DELETE endpoints |
| `CONFLICT` | 409 | Resource conflict (e.g., duplicate entry) | POST /auth/register |

### Validation Errors

| Code | HTTP Status | Description | Endpoints |
|------|-------------|-------------|-----------|
| `VALIDATION_ERROR` | 422 | Request validation failed | All POST, PUT endpoints |
| `BAD_REQUEST` | 400 | Malformed request body | All endpoints |

### Business Logic Errors

| Code | HTTP Status | Description | Endpoints |
|------|-------------|-------------|-----------|
| `INVALID_STATE` | 422 | Resource is in invalid state for operation | Training, fight booking |
| `LIMIT_EXCEEDED` | 422 | Rate limit or business limit exceeded | Training sessions |

### Server Errors

| Code | HTTP Status | Description | Endpoints |
|------|-------------|-------------|-----------|
| `INTERNAL_ERROR` | 500 | Unexpected server error | All endpoints |

## Endpoint-Specific Error Codes

### Authentication Endpoints

#### POST /auth/register

| Error Code | Condition |
|------------|-----------|
| `VALIDATION_ERROR` | Missing or invalid email, password, or username |
| `CONFLICT` | Email or username already exists |
| `INTERNAL_ERROR` | Database error during registration |

#### POST /auth/login

| Error Code | Condition |
|------------|-----------|
| `VALIDATION_ERROR` | Missing email or password |
| `UNAUTHORIZED` | Invalid credentials |
| `INTERNAL_ERROR` | Authentication service error |

### Boxer Endpoints

#### GET /boxers/{id}

| Error Code | Condition |
|------------|-----------|
| `NOT_FOUND` | Boxer with ID does not exist |
| `INTERNAL_ERROR` | Database query failed |

#### POST /boxers

| Error Code | Condition |
|------------|-----------|
| `VALIDATION_ERROR` | Missing required fields (name, stats) |
| `INTERNAL_ERROR` | Database insert failed |

#### PUT /boxers/{id}

| Error Code | Condition |
|------------|-----------|
| `NOT_FOUND` | Boxer with ID does not exist |
| `VALIDATION_ERROR` | Invalid stat values (out of range) |
| `INTERNAL_ERROR` | Database update failed |

#### DELETE /boxers/{id}

| Error Code | Condition |
|------------|-----------|
| `NOT_FOUND` | Boxer with ID does not exist |
| `INVALID_STATE` | Boxer has active fights or bookings |
| `INTERNAL_ERROR` | Database delete failed |

### Training Endpoints

#### POST /boxers/{id}/training

| Error Code | Condition |
|------------|-----------|
| `NOT_FOUND` | Boxer with ID does not exist |
| `VALIDATION_ERROR` | Invalid training type or duration |
| `INVALID_STATE` | Boxer is resting or in a fight |
| `LIMIT_EXCEEDED` | Maximum training sessions exceeded |

### Fight Booking Endpoints

#### POST /fights/book

| Error Code | Condition |
|------------|-----------|
| `NOT_FOUND` | One or both boxers not found |
| `INVALID_STATE` | Boxer is unavailable (fighting, resting, training) |
| `CONFLICT` | Fight already scheduled for this time |
| `VALIDATION_ERROR` | Invalid fight parameters |

### Rankings Endpoints

#### GET /rankings

| Error Code | Condition |
|------------|-----------|
| `INTERNAL_ERROR` | Failed to fetch or calculate rankings |
| `VALIDATION_ERROR` | Invalid pagination parameters |

## Implementation Guidelines

### Creating New Error Responses

```go
import "github.com/boxingsim/internal/pkg/errors"

// In your handler:
func GetBoxer(c *gin.Context) {
    id := c.Param("id")
    
    boxer, err := boxerService.GetByID(id)
    if err != nil {
        if errors.IsNotFound(err) {
            apiErr := pkgerrors.NotFound("boxer")
            responseHandler.WriteError(c, apiErr)
            return
        }
        
        apiErr := pkgerrors.Internal("Failed to fetch boxer")
        responseHandler.WriteError(c, apiErr)
        return
    }
    
    responseHandler.WriteSuccess(c, http.StatusOK, boxer)
}
```

### Wrapping Errors with Context

```go
// In your service layer:
func (s *BoxerService) GetByID(id string) (*boxer.Boxer, error) {
    boxer, err := s.store.GetByID(id)
    if err != nil {
        if db.IsNotFound(err) {
            return nil, pkgerrors.NotFound("boxer")
        }
        return nil, pkgerrors.Wrap(err, pkgerrors.ErrorCodeInternal, "Database query failed")
    }
    return boxer, nil
}
```

### Adding New Error Codes

1. Add the error code constant to `internal/pkg/errors/errors.go`
2. Update `GetHTTPStatus()` function with appropriate HTTP status mapping
3. Document in this file under appropriate category
4. List affected endpoints in the table

## Testing Guidelines

All error responses should be tested for:

1. Correct HTTP status code
2. Correct error code in response body
3. Helpful, non-sensitive error message
4. Optional details field populated when applicable

Example test:

```go
func TestGetBoxer_NotFound(t *testing.T) {
    w := PerformRequest(httptest.NewRequest("GET", "/boxers/999", nil))
    
    assert.Equal(t, http.StatusNotFound, w.Code)
    
    var response pkgerrors.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Equal(t, "NOT_FOUND", string(response.Error.Code))
    assert.Contains(t, response.Error.Message, "not found")
}
```
