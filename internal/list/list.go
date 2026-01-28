package list

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shepbook/ghissues/internal/database"
)

// Config interface for accessing configuration
type Config interface {
	GetDisplayColumns() []string
	GetDefaultRepository() string
	GetSortField() string
	GetSortDescending() bool
	SaveSort(field string, descending bool) error
}

// Model represents the issue list TUI state
type Model struct {
	dbPath      string
	repo        string
	columns     []string
	issues      []database.ListIssue
	selected    int
	width       int
	height      int
	db          *sql.DB
	sortField   string
	sortDesc    bool
	sortFields  []string
	configPath  string
	// saveSort is a callback to persist sort settings
	saveSort func(field string, descending bool) error
}

// Styles for the list view
var (
	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	normalStyle = lipgloss.NewStyle()

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

)

// NewModel creates a new list model
func NewModel(cfg Config, dbPath, configPath string) Model {
	columns := validateColumns(cfg.GetDisplayColumns())
	if len(columns) == 0 {
		columns = []string{"number", "title", "author", "updated", "comments"}
	}

	// Validate sort field from config
	sortField := cfg.GetSortField()
	if sortField == "" {
		sortField = "updated"
	}

	return Model{
		dbPath:     dbPath,
		repo:       cfg.GetDefaultRepository(),
		columns:    columns,
		issues:     []database.ListIssue{},
		selected:   0,
		width:      80,
		height:     24,
		sortField:  sortField,
		sortDesc:   cfg.GetSortDescending(),
		sortFields: []string{"updated", "created", "number", "comments"},
		configPath: configPath,
		saveSort:   cfg.SaveSort,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.loadIssues()
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if m.selected > 0 {
				m.selected--
			}
		case tea.KeyDown:
			if m.selected < len(m.issues)-1 {
				m.selected++
			}
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return m, tea.Quit
			case "k":
				if m.selected > 0 {
					m.selected--
				}
			case "j":
				if m.selected < len(m.issues)-1 {
					m.selected++
				}
			case "s":
				// Cycle to next sort field
				m.cycleSortField()
				// Save sort preference and reload issues
				return m, tea.Batch(
					m.saveSortConfig(),
					m.loadIssues(),
				)
			case "S":
				// Toggle sort order
				m.sortDesc = !m.sortDesc
				// Save sort preference and reload issues
				return m, tea.Batch(
					m.saveSortConfig(),
					m.loadIssues(),
				)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case issuesLoadedMsg:
		m.issues = msg.issues
		if m.selected >= len(m.issues) {
			m.selected = 0
		}
	}

	return m, nil
}

// View renders the list UI
func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(headerStyle.Render(fmt.Sprintf("📋 %s", m.repo)))
	b.WriteString("\n\n")

	// Calculate available height for issue list
	headerLines := 3 // Title + blank line + separator
	statusLines := 2 // Status line + separator
	availableHeight := m.height - headerLines - statusLines

	if availableHeight < 5 {
		availableHeight = 10 // Minimum height
	}

	// Issue list panel
	if len(m.issues) == 0 {
		b.WriteString("  No issues found. Run 'ghissues sync' to fetch issues.\n")
	} else {
		// Render each issue
		for i, issue := range m.issues {
			if i >= availableHeight {
				break
			}

			line := m.renderIssueLine(issue, i == m.selected)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Status bar at the bottom
	b.WriteString("\n")
	orderIcon := "↓"
	if !m.sortDesc {
		orderIcon = "↑"
	}
	status := fmt.Sprintf("%d issues | sort:%s %s | j/k to navigate | s to sort | q to quit", len(m.issues), m.sortField, orderIcon)
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n")

	return b.String()
}

// renderIssueLine renders a single issue line
func (m Model) renderIssueLine(issue database.ListIssue, isSelected bool) string {
	columns := renderColumns(issue, m.columns)

	// Join columns with spacing
	line := strings.Join(columns, "  ")

	// Truncate if too long
	maxWidth := m.width - 2 // Leave room for selection indicator
	if lipgloss.Width(line) > maxWidth {
		line = lipgloss.NewStyle().MaxWidth(maxWidth).Render(line)
	}

	// Add selection indicator
	if isSelected {
		return selectedStyle.Render("> " + line)
	}
	return normalStyle.Render("  " + line)
}

// renderColumns extracts and formats the requested columns from an issue
func renderColumns(issue database.ListIssue, columns []string) []string {
	var result []string

	for _, col := range columns {
		var value string
		switch col {
		case "number":
			value = fmt.Sprintf("#%d", issue.Number)
		case "title":
			value = issue.Title
		case "author":
			value = issue.Author
		case "created":
			value = database.FormatDate(issue.CreatedAt)
		case "updated":
			value = database.FormatDate(issue.UpdatedAt)
		case "comments":
			if issue.CommentCount > 0 {
				value = fmt.Sprintf("💬 %d", issue.CommentCount)
			} else {
				value = ""
			}
		case "state":
			value = issue.State
		default:
			continue // Skip unknown columns
		}
		result = append(result, value)
	}

	return result
}

// validateColumns filters out invalid column names
func validateColumns(columns []string) []string {
	valid := map[string]bool{
		"number":   true,
		"title":    true,
		"author":   true,
		"created":  true,
		"updated":  true,
		"comments": true,
		"state":    true,
	}

	var result []string
	for _, col := range columns {
		if valid[col] {
			result = append(result, col)
		}
	}
	return result
}

// issuesLoadedMsg is sent when issues are loaded from the database
type issuesLoadedMsg struct {
	issues []database.ListIssue
}

// loadIssues loads issues from the database with current sort settings
func (m Model) loadIssues() tea.Cmd {
	return func() tea.Msg {
		db, err := database.InitializeSchema(m.dbPath)
		if err != nil {
			return issuesLoadedMsg{issues: []database.ListIssue{}}
		}

		issues, err := database.ListIssuesSorted(db, m.repo, m.sortField, m.sortDesc)
		if err != nil {
			db.Close()
			return issuesLoadedMsg{issues: []database.ListIssue{}}
		}

		return issuesLoadedMsg{issues: issues}
	}
}

// Selected returns the currently selected issue
func (m Model) Selected() *database.ListIssue {
	if m.selected < 0 || m.selected >= len(m.issues) {
		return nil
	}
	return &m.issues[m.selected]
}

// SetDimensions updates the model dimensions
func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// cycleSortField moves to the next sort field in the cycle
func (m *Model) cycleSortField() {
	for i, field := range m.sortFields {
		if field == m.sortField {
			// Move to next field
			nextIndex := (i + 1) % len(m.sortFields)
			m.sortField = m.sortFields[nextIndex]
			// Reset to descending when changing field
			m.sortDesc = true
			return
		}
	}
	// If current sort field not in list, start from beginning
	m.sortField = m.sortFields[0]
	m.sortDesc = true
}

// GetSortField returns the current sort field
func (m Model) GetSortField() string {
	return m.sortField
}

// GetSortDescending returns whether sort is descending
func (m Model) GetSortDescending() bool {
	return m.sortDesc
}

// saveSortConfig returns a command that saves the sort configuration
func (m Model) saveSortConfig() tea.Cmd {
	return func() tea.Msg {
		if m.saveSort != nil {
			if err := m.saveSort(m.sortField, m.sortDesc); err != nil {
				// Log error but don't fail - we can continue without persistence
				return sortSavedMsg{err: err}
			}
		}
		return sortSavedMsg{}
	}
}

// sortSavedMsg indicates that sort preferences were saved (or failed)
type sortSavedMsg struct {
	err error
}
