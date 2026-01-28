package config

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestThemeModelNew(t *testing.T) {
	model := NewThemeModel("default")

	if model.currentTheme != "default" {
		t.Errorf("expected currentTheme to be 'default', got %s", model.currentTheme)
	}

	if len(model.availableThemes) != 6 {
		t.Errorf("expected 6 themes, got %d", len(model.availableThemes))
	}

	if model.selected != 0 {
		t.Errorf("expected selected to be 0, got %d", model.selected)
	}

	if model.width != 80 {
		t.Errorf("expected width to be 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("expected height to be 24, got %d", model.height)
	}

	if model.quitting {
		t.Error("expected quitting to be false")
	}

	if model.saved {
		t.Error("expected saved to be false")
	}
}

func TestThemeModelUpdateNavigation(t *testing.T) {
	model := NewThemeModel("default")

	// Test moving down
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(ThemeModel)
	if model.selected != 1 {
		t.Errorf("expected selected to be 1 after down, got %d", model.selected)
	}

	// Test moving up
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = newModel.(ThemeModel)
	if model.selected != 0 {
		t.Errorf("expected selected to be 0 after up, got %d", model.selected)
	}

	// Test wrapping at top
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = newModel.(ThemeModel)
	if model.selected != len(model.availableThemes)-1 {
		t.Errorf("expected selected to wrap to last item, got %d", model.selected)
	}

	// Test wrapping at bottom
	model.selected = len(model.availableThemes) - 1
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(ThemeModel)
	if model.selected != 0 {
		t.Errorf("expected selected to wrap to 0, got %d", model.selected)
	}
}

func TestThemeModelUpdateVimKeys(t *testing.T) {
	model := NewThemeModel("default")

	// Test 'j' for down
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = newModel.(ThemeModel)
	if model.selected != 1 {
		t.Errorf("expected selected to be 1 after j, got %d", model.selected)
	}

	// Test 'k' for up
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = newModel.(ThemeModel)
	if model.selected != 0 {
		t.Errorf("expected selected to be 0 after k, got %d", model.selected)
	}
}


func TestThemeModelUpdateEnter(t *testing.T) {
	model := NewThemeModel("default")

	// Select a different theme (e.g., dracula which is at index 1)
	model.selected = 1
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = newModel.(ThemeModel)

	if !model.quitting {
		t.Error("expected quitting to be true after Enter")
	}

	if !model.saved {
		t.Error("expected saved to be true after Enter")
	}

	if model.GetSelectedTheme() != "dracula" {
		t.Errorf("expected theme to be 'dracula', got %s", model.GetSelectedTheme())
	}

	// Verify the command is tea.Quit
	if cmd == nil {
		t.Error("expected command to be returned")
	}
}

func TestThemeModelUpdateQKey(t *testing.T) {
	model := NewThemeModel("default")

	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = newModel.(ThemeModel)

	if !model.quitting {
		t.Error("expected quitting to be true after q")
	}

	if model.saved {
		t.Error("expected saved to be false after q (cancel)")
	}

	// Verify the command is tea.Quit
	if cmd == nil {
		t.Error("expected command to be returned")
	}
}

func TestThemeModelUpdateCtrlC(t *testing.T) {
	model := NewThemeModel("default")

	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model = newModel.(ThemeModel)

	if !model.quitting {
		t.Error("expected quitting to be true after Ctrl+C")
	}

	if model.saved {
		t.Error("expected saved to be false after Ctrl+C (cancel)")
	}

	// Verify the command is tea.Quit
	if cmd == nil {
		t.Error("expected command to be returned")
	}
}

func TestThemeModelUpdateWindowSize(t *testing.T) {
	model := NewThemeModel("default")

	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model = newModel.(ThemeModel)

	if model.width != 100 {
		t.Errorf("expected width to be 100, got %d", model.width)
	}

	if model.height != 30 {
		t.Errorf("expected height to be 30, got %d", model.height)
	}
}

func TestThemeModelGetSelectedTheme(t *testing.T) {
	model := NewThemeModel("default")

	// Test getting theme at different positions
	tests := []struct {
		selected int
		expected string
	}{
		{0, "default"},
		{1, "dracula"},
		{2, "gruvbox"},
		{3, "nord"},
		{4, "solarized-dark"},
		{5, "solarized-light"},
	}

	for _, tt := range tests {
		model.selected = tt.selected
		got := model.GetSelectedTheme()
		if got != tt.expected {
			t.Errorf("GetSelectedTheme() with selected=%d: got %s, want %s", tt.selected, got, tt.expected)
		}
	}
}

func TestThemeModelIsSaved(t *testing.T) {
	model := NewThemeModel("default")

	if model.IsSaved() {
		t.Error("expected IsSaved to be false initially")
	}

	model.saved = true

	if !model.IsSaved() {
		t.Error("expected IsSaved to be true after setting saved")
	}
}

func TestThemeModelViewContainsContent(t *testing.T) {
	model := NewThemeModel("default")
	model.width = 80
	model.height = 24

	view := model.View()

	// Check that view contains expected elements
	expectedStrings := []string{
		"Theme",        // Title
		"default",      // First theme
		"dracula",      // Second theme
		"gruvbox",
		"nord",
		"solarized",
	}

	for _, s := range expectedStrings {
		if !contains(view, s) {
			t.Errorf("View does not contain expected string: %s", s)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(findSubstring(s, substr) >= 0)
}

func findSubstring(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestSaveThemeToConfig(t *testing.T) {
	// Test with invalid theme name
	err := SaveThemeToConfig("invalid-theme")
	if err == nil {
		t.Error("expected error for invalid theme name")
	}

	// Note: We can't easily test valid theme saving without
	// a config file. The main test would be integration testing.
}
