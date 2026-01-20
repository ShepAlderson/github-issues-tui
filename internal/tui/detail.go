package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/git/github-issues-tui/internal/db"
)

// DetailModel represents the TUI model for the issue detail view
type DetailModel struct {
	db            *db.DB
	issue         *db.Issue
	viewport      viewport.Model
	showRendered  bool // Toggle between raw markdown and rendered
	quitting      bool
	err           error
	viewComments  bool // Flag to navigate to comments
	width         int
	height        int
}

// NewDetailModel creates a new issue detail model
func NewDetailModel(dbPath string, issueNum int) (*DetailModel, error) {
	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Fetch the issue
	issue, err := database.GetIssue(issueNum)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to fetch issue: %w", err)
	}

	// Create viewport
	vp := viewport.New(80, 20) // Default size, will be resized with WindowSizeMsg

	// Create model
	model := &DetailModel{
		db:           database,
		issue:        issue,
		viewport:     vp,
		showRendered: false,
		quitting:     false,
	}

	// Set initial content
	model.viewport.SetContent(model.getContent())

	return model, nil
}

// getContent returns the content to display (raw or rendered)
func (m DetailModel) getContent() string {
	header := m.renderHeader()
	body := m.renderBody()
	return header + "\n\n" + body
}

// renderHeader renders the issue header section
func (m DetailModel) renderHeader() string {
	if m.issue == nil {
		return ""
	}

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	labelStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("63")).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1).
		MarginRight(1)

	assigneeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("226"))

	// Build header parts
	var header strings.Builder

	// Title line: #number Title
	titleLine := fmt.Sprintf("#%d %s", m.issue.Number, m.issue.Title)
	header.WriteString(titleStyle.Render(titleLine))
	header.WriteRune('\n')

	// Status line: state, author, dates, comments
	statusLine := fmt.Sprintf("State: %s | Author: %s | Created: %.10s | Updated: %.10s | Comments: %d",
		m.issue.State, m.issue.Author, m.issue.CreatedAt, m.issue.UpdatedAt, m.issue.CommentCount)
	header.WriteString(infoStyle.Render(statusLine))
	header.WriteRune('\n')

	// Labels if present
	if len(m.issue.Labels) > 0 {
		header.WriteRune('\n')
		header.WriteString(infoStyle.Render("Labels: "))
		for i, label := range m.issue.Labels {
			if i > 0 {
				header.WriteRune(' ')
			}
			header.WriteString(labelStyle.Render(label))
		}
		header.WriteRune('\n')
	}

	// Assignees if present
	if len(m.issue.Assignees) > 0 {
		assigneeLine := fmt.Sprintf("Assignees: %s", strings.Join(m.issue.Assignees, ", "))
		header.WriteString(assigneeStyle.Render(assigneeLine))
		header.WriteRune('\n')
	}

	// Render mode indicator
	mode := "Raw Markdown"
	if m.showRendered {
		mode = "Rendered Markdown"
	}
	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		MarginLeft(1)
	header.WriteRune('\n')
	header.WriteString(modeStyle.Render("View: " + mode + " (press 'm' to toggle, 'c' for comments)"))

	return header.String()
}

// renderBody renders the issue body
func (m DetailModel) renderBody() string {
	if m.issue == nil || m.issue.Body == "" {
		return "No description provided."
	}

	if !m.showRendered {
		// Return raw markdown with border
		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
		return borderStyle.Render(m.issue.Body)
	}

	// Render markdown with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.viewport.Width-4), // Account for padding
	)
	if err != nil {
		return fmt.Sprintf("Error rendering markdown: %v", err)
	}

	rendered, err := renderer.Render(m.issue.Body)
	if err != nil {
		return fmt.Sprintf("Error rendering markdown: %v", err)
	}

	return rendered
}

// Init initializes the model
func (m DetailModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport size accounting for header
		headerHeight := m.calculateHeaderHeight()
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - 1
		m.viewport.SetContent(m.getContent())

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "m":
			// Toggle between raw and rendered markdown
			m.showRendered = !m.showRendered
			m.viewport.SetContent(m.getContent())

		case "c":
			// Navigate to comments view
			m.viewComments = true
			m.quitting = true
			return m, tea.Quit

		default:
			// Pass other keys to viewport for scrolling
			m.viewport, cmd = m.viewport.Update(msg)
		}

	default:
		// Pass other messages to viewport
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

// calculateHeaderHeight calculates the height needed for the header
func (m DetailModel) calculateHeaderHeight() int {
	header := m.renderHeader()
	lines := strings.Split(header, "\n")
	return len(lines)
}

// View renders the UI
func (m DetailModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.issue == nil {
		return "No issue to display.\n"
	}

	// Build the full view
	var view strings.Builder

	// Header (title and metadata)
	header := m.renderHeader()
	view.WriteString(header)

	// Body (in viewport)
	view.WriteRune('\n')
	view.WriteString(m.viewport.View())

	return view.String()
}

// Close closes the database connection
func (m *DetailModel) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
