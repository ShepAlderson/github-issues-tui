package list

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shepbook/ghissues/internal/comments"
	"github.com/shepbook/ghissues/internal/database"
	"github.com/shepbook/ghissues/internal/detail"
)

// Config interface for accessing configuration
type Config interface {
	GetDisplayColumns() []string
	GetDefaultRepository() string
	GetSortField() string
	GetSortDescending() bool
	SaveSort(field string, descending bool) error
}

// Model represents the issue list TUI state
type Model struct {
	dbPath     string
	repo       string
	columns    []string
	issues     []database.ListIssue
	selected   int
	width      int
	height     int
	db         *sql.DB
	sortField  string
	sortDesc   bool
	sortFields []string
	configPath string
	// saveSort is a callback to persist sort settings
	saveSort func(field string, descending bool) error
	// openComments is a callback to open comments view for an issue
	openComments func(issueNumber int, issueTitle string)
	// detail fields
	showDetail   bool
	detailModel  *detail.Model
	detailIssue  *database.IssueDetail
	renderedMode bool
	// comments fields
	// showingComments indicates if we're in the comments view
	showingComments     bool
	commentsModel       *comments.Model
	commentsOpenPending bool
	// refresh fields
	refreshing      bool
	refreshPending  bool
	refreshProgress string
	token           string
}

// Styles for the list view
var (
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	normalStyle = lipgloss.NewStyle()

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)

