package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/storage"
)

// Model represents the main application state
type Model struct {
	IssueList   *IssueList
	Quitting    bool
	Width       int
	Height      int
}

// NewModel creates a new TUI model
func NewModel(issues []storage.Issue, columns []Column) Model {
	issueList := NewIssueList(issues, columns)
	return Model{
		IssueList: issueList,
		Quitting:  false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "j", "down":
			m.IssueList.MoveCursor(1)
			return m, nil

		case "k", "up":
			m.IssueList.MoveCursor(-1)
			return m, nil

		case "enter", " ":
			m.IssueList.SelectCurrent()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// Reserve space for header and status (3 lines)
		m.IssueList.SetViewport(msg.Height - 3)
		return m, nil
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.Quitting {
		return "Goodbye!\n"
	}

	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Build header
	header := m.renderHeader()

	// Build issue list
	issuesView := m.renderIssueList()

	// Build status bar
	status := m.renderStatusBar()

	// Combine all parts
	return header + "\n" + issuesView + "\n" + status
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

	status := lipgloss.NewStyle().
		Faint(true).
		Render("Issues: " + formatNumber(issueCount) +
			" | ↑↓/jk: navigate | Enter: select | q: quit" +
			selectedInfo)

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
