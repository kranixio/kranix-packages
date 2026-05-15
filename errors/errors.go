package errors

import "fmt"

// KraneError represents a typed error with code and HTTP status.
type KraneError struct {
	Code string `json:"code"`
	HTTP int    `json:"http"`
	Msg  string `json:"message,omitempty"`
}

// Error implements the error interface.
func (e *KraneError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Msg)
	}
	return e.Code
}

// Unwrap allows error wrapping.
func (e *KraneError) Unwrap() error {
	return nil
}

// Predefined error codes.
var (
	ErrWorkloadNotFound    = &KraneError{Code: "WORKLOAD_NOT_FOUND", HTTP: 404}
	ErrNamespaceNotFound   = &KraneError{Code: "NAMESPACE_NOT_FOUND", HTTP: 404}
	ErrInvalidSpec         = &KraneError{Code: "INVALID_SPEC", HTTP: 400}
	ErrBackendUnavailable  = &KraneError{Code: "BACKEND_UNAVAILABLE", HTTP: 503}
	ErrReconcileFailed     = &KraneError{Code: "RECONCILE_FAILED", HTTP: 500}
	ErrUnauthorized        = &KraneError{Code: "UNAUTHORIZED", HTTP: 401}
	ErrForbidden           = &KraneError{Code: "FORBIDDEN", HTTP: 403}
	ErrPodNotFound         = &KraneError{Code: "POD_NOT_FOUND", HTTP: 404}
	ErrConflict            = &KraneError{Code: "CONFLICT", HTTP: 409}
	ErrInternalError       = &KraneError{Code: "INTERNAL_ERROR", HTTP: 500}
	ErrBadRequest          = &KraneError{Code: "BAD_REQUEST", HTTP: 400}
)

// New creates a new KraneError with a custom message.
func New(code string, httpStatus int, message string) *KraneError {
	return &KraneError{
		Code: code,
		HTTP: httpStatus,
		Msg:  message,
	}
}

// Wrap wraps an error with additional context.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsKraneError checks if an error is a KraneError.
func IsKraneError(err error) (*KraneError, bool) {
	if kerr, ok := err.(*KraneError); ok {
		return kerr, true
	}
	return nil, false
}
