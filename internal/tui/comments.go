package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/database"
)

// CommentsComponent represents the comments view component
type CommentsComponent struct {
	dbManager         *database.DBManager
	comments          []database.Comment
	currentIssueNumber int
	currentIssueTitle  string
	showRawMarkdown   bool
	scrollOffset      int
	maxVisibleLines   int
	width             int
	height            int
}

// NewCommentsComponent creates a new comments component
func NewCommentsComponent(dbManager *database.DBManager) *CommentsComponent {
	return &CommentsComponent{
		dbManager:         dbManager,
		comments:          nil,
		currentIssueNumber: 0,
		currentIssueTitle:  "",
		showRawMarkdown:   false,
		scrollOffset:      0,
		maxVisibleLines:   0,
	}
}

// SetIssue sets the current issue to display comments for
func (cc *CommentsComponent) SetIssue(issueNumber int, issueTitle string) error {
	if cc.dbManager == nil {
		return fmt.Errorf("database manager not initialized")
	}

	// Get issue details first to get title
	issue, err := cc.dbManager.GetIssueByNumber(issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get issue #%d: %w", issueNumber, err)
	}

	// Get comments for the issue
	comments, err := cc.dbManager.GetCommentsByIssueNumber(issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get comments for issue #%d: %w", issueNumber, err)
	}

	cc.comments = comments
	cc.currentIssueNumber = issueNumber
	cc.currentIssueTitle = issue.Title
	cc.scrollOffset = 0
	return nil
}

// Init implements the tea.Model interface
func (cc *CommentsComponent) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface
func (cc *CommentsComponent) Update(msg tea.Msg) (*CommentsComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cc.width = msg.Width
		cc.height = msg.Height
		// Calculate max visible lines based on height
		cc.maxVisibleLines = max(0, cc.height-8) // Reserve space for header and footer

	case tea.KeyMsg:
		switch msg.String() {
		case "m":
			cc.toggleMarkdownView()
			return cc, nil
		case "down", "j":
			cc.scrollDown()
			return cc, nil
		case "up", "k":
			cc.scrollUp()
			return cc, nil
		case "g":
			cc.scrollToTop()
			return cc, nil
		case "G":
			cc.scrollToBottom()
			return cc, nil
		}
	}

	return cc, nil
}

// View renders the comments component
func (cc *CommentsComponent) View() string {
	if cc.currentIssueNumber == 0 || len(cc.comments) == 0 {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Render("No comments to display\n\nPress 'q' or Ctrl+C to quit")
	}

	var builder strings.Builder

	// Render header with issue title/number
	builder.WriteString(cc.renderHeader())
	builder.WriteString("\n\n")

	// Render comments
	commentsView := cc.renderComments()
	builder.WriteString(commentsView)

	// Add footer with navigation hints
	builder.WriteString("\n\n")
	builder.WriteString(cc.renderFooter())

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(cc.width).
		Render(builder.String())
}

// renderHeader renders the issue header with title and number
func (cc *CommentsComponent) renderHeader() string {
	// Style for different elements
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")) // Bright blue

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")) // Orange

	header := fmt.Sprintf("%s\nComments: %s",
		titleStyle.Render(fmt.Sprintf("#%d %s", cc.currentIssueNumber, cc.currentIssueTitle)),
		countStyle.Render(fmt.Sprintf("%d", len(cc.comments))),
	)

	return header
}

// renderComments renders all comments chronologically
func (cc *CommentsComponent) renderComments() string {
	if len(cc.comments) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No comments yet.")
	}

	var builder strings.Builder

	// Calculate visible range based on scroll
	startIdx := cc.scrollOffset
	endIdx := min(len(cc.comments), startIdx+cc.maxVisibleLines)
	if startIdx >= len(cc.comments) {
		startIdx = max(0, len(cc.comments)-cc.maxVisibleLines)
		endIdx = len(cc.comments)
	}

	for i := startIdx; i < endIdx; i++ {
		comment := cc.comments[i]
		builder.WriteString(cc.renderComment(comment))
		if i < endIdx-1 {
			builder.WriteString("\n\n")
			// Add separator between comments
			builder.WriteString(strings.Repeat("─", cc.width-4))
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}

// renderComment renders a single comment with author, date, and body
func (cc *CommentsComponent) renderComment(comment database.Comment) string {
	var builder strings.Builder

	// Render comment header with author and date
	authorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214")) // Orange

	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Dark gray

	createdStr := comment.CreatedAt.Format("2006-01-02 15:04")
	builder.WriteString(fmt.Sprintf("%s • %s\n",
		authorStyle.Render(comment.Author),
		dateStyle.Render(createdStr),
	))

	// Render comment body
	body := cc.renderCommentBody(comment.Body)
	builder.WriteString(body)

	return builder.String()
}

// renderCommentBody renders the comment body, either as raw markdown or rendered
func (cc *CommentsComponent) renderCommentBody(body string) string {
	if body == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("(No content)")
	}

	if cc.showRawMarkdown {
		// Show raw markdown
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(body)
	}

	// Render markdown with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(cc.width-4), // Account for padding
	)
	if err != nil {
		// Fall back to raw markdown if glamour fails
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(body)
	}

	rendered, err := renderer.Render(body)
	if err != nil {
		// Fall back to raw markdown if rendering fails
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(body)
	}

	return rendered
}

// renderFooter renders navigation hints
func (cc *CommentsComponent) renderFooter() string {
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	hints := []string{}
	if len(cc.comments) > 0 {
		if cc.showRawMarkdown {
			hints = append(hints, "m: view rendered markdown")
		} else {
			hints = append(hints, "m: view raw markdown")
		}
		hints = append(hints, "↑/↓: scroll comments")
		hints = append(hints, "g/G: top/bottom")
	}
	hints = append(hints, "q/esc: back to issue list")

	return hintStyle.Render(strings.Join(hints, " • "))
}

// toggleMarkdownView toggles between raw and rendered markdown
func (cc *CommentsComponent) toggleMarkdownView() {
	if len(cc.comments) > 0 {
		cc.showRawMarkdown = !cc.showRawMarkdown
		cc.scrollOffset = 0
	}
}

// scrollDown scrolls the comments down
func (cc *CommentsComponent) scrollDown() {
	if len(cc.comments) == 0 {
		return
	}

	cc.scrollOffset = min(cc.scrollOffset+1, max(0, len(cc.comments)-cc.maxVisibleLines))
}

// scrollUp scrolls the comments up
func (cc *CommentsComponent) scrollUp() {
	cc.scrollOffset = max(0, cc.scrollOffset-1)
}

// scrollToTop scrolls to the top of the comments
func (cc *CommentsComponent) scrollToTop() {
	cc.scrollOffset = 0
}

// scrollToBottom scrolls to the bottom of the comments
func (cc *CommentsComponent) scrollToBottom() {
	if len(cc.comments) == 0 {
		return
	}

	cc.scrollOffset = max(0, len(cc.comments)-cc.maxVisibleLines)
}

