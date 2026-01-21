package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

// IssueList represents the issue list component
type IssueList struct {
	dbManager      *database.DBManager
	config         *config.Config
	configManager  *config.Manager
	columnRenderer *ColumnRenderer
	issues         []IssueItem
	selectedIdx    int
	width          int
	height         int
	sortField      string
	sortAscending  bool
}

// IssueItem represents an item in the issue list
type IssueItem struct {
	Number       int
	TitleText    string
	Author       string
	Created      time.Time
	Updated      time.Time
	CommentCount int
}

// NewIssueList creates a new issue list component
func NewIssueList(dbManager *database.DBManager, cfg *config.Config, cfgMgr *config.Manager) *IssueList {
	// Create column renderer from config
	var columnKeys []string
	if cfg.Display.Columns != nil {
		columnKeys = cfg.Display.Columns
	} else {
		// Use defaults if not specified
		columnKeys = []string{"number", "title", "author", "date", "comments"}
	}

	columnRenderer := NewColumnRenderer(columnKeys)

	// Get sort configuration from config
	sortField := "updated"
	sortAscending := false
	if cfg.Display.SortField != "" {
		sortField = cfg.Display.SortField
	}
	// Note: bool defaults to false, so we only need to set if true
	sortAscending = cfg.Display.SortAscending

	return &IssueList{
		dbManager:      dbManager,
		config:         cfg,
		configManager:  cfgMgr,
		columnRenderer: columnRenderer,
		selectedIdx:    0,
		sortField:      sortField,
		sortAscending:  sortAscending,
	}
}

// Init loads issues from the database
func (il *IssueList) Init() tea.Cmd {
	return il.loadIssues()
}

// Update handles messages for the issue list
func (il *IssueList) Update(msg tea.Msg) (*IssueList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		il.width = msg.Width
		il.height = msg.Height
		return il, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if il.selectedIdx < len(il.issues)-1 {
				il.selectedIdx++
			}
			return il, nil
		case "k", "up":
			if il.selectedIdx > 0 {
				il.selectedIdx--
			}
			return il, nil
		case "s":
			il.cycleSortField()
			return il, nil
		case "S":
			il.toggleSortOrder()
			return il, nil
		}
	}

	return il, nil
}

// View renders the issue list
func (il *IssueList) View() string {
	if len(il.issues) == 0 {
		return "No issues found"
	}

	var builder strings.Builder

	// Render header
	builder.WriteString(il.columnRenderer.RenderHeader())
	builder.WriteString("\n")

	// Render separator
	builder.WriteString(strings.Repeat("─", il.columnRenderer.TotalWidth()))
	builder.WriteString("\n")

	// Calculate visible range based on height
	headerHeight := 2 // header + separator
	availableHeight := il.height - headerHeight
	if availableHeight <= 0 {
		availableHeight = 10 // default minimum
	}

	startIdx := max(0, il.selectedIdx-(availableHeight/2))
	endIdx := min(len(il.issues), startIdx+availableHeight)

	// Render visible issues
	for i := startIdx; i < endIdx; i++ {
		issue := il.issues[i]
		selected := i == il.selectedIdx
		row := il.columnRenderer.RenderIssue(issue, selected)
		builder.WriteString(row)
		builder.WriteString("\n")
	}

	// Add status bar with issue count, sort status, and last sync time
	lastSyncText := il.getLastSyncText()
	status := fmt.Sprintf("Issues: %d/%d | Sort: %s | Last synced: %s", il.selectedIdx+1, len(il.issues), il.sortStatus(), lastSyncText)
	builder.WriteString("\n")
	builder.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(status))

	// Add footer hints
	builder.WriteString("\n\n")
	builder.WriteString(GetFooterHints(FooterContextList))

	return builder.String()
}

// loadIssues loads issues from the database
func (il *IssueList) loadIssues() tea.Cmd {
	if il.dbManager == nil {
		// Fall back to dummy data if no database manager
		return il.loadDummyIssues()
	}

	dbIssues, err := il.dbManager.GetAllIssues()
	if err != nil {
		// Fall back to dummy data on error
		return il.loadDummyIssues()
	}

	// Convert database issues to IssueItem structs
	var issues []IssueItem
	for _, dbIssue := range dbIssues {
		issues = append(issues, IssueItem{
			Number:       dbIssue.Number,
			TitleText:    dbIssue.Title,
			Author:       dbIssue.Author,
			Created:      dbIssue.CreatedAt,
			Updated:      dbIssue.UpdatedAt,
			CommentCount: dbIssue.CommentCount,
		})
	}

	il.issues = issues
	il.sortIssues()
	return nil
}

