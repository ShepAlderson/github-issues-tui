package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewHelpComponent(t *testing.T) {
	hc := NewHelpComponent()
	if hc == nil {
		t.Fatal("NewHelpComponent returned nil")
	}
	if hc.showHelp {
		t.Error("New help component should not show help by default")
	}
}

func TestHelpComponent_ShowHelp(t *testing.T) {
	hc := NewHelpComponent()
	hc.ShowHelp()
	if !hc.showHelp {
		t.Error("ShowHelp should set showHelp to true")
	}
}

func TestHelpComponent_HideHelp(t *testing.T) {
	hc := NewHelpComponent()
	hc.ShowHelp()
	hc.HideHelp()
	if hc.showHelp {
		t.Error("HideHelp should set showHelp to false")
	}
}

func TestHelpComponent_Update_WindowSize(t *testing.T) {
	hc := NewHelpComponent()
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := hc.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil component")
	}
	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}
	if updated.width != 80 || updated.height != 24 {
		t.Errorf("Expected width=80, height=24, got width=%d, height=%d", updated.width, updated.height)
	}
}

func TestHelpComponent_Update_KeyMsg_HelpHidden(t *testing.T) {
	hc := NewHelpComponent()
	hc.HideHelp()

	// Test that key messages don't affect hidden help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, cmd := hc.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil component")
	}
	if cmd != nil {
		t.Error("KeyMsg should not return a command when help is hidden")
	}
	if updated.showHelp {
		t.Error("KeyMsg should not show help when component is hidden")
	}
}

func TestHelpComponent_Update_KeyMsg_HelpShown(t *testing.T) {
	hc := NewHelpComponent()
	hc.ShowHelp()

	// Test that ? does NOT hide help when shown (App handles ?)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, cmd := hc.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil component")
	}
	if cmd != nil {
		t.Error("KeyMsg should not return a command")
	}
	if !updated.showHelp {
		t.Error("? key should NOT hide help in help component (App handles ?)")
	}

	// Test that Esc hides help
	hc.ShowHelp()
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd = hc.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil component")
	}
	if cmd != nil {
		t.Error("KeyMsg should not return a command")
	}
	if updated.showHelp {
		t.Error("Esc key should hide help when shown")
	}

	// Test that Enter hides help
	hc.ShowHelp()
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd = hc.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil component")
	}
	if cmd != nil {
		t.Error("KeyMsg should not return a command")
	}
	if updated.showHelp {
		t.Error("Enter key should hide help when shown")
	}

	// Test that Space hides help
	hc.ShowHelp()
	msg = tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd = hc.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil component")
	}
	if cmd != nil {
		t.Error("KeyMsg should not return a command")
	}
	if updated.showHelp {
		t.Error("Space key should hide help when shown")
	}
}

func TestHelpComponent_View_Hidden(t *testing.T) {
	hc := NewHelpComponent()
	hc.HideHelp()

	view := hc.View()
	if view != "" {
		t.Errorf("View should return empty string when help is hidden, got: %q", view)
	}
}

func TestHelpComponent_View_Shown(t *testing.T) {
	hc := NewHelpComponent()
	hc.width = 80
	hc.height = 24
	hc.ShowHelp()

	view := hc.View()
	if view == "" {
		t.Error("View should return non-empty string when help is shown")
	}
	if len(view) < 100 {
		t.Errorf("View should return reasonably sized help content, got length: %d", len(view))
	}
}

func TestGetFooterHints(t *testing.T) {
	// Test list context
	listHints := GetFooterHints(FooterContextList)
	if listHints == "" {
		t.Error("GetFooterHints should return non-empty string for list context")
	}
	if !containsAll(listHints, []string{"help", "quit", "navigate", "open", "sort", "refresh"}) {
		t.Errorf("List hints should contain key terms, got: %s", listHints)
	}

	// Test detail context
	detailHints := GetFooterHints(FooterContextDetail)
	if detailHints == "" {
		t.Error("GetFooterHints should return non-empty string for detail context")
	}
	if !containsAll(detailHints, []string{"help", "quit", "scroll", "comments", "back", "toggle"}) {
		t.Errorf("Detail hints should contain key terms, got: %s", detailHints)
	}

	// Test comments context
	commentsHints := GetFooterHints(FooterContextComments)
	if commentsHints == "" {
		t.Error("GetFooterHints should return non-empty string for comments context")
	}
	if !containsAll(commentsHints, []string{"help", "quit", "scroll", "comments", "back", "toggle"}) {
		t.Errorf("Comments hints should contain key terms, got: %s", commentsHints)
	}

	// Test that all contexts include help and quit
	if !containsAll(listHints, []string{"help", "quit"}) {
		t.Error("All contexts should include help and quit hints")
	}
	if !containsAll(detailHints, []string{"help", "quit"}) {
		t.Error("All contexts should include help and quit hints")
	}
	if !containsAll(commentsHints, []string{"help", "quit"}) {
		t.Error("All contexts should include help and quit hints")
	}
}

func containsAll(str string, substrings []string) bool {
	for _, substr := range substrings {
		if !contains(str, substr) {
			return false
		}
	}
	return true
}

func contains(str, substr string) bool {
	// Simple check for test purposes
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}