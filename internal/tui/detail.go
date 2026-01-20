package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/storage"
	"github.com/shepbook/ghissues/internal/theme"
)

// DetailPanel represents the state of the issue detail view
type DetailPanel struct {
	Issue          storage.Issue
	RenderMarkdown bool          // Whether to render markdown or show raw
	ScrollOffset   int           // Current scroll position
	ViewportHeight int           // Height of the viewport
}

// NewDetailPanel creates a new detail panel for an issue
func NewDetailPanel(issue storage.Issue) *DetailPanel {
	return &DetailPanel{
		Issue:          issue,
		RenderMarkdown: false, // Default to raw markdown
		ScrollOffset:   0,
		ViewportHeight: 20,
	}
}

// ToggleMarkdown toggles between rendered and raw markdown
func (p *DetailPanel) ToggleMarkdown() {
	p.RenderMarkdown = !p.RenderMarkdown
	// Reset scroll when toggling
	p.ScrollOffset = 0
}

// ScrollDown scrolls the detail panel down by one line
func (p *DetailPanel) ScrollDown() {
	// Don't scroll if there's no content
	if p.Issue.Body == "" && p.Issue.Title == "" {
		return
	}
	// For now, just increment. We'll limit this later based on content height
	p.ScrollOffset++
}

// ScrollUp scrolls the detail panel up by one line
func (p *DetailPanel) ScrollUp() {
	if p.ScrollOffset > 0 {
		p.ScrollOffset--
	}
}

// SetViewport sets the viewport height
func (p *DetailPanel) SetViewport(height int) {
	p.ViewportHeight = height
}

// View renders the detail panel
func (p *DetailPanel) View(theme *theme.Theme) string {
	if p.Issue.Number == 0 {
		return "No issue selected"
	}

	var parts []string

	// Header with issue number and title
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Title))
	header := headerStyle.Render(fmt.Sprintf("#%d %s", p.Issue.Number, p.Issue.Title))
	parts = append(parts, header)

	// Metadata row
	metaParts := []string{
		fmt.Sprintf("Author: %s", p.Issue.Author),
		fmt.Sprintf("Status: %s", p.Issue.State),
		fmt.Sprintf("Created: %s", p.Issue.CreatedAt.Format("2006-01-02 15:04")),
	}
	if p.Issue.State == "closed" && p.Issue.ClosedAt != nil {
		metaParts = append(metaParts, fmt.Sprintf("Closed: %s", p.Issue.ClosedAt.Format("2006-01-02 15:04")))
	}
	metaRow := strings.Join(metaParts, " | ")
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint))
	parts = append(parts, metaStyle.Render(metaRow))

	// Labels if present
	if p.Issue.Labels != "" {
		labels := strings.Split(p.Issue.Labels, ",")
		labelParts := []string{}
		for _, label := range labels {
			label = strings.TrimSpace(label)
			if label != "" {
				labelStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(theme.Label)).
					Foreground(lipgloss.Color("white")).
					Padding(0, 1)
				labelParts = append(labelParts, labelStyle.Render(label))
			}
		}
		if len(labelParts) > 0 {
			parts = append(parts, strings.Join(labelParts, " "))
		}
	}

	// Assignees if present
	if p.Issue.Assignees != "" {
		assignees := strings.Split(p.Issue.Assignees, ",")
		assigneeList := []string{}
		for _, assignee := range assignees {
			assignee = strings.TrimSpace(assignee)
			if assignee != "" {
				assigneeList = append(assigneeList, assignee)
			}
		}
		if len(assigneeList) > 0 {
			assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint))
			parts = append(parts, assigneeStyle.Render("Assignees: "+strings.Join(assigneeList, ", ")))
		}
	}

	// Separator
	separator := strings.Repeat("─", 80)
	parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint)).Render(separator))

	// Body
	body := p.renderBody(theme)
	parts = append(parts, body)

	return strings.Join(parts, "\n")
}

// renderBody renders the issue body, either as raw markdown or rendered
func (p *DetailPanel) renderBody(theme *theme.Theme) string {
	if p.Issue.Body == "" {
		bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Faint))
		return bodyStyle.Render("No description provided.")
	}

	if p.RenderMarkdown {
		// Use glamour to render markdown
		rendered, err := p.renderWithGlamour(p.Issue.Body)
		if err != nil {
			// Fallback to raw if rendering fails
			return p.Issue.Body
		}
		return rendered
	}

	return p.Issue.Body
}

// renderWithGlamour renders markdown using glamour
func (p *DetailPanel) renderWithGlamour(markdown string) (string, error) {
	// Create a glamour renderer with terminal width
	// Use 80 as default width, will be adjusted by viewport
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
func (p *DetailPanel) GetVisibleLines(theme *theme.Theme) []string {
	view := p.View(theme)
	lines := strings.Split(view, "\n")

	start := p.ScrollOffset
	if start >= len(lines) {
		return []string{}
	}

	end := start + p.ViewportHeight
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end]
}
