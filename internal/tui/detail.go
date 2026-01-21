package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/database"
)

// IssueDetailComponent represents the issue detail view component
type IssueDetailComponent struct {
	dbManager        *database.DBManager
	currentIssue     *database.IssueDetail
	showRawMarkdown  bool
	scrollOffset     int
	maxVisibleLines  int
	width            int
	height           int
}

// NewIssueDetailComponent creates a new issue detail component
func NewIssueDetailComponent(dbManager *database.DBManager) *IssueDetailComponent {
	return &IssueDetailComponent{
		dbManager:       dbManager,
		currentIssue:    nil,
		showRawMarkdown: false,
		scrollOffset:    0,
		maxVisibleLines: 0,
	}
}

// SetIssue sets the current issue to display
func (idc *IssueDetailComponent) SetIssue(issueNumber int) error {
	if idc.dbManager == nil {
		return fmt.Errorf("database manager not initialized")
	}

	issue, err := idc.dbManager.GetIssueByNumber(issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get issue #%d: %w", issueNumber, err)
	}

	idc.currentIssue = issue
	idc.scrollOffset = 0
	return nil
}

// Init implements the tea.Model interface
func (idc *IssueDetailComponent) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface
func (idc *IssueDetailComponent) Update(msg tea.Msg) (*IssueDetailComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		idc.width = msg.Width
		idc.height = msg.Height
		// Calculate max visible lines based on height
		idc.maxVisibleLines = max(0, idc.height-10) // Reserve space for header and footer

	case tea.KeyMsg:
		switch msg.String() {
		case "m":
			idc.toggleMarkdownView()
			return idc, nil
		case "down", "j":
			idc.scrollDown()
			return idc, nil
		case "up", "k":
			idc.scrollUp()
			return idc, nil
		case "g":
			idc.scrollToTop()
			return idc, nil
		case "G":
			idc.scrollToBottom()
			return idc, nil
		}
	}

	return idc, nil
}

// View renders the issue detail component
func (idc *IssueDetailComponent) View() string {
	if idc.currentIssue == nil {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Render("Select an issue to view details\n\nPress 'q' or Ctrl+C to quit")
	}

	var builder strings.Builder

	// Render header
	builder.WriteString(idc.renderHeader())
	builder.WriteString("\n\n")

	// Render labels and assignees if present
	metadata := idc.renderMetadata()
	if metadata != "" {
		builder.WriteString(metadata)
		builder.WriteString("\n\n")
	}

	// Render body
	body := idc.renderBody()
	builder.WriteString(body)

	// Add footer with navigation hints
	builder.WriteString("\n\n")
	builder.WriteString(idc.renderFooter())

	return lipgloss.NewStyle().
		Padding(1, 2).
		Width(idc.width).
		Render(builder.String())
}

// renderHeader renders the issue header with number, title, author, status, and dates
func (idc *IssueDetailComponent) renderHeader() string {
	issue := idc.currentIssue
	if issue == nil {
		return ""
	}

	// Format dates
	createdStr := issue.CreatedAt.Format("2006-01-02 15:04")
	updatedStr := issue.UpdatedAt.Format("2006-01-02 15:04")

	// Style for different elements
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")) // Bright blue

	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")) // Gray

	authorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")) // Orange

	stateStyle := lipgloss.NewStyle().
		Bold(true)
	if issue.State == "open" {
		stateStyle = stateStyle.Foreground(lipgloss.Color("46")) // Green
	} else {
		stateStyle = stateStyle.Foreground(lipgloss.Color("196")) // Red
	}

	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")) // Dark gray

	header := fmt.Sprintf("%s • %s\n%s • %s • %s comments\nCreated: %s • Updated: %s",
		titleStyle.Render(fmt.Sprintf("#%d %s", issue.Number, issue.Title)),
		stateStyle.Render(strings.ToUpper(issue.State)),
		numberStyle.Render(fmt.Sprintf("#%d", issue.Number)),
		authorStyle.Render(issue.Author),
		numberStyle.Render(fmt.Sprintf("%d", issue.CommentCount)),
		dateStyle.Render(createdStr),
		dateStyle.Render(updatedStr),
	)

	return header
}

