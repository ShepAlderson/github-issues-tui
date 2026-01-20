package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/shepbook/github-issues-tui/internal/sync"
	"github.com/shepbook/github-issues-tui/internal/theme"
)

// View mode constants
const (
	viewModeList = iota
	viewModeComments
	viewModeHelp
)

// Model represents the TUI state
type Model struct {
	issues              []*sync.Issue
	columns             []string
	cursor              int
	width               int
	height              int
	sortBy              string // "updated", "created", "number", "comments"
	sortAscending       bool   // true for ascending, false for descending
	showRawMarkdown     bool   // true to show raw markdown, false for rendered
	detailScrollOffset  int    // scroll offset for detail panel
	viewMode            int    // viewModeList, viewModeComments, or viewModeHelp
	commentsScrollOffset int    // scroll offset for comments view
	currentComments     []*sync.Comment // cached comments for current issue
	store               *sync.IssueStore // for loading comments
	statusError         string // minor error shown in status bar
	modalError          string // critical error shown as modal (blocks interaction)
	lastSyncTime        time.Time // last successful sync time
	previousViewMode    int    // view mode before opening help
	theme               theme.Theme // color theme for the TUI
}

// StatusErrorMsg is a message for minor errors displayed in status bar
type StatusErrorMsg struct {
	Err string
}

// ClearStatusErrorMsg clears the status bar error
type ClearStatusErrorMsg struct{}

// ModalErrorMsg is a message for critical errors displayed as modal
type ModalErrorMsg struct {
	Err string
}

// NewModel creates a new TUI model
func NewModel(issues []*sync.Issue, columns []string, sortBy string, sortAscending bool, store *sync.IssueStore, lastSyncTime time.Time, themeName string) Model {
	m := Model{
		issues:        issues,
		columns:       columns,
		cursor:        0,
		sortBy:        sortBy,
		sortAscending: sortAscending,
		viewMode:      viewModeList,
		store:         store,
		lastSyncTime:  lastSyncTime,
		theme:         theme.GetTheme(themeName),
	}
	m.sortIssues() // Apply initial sort
	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// DEBUG: Log all incoming messages
	fmt.Fprintf(os.Stderr, "DEBUG: Received message type: %T\n", msg)

	switch msg := msg.(type) {
	case StatusErrorMsg:
		m.statusError = msg.Err
		return m, nil

	case ClearStatusErrorMsg:
		m.statusError = ""
		return m, nil

	case ModalErrorMsg:
		m.modalError = msg.Err
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// If modal error is active, only allow Enter to acknowledge and Ctrl+C to quit
		if m.modalError != "" {
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEnter:
				// Acknowledge modal error
				m.modalError = ""
				return m, nil
			}
			// Block all other keys when modal is active
			return m, nil
		}
		// Handle view-specific keys
		if m.viewMode == viewModeHelp {
			return m.handleHelpViewKeys(msg)
		}
		if m.viewMode == viewModeComments {
			return m.handleCommentsViewKeys(msg)
		}

		// List view keys
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			// Enter comments view for selected issue
			if len(m.issues) > 0 && m.store != nil {
				selected := m.SelectedIssue()
				if selected != nil {
					// Load comments for the selected issue
					comments, err := m.store.LoadComments(selected.Number)
					if err != nil {
						// Show error in status bar
						m.statusError = fmt.Sprintf("Failed to load comments: %v", err)
						return m, nil
					}
					m.currentComments = comments
					m.viewMode = viewModeComments
					m.commentsScrollOffset = 0
				}
			}
			return m, nil

		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				return m, tea.Quit
			case "?":
				// Open help overlay
				m.previousViewMode = m.viewMode
				m.viewMode = viewModeHelp
			case "j":
				if m.cursor < len(m.issues)-1 {
					m.cursor++
					m.detailScrollOffset = 0 // Reset scroll when changing issue
				}
			case "k":
				if m.cursor > 0 {
					m.cursor--
					m.detailScrollOffset = 0 // Reset scroll when changing issue
				}
			case "s":
				// Cycle through sort options: updated -> created -> number -> comments -> updated
				switch m.sortBy {
				case "updated":
					m.sortBy = "created"
				case "created":
					m.sortBy = "number"
				case "number":
					m.sortBy = "comments"
				case "comments":
					m.sortBy = "updated"
				}
				m.sortIssues()
				m.cursor = 0 // Reset cursor to top after sorting
			case "S":
				// Toggle sort order
				m.sortAscending = !m.sortAscending
				m.sortIssues()
				m.cursor = 0 // Reset cursor to top after sorting
			case "m":
				// Toggle between raw and rendered markdown
				m.showRawMarkdown = !m.showRawMarkdown
			}

		case tea.KeyDown:
			if m.cursor < len(m.issues)-1 {
				m.cursor++
				m.detailScrollOffset = 0 // Reset scroll when changing issue
			}

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
				m.detailScrollOffset = 0 // Reset scroll when changing issue
			}

		case tea.KeyPgDown:
			// Scroll down in detail panel
			m.detailScrollOffset += 10
			// Max scroll is enforced in rendering

		case tea.KeyPgUp:
			// Scroll up in detail panel
			m.detailScrollOffset -= 10
			if m.detailScrollOffset < 0 {
				m.detailScrollOffset = 0
			}
		}
	}

	return m, nil
}

