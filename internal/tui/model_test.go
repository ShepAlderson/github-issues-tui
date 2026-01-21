package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestIssues() []github.Issue {
	return []github.Issue{
		{
			Number:       1,
			Title:        "First issue",
			Author:       github.User{Login: "user1"},
			CreatedAt:    "2024-01-01T12:00:00Z",
			UpdatedAt:    "2024-01-15T12:00:00Z",
			CommentCount: 5,
		},
		{
			Number:       2,
			Title:        "Second issue",
			Author:       github.User{Login: "user2"},
			CreatedAt:    "2024-01-02T12:00:00Z",
			UpdatedAt:    "2024-01-16T12:00:00Z",
			CommentCount: 3,
		},
		{
			Number:       3,
			Title:        "Third issue",
			Author:       github.User{Login: "user3"},
			CreatedAt:    "2024-01-03T12:00:00Z",
			UpdatedAt:    "2024-01-17T12:00:00Z",
			CommentCount: 0,
		},
	}
}

func TestNewModel(t *testing.T) {
	issues := createTestIssues()
	columns := []string{"number", "title", "author"}

	m := NewModel(issues, columns)

	// Issues are sorted by updated date descending by default
	// createTestIssues has: 1 (Jan 15), 2 (Jan 16), 3 (Jan 17)
	// So sorted: 3, 2, 1
	assert.Equal(t, 3, m.issues[0].Number)
	assert.Equal(t, 2, m.issues[1].Number)
	assert.Equal(t, 1, m.issues[2].Number)
	assert.Equal(t, columns, m.columns)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 3, m.IssueCount())
}

func TestNewModelEmpty(t *testing.T) {
	m := NewModel(nil, nil)

	assert.Empty(t, m.issues)
	assert.Equal(t, DefaultColumns(), m.columns)
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 0, m.IssueCount())
}

func TestModelNavigationDown(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)

	// Initial cursor position
	assert.Equal(t, 0, m.cursor)

	// Move down with j key
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	assert.Equal(t, 1, m.cursor)

	// Move down with down arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	assert.Equal(t, 2, m.cursor)

	// Should not go past last item
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	assert.Equal(t, 2, m.cursor)
}

func TestModelNavigationUp(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.cursor = 2 // Start at bottom

	// Move up with k key
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	assert.Equal(t, 1, m.cursor)

	// Move up with up arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	assert.Equal(t, 0, m.cursor)

	// Should not go past first item
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	assert.Equal(t, 0, m.cursor)
}

func TestModelQuitKeys(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)

	// q should quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd)
	msg := cmd()
	assert.IsType(t, tea.QuitMsg{}, msg)

	// Ctrl+C should quit
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)
	msg = cmd()
	assert.IsType(t, tea.QuitMsg{}, msg)
}

func TestModelSelectedIssue(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)

	// Get selected issue at cursor 0
	// Default sort is by updated date descending, so issue 3 is first (most recently updated)
	issue := m.SelectedIssue()
	require.NotNil(t, issue)
	assert.Equal(t, 3, issue.Number)
	assert.Equal(t, "Third issue", issue.Title)

	// Move cursor and check again
	m.cursor = 1
	issue = m.SelectedIssue()
	require.NotNil(t, issue)
	assert.Equal(t, 2, issue.Number) // Second most recently updated
}

func TestModelSelectedIssueEmpty(t *testing.T) {
	m := NewModel(nil, nil)

	issue := m.SelectedIssue()
	assert.Nil(t, issue)
}

func TestModelIssueCount(t *testing.T) {
	tests := []struct {
		name     string
		issues   []github.Issue
		expected int
	}{
		{name: "three issues", issues: createTestIssues(), expected: 3},
		{name: "empty", issues: nil, expected: 0},
		{name: "one issue", issues: []github.Issue{{Number: 1}}, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(tt.issues, nil)
			assert.Equal(t, tt.expected, m.IssueCount())
		})
	}
}

