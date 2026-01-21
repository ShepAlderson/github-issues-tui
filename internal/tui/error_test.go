package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewErrorComponent(t *testing.T) {
	errorComp := NewErrorComponent()
	if errorComp == nil {
		t.Fatal("NewErrorComponent returned nil")
	}

	// Check initial state
	if errorComp.showModal != false {
		t.Error("New error component should not show modal initially")
	}
	if errorComp.modalTitle != "" {
		t.Error("Modal title should be empty initially")
	}
	if errorComp.modalMessage != "" {
		t.Error("Modal message should be empty initially")
	}
}

func TestErrorComponent_ShowModal(t *testing.T) {
	errorComp := NewErrorComponent()

	// Show a modal error
	errorComp.ShowModal("Critical Error", "Database corruption detected")

	if errorComp.showModal != true {
		t.Error("ShowModal should set showModal to true")
	}
	if errorComp.modalTitle != "Critical Error" {
		t.Error("ShowModal should set correct title")
	}
	if errorComp.modalMessage != "Database corruption detected" {
		t.Error("ShowModal should set correct message")
	}
}

func TestErrorComponent_ShowStatus(t *testing.T) {
	errorComp := NewErrorComponent()

	// Show a status error
	errorComp.ShowStatus("Network timeout", time.Second*5)

	if errorComp.statusMessage != "Network timeout" {
		t.Error("ShowStatus should set status message")
	}
	if errorComp.statusExpiresAt.IsZero() {
		t.Error("ShowStatus should set expiration time")
	}
}

func TestErrorComponent_Init(t *testing.T) {
	errorComp := NewErrorComponent()
	cmd := errorComp.Init()

	// Error component init should return nil command
	if cmd != nil {
		t.Error("ErrorComponent Init should return nil command")
	}
}

func TestErrorComponent_Update(t *testing.T) {
	errorComp := NewErrorComponent()

	// Test window size message
	windowSizeMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updatedComp, cmd := errorComp.Update(windowSizeMsg)
	if updatedComp == nil {
		t.Fatal("Update should return non-nil component")
	}
	if cmd != nil {
		t.Error("Update with WindowSizeMsg should return nil command")
	}

	// Test key message when modal is shown - Enter should dismiss
	errorComp.ShowModal("Test Error", "This is a test error")
	keyMsg := tea.KeyMsg{
		Type: tea.KeyEnter,
	}
	updatedComp, cmd = errorComp.Update(keyMsg)
	if updatedComp.showModal != false {
		t.Error("Enter key should dismiss modal")
	}

	// Test key message when modal is shown - Space should dismiss
	errorComp.ShowModal("Test Error", "This is a test error")
	keyMsg = tea.KeyMsg{
		Type: tea.KeySpace,
	}
	updatedComp, cmd = errorComp.Update(keyMsg)
	if updatedComp.showModal != false {
		t.Error("Space key should dismiss modal")
	}

	// Test key message when modal is shown - Escape should dismiss
	errorComp.ShowModal("Test Error", "This is a test error")
	keyMsg = tea.KeyMsg{
		Type: tea.KeyEscape,
	}
	updatedComp, cmd = errorComp.Update(keyMsg)
	if updatedComp.showModal != false {
		t.Error("Escape key should dismiss modal")
	}

	// Test key message when modal is shown - Other key should not dismiss
	errorComp.ShowModal("Test Error", "This is a test error")
	keyMsg = tea.KeyMsg{
		Type: tea.KeyDown,
	}
	updatedComp, cmd = errorComp.Update(keyMsg)
	if updatedComp.showModal != true {
		t.Error("Down arrow key should not dismiss modal")
	}
}

func TestErrorComponent_View(t *testing.T) {
	errorComp := NewErrorComponent()

	// Test empty view (no errors)
	view := errorComp.View()
	if view != "" {
		t.Error("View should return empty string when no errors")
	}

	// Test modal view
	errorComp.ShowModal("Critical Error", "Database corruption detected\nPlease check your database file.")
	view = errorComp.View()
	if view == "" {
		t.Error("View should not be empty when modal is shown")
	}

	// Check that modal contains expected text
	if !strings.Contains(view, "Critical Error") {
		t.Error("Modal view should contain title")
	}
	if !strings.Contains(view, "Database corruption detected") {
		t.Error("Modal view should contain message")
	}

	// Test status view
	errorComp.ShowStatus("Network timeout occurred", time.Second*3)
	view = errorComp.View()
	if view == "" {
		t.Error("View should not be empty when status message is shown")
	}
	if !strings.Contains(view, "Network timeout occurred") {
		t.Error("Status view should contain message")
	}
}

func TestErrorComponent_CategorizeError(t *testing.T) {
	// Test network errors
	networkError := "dial tcp github.com:443: connect: connection timed out"
	title, message, isCritical := CategorizeError(networkError)
	if title != "Network Error" {
		t.Errorf("Expected 'Network Error', got %s", title)
	}
	if !strings.Contains(message, "Check your internet connection") {
		t.Error("Network error should include connectivity guidance")
	}
	if isCritical != false {
		t.Error("Network error should not be critical")
	}

	// Test authentication errors
	authError := "failed to get authenticated client: invalid token"
	title, message, isCritical = CategorizeError(authError)
	if title != "Authentication Error" {
		t.Errorf("Expected 'Authentication Error', got %s", title)
	}
	if !strings.Contains(message, "Check your GitHub token is valid") {
		t.Error("Authentication error should include token guidance")
	}
	if isCritical != true {
		t.Error("Authentication error should be critical")
	}

	// Test database errors
	dbError := "database disk image is malformed"
	title, message, isCritical = CategorizeError(dbError)
	if title != "Database Error" {
		t.Errorf("Expected 'Database Error', got %s", title)
	}
	if !strings.Contains(message, "Your database file may be corrupted") {
		t.Error("Database error should include database guidance")
	}
	if isCritical != true {
		t.Error("Database error should be critical")
	}

	// Test generic errors
	genericError := "something went wrong"
	title, message, isCritical = CategorizeError(genericError)
	if title != "Error" {
		t.Errorf("Expected 'Error', got %s", title)
	}
	if isCritical != false {
		t.Error("Generic error should not be critical")
	}
}

