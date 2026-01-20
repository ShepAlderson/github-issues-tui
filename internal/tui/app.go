package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/sort"
	"github.com/shepbook/ghissues/internal/storage"
)

// Model represents the main application state
type Model struct {
	IssueList     *IssueList
	DetailPanel   *DetailPanel
	CommentsView  *CommentsView // Comments view for drill-down
	AllComments   map[int][]storage.Comment // Cache comments by issue number
	Quitting      bool
	Width         int
	Height        int
}

// NewModel creates a new TUI model
func NewModel(issues []storage.Issue, columns []Column) Model {
	issueList := NewIssueList(issues, columns)
	model := Model{
		IssueList:   issueList,
		DetailPanel: nil,
		AllComments: make(map[int][]storage.Comment),
		Quitting:    false,
	}
	// Initialize detail panel with first issue if available
	if len(issues) > 0 {
		model.DetailPanel = NewDetailPanel(issues[0])
	}
	return model
}

// NewModelWithSort creates a new TUI model with specific sort settings
func NewModelWithSort(issues []storage.Issue, columns []Column, sortField string, sortDescending bool) Model {
	issueList := NewIssueListWithSort(issues, columns, sortField, sortDescending)
	model := Model{
		IssueList:   issueList,
		DetailPanel: nil,
		AllComments: make(map[int][]storage.Comment),
		Quitting:    false,
	}
	// Initialize detail panel with first issue if available
	if len(issueList.Issues) > 0 {
		model.DetailPanel = NewDetailPanel(issueList.Issues[0])
	}
	return model
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in comments view, handle comments-specific keybindings
		if m.CommentsView != nil {
			return m.updateCommentsView(msg)
		}

		// Otherwise handle main view keybindings
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "j", "down":
			m.IssueList.MoveCursor(1)
			m.updateDetailPanel()
			return m, nil

		case "k", "up":
			m.IssueList.MoveCursor(-1)
			m.updateDetailPanel()
			return m, nil

		case "enter":
			// Open comments view for current issue
			return m.openCommentsView()

		case " ":
			m.IssueList.SelectCurrent()
			m.updateDetailPanel()
			return m, nil

		case "s":
			// Cycle to next sort field
			m.IssueList.CycleSortField()
			return m, nil

		case "S":
			// Toggle sort order (shift+s)
			m.IssueList.ToggleSortOrder()
			return m, nil

		case "m":
			// Toggle markdown rendering in detail panel
			if m.DetailPanel != nil {
				m.DetailPanel.ToggleMarkdown()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Reserve space for header and status (3 lines)
		// Split remaining space between list and detail (60/40)
		listHeight := (msg.Height - 3) * 6 / 10
		detailHeight := (msg.Height - 3) * 4 / 10
		m.IssueList.SetViewport(listHeight)
		if m.DetailPanel != nil {
			m.DetailPanel.SetViewport(detailHeight)
		}
		if m.CommentsView != nil {
			m.CommentsView.SetViewport(m.Height - 3)
		}
		return m, nil
	}

	return m, nil
}

// updateCommentsView handles keybindings when in comments view
func (m Model) updateCommentsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		// Close comments view and return to main view
		m.CommentsView = nil
		return m, nil

	case "j", "down":
		m.CommentsView.ScrollDown()
		return m, nil

	case "k", "up":
		m.CommentsView.ScrollUp()
		return m, nil

	case "m":
		// Toggle markdown rendering in comments view
		m.CommentsView.ToggleMarkdown()
		return m, nil
	}

	return m, nil
}

// openCommentsView opens the comments view for the current issue
func (m Model) openCommentsView() (tea.Model, tea.Cmd) {
	if len(m.IssueList.Issues) == 0 {
		return m, nil
	}

	if m.IssueList.Cursor < 0 || m.IssueList.Cursor >= len(m.IssueList.Issues) {
		return m, nil
	}

	issue := m.IssueList.Issues[m.IssueList.Cursor]

	// Get comments from cache
	comments, ok := m.AllComments[issue.Number]
	if !ok {
		// No comments loaded yet, create empty slice
		comments = []storage.Comment{}
	}

	m.CommentsView = NewCommentsView(issue, comments)
	m.CommentsView.SetViewport(m.Height - 3)

	return m, nil
}

// SetComments sets the comments for a specific issue in the cache
func (m *Model) SetComments(issueNumber int, comments []storage.Comment) {
	if m.AllComments == nil {
		m.AllComments = make(map[int][]storage.Comment)
	}
	m.AllComments[issueNumber] = comments
}

// View renders the UI
func (m Model) View() string {
	if m.Quitting {
		return "Goodbye!\n"
	}

	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// If in comments view, render comments view
	if m.CommentsView != nil {
		return m.renderCommentsView()
	}

	// Build header
	header := m.renderHeader()

	// Build main content (split layout)
	content := m.renderSplitLayout()

	// Build status bar
	status := m.renderStatusBar()

	// Combine all parts
	return header + "\n" + content + "\n" + status
}

// renderHeader renders the column headers
func (m Model) renderHeader() string {
	if len(m.IssueList.Columns) == 0 {
		return ""
	}

	var parts []string
	for _, col := range m.IssueList.Columns {
		if col.Width > 0 {
			parts = append(parts, lipgloss.NewStyle().Width(col.Width).Render(col.Title))
		} else {
			parts = append(parts, col.Title)
		}
	}

	header := strings.Join(parts, "  ")

	// Add separator line
	separator := strings.Repeat("─", m.Width)

	style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("blue"))
	return style.Render(header) + "\n" + lipgloss.NewStyle().Faint(true).Render(separator)
}

