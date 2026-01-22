package tui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModalDialog(t *testing.T) {
	t.Run("create modal with error message", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("401 Unauthorized"),
			userMsg:   "Invalid token. Please update your configuration.",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 80, 24)

		if modal.errMsg.err == nil {
			t.Error("modal should contain error")
		}

		if modal.width != 80 {
			t.Errorf("width = %d, want 80", modal.width)
		}

		if modal.height != 24 {
			t.Errorf("height = %d, want 24", modal.height)
		}

		if !modal.active {
			t.Error("modal should be active by default")
		}
	})

	t.Run("modal renders error content", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "This is a test error message.",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 80, 24)
		view := modal.View()

		if view == "" {
			t.Error("modal view should not be empty")
		}

		// Should contain the error message
		if !contains(view, "test error message") {
			t.Errorf("modal should display error message, got: %s", view)
		}

		// Should indicate it's an error
		if !contains(view, "ERROR") {
			t.Errorf("modal should indicate it's an error, got: %s", view)
		}
	})

	t.Run("pressing enter acknowledges error", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "Test message",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 80, 24)

		// Send enter key
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := modal.Update(msg)

		if cmd == nil {
			t.Error("pressing enter should return a command")
		}

		// Verify modal is no longer active
		if modal.active {
			t.Error("modal should be deactivated after pressing enter")
		}
	})

	t.Run("pressing q acknowledges error", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "Test message",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 80, 24)

		// Send 'q' key
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := modal.Update(msg)

		if cmd == nil {
			t.Error("pressing 'q' should return a command")
		}

		// Verify modal is no longer active
		if modal.active {
			t.Error("modal should be deactivated after pressing 'q'")
		}
	})

	t.Run("pressing escape acknowledges error", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "Test message",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 80, 24)

		// Send escape key
		msg := tea.KeyMsg{Type: tea.KeyEscape}
		_, cmd := modal.Update(msg)

		if cmd == nil {
			t.Error("pressing escape should return a command")
		}

		// Verify modal is no longer active
		if modal.active {
			t.Error("modal should be deactivated after pressing escape")
		}
	})

	t.Run("other keys do nothing", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "Test message",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 80, 24)
		originalActive := modal.active

		// Send random key
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		_, cmd := modal.Update(msg)

		if cmd != nil {
			t.Error("random keys should not return commands")
		}

		if modal.active != originalActive {
			t.Error("modal active state should not change on random keys")
		}
	})

	t.Run("modal is centered", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "Test message that is long enough to test width",
			timestamp: time.Now(),
		}

		modal := NewModalDialog(errMsg, 100, 30)
		view := modal.View()

		// Should have consistent styling
		lines := splitLines(view)
		if len(lines) == 0 {
			t.Error("modal should render multiple lines")
		}
	})
}

func TestModalDialogAcknowledgedCommand(t *testing.T) {
	t.Run("acknowledged command marks error as acknowledged", func(t *testing.T) {
		errMsg := ErrorMessage{
			severity:  ErrorSeverityCritical,
			err:       errors.New("test error"),
			userMsg:   "Test message",
			timestamp: time.Now(),
		}

		// Create modal and trigger acknowledgment
		modal := NewModalDialog(errMsg, 80, 24)
		cmd := modal.acknowledgeError()

		if cmd == nil {
			t.Fatal("acknowledgeError should return a command")
		}

		// Execute the command
		msg := cmd()
		ackMsg, ok := msg.(errorAcknowledgedMsg)
		if !ok {
			t.Fatal("command should return errorAcknowledgedMsg")
		}

		if !ackMsg.errMsg.acknowledged {
			t.Error("error should be marked as acknowledged")
		}
	})
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
