package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/sync"
)

// Model represents the TUI state
type Model struct {
	issues            []*sync.Issue
	columns           []string
	cursor            int
	width             int
	height            int
	sortBy            string // "updated", "created", "number", "comments"
	sortAscending     bool   // true for ascending, false for descending
	showRawMarkdown   bool   // true to show raw markdown, false for rendered
	detailScrollOffset int    // scroll offset for detail panel
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
					m.detailScrollOffset = 0 // Reset scroll when changing issue
				}
			case "k":
				if m.cursor > 0 {
					m.cursor--
					m.detailScrollOffset = 0 // Reset scroll when changing issue
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
			case "m":
				// Toggle between raw and rendered markdown
				m.showRawMarkdown = !m.showRawMarkdown
			}

		case tea.KeyDown:
			if m.cursor < len(m.issues)-1 {
				m.cursor++
				m.detailScrollOffset = 0 // Reset scroll when changing issue
			}

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
				m.detailScrollOffset = 0 // Reset scroll when changing issue
			}

		case tea.KeyPgDown:
			// Scroll down in detail panel
			m.detailScrollOffset += 10
			// Max scroll is enforced in rendering

		case tea.KeyPgUp:
			// Scroll up in detail panel
			m.detailScrollOffset -= 10
			if m.detailScrollOffset < 0 {
				m.detailScrollOffset = 0
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

	// If no dimensions set yet, render simple list (will be set on first WindowSizeMsg)
	if m.width == 0 || m.height == 0 {
		return m.renderSimpleList()
	}

	// Render split-pane layout: list on left, detail on right
	return m.renderSplitPane()
}

// renderSimpleList renders the simple list view (before dimensions are known)
func (m Model) renderSimpleList() string {
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

// renderSplitPane renders the split-pane layout with list and detail
func (m Model) renderSplitPane() string {
	// Calculate widths for split pane (60/40 split)
	listWidth := m.width * 60 / 100
	detailWidth := m.width - listWidth - 2 // -2 for separator

	// Calculate heights (leaving room for status and footer)
	contentHeight := m.height - 5 // -5 for header, status, footer

	// Render list panel
	listPanel := m.renderListPanel(listWidth, contentHeight)

	// Render detail panel
	detailPanel := m.renderDetailPanel(detailWidth, contentHeight)

	// Combine panels side by side
	listLines := strings.Split(listPanel, "\n")
	detailLines := strings.Split(detailPanel, "\n")

	var b strings.Builder

	// Ensure both panels have same number of lines
	maxLines := len(listLines)
	if len(detailLines) > maxLines {
		maxLines = len(detailLines)
	}

	for i := 0; i < maxLines; i++ {
		// List panel
		if i < len(listLines) {
			b.WriteString(listLines[i])
		} else {
			b.WriteString(strings.Repeat(" ", listWidth))
		}

		// Separator
		b.WriteString(" │ ")

		// Detail panel
		if i < len(detailLines) {
			b.WriteString(detailLines[i])
		}

		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(statusStyle.Render(m.renderStatus()))

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("j/k, ↑/↓: navigate • PgUp/PgDn: scroll • m: toggle markdown • s: sort • S: reverse • q: quit"))

	return b.String()
}

// renderListPanel renders the left panel with issue list
func (m Model) renderListPanel(width, height int) string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render(m.renderHeader()))
	b.WriteString("\n\n")

	// Issue list (limit to height)
	linesRendered := 2 // header + blank line
	for i, issue := range m.issues {
		if linesRendered >= height {
			break
		}

		cursor := " "
		style := normalStyle
		if i == m.cursor {
			cursor = ">"
			style = selectedStyle
		}

		line := fmt.Sprintf("%s %s", cursor, m.renderIssue(issue))
		// Truncate to width
		if len(line) > width {
			line = line[:width]
		}
		b.WriteString(style.Render(line))
		b.WriteString("\n")
		linesRendered++
	}

	return b.String()
}

// renderDetailPanel renders the right panel with issue details
func (m Model) renderDetailPanel(width, height int) string {
	selected := m.SelectedIssue()
	if selected == nil {
		return detailPanelStyle.Width(width).Height(height).Render("No issue selected")
	}

	var b strings.Builder

	// Header: issue number, title, state
	b.WriteString(detailHeaderStyle.Render(fmt.Sprintf("#%d • %s", selected.Number, selected.State)))
	b.WriteString("\n")
	b.WriteString(detailTitleStyle.Render(selected.Title))
	b.WriteString("\n\n")

	// Metadata: author, dates
	b.WriteString(detailMetaStyle.Render(fmt.Sprintf("Author: %s", selected.Author)))
	b.WriteString("\n")
	b.WriteString(detailMetaStyle.Render(fmt.Sprintf("Created: %s", formatRelativeTime(selected.CreatedAt))))
	b.WriteString("\n")
	b.WriteString(detailMetaStyle.Render(fmt.Sprintf("Updated: %s", formatRelativeTime(selected.UpdatedAt))))
	b.WriteString("\n")

	// Labels (if any)
	if len(selected.Labels) > 0 {
		b.WriteString(detailMetaStyle.Render(fmt.Sprintf("Labels: %s", strings.Join(selected.Labels, ", "))))
		b.WriteString("\n")
	}

	// Assignees (if any)
	if len(selected.Assignees) > 0 {
		b.WriteString(detailMetaStyle.Render(fmt.Sprintf("Assignees: %s", strings.Join(selected.Assignees, ", "))))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Body
	bodyContent := selected.Body
	if bodyContent == "" {
		bodyContent = "(No description provided)"
	}

	// Render markdown or show raw
	if m.showRawMarkdown {
		b.WriteString(bodyContent)
	} else {
		// Use glamour to render markdown
		rendered, err := renderMarkdown(bodyContent, width)
		if err != nil {
			// Fall back to raw if rendering fails
			b.WriteString(bodyContent)
		} else {
			b.WriteString(rendered)
		}
	}

	// Split into lines for scrolling
	content := b.String()
	lines := strings.Split(content, "\n")

	// Apply scroll offset
	startLine := m.detailScrollOffset
	if startLine >= len(lines) {
		startLine = len(lines) - 1
	}
	if startLine < 0 {
		startLine = 0
	}

	endLine := startLine + height
	if endLine > len(lines) {
		endLine = len(lines)
	}

	scrolledLines := lines[startLine:endLine]

	// Truncate each line to width
	for i, line := range scrolledLines {
		if len(line) > width {
			scrolledLines[i] = line[:width]
		}
	}

	return strings.Join(scrolledLines, "\n")
}

// renderMarkdown renders markdown content using glamour
func renderMarkdown(content string, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}

	rendered, err := r.Render(content)
	if err != nil {
		return "", err
	}

	return rendered, nil
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

	detailPanelStyle = lipgloss.NewStyle().
				Padding(1)

	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170"))

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	detailMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
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
