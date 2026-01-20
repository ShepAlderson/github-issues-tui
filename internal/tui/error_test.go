package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewErrorModel(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		severity      ErrorSeverity
		wantActive    bool
		wantMessage   string
		wantGuidance  string
	}{
		{
			name:         "critical error with guidance",
			err:          errors.New("authentication failed: invalid GitHub token"),
			severity:     ErrorSeverityCritical,
			wantActive:   true,
			wantMessage:  "authentication failed: invalid GitHub token",
			wantGuidance: "Please check your GitHub token:\n  • Set GITHUB_TOKEN environment variable\n  • Run 'ghissues config' to update token\n  • Verify token has 'repo' scope",
		},
		{
			name:         "minor error",
			err:          errors.New("network timeout"),
			severity:     ErrorSeverityMinor,
			wantActive:   false, // Minor errors don't show modal
			wantMessage:  "network timeout",
			wantGuidance: "Network issue detected:\n  • Check your internet connection\n  • Verify GitHub.com is accessible\n  • Try again in a moment (press 'r' to refresh)",
		},
		{
			name:         "unknown error defaults to critical",
			err:          errors.New("something broke"),
			severity:     ErrorSeverityCritical,
			wantActive:   true,
			wantMessage:  "something broke",
			wantGuidance: "An unexpected error occurred.\n  • Try again (press 'r' to refresh)\n  • Run 'ghissues sync' to re-sync data\n  • Check logs for more details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewErrorModel(tt.err, tt.severity)

			if model.Active != tt.wantActive {
				t.Errorf("NewErrorModel() Active = %v, want %v", model.Active, tt.wantActive)
			}

			if model.Message != tt.wantMessage {
				t.Errorf("NewErrorModel() Message = %v, want %v", model.Message, tt.wantMessage)
			}

			if model.Guidance != tt.wantGuidance {
				t.Errorf("NewErrorModel() Guidance = %v, want %v", model.Guidance, tt.wantGuidance)
			}

			if model.Severity != tt.severity {
				t.Errorf("NewErrorModel() Severity = %v, want %v", model.Severity, tt.severity)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantSev  ErrorSeverity
		wantText string
	}{
		{
			name:     "authentication error - critical",
			err:      errors.New("authentication failed: invalid GitHub token"),
			wantSev:  ErrorSeverityCritical,
			wantText: "invalid GitHub token",
		},
		{
			name:     "unauthorized error - critical",
			err:      errors.New("authentication failed: invalid GitHub token"),
			wantSev:  ErrorSeverityCritical,
			wantText: "invalid GitHub token",
		},
		{
			name:     "repository not found - critical",
			err:      errors.New("repository not found: owner/repo"),
			wantSev:  ErrorSeverityCritical,
			wantText: "repository not found",
		},
		{
			name:     "database corruption - critical",
			err:      errors.New("database is malformed"),
			wantSev:  ErrorSeverityCritical,
			wantText: "database corrupted",
		},
		{
			name:     "network timeout - minor",
			err:      errors.New("failed to fetch issues: timeout"),
			wantSev:  ErrorSeverityMinor,
			wantText: "network timeout",
		},
		{
			name:     "rate limit - minor",
			err:      errors.New("rate limit exceeded"),
			wantSev:  ErrorSeverityMinor,
			wantText: "rate limit",
		},
		{
			name:     "context canceled - minor",
			err:      errors.New("sync cancelled"),
			wantSev:  ErrorSeverityMinor,
			wantText: "cancelled",
		},
		{
			name:     "unknown error - defaults to critical",
			err:      errors.New("something unknown happened"),
			wantSev:  ErrorSeverityCritical,
			wantText: "unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity, guidance := ClassifyError(tt.err)

			if severity != tt.wantSev {
				t.Errorf("ClassifyError() severity = %v, want %v", severity, tt.wantSev)
			}

			if guidance == "" && tt.wantSev == ErrorSeverityCritical {
				t.Errorf("ClassifyError() guidance should not be empty for critical errors")
			}
		})
	}
}

