package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/github"
)

// DefaultColumns returns the default columns to display
func DefaultColumns() []string {
	return []string{"number", "title", "author", "date", "comments"}
}

// Model represents the TUI application state
type Model struct {
	issues      []github.Issue
	columns     []string
	cursor      int
	width       int
	height      int
	sortField   config.SortField
	sortOrder   config.SortOrder
	sortChanged bool // Track if sort was changed during session
}

// NewModel creates a new TUI model with the given issues and columns
// Uses default sort: most recently updated first (updated descending)
func NewModel(issues []github.Issue, columns []string) Model {
	sortField, sortOrder := config.DefaultSortConfig()
	return NewModelWithSort(issues, columns, sortField, sortOrder)
}

// NewModelWithSort creates a new TUI model with the given issues, columns, and sort options
func NewModelWithSort(issues []github.Issue, columns []string, sortField config.SortField, sortOrder config.SortOrder) Model {
	if columns == nil {
		columns = DefaultColumns()
	}
	if sortField == "" {
		sortField, _ = config.DefaultSortConfig()
	}
	if sortOrder == "" {
		_, sortOrder = config.DefaultSortConfig()
	}

	m := Model{
		issues:    make([]github.Issue, len(issues)),
		columns:   columns,
		cursor:    0,
		sortField: sortField,
		sortOrder: sortOrder,
	}

	// Copy issues to avoid modifying the original slice
	copy(m.issues, issues)

	// Apply initial sort
	m.sortIssues()

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'j'):
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}
		case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'k'):
			if m.cursor > 0 {
				m.cursor--
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'q':
			return m, tea.Quit
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 's':
			// Cycle sort field
			m.sortField = config.NextSortField(m.sortField)
			m.sortIssues()
			m.cursor = 0 // Reset cursor after sort change
			m.sortChanged = true
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'S':
			// Toggle sort order
			m.sortOrder = config.ToggleSortOrder(m.sortOrder)
			m.sortIssues()
			m.cursor = 0 // Reset cursor after sort change
			m.sortChanged = true
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("238"))
	normalStyle := lipgloss.NewStyle()
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))

	// Title
	title := titleStyle.Render("GitHub Issues")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Handle empty state
	if len(m.issues) == 0 {
		b.WriteString("No issues found. Run 'ghissues sync' to fetch issues.\n")
	} else {
		// Calculate column widths
		colWidths := m.calculateColumnWidths()

		// Render header
		header := m.renderHeader(colWidths, headerStyle)
		b.WriteString(header)
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", min(m.width, 120)))
		b.WriteString("\n")

		// Render issue list
		visibleHeight := m.height - 6 // Account for title, header, separator, status
		if visibleHeight < 1 {
			visibleHeight = 10
		}

		startIdx := 0
		if m.cursor >= visibleHeight {
			startIdx = m.cursor - visibleHeight + 1
		}

		endIdx := startIdx + visibleHeight
		if endIdx > len(m.issues) {
			endIdx = len(m.issues)
		}

		for i := startIdx; i < endIdx; i++ {
			issue := m.issues[i]
			row := m.renderIssueRow(issue, colWidths)

			if i == m.cursor {
				b.WriteString(selectedStyle.Render("> " + row))
			} else {
				b.WriteString(normalStyle.Render("  " + row))
			}
			b.WriteString("\n")
		}
	}

	// Status bar
	b.WriteString("\n")
	sortIndicator := "↓"
	if m.sortOrder == config.SortAsc {
		sortIndicator = "↑"
	}
	status := fmt.Sprintf("%d issues | %s %s | s: sort | S: reverse | j/k: navigate | q: quit",
		len(m.issues), m.sortField.DisplayName(), sortIndicator)
	b.WriteString(statusStyle.Render(status))

	return b.String()
}