// handleCommentsViewKeys handles key presses in comments view
func (m Model) handleCommentsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		// Return to list view
		m.viewMode = viewModeList
		m.currentComments = nil
		return m, nil

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "q":
			// Return to list view (not quit in comments view)
			m.viewMode = viewModeList
			m.currentComments = nil
		case "m":
			// Toggle markdown rendering
			m.showRawMarkdown = !m.showRawMarkdown
		case "?":
			// Open help overlay
			m.previousViewMode = m.viewMode
			m.viewMode = viewModeHelp
		}

	case tea.KeyPgDown:
		// Scroll down
		m.commentsScrollOffset += 10

	case tea.KeyPgUp:
		// Scroll up
		m.commentsScrollOffset -= 10
		if m.commentsScrollOffset < 0 {
			m.commentsScrollOffset = 0
		}
	}

	return m, nil
}

// handleHelpViewKeys handles key presses in help view
func (m Model) handleHelpViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		// Return to previous view
		m.viewMode = m.previousViewMode
		return m, nil

	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "?":
			// Toggle help (dismiss)
			m.viewMode = m.previousViewMode
		}
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	// If modal error is active, show modal overlay
	if m.modalError != "" {
		// Render the underlying view first
		var baseView string
		if m.viewMode == viewModeComments {
			baseView = m.renderCommentsView()
		} else if len(m.issues) == 0 {
			baseView = m.theme.NoIssuesStyle.Render("No issues found. Run 'ghissues sync' to fetch issues.")
		} else if m.width == 0 || m.height == 0 {
			baseView = m.renderSimpleList()
		} else {
			baseView = m.renderSplitPane()
		}

		// Overlay modal error
		return m.renderModalError(baseView)
	}

	// Handle help view
	if m.viewMode == viewModeHelp {
		return m.renderHelpView()
	}

	// Handle comments view
	if m.viewMode == viewModeComments {
		return m.renderCommentsView()
	}

	// List view
	if len(m.issues) == 0 {
		return m.theme.NoIssuesStyle.Render("No issues found. Run 'ghissues sync' to fetch issues.")
	}

	// If no dimensions set yet, render simple list (will be set on first WindowSizeMsg)
	if m.width == 0 || m.height == 0 {
		return m.renderSimpleList()
	}

	// Render split-pane layout: list on left, detail on right
	return m.renderSplitPane()
}

// renderSimpleList renders the simple list view (before dimensions are known)
func (m Model) renderSimpleList() string {
	var b strings.Builder

	// Header
	b.WriteString(m.theme.HeaderStyle.Render(m.renderHeader()))
	b.WriteString("\n\n")

	// Issue list
	for i, issue := range m.issues {
		cursor := " "
		style := m.theme.NormalStyle
		if i == m.cursor {
			cursor = ">"
			style = m.theme.SelectedStyle
		}

		line := fmt.Sprintf("%s %s", cursor, m.renderIssue(issue))
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.theme.StatusStyle.Render(m.renderStatus()))

	// Footer
	b.WriteString("\n")
	b.WriteString(m.theme.FooterStyle.Render(m.renderFooter()))

	return b.String()
}