// loadDummyIssues loads dummy issues for testing (fallback)
func (il *IssueList) loadDummyIssues() tea.Cmd {
	// Create dummy issues for testing
	now := time.Now()
	dummyIssues := []IssueItem{
		{Number: 1, TitleText: "Fix bug in login flow", Author: "alice", Created: now.Add(-24 * time.Hour), Updated: now.Add(-2 * time.Hour), CommentCount: 3},
		{Number: 2, TitleText: "Add dark mode support", Author: "bob", Created: now.Add(-48 * time.Hour), Updated: now.Add(-4 * time.Hour), CommentCount: 0},
		{Number: 3, TitleText: "Improve performance", Author: "charlie", Created: now.Add(-72 * time.Hour), Updated: now.Add(-6 * time.Hour), CommentCount: 7},
		{Number: 4, TitleText: "Update documentation", Author: "david", Created: now.Add(-96 * time.Hour), Updated: now.Add(-8 * time.Hour), CommentCount: 2},
		{Number: 5, TitleText: "Refactor API endpoints", Author: "eve", Created: now.Add(-120 * time.Hour), Updated: now.Add(-10 * time.Hour), CommentCount: 1},
	}

	il.issues = dummyIssues
	il.sortIssues()
	return nil
}

// SelectedIssue returns the currently selected issue
func (il *IssueList) SelectedIssue() *IssueItem {
	if len(il.issues) == 0 {
		return nil
	}

	if il.selectedIdx < 0 || il.selectedIdx >= len(il.issues) {
		return nil
	}

	issue := il.issues[il.selectedIdx]
	return &issue
}

// IssueCount returns the total number of issues
func (il *IssueList) IssueCount() int {
	return len(il.issues)
}

// sortIssues sorts the issues based on current sort configuration
func (il *IssueList) sortIssues() {
	// Sort using the appropriate comparison function
	switch il.sortField {
	case "updated":
		il.sortByUpdated()
	case "created":
		il.sortByCreated()
	case "number":
		il.sortByNumber()
	case "comments":
		il.sortByComments()
	default:
		// Default to updated date
		il.sortByUpdated()
	}
}

// sortByUpdated sorts issues by updated date
func (il *IssueList) sortByUpdated() {
	sort.Slice(il.issues, func(i, j int) bool {
		if il.sortAscending {
			return il.issues[i].Updated.Before(il.issues[j].Updated)
		}
		return il.issues[i].Updated.After(il.issues[j].Updated)
	})
}

// sortByCreated sorts issues by created date
func (il *IssueList) sortByCreated() {
	sort.Slice(il.issues, func(i, j int) bool {
		if il.sortAscending {
			return il.issues[i].Created.Before(il.issues[j].Created)
		}
		return il.issues[i].Created.After(il.issues[j].Created)
	})
}

// sortByNumber sorts issues by issue number
func (il *IssueList) sortByNumber() {
	sort.Slice(il.issues, func(i, j int) bool {
		if il.sortAscending {
			return il.issues[i].Number < il.issues[j].Number
		}
		return il.issues[i].Number > il.issues[j].Number
	})
}

// sortByComments sorts issues by comment count
func (il *IssueList) sortByComments() {
	sort.Slice(il.issues, func(i, j int) bool {
		if il.sortAscending {
			return il.issues[i].CommentCount < il.issues[j].CommentCount
		}
		return il.issues[i].CommentCount > il.issues[j].CommentCount
	})
}

// cycleSortField cycles through available sort fields
func (il *IssueList) cycleSortField() {
	// Define the order of sort fields
	sortFields := []string{"updated", "created", "number", "comments"}

	// Find current field index
	currentIdx := -1
	for i, field := range sortFields {
		if field == il.sortField {
			currentIdx = i
			break
		}
	}

	// If not found, start from beginning
	if currentIdx == -1 {
		il.sortField = sortFields[0]
	} else {
		// Move to next field, wrapping around
		nextIdx := (currentIdx + 1) % len(sortFields)
		il.sortField = sortFields[nextIdx]
	}

	// Update config and save
	il.updateConfigAndSave()

	// Re-sort issues with new field
	il.sortIssues()
}

// toggleSortOrder toggles between ascending and descending order
func (il *IssueList) toggleSortOrder() {
	il.sortAscending = !il.sortAscending

	// Update config and save
	il.updateConfigAndSave()

	il.sortIssues()
}

// sortStatus returns a string representation of current sort state
func (il *IssueList) sortStatus() string {
	arrow := "↑"
	if il.sortAscending {
		arrow = "↓"
	}
	return fmt.Sprintf("%s %s", arrow, il.sortField)
}

// updateConfigAndSave updates the config with current sort settings and saves to disk
func (il *IssueList) updateConfigAndSave() {
	// Update config in memory
	il.config.Display.SortField = il.sortField
	il.config.Display.SortAscending = il.sortAscending

	// Save to disk if config manager is available
	if il.configManager != nil {
		// Ignore error for now - in production we might want to log it
		_ = il.configManager.Save(il.config)
	}
}

// getLastSyncText returns a formatted string for the last sync time
func (il *IssueList) getLastSyncText() string {
	if il.dbManager == nil {
		return "unknown"
	}

	lastSyncTime, err := il.dbManager.GetLastSyncTime()
	if err != nil {
		// Log error but don't crash - return generic message
		return "unknown"
	}

	return FormatRelativeTime(lastSyncTime)
}
