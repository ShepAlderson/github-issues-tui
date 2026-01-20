package tui

import (
	"errors"
	"testing"
	"time"
)

func TestErrorSeverity(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantSeverity  ErrorSeverity
		wantUserMsg   string
		wantShowModal bool
	}{
		{
			name:          "network timeout",
			err:           errors.New("dial tcp: i/o timeout"),
			wantSeverity:  ErrorSeverityMinor,
			wantUserMsg:   "Network timeout. Please check your connection and try again.",
			wantShowModal: false,
		},
		{
			name:          "rate limit exceeded",
			err:           errors.New("API rate limit exceeded"),
			wantSeverity:  ErrorSeverityMinor,
			wantUserMsg:   "GitHub API rate limit exceeded. Please wait a moment and try again.",
			wantShowModal: false,
		},
		{
			name:          "invalid token",
			err:           errors.New("401 Unauthorized"),
			wantSeverity:  ErrorSeverityCritical,
			wantUserMsg:   "Invalid GitHub token. Please run 'ghissies config' to update your token.",
			wantShowModal: true,
		},
		{
			name:          "token permission error",
			err:           errors.New("403: Token does not have required permissions"),
			wantSeverity:  ErrorSeverityCritical,
			wantUserMsg:   "GitHub token missing required permissions. Please update token with 'repo' scope.",
			wantShowModal: true,
		},
		{
			name:          "database corruption",
			err:           errors.New("database disk image is malformed"),
			wantSeverity:  ErrorSeverityCritical,
			wantUserMsg:   "Database corruption detected. Please run 'ghissies sync' to rebuild.",
			wantShowModal: true,
		},
		{
			name:          "database locked",
			err:           errors.New("database is locked"),
			wantSeverity:  ErrorSeverityMinor,
			wantUserMsg:   "Database temporarily locked. Retry in a moment.",
			wantShowModal: false,
		},
		{
			name:          "connection refused",
			err:           errors.New("connect: connection refused"),
			wantSeverity:  ErrorSeverityMinor,
			wantUserMsg:   "Cannot connect to GitHub. Please check your internet connection.",
			wantShowModal: false,
		},
		{
			name:          "repository not found",
			err:           errors.New("404 Not Found"),
			wantSeverity:  ErrorSeverityCritical,
			wantUserMsg:   "Repository not found. Please check the repository name in your config.",
			wantShowModal: true,
		},
		{
			name:          "generic wrapped error",
			err:           errors.New("some other error"),
			wantSeverity:  ErrorSeverityMinor,
			wantUserMsg:   "An unexpected error occurred: some other error",
			wantShowModal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity, userMsg, showModal := classifyError(tt.err)

			if severity != tt.wantSeverity {
				t.Errorf("classifyError() severity = %v, want %v", severity, tt.wantSeverity)
			}

			if userMsg != tt.wantUserMsg {
				t.Errorf("classifyError() userMsg = %q, want %q", userMsg, tt.wantUserMsg)
			}

			if showModal != tt.wantShowModal {
				t.Errorf("classifyError() showModal = %v, want %v", showModal, tt.wantShowModal)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	t.Run("create minor error message", func(t *testing.T) {
		err := errors.New("network timeout")
		msg := newErrorMessage(err)

		if msg.severity != ErrorSeverityMinor {
			t.Errorf("severity = %v, want ErrorSeverityMinor", msg.severity)
		}

		if !msg.timestamp.Before(time.Now().Add(time.Second)) {
			t.Errorf("timestamp not set correctly")
		}

		if msg.acknowledged != false {
			t.Errorf("acknowledged = %v, want false", msg.acknowledged)
		}
	})

	t.Run("create critical error message", func(t *testing.T) {
		err := errors.New("401 Unauthorized")
		msg := newErrorMessage(err)

		if msg.severity != ErrorSeverityCritical {
			t.Errorf("severity = %v, want ErrorSeverityCritical", msg.severity)
		}
	})

	t.Run("error message includes user-friendly text", func(t *testing.T) {
		err := errors.New("API rate limit exceeded")
		msg := newErrorMessage(err)

		if msg.err == nil {
			t.Error("error message should wrap original error")
		}

		if msg.userMsg == "" {
			t.Error("userMsg should not be empty")
		}
	})
}

func TestErrorSeverityString(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		want     string
	}{
		{ErrorSeverityMinor, "Minor"},
		{ErrorSeverityCritical, "Critical"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("ErrorSeverity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