// NewModel creates a new list model
func NewModel(cfg Config, dbPath, configPath string) Model {
	columns := validateColumns(cfg.GetDisplayColumns())
	if len(columns) == 0 {
		columns = []string{"number", "title", "author", "updated", "comments"}
	}

	// Validate sort field from config
	sortField := cfg.GetSortField()
	if sortField == "" {
		sortField = "updated"
	}

	return Model{
		dbPath:              dbPath,
		repo:                cfg.GetDefaultRepository(),
		columns:             columns,
		issues:              []database.ListIssue{},
		selected:            0,
		width:               80,
		height:              24,
		sortField:           sortField,
		sortDesc:            cfg.GetSortDescending(),
		sortFields:          []string{"updated", "created", "number", "comments"},
		configPath:          configPath,
		saveSort:            cfg.SaveSort,
		showDetail:          true, // Default to showing detail panel
		renderedMode:        true, // Default to rendered markdown
		openComments:        nil,  // Will be set by caller
		showingComments:     false,
		commentsModel:       nil,
		commentsOpenPending: false,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return m.loadIssues()
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if m.selected > 0 {
				m.selected--
			}
		case tea.KeyDown:
			if m.selected < len(m.issues)-1 {
				m.selected++
			}
		case tea.KeyRunes:
			switch msg.String() {
			case "q", "Q":
				return m, tea.Quit
			case "k":
				if m.selected > 0 {
					m.selected--
					m.updateDetailIssue()
				}
			case "j":
				if m.selected < len(m.issues)-1 {
					m.selected++
					m.updateDetailIssue()
				}
			case "s":
				// Cycle to next sort field
				m.cycleSortField()
				// Save sort preference and reload issues
				return m, tea.Batch(
					m.saveSortConfig(),
					m.loadIssues(),
				)
			case "S":
				// Toggle sort order
				m.sortDesc = !m.sortDesc
				// Save sort preference and reload issues
				return m, tea.Batch(
					m.saveSortConfig(),
					m.loadIssues(),
				)
			case "m", "M":
				// Toggle markdown mode
				m.renderedMode = !m.renderedMode
				if m.detailModel != nil {
					m.detailModel.ToggleRenderedMode()
				}
			case "r", "R":
				// Trigger refresh
				m.refreshPending = true
				return m, m.startRefresh()
			}
		case tea.KeyEnter:
			// Open comments view for selected issue
			if m.selected >= 0 && m.selected < len(m.issues) {
				m.commentsOpenPending = true
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case issuesLoadedMsg:
		m.issues = msg.issues
		if m.selected >= len(m.issues) {
			m.selected = 0
		}
		// Update detail view for selected issue
		if len(m.issues) > 0 && m.selected < len(m.issues) {
			cmd := m.loadDetailIssue()
			return m, cmd
		}

	case detailLoadedMsg:
		if msg.err != nil {
			// Detail failed to load, clear detail model
			m.detailModel = nil
			m.detailIssue = nil
		} else {
			m.detailIssue = msg.issue
			// Create/update detail model
			if m.detailIssue != nil {
				// Calculate detail panel dimensions
				listPanelWidth := m.calculateListPanelWidth()
				detailWidth := m.width - listPanelWidth - 2
				detailHeight := m.height - 2
				if detailWidth < 20 {
					detailWidth = 20
				}
				if detailHeight < 5 {
					detailHeight = 5
				}

				detailModel := detail.NewModel(*m.detailIssue, detailWidth, detailHeight)
				detailModel.RenderedMode = m.renderedMode
				m.detailModel = &detailModel
			}
		}
	}

	return m, nil
}

// View renders the split UI with issue list and detail panel
func (m Model) View() string {
	if !m.showDetail || m.detailModel == nil {
		return m.renderListOnlyView()
	}
	return m.renderSplitView()
}

// renderListOnlyView renders just the issue list (when detail not available)
func (m Model) renderListOnlyView() string {
	var b strings.Builder

	// Title
	b.WriteString(headerStyle.Render(fmt.Sprintf("📋 %s", m.repo)))
	b.WriteString("\n\n")

	// Calculate available height
	headerLines := 3
	statusLines := 2
	availableHeight := m.height - headerLines - statusLines

	if availableHeight < 5 {
		availableHeight = 10
	}

	// Issue list
	if len(m.issues) == 0 {
		b.WriteString("  No issues found. Run 'ghissues sync' to fetch issues.\n")
	} else {
		for i, issue := range m.issues {
			if i >= availableHeight {
				break
			}
			line := m.renderIssueLine(issue, i == m.selected)
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Status bar
	b.WriteString("\n")
	orderIcon := "↓"
	if !m.sortDesc {
		orderIcon = "↑"
	}
	status := fmt.Sprintf("%d issues | sort:%s %s | j/k to navigate | s to sort | q to quit", len(m.issues), m.sortField, orderIcon)
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n")

	return b.String()
}

// renderSplitView renders the split layout with list and detail
func (m Model) renderSplitView() string {
	listWidth := m.calculateListPanelWidth()
	detailWidth := m.calculateDetailPanelWidth()

	// Calculate heights
	headerHeight := 3
	statusHeight := 2
	contentHeight := m.height - headerHeight - statusHeight

	if contentHeight < 5 {
		contentHeight = 10
	}

	// Build list panel
	var listBuilder strings.Builder

	// Title
	listBuilder.WriteString(headerStyle.Render(fmt.Sprintf("📋 %s", m.repo)))
	listBuilder.WriteString("\n\n")

	// Issue list
	listContentHeight := contentHeight - 2 // Account for title
	if len(m.issues) == 0 {
		listBuilder.WriteString("  No issues found.\n")
	} else {
		for i, issue := range m.issues {
			if i >= listContentHeight {
				break
			}
			line := m.renderIssueLine(issue, i == m.selected)
			listBuilder.WriteString(line)
			listBuilder.WriteString("\n")
		}
	}

	// Status bar for list
	listBuilder.WriteString("\n")
	orderIcon := "↓"
	if !m.sortDesc {
		orderIcon = "↑"
	}
	status := fmt.Sprintf("%d issues | sort:%s %s | m markdown | r refresh | enter comments | q quit", len(m.issues), m.sortField, orderIcon)
	listBuilder.WriteString(statusStyle.Render(status))

	// Style the list panel with border
	listStyle := lipgloss.NewStyle().
		Width(listWidth).
		Height(m.height).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRight(true).
		Padding(0, 1)

	listPanel := listStyle.Render(listBuilder.String())

	// Build detail panel
	detailPanel := ""
	if m.detailModel != nil {
		// Set dimensions for detail model
		m.detailModel.SetDimensions(detailWidth-2, contentHeight)
		detailContent := m.detailModel.View()

		detailStyle := lipgloss.NewStyle().
			Width(detailWidth-2).
			Height(m.height).
			Padding(0, 1)

		detailPanel = detailStyle.Render(detailContent)
	}

	// Join panels horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
}

// renderIssueLine renders a single issue line
func (m Model) renderIssueLine(issue database.ListIssue, isSelected bool) string {
	columns := renderColumns(issue, m.columns)

	// Join columns with spacing
	line := strings.Join(columns, "  ")

	// Truncate if too long
	maxWidth := m.width - 2 // Leave room for selection indicator
	if lipgloss.Width(line) > maxWidth {
		line = lipgloss.NewStyle().MaxWidth(maxWidth).Render(line)
	}

	// Add selection indicator
	if isSelected {
		return selectedStyle.Render("> " + line)
	}
	return normalStyle.Render("  " + line)
}

// renderColumns extracts and formats the requested columns from an issue
func renderColumns(issue database.ListIssue, columns []string) []string {
	var result []string

	for _, col := range columns {
		var value string
		switch col {
		case "number":
			value = fmt.Sprintf("#%d", issue.Number)
		case "title":
			value = issue.Title
		case "author":
			value = issue.Author
		case "created":
			value = database.FormatDate(issue.CreatedAt)
		case "updated":
			value = database.FormatDate(issue.UpdatedAt)
		case "comments":
			if issue.CommentCount > 0 {
				value = fmt.Sprintf("💬 %d", issue.CommentCount)
			} else {
				value = ""
			}
		case "state":
			value = issue.State
		default:
			continue // Skip unknown columns
		}
		result = append(result, value)
	}

	return result
}

// validateColumns filters out invalid column names
func validateColumns(columns []string) []string {
	valid := map[string]bool{
		"number":   true,
		"title":    true,
		"author":   true,
		"created":  true,
		"updated":  true,
		"comments": true,
		"state":    true,
	}

	var result []string
	for _, col := range columns {
		if valid[col] {
			result = append(result, col)
		}
	}
	return result
}

// issuesLoadedMsg is sent when issues are loaded from the database
type issuesLoadedMsg struct {
	issues []database.ListIssue
}

// loadIssues loads issues from the database with current sort settings
func (m Model) loadIssues() tea.Cmd {
	return func() tea.Msg {
		db, err := database.InitializeSchema(m.dbPath)
		if err != nil {
			return issuesLoadedMsg{issues: []database.ListIssue{}}
		}

		issues, err := database.ListIssuesSorted(db, m.repo, m.sortField, m.sortDesc)
		if err != nil {
			db.Close()
			return issuesLoadedMsg{issues: []database.ListIssue{}}
		}

		return issuesLoadedMsg{issues: issues}
	}
}

// Selected returns the currently selected issue
func (m Model) Selected() *database.ListIssue {
	if m.selected < 0 || m.selected >= len(m.issues) {
		return nil
	}
	return &m.issues[m.selected]
}

// updateDetailIssue fetches the full issue details for the selected issue
func (m *Model) updateDetailIssue() {
	if m.selected < 0 || m.selected >= len(m.issues) {
		m.detailModel = nil
		m.detailIssue = nil
		return
	}

	// Defer fetching to a command so we don't block
	// The actual fetch will be done via a command in loadDetailIssue
}

// loadDetailIssue returns a command to fetch issue details
func (m *Model) loadDetailIssue() tea.Cmd {
	if m.selected < 0 || m.selected >= len(m.issues) {
		return nil
	}

	issueNumber := m.issues[m.selected].Number
	return func() tea.Msg {
		db, err := database.InitializeSchema(m.dbPath)
		if err != nil {
			return detailLoadedMsg{err: err}
		}
		defer db.Close()

		issue, err := database.GetIssueDetail(db, m.repo, issueNumber)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		return detailLoadedMsg{issue: issue}
	}
}

// detailLoadedMsg is sent when issue detail is loaded
type detailLoadedMsg struct {
	issue *database.IssueDetail
	err   error
}

// ShouldOpenComments returns true if the user requested to open comments view
func (m Model) ShouldOpenComments() bool {
	return m.commentsOpenPending
}

// GetSelectedIssueForComments returns the selected issue number and title
// Call ResetCommentsPending() after using this information
func (m Model) GetSelectedIssueForComments() (int, string, bool) {
	if m.selected < 0 || m.selected >= len(m.issues) {
		return 0, "", false
	}
	issue := m.issues[m.selected]
	return issue.Number, issue.Title, true
}

// ResetCommentsPending resets the comments pending flag
func (m *Model) ResetCommentsPending() {
	m.commentsOpenPending = false
}

// SetDimensions updates the model dimensions
func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	// Update detail model dimensions if it exists
	if m.detailModel != nil {
		listPanelWidth := m.calculateListPanelWidth()
		detailWidth := m.width - listPanelWidth - 2
		detailHeight := m.height - 2
		if detailWidth < 20 {
			detailWidth = 20
		}
		if detailHeight < 5 {
			detailHeight = 5
		}
		m.detailModel.SetDimensions(detailWidth, detailHeight)
	}
}

// calculateListPanelWidth returns the width for the list panel
func (m Model) calculateListPanelWidth() int {
	// List panel takes 40% of width, min 30, max 50
	listWidth := int(float64(m.width) * 0.4)
	if listWidth < 30 {
		listWidth = 30
	}
	if listWidth > 50 {
		listWidth = 50
	}
	return listWidth
}

// calculateDetailPanelWidth returns the width for the detail panel
func (m Model) calculateDetailPanelWidth() int {
	listWidth := m.calculateListPanelWidth()
	return m.width - listWidth - 2 // Account for separator
}

// cycleSortField moves to the next sort field in the cycle
func (m *Model) cycleSortField() {
	for i, field := range m.sortFields {
		if field == m.sortField {
			// Move to next field
			nextIndex := (i + 1) % len(m.sortFields)
			m.sortField = m.sortFields[nextIndex]
			// Reset to descending when changing field
			m.sortDesc = true
			return
		}
	}
	// If current sort field not in list, start from beginning
	m.sortField = m.sortFields[0]
	m.sortDesc = true
}

// GetSortField returns the current sort field
func (m Model) GetSortField() string {
	return m.sortField
}

// GetSortDescending returns whether sort is descending
func (m Model) GetSortDescending() bool {
	return m.sortDesc
}

// saveSortConfig returns a command that saves the sort configuration
func (m Model) saveSortConfig() tea.Cmd {
	return func() tea.Msg {
		if m.saveSort != nil {
			if err := m.saveSort(m.sortField, m.sortDesc); err != nil {
				// Log error but don't fail - we can continue without persistence
				return sortSavedMsg{err: err}
			}
		}
		return sortSavedMsg{}
	}
}

// sortSavedMsg indicates that sort preferences were saved (or failed)
type sortSavedMsg struct {
	err error
}

// IsRefreshing returns whether the model is currently refreshing
func (m Model) IsRefreshing() bool {
	return m.refreshing
}

// SetRefreshing sets the refreshing state
func (m *Model) SetRefreshing(refreshing bool) {
	m.refreshing = refreshing
}

// SetToken sets the GitHub token for refresh operations
func (m *Model) SetToken(token string) {
	m.token = token
}

// ShouldRefresh returns true if the user requested a refresh
func (m Model) ShouldRefresh() bool {
	return m.refreshPending
}

// ResetRefresh resets the refresh pending flag
func (m *Model) ResetRefresh() {
	m.refreshPending = false
}

// startRefresh starts the refresh process
func (m Model) startRefresh() tea.Cmd {
	return func() tea.Msg {
		// This is a placeholder - the actual refresh will be handled
		// by the main loop which has access to the GitHub client
		m.refreshing = true
		return refreshStartedMsg{}
	}
}

// refreshStartedMsg is sent when refresh starts
type refreshStartedMsg struct{}