func TestModelSetWindowSize(t *testing.T) {
	m := NewModel(createTestIssues(), nil)

	m.SetWindowSize(100, 50)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}

func TestModelWindowSizeMsg(t *testing.T) {
	m := NewModel(createTestIssues(), nil)

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = newModel.(Model)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestDefaultColumns(t *testing.T) {
	cols := DefaultColumns()
	assert.Equal(t, []string{"number", "title", "author", "date", "comments"}, cols)
}

func TestModelViewContainsIssueCount(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(80, 24)

	view := m.View()

	// Status area should show issue count
	assert.Contains(t, view, "3 issues")
}

func TestModelViewContainsSelectedHighlight(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(80, 24)

	view := m.View()

	// Selected issue should be in the view
	assert.Contains(t, view, "First issue")
}

func TestModelViewEmptyState(t *testing.T) {
	m := NewModel(nil, nil)
	m.SetWindowSize(80, 24)

	view := m.View()

	// Should show no issues message
	assert.Contains(t, view, "No issues")
	assert.Contains(t, view, "0 issues")
}

// Helper to create test issues with varied dates and counts for sorting tests
func createSortTestIssues() []github.Issue {
	return []github.Issue{
		{
			Number:       10,
			Title:        "Issue ten",
			Author:       github.User{Login: "alice"},
			CreatedAt:    "2024-01-05T12:00:00Z",
			UpdatedAt:    "2024-01-20T12:00:00Z",
			CommentCount: 2,
		},
		{
			Number:       5,
			Title:        "Issue five",
			Author:       github.User{Login: "bob"},
			CreatedAt:    "2024-01-10T12:00:00Z",
			UpdatedAt:    "2024-01-15T12:00:00Z",
			CommentCount: 10,
		},
		{
			Number:       15,
			Title:        "Issue fifteen",
			Author:       github.User{Login: "charlie"},
			CreatedAt:    "2024-01-01T12:00:00Z",
			UpdatedAt:    "2024-01-25T12:00:00Z",
			CommentCount: 5,
		},
	}
}

func TestNewModelWithSort(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByUpdated, config.SortDesc)

	assert.Equal(t, config.SortByUpdated, m.sortField)
	assert.Equal(t, config.SortDesc, m.sortOrder)
	assert.Equal(t, 3, m.IssueCount())
}

func TestDefaultSortIsUpdatedDescending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	// Default sort should be by updated date descending (most recently updated first)
	assert.Equal(t, config.SortByUpdated, m.sortField)
	assert.Equal(t, config.SortDesc, m.sortOrder)

	// First issue should be the most recently updated (issue 15)
	selected := m.SelectedIssue()
	require.NotNil(t, selected)
	assert.Equal(t, 15, selected.Number)
}

func TestSortByUpdatedDescending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByUpdated, config.SortDesc)
	m.SetWindowSize(80, 24)

	// Most recently updated first: 15 (Jan 25), 10 (Jan 20), 5 (Jan 15)
	assert.Equal(t, 15, m.issues[0].Number)
	assert.Equal(t, 10, m.issues[1].Number)
	assert.Equal(t, 5, m.issues[2].Number)
}

func TestSortByUpdatedAscending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByUpdated, config.SortAsc)

	// Oldest updated first: 5 (Jan 15), 10 (Jan 20), 15 (Jan 25)
	assert.Equal(t, 5, m.issues[0].Number)
	assert.Equal(t, 10, m.issues[1].Number)
	assert.Equal(t, 15, m.issues[2].Number)
}

func TestSortByCreatedDescending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByCreated, config.SortDesc)

	// Most recently created first: 5 (Jan 10), 10 (Jan 5), 15 (Jan 1)
	assert.Equal(t, 5, m.issues[0].Number)
	assert.Equal(t, 10, m.issues[1].Number)
	assert.Equal(t, 15, m.issues[2].Number)
}

func TestSortByCreatedAscending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByCreated, config.SortAsc)

	// Oldest created first: 15 (Jan 1), 10 (Jan 5), 5 (Jan 10)
	assert.Equal(t, 15, m.issues[0].Number)
	assert.Equal(t, 10, m.issues[1].Number)
	assert.Equal(t, 5, m.issues[2].Number)
}