// calculateColumnWidths calculates the width for each column
func (m Model) calculateColumnWidths() map[string]int {
	widths := map[string]int{
		"number":   6,
		"title":    40,
		"author":   15,
		"date":     12,
		"comments": 8,
	}

	// Adjust title width based on available space
	totalFixed := 0
	for col, w := range widths {
		if col != "title" {
			totalFixed += w + 2 // +2 for separator
		}
	}

	availableWidth := m.width - totalFixed - 4 // -4 for padding and cursor
	if availableWidth > 20 {
		widths["title"] = min(availableWidth, 80)
	}

	return widths
}

// renderHeader renders the column header row
func (m Model) renderHeader(widths map[string]int, style lipgloss.Style) string {
	var parts []string
	for _, col := range m.columns {
		width := widths[col]
		header := columnHeader(col)
		parts = append(parts, style.Render(padOrTruncate(header, width)))
	}
	return "  " + strings.Join(parts, " │ ")
}

// renderIssueRow renders a single issue row
func (m Model) renderIssueRow(issue github.Issue, widths map[string]int) string {
	var parts []string
	for _, col := range m.columns {
		width := widths[col]
		value := m.getColumnValue(issue, col)
		parts = append(parts, padOrTruncate(value, width))
	}
	return strings.Join(parts, " │ ")
}

// getColumnValue returns the display value for a column
func (m Model) getColumnValue(issue github.Issue, col string) string {
	switch col {
	case "number":
		return fmt.Sprintf("#%d", issue.Number)
	case "title":
		return issue.Title
	case "author":
		return issue.Author.Login
	case "date":
		t, err := time.Parse(time.RFC3339, issue.UpdatedAt)
		if err != nil {
			return issue.UpdatedAt
		}
		return t.Format("2006-01-02")
	case "comments":
		return fmt.Sprintf("%d", issue.CommentCount)
	default:
		return ""
	}
}

// columnHeader returns the header text for a column
func columnHeader(col string) string {
	switch col {
	case "number":
		return "#"
	case "title":
		return "Title"
	case "author":
		return "Author"
	case "date":
		return "Updated"
	case "comments":
		return "Comments"
	default:
		return col
	}
}

// padOrTruncate pads or truncates a string to the given width
func padOrTruncate(s string, width int) string {
	if len(s) > width {
		if width > 3 {
			return s[:width-3] + "..."
		}
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

// SelectedIssue returns the currently selected issue, or nil if no issues
func (m Model) SelectedIssue() *github.Issue {
	if len(m.issues) == 0 || m.cursor >= len(m.issues) {
		return nil
	}
	return &m.issues[m.cursor]
}

// IssueCount returns the total number of issues
func (m Model) IssueCount() int {
	return len(m.issues)
}

// SetWindowSize sets the terminal window size
func (m *Model) SetWindowSize(width, height int) {
	m.width = width
	m.height = height
}

// GetSortConfig returns the current sort field and order
func (m Model) GetSortConfig() (config.SortField, config.SortOrder) {
	return m.sortField, m.sortOrder
}

// SortChanged returns true if the sort settings were changed during the session
func (m Model) SortChanged() bool {
	return m.sortChanged
}

// sortIssues sorts the issues based on the current sort field and order
func (m *Model) sortIssues() {
	if len(m.issues) == 0 {
		return
	}

	sort.Slice(m.issues, func(i, j int) bool {
		var less bool

		switch m.sortField {
		case config.SortByUpdated:
			ti, _ := m.issues[i].UpdatedAtTime()
			tj, _ := m.issues[j].UpdatedAtTime()
			less = ti.Before(tj)
		case config.SortByCreated:
			ti, _ := m.issues[i].CreatedAtTime()
			tj, _ := m.issues[j].CreatedAtTime()
			less = ti.Before(tj)
		case config.SortByNumber:
			less = m.issues[i].Number < m.issues[j].Number
		case config.SortByComments:
			less = m.issues[i].CommentCount < m.issues[j].CommentCount
		default:
			// Default to updated date
			ti, _ := m.issues[i].UpdatedAtTime()
			tj, _ := m.issues[j].UpdatedAtTime()
			less = ti.Before(tj)
		}

		// Descending order reverses the comparison
		if m.sortOrder == config.SortDesc {
			return !less
		}
		return less
	})
}
