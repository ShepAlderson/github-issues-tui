package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/git/github-issues-tui/internal/config"
	"github.com/shepbook/git/github-issues-tui/internal/db"
)

// ListModel represents the TUI model for the issue list view
type ListModel struct {
	db        *db.DB
	config    *config.Config
	issues    []*db.Issue
	table     table.Model
	selected  int
	width     int
	height    int
	quitting  bool
	err       error
}

// NewListModel creates a new issue list model
func NewListModel(dbPath string, cfg *config.Config) (*ListModel, error) {
	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Fetch issues for display
	issues, err := database.GetIssuesForDisplay()
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	// Use configured columns or defaults
	columns := cfg.Display.Columns
	if len(columns) == 0 {
		columns = config.GetDefaultDisplayColumns()
	}

	// Create table columns
	tableCols := make([]table.Column, 0, len(columns))
	for _, col := range columns {
		tableCols = append(tableCols, table.Column{
			Title: formatColumnTitle(col),
			Width: getColumnWidth(col),
		})
	}

	// Create table rows
	rows := make([]table.Row, 0, len(issues))
	for _, issue := range issues {
		row := make(table.Row, 0, len(columns))
		for _, col := range columns {
			row = append(row, formatIssueField(issue, col))
		}
		rows = append(rows, row)
	}

	// Create table model
	t := table.New(
		table.WithColumns(tableCols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	// Set initial cursor if there are issues
	if len(rows) > 0 {
		t.SetCursor(0)
	}

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return &ListModel{
		db:       database,
		config:   cfg,
		issues:   issues,
		table:    t,
		selected: 0,
	}, nil
}

// Init initializes the model
func (m ListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetHeight(msg.Height - 4) // Leave room for header/status

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "j", "down":
			// Move down with 'j' or arrow down
			m.table, cmd = m.table.Update(tea.KeyMsg{Type: tea.KeyDown})
			m.selected = m.table.Cursor()

		case "k", "up":
			// Move up with 'k' or arrow up
			m.table, cmd = m.table.Update(tea.KeyMsg{Type: tea.KeyUp})
			m.selected = m.table.Cursor()

		case "enter":
			// TODO: In future stories, show issue detail view
			if m.selected < len(m.issues) && m.selected >= 0 {
				m.err = fmt.Errorf("issue detail view not yet implemented (selected issue #%d)", m.issues[m.selected].Number)
			}
		}
	}

	return m, cmd
}

// View renders the UI
func (m ListModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if len(m.issues) == 0 {
		return "No issues found. Run 'ghissues sync' to fetch issues.\n"
	}

	// Build the UI
	var b strings.Builder

	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n")

	// Status bar
	status := m.renderStatusBar()
	b.WriteString(status)

	return b.String()
}

// Close closes the database connection
func (m *ListModel) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// renderStatusBar renders the status bar at the bottom
func (m ListModel) renderStatusBar() string {
	if len(m.issues) == 0 {
		return ""
	}

	status := fmt.Sprintf(" %d/%d issues ", m.selected+1, len(m.issues))

	// Style the status bar
	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("229"))

	return statusStyle.Render(status)
}

// formatColumnTitle formats a column name for display
func formatColumnTitle(col string) string {
	switch col {
	case "number":
		return "#"
	case "title":
		return "Title"
	case "author":
		return "Author"
	case "created_at":
		return "Created"
	case "comment_count":
		return "Comments"
	default:
		// Simple capitalization instead of deprecated strings.Title
		words := strings.Split(strings.ReplaceAll(col, "_", " "), " ")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		return strings.Join(words, " ")
	}
}

// getColumnWidth returns the default width for a column
func getColumnWidth(col string) int {
	switch col {
	case "number":
		return 6
	case "author":
		return 15
	case "created_at":
		return 12
	case "comment_count":
		return 10
	default:
		return 30 // Flexible width for title
	}
}

// formatIssueField formats an issue field for display
func formatIssueField(issue *db.Issue, field string) string {
	switch field {
	case "number":
		return fmt.Sprintf("#%d", issue.Number)
	case "title":
		// Truncate long titles
		if len(issue.Title) > 50 {
			return issue.Title[:48] + "..."
		}
		return issue.Title
	case "author":
		return issue.Author
	case "created_at":
		// Format date as YYYY-MM-DD
		if len(issue.CreatedAt) >= 10 {
			return issue.CreatedAt[:10]
		}
		return issue.CreatedAt
	case "comment_count":
		return fmt.Sprintf("%d", issue.CommentCount)
	default:
		return ""
	}
}

// RunListView runs the issue list TUI
func RunListView(dbPath string, cfg *config.Config, output io.Writer) error {
	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		return fmt.Errorf("failed to create list model: %w", err)
	}
	defer model.Close()

	// Check if we're in a test environment
	if os.Getenv("GHISSIES_TEST") == "1" {
		// In test mode, just show a simple message
		fmt.Fprintln(output, "Issue list TUI would be displayed here")
		fmt.Fprintf(output, "Found %d issues\n", len(model.issues))
		return nil
	}

	// Create tea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running list view: %w", err)
	}

	return nil
}