func TestSortByNumberDescending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByNumber, config.SortDesc)

	// Highest number first: 15, 10, 5
	assert.Equal(t, 15, m.issues[0].Number)
	assert.Equal(t, 10, m.issues[1].Number)
	assert.Equal(t, 5, m.issues[2].Number)
}

func TestSortByNumberAscending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByNumber, config.SortAsc)

	// Lowest number first: 5, 10, 15
	assert.Equal(t, 5, m.issues[0].Number)
	assert.Equal(t, 10, m.issues[1].Number)
	assert.Equal(t, 15, m.issues[2].Number)
}

func TestSortByCommentsDescending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByComments, config.SortDesc)

	// Most comments first: 5 (10 comments), 15 (5 comments), 10 (2 comments)
	assert.Equal(t, 5, m.issues[0].Number)
	assert.Equal(t, 15, m.issues[1].Number)
	assert.Equal(t, 10, m.issues[2].Number)
}

func TestSortByCommentsAscending(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByComments, config.SortAsc)

	// Fewest comments first: 10 (2 comments), 15 (5 comments), 5 (10 comments)
	assert.Equal(t, 10, m.issues[0].Number)
	assert.Equal(t, 15, m.issues[1].Number)
	assert.Equal(t, 5, m.issues[2].Number)
}

func TestCycleSortFieldWithSKey(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	// Initial sort: updated
	assert.Equal(t, config.SortByUpdated, m.sortField)

	// Press 's' to cycle to created
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)
	assert.Equal(t, config.SortByCreated, m.sortField)

	// Press 's' to cycle to number
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)
	assert.Equal(t, config.SortByNumber, m.sortField)

	// Press 's' to cycle to comments
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)
	assert.Equal(t, config.SortByComments, m.sortField)

	// Press 's' to wrap back to updated
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)
	assert.Equal(t, config.SortByUpdated, m.sortField)
}

func TestReverseSortOrderWithShiftSKey(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	// Initial order: descending
	assert.Equal(t, config.SortDesc, m.sortOrder)

	// Press 'S' to toggle to ascending
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m = newModel.(Model)
	assert.Equal(t, config.SortAsc, m.sortOrder)

	// Press 'S' to toggle back to descending
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m = newModel.(Model)
	assert.Equal(t, config.SortDesc, m.sortOrder)
}

func TestSortReordersIssuesAndResetsCursor(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	// Move cursor to second item
	m.cursor = 1

	// Press 's' to change sort
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)

	// Cursor should reset to 0 after sort change
	assert.Equal(t, 0, m.cursor)
}

func TestStatusBarShowsCurrentSort(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 24)

	view := m.View()

	// Status bar should show current sort field and order
	assert.Contains(t, view, "Updated")
	assert.Contains(t, view, "↓") // Descending indicator
}

func TestStatusBarShowsAscendingIndicator(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByNumber, config.SortAsc)
	m.SetWindowSize(120, 24)

	view := m.View()

	assert.Contains(t, view, "Number")
	assert.Contains(t, view, "↑") // Ascending indicator
}

func TestGetSortConfig(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModelWithSort(issues, nil, config.SortByComments, config.SortAsc)

	field, order := m.GetSortConfig()

	assert.Equal(t, config.SortByComments, field)
	assert.Equal(t, config.SortAsc, order)
}

func TestSortChangedInitiallyFalse(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	assert.False(t, m.SortChanged())
}

func TestSortChangedAfterCycleField(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	// Press 's' to cycle sort field
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newModel.(Model)

	assert.True(t, m.SortChanged())
}

func TestSortChangedAfterToggleOrder(t *testing.T) {
	issues := createSortTestIssues()
	m := NewModel(issues, nil)

	// Press 'S' to toggle sort order
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	m = newModel.(Model)

	assert.True(t, m.SortChanged())
}