// renderIssueList renders the visible issues
func (m Model) renderIssueList() string {
	if len(m.IssueList.Issues) == 0 {
		return "No issues found. Run 'ghissues sync' to fetch issues."
	}

	visibleIssues := m.IssueList.GetVisibleIssues()
	var lines []string

	visibleStart := m.IssueList.ViewportOffset
	for i, issue := range visibleIssues {
		globalIndex := visibleStart + i
		isCursor := globalIndex == m.IssueList.Cursor
		line := m.renderIssueRow(issue, isCursor)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderIssueRow renders a single issue row
func (m Model) renderIssueRow(issue storage.Issue, isCursor bool) string {
	if len(m.IssueList.Columns) == 0 {
		return ""
	}

	var parts []string
	for _, col := range m.IssueList.Columns {
		value := GetColumnValue(issue, col.Name)

		if col.Width > 0 {
			parts = append(parts, lipgloss.NewStyle().Width(col.Width).Render(value))
		} else {
			// For flexible width columns, truncate if needed
			maxWidth := m.Width - sumFixedWidths(m.IssueList.Columns)
			if len(value) > maxWidth && maxWidth > 0 {
				value = value[:maxWidth-3] + "..."
			}
			parts = append(parts, value)
		}
	}

	row := strings.Join(parts, "  ")

	if isCursor {
		row = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("white")).
			Render(row)
	}

	return row
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	issueCount := len(m.IssueList.Issues)
	selectedInfo := ""

	if m.IssueList.Selected != nil {
		selectedInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("green")).
			Render(" | Selected: #" + formatNumber(m.IssueList.Selected.Number))
	}

	// Build sort info
	sortOrder := "▼"
	if m.IssueList.SortDescending {
		sortOrder = "▼" // Descending
	} else {
		sortOrder = "▲" // Ascending
	}
	sortInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("yellow")).
		Render(" | Sort: " + sort.GetSortFieldLabel(m.IssueList.SortField) + sortOrder)

	markdownHint := ""
	if m.DetailPanel != nil {
		mode := "raw"
		if m.DetailPanel.RenderMarkdown {
			mode = "rendered"
		}
		markdownHint = " | m: toggle markdown (" + mode + ")"
	}

	status := lipgloss.NewStyle().
		Faint(true).
		Render("Issues: " + formatNumber(issueCount) +
			" | ↑↓/jk: navigate | s: sort field | S: sort order | Enter: comments | Space: select | q: quit" +
			markdownHint + selectedInfo + sortInfo)

	return status
}

// renderCommentsView renders the comments view
func (m Model) renderCommentsView() string {
	if m.CommentsView == nil {
		return "Error: Comments view is nil"
	}

	// Get the full view content
	content := m.CommentsView.View()

	// Build status bar for comments view
	status := m.renderCommentsStatusBar()

	return content + "\n" + status
}

// renderCommentsStatusBar renders the status bar for comments view
func (m Model) renderCommentsStatusBar() string {
	commentCount := len(m.CommentsView.Comments)

	mode := "raw"
	if m.CommentsView.RenderMarkdown {
		mode = "rendered"
	}

	status := lipgloss.NewStyle().
		Faint(true).
		Render(fmt.Sprintf("Comments: %d | ↑↓/jk: scroll | m: toggle markdown (%s) | Esc/q: back to list", commentCount, mode))

	return status
}

// sumFixedWidths calculates the total width of fixed-width columns
func sumFixedWidths(columns []Column) int {
	total := 0
	for _, col := range columns {
		if col.Width > 0 {
			total += col.Width
		}
	}
	// Add spacing between columns (2 spaces per column separator)
	if len(columns) > 0 {
		total += (len(columns) - 1) * 2
	}
	return total
}

// updateDetailPanel updates the detail panel with the currently selected issue
func (m *Model) updateDetailPanel() {
	if len(m.IssueList.Issues) == 0 {
		m.DetailPanel = nil
		return
	}

	if m.IssueList.Cursor >= 0 && m.IssueList.Cursor < len(m.IssueList.Issues) {
		issue := m.IssueList.Issues[m.IssueList.Cursor]
		m.DetailPanel = NewDetailPanel(issue)
	}
}

// renderSplitLayout renders the split layout with list and detail panels
func (m Model) renderSplitLayout() string {
	if len(m.IssueList.Issues) == 0 {
		return "No issues found. Run 'ghissues sync' to fetch issues."
	}

	// Split width: 60% for list, 40% for detail
	listWidth := m.Width * 6 / 10
	detailWidth := m.Width - listWidth - 1 // -1 for separator

	// Render issue list
	listView := m.renderIssueList()

	// Render detail panel
	detailView := ""
	if m.DetailPanel != nil {
		detailView = m.DetailPanel.View()
	} else {
		detailView = "No issue selected"
	}

	// Combine with vertical separator
	leftPanel := lipgloss.NewStyle().Width(listWidth).Height(m.Height - 3).Render(listView)
	rightPanel := lipgloss.NewStyle().Width(detailWidth).Height(m.Height - 3).Render(detailView)
	separator := lipgloss.NewStyle().Faint(true).Render("│")

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, separator, rightPanel)
}

