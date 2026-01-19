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

	model := NewModel(issues, []string{"number", "title", "author"}, "updated", false)

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

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
	model := NewModel([]*sync.Issue{}, []string{"number", "title"}, "updated", false)

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
	now := time.Now()
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour)},
		{Number: 2, Title: "Second", State: "open", Author: "user2", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

	// Get initially selected issue (should be #2 since it's most recently updated)
	selected := model.SelectedIssue()
	if selected == nil {
		t.Fatal("SelectedIssue() returned nil")
	}
	if selected.Number != 2 {
		t.Errorf("Initially selected issue number = %d, want 2 (most recently updated)", selected.Number)
	}

	// Move cursor and get selected issue
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m := updatedModel.(Model)
	selected = m.SelectedIssue()
	if selected == nil {
		t.Fatal("SelectedIssue() returned nil after navigation")
	}
	if selected.Number != 1 {
		t.Errorf("After moving down, selected issue number = %d, want 1", selected.Number)
	}
}

func TestModel_SelectedIssue_Empty(t *testing.T) {
	model := NewModel([]*sync.Issue{}, []string{"number", "title"}, "updated", false)

	selected := model.SelectedIssue()
	if selected != nil {
		t.Errorf("SelectedIssue() on empty model = %v, want nil", selected)
	}
}

func TestModel_SortKeyCycling(t *testing.T) {
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

	// Initial sort should be "updated"
	if model.sortBy != "updated" {
		t.Errorf("Initial sortBy = %s, want updated", model.sortBy)
	}
	if model.sortAscending != false {
		t.Errorf("Initial sortAscending = %v, want false", model.sortAscending)
	}

	// Press 's' to cycle to "created"
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m := updatedModel.(Model)
	if m.sortBy != "created" {
		t.Errorf("After first 's', sortBy = %s, want created", m.sortBy)
	}

	// Press 's' again to cycle to "number"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedModel.(Model)
	if m.sortBy != "number" {
		t.Errorf("After second 's', sortBy = %s, want number", m.sortBy)
	}

	// Press 's' again to cycle to "comments"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedModel.(Model)
	if m.sortBy != "comments" {
		t.Errorf("After third 's', sortBy = %s, want comments", m.sortBy)
	}

	// Press 's' again to cycle back to "updated"
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedModel.(Model)
	if m.sortBy != "updated" {
		t.Errorf("After fourth 's', sortBy = %s, want updated", m.sortBy)
	}
}

func TestModel_SortOrderReversal(t *testing.T) {
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

	// Initial should be descending (false)
	if model.sortAscending {
		t.Errorf("Initial sortAscending = %v, want false", model.sortAscending)
	}

	// Press 'S' to reverse to ascending
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m := updatedModel.(Model)
	if !m.sortAscending {
		t.Errorf("After 'S', sortAscending = %v, want true", m.sortAscending)
	}

	// Press 'S' again to reverse back to descending
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m = updatedModel.(Model)
	if m.sortAscending {
		t.Errorf("After second 'S', sortAscending = %v, want false", m.sortAscending)
	}
}

func TestModel_IssueSorting(t *testing.T) {
	now := time.Now()
	issues := []*sync.Issue{
		{Number: 3, Title: "Third", State: "open", Author: "user3", CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour), CommentCount: 5},
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: now.Add(-5 * time.Hour), UpdatedAt: now.Add(-3 * time.Hour), CommentCount: 10},
		{Number: 2, Title: "Second", State: "open", Author: "user2", CreatedAt: now.Add(-4 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), CommentCount: 2},
	}

	tests := []struct {
		name          string
		sortBy        string
		sortAscending bool
		wantOrder     []int // Expected issue numbers in order
	}{
		{
			name:          "sort by updated desc",
			sortBy:        "updated",
			sortAscending: false,
			wantOrder:     []int{3, 2, 1}, // Most recently updated first
		},
		{
			name:          "sort by updated asc",
			sortBy:        "updated",
			sortAscending: true,
			wantOrder:     []int{1, 2, 3}, // Least recently updated first
		},
		{
			name:          "sort by created desc",
			sortBy:        "created",
			sortAscending: false,
			wantOrder:     []int{3, 2, 1}, // Most recently created first
		},
		{
			name:          "sort by created asc",
			sortBy:        "created",
			sortAscending: true,
			wantOrder:     []int{1, 2, 3}, // Least recently created first
		},
		{
			name:          "sort by number desc",
			sortBy:        "number",
			sortAscending: false,
			wantOrder:     []int{3, 2, 1}, // Highest number first
		},
		{
			name:          "sort by number asc",
			sortBy:        "number",
			sortAscending: true,
			wantOrder:     []int{1, 2, 3}, // Lowest number first
		},
		{
			name:          "sort by comments desc",
			sortBy:        "comments",
			sortAscending: false,
			wantOrder:     []int{1, 3, 2}, // Most comments first (10, 5, 2)
		},
		{
			name:          "sort by comments asc",
			sortBy:        "comments",
			sortAscending: true,
			wantOrder:     []int{2, 3, 1}, // Least comments first (2, 5, 10)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create model and set sort parameters
			model := NewModel(issues, []string{"number", "title"}, "updated", false)
			model.sortBy = tt.sortBy
			model.sortAscending = tt.sortAscending

			// Manually trigger sort
			model.sortIssues()

			// Check order
			for i, wantNumber := range tt.wantOrder {
				if model.issues[i].Number != wantNumber {
					t.Errorf("After sorting, issues[%d].Number = %d, want %d", i, model.issues[i].Number, wantNumber)
				}
			}
		})
	}
}

