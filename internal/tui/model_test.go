package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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

	assert.Equal(t, issues, m.issues)
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
	issue := m.SelectedIssue()
	require.NotNil(t, issue)
	assert.Equal(t, 1, issue.Number)
	assert.Equal(t, "First issue", issue.Title)

	// Move cursor and check again
	m.cursor = 1
	issue = m.SelectedIssue()
	require.NotNil(t, issue)
	assert.Equal(t, 2, issue.Number)
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
