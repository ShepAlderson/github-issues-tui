package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	issues          []github.Issue
	columns         []string
	cursor          int
	width           int
	height          int
	sortField       config.SortField
	sortOrder       config.SortOrder
	sortChanged     bool // Track if sort was changed during session
	rawMarkdown     bool // Toggle between raw and rendered markdown
	detailScrollY   int  // Scroll offset for detail panel
	inCommentsView  bool // Whether we're in the comments view
	glamourRenderer *glamour.TermRenderer
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

	// Create a glamour renderer for markdown rendering
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(60),
	)

	m := Model{
		issues:          make([]github.Issue, len(issues)),
		columns:         columns,
		cursor:          0,
		sortField:       sortField,
		sortOrder:       sortOrder,
		glamourRenderer: renderer,
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
		case msg.Type == tea.KeyEscape:
			// Exit comments view if in it
			if m.inCommentsView {
				m.inCommentsView = false
			}
		case msg.Type == tea.KeyEnter:
			// Open comments view for selected issue
			if len(m.issues) > 0 {
				m.inCommentsView = true
			}
		case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'j'):
			if m.cursor < len(m.issues)-1 {
				m.cursor++
				m.detailScrollY = 0 // Reset detail scroll when changing issue
			}
		case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'k'):
			if m.cursor > 0 {
				m.cursor--
				m.detailScrollY = 0 // Reset detail scroll when changing issue
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
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm':
			// Toggle raw/rendered markdown
			m.rawMarkdown = !m.rawMarkdown
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'l':
			// Scroll detail panel down
			m.detailScrollY++
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'h':
			// Scroll detail panel up
			if m.detailScrollY > 0 {
				m.detailScrollY--
			}
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
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Title
	title := titleStyle.Render("GitHub Issues")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Handle empty state
	if len(m.issues) == 0 {
		b.WriteString("No issues found. Run 'ghissues sync' to fetch issues.\n")
	} else {
		// Calculate the panel widths (50/50 split, or use available space)
		listWidth := m.width / 2
		detailWidth := m.width - listWidth - 3 // 3 for separator

		// Render issue list panel
		listPanel := m.renderIssueListPanel(listWidth)

		// Render detail panel
		detailPanel := m.renderDetailPanel(detailWidth)

		// Combine panels side by side
		listLines := strings.Split(listPanel, "\n")
		detailLines := strings.Split(detailPanel, "\n")

		// Get the max lines to display
		maxLines := max(len(listLines), len(detailLines))
		contentHeight := m.height - 6 // Account for title, status
		if contentHeight < 5 {
			contentHeight = 15
		}
		maxLines = min(maxLines, contentHeight)

		for i := 0; i < maxLines; i++ {
			listLine := ""
			if i < len(listLines) {
				listLine = listLines[i]
			}
			// Pad list line to width
			listLine = padToWidth(listLine, listWidth)

			detailLine := ""
			if i < len(detailLines) {
				detailLine = detailLines[i]
			}

			b.WriteString(listLine)
			b.WriteString(" │ ")
			b.WriteString(detailLine)
			b.WriteString("\n")
		}
	}

	// Status bar
	b.WriteString("\n")
	sortIndicator := "↓"
	if m.sortOrder == config.SortAsc {
		sortIndicator = "↑"
	}
	status := fmt.Sprintf("%d issues | %s %s | s: sort | S: reverse | m: markdown | j/k: nav | h/l: scroll | Enter: comments | q: quit",
		len(m.issues), m.sortField.DisplayName(), sortIndicator)
	b.WriteString(statusStyle.Render(status))

	return b.String()
}

