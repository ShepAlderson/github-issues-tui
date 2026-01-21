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

// RefreshProgress contains progress information during a refresh operation
type RefreshProgress struct {
	Phase   string // "issues" or "comments"
	Current int
	Total   int
}

// RefreshProgressMsg is sent when refresh progress updates
type RefreshProgressMsg struct {
	Progress RefreshProgress
}

// RefreshDoneMsg is sent when refresh completes successfully
type RefreshDoneMsg struct {
	Issues []github.Issue
}

// RefreshErrorMsg is sent when refresh fails
type RefreshErrorMsg struct {
	Err error
}

// RefreshStartMsg is sent to initiate a refresh (used by Init for auto-refresh)
type RefreshStartMsg struct{}

// CriticalErrorMsg is sent when a critical error occurs that requires user acknowledgment
type CriticalErrorMsg struct {
	Err      error  // The underlying error
	Title    string // Title for the error modal (e.g., "Authentication Error")
	Guidance string // Optional actionable guidance for resolving the error
}

// Model represents the TUI application state
type Model struct {
	issues           []github.Issue
	comments         []github.Comment
	columns          []string
	cursor           int
	width            int
	height           int
	sortField        config.SortField
	sortOrder        config.SortOrder
	sortChanged      bool // Track if sort was changed during session
	rawMarkdown      bool // Toggle between raw and rendered markdown
	detailScrollY    int  // Scroll offset for detail panel
	commentsScrollY  int  // Scroll offset for comments view
	inCommentsView   bool // Whether we're in the comments view
	glamourRenderer  *glamour.TermRenderer

	// Refresh state
	isRefreshing    bool            // Whether a refresh is in progress
	refreshProgress RefreshProgress // Current refresh progress
	refreshError    string          // Last refresh error message
	refreshFunc     func() tea.Msg  // Function to call to perform refresh

	// Error modal state (for critical errors)
	showErrorModal    bool   // Whether the error modal is visible
	errorModalTitle   string // Title of the error modal
	errorModalMessage string // Error message to display
	errorModalGuidance string // Optional actionable guidance
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
		// Handle error modal first - block most keys when modal is shown
		if m.showErrorModal {
			switch {
			case msg.Type == tea.KeyCtrlC:
				return m, tea.Quit
			case msg.Type == tea.KeyEscape, msg.Type == tea.KeyEnter:
				// Dismiss the error modal
				m.showErrorModal = false
				m.errorModalTitle = ""
				m.errorModalMessage = ""
				m.errorModalGuidance = ""
			case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'q':
				// Dismiss modal with 'q' as well
				m.showErrorModal = false
				m.errorModalTitle = ""
				m.errorModalMessage = ""
				m.errorModalGuidance = ""
			}
			// Block all other keys while modal is shown
			return m, nil
		}

		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyEscape:
			// Exit comments view if in it
			if m.inCommentsView {
				m.inCommentsView = false
				m.commentsScrollY = 0
			}
		case msg.Type == tea.KeyEnter:
			// Open comments view for selected issue
			if len(m.issues) > 0 && !m.inCommentsView {
				m.inCommentsView = true
				m.commentsScrollY = 0
			}
		case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'j'):
			if !m.inCommentsView && m.cursor < len(m.issues)-1 {
				m.cursor++
				m.detailScrollY = 0 // Reset detail scroll when changing issue
			}
		case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'k'):
			if !m.inCommentsView && m.cursor > 0 {
				m.cursor--
				m.detailScrollY = 0 // Reset detail scroll when changing issue
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'q':
			// In comments view, 'q' returns to issue list; otherwise, quit
			if m.inCommentsView {
				m.inCommentsView = false
				m.commentsScrollY = 0
			} else {
				return m, tea.Quit
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 's':
			// Cycle sort field (only in issue list view)
			if !m.inCommentsView {
				m.sortField = config.NextSortField(m.sortField)
				m.sortIssues()
				m.cursor = 0 // Reset cursor after sort change
				m.sortChanged = true
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'S':
			// Toggle sort order (only in issue list view)
			if !m.inCommentsView {
				m.sortOrder = config.ToggleSortOrder(m.sortOrder)
				m.sortIssues()
				m.cursor = 0 // Reset cursor after sort change
				m.sortChanged = true
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm':
			// Toggle raw/rendered markdown
			m.rawMarkdown = !m.rawMarkdown
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'l':
			// Scroll down - either detail panel or comments view
			if m.inCommentsView {
				m.commentsScrollY++
			} else {
				m.detailScrollY++
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'h':
			// Scroll up - either detail panel or comments view
			if m.inCommentsView {
				if m.commentsScrollY > 0 {
					m.commentsScrollY--
				}
			} else {
				if m.detailScrollY > 0 {
					m.detailScrollY--
				}
			}
		case msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && (msg.Runes[0] == 'r' || msg.Runes[0] == 'R'):
			// Trigger refresh (only in issue list view, not while already refreshing)
			if !m.inCommentsView && !m.isRefreshing {
				m.isRefreshing = true
				m.refreshError = "" // Clear previous error
				m.refreshProgress = RefreshProgress{}
				if m.refreshFunc != nil {
					return m, m.refreshFunc
				}
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case RefreshProgressMsg:
		m.refreshProgress = msg.Progress
	case RefreshDoneMsg:
		m.isRefreshing = false
		m.refreshProgress = RefreshProgress{}
		// Remember the currently selected issue number
		var selectedNumber int
		if m.cursor < len(m.issues) {
			selectedNumber = m.issues[m.cursor].Number
		}
		// Update issues
		m.issues = make([]github.Issue, len(msg.Issues))
		copy(m.issues, msg.Issues)
		m.sortIssues()
		// Try to restore cursor to the same issue
		m.cursor = 0
		for i, issue := range m.issues {
			if issue.Number == selectedNumber {
				m.cursor = i
				break
			}
		}
	case RefreshErrorMsg:
		m.isRefreshing = false
		m.refreshProgress = RefreshProgress{}
		if msg.Err != nil {
			m.refreshError = msg.Err.Error()
		}
	case CriticalErrorMsg:
		// Show error modal for critical errors
		m.showErrorModal = true
		m.errorModalTitle = msg.Title
		if msg.Err != nil {
			m.errorModalMessage = msg.Err.Error()
		}
		m.errorModalGuidance = msg.Guidance
	case RefreshStartMsg:
		// Auto-refresh trigger from Init
		if !m.isRefreshing {
			m.isRefreshing = true
			m.refreshError = ""
			m.refreshProgress = RefreshProgress{}
			if m.refreshFunc != nil {
				return m, m.refreshFunc
			}
		}
	}
	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	// If error modal is shown, render it over everything
	if m.showErrorModal {
		return m.renderErrorModal()
	}

	// If in comments view, render the drill-down view instead
	if m.inCommentsView {
		return m.renderCommentsView()
	}

	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

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

	// Build status line with refresh indicator or error
	var status string
	if m.isRefreshing {
		if m.refreshProgress.Total > 0 {
			status = fmt.Sprintf("Refreshing %s: %d/%d | %d issues | %s %s | r: refresh | q: quit",
				m.refreshProgress.Phase, m.refreshProgress.Current, m.refreshProgress.Total,
				len(m.issues), m.sortField.DisplayName(), sortIndicator)
		} else {
			status = fmt.Sprintf("Refreshing... | %d issues | %s %s | r: refresh | q: quit",
				len(m.issues), m.sortField.DisplayName(), sortIndicator)
		}
		b.WriteString(statusStyle.Render(status))
	} else if m.refreshError != "" {
		// Show minor error in status bar with retry hint
		errMsg := m.refreshError
		// Truncate if too long for status bar
		maxErrLen := m.width - 30
		if maxErrLen < 20 {
			maxErrLen = 40
		}
		if len(errMsg) > maxErrLen {
			errMsg = errMsg[:maxErrLen-3] + "..."
		}
		status = fmt.Sprintf("Error: %s | r: retry | q: quit", errMsg)
		b.WriteString(errorStyle.Render(status))
	} else {
		status = fmt.Sprintf("%d issues | %s %s | s: sort | S: reverse | r: refresh | m: markdown | j/k: nav | h/l: scroll | Enter: comments | q: quit",
			len(m.issues), m.sortField.DisplayName(), sortIndicator)
		b.WriteString(statusStyle.Render(status))
	}

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

// renderCommentsView renders the full-screen comments drill-down view
func (m Model) renderCommentsView() string {
	if len(m.issues) == 0 || m.cursor >= len(m.issues) {
		return "No issue selected"
	}

	issue := m.issues[m.cursor]
	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	authorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Title: Issue number and title
	b.WriteString(titleStyle.Render(fmt.Sprintf("Comments for #%d: %s", issue.Number, issue.Title)))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("═", min(m.width, 80)))
	b.WriteString("\n\n")

	// Handle empty comments
	if len(m.comments) == 0 {
		b.WriteString(dimStyle.Render("No comments on this issue."))
		b.WriteString("\n")
	} else {
		// Render comments
		var allCommentLines []string

		for i, comment := range m.comments {
			// Comment header: author and date
			createdAt, _ := time.Parse(time.RFC3339, comment.CreatedAt)
			header := fmt.Sprintf("%s  %s",
				authorStyle.Render(comment.Author.Login),
				dimStyle.Render(createdAt.Format("2006-01-02 15:04")))
			allCommentLines = append(allCommentLines, header)

			// Comment body
			bodyContent := m.renderBody(comment.Body)
			bodyLines := strings.Split(bodyContent, "\n")
			allCommentLines = append(allCommentLines, bodyLines...)

			// Add separator between comments (except after last one)
			if i < len(m.comments)-1 {
				allCommentLines = append(allCommentLines, "")
				allCommentLines = append(allCommentLines, headerStyle.Render(strings.Repeat("─", min(m.width-4, 60))))
				allCommentLines = append(allCommentLines, "")
			}
		}

		// Apply scroll offset
		visibleHeight := m.height - 8 // Account for header and status bar
		if visibleHeight < 5 {
			visibleHeight = 10
		}

		startLine := m.commentsScrollY
		if startLine >= len(allCommentLines) {
			startLine = max(0, len(allCommentLines)-1)
		}
		endLine := min(startLine+visibleHeight, len(allCommentLines))

		for i := startLine; i < endLine; i++ {
			b.WriteString(allCommentLines[i])
			b.WriteString("\n")
		}
	}

	// Status bar
	b.WriteString("\n")
	status := fmt.Sprintf("%d comments | m: toggle markdown | h/l: scroll | Esc/q: back",
		len(m.comments))
	b.WriteString(statusStyle.Render(status))

	return b.String()
}

// renderErrorModal renders a modal dialog for critical errors
func (m Model) renderErrorModal() string {
	var b strings.Builder

	// Styles for the error modal
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	guidanceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Italic(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Calculate modal dimensions
	modalWidth := min(m.width-4, 70)
	if modalWidth < 40 {
		modalWidth = 40
	}

	// Calculate padding for centering
	leftPadding := (m.width - modalWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	pad := strings.Repeat(" ", leftPadding)

	// Top spacing for vertical centering
	topPadding := (m.height - 12) / 2
	if topPadding < 2 {
		topPadding = 2
	}
	for i := 0; i < topPadding; i++ {
		b.WriteString("\n")
	}

	// Top border
	topBorder := "╔" + strings.Repeat("═", modalWidth-2) + "╗"
	b.WriteString(pad + borderStyle.Render(topBorder) + "\n")

	// Title line
	b.WriteString(pad + borderStyle.Render("║") + " " + titleStyle.Render(m.errorModalTitle) + strings.Repeat(" ", max(0, modalWidth-4-len(m.errorModalTitle))) + " " + borderStyle.Render("║") + "\n")

	// Separator line
	sepLine := "║" + strings.Repeat("─", modalWidth-2) + "║"
	b.WriteString(pad + borderStyle.Render(sepLine) + "\n")

	// Message - wrap to fit modal width
	msgWidth := modalWidth - 4
	msgLines := wrapText(m.errorModalMessage, msgWidth)
	for _, line := range msgLines {
		paddedLine := line + strings.Repeat(" ", max(0, msgWidth-len(line)))
		b.WriteString(pad + borderStyle.Render("║") + " " + messageStyle.Render(paddedLine) + " " + borderStyle.Render("║") + "\n")
	}

	// Empty line before guidance
	emptyLine := strings.Repeat(" ", modalWidth-2)
	b.WriteString(pad + borderStyle.Render("║") + emptyLine + borderStyle.Render("║") + "\n")

	// Guidance if present
	if m.errorModalGuidance != "" {
		guidanceLines := wrapText(m.errorModalGuidance, msgWidth)
		for _, line := range guidanceLines {
			paddedLine := line + strings.Repeat(" ", max(0, msgWidth-len(line)))
			b.WriteString(pad + borderStyle.Render("║") + " " + guidanceStyle.Render(paddedLine) + " " + borderStyle.Render("║") + "\n")
		}
		b.WriteString(pad + borderStyle.Render("║") + emptyLine + borderStyle.Render("║") + "\n")
	}

	// Instructions line
	instructions := "Press Enter or Esc to dismiss"
	instrPadLen := modalWidth - 4 - len(instructions)
	if instrPadLen < 0 {
		instrPadLen = 0
	}
	instrLine := strings.Repeat(" ", instrPadLen/2) + instructions + strings.Repeat(" ", instrPadLen-instrPadLen/2)
	b.WriteString(pad + borderStyle.Render("║") + " " + dimStyle.Render(instrLine) + " " + borderStyle.Render("║") + "\n")

	// Bottom border
	bottomBorder := "╚" + strings.Repeat("═", modalWidth-2) + "╝"
	b.WriteString(pad + borderStyle.Render(bottomBorder) + "\n")

	return b.String()
}

// wrapText wraps text to fit within a specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
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

// SetComments sets the comments for the currently selected issue
func (m *Model) SetComments(comments []github.Comment) {
	m.comments = comments
	m.commentsScrollY = 0 // Reset scroll when setting new comments
}

// GetComments returns the current comments
func (m Model) GetComments() []github.Comment {
	return m.comments
}

// CommentsScrollOffset returns the current scroll offset for the comments view
func (m Model) CommentsScrollOffset() int {
	return m.commentsScrollY
}

// IsRefreshing returns whether a refresh is in progress
func (m Model) IsRefreshing() bool {
	return m.isRefreshing
}

// GetRefreshProgress returns the current refresh progress
func (m Model) GetRefreshProgress() RefreshProgress {
	return m.refreshProgress
}

// GetRefreshError returns the last refresh error message
func (m Model) GetRefreshError() string {
	return m.refreshError
}

// SetRefreshFunc sets the function to be called when refresh is triggered
func (m *Model) SetRefreshFunc(fn func() tea.Msg) {
	m.refreshFunc = fn
}

// HasErrorModal returns whether the error modal is visible
func (m Model) HasErrorModal() bool {
	return m.showErrorModal
}

// GetErrorModalTitle returns the title of the error modal
func (m Model) GetErrorModalTitle() string {
	return m.errorModalTitle
}

// GetErrorModalMessage returns the error message in the modal
func (m Model) GetErrorModalMessage() string {
	return m.errorModalMessage
}

// GetErrorModalGuidance returns the guidance text for the error modal
func (m Model) GetErrorModalGuidance() string {
	return m.errorModalGuidance
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
