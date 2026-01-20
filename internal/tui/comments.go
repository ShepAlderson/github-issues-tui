package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/storage"
	"github.com/shepbook/ghissues/internal/theme"
)

// CommentsView represents the state of the comments view for an issue
type CommentsView struct {
	Issue          storage.Issue
	Comments       []storage.Comment
	RenderMarkdown bool
	ScrollOffset   int
	ViewportHeight int
}

// NewCommentsView creates a new comments view for an issue
func NewCommentsView(issue storage.Issue, comments []storage.Comment) *CommentsView {
	return &CommentsView{
		Issue:          issue,
		Comments:       comments,
		RenderMarkdown: false, // Default to raw markdown
		ScrollOffset:   0,
		ViewportHeight: 20,
	}
}

// ToggleMarkdown toggles between rendered and raw markdown
func (v *CommentsView) ToggleMarkdown() {
	v.RenderMarkdown = !v.RenderMarkdown
	// Reset scroll when toggling
	v.ScrollOffset = 0
}

// ScrollDown scrolls the view down by one line
func (v *CommentsView) ScrollDown() {
	v.ScrollOffset++
}

// ScrollUp scrolls the view up by one line
func (v *CommentsView) ScrollUp() {
	if v.ScrollOffset > 0 {
		v.ScrollOffset--
	}
}

// SetViewport sets the viewport height
func (v *CommentsView) SetViewport(height int) {
	v.ViewportHeight = height
}

// View renders the comments view
func (v *CommentsView) View(theme *theme.Theme) string {
	var parts []string

	// Header with issue number and title
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Title))
	header := headerStyle.Render(fmt.Sprintf("#%d %s", v.Issue.Number, v.Issue.Title))
	parts = append(parts, header)

	// Separator
	separator := strings.Repeat("─", 80)
	parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint)).Render(separator))

	// Comments
	if len(v.Comments) == 0 {
		noCommentsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint))
		parts = append(parts, noCommentsStyle.Render("No comments on this issue."))
	} else {
		for i, comment := range v.Comments {
			commentView := v.renderComment(comment, i, theme)
			parts = append(parts, commentView)
			// Add separator between comments
			if i < len(v.Comments)-1 {
				parts = append(parts, "")
			}
		}
	}

	return strings.Join(parts, "\n")
}

// renderComment renders a single comment
func (v *CommentsView) renderComment(comment storage.Comment, index int, theme *theme.Theme) string {
	var parts []string

	// Comment header with author and date
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Success))
	commentHeader := headerStyle.Render(fmt.Sprintf("%s commented on %s",
		comment.Author,
		comment.CreatedAt.Format("2006-01-02 15:04:05")))
	parts = append(parts, commentHeader)

	// Comment body
	body := v.renderCommentBody(comment, theme)
	parts = append(parts, body)

	return strings.Join(parts, "\n")
}

// renderCommentBody renders the comment body, either as raw markdown or rendered
func (v *CommentsView) renderCommentBody(comment storage.Comment, theme *theme.Theme) string {
	if comment.Body == "" {
		bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint))
		return bodyStyle.Render("No content.")
	}

	if v.RenderMarkdown {
		// Use glamour to render markdown
		rendered, err := v.renderWithGlamour(comment.Body)
		if err != nil {
			// Fallback to raw if rendering fails
			return comment.Body
		}
		return rendered
	}

	return comment.Body
}

// renderWithGlamour renders markdown using glamour
func (v *CommentsView) renderWithGlamour(markdown string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return "", err
	}

	return renderer.Render(markdown)
}

// GetVisibleLines returns the visible lines based on scroll offset
func (v *CommentsView) GetVisibleLines(theme *theme.Theme) []string {
	view := v.View(theme)
	lines := strings.Split(view, "\n")

	start := v.ScrollOffset
	if start >= len(lines) {
		return []string{}
	}

	end := start + v.ViewportHeight
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end]
}
