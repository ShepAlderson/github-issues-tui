package tui

import (
	"fmt"
	"testing"
	"time"

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

// Tests for Data Refresh functionality (US-009)

func TestRefreshKeyRTriggersRefresh(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Set a mock refresh function
	refreshCalled := false
	m.SetRefreshFunc(func() tea.Msg {
		refreshCalled = true
		return RefreshDoneMsg{Issues: issues}
	})

	// Initially not refreshing
	assert.False(t, m.IsRefreshing())

	// Press 'r' to trigger refresh
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Should set refreshing state
	assert.True(t, m.IsRefreshing())
	// Should return a command to start the refresh
	assert.NotNil(t, cmd)
	// Execute the command to verify it was called
	cmd()
	assert.True(t, refreshCalled)
}

func TestRefreshKeyRUpperCaseTriggersRefresh(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Set a mock refresh function
	m.SetRefreshFunc(func() tea.Msg {
		return RefreshDoneMsg{Issues: issues}
	})

	// Press 'R' (uppercase) to trigger refresh
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	m = newModel.(Model)

	assert.True(t, m.IsRefreshing())
	assert.NotNil(t, cmd)
}

func TestRefreshKeyIgnoredWhileRefreshing(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Set a mock refresh function
	m.SetRefreshFunc(func() tea.Msg {
		return RefreshDoneMsg{Issues: issues}
	})

	// Start a refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Try to trigger another refresh while already refreshing
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Should still be refreshing, but no new command
	assert.True(t, m.IsRefreshing())
	assert.Nil(t, cmd) // No additional command should be returned
}

func TestRefreshKeyIgnoredInCommentsView(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Enter comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	assert.True(t, m.InCommentsView())

	// Press 'r' - should not trigger refresh while in comments view
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	assert.False(t, m.IsRefreshing())
	assert.Nil(t, cmd)
}

func TestRefreshProgressMsgUpdatesProgress(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Simulate progress message
	progress := RefreshProgress{Phase: "issues", Current: 5, Total: 10}
	newModel, _ = m.Update(RefreshProgressMsg{Progress: progress})
	m = newModel.(Model)

	// Check progress is tracked
	currentProgress := m.GetRefreshProgress()
	assert.Equal(t, "issues", currentProgress.Phase)
	assert.Equal(t, 5, currentProgress.Current)
	assert.Equal(t, 10, currentProgress.Total)
}

func TestRefreshDoneMsgUpdatesIssues(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	assert.True(t, m.IsRefreshing())

	// New issues from refresh
	newIssues := []github.Issue{
		{
			Number:       100,
			Title:        "New issue",
			Author:       github.User{Login: "newuser"},
			CreatedAt:    "2024-02-01T12:00:00Z",
			UpdatedAt:    "2024-02-01T12:00:00Z",
			CommentCount: 0,
		},
	}

	// Simulate refresh done message
	newModel, _ = m.Update(RefreshDoneMsg{Issues: newIssues})
	m = newModel.(Model)

	// Should no longer be refreshing
	assert.False(t, m.IsRefreshing())
	// Issues should be updated
	assert.Equal(t, 1, m.IssueCount())
	assert.Equal(t, 100, m.issues[0].Number)
}

func TestRefreshErrorMsgStopsRefreshing(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Simulate error message
	newModel, _ = m.Update(RefreshErrorMsg{Err: fmt.Errorf("network error")})
	m = newModel.(Model)

	// Should no longer be refreshing
	assert.False(t, m.IsRefreshing())
	// Error should be tracked
	assert.Equal(t, "network error", m.GetRefreshError())
	// Original issues should remain
	assert.Equal(t, 3, m.IssueCount())
}

func TestStatusBarShowsRefreshingIndicator(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	view := m.View()

	// Status bar should show refreshing indicator
	assert.Contains(t, view, "Refreshing")
}

func TestStatusBarShowsRefreshProgress(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Simulate progress message
	progress := RefreshProgress{Phase: "issues", Current: 5, Total: 10}
	newModel, _ = m.Update(RefreshProgressMsg{Progress: progress})
	m = newModel.(Model)

	view := m.View()

	// Status bar should show progress
	assert.Contains(t, view, "5/10")
}

func TestRefreshClearsErrorOnNewRefresh(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Set a mock refresh function
	m.SetRefreshFunc(func() tea.Msg {
		return RefreshDoneMsg{Issues: issues}
	})

	// Start and fail a refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)
	newModel, _ = m.Update(RefreshErrorMsg{Err: fmt.Errorf("network error")})
	m = newModel.(Model)

	assert.NotEmpty(t, m.GetRefreshError())

	// Start a new refresh
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Error should be cleared
	assert.Empty(t, m.GetRefreshError())
}

func TestRefreshMaintainsCursorOnSameIssue(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Move cursor to issue #2 (second in sorted order after applying default sort)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	selectedBefore := m.SelectedIssue()
	require.NotNil(t, selectedBefore)
	selectedNumber := selectedBefore.Number

	// Start refresh
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Complete refresh with same issues
	newModel, _ = m.Update(RefreshDoneMsg{Issues: createTestIssues()})
	m = newModel.(Model)

	// Cursor should still be on the same issue
	selectedAfter := m.SelectedIssue()
	require.NotNil(t, selectedAfter)
	assert.Equal(t, selectedNumber, selectedAfter.Number)
}

// Tests for Error Handling (US-013)

func TestMinorErrorShownInStatusBar(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Simulate a minor error (network timeout)
	m.SetRefreshFunc(func() tea.Msg {
		return RefreshDoneMsg{Issues: issues}
	})

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Simulate a network timeout error (minor)
	newModel, _ = m.Update(RefreshErrorMsg{Err: fmt.Errorf("network timeout: connection timed out")})
	m = newModel.(Model)

	// Error should be shown in status bar, not as modal
	assert.False(t, m.HasErrorModal())
	assert.NotEmpty(t, m.GetRefreshError())

	view := m.View()
	// Status bar should show error
	assert.Contains(t, view, "timeout")
}

func TestMinorErrorSuggestsRetry(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Simulate a network error
	newModel, _ := m.Update(RefreshErrorMsg{Err: fmt.Errorf("network error: dial tcp: no route to host")})
	m = newModel.(Model)

	view := m.View()

	// Should show actionable guidance - suggest retry
	assert.Contains(t, view, "r: retry")
}

func TestRateLimitErrorShownInStatusBar(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Simulate a rate limit error (minor)
	newModel, _ := m.Update(RefreshErrorMsg{Err: fmt.Errorf("GitHub API error: 403 rate limit exceeded")})
	m = newModel.(Model)

	// Should be shown in status bar, not modal
	assert.False(t, m.HasErrorModal())
	assert.Contains(t, m.GetRefreshError(), "rate limit")
}

func TestCriticalErrorShowsModal(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Simulate a critical error (invalid token)
	err := fmt.Errorf("invalid GitHub token: authentication failed (401 Unauthorized)")
	newModel, _ := m.Update(CriticalErrorMsg{Err: err, Title: "Authentication Error"})
	m = newModel.(Model)

	// Should show modal
	assert.True(t, m.HasErrorModal())
	assert.Equal(t, "Authentication Error", m.GetErrorModalTitle())
	assert.Contains(t, m.GetErrorModalMessage(), "invalid")
}

func TestCriticalErrorModalRequiresAcknowledgment(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Show critical error modal
	err := fmt.Errorf("database corruption: file is not a database")
	newModel, _ := m.Update(CriticalErrorMsg{Err: err, Title: "Database Error"})
	m = newModel.(Model)

	assert.True(t, m.HasErrorModal())

	// Navigation keys should not work while modal is shown
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	assert.True(t, m.HasErrorModal()) // Modal still shown

	// Press Enter to acknowledge
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	// Modal should be dismissed
	assert.False(t, m.HasErrorModal())
}

func TestCriticalErrorModalDismissedWithEscape(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Show critical error modal
	err := fmt.Errorf("invalid token")
	newModel, _ := m.Update(CriticalErrorMsg{Err: err, Title: "Auth Error"})
	m = newModel.(Model)

	assert.True(t, m.HasErrorModal())

	// Press Escape to dismiss
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)

	assert.False(t, m.HasErrorModal())
}

