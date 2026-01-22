package tui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppModelErrorHandling(t *testing.T) {
	t.Run("display minor error in status bar", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		// Simulate a network error
		err := errors.New("dial tcp: i/o timeout")
		app.setError(err)

		if app.errorMsg.err == nil {
			t.Error("error should be set")
		}

		if app.errorMsg.severity != ErrorSeverityMinor {
			t.Errorf("error severity = %v, want ErrorSeverityMinor", app.errorMsg.severity)
		}

		// Should not create modal for minor errors
		if app.errorModal.IsActive() {
			t.Error("modal should not be active for minor errors")
		}
	})

	t.Run("display critical error as modal", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		// Simulate an auth error
		err := errors.New("401 Unauthorized")
		app.setError(err)

		if app.errorMsg.severity != ErrorSeverityCritical {
			t.Errorf("error severity = %v, want ErrorSeverityCritical", app.errorMsg.severity)
		}

		// Should create modal for critical errors
		if !app.errorModal.IsActive() {
			t.Error("modal should be active for critical errors")
		}
	})

	t.Run("clear error", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		app.setError(errors.New("test error"))

		// Clear the error
		app.clearError()

		// After clear, errorMsg should be zero value (no error, not acknowledged)
		// which means the error should not be valid
		if app.errorMsg.IsValid() {
			t.Error("error should not be valid after clear")
		}

		if app.errorModal.IsActive() {
			t.Error("modal should be deactivated")
		}
	})

	t.Run("modal blocks other input when active", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		// Set a critical error
		app.setError(errors.New("401 Unauthorized"))

		// Try to send a regular key
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		_, _ = app.Update(msg)

		// Modal should handle the key or ignore it (no crash)
		// The test passes if we reach this point without errors
	})

	t.Run("acknowledging modal error clears it", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		// Set a critical error
		app.setError(errors.New("401 Unauthorized"))

		// Verify modal is active
		if !app.errorModal.IsActive() {
			t.Fatal("modal should be active")
		}

		// Press enter to acknowledge
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		model, cmd := app.Update(msg)

		// Command should be for acknowledgment
		if cmd == nil {
			t.Error("should return acknowledgment command")
		}

		// Cast back to AppModel
		app, ok := model.(*AppModel)
		if !ok {
			t.Fatal("model should be *AppModel")
		}

		// Modal should be deactivated
		if app.errorModal.IsActive() {
			t.Error("modal should be deactivated after acknowledgment")
		}
	})

	t.Run("status bar shows error for minor issues", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		// Set a network timeout error (minor)
		app.setError(errors.New("dial tcp: i/o timeout"))

		// Create a minimal list model for rendering
		app.listModel = &ListModel{
			width:  80,
			height: 24,
		}

		view := app.View()

		// Should show network error guidance in status bar
		if !contains(view, "Network timeout") && !contains(view, "connection") {
			t.Errorf("status bar should show network error message, got: %s", view)
		}
	})

	t.Run("error expires after timeout", func(t *testing.T) {
		app := &AppModel{
			currentView: ListView,
			width:       80,
			height:      24,
		}

		// Set an error
		app.setError(errors.New("test error"))

		// Fast forward time
		app.errorMsg.timestamp = time.Now().Add(-15 * time.Second)

		// Error should be expired
		if app.errorMsg.IsValid() {
			t.Error("error should be expired after timeout")
		}
	})
}
