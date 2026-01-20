package errors

import (
	"errors"
	"strings"
)

// ErrorCategory categorizes errors for appropriate UI display
type ErrorCategory int

const (
	// CategoryMinor represents non-critical errors that can be shown in status bar
	CategoryMinor ErrorCategory = iota
	// CategoryCritical represents critical errors that require a modal dialog
	CategoryCritical
)

// ErrorHint provides actionable guidance for error resolution
type ErrorHint struct {
	Message  string
	Action   string
	CanRetry bool
}

// UIError wraps an error with UI-specific metadata
type UIError struct {
	Err      error
	Category ErrorCategory
	Hint     *ErrorHint
}

// Error returns the underlying error message
func (e *UIError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error for error chain handling
func (e *UIError) Unwrap() error {
	return e.Err
}

// IsMinorError returns true if the error is a minor error (network timeout, rate limit)
func IsMinorError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	lowerMsg := strings.ToLower(msg)

	// Check critical errors first (these should not be considered minor)
	if IsCriticalError(err) {
		return false
	}

	// Network-related errors
	if strings.Contains(lowerMsg, "connection refused") ||
		strings.Contains(lowerMsg, "connection timeout") ||
		strings.Contains(lowerMsg, "no such host") ||
		strings.Contains(lowerMsg, "network is unreachable") ||
		strings.Contains(lowerMsg, "i/o timeout") ||
		strings.Contains(lowerMsg, "timeout") {
		return true
	}

	// Rate limit errors
	if strings.Contains(lowerMsg, "rate limit") ||
		strings.Contains(lowerMsg, "rate limit exceeded") ||
		strings.Contains(lowerMsg, "403") ||
		strings.Contains(lowerMsg, "too many requests") {
		return true
	}

	return false
}

// IsCriticalError returns true if the error is critical (invalid token, database corruption)
func IsCriticalError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	lowerMsg := strings.ToLower(msg)

	// Authentication errors - check for invalid token/bad credentials
	if strings.Contains(lowerMsg, "invalid token") ||
		strings.Contains(lowerMsg, "bad credentials") ||
		strings.Contains(lowerMsg, "unauthorized") ||
		strings.Contains(lowerMsg, "401") {
		return true
	}

	// Database errors
	if strings.Contains(lowerMsg, "database") {
		if strings.Contains(lowerMsg, "corrupt") ||
			strings.Contains(lowerMsg, "locked") ||
			strings.Contains(lowerMsg, "cannot open") ||
			strings.Contains(lowerMsg, "no such file") {
			return true
		}
	}

	return false
}

// ClassifyError returns the appropriate category for an error
func ClassifyError(err error) ErrorCategory {
	if err == nil {
		return CategoryMinor
	}

	if IsCriticalError(err) {
		return CategoryCritical
	}

	return CategoryMinor
}

// GetErrorHint returns actionable guidance for an error
func GetErrorHint(err error) *ErrorHint {
	if err == nil {
		return nil
	}

	msg := err.Error()
	lowerMsg := strings.ToLower(msg)

	// Network errors
	if strings.Contains(lowerMsg, "connection refused") ||
		strings.Contains(lowerMsg, "no such host") ||
		strings.Contains(lowerMsg, "network is unreachable") {
		return &ErrorHint{
			Message:  "Unable to connect to GitHub",
			Action:   "Check your internet connection and try again",
			CanRetry: true,
		}
	}

	if strings.Contains(lowerMsg, "timeout") ||
		strings.Contains(lowerMsg, "i/o timeout") {
		return &ErrorHint{
			Message:  "Connection timed out",
			Action:   "The request took too long. Check your connection and try again",
			CanRetry: true,
		}
	}

	// Rate limit errors
	if strings.Contains(lowerMsg, "rate limit") ||
		strings.Contains(lowerMsg, "403") ||
		strings.Contains(lowerMsg, "too many requests") {
		return &ErrorHint{
			Message:  "GitHub API rate limit exceeded",
			Action:   "Wait for the rate limit to reset, or use authentication for higher limits",
			CanRetry: true,
		}
	}

	// Authentication errors
	if strings.Contains(lowerMsg, "invalid token") ||
		strings.Contains(lowerMsg, "bad credentials") ||
		strings.Contains(lowerMsg, "unauthorized") ||
		strings.Contains(lowerMsg, "401") {
		return &ErrorHint{
			Message:  "Invalid or expired GitHub token",
			Action:   "Update your token: set GITHUB_TOKEN env var, run 'ghissues config', or ensure 'gh auth' is active",
			CanRetry: false,
		}
	}

	// Database errors
	if strings.Contains(lowerMsg, "database") {
		if strings.Contains(lowerMsg, "locked") {
			return &ErrorHint{
				Message:  "Database is locked",
				Action:   "Another instance may be running. Close it and try again",
				CanRetry: false,
			}
		}
		if strings.Contains(lowerMsg, "corrupt") {
			return &ErrorHint{
				Message:  "Database is corrupted",
				Action:   "Delete the database file and run 'ghissues sync' to recreate it",
				CanRetry: false,
			}
		}
		if strings.Contains(lowerMsg, "no such file") || strings.Contains(lowerMsg, "cannot open") {
			return &ErrorHint{
				Message:  "Database file not found",
				Action:   "Run 'ghissues sync' to create the database",
				CanRetry: false,
			}
		}
		return &ErrorHint{
			Message:  "Database error occurred",
			Action:   "Check the database path and file permissions",
			CanRetry: false,
		}
	}

	// Default fallback
	return &ErrorHint{
		Message:  "An error occurred",
		Action:   "Check your connection and try again",
		CanRetry: true,
	}
}

// NewUIError creates a new UIError with appropriate categorization
func NewUIError(err error) *UIError {
	return &UIError{
		Err:      err,
		Category: ClassifyError(err),
		Hint:     GetErrorHint(err),
	}
}

// NewUIErrorf creates a new UIError with a formatted message
func NewUIErrorf(format string, args ...any) *UIError {
	err := errors.New(format)
	return NewUIError(err)
}

// IsRetryable returns true if the error allows retry
func IsRetryable(err error) bool {
	hint := GetErrorHint(err)
	if hint != nil {
		return hint.CanRetry
	}
	return !IsCriticalError(err)
}