// renderMetadata renders labels and assignees
func (idc *IssueDetailComponent) renderMetadata() string {
	issue := idc.currentIssue
	if issue == nil {
		return ""
	}

	var parts []string

	// Render labels if present
	if len(issue.Labels) > 0 {
		var labels []string
		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("240")).
			Padding(0, 1).
			MarginRight(1)

		for _, label := range issue.Labels {
			labels = append(labels, labelStyle.Render(label))
		}
		parts = append(parts, fmt.Sprintf("Labels: %s", strings.Join(labels, " ")))
	}

	// Render assignees if present
	if len(issue.Assignees) > 0 {
		assigneeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // Orange
		assigneesStr := assigneeStyle.Render(strings.Join(issue.Assignees, ", "))
		parts = append(parts, fmt.Sprintf("Assignees: %s", assigneesStr))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n")
}

// renderBody renders the issue body, either as raw markdown or rendered
func (idc *IssueDetailComponent) renderBody() string {
	issue := idc.currentIssue
	if issue == nil || issue.Body == "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("No description provided.")
	}

	if idc.showRawMarkdown {
		// Show raw markdown
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(issue.Body)
	}

	// Render markdown with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(idc.width-4), // Account for padding
	)
	if err != nil {
		// Fall back to raw markdown if glamour fails
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(issue.Body)
	}

	rendered, err := renderer.Render(issue.Body)
	if err != nil {
		// Fall back to raw markdown if rendering fails
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Render(issue.Body)
	}

	// Apply scrolling
	lines := strings.Split(rendered, "\n")
	start := idc.scrollOffset
	end := min(len(lines), start+idc.maxVisibleLines)
	if start >= len(lines) {
		start = max(0, len(lines)-idc.maxVisibleLines)
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n")
}

// renderFooter renders navigation hints
func (idc *IssueDetailComponent) renderFooter() string {
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	hints := []string{}
	if idc.currentIssue != nil && idc.currentIssue.Body != "" {
		if idc.showRawMarkdown {
			hints = append(hints, "m: view rendered markdown")
		} else {
			hints = append(hints, "m: view raw markdown")
		}
		hints = append(hints, "↑/↓: scroll")
		hints = append(hints, "g/G: top/bottom")
	}
	hints = append(hints, "q: quit")

	return hintStyle.Render(strings.Join(hints, " • "))
}

// toggleMarkdownView toggles between raw and rendered markdown
func (idc *IssueDetailComponent) toggleMarkdownView() {
	if idc.currentIssue != nil && idc.currentIssue.Body != "" {
		idc.showRawMarkdown = !idc.showRawMarkdown
		idc.scrollOffset = 0
	}
}

// scrollDown scrolls the body down
func (idc *IssueDetailComponent) scrollDown() {
	if idc.currentIssue == nil || idc.currentIssue.Body == "" {
		return
	}

	// Estimate line count
	body := idc.currentIssue.Body
	if !idc.showRawMarkdown {
		// For rendered markdown, we'd need to know actual line count
		// For now, use a simple heuristic
		lines := strings.Split(body, "\n")
		idc.scrollOffset = min(idc.scrollOffset+5, max(0, len(lines)-idc.maxVisibleLines))
	} else {
		lines := strings.Split(body, "\n")
		idc.scrollOffset = min(idc.scrollOffset+5, max(0, len(lines)-idc.maxVisibleLines))
	}
}

// scrollUp scrolls the body up
func (idc *IssueDetailComponent) scrollUp() {
	idc.scrollOffset = max(0, idc.scrollOffset-5)
}

// scrollToTop scrolls to the top of the body
func (idc *IssueDetailComponent) scrollToTop() {
	idc.scrollOffset = 0
}

// scrollToBottom scrolls to the bottom of the body
func (idc *IssueDetailComponent) scrollToBottom() {
	if idc.currentIssue == nil || idc.currentIssue.Body == "" {
		return
	}

	// Simple heuristic for max scroll
	lines := strings.Split(idc.currentIssue.Body, "\n")
	idc.scrollOffset = max(0, len(lines)-idc.maxVisibleLines)
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}