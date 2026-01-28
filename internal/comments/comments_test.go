package comments

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/ghissues/internal/database"
)

func TestNewModel(t *testing.T) {
	dbPath := "/tmp/test.db"
	repo := "owner/repo"
	issueNumber := 42

	model := NewModel(dbPath, repo, issueNumber, "Test Issue Title")

	if model.dbPath != dbPath {
		t.Errorf("expected dbPath to be '/tmp/test.db', got %q", model.dbPath)
	}

	if model.repo != repo {
		t.Errorf("expected repo to be 'owner/repo', got %q", model.repo)
	}

	if model.issueNumber != issueNumber {
		t.Errorf("expected issueNumber to be 42, got %d", model.issueNumber)
	}

	if model.issueTitle != "Test Issue Title" {
		t.Errorf("expected title 'Test Issue Title', got %q", model.issueTitle)
	}

	if model.renderedMode != true {
		t.Error("expected rendered mode to be true by default")
	}

	if model.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0, got %d", model.scrollOffset)
	}
}

func TestModel_ToggleRenderedMode(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")

	// Initially in rendered mode
	if !model.renderedMode {
		t.Error("expected rendered mode initially")
	}

	// Toggle to raw mode
	model.ToggleRenderedMode()
	if model.renderedMode {
		t.Error("expected raw mode after toggle")
	}

	// Toggle back to rendered mode
	model.ToggleRenderedMode()
	if !model.renderedMode {
		t.Error("expected rendered mode after second toggle")
	}
}

func TestModel_SetDimensions(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.SetDimensions(100, 30)

	if model.width != 100 {
		t.Errorf("expected width 100, got %d", model.width)
	}

	if model.height != 30 {
		t.Errorf("expected height 30, got %d", model.height)
	}
}

func TestModel_QuitWithQ(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.comments = []database.Comment{
		{ID: 1, IssueNumber: 1, Body: "Test comment", Author: "user"},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)

	// Should return quit command
	if cmd == nil {
		t.Error("expected quit command for 'q' key")
	}
}

func TestModel_QuitWithEsc(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.comments = []database.Comment{
		{ID: 1, IssueNumber: 1, Body: "Test comment", Author: "user"},
	}

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := model.Update(msg)

	// Should return quit command
	if cmd == nil {
		t.Error("expected quit command for Escape key")
	}
}

func TestModel_QuitWithCtrlC(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	// Should return quit command
	if cmd == nil {
		t.Error("expected quit command for Ctrl+C")
	}
}

func TestModel_MKeyTogglesRenderedMode(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.comments = []database.Comment{
		{ID: 1, IssueNumber: 1, Body: "Test **markdown** comment", Author: "user"},
	}

	// Initially in rendered mode
	if !model.renderedMode {
		t.Error("expected rendered mode initially")
	}

	// Press 'm' to toggle
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.renderedMode {
		t.Error("expected raw mode after 'm' key")
	}

	// Press 'm' again to toggle back
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if !m.renderedMode {
		t.Error("expected rendered mode after second 'm' key")
	}
}

func TestModel_ScrollDown(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.SetDimensions(80, 5) // Small height to force scrolling
	model.comments = []database.Comment{
		{ID: 1, Author: "user1", Body: "Comment 1"},
		{ID: 2, Author: "user2", Body: "Comment 2"},
		{ID: 3, Author: "user3", Body: "Comment 3"},
		{ID: 4, Author: "user4", Body: "Comment 4"},
		{ID: 5, Author: "user5", Body: "Comment 5"},
	}

	// Scroll down with 'j' and arrow down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.scrollOffset != 1 {
		t.Errorf("expected scrollOffset to be 1, got %d", m.scrollOffset)
	}

	msg = tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.scrollOffset != 2 {
		t.Errorf("expected scrollOffset to be 2, got %d", m.scrollOffset)
	}
}

func TestModel_ScrollUp(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.SetDimensions(80, 5)
	model.comments = []database.Comment{
		{ID: 1, Author: "user1", Body: "Comment 1"},
		{ID: 2, Author: "user2", Body: "Comment 2"},
	}
	model.scrollOffset = 2

	// Scroll up with 'k' and arrow up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.scrollOffset != 1 {
		t.Errorf("expected scrollOffset to be 1, got %d", m.scrollOffset)
	}

	msg = tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0, got %d", m.scrollOffset)
	}
}

func TestModel_ScrollDoesNotGoNegative(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.SetDimensions(80, 10)
	model.comments = []database.Comment{
		{ID: 1, Author: "user1", Body: "Comment 1"},
		{ID: 2, Author: "user2", Body: "Comment 2"},
	}
	model.scrollOffset = 1

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to be 0, got %d", m.scrollOffset)
	}

	// Try to scroll up again
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to stay at 0, got %d", m.scrollOffset)
	}
}

func TestModel_ViewRendersHeader(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 42, "Test Issue Title")
	model.SetDimensions(80, 10)
	model.comments = []database.Comment{
		{ID: 1, Author: "user1", Body: "Test comment"},
	}

	view := model.View()

	// Check header contains issue number and title
	if !contains(view, "#42") {
		t.Error("expected view to contain issue number '#42'")
	}
	if !contains(view, "Test Issue Title") {
		t.Error("expected view to contain issue title")
	}
}

func TestModel_ViewRendersComments(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.SetDimensions(80, 20)
	model.comments = []database.Comment{
		{ID: 1, Author: "alice", Body: "First comment", CreatedAt: "2024-01-15T10:00:00Z"},
		{ID: 2, Author: "bob", Body: "Second comment", CreatedAt: "2024-01-16T11:00:00Z"},
	}

	view := model.View()

	// Check comments appear in view
	if !contains(view, "alice") {
		t.Error("expected view to contain author 'alice'")
	}
	if !contains(view, "bob") {
		t.Error("expected view to contain author 'bob'")
	}
	if !contains(view, "First comment") {
		t.Error("expected view to contain 'First comment'")
	}
	if !contains(view, "Second comment") {
		t.Error("expected view to contain 'Second comment'")
	}
}

func TestModel_ViewEmptyComments(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 42, "Test Issue")
	model.SetDimensions(80, 10)
	model.comments = []database.Comment{}

	view := model.View()

	// Should show "No comments" message
	if !contains(view, "No comments") {
		t.Error("expected view to contain 'No comments' message")
	}
}

func TestModel_ViewShowsScrollHint(t *testing.T) {
	model := NewModel("/tmp/test.db", "owner/repo", 1, "Test Issue")
	model.SetDimensions(80, 3) // Small height
	model.comments = []database.Comment{
		{ID: 1, Author: "user1", Body: "Comment 1"},
		{ID: 2, Author: "user2", Body: "Comment 2"},
		{ID: 3, Author: "user3", Body: "Comment 3"},
	}

	view := model.View()

	// Should show help text for navigation
	if !contains(view, "j/k") {
		t.Error("expected view to contain scroll navigation hint")
	}
	if !contains(view, "q") {
		t.Error("expected view to contain quit hint")
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "formats RFC3339 date",
			input:    "2024-01-15T10:30:00Z",
			expected: "2024-01-15",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "date unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
