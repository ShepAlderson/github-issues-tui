package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/database"
)

// Model represents the issue detail view
type Model struct {
	Issue        database.IssueDetail
	Width        int
	Height       int
	RenderedMode bool
}

// Styles for the detail view
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

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	stateOpenStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#238636")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	stateClosedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#8957E5")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F6FEB")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			MarginRight(1)

	bodyStyle = lipgloss.NewStyle().
			Padding(1, 0)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			MarginTop(1)
)

// NewModel creates a new detail model
func NewModel(issue database.IssueDetail, width, height int) Model {
	return Model{
		Issue:        issue,
		Width:        width,
		Height:       height,
		RenderedMode: true, // Default to rendered mode
	}
}

// SetDimensions updates the model dimensions
func (m *Model) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
}

// ToggleRenderedMode toggles between rendered and raw markdown mode
func (m *Model) ToggleRenderedMode() {
	m.RenderedMode = !m.RenderedMode
}

// View renders the detail view
func (m Model) View() string {
	var b strings.Builder

	// Header section with title and meta info
	header := m.renderHeader()
	b.WriteString(header)

	// Labels section
	if len(m.Issue.Labels) > 0 {
		b.WriteString(m.renderLabels())
		b.WriteString("\n")
	}

	// Assignees section
	if len(m.Issue.Assignees) > 0 {
		b.WriteString(m.renderAssignees())
		b.WriteString("\n")
	}

	// Body section
	body := m.renderBody()
	b.WriteString(body)

	// Footer with instructions
	modeText := "rendered"
	if !m.RenderedMode {
		modeText = "raw"
	}
	footer := footerStyle.Render(fmt.Sprintf("Mode: %s | m to toggle | q to quit", modeText))
	b.WriteString(footer)

	return b.String()
}

// renderHeader creates the header section with issue details
func (m Model) renderHeader() string {
	var parts []string

	// Title with issue number
	title := fmt.Sprintf("#%d %s", m.Issue.Number, m.Issue.Title)
	parts = append(parts, titleStyle.Render(title))

	// State badge
	var stateBadge string
	if m.Issue.State == "open" {
		stateBadge = stateOpenStyle.Render("● open")
	} else {
		stateBadge = stateClosedStyle.Render("● closed")
	}

	// Meta line: author and dates
	created := formatDate(m.Issue.CreatedAt)
	updated := formatDate(m.Issue.UpdatedAt)
	meta := fmt.Sprintf("by %s • created %s • updated %s", m.Issue.Author, created, updated)

	// Add closed date if present
	if m.Issue.ClosedAt != "" {
		closed := formatDate(m.Issue.ClosedAt)
		meta += fmt.Sprintf(" • closed %s", closed)
	}

	parts = append(parts, stateBadge)
	parts = append(parts, metaStyle.Render(meta))

	return headerStyle.Render(strings.Join(parts, "\n"))
}

// renderLabels creates the labels section
func (m Model) renderLabels() string {
	if len(m.Issue.Labels) == 0 {
		return ""
	}

	var labels []string
	for _, label := range m.Issue.Labels {
		labels = append(labels, labelStyle.Render(label))
	}

	return strings.Join(labels, " ")
}

// renderAssignees creates the assignees section
func (m Model) renderAssignees() string {
	if len(m.Issue.Assignees) == 0 {
		return ""
	}

	assigneesList := strings.Join(m.Issue.Assignees, ", ")
	return metaStyle.Render(fmt.Sprintf("Assignees: %s", assigneesList))
}

// renderBody renders the issue body
func (m Model) renderBody() string {
	if m.Issue.Body == "" {
		return bodyStyle.Render("*No description provided*")
	}

	// Calculate available height for body
	headerLines := 6 // Approximate lines for header, labels, assignees
	footerLines := 3 // Footer + padding
	availableHeight := m.Height - headerLines - footerLines
	if availableHeight < 5 {
		availableHeight = 10
	}

	if m.RenderedMode {
		// Use glamour for markdown rendering
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.Width-4),
		)
		if err != nil {
			// Fall back to raw if glamour fails
			body := truncateBody(m.Issue.Body, availableHeight)
			return bodyStyle.Render(body)
		}

		rendered, err := renderer.Render(m.Issue.Body)
		if err != nil {
			body := truncateBody(m.Issue.Body, availableHeight)
			return bodyStyle.Render(body)
		}

		return bodyStyle.Render(rendered)
	}

	// Raw mode - show markdown as-is
	body := truncateBody(m.Issue.Body, availableHeight)
	return bodyStyle.Render(body)
}

// formatDate formats a date string for display
func formatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02")
}

// truncateBody truncates body text to maxLines lines
func truncateBody(body string, maxLines int) string {
	lines := strings.Split(body, "\n")
	if len(lines) <= maxLines {
		return body
	}
	return strings.Join(lines[:maxLines], "\n") + "\n..."
}

// IssueKey returns a unique key for the issue detail view
func IssueKey(number int) string {
	return fmt.Sprintf("detail_%d", number)
}
