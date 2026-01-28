package comments

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/database"
)

// Model represents the comments view TUI state
type Model struct {
	dbPath       string
	repo         string
	issueNumber  int
	issueTitle   string
	comments     []database.Comment
	width        int
	height       int
	scrollOffset int
	renderedMode bool
	db           *sql.DB // Will be set when loading
}

// Styles for the comments view
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginBottom(1)

	commentHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	commentMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888"))

	commentBodyStyle = lipgloss.NewStyle().
				Padding(1, 0)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginTop(1)
)

// NewModel creates a new comments model
func NewModel(dbPath, repo string, issueNumber int, issueTitle string) Model {
	return Model{
		dbPath:       dbPath,
		repo:         repo,
		issueNumber:  issueNumber,
		issueTitle:   issueTitle,
		comments:     []database.Comment{},
		width:        80,
		height:       24,
		scrollOffset: 0,
		renderedMode: true,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.loadComments()
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			// Return to issue list
			return m, tea.Quit
		case tea.KeyUp:
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case tea.KeyDown:
			// Allow scrolling past content
			m.scrollOffset++
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return m, tea.Quit
			case "j":
				m.scrollOffset++
			case "k":
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			case "m", "M":
				m.ToggleRenderedMode()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case commentsLoadedMsg:
		m.comments = msg.comments
	}

	return m, nil
}

// View renders the comments view
func (m Model) View() string {
	var b strings.Builder

	// Header with issue info
	header := m.renderHeader()
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Comments section
	if len(m.comments) == 0 {
		b.WriteString("  *No comments on this issue*\n")
	} else {
		commentsView := m.renderComments()
		b.WriteString(commentsView)
	}

	// Footer with navigation hints
	b.WriteString("\n")
	modeText := "rendered"
	if !m.renderedMode {
		modeText = "raw"
	}
	footer := fmt.Sprintf("Mode: %s | m toggle | j/k scroll | q/esc back", modeText)
	b.WriteString(statusStyle.Render(footer))

	return b.String()
}

// renderHeader creates the header with issue number and title
func (m Model) renderHeader() string {
	var parts []string

	// Issue number and title
	title := fmt.Sprintf("#%d %s", m.issueNumber, m.issueTitle)
	parts = append(parts, titleStyle.Render(title))

	// Comment count
	commentCount := len(m.comments)
	countText := fmt.Sprintf("💬 %d comment", commentCount)
	if commentCount != 1 {
		countText += "s"
	}
	parts = append(parts, commentMetaStyle.Render(countText))

	return strings.Join(parts, "\n")
}

// renderComments renders the list of comments
func (m Model) renderComments() string {
	var b strings.Builder

	// Calculate available height for comments
	headerLines := 4
	footerLines := 3
	availableHeight := m.height - headerLines - footerLines
	if availableHeight < 5 {
		availableHeight = 10
	}

	// Render comments starting from scroll offset
	for i := m.scrollOffset; i < len(m.comments) && i < m.scrollOffset+availableHeight; i++ {
		comment := m.comments[i]
		commentView := m.renderComment(comment)
		b.WriteString(commentView)

		// Add separator between comments
		if i < len(m.comments)-1 {
			b.WriteString(separatorStyle.Render(strings.Repeat("─", m.width-4)))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderComment renders a single comment
func (m Model) renderComment(comment database.Comment) string {
	var b strings.Builder

	// Comment header with author and date
	date := formatDate(comment.CreatedAt)
	header := fmt.Sprintf("@%s • %s", comment.Author, date)
	b.WriteString(commentHeaderStyle.Render(header))
	b.WriteString("\n")

	// Comment body
	body := m.renderCommentBody(comment.Body)
	b.WriteString(commentBodyStyle.Render(body))
	b.WriteString("\n")

	return b.String()
}

// renderCommentBody renders a comment body (rendered or raw markdown)
func (m Model) renderCommentBody(body string) string {
	if body == "" {
		return "*No comment body*"
	}

	if m.renderedMode {
		// Use glamour for markdown rendering
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width-6),
		)
		if err != nil {
			// Fall back to raw if glamour fails
			return body
		}

		rendered, err := renderer.Render(body)
		if err != nil {
			return body
		}

		return rendered
	}

	// Raw mode - show markdown as-is
	return body
}

// ToggleRenderedMode toggles between rendered and raw markdown mode
func (m *Model) ToggleRenderedMode() {
	m.renderedMode = !m.renderedMode
}

// SetDimensions updates the model dimensions
func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

// GetIssueNumber returns the issue number for this view
func (m Model) GetIssueNumber() int {
	return m.issueNumber
}

// commentsLoadedMsg is sent when comments are loaded from the database
type commentsLoadedMsg struct {
	comments []database.Comment
}

// loadComments loads comments from the database
func (m Model) loadComments() tea.Cmd {
	return func() tea.Msg {
		db, err := database.InitializeSchema(m.dbPath)
		if err != nil {
			return commentsLoadedMsg{comments: []database.Comment{}}
		}
		defer db.Close()

		comments, err := database.GetCommentsForIssue(db, m.repo, m.issueNumber)
		if err != nil {
			return commentsLoadedMsg{comments: []database.Comment{}}
		}

		return commentsLoadedMsg{comments: comments}
	}
}

// formatDate formats a date string for display
func formatDate(dateStr string) string {
	if dateStr == "" {
		return "date unknown"
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02")
}
