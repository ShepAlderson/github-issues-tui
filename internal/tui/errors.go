package tui

import (
	"strings"
	"time"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	// ErrorSeverityMinor represents non-critical errors that can be shown in status bar
	ErrorSeverityMinor ErrorSeverity = iota
	// ErrorSeverityCritical represents critical errors that require modal display
	ErrorSeverityCritical
)

func (s ErrorSeverity) String() string {
	switch s {
	case ErrorSeverityMinor:
		return "Minor"
	case ErrorSeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// ErrorMessage represents a classified error with user-friendly messaging
type ErrorMessage struct {
	severity     ErrorSeverity
	err          error
	userMsg      string
	timestamp    time.Time
	acknowledged bool
}

// newErrorMessage creates a new ErrorMessage from an error
func newErrorMessage(err error) ErrorMessage {
	severity, userMsg, _ := classifyError(err)
	return ErrorMessage{
		severity:     severity,
		err:          err,
		userMsg:      userMsg,
		timestamp:    time.Now(),
		acknowledged: false,
	}
}

// classifyError determines the severity and user message for an error
func classifyError(err error) (ErrorSeverity, string, bool) {
	errStr := err.Error()

	// Critical errors - authentication and configuration issues
	if strings.Contains(errStr, "401 Unauthorized") {
		return ErrorSeverityCritical,
			"Invalid GitHub token. Please run 'ghissies config' to update your token.",
			true
	}

	if strings.Contains(errStr, "403") && strings.Contains(errStr, "permissions") {
		return ErrorSeverityCritical,
			"GitHub token missing required permissions. Please update token with 'repo' scope.",
			true
	}

	if strings.Contains(errStr, "404 Not Found") {
		return ErrorSeverityCritical,
			"Repository not found. Please check the repository name in your config.",
			true
	}

	if strings.Contains(errStr, "database disk image is malformed") {
		return ErrorSeverityCritical,
			"Database corruption detected. Please run 'ghissies sync' to rebuild.",
			true
	}

	// Minor errors - transient network and rate limit issues
	if strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "timeout") {
		return ErrorSeverityMinor,
			"Network timeout. Please check your connection and try again.",
			false
	}

	if strings.Contains(errStr, "connection refused") {
		return ErrorSeverityMinor,
			"Cannot connect to GitHub. Please check your internet connection.",
			false
	}

	if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "Rate limit") {
		return ErrorSeverityMinor,
			"GitHub API rate limit exceeded. Please wait a moment and try again.",
			false
	}

	if strings.Contains(errStr, "database is locked") {
		return ErrorSeverityMinor,
			"Database temporarily locked. Retry in a moment.",
			false
	}

	// Default to minor for unknown errors
	return ErrorSeverityMinor,
		"An unexpected error occurred: " + err.Error(),
		false
}

// IsValid checks if the error message is valid (not acknowledged and recent)
func (e ErrorMessage) IsValid() bool {
	if e.acknowledged {
		return false
	}
	// Error is valid for 10 seconds
	return time.Since(e.timestamp) < 10*time.Second
}

// MarkAcknowledged marks the error as acknowledged
func (e *ErrorMessage) MarkAcknowledged() {
	e.acknowledged = true
}
