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

	model := NewModel(issues, []string{"number", "title", "author"}, "updated", false, nil, time.Time{})

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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
	model := NewModel([]*sync.Issue{}, []string{"number", "title"}, "updated", false, nil, time.Time{})

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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
	model := NewModel([]*sync.Issue{}, []string{"number", "title"}, "updated", false, nil, time.Time{})

	selected := model.SelectedIssue()
	if selected != nil {
		t.Errorf("SelectedIssue() on empty model = %v, want nil", selected)
	}
}

func TestModel_SortKeyCycling(t *testing.T) {
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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
			model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})
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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

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

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})
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

func TestModel_CommentsViewNavigation(t *testing.T) {
	// Create temporary database with test data
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	store, err := sync.NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store test issue
	issue := &sync.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Store test comments
	comments := []*sync.Comment{
		{ID: 1, IssueNumber: 1, Body: "First comment", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: 2, IssueNumber: 1, Body: "Second comment", Author: "user2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, comment := range comments {
		if err := store.StoreComment(comment); err != nil {
			t.Fatalf("StoreComment() error = %v", err)
		}
	}

	// Create model with store
	issues := []*sync.Issue{issue}
	model := NewModel(issues, []string{"number", "title"}, "updated", false, store, time.Time{})

	// Initial view mode should be list
	if model.viewMode != viewModeList {
		t.Errorf("Initial viewMode = %d, want %d (viewModeList)", model.viewMode, viewModeList)
	}

	// Press Enter to switch to comments view
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)
	if m.viewMode != viewModeComments {
		t.Errorf("After Enter, viewMode = %d, want %d (viewModeComments)", m.viewMode, viewModeComments)
	}
	if m.commentsScrollOffset != 0 {
		t.Errorf("After entering comments view, commentsScrollOffset = %d, want 0", m.commentsScrollOffset)
	}
	if len(m.currentComments) != 2 {
		t.Errorf("After entering comments view, len(currentComments) = %d, want 2", len(m.currentComments))
	}

	// Press Esc to return to list view
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)
	if m.viewMode != viewModeList {
		t.Errorf("After Esc, viewMode = %d, want %d (viewModeList)", m.viewMode, viewModeList)
	}

	// Enter comments view again
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)

	// Press 'q' to return to list view
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updatedModel.(Model)
	if m.viewMode != viewModeList {
		t.Errorf("After 'q' in comments view, viewMode = %d, want %d (viewModeList)", m.viewMode, viewModeList)
	}
}

func TestModel_CommentsViewScrolling(t *testing.T) {
	// Create temporary database with test data
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	store, err := sync.NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store test issue
	issue := &sync.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Store test comment
	comment := &sync.Comment{
		ID:          1,
		IssueNumber: 1,
		Body:        "Test comment",
		Author:      "user1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.StoreComment(comment); err != nil {
		t.Fatalf("StoreComment() error = %v", err)
	}

	// Create model with store
	issues := []*sync.Issue{issue}
	model := NewModel(issues, []string{"number", "title"}, "updated", false, store, time.Time{})
	model.width = 120
	model.height = 30

	// Enter comments view
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	// Initial scroll offset should be 0
	if m.commentsScrollOffset != 0 {
		t.Errorf("Initial commentsScrollOffset = %d, want 0", m.commentsScrollOffset)
	}

	// Press PageDown to scroll down
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updatedModel.(Model)
	if m.commentsScrollOffset <= 0 {
		t.Errorf("After PageDown, commentsScrollOffset = %d, want > 0", m.commentsScrollOffset)
	}

	scrollAfterPageDown := m.commentsScrollOffset

	// Press PageUp to scroll back up
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = updatedModel.(Model)
	if m.commentsScrollOffset >= scrollAfterPageDown {
		t.Errorf("After PageUp, commentsScrollOffset = %d, want < %d", m.commentsScrollOffset, scrollAfterPageDown)
	}

	// Scroll offset should not go below 0
	for i := 0; i < 10; i++ {
		updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
		m = updatedModel.(Model)
	}
	if m.commentsScrollOffset < 0 {
		t.Errorf("After multiple PageUp, commentsScrollOffset = %d, want >= 0", m.commentsScrollOffset)
	}
}

func TestModel_CommentsViewMarkdownToggle(t *testing.T) {
	// Test markdown toggle in comments view
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Enter comments view
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	// Initial state should be rendered markdown
	if m.showRawMarkdown {
		t.Errorf("Initial showRawMarkdown = %v, want false", m.showRawMarkdown)
	}

	// Press 'm' to toggle to raw
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedModel.(Model)
	if !m.showRawMarkdown {
		t.Errorf("After 'm' in comments view, showRawMarkdown = %v, want true", m.showRawMarkdown)
	}

	// Press 'm' again to toggle back
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updatedModel.(Model)
	if m.showRawMarkdown {
		t.Errorf("After second 'm' in comments view, showRawMarkdown = %v, want false", m.showRawMarkdown)
	}
}

