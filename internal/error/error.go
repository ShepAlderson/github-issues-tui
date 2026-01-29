package error

import (
	"errors"
	"net"
	"strings"
)

// ErrorSeverity indicates the severity level of an error
type ErrorSeverity int

const (
	// SeverityMinor indicates non-blocking errors (network timeout, rate limit)
	// These are shown in the status bar
	SeverityMinor ErrorSeverity = iota

	// SeverityCritical indicates blocking errors (invalid token, database corruption)
	// These are shown as modal dialogs
	SeverityCritical
)

// IsMinor returns true if the error is minor
func (s ErrorSeverity) IsMinor() bool {
	return s == SeverityMinor
}

// IsCritical returns true if the error is critical
func (s ErrorSeverity) IsCritical() bool {
	return s == SeverityCritical
}

// AppError represents a classified application error with user-friendly messaging
type AppError struct {
	Original error
	Severity ErrorSeverity
	Display  string
	Guidance string
	Actionable bool
	Retryable  bool
}

// Error returns the error string (implements error interface)
func (e AppError) Error() string {
	if e.Guidance != "" {
		return e.Display + ": " + e.Original.Error()
	}
	return e.Display + ": " + e.Original.Error()
}

// DisplayMessage returns the user-facing message with guidance
func (e AppError) DisplayMessage() string {
	if e.Guidance != "" {
		return e.Display + "\n\n" + e.Guidance
	}
	return e.Display
}

// Unwrap returns the original error
func (e AppError) Unwrap() error {
	return e.Original
}

// Classify analyzes an error and returns a classified AppError
func Classify(err error) AppError {
	if err == nil {
		return AppError{}
	}

	errStr := err.Error()

	// Check for authentication errors
	if isAuthError(errStr) {
		return AppError{
			Original:   err,
			Severity:   SeverityCritical,
			Display:    "Authentication failed",
			Guidance:   "Please check your GitHub token.\nRun 'gh auth login' or set GITHUB_TOKEN environment variable.",
			Actionable: true,
			Retryable:  false,
		}
	}

	// Check for database corruption/errors
	if isDatabaseError(errStr) {
		return AppError{
			Original:   err,
			Severity:   SeverityCritical,
			Display:    "Database error",
			Guidance:   "The local database may be corrupted.\nTry removing the database file and run 'ghissues sync' again.",
			Actionable: true,
			Retryable:  false,
		}
	}

	// Check for rate limit errors
	if isRateLimitError(errStr) {
		return AppError{
			Original:   err,
			Severity:   SeverityMinor,
			Display:    "Rate limit exceeded",
			Guidance:   "GitHub API rate limit reached. Wait a few minutes and try again.",
			Actionable: true,
			Retryable:  true,
		}
	}

	// Check for network errors
	if isNetworkError(err) {
		return AppError{
			Original:   err,
			Severity:   SeverityMinor,
			Display:    "Network error",
			Guidance:   "Check your internet connection and try again with 'r'.",
			Actionable: true,
			Retryable:  true,
		}
	}

	// Default: minor error
	return AppError{
		Original:   err,
		Severity:   SeverityMinor,
		Display:    "An error occurred",
		Guidance:   "",
		Actionable: false,
		Retryable:  false,
	}
}

// isAuthError checks if error is related to authentication
func isAuthError(errStr string) bool {
	authPatterns := []string{
		"Bad credentials",
		"status 401",
		"Unauthorized",
		"token is invalid",
		"token is expired",
	}
	lowerErr := strings.ToLower(errStr)
	for _, pattern := range authPatterns {
		if strings.Contains(lowerErr, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isDatabaseError checks if error is related to database corruption
func isDatabaseError(errStr string) bool {
	dbPatterns := []string{
		"database disk image is malformed",
		"database corruption",
		"database is locked",
		"readonly database",
		"SQLITE_CORRUPT",
	}
	lowerErr := strings.ToLower(errStr)
	for _, pattern := range dbPatterns {
		if strings.Contains(lowerErr, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isRateLimitError checks if error is related to rate limiting
func isRateLimitError(errStr string) bool {
	ratePatterns := []string{
		"rate limit",
		"status 403",
		"Forbidden",
	}
	lowerErr := strings.ToLower(errStr)
	for _, pattern := range ratePatterns {
		if strings.Contains(lowerErr, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isNetworkError checks if error is related to network issues
func isNetworkError(err error) bool {
	// Check for net.Error interface
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Check for *net.OpError
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check error string patterns
	errStr := strings.ToLower(err.Error())
	networkPatterns := []string{
		"timeout",
		"timed out",
		"connection refused",
		"no such host",
		"connection reset",
		"network is unreachable",
		"temporary failure in name resolution",
		"dial tcp",
		"i/o timeout",
	}

	for _, pattern := range networkPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
