package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestNewHelpModel verifies that a new help model is created with proper defaults
func TestNewHelpModel(t *testing.T) {
	model := NewHelpModel()

	if model.Active != false {
		t.Errorf("Expected Active to be false, got %v", model.Active)
	}

	if model.Width != 80 {
		t.Errorf("Expected default Width to be 80, got %d", model.Width)
	}

	if model.Height != 24 {
		t.Errorf("Expected default Height to be 24, got %d", model.Height)
	}
}

// TestHelpModelToggle verifies that help can be toggled on and off
func TestHelpModelToggle(t *testing.T) {
	model := NewHelpModel()

	// Toggle on
	model.Toggle()
	if !model.Active {
		t.Errorf("Expected Active to be true after toggle, got false")
	}

	// Toggle off
	model.Toggle()
	if model.Active {
		t.Errorf("Expected Active to be false after second toggle, got true")
	}
}

// TestHelpModelUpdateDismissOnQuestion verifies help dismisses with '?' key
func TestHelpModelUpdateDismissOnQuestion(t *testing.T) {
	model := NewHelpModel()
	model.Active = true

	// Send '?' key to dismiss
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, _ := model.Update(msg)

	if updated.Active {
		t.Errorf("Expected help to be dismissed with '?', but Active is still true")
	}
}

// TestHelpModelUpdateDismissOnEsc verifies help dismisses with 'Esc' key
func TestHelpModelUpdateDismissOnEsc(t *testing.T) {
	model := NewHelpModel()
	model.Active = true

	// Send 'esc' key to dismiss
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updated, _ := model.Update(msg)

	if updated.Active {
		t.Errorf("Expected help to be dismissed with 'Esc', but Active is still true")
	}
}

// TestHelpModelUpdateWindowSize verifies window size messages update dimensions
func TestHelpModelUpdateWindowSize(t *testing.T) {
	model := NewHelpModel()

	msg := tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	}

	updated, _ := model.Update(msg)

	if updated.Width != 120 {
		t.Errorf("Expected Width to be 120, got %d", updated.Width)
	}

	if updated.Height != 40 {
		t.Errorf("Expected Height to be 40, got %d", updated.Height)
	}
}

// TestHelpModelViewWhenInactive verifies view returns empty string when inactive
func TestHelpModelViewWhenInactive(t *testing.T) {
	model := NewHelpModel()
	model.Active = false

	view := model.View()

	if view != "" {
		t.Errorf("Expected empty view when inactive, got %q", view)
	}
}

// TestHelpModelViewWhenActive verifies view renders help content when active
func TestHelpModelViewWhenActive(t *testing.T) {
	model := NewHelpModel()
	model.Active = true
	model.Width = 100
	model.Height = 30

	view := model.View()

	if view == "" {
		t.Errorf("Expected non-empty view when active")
	}

	// Check for key sections
	viewLower := strings.ToLower(view)

	// Should contain main view section
	if !strings.Contains(viewLower, "main view") {
		t.Errorf("Expected view to contain 'Main View' section")
	}

	// Should contain common keys like j/k for navigation
	if !strings.Contains(view, "j") || !strings.Contains(view, "k") {
		t.Errorf("Expected view to contain j/k navigation keys")
	}

	// Should contain help hint at bottom
	if !strings.Contains(viewLower, "dismiss") {
		t.Errorf("Expected view to contain dismissal hint")
	}
}

// TestHelpModelContainsAllKeybindings verifies all main view keybindings are documented
func TestHelpModelContainsAllKeybindings(t *testing.T) {
	model := NewHelpModel()
	model.Active = true
	model.Width = 120
	model.Height = 40

	view := model.View()

	// Main view keybindings
	expectedKeys := []string{
		"j", "k",           // Navigation
		"Enter",            // Open comments (capitalized in help)
		"Space",            // Select (shown as "Space" in help)
		"s", "S",           // Sort
		"m",                // Toggle markdown
		"r", "R",           // Refresh
		"q",                // Quit
		"?",                // Help
	}

	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("Expected help view to contain key '%s'", key)
		}
	}
}

// TestHelpModelContainsCommentsKeybindings verifies comments view keybindings are documented
func TestHelpModelContainsCommentsKeybindings(t *testing.T) {
	model := NewHelpModel()
	model.Active = true
	model.Width = 120
	model.Height = 40

	view := model.View()

	// Comments view keybindings
	expectedKeys := []string{
		"j", "k",      // Scroll
		"m",           // Toggle markdown
		"esc", "q",    // Back to list
	}

	for _, key := range expectedKeys {
		if !strings.Contains(strings.ToLower(view), strings.ToLower(key)) {
			t.Errorf("Expected help view to contain comments key '%s'", key)
		}
	}
}

// TestHelpModelViewOrganizedByContext verifies keybindings are organized by context
func TestHelpModelViewOrganizedByContext(t *testing.T) {
	model := NewHelpModel()
	model.Active = true
	model.Width = 120
	model.Height = 40

	view := model.View()

	// Should have sections for different contexts
	expectedSections := []string{
		"Main View",
		"Comments View",
	}

	viewLower := strings.ToLower(view)
	for _, section := range expectedSections {
		if !strings.Contains(viewLower, strings.ToLower(section)) {
			t.Errorf("Expected help view to have section '%s'", section)
		}
	}
}

// TestHelpModelCentered verifies help overlay is centered in viewport
func TestHelpModelCentered(t *testing.T) {
	model := NewHelpModel()
	model.Active = true
	model.Width = 100
	model.Height = 30

	view := model.View()

	// Should have vertical padding (newlines at start)
	// This tests vertical centering
	lines := strings.Split(view, "\n")
	verticalPadding := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			verticalPadding++
		} else {
			break
		}
	}

	if verticalPadding < 1 {
		t.Errorf("Expected vertical padding for centering, got %d lines", verticalPadding)
	}

	// Check that the view is roughly centered by finding a non-empty line
	// and verifying it's not at the very top or bottom
	firstContentLine := 0
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			firstContentLine = i
			break
		}
	}

	// First content should not be at very top (allows for some padding)
	if firstContentLine < 1 {
		t.Errorf("Expected help to be vertically centered, first content at line %d", firstContentLine)
	}
}

// TestHelpModelDoesNotBlockWhenInactive verifies other keys don't affect inactive help
func TestHelpModelUpdateDoesNotBlockWhenInactive(t *testing.T) {
	model := NewHelpModel()
	model.Active = false

	// Send various keys
	keys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'j'}},
		{Type: tea.KeyRunes, Runes: []rune{'k'}},
		{Type: tea.KeyEnter},
		{Type: tea.KeySpace},
	}

	for _, key := range keys {
		updated, _ := model.Update(key)
		if updated.Active {
			t.Errorf("Expected inactive help to stay inactive on key press")
		}
	}
}
