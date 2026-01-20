package tui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRefreshModel(t *testing.T) {
	model := NewRefreshModel()

	if model.Active {
		t.Errorf("Expected new model to be inactive, got active")
	}
	if model.Complete {
		t.Errorf("Expected new model to be incomplete, got complete")
	}
}

func TestRefreshModelUpdateProgress(t *testing.T) {
	model := NewRefreshModel()

	// Send progress message
	msg := RefreshProgressMsg{
		Current: 5,
		Total:   10,
		Message: "Fetching...",
	}
	updated, _ := model.Update(msg)

	if !updated.Active {
		t.Errorf("Expected model to be active after progress message")
	}
	if updated.Current != 5 {
		t.Errorf("Expected current to be 5, got %d", updated.Current)
	}
	if updated.Total != 10 {
		t.Errorf("Expected total to be 10, got %d", updated.Total)
	}
	if updated.Message != "Fetching..." {
		t.Errorf("Expected message 'Fetching...', got '%s'", updated.Message)
	}
}

func TestRefreshModelUpdateComplete(t *testing.T) {
	model := NewRefreshModel()

	// Send complete message
	msg := RefreshCompleteMsg{
		Success: true,
		Error:   nil,
	}
	updated, _ := model.Update(msg)

	if updated.Active {
		t.Errorf("Expected model to be inactive after complete message")
	}
	if !updated.Complete {
		t.Errorf("Expected model to be complete after complete message")
	}
	if !updated.Success {
		t.Errorf("Expected success to be true, got false")
	}
}

func TestRefreshModelUpdateCompleteWithError(t *testing.T) {
	model := NewRefreshModel()

	// Send complete message with error
	err := fmt.Errorf("test error")
	msg := RefreshCompleteMsg{
		Success: false,
		Error:   err,
	}
	updated, _ := model.Update(msg)

	if updated.Success {
		t.Errorf("Expected success to be false with error, got true")
	}
}

func TestRefreshModelView(t *testing.T) {
	model := NewRefreshModel()

	// Inactive model should return empty string
	view := model.View()
	if view != "" {
		t.Errorf("Expected empty view for inactive model, got '%s'", view)
	}

	// Send progress message
	msg := RefreshProgressMsg{
		Current: 5,
		Total:   10,
	}
	updated, _ := model.Update(msg)

	// Active model should return non-empty view
	view = updated.View()
	if view == "" {
		t.Errorf("Expected non-empty view for active model")
	}
}

func TestRefreshModelReset(t *testing.T) {
	model := NewRefreshModel()

	// Activate and complete the model
	msg1 := RefreshProgressMsg{Current: 5, Total: 10}
	updated, _ := model.Update(msg1)
	model = updated

	msg2 := RefreshCompleteMsg{Success: true}
	updated, _ = model.Update(msg2)
	model = updated

	// Reset the model
	model.Reset()

	if model.Active {
		t.Errorf("Expected model to be inactive after reset")
	}
	if model.Complete {
		t.Errorf("Expected model to be incomplete after reset")
	}
	if model.Current != 0 {
		t.Errorf("Expected current to be 0 after reset, got %d", model.Current)
	}
	if model.Total != 0 {
		t.Errorf("Expected total to be 0 after reset, got %d", model.Total)
	}
	if model.Message != "" {
		t.Errorf("Expected message to be empty after reset, got '%s'", model.Message)
	}
}

func TestRefreshModelWindowSize(t *testing.T) {
	model := NewRefreshModel()

	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}
	updated, _ := model.Update(msg)

	if updated.Width != 80 {
		t.Errorf("Expected width 80, got %d", updated.Width)
	}
	if updated.Height != 24 {
		t.Errorf("Expected height 24, got %d", updated.Height)
	}
}
