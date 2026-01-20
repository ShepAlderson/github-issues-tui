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

// CommentsModel represents the TUI model for the issue comments view
type CommentsModel struct {
	db           *db.DB
	issue        *db.Issue
	comments     []*db.Comment
	viewport     viewport.Model
	showRendered bool // Toggle between raw markdown and rendered
	quitting     bool
	err          error
	width        int
	height       int
}

// NewCommentsModel creates a new issue comments model
func NewCommentsModel(dbPath string, issue *db.Issue) (*CommentsModel, error) {
	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Fetch comments for the issue
	comments, err := database.GetComments(issue.Number)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Create viewport
	vp := viewport.New(80, 20) // Default size, will be resized with WindowSizeMsg

	// Create model
	model := &CommentsModel{
		db:           database,
		issue:        issue,
		comments:     comments,
		viewport:     vp,
		showRendered: false,
		quitting:     false,
	}

	// Set initial content
	model.viewport.SetContent(model.getContent())

	return model, nil
}

// getContent returns the content to display (raw or rendered)
func (m CommentsModel) getContent() string {
	header := m.renderHeader()
	comments := m.renderComments()
	return header + "\n\n" + comments
}

// renderHeader renders the issue header section
func (m CommentsModel) renderHeader() string {
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

	modeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true).
		MarginTop(1)

	// Build header parts
	var header strings.Builder

	// Title line: Comments for #number Title
	titleLine := fmt.Sprintf("Comments for #%d %s", m.issue.Number, m.issue.Title)
	header.WriteString(titleStyle.Render(titleLine))
	header.WriteRune('\n')

	// Comment count
	fmt.Fprintf(&header, "%s", infoStyle.Render(fmt.Sprintf("%d %s", len(m.comments), func() string {
		if len(m.comments) == 1 {
			return "comment"
		}
		return "comments"
	}())))

	// Render mode indicator
	mode := "Raw Markdown"
	if m.showRendered {
		mode = "Rendered Markdown"
	}
	header.WriteString(modeStyle.Render("\n(View: " + mode + " | press 'm' to toggle)"))

	return header.String()
}

// renderComments renders all comments
func (m CommentsModel) renderComments() string {
	if len(m.comments) == 0 {
		return "\nNo comments found for this issue."
	}

	var comments strings.Builder

	for i, comment := range m.comments {
		// Add separator between comments
		if i > 0 {
			comments.WriteString("\n\n" + strings.Repeat("─", m.viewport.Width) + "\n\n")
		}

		// Render comment header
		authorStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("87"))

		dateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

		commentHeader := fmt.Sprintf("%s • %s",
			authorStyle.Render(comment.Author),
			dateStyle.Render(comment.CreatedAt))
		comments.WriteString(commentHeader)
		comments.WriteRune('\n')

		// Render comment body
		if m.showRendered {
			// Render markdown with glamour
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.viewport.Width-2), // Account for padding
			)
			if err != nil {
				fmt.Fprintf(&comments, "Error rendering markdown: %v\n", err)
				continue
			}

			rendered, err := renderer.Render(comment.Body)
			if err != nil {
				fmt.Fprintf(&comments, "Error rendering markdown: %v\n", err)
				continue
			}

			comments.WriteString(rendered)
		} else {
			// Raw markdown with indentation
			for _, line := range strings.Split(comment.Body, "\n") {
				fmt.Fprintf(&comments, "  %s\n", line)
			}
		}
	}

	return comments.String()
}

// Init initializes the model
func (m CommentsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *CommentsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "m":
			// Toggle between raw and rendered markdown
			m.showRendered = !m.showRendered
			m.viewport.SetContent(m.getContent())

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
func (m CommentsModel) calculateHeaderHeight() int {
	header := m.renderHeader()
	lines := strings.Split(header, "\n")
	return len(lines)
}

// View renders the UI
func (m CommentsModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.issue == nil {
		return "No issue to display comments for.\n"
	}

	// Build the full view
	var view strings.Builder

	// Header (title and metadata)
	header := m.renderHeader()
	view.WriteString(header)

	// Comments (in viewport)
	view.WriteRune('\n')
	view.WriteString(m.viewport.View())

	return view.String()
}

// Close closes the database connection
func (m *CommentsModel) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
