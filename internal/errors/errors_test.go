package errors_test

import (
	errlib "errors"
	"testing"

	"github.com/shepbook/ghissues/internal/errors"
)

func TestIsMinorError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"nil error", "", false},
		{"connection refused", "connection refused", true},
		{"connection timeout", "connection timeout", true},
		{"no such host", "no such host", true},
		{"network unreachable", "network is unreachable", true},
		{"i/o timeout", "i/o timeout", true},
		{"generic timeout", "timeout waiting for response", true},
		{"rate limit", "GitHub API rate limit exceeded", true},
		{"rate limit exceeded", "rate limit exceeded", true},
		{"403 forbidden", "403 Forbidden", true},
		{"too many requests", "429 too many requests", true},
		{"invalid token", "invalid token", false}, // critical error, not minor
		{"unauthorized", "401 unauthorized", false},
		{"database locked", "database locked", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errlib.New(tt.errMsg)
			result := errors.IsMinorError(err)
			if result != tt.expected {
				t.Errorf("IsMinorError(%q) = %v, expected %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestIsCriticalError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"nil error", "", false},
		{"invalid token", "invalid token", true},
		{"github token invalid", "invalid GitHub token", false}, // does not match "invalid token" exactly
		{"unauthorized", "401 unauthorized", true},
		{"bad credentials", "bad credentials", true},
		{"database corrupted", "database corrupted", true},
		{"database locked", "database locked", true},
		{"cannot open database", "cannot open database", true},
		{"no such file database", "no such file database", true},
		{"connection refused", "connection refused", false},
		{"rate limit", "rate limit exceeded", false},
		{"generic error", "something went wrong", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errlib.New(tt.errMsg)
			result := errors.IsCriticalError(err)
			if result != tt.expected {
				t.Errorf("IsCriticalError(%q) = %v, expected %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected errors.ErrorCategory
	}{
		{"nil error", "", errors.CategoryMinor},
		{"network error", "connection refused", errors.CategoryMinor},
		{"rate limit", "rate limit exceeded", errors.CategoryMinor},
		{"invalid token", "invalid token", errors.CategoryCritical},
		{"unauthorized", "401 unauthorized", errors.CategoryCritical},
		{"database corrupted", "database corrupted", errors.CategoryCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errlib.New(tt.errMsg)
			result := errors.ClassifyError(err)
			if result != tt.expected {
				t.Errorf("ClassifyError(%q) = %v, expected %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestGetErrorHint(t *testing.T) {
	tests := []struct {
		name           string
		errMsg         string
		expectedMsg    string
		expectedAction string
		expectsRetry   bool
	}{
		{
			name:           "connection refused",
			errMsg:         "connection refused",
			expectedMsg:    "Unable to connect to GitHub",
			expectedAction: "Check your internet connection and try again",
			expectsRetry:   true,
		},
		{
			name:           "timeout",
			errMsg:         "i/o timeout",
			expectedMsg:    "Connection timed out",
			expectedAction: "The request took too long. Check your connection and try again",
			expectsRetry:   true,
		},
		{
			name:           "rate limit",
			errMsg:         "rate limit exceeded",
			expectedMsg:    "GitHub API rate limit exceeded",
			expectedAction: "Wait for the rate limit to reset, or use authentication for higher limits",
			expectsRetry:   true,
		},
		{
			name:           "invalid token",
			errMsg:         "invalid token",
			expectedMsg:    "Invalid or expired GitHub token",
			expectedAction: "Update your token: set GITHUB_TOKEN env var, run 'ghissues config', or ensure 'gh auth' is active",
			expectsRetry:   false,
		},
		{
			name:           "database locked",
			errMsg:         "database locked",
			expectedMsg:    "Database is locked",
			expectedAction: "Another instance may be running. Close it and try again",
			expectsRetry:   false,
		},
		{
			name:           "database corrupted",
			errMsg:         "database corrupted",
			expectedMsg:    "Database is corrupted",
			expectedAction: "Delete the database file and run 'ghissues sync' to recreate it",
			expectsRetry:   false,
		},
		{
			name:           "generic error",
			errMsg:         "something went wrong",
			expectedMsg:    "An error occurred",
			expectedAction: "Check your connection and try again",
			expectsRetry:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errlib.New(tt.errMsg)
			hint := errors.GetErrorHint(err)
			if hint == nil {
				t.Fatalf("GetErrorHint(%q) returned nil", tt.errMsg)
			}
			if hint.Message != tt.expectedMsg {
				t.Errorf("GetErrorHint(%q).Message = %q, expected %q", tt.errMsg, hint.Message, tt.expectedMsg)
			}
			if hint.Action != tt.expectedAction {
				t.Errorf("GetErrorHint(%q).Action = %q, expected %q", tt.errMsg, hint.Action, tt.expectedAction)
			}
			if hint.CanRetry != tt.expectsRetry {
				t.Errorf("GetErrorHint(%q).CanRetry = %v, expected %v", tt.errMsg, hint.CanRetry, tt.expectsRetry)
			}
		})
	}
}

func TestGetErrorHintNil(t *testing.T) {
	hint := errors.GetErrorHint(nil)
	if hint != nil {
		t.Errorf("GetErrorHint(nil) should return nil, got %v", hint)
	}
}

func TestNewUIError(t *testing.T) {
	err := errlib.New("connection refused")
	uiErr := errors.NewUIError(err)

	if uiErr.Err != err {
		t.Errorf("NewUIError.Err = %v, expected %v", uiErr.Err, err)
	}
	if uiErr.Category != errors.CategoryMinor {
		t.Errorf("NewUIError.Category = %v, expected CategoryMinor", uiErr.Category)
	}
	if uiErr.Hint == nil {
		t.Error("NewUIError.Hint should not be nil for connection error")
	}
}

func TestNewUIErrorf(t *testing.T) {
	uiErr := errors.NewUIErrorf("test error occurred")

	if uiErr.Err.Error() != "test error occurred" {
		t.Errorf("NewUIErrorf.Error() = %q, expected %q", uiErr.Err.Error(), "test error occurred")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"network error", "connection refused", true},
		{"timeout", "timeout", true},
		{"rate limit", "rate limit exceeded", true},
		{"invalid token", "invalid token", false},
		{"database corrupted", "database corrupted", false},
		{"generic error", "something went wrong", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errlib.New(tt.errMsg)
			result := errors.IsRetryable(err)
			if result != tt.expected {
				t.Errorf("IsRetryable(%q) = %v, expected %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}