func TestModel_DetailPanelWithDimensions(t *testing.T) {
	// Test that detail panel is rendered when window size is set
	now := time.Now()
	issues := []*sync.Issue{
		{
			Number:       123,
			Title:        "Test Issue",
			Body:         "This is a test issue body",
			State:        "open",
			Author:       "testuser",
			CreatedAt:    now.Add(-24 * time.Hour),
			UpdatedAt:    now.Add(-1 * time.Hour),
			CommentCount: 5,
			Labels:       []string{"bug", "enhancement"},
			Assignees:    []string{"user1", "user2"},
		},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	m := updatedModel.(Model)

	if m.width != 120 {
		t.Errorf("Width = %d, want 120", m.width)
	}
	if m.height != 30 {
		t.Errorf("Height = %d, want 30", m.height)
	}

	// Verify View() returns non-empty string (actual split-pane rendering tested visually)
	view := m.View()
	if view == "" {
		t.Error("View() returned empty string after setting dimensions")
	}
}

func TestModel_MarkdownToggle(t *testing.T) {
	// Test toggling between raw and rendered markdown with 'm' key
	issues := []*sync.Issue{
		{
			Number:    1,
			Title:     "Test",
			Body:      "# Heading\n\n**Bold text**",
			State:     "open",
			Author:    "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false)

	// Initial state should be rendered markdown (showRawMarkdown = false)
	if model.showRawMarkdown {
		t.Errorf("Initial showRawMarkdown = %v, want false", model.showRawMarkdown)
	}

	// Press 'm' to toggle to raw
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m := updatedModel.(Model)
	if !m.showRawMarkdown {
		t.Errorf("After 'm', showRawMarkdown = %v, want true", m.showRawMarkdown)
	}

	// Press 'm' again to toggle back to rendered
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedModel.(Model)
	if m.showRawMarkdown {
		t.Errorf("After second 'm', showRawMarkdown = %v, want false", m.showRawMarkdown)
	}
}

func TestModel_DetailPanelScrolling(t *testing.T) {
	// Test scrolling in detail panel with PageDown/PageUp
	longBody := ""
	for i := 0; i < 100; i++ {
		longBody += "Line " + string(rune(i)) + "\n"
	}

	issues := []*sync.Issue{
		{
			Number:    1,
			Title:     "Test",
			Body:      longBody,
			State:     "open",
			Author:    "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false)
	model.width = 120
	model.height = 30

	// Initial scroll offset should be 0
	if model.detailScrollOffset != 0 {
		t.Errorf("Initial detailScrollOffset = %d, want 0", model.detailScrollOffset)
	}

	// Press PageDown to scroll down
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m := updatedModel.(Model)
	if m.detailScrollOffset <= 0 {
		t.Errorf("After PageDown, detailScrollOffset = %d, want > 0", m.detailScrollOffset)
	}

	scrollAfterPageDown := m.detailScrollOffset

	// Press PageUp to scroll back up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = updatedModel.(Model)
	if m.detailScrollOffset >= scrollAfterPageDown {
		t.Errorf("After PageUp, detailScrollOffset = %d, want < %d", m.detailScrollOffset, scrollAfterPageDown)
	}

	// Scroll offset should not go below 0
	for i := 0; i < 10; i++ {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
		m = updatedModel.(Model)
	}
	if m.detailScrollOffset < 0 {
		t.Errorf("After multiple PageUp, detailScrollOffset = %d, want >= 0", m.detailScrollOffset)
	}
}