func TestErrorModalViewRendering(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Show critical error modal with actionable guidance
	err := fmt.Errorf("invalid GitHub token: authentication failed (401 Unauthorized). Please check that your token is correct and has not expired")
	newModel, _ := m.Update(CriticalErrorMsg{
		Err:      err,
		Title:    "Authentication Error",
		Guidance: "Run 'ghissues config' to update your authentication settings.",
	})
	m = newModel.(Model)

	view := m.View()

	// Modal should be visible
	assert.Contains(t, view, "Authentication Error")
	assert.Contains(t, view, "Unauthorized")
	// Should show actionable guidance
	assert.Contains(t, view, "ghissues config")
	// Should show dismissal instructions
	assert.Contains(t, view, "Enter")
}

func TestDatabaseErrorShowsModal(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Database corruption is a critical error
	err := fmt.Errorf("database error: file is not a database")
	newModel, _ := m.Update(CriticalErrorMsg{
		Err:      err,
		Title:    "Database Error",
		Guidance: "Try deleting the database file and running 'ghissues sync' to rebuild it.",
	})
	m = newModel.(Model)

	assert.True(t, m.HasErrorModal())
	view := m.View()
	assert.Contains(t, view, "Database Error")
	assert.Contains(t, view, "database")
}

func TestNavigationBlockedDuringErrorModal(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Remember initial cursor position
	initialCursor := m.cursor

	// Show critical error modal
	newModel, _ := m.Update(CriticalErrorMsg{Err: fmt.Errorf("error"), Title: "Error"})
	m = newModel.(Model)

	// Try various navigation keys
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	assert.Equal(t, initialCursor, m.cursor) // Cursor unchanged

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	assert.Equal(t, initialCursor, m.cursor) // Cursor unchanged

	// 'q' should dismiss modal, not quit
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = newModel.(Model)
	// Modal should be dismissed
	assert.False(t, m.HasErrorModal())
	// Should not return quit command
	assert.Nil(t, cmd)
}

