package help

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	model := NewModel()

	if model.showing != false {
		t.Errorf("NewModel() showing = %v, want false", model.showing)
	}

	if model.width != 80 {
		t.Errorf("NewModel() width = %d, want 80", model.width)
	}

	if model.height != 24 {
		t.Errorf("NewModel() height = %d, want 24", model.height)
	}
}

func TestToggleHelp(t *testing.T) {
	model := NewModel()

	// Initially not showing
	if model.IsShowing() {
		t.Error("Expected help to not be showing initially")
	}

	// Toggle on
	model.ToggleHelp()
	if !model.IsShowing() {
		t.Error("Expected help to be showing after toggle")
	}

	// Toggle off
	model.ToggleHelp()
	if model.IsShowing() {
		t.Error("Expected help to not be showing after second toggle")
	}
}

func TestShowHelp(t *testing.T) {
	model := NewModel()

	model.ShowHelp()
	if !model.IsShowing() {
		t.Error("Expected help to be showing after ShowHelp()")
	}
}

func TestHideHelp(t *testing.T) {
	model := NewModel()

	model.ShowHelp()
	model.HideHelp()
	if model.IsShowing() {
		t.Error("Expected help to not be showing after HideHelp()")
	}
}

func TestSetDimensions(t *testing.T) {
	model := NewModel()

	model.SetDimensions(100, 40)

	if model.width != 100 {
		t.Errorf("SetDimensions() width = %d, want 100", model.width)
	}

	if model.height != 40 {
		t.Errorf("SetDimensions() height = %d, want 40", model.height)
	}
}

func TestViewWhenNotShowing(t *testing.T) {
	model := NewModel()

	view := model.View()

	if view != "" {
		t.Errorf("View() when not showing = %q, want empty string", view)
	}
}

func TestViewWhenShowing(t *testing.T) {
	model := NewModel()
	model.ShowHelp()

	view := model.View()

	// Should contain title
	if !strings.Contains(view, "Keybindings") {
		t.Error("View() should contain 'Keybindings' title")
	}

	// Should contain navigation section
	if !strings.Contains(view, "Navigation") {
		t.Error("View() should contain 'Navigation' section")
	}

	// Should contain list view section
	if !strings.Contains(view, "List View") {
		t.Error("View() should contain 'List View' section")
	}

	// Should contain detail view section
	if !strings.Contains(view, "Detail View") {
		t.Error("View() should contain 'Detail View' section")
	}

	// Should contain comments view section
	if !strings.Contains(view, "Comments View") {
		t.Error("View() should contain 'Comments View' section")
	}

	// Should contain dismiss hint (without ANSI codes)
	if !strings.Contains(view, "Press") || !strings.Contains(view, "to dismiss") {
		t.Error("View() should contain dismiss hint")
	}
}

func TestViewContainsKeybindings(t *testing.T) {
	model := NewModel()
	model.ShowHelp()

	view := model.View()

	// Check for specific keybindings
	expectedKeys := []string{
		"j, k",
		"↑, ↓",
		"Enter",
		"s",
		"S",
		"m",
		"r",
		"q",
		"?",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(view, key) {
			t.Errorf("View() should contain keybinding %q", key)
		}
	}
}

func TestHelpModelAsTeaModel(t *testing.T) {
	// Test that HelpModel implements tea.Model interface
	var _ tea.Model = NewModel()

	model := NewModel()

	// Test Init
	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}

	// Test Update with WindowSizeMsg
	newModel, cmd := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m := newModel.(Model)

	if m.width != 100 {
		t.Errorf("Update(WindowSizeMsg) width = %d, want 100", m.width)
	}

	if m.height != 40 {
		t.Errorf("Update(WindowSizeMsg) height = %d, want 40", m.height)
	}

	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
}

func TestRenderKeybinding(t *testing.T) {
	key := renderKeybinding("j, k", "Move up/down")

	if !strings.Contains(key, "j, k") {
		t.Error("renderKeybinding() should contain key")
	}

	if !strings.Contains(key, "Move up/down") {
		t.Error("renderKeybinding() should contain description")
	}
}

func TestRenderSection(t *testing.T) {
	items := []string{"item1", "item2"}
	section := renderSection("Test Section", items)

	if !strings.Contains(section, "Test Section") {
		t.Error("renderSection() should contain section title")
	}

	if !strings.Contains(section, "item1") {
		t.Error("renderSection() should contain item1")
	}

	if !strings.Contains(section, "item2") {
		t.Error("renderSection() should contain item2")
	}
}
