package apiclient

import (
	"fmt"
	"strings"
)

// APIError represents an error response from the StatusDrift API.
type APIError struct {
	StatusCode int
	Message    string
	Violations []FieldViolation
}

// FieldViolation represents a single field-level validation error.
type FieldViolation struct {
	Field   string `json:"propertyPath"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if len(e.Violations) > 0 {
		var parts []string
		for _, v := range e.Violations {
			parts = append(parts, fmt.Sprintf("%s: %s", v.Field, v.Message))
		}
		return fmt.Sprintf("API error %d: %s", e.StatusCode, strings.Join(parts, "; "))
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404.
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsBadRequest returns true if the error is a 400.
func IsBadRequest(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 400
	}
	return false
}

// IsConflict returns true if the error is a 409.
func IsConflict(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 409
	}
	return false
}