func TestModel_CommentsViewWithEmptyIssueList(t *testing.T) {
	// Test that Enter doesn't crash with empty issue list
	model := NewModel([]*sync.Issue{}, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Try to enter comments view
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	// Should remain in list view
	if m.viewMode != viewModeList {
		t.Errorf("After Enter with empty issues, viewMode = %d, want %d (viewModeList)", m.viewMode, viewModeList)
	}
}

func TestModel_StatusBarError(t *testing.T) {
	// Test that status bar errors are displayed correctly
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Should have no error initially
	if model.statusError != "" {
		t.Errorf("Initial statusError = %q, want empty string", model.statusError)
	}

	// Send error message
	updatedModel, _ := model.Update(StatusErrorMsg{Err: "Network timeout"})
	m := updatedModel.(Model)

	if m.statusError != "Network timeout" {
		t.Errorf("After StatusErrorMsg, statusError = %q, want \"Network timeout\"", m.statusError)
	}

	// Status bar should include error
	status := m.renderStatus()
	if status == "" {
		t.Error("renderStatus() returned empty string with status error")
	}

	// Clear error message
	updatedModel, _ = m.Update(ClearStatusErrorMsg{})
	m = updatedModel.(Model)

	if m.statusError != "" {
		t.Errorf("After ClearStatusErrorMsg, statusError = %q, want empty string", m.statusError)
	}
}

func TestModel_ModalError(t *testing.T) {
	// Test that modal errors require acknowledgment
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Should have no error initially
	if model.modalError != "" {
		t.Errorf("Initial modalError = %q, want empty string", model.modalError)
	}

	// Send modal error message
	updatedModel, _ := model.Update(ModalErrorMsg{Err: "Database corruption detected"})
	m := updatedModel.(Model)

	if m.modalError != "Database corruption detected" {
		t.Errorf("After ModalErrorMsg, modalError = %q, want \"Database corruption detected\"", m.modalError)
	}

	// Navigation should not work when modal is active
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(Model)
	if m.cursor != 0 {
		t.Errorf("With modal error, cursor = %d after 'j', want 0 (navigation blocked)", m.cursor)
	}

	// Press Enter to acknowledge error
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updatedModel.(Model)

	if m.modalError != "" {
		t.Errorf("After Enter to acknowledge, modalError = %q, want empty string", m.modalError)
	}

	// Navigation should work again after acknowledgment
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(Model)
	if len(m.issues) > 1 && m.cursor == 0 {
		t.Error("After acknowledging modal error, navigation still blocked")
	}
}

func TestModel_LoadCommentsError(t *testing.T) {
	// Test error handling when loading comments fails
	// Create a store that will fail to load comments (closed database)
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	store, err := sync.NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}

	// Store test issue
	issue := &sync.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Close the store to cause errors
	store.Close()

	// Create model with closed store
	issues := []*sync.Issue{issue}
	model := NewModel(issues, []string{"number", "title"}, "updated", false, store, time.Time{})

	// Try to enter comments view (should fail to load comments)
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	// Should have status error set
	if m.statusError == "" {
		t.Error("Expected statusError after failing to load comments, got empty string")
	}

	// Should remain in list view
	if m.viewMode != viewModeList {
		t.Errorf("After failed comment load, viewMode = %d, want %d (viewModeList)", m.viewMode, viewModeList)
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		t        time.Time
		expected string
	}{
		{
			name:     "zero time",
			t:        time.Time{},
			expected: "never",
		},
		{
			name:     "just now (5 seconds ago)",
			t:        now.Add(-5 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			t:        now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			t:        now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			t:        now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			t:        now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			t:        now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "5 days ago",
			t:        now.Add(-5 * 24 * time.Hour),
			expected: "5 days ago",
		},
		{
			name:     "1 week ago",
			t:        now.Add(-7 * 24 * time.Hour),
			expected: "1 week ago",
		},
		{
			name:     "2 weeks ago",
			t:        now.Add(-14 * 24 * time.Hour),
			expected: "2 weeks ago",
		},
		{
			name:     "35 days ago (shows weeks)",
			t:        now.Add(-35 * 24 * time.Hour),
			expected: "5 weeks ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(tt.t)
			if result != tt.expected {
				t.Errorf("formatRelativeTime(%v) = %q, want %q", tt.t, result, tt.expected)
			}
		})
	}
}