func TestErrorModelUpdate(t *testing.T) {
	t.Run("dismiss error on any key when complete", func(t *testing.T) {
		model := NewErrorModel(errors.New("test error"), ErrorSeverityCritical)
		model.Active = true
		model.Complete = true

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updated, cmd := model.Update(msg)

		if updated.Active {
			t.Error("Error modal should be dismissed after key press when complete")
		}

		if cmd != nil {
			t.Error("Expected no command, got non-nil")
		}
	})

	t.Run("do not dismiss when not complete", func(t *testing.T) {
		model := NewErrorModel(errors.New("test error"), ErrorSeverityCritical)
		model.Active = true
		model.Complete = false

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updated, _ := model.Update(msg)

		if !updated.Active {
			t.Error("Error modal should not be dismissed when not complete")
		}
	})

	t.Run("window size message updates dimensions", func(t *testing.T) {
		model := NewErrorModel(errors.New("test error"), ErrorSeverityCritical)

		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		updated, _ := model.Update(msg)

		if updated.Width != 100 {
			t.Errorf("Expected Width = 100, got %d", updated.Width)
		}
		if updated.Height != 50 {
			t.Errorf("Expected Height = 50, got %d", updated.Height)
		}
	})
}

func TestErrorModelView(t *testing.T) {
	t.Run("view returns empty string when not active", func(t *testing.T) {
		model := NewErrorModel(errors.New("test error"), ErrorSeverityCritical)
		model.Active = false

		view := model.View()
		if view != "" {
			t.Errorf("View() should return empty string when not active, got %q", view)
		}
	})

	t.Run("view renders error modal when active", func(t *testing.T) {
		model := NewErrorModel(errors.New("test error"), ErrorSeverityCritical)
		model.Active = true
		model.Width = 80
		model.Height = 24

		view := model.View()
		if view == "" {
			t.Error("View() should not return empty string when active")
		}

		// Check that key elements are present
		if !contains(view, "Error") {
			t.Error("View should contain 'Error'")
		}
		if !contains(view, "test error") {
			t.Error("View should contain the error message")
		}
	})

	t.Run("view shows dismiss hint when complete", func(t *testing.T) {
		model := NewErrorModel(errors.New("test error"), ErrorSeverityCritical)
		model.Active = true
		model.Complete = true
		model.Width = 80
		model.Height = 24

		view := model.View()
		if !contains(view, "Press") || !contains(view, "key") {
			t.Error("View should show dismiss hint when complete")
		}
	})
}

func TestGetErrorMessageForStatusBar(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		severity ErrorSeverity
		want     string
	}{
		{
			name:     "minor error returns message",
			err:      errors.New("network timeout"),
			severity: ErrorSeverityMinor,
			want:     "Error: network timeout",
		},
		{
			name:     "critical error returns empty for status bar",
			err:      errors.New("authentication failed"),
			severity: ErrorSeverityCritical,
			want:     "", // Critical errors show modal, not status bar
		},
		{
			name:     "nil error returns empty",
			err:      nil,
			severity: ErrorSeverityMinor,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorMessageForStatusBar(tt.err, tt.severity)
			if got != tt.want {
				t.Errorf("GetErrorMessageForStatusBar() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetActionableGuidance(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantType string // "auth", "repo", "db", "network", or "default"
	}{
		{
			name:     "authentication error",
			err:      errors.New("authentication failed: invalid GitHub token"),
			wantType: "auth",
		},
		{
			name:     "repository not found",
			err:      errors.New("repository not found: owner/repo"),
			wantType: "repo",
		},
		{
			name:     "database error",
			err:      errors.New("database is locked"),
			wantType: "default", // "locked" doesn't match our DB error patterns, so it falls to default
		},
		{
			name:     "network error",
			err:      errors.New("timeout"),
			wantType: "network",
		},
		{
			name:     "unknown error",
			err:      errors.New("something broke"),
			wantType: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guidance := GetActionableGuidance(tt.err)

			// Check that guidance is not empty for all error types
			if guidance == "" {
				t.Error("GetActionableGuidance() should not return empty string")
			}

			// Verify the guidance contains relevant keywords based on type
			switch tt.wantType {
			case "auth":
				if !contains(guidance, "token") && !contains(guidance, "GITHUB_TOKEN") {
					t.Error("Auth guidance should mention token or GITHUB_TOKEN")
				}
			case "repo":
				if !contains(guidance, "repository") && !contains(guidance, "repo") {
					t.Error("Repo guidance should mention repository")
				}
			case "db", "default":
				// Default guidance should have general helpful suggestions
				if !contains(guidance, "Try again") && !contains(guidance, "refresh") {
					t.Error("Default/DB guidance should mention trying again or refresh")
				}
			case "network":
				if !contains(guidance, "network") && !contains(guidance, "connection") && !contains(guidance, "retry") {
					t.Error("Network guidance should mention network, connection, or retry")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
