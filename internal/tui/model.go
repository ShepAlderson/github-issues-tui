package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/sync"
)

// Model represents the TUI state
type Model struct {
	issues        []*sync.Issue
	columns       []string
	cursor        int
	width         int
	height        int
	sortBy        string // "updated", "created", "number", "comments"
	sortAscending bool   // true for ascending, false for descending
}

// NewModel creates a new TUI model
func NewModel(issues []*sync.Issue, columns []string, sortBy string, sortAscending bool) Model {
	m := Model{
		issues:        issues,
		columns:       columns,
		cursor:        0,
		sortBy:        sortBy,
		sortAscending: sortAscending,
	}
	m.sortIssues() // Apply initial sort
	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				return m, tea.Quit
			case "j":
				if m.cursor < len(m.issues)-1 {
					m.cursor++
				}
			case "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "s":
				// Cycle through sort options: updated -> created -> number -> comments -> updated
				switch m.sortBy {
				case "updated":
					m.sortBy = "created"
				case "created":
					m.sortBy = "number"
				case "number":
					m.sortBy = "comments"
				case "comments":
					m.sortBy = "updated"
				}
				m.sortIssues()
				m.cursor = 0 // Reset cursor to top after sorting
			case "S":
				// Toggle sort order
				m.sortAscending = !m.sortAscending
				m.sortIssues()
				m.cursor = 0 // Reset cursor to top after sorting
			}

		case tea.KeyDown:
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if len(m.issues) == 0 {
		return noIssuesStyle.Render("No issues found. Run 'ghissues sync' to fetch issues.")
	}

	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render(m.renderHeader()))
	b.WriteString("\n\n")

	// Issue list
	for i, issue := range m.issues {
		cursor := " "
		style := normalStyle
		if i == m.cursor {
			cursor = ">"
			style = selectedStyle
		}

		line := fmt.Sprintf("%s %s", cursor, m.renderIssue(issue))
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(m.renderStatus()))

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("j/k, ↑/↓: navigate • s: cycle sort • S: reverse order • q: quit"))

	return b.String()
}

// renderHeader renders the column headers
func (m Model) renderHeader() string {
	var parts []string
	for _, col := range m.columns {
		switch col {
		case "number":
			parts = append(parts, padRight("#", 8))
		case "title":
			parts = append(parts, padRight("Title", 50))
		case "author":
			parts = append(parts, padRight("Author", 20))
		case "date":
			parts = append(parts, padRight("Updated", 20))
		case "comments":
			parts = append(parts, padRight("Comments", 10))
		}
	}
	return strings.Join(parts, " ")
}

// renderIssue renders a single issue row
func (m Model) renderIssue(issue *sync.Issue) string {
	var parts []string
	for _, col := range m.columns {
		switch col {
		case "number":
			parts = append(parts, padRight(fmt.Sprintf("#%d", issue.Number), 8))
		case "title":
			parts = append(parts, padRight(truncate(issue.Title, 48), 50))
		case "author":
			parts = append(parts, padRight(truncate(issue.Author, 18), 20))
		case "date":
			parts = append(parts, padRight(formatRelativeTime(issue.UpdatedAt), 20))
		case "comments":
			parts = append(parts, padRight(fmt.Sprintf("%d", issue.CommentCount), 10))
		}
	}
	return strings.Join(parts, " ")
}

// renderStatus renders the status bar
func (m Model) renderStatus() string {
	total := len(m.issues)
	current := m.cursor + 1

	// Build sort description
	sortOrder := "desc"
	if m.sortAscending {
		sortOrder = "asc"
	}
	sortDesc := fmt.Sprintf("sort: %s (%s)", m.sortBy, sortOrder)

	if total == 0 {
		return fmt.Sprintf("No issues • %s", sortDesc)
	}
	return fmt.Sprintf("Issue %d of %d • %s", current, total, sortDesc)
}

// SelectedIssue returns the currently selected issue
func (m Model) SelectedIssue() *sync.Issue {
	if len(m.issues) == 0 {
		return nil
	}
	return m.issues[m.cursor]
}

// sortIssues sorts the issues slice based on sortBy and sortAscending fields
func (m *Model) sortIssues() {
	if len(m.issues) == 0 {
		return
	}

	// Use a stable sort to maintain order for equal elements
	sortFunc := func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case "updated":
			less = m.issues[i].UpdatedAt.Before(m.issues[j].UpdatedAt)
		case "created":
			less = m.issues[i].CreatedAt.Before(m.issues[j].CreatedAt)
		case "number":
			less = m.issues[i].Number < m.issues[j].Number
		case "comments":
			less = m.issues[i].CommentCount < m.issues[j].CommentCount
		default:
			// Default to updated
			less = m.issues[i].UpdatedAt.Before(m.issues[j].UpdatedAt)
		}

		// If ascending, use the comparison as-is; if descending, invert it
		if m.sortAscending {
			return less
		}
		return !less
	}

	// Sort issues in place
	for i := 0; i < len(m.issues)-1; i++ {
		for j := i + 1; j < len(m.issues); j++ {
			if sortFunc(j, i) {
				m.issues[i], m.issues[j] = m.issues[j], m.issues[i]
			}
		}
	}
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39"))

	normalStyle = lipgloss.NewStyle()

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	noIssuesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
)

// Helper functions

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatRelativeTime(t time.Time) string {
	dur := time.Since(t)
	if dur < time.Minute {
		return "just now"
	}
	if dur < time.Hour {
		mins := int(dur.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if dur < 24*time.Hour {
		hours := int(dur.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(dur.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}
	months := days / 30
	if months == 1 {
		return "1 month ago"
	}
	if months < 12 {
		return fmt.Sprintf("%d months ago", months)
	}
	years := months / 12
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}