// renderIssueListPanel renders the left panel with the issue list
func (m Model) renderIssueListPanel(width int) string {
	var b strings.Builder

	selectedStyle := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("238"))
	normalStyle := lipgloss.NewStyle()
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))

	// Render header
	header := fmt.Sprintf("  %-6s %-*s", "#", width-10, "Title")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	// Calculate visible height
	visibleHeight := m.height - 8
	if visibleHeight < 1 {
		visibleHeight = 10
	}

	startIdx := 0
	if m.cursor >= visibleHeight {
		startIdx = m.cursor - visibleHeight + 1
	}

	endIdx := min(startIdx+visibleHeight, len(m.issues))

	for i := startIdx; i < endIdx; i++ {
		issue := m.issues[i]
		titleWidth := width - 12
		title := issue.Title
		if len(title) > titleWidth {
			title = title[:titleWidth-3] + "..."
		}
		row := fmt.Sprintf("#%-5d %-*s", issue.Number, titleWidth, title)

		if i == m.cursor {
			b.WriteString(selectedStyle.Render("> " + row))
		} else {
			b.WriteString(normalStyle.Render("  " + row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderDetailPanel renders the right panel with issue details
func (m Model) renderDetailPanel(width int) string {
	if len(m.issues) == 0 || m.cursor >= len(m.issues) {
		return "No issue selected"
	}

	issue := m.issues[m.cursor]
	var b strings.Builder

	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Issue header: number and title
	b.WriteString(headerStyle.Render(fmt.Sprintf("#%d", issue.Number)))
	b.WriteString(" ")
	title := issue.Title
	if len(title) > width-10 {
		title = title[:width-13] + "..."
	}
	b.WriteString(headerStyle.Render(title))
	b.WriteString("\n")

	// Author
	b.WriteString(dimStyle.Render("Author: "))
	b.WriteString(issue.Author.Login)
	b.WriteString("\n")

	// Dates
	createdAt, _ := time.Parse(time.RFC3339, issue.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, issue.UpdatedAt)
	b.WriteString(dimStyle.Render("Created: "))
	b.WriteString(createdAt.Format("2006-01-02"))
	b.WriteString("  ")
	b.WriteString(dimStyle.Render("Updated: "))
	b.WriteString(updatedAt.Format("2006-01-02"))
	b.WriteString("\n")

	// Labels
	if len(issue.Labels) > 0 {
		b.WriteString(dimStyle.Render("Labels: "))
		var labels []string
		for _, label := range issue.Labels {
			labels = append(labels, labelStyle.Render(label.Name))
		}
		b.WriteString(strings.Join(labels, ", "))
		b.WriteString("\n")
	}

	// Assignees
	if len(issue.Assignees) > 0 {
		b.WriteString(dimStyle.Render("Assignees: "))
		var assignees []string
		for _, a := range issue.Assignees {
			assignees = append(assignees, a.Login)
		}
		b.WriteString(strings.Join(assignees, ", "))
		b.WriteString("\n")
	}

	// Separator
	b.WriteString(strings.Repeat("─", min(width, 60)))
	b.WriteString("\n")

	// Body
	if issue.Body != "" {
		bodyContent := m.renderBody(issue.Body)
		bodyLines := strings.Split(bodyContent, "\n")

		// Apply scroll offset
		startLine := m.detailScrollY
		if startLine >= len(bodyLines) {
			startLine = max(0, len(bodyLines)-1)
		}

		// Calculate visible body height
		visibleHeight := m.height - 15
		if visibleHeight < 3 {
			visibleHeight = 5
		}

		endLine := min(startLine+visibleHeight, len(bodyLines))
		for i := startLine; i < endLine; i++ {
			b.WriteString(bodyLines[i])
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderBody renders the issue body, either raw or with glamour
func (m Model) renderBody(body string) string {
	if m.rawMarkdown {
		return body
	}

	// Render markdown with glamour
	if m.glamourRenderer != nil {
		rendered, err := m.glamourRenderer.Render(body)
		if err == nil {
			return strings.TrimSpace(rendered)
		}
	}

	// Fallback to raw if rendering fails
	return body
}

// padToWidth pads a string to a specific width, accounting for ANSI codes
func padToWidth(s string, width int) string {
	// Get visual width (lipgloss handles ANSI codes)
	visualWidth := lipgloss.Width(s)
	if visualWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visualWidth)
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

// IsRawMarkdown returns whether the detail panel is showing raw markdown
func (m Model) IsRawMarkdown() bool {
	return m.rawMarkdown
}

// DetailScrollOffset returns the current scroll offset for the detail panel
func (m Model) DetailScrollOffset() int {
	return m.detailScrollY
}

// InCommentsView returns whether the comments view is active
func (m Model) InCommentsView() bool {
	return m.inCommentsView
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