// renderSplitPane renders the split-pane layout with list and detail
func (m Model) renderSplitPane() string {
	// Calculate widths for split pane (60/40 split)
	listWidth := m.width * 60 / 100
	detailWidth := m.width - listWidth - 2 // -2 for separator

	// Calculate heights (leaving room for status and footer)
	contentHeight := m.height - 5 // -5 for header, status, footer

	// Render list panel
	listPanel := m.renderListPanel(listWidth, contentHeight)

	// Render detail panel
	detailPanel := m.renderDetailPanel(detailWidth, contentHeight)

	// Combine panels side by side
	listLines := strings.Split(listPanel, "\n")
	detailLines := strings.Split(detailPanel, "\n")

	var b strings.Builder

	// Ensure both panels have same number of lines
	maxLines := len(listLines)
	if len(detailLines) > maxLines {
		maxLines = len(detailLines)
	}

	for i := 0; i < maxLines; i++ {
		// List panel
		if i < len(listLines) {
			b.WriteString(listLines[i])
		} else {
			b.WriteString(strings.Repeat(" ", listWidth))
		}

		// Separator
		b.WriteString(" │ ")

		// Detail panel
		if i < len(detailLines) {
			b.WriteString(detailLines[i])
		}

		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(m.theme.StatusStyle.Render(m.renderStatus()))

	// Footer
	b.WriteString("\n")
	b.WriteString(m.theme.FooterStyle.Render(m.renderFooter()))

	return b.String()
}

// renderModalError overlays a modal error dialog on top of the base view
func (m Model) renderModalError(baseView string) string {
	// Build modal content
	var b strings.Builder
	b.WriteString(m.theme.ModalTitleStyle.Render("Error"))
	b.WriteString("\n\n")
	b.WriteString(m.modalError)
	b.WriteString("\n\n")
	b.WriteString(m.theme.FooterStyle.Render("Press Enter to continue"))

	modal := m.theme.ModalStyle.Render(b.String())

	// For simple overlay, just append modal to base view
	// In a full implementation, you'd center the modal on screen
	var output strings.Builder
	output.WriteString(baseView)
	output.WriteString("\n\n")
	output.WriteString(modal)
	output.WriteString("\n")

	return output.String()
}

// renderCommentsView renders the full-screen comments view
func (m Model) renderCommentsView() string {
	selected := m.SelectedIssue()
	if selected == nil {
		return m.theme.NoIssuesStyle.Render("No issue selected")
	}

	var b strings.Builder

	// Header with issue info
	b.WriteString(m.theme.CommentsHeaderStyle.Render(fmt.Sprintf("#%d - %s", selected.Number, selected.Title)))
	b.WriteString("\n")
	b.WriteString(m.theme.CommentsMetaStyle.Render(fmt.Sprintf("State: %s | Author: %s | Comments: %d", selected.State, selected.Author, len(m.currentComments))))
	b.WriteString("\n\n")

	// Render comments
	if len(m.currentComments) == 0 {
		b.WriteString(m.theme.CommentsMetaStyle.Render("No comments on this issue"))
		b.WriteString("\n\n")
	} else {
		// Build full content first
		var contentBuilder strings.Builder
		for i, comment := range m.currentComments {
			// Comment header
			contentBuilder.WriteString(m.theme.CommentAuthorStyle.Render(fmt.Sprintf("@%s", comment.Author)))
			contentBuilder.WriteString(" • ")
			contentBuilder.WriteString(m.theme.CommentsMetaStyle.Render(formatRelativeTime(comment.CreatedAt)))
			contentBuilder.WriteString("\n")

			// Comment body
			body := comment.Body
			if body == "" {
				body = "(No comment body)"
			}

			// Render markdown or show raw
			if m.showRawMarkdown {
				contentBuilder.WriteString(body)
			} else {
				// Use glamour to render markdown
				rendered, err := renderMarkdown(body, m.width-4) // -4 for padding
				if err != nil {
					// Fall back to raw if rendering fails
					contentBuilder.WriteString(body)
				} else {
					contentBuilder.WriteString(rendered)
				}
			}

			// Separator between comments (except last one)
			if i < len(m.currentComments)-1 {
				contentBuilder.WriteString("\n")
				contentBuilder.WriteString(m.theme.CommentSeparatorStyle.Render(strings.Repeat("─", min(m.width, 80))))
				contentBuilder.WriteString("\n\n")
			}
		}

		// Apply scrolling
		content := contentBuilder.String()
		lines := strings.Split(content, "\n")

		// Calculate visible range
		contentHeight := m.height - 6 // Leave room for header, status, footer
		if contentHeight < 1 {
			contentHeight = 1
		}

		startLine := m.commentsScrollOffset
		if startLine >= len(lines) {
			startLine = len(lines) - 1
		}
		if startLine < 0 {
			startLine = 0
		}

		endLine := startLine + contentHeight
		if endLine > len(lines) {
			endLine = len(lines)
		}

		visibleLines := lines[startLine:endLine]
		b.WriteString(strings.Join(visibleLines, "\n"))
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	statusText := fmt.Sprintf("Comments View | Scroll: %d", m.commentsScrollOffset)
	if len(m.currentComments) > 0 {
		statusText += fmt.Sprintf(" | Total: %d", len(m.currentComments))
	}
	b.WriteString(m.theme.StatusStyle.Render(statusText))

	// Footer
	b.WriteString("\n")
	b.WriteString(m.theme.FooterStyle.Render(m.renderFooter()))

	return b.String()
}

// renderHelpView renders the help overlay with all keybindings
func (m Model) renderHelpView() string {
	var b strings.Builder

	// Title
	b.WriteString(m.theme.HelpTitleStyle.Render("Keybinding Help"))
	b.WriteString("\n\n")

	// Determine which context to show based on previous view
	var context string
	switch m.previousViewMode {
	case viewModeList:
		context = "Issue List View"
	case viewModeComments:
		context = "Comments View"
	default:
		context = "Issue List View"
	}

	b.WriteString(m.theme.HelpSectionStyle.Render("Current Context: " + context))
	b.WriteString("\n\n")

	// Navigation section
	b.WriteString(m.theme.HelpSectionStyle.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  j, ↓        "))
	b.WriteString(m.theme.HelpDescStyle.Render("Move down (list view)"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  k, ↑        "))
	b.WriteString(m.theme.HelpDescStyle.Render("Move up (list view)"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  Enter       "))
	b.WriteString(m.theme.HelpDescStyle.Render("Open comments view for selected issue"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  Esc, q      "))
	b.WriteString(m.theme.HelpDescStyle.Render("Return to list view (from comments)"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  PgUp/PgDn   "))
	b.WriteString(m.theme.HelpDescStyle.Render("Scroll detail panel or comments"))
	b.WriteString("\n\n")

	// Sorting section
	b.WriteString(m.theme.HelpSectionStyle.Render("Sorting"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  s           "))
	b.WriteString(m.theme.HelpDescStyle.Render("Cycle sort field (updated → created → number → comments)"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  S           "))
	b.WriteString(m.theme.HelpDescStyle.Render("Reverse sort order (ascending ↔ descending)"))
	b.WriteString("\n\n")

	// View options section
	b.WriteString(m.theme.HelpSectionStyle.Render("View Options"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  m           "))
	b.WriteString(m.theme.HelpDescStyle.Render("Toggle markdown rendering (raw ↔ rendered)"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  ?           "))
	b.WriteString(m.theme.HelpDescStyle.Render("Show/hide this help"))
	b.WriteString("\n\n")

	// Application section
	b.WriteString(m.theme.HelpSectionStyle.Render("Application"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  q           "))
	b.WriteString(m.theme.HelpDescStyle.Render("Quit application (from list view)"))
	b.WriteString("\n")
	b.WriteString(m.theme.HelpKeyStyle.Render("  Ctrl+C      "))
	b.WriteString(m.theme.HelpDescStyle.Render("Quit application (from any view)"))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.theme.HelpFooterStyle.Render("Press ? or Esc to close this help"))

	return m.theme.HelpOverlayStyle.Render(b.String())
}

// renderListPanel renders the left panel with issue list
func (m Model) renderListPanel(width, height int) string {
	var b strings.Builder

	// Header
	b.WriteString(m.theme.HeaderStyle.Render(m.renderHeader()))
	b.WriteString("\n\n")

	// Issue list (limit to height)
	linesRendered := 2 // header + blank line
	for i, issue := range m.issues {
		if linesRendered >= height {
			break
		}

		cursor := " "
		style := m.theme.NormalStyle
		if i == m.cursor {
			cursor = ">"
			style = m.theme.SelectedStyle
		}

		line := fmt.Sprintf("%s %s", cursor, m.renderIssue(issue))
		// Truncate to width
		if len(line) > width {
			line = line[:width]
		}
		b.WriteString(style.Render(line))
		b.WriteString("\n")
		linesRendered++
	}

	return b.String()
}

// renderDetailPanel renders the right panel with issue details
func (m Model) renderDetailPanel(width, height int) string {
	selected := m.SelectedIssue()
	if selected == nil {
		return m.theme.DetailPanelStyle.Width(width).Height(height).Render("No issue selected")
	}

	var b strings.Builder

	// Header: issue number, title, state
	b.WriteString(m.theme.DetailHeaderStyle.Render(fmt.Sprintf("#%d • %s", selected.Number, selected.State)))
	b.WriteString("\n")
	b.WriteString(m.theme.DetailTitleStyle.Render(selected.Title))
	b.WriteString("\n\n")

	// Metadata: author, dates
	b.WriteString(m.theme.DetailMetaStyle.Render(fmt.Sprintf("Author: %s", selected.Author)))
	b.WriteString("\n")
	b.WriteString(m.theme.DetailMetaStyle.Render(fmt.Sprintf("Created: %s", formatRelativeTime(selected.CreatedAt))))
	b.WriteString("\n")
	b.WriteString(m.theme.DetailMetaStyle.Render(fmt.Sprintf("Updated: %s", formatRelativeTime(selected.UpdatedAt))))
	b.WriteString("\n")

	// Labels (if any)
	if len(selected.Labels) > 0 {
		b.WriteString(m.theme.DetailMetaStyle.Render(fmt.Sprintf("Labels: %s", strings.Join(selected.Labels, ", "))))
		b.WriteString("\n")
	}

	// Assignees (if any)
	if len(selected.Assignees) > 0 {
		b.WriteString(m.theme.DetailMetaStyle.Render(fmt.Sprintf("Assignees: %s", strings.Join(selected.Assignees, ", "))))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Body
	bodyContent := selected.Body
	if bodyContent == "" {
		bodyContent = "(No description provided)"
	}

	// Render markdown or show raw
	if m.showRawMarkdown {
		b.WriteString(bodyContent)
	} else {
		// Use glamour to render markdown
		rendered, err := renderMarkdown(bodyContent, width)
		if err != nil {
			// Fall back to raw if rendering fails
			b.WriteString(bodyContent)
		} else {
			b.WriteString(rendered)
		}
	}

	// Split into lines for scrolling
	content := b.String()
	lines := strings.Split(content, "\n")

	// Apply scroll offset
	startLine := m.detailScrollOffset
	if startLine >= len(lines) {
		startLine = len(lines) - 1
	}
	if startLine < 0 {
		startLine = 0
	}

	endLine := startLine + height
	if endLine > len(lines) {
		endLine = len(lines)
	}

	scrolledLines := lines[startLine:endLine]

	// Truncate each line to width
	for i, line := range scrolledLines {
		if len(line) > width {
			scrolledLines[i] = line[:width]
		}
	}

	return strings.Join(scrolledLines, "\n")
}

// renderMarkdown renders markdown content using glamour
func renderMarkdown(content string, width int) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return "", err
	}

	rendered, err := r.Render(content)
	if err != nil {
		return "", err
	}

	return rendered, nil
}

// renderHeader renders the column headers
func (m Model) renderHeader() string {
	var parts []string
	for _, col := range m.columns {
		switch col {
		case "number":
			parts = append(parts, padRight("#", 8))
		case "title":
			parts = append(parts, padRight("Title", 50))
		case "author":
			parts = append(parts, padRight("Author", 20))
		case "date":
			parts = append(parts, padRight("Updated", 20))
		case "comments":
			parts = append(parts, padRight("Comments", 10))
		}
	}
	return strings.Join(parts, " ")
}

// renderIssue renders a single issue row
func (m Model) renderIssue(issue *sync.Issue) string {
	var parts []string
	for _, col := range m.columns {
		switch col {
		case "number":
			parts = append(parts, padRight(fmt.Sprintf("#%d", issue.Number), 8))
		case "title":
			parts = append(parts, padRight(truncate(issue.Title, 48), 50))
		case "author":
			parts = append(parts, padRight(truncate(issue.Author, 18), 20))
		case "date":
			parts = append(parts, padRight(formatRelativeTime(issue.UpdatedAt), 20))
		case "comments":
			parts = append(parts, padRight(fmt.Sprintf("%d", issue.CommentCount), 10))
		}
	}
	return strings.Join(parts, " ")
}

// renderFooter renders context-sensitive footer with keybindings
func (m Model) renderFooter() string {
	switch m.viewMode {
	case viewModeList:
		// Split-pane view or simple list
		if m.width > 0 && m.height > 0 {
			return "j/k, ↑/↓: navigate • Enter: comments • PgUp/PgDn: scroll • m: markdown • s: sort • S: reverse • ?: help • q: quit"
		}
		return "j/k, ↑/↓: navigate • s: sort • S: reverse • ?: help • q: quit"
	case viewModeComments:
		return "PgUp/PgDn: scroll • m: markdown • ?: help • Esc/q: back"
	case viewModeHelp:
		return "Press ? or Esc to close help"
	default:
		return ""
	}
}

// renderStatus renders the status bar
func (m Model) renderStatus() string {
	total := len(m.issues)
	current := m.cursor + 1

	// Build sort description
	sortOrder := "desc"
	if m.sortAscending {
		sortOrder = "asc"
	}
	sortDesc := fmt.Sprintf("sort: %s (%s)", m.sortBy, sortOrder)

	// Build last synced indicator
	lastSynced := fmt.Sprintf("Last synced: %s", formatRelativeTime(m.lastSyncTime))

	// Build base status
	var baseStatus string
	if total == 0 {
		baseStatus = fmt.Sprintf("No issues • %s • %s", sortDesc, lastSynced)
	} else {
		baseStatus = fmt.Sprintf("Issue %d of %d • %s • %s", current, total, sortDesc, lastSynced)
	}

	// Append status error if present
	if m.statusError != "" {
		return baseStatus + " • " + m.theme.ErrorStyle.Render("Error: "+m.statusError)
	}

	return baseStatus
}

// SelectedIssue returns the currently selected issue
func (m Model) SelectedIssue() *sync.Issue {
	if len(m.issues) == 0 {
		return nil
	}
	return m.issues[m.cursor]
}

// sortIssues sorts the issues slice based on sortBy and sortAscending fields
func (m *Model) sortIssues() {
	if len(m.issues) == 0 {
		return
	}

	// Use a stable sort to maintain order for equal elements
	sortFunc := func(i, j int) bool {
		var less bool
		switch m.sortBy {
		case "updated":
			less = m.issues[i].UpdatedAt.Before(m.issues[j].UpdatedAt)
		case "created":
			less = m.issues[i].CreatedAt.Before(m.issues[j].CreatedAt)
		case "number":
			less = m.issues[i].Number < m.issues[j].Number
		case "comments":
			less = m.issues[i].CommentCount < m.issues[j].CommentCount
		default:
			// Default to updated
			less = m.issues[i].UpdatedAt.Before(m.issues[j].UpdatedAt)
		}

		// If ascending, use the comparison as-is; if descending, invert it
		if m.sortAscending {
			return less
		}
		return !less
	}

	// Sort issues in place
	for i := 0; i < len(m.issues)-1; i++ {
		for j := i + 1; j < len(m.issues); j++ {
			if sortFunc(j, i) {
				m.issues[i], m.issues[j] = m.issues[j], m.issues[i]
			}
		}
	}
}

// formatRelativeTime converts a timestamp to a human-readable relative time string
func formatRelativeTime(t time.Time) string {
	// Handle zero time (never synced)
	if t.IsZero() {
		return "never"
	}

	elapsed := time.Since(t)

	// Less than 1 minute
	if elapsed < time.Minute {
		return "just now"
	}

	// Minutes
	if elapsed < time.Hour {
		minutes := int(elapsed.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	// Hours
	if elapsed < 24*time.Hour {
		hours := int(elapsed.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Days
	if elapsed < 7*24*time.Hour {
		days := int(elapsed.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	// Weeks
	weeks := int(elapsed.Hours() / 24 / 7)
	if weeks == 1 {
		return "1 week ago"
	}
	return fmt.Sprintf("%d weeks ago", weeks)
}

// Helper functions

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
