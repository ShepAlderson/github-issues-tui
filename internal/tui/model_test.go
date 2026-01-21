package tui

import (
	"fmt"
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

// Helper to create test issues with body content for detail view tests
func createDetailTestIssues() []github.Issue {
	return []github.Issue{
		{
			Number:       42,
			Title:        "Test issue with markdown body",
			Body:         "## Description\n\nThis is a **bold** description with `code`.\n\n- Item 1\n- Item 2",
			Author:       github.User{Login: "testuser"},
			CreatedAt:    "2024-01-15T10:30:00Z",
			UpdatedAt:    "2024-01-20T14:45:00Z",
			CommentCount: 5,
			Labels:       []github.Label{{Name: "bug", Color: "d73a4a"}, {Name: "priority", Color: "0052cc"}},
			Assignees:    []github.User{{Login: "assignee1"}, {Login: "assignee2"}},
		},
		{
			Number:       43,
			Title:        "Another issue",
			Body:         "Simple body text",
			Author:       github.User{Login: "user2"},
			CreatedAt:    "2024-01-16T10:30:00Z",
			UpdatedAt:    "2024-01-21T14:45:00Z",
			CommentCount: 0,
			Labels:       nil,
			Assignees:    nil,
		},
	}
}

func TestDetailPanelShowsSelectedIssue(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	view := m.View()

	// Detail panel should show the selected issue info
	// Issue 43 is first because it has a more recent UpdatedAt date
	assert.Contains(t, view, "#43")
	assert.Contains(t, view, "Another issue")
	assert.Contains(t, view, "user2")
}

func TestDetailPanelShowsHeader(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Select issue 42 (second in sorted order)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	view := m.View()

	// Header should show: issue number, title, author, status indicators
	assert.Contains(t, view, "#42")
	assert.Contains(t, view, "Test issue with markdown body")
	assert.Contains(t, view, "testuser")
}

func TestDetailPanelShowsDates(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Select issue 42 (second in sorted order)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	view := m.View()

	// Should show created and updated dates
	assert.Contains(t, view, "2024-01-15")
	assert.Contains(t, view, "2024-01-20")
}

func TestDetailPanelShowsLabels(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Select issue 42 (second in sorted order)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	view := m.View()

	// Should show labels
	assert.Contains(t, view, "bug")
	assert.Contains(t, view, "priority")
}

func TestDetailPanelShowsAssignees(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Select issue 42 (second in sorted order)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	view := m.View()

	// Should show assignees
	assert.Contains(t, view, "assignee1")
	assert.Contains(t, view, "assignee2")
}

func TestDetailPanelShowsBody(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Select issue 42 (second in sorted order)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	view := m.View()

	// Should show body content (rendered markdown may have transformed text)
	assert.Contains(t, view, "Description")
}

func TestDetailPanelToggleRawMarkdown(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Select issue 42 (second in sorted order)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	// Initial state should be rendered markdown
	assert.False(t, m.IsRawMarkdown())

	// Press 'm' to toggle to raw markdown
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = newModel.(Model)

	assert.True(t, m.IsRawMarkdown())
	view := m.View()
	// Raw view should contain markdown syntax
	assert.Contains(t, view, "**bold**")
	assert.Contains(t, view, "`code`")

	// Press 'm' again to toggle back to rendered
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = newModel.(Model)

	assert.False(t, m.IsRawMarkdown())
}

func TestDetailPanelScroll(t *testing.T) {
	// Create an issue with a long body
	issues := []github.Issue{
		{
			Number:       1,
			Title:        "Long issue",
			Body:         "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\nLine 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20",
			Author:       github.User{Login: "testuser"},
			CreatedAt:    "2024-01-15T10:30:00Z",
			UpdatedAt:    "2024-01-20T14:45:00Z",
			CommentCount: 0,
		},
	}

	m := NewModel(issues, nil)
	m.SetWindowSize(120, 15) // Small height to trigger scrolling

	// Initial scroll should be at 0
	assert.Equal(t, 0, m.DetailScrollOffset())

	// Press 'l' to scroll down in the detail panel
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = newModel.(Model)

	assert.Greater(t, m.DetailScrollOffset(), 0)

	// Press 'h' to scroll up in the detail panel
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = newModel.(Model)

	assert.Equal(t, 0, m.DetailScrollOffset())
}

func TestEnterKeyOpensCommentsView(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Initial state: not in comments view
	assert.False(t, m.InCommentsView())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	assert.True(t, m.InCommentsView())

	// Press Escape to go back
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)

	assert.False(t, m.InCommentsView())
}

func TestDetailPanelNoLabelsOrAssignees(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Issue 43 is first (more recent UpdatedAt) and has no labels or assignees
	view := m.View()

	// Should still render without crashing, just not show label/assignee sections
	assert.Contains(t, view, "#43")
	assert.Contains(t, view, "Another issue")
}

