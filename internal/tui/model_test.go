package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/github-issues-tui/internal/sync"
)

func TestModel_Navigation(t *testing.T) {
	// Create test issues
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Number: 2, Title: "Second", State: "open", Author: "user2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Number: 3, Title: "Third", State: "open", Author: "user3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title", "author"})

	// Test initial state
	if model.cursor != 0 {
		t.Errorf("Initial cursor = %d, want 0", model.cursor)
	}

	// Test moving down with 'j'
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updatedModel.(Model)
	if m.cursor != 1 {
		t.Errorf("After 'j', cursor = %d, want 1", m.cursor)
	}

	// Test moving down again
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(Model)
	if m.cursor != 2 {
		t.Errorf("After second 'j', cursor = %d, want 2", m.cursor)
	}

	// Test boundary: moving down at last item should not change cursor
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(Model)
	if m.cursor != 2 {
		t.Errorf("After 'j' at end, cursor = %d, want 2", m.cursor)
	}

	// Test moving up with 'k'
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(Model)
	if m.cursor != 1 {
		t.Errorf("After 'k', cursor = %d, want 1", m.cursor)
	}

	// Move to top
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("After second 'k', cursor = %d, want 0", m.cursor)
	}

	// Test boundary: moving up at first item should not change cursor
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("After 'k' at top, cursor = %d, want 0", m.cursor)
	}
}

func TestModel_ArrowKeyNavigation(t *testing.T) {
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Number: 2, Title: "Second", State: "open", Author: "user2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"})

	// Test down arrow
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updatedModel.(Model)
	if m.cursor != 1 {
		t.Errorf("After KeyDown, cursor = %d, want 1", m.cursor)
	}

	// Test up arrow
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("After KeyUp, cursor = %d, want 0", m.cursor)
	}
}

func TestModel_Quit(t *testing.T) {
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"})

	// Test 'q' quits
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("Expected quit command after 'q', got nil")
	}

	// Test Ctrl+C quits
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("Expected quit command after Ctrl+C, got nil")
	}
}

func TestModel_EmptyIssues(t *testing.T) {
	// Test with no issues
	model := NewModel([]*sync.Issue{}, []string{"number", "title"})

	if model.cursor != 0 {
		t.Errorf("Empty model cursor = %d, want 0", model.cursor)
	}

	// Navigation should not crash
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("After 'j' on empty, cursor = %d, want 0", m.cursor)
	}
}

func TestModel_SelectedIssue(t *testing.T) {
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Number: 2, Title: "Second", State: "open", Author: "user2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"})

	// Get initially selected issue
	selected := model.SelectedIssue()
	if selected == nil {
		t.Fatal("SelectedIssue() returned nil")
	}
	if selected.Number != 1 {
		t.Errorf("Initially selected issue number = %d, want 1", selected.Number)
	}

	// Move cursor and get selected issue
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updatedModel.(Model)
	selected = m.SelectedIssue()
	if selected == nil {
		t.Fatal("SelectedIssue() returned nil after navigation")
	}
	if selected.Number != 2 {
		t.Errorf("After moving down, selected issue number = %d, want 2", selected.Number)
	}
}

func TestModel_SelectedIssue_Empty(t *testing.T) {
	model := NewModel([]*sync.Issue{}, []string{"number", "title"})

	selected := model.SelectedIssue()
	if selected != nil {
		t.Errorf("SelectedIssue() on empty model = %v, want nil", selected)
	}
}
