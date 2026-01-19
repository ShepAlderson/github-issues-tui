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
	issues  []*sync.Issue
	columns []string
	cursor  int
	width   int
	height  int
}

// NewModel creates a new TUI model
func NewModel(issues []*sync.Issue, columns []string) Model {
	return Model{
		issues:  issues,
		columns: columns,
		cursor:  0,
	}
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
	b.WriteString(footerStyle.Render("j/k, ↑/↓: navigate • q: quit"))

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
	if total == 0 {
		return "No issues"
	}
	return fmt.Sprintf("Issue %d of %d", current, total)
}

// SelectedIssue returns the currently selected issue
func (m Model) SelectedIssue() *sync.Issue {
	if len(m.issues) == 0 {
		return nil
	}
	return m.issues[m.cursor]
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