// Helper to create test comments
func createTestComments() []github.Comment {
	return []github.Comment{
		{
			ID:        "comment1",
			Body:      "This is the **first** comment with `code`.",
			Author:    github.User{Login: "commenter1"},
			CreatedAt: "2024-01-10T10:00:00Z",
		},
		{
			ID:        "comment2",
			Body:      "Second comment here.",
			Author:    github.User{Login: "commenter2"},
			CreatedAt: "2024-01-11T14:30:00Z",
		},
		{
			ID:        "comment3",
			Body:      "Third comment with more details.\n\n- Point 1\n- Point 2",
			Author:    github.User{Login: "commenter3"},
			CreatedAt: "2024-01-12T09:15:00Z",
		},
	}
}

func TestSetComments(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)

	comments := createTestComments()
	m.SetComments(comments)

	assert.Equal(t, 3, len(m.GetComments()))
}

func TestCommentsViewShowsIssueHeader(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view := m.View()

	// Comments view should replace main interface (no issue list panel)
	// and show issue title/number as header
	assert.True(t, m.InCommentsView())
	assert.Contains(t, view, "#43") // Issue number
	assert.Contains(t, view, "Another issue") // Issue title
}

func TestCommentsViewShowsCommentsChronologically(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view := m.View()

	// Comments should be displayed chronologically
	// Check that all comment authors are visible
	assert.Contains(t, view, "commenter1")
	assert.Contains(t, view, "commenter2")
	assert.Contains(t, view, "commenter3")
}

func TestCommentsViewShowsCommentDate(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view := m.View()

	// Comments should show dates
	assert.Contains(t, view, "2024-01-10")
	assert.Contains(t, view, "2024-01-11")
	assert.Contains(t, view, "2024-01-12")
}

func TestCommentsViewShowsCommentBody(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view := m.View()

	// Comments should show body content
	assert.Contains(t, view, "first")
	assert.Contains(t, view, "Second comment")
}

func TestCommentsViewToggleMarkdown(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	// Default is rendered markdown
	assert.False(t, m.IsRawMarkdown())

	// Press 'm' to toggle to raw markdown
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = newModel.(Model)

	assert.True(t, m.IsRawMarkdown())
	view := m.View()
	// Raw view should contain markdown syntax
	assert.Contains(t, view, "**first**")
	assert.Contains(t, view, "`code`")

	// Press 'm' again to toggle back
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = newModel.(Model)

	assert.False(t, m.IsRawMarkdown())
}

func TestCommentsViewScrollable(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 15) // Small height to test scrolling

	// Create many comments to test scrolling
	manyComments := []github.Comment{}
	for i := 0; i < 20; i++ {
		manyComments = append(manyComments, github.Comment{
			ID:        fmt.Sprintf("comment%d", i),
			Body:      fmt.Sprintf("Comment %d body text", i),
			Author:    github.User{Login: fmt.Sprintf("user%d", i)},
			CreatedAt: fmt.Sprintf("2024-01-%02dT10:00:00Z", i+1),
		})
	}
	m.SetComments(manyComments)

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	// Initial scroll should be 0
	assert.Equal(t, 0, m.CommentsScrollOffset())

	// Press 'l' to scroll down
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = newModel.(Model)

	assert.Greater(t, m.CommentsScrollOffset(), 0)

	// Press 'h' to scroll back up
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = newModel.(Model)

	assert.Equal(t, 0, m.CommentsScrollOffset())
}

func TestCommentsViewEscapeReturnsToIssueList(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	assert.True(t, m.InCommentsView())

	// Press Escape to return to issue list
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)

	assert.False(t, m.InCommentsView())
}

func TestCommentsViewQKeyReturnsToIssueList(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)
	m.SetComments(createTestComments())

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	assert.True(t, m.InCommentsView())

	// Press 'q' to return to issue list (not quit the app when in comments view)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = newModel.(Model)

	assert.False(t, m.InCommentsView())
}

func TestCommentsViewEmptyState(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)
	// No comments set

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view := m.View()

	// Should show a message when there are no comments
	assert.Contains(t, view, "No comments")
}

func TestCommentsViewReplacesMainInterface(t *testing.T) {
	issues := createDetailTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)
	m.SetComments(createTestComments())

	// Before entering comments view, should see the issue list
	view := m.View()
	assert.Contains(t, view, "#") // Issue list column header

	// Press Enter to open comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view = m.View()

	// Comments view should be a drill-down (full screen), not split panel
	// The issue list should not be visible
	// Check that the comments header is present
	assert.Contains(t, view, "Comments")
}