func TestModel_LastSyncedIndicator(t *testing.T) {
	// Create test issues
	issues := []*sync.Issue{
		{Number: 1, Title: "First", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	// Test with recent sync (5 minutes ago)
	lastSync := time.Now().Add(-5 * time.Minute)
	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, lastSync)

	status := model.renderStatus()

	// Status should include "Last synced: 5 minutes ago"
	if !contains(status, "Last synced:") {
		t.Errorf("Status bar should contain 'Last synced:', got: %s", status)
	}
	if !contains(status, "minutes ago") {
		t.Errorf("Status bar should contain relative time, got: %s", status)
	}

	// Test with zero time (never synced)
	model = NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})
	status = model.renderStatus()

	// Status should include "Last synced: never"
	if !contains(status, "Last synced: never") {
		t.Errorf("Status bar should contain 'Last synced: never' for zero time, got: %s", status)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestModel_HelpOverlayFromListView(t *testing.T) {
	// Test that ? key opens help overlay from list view
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Should start in list view
	if model.viewMode != viewModeList {
		t.Errorf("Initial viewMode = %d, want %d (viewModeList)", model.viewMode, viewModeList)
	}

	// Press ? to open help
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m := updatedModel.(Model)

	if m.viewMode != viewModeHelp {
		t.Errorf("After '?', viewMode = %d, want %d (viewModeHelp)", m.viewMode, viewModeHelp)
	}
}

func TestModel_HelpOverlayFromCommentsView(t *testing.T) {
	// Test that ? key opens help overlay from comments view
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	store, err := sync.NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store test issue
	issue := &sync.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Store test comment
	comment := &sync.Comment{
		ID:          1,
		IssueNumber: 1,
		Body:        "Test comment",
		Author:      "user1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.StoreComment(comment); err != nil {
		t.Fatalf("StoreComment() error = %v", err)
	}

	issues := []*sync.Issue{issue}
	model := NewModel(issues, []string{"number", "title"}, "updated", false, store, time.Time{})
	model.width = 120
	model.height = 30

	// Enter comments view
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	if m.viewMode != viewModeComments {
		t.Errorf("After Enter, viewMode = %d, want %d (viewModeComments)", m.viewMode, viewModeComments)
	}

	// Press ? to open help
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updatedModel.(Model)

	if m.viewMode != viewModeHelp {
		t.Errorf("After '?' in comments view, viewMode = %d, want %d (viewModeHelp)", m.viewMode, viewModeHelp)
	}
}

func TestModel_HelpOverlayDismissWithQuestionMark(t *testing.T) {
	// Test that ? key dismisses help overlay
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Open help
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m := updatedModel.(Model)

	// Press ? again to dismiss
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updatedModel.(Model)

	if m.viewMode != viewModeList {
		t.Errorf("After second '?', viewMode = %d, want %d (viewModeList)", m.viewMode, viewModeList)
	}
}

func TestModel_HelpOverlayDismissWithEsc(t *testing.T) {
	// Test that Esc key dismisses help overlay
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Open help
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m := updatedModel.(Model)

	// Press Esc to dismiss
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)

	if m.viewMode != viewModeList {
		t.Errorf("After Esc in help view, viewMode = %d, want %d (viewModeList)", m.viewMode, viewModeList)
	}
}

func TestModel_HelpOverlayReturnsToCorrectView(t *testing.T) {
	// Test that help overlay returns to the view it was opened from
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	store, err := sync.NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store test issue with comment
	issue := &sync.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	comment := &sync.Comment{
		ID:          1,
		IssueNumber: 1,
		Body:        "Test comment",
		Author:      "user1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.StoreComment(comment); err != nil {
		t.Fatalf("StoreComment() error = %v", err)
	}

	issues := []*sync.Issue{issue}
	model := NewModel(issues, []string{"number", "title"}, "updated", false, store, time.Time{})
	model.width = 120
	model.height = 30

	// Enter comments view
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updatedModel.(Model)

	// Open help from comments view
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updatedModel.(Model)

	// Dismiss help - should return to comments view
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updatedModel.(Model)

	if m.viewMode != viewModeComments {
		t.Errorf("After dismissing help from comments view, viewMode = %d, want %d (viewModeComments)", m.viewMode, viewModeComments)
	}
}

func TestModel_HelpOverlayBlocksOtherKeys(t *testing.T) {
	// Test that other keys don't work when help overlay is active
	issues := []*sync.Issue{
		{Number: 1, Title: "Test", State: "open", Author: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Number: 2, Title: "Second", State: "open", Author: "user2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	model := NewModel(issues, []string{"number", "title"}, "updated", false, nil, time.Time{})

	// Open help
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m := updatedModel.(Model)

	// Try navigation (should not work)
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updatedModel.(Model)

	if m.cursor != 0 {
		t.Errorf("With help overlay active, cursor = %d after 'j', want 0 (navigation blocked)", m.cursor)
	}

	// Try sort (should not work)
	initialSortBy := m.sortBy
	updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updatedModel.(Model)

	if m.sortBy != initialSortBy {
		t.Errorf("With help overlay active, sortBy changed from %s to %s (sort blocked)", initialSortBy, m.sortBy)
	}
}