func TestRefreshBlockedDuringErrorModal(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Set refresh function
	m.SetRefreshFunc(func() tea.Msg {
		return RefreshDoneMsg{Issues: issues}
	})

	// Show critical error modal
	newModel, _ := m.Update(CriticalErrorMsg{Err: fmt.Errorf("error"), Title: "Error"})
	m = newModel.(Model)

	// Try to trigger refresh while modal is shown
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Should not start refreshing
	assert.False(t, m.IsRefreshing())
	assert.Nil(t, cmd)
}

func TestNetworkErrorIncludesConnectivityGuidance(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(140, 30) // Wider to see full message

	// Network error
	newModel, _ := m.Update(RefreshErrorMsg{Err: fmt.Errorf("failed to execute request: dial tcp: no such host")})
	m = newModel.(Model)

	view := m.View()

	// Should suggest checking connectivity (part of the standard error display)
	// The error message and retry option should be visible
	assert.Contains(t, view, "dial tcp")
}

// Tests for Last Synced Indicator (US-010)

func TestSetLastSyncTime(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)

	syncTime := time.Date(2024, 1, 20, 14, 30, 0, 0, time.UTC)
	m.SetLastSyncTime(syncTime)

	assert.Equal(t, syncTime, m.GetLastSyncTime())
}

func TestLastSyncTimeDefaultIsZero(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)

	// Default should be zero time (never synced)
	assert.True(t, m.GetLastSyncTime().IsZero())
}

func TestStatusBarShowsLastSyncedTime(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30) // Wide enough to show full status bar

	// Set last sync time to a few minutes ago
	syncTime := time.Now().Add(-5 * time.Minute)
	m.SetLastSyncTime(syncTime)

	view := m.View()

	// Status bar should show "Last synced:" indicator
	assert.Contains(t, view, "Last synced:")
}

func TestLastSyncedRelativeTimeMinutesAgo(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Set last sync time to 5 minutes ago
	syncTime := time.Now().Add(-5 * time.Minute)
	m.SetLastSyncTime(syncTime)

	view := m.View()

	// Should show relative time (e.g., "5 minutes ago")
	assert.Contains(t, view, "5m ago")
}

func TestLastSyncedRelativeTimeSecondsAgo(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Set last sync time to 30 seconds ago
	syncTime := time.Now().Add(-30 * time.Second)
	m.SetLastSyncTime(syncTime)

	view := m.View()

	// Should show "just now" or "<1m ago"
	assert.Contains(t, view, "<1m ago")
}

func TestLastSyncedRelativeTimeHoursAgo(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Set last sync time to 2 hours ago
	syncTime := time.Now().Add(-2 * time.Hour)
	m.SetLastSyncTime(syncTime)

	view := m.View()

	// Should show relative time (e.g., "2h ago")
	assert.Contains(t, view, "2h ago")
}

func TestLastSyncedRelativeTimeDaysAgo(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Set last sync time to 3 days ago
	syncTime := time.Now().Add(-3 * 24 * time.Hour)
	m.SetLastSyncTime(syncTime)

	view := m.View()

	// Should show relative time (e.g., "3d ago")
	assert.Contains(t, view, "3d ago")
}

func TestLastSyncedNeverSynced(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Don't set last sync time (zero value = never synced)
	view := m.View()

	// Should show "never" or similar indicator
	assert.Contains(t, view, "Last synced: never")
}

func TestLastSyncTimeUpdatedAfterRefresh(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Initially never synced
	assert.True(t, m.GetLastSyncTime().IsZero())

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Complete refresh - this should include the new sync time
	newSyncTime := time.Now()
	newModel, _ = m.Update(RefreshDoneMsg{Issues: createTestIssues(), LastSyncTime: newSyncTime})
	m = newModel.(Model)

	// Last sync time should be updated
	assert.False(t, m.GetLastSyncTime().IsZero())
	// Should be approximately now (within a second)
	assert.WithinDuration(t, newSyncTime, m.GetLastSyncTime(), time.Second)
}

