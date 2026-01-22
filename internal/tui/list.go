package tui

import (
	"github.com/shepbook/ghissues/internal/sort"
	"github.com/shepbook/ghissues/internal/storage"
)

// IssueList represents the state of the issue list view
type IssueList struct {
	Issues         []storage.Issue
	UnsortedIssues []storage.Issue // Keep original unsorted issues
	Columns        []Column
	Cursor         int
	Selected       *storage.Issue
	ViewportHeight int
	ViewportOffset int
	SortField      string
	SortDescending bool
}

// NewIssueList creates a new issue list model
func NewIssueList(issues []storage.Issue, columns []Column) *IssueList {
	// Apply default sorting
	sortedIssues := sort.SortIssues(issues, sort.GetDefaultSortField(), sort.GetDefaultSortDescending())

	return &IssueList{
		Issues:         sortedIssues,
		UnsortedIssues: issues,
		Columns:        columns,
		Cursor:         0,
		Selected:       nil,
		ViewportHeight: 10,
		ViewportOffset: 0,
		SortField:      sort.GetDefaultSortField(),
		SortDescending: sort.GetDefaultSortDescending(),
	}
}

// NewIssueListWithSort creates a new issue list model with specific sort settings
func NewIssueListWithSort(issues []storage.Issue, columns []Column, sortField string, sortDescending bool) *IssueList {
	// Validate sort field
	if err := sort.ValidateSortField(sortField); err != nil {
		// Fall back to default if invalid
		sortField = sort.GetDefaultSortField()
		sortDescending = sort.GetDefaultSortDescending()
	}

	// Apply specified sorting
	sortedIssues := sort.SortIssues(issues, sortField, sortDescending)

	return &IssueList{
		Issues:         sortedIssues,
		UnsortedIssues: issues,
		Columns:        columns,
		Cursor:         0,
		Selected:       nil,
		ViewportHeight: 10,
		ViewportOffset: 0,
		SortField:      sortField,
		SortDescending: sortDescending,
	}
}

// MoveCursor moves the cursor up or down by the specified delta
func (m *IssueList) MoveCursor(delta int) {
	if len(m.Issues) == 0 {
		return
	}

	newCursor := m.Cursor + delta

	// Boundary checks
	if newCursor < 0 {
		newCursor = 0
	} else if newCursor >= len(m.Issues) {
		newCursor = len(m.Issues) - 1
	}

	m.Cursor = newCursor
	m.updateViewportOffset()
}

// SelectCurrent marks the currently cursor-positioned issue as selected
func (m *IssueList) SelectCurrent() {
	if len(m.Issues) == 0 {
		return
	}

	if m.Cursor >= 0 && m.Cursor < len(m.Issues) {
		issue := m.Issues[m.Cursor]
		m.Selected = &issue
	}
}

// SetViewport sets the viewport height
func (m *IssueList) SetViewport(height int) {
	m.ViewportHeight = height
	m.updateViewportOffset()
}

// VisibleRange returns the start and end indices of visible issues
func (m *IssueList) VisibleRange() (start, end int) {
	if m.ViewportHeight <= 0 {
		return 0, len(m.Issues)
	}

	start = m.ViewportOffset
	end = start + m.ViewportHeight

	if end > len(m.Issues) {
		end = len(m.Issues)
	}

	return start, end
}

// GetVisibleIssues returns the list of issues currently visible in the viewport
func (m *IssueList) GetVisibleIssues() []storage.Issue {
	start, end := m.VisibleRange()
	if start >= len(m.Issues) {
		return []storage.Issue{}
	}

	return m.Issues[start:end]
}

// updateViewportOffset adjusts the viewport offset to keep the cursor visible
func (m *IssueList) updateViewportOffset() {
	if m.ViewportHeight <= 0 || len(m.Issues) == 0 {
		return
	}

	// If cursor is above viewport, scroll up
	if m.Cursor < m.ViewportOffset {
		m.ViewportOffset = m.Cursor
	}

	// If cursor is below viewport, scroll down
	if m.Cursor >= m.ViewportOffset+m.ViewportHeight {
		m.ViewportOffset = m.Cursor - m.ViewportHeight + 1
	}

	// Ensure offset doesn't go below 0
	if m.ViewportOffset < 0 {
		m.ViewportOffset = 0
	}

	// Ensure offset doesn't go past the end
	maxOffset := len(m.Issues) - m.ViewportHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.ViewportOffset > maxOffset {
		m.ViewportOffset = maxOffset
	}
}

// SetSort sets the sort field and order, re-sorts the issues
func (m *IssueList) SetSort(field string, descending bool) {
	m.SortField = field
	m.SortDescending = descending
	m.resortIssues()
}

// CycleSortField cycles to the next sort field
func (m *IssueList) CycleSortField() {
	m.SortField = sort.CycleSortField(m.SortField)
	m.resortIssues()
}

// ToggleSortOrder toggles between ascending and descending sort order
func (m *IssueList) ToggleSortOrder() {
	m.SortDescending = !m.SortDescending
	m.resortIssues()
}

// resortIssues re-sorts the issues based on current sort settings
func (m *IssueList) resortIssues() {
	m.Issues = sort.SortIssues(m.UnsortedIssues, m.SortField, m.SortDescending)
	// Reset cursor to avoid out of bounds
	if m.Cursor >= len(m.Issues) {
		m.Cursor = 0
		m.ViewportOffset = 0
	}
}
