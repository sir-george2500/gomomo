package gomomo

import (
	"errors"
	"fmt"
)

// Pre-defined errors
var (
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrAPIRequestFailed     = errors.New("API request failed")
	ErrInvalidResponse      = errors.New("invalid response from API")
	ErrTransactionFailed    = errors.New("transaction failed")
)

// MoMoError represents a MTN MoMo API error
type MoMoError struct {
	Code       string
	Message    string
	StatusCode int
	Details    map[string]interface{}
}

// Error implements the error interface
func (e *MoMoError) Error() string {
	return fmt.Sprintf("MTN MoMo API error: %s (%s), status: %d", e.Message, e.Code, e.StatusCode)
}

// NewMoMoError creates a new MoMo error
func NewMoMoError(code, message string, statusCode int, details map[string]interface{}) *MoMoError {
	return &MoMoError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