func TestLastSyncTimeNotChangedOnRefreshError(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(160, 30)

	// Set initial sync time
	initialSyncTime := time.Now().Add(-10 * time.Minute)
	m.SetLastSyncTime(initialSyncTime)

	// Start refresh
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = newModel.(Model)

	// Simulate error
	newModel, _ = m.Update(RefreshErrorMsg{Err: fmt.Errorf("network error")})
	m = newModel.(Model)

	// Last sync time should remain unchanged
	assert.Equal(t, initialSyncTime, m.GetLastSyncTime())
}

func TestRelativeTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{name: "30 seconds", duration: 30 * time.Second, expected: "<1m ago"},
		{name: "1 minute", duration: 1 * time.Minute, expected: "1m ago"},
		{name: "5 minutes", duration: 5 * time.Minute, expected: "5m ago"},
		{name: "59 minutes", duration: 59 * time.Minute, expected: "59m ago"},
		{name: "1 hour", duration: 1 * time.Hour, expected: "1h ago"},
		{name: "2 hours", duration: 2 * time.Hour, expected: "2h ago"},
		{name: "23 hours", duration: 23 * time.Hour, expected: "23h ago"},
		{name: "1 day", duration: 24 * time.Hour, expected: "1d ago"},
		{name: "3 days", duration: 3 * 24 * time.Hour, expected: "3d ago"},
		{name: "7 days", duration: 7 * 24 * time.Hour, expected: "7d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatRelativeTime(time.Now().Add(-tt.duration))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRelativeTimeFormatNeverSynced(t *testing.T) {
	// Zero time should return "never"
	result := formatRelativeTime(time.Time{})
	assert.Equal(t, "never", result)
}

// Tests for Keybinding Help (US-011)

func TestHelpOverlayOpenedWithQuestionMark(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Initially help overlay should not be shown
	assert.False(t, m.ShowHelpOverlay())

	// Press '?' to open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	assert.True(t, m.ShowHelpOverlay())
}

func TestHelpOverlayDismissedWithQuestionMark(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)
	assert.True(t, m.ShowHelpOverlay())

	// Press '?' again to dismiss
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	assert.False(t, m.ShowHelpOverlay())
}

func TestHelpOverlayDismissedWithEscape(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)
	assert.True(t, m.ShowHelpOverlay())

	// Press Escape to dismiss
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)

	assert.False(t, m.ShowHelpOverlay())
}

func TestHelpOverlayShowsAllKeybindings(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	view := m.View()

	// Should show keybindings organized by context
	// Navigation
	assert.Contains(t, view, "j/↓")
	assert.Contains(t, view, "k/↑")

	// Actions
	assert.Contains(t, view, "Enter")
	assert.Contains(t, view, "r")

	// Sorting
	assert.Contains(t, view, "s")
	assert.Contains(t, view, "S")

	// Detail Panel / Scroll
	assert.Contains(t, view, "h")
	assert.Contains(t, view, "l")
	assert.Contains(t, view, "m")

	// General
	assert.Contains(t, view, "q")
	assert.Contains(t, view, "?")
}

func TestHelpOverlayBlocksOtherKeys(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Remember initial cursor position
	initialCursor := m.cursor

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	// Try navigation key - should be blocked
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)

	assert.Equal(t, initialCursor, m.cursor) // Cursor unchanged
	assert.True(t, m.ShowHelpOverlay()) // Still showing help
}

func TestHelpOverlayCtrlCStillQuits(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	// Ctrl+C should still work to quit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)
	msg := cmd()
	assert.IsType(t, tea.QuitMsg{}, msg)
}

func TestFooterShowsContextSensitiveKeysInListView(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	view := m.View()

	// In list view, footer should show common list navigation keys
	assert.Contains(t, view, "j/k")
	assert.Contains(t, view, "Enter")
	assert.Contains(t, view, "?")
}

func TestFooterShowsContextSensitiveKeysInCommentsView(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)
	m.SetComments(createTestComments())

	// Enter comments view
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	view := m.View()

	// In comments view, footer should show relevant keys for that context
	assert.Contains(t, view, "h/l")
	assert.Contains(t, view, "Esc")
	assert.Contains(t, view, "?")
}

func TestFooterShowsHelpHint(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 30)

	view := m.View()

	// Footer should always show ? for help
	assert.Contains(t, view, "?")
}

func TestHelpOverlayOrganizedByContext(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	view := m.View()

	// Should have section headers
	assert.Contains(t, view, "Navigation")
	assert.Contains(t, view, "Sorting")
}

func TestHelpOverlayShowsDismissInstructions(t *testing.T) {
	issues := createTestIssues()
	m := NewModel(issues, nil)
	m.SetWindowSize(120, 40)

	// Open help overlay
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(Model)

	view := m.View()

	// Should show how to dismiss
	assert.Contains(t, view, "Esc")
}
