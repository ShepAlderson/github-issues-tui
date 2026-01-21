package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

// IssueList represents the issue list component
type IssueList struct {
	dbManager      *database.DBManager
	config         *config.Config
	columnRenderer *ColumnRenderer
	issues         []IssueItem
	selectedIdx    int
	width          int
	height         int
}

// IssueItem represents an item in the issue list
type IssueItem struct {
	Number       int
	TitleText    string
	Author       string
	Date         string
	CommentCount int
}

// NewIssueList creates a new issue list component
func NewIssueList(dbManager *database.DBManager, cfg *config.Config) *IssueList {
	// Create column renderer from config
	var columnKeys []string
	if cfg.Display.Columns != nil {
		columnKeys = cfg.Display.Columns
	} else {
		// Use defaults if not specified
		columnKeys = []string{"number", "title", "author", "date", "comments"}
	}

	columnRenderer := NewColumnRenderer(columnKeys)

	return &IssueList{
		dbManager:      dbManager,
		config:         cfg,
		columnRenderer: columnRenderer,
		selectedIdx:    0,
	}
}

// Init loads issues from the database
func (il *IssueList) Init() tea.Cmd {
	// TODO: Load issues from database
	// For now, return dummy data for testing
	return il.loadDummyIssues()
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

	// Add status bar with issue count
	status := fmt.Sprintf("Issues: %d/%d", il.selectedIdx+1, len(il.issues))
	builder.WriteString("\n")
	builder.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(status))

	return builder.String()
}

// loadDummyIssues loads dummy issues for testing
func (il *IssueList) loadDummyIssues() tea.Cmd {
	// Create dummy issues for testing
	dummyIssues := []IssueItem{
		{Number: 1, TitleText: "Fix bug in login flow", Author: "alice", Date: "2024-01-15", CommentCount: 3},
		{Number: 2, TitleText: "Add dark mode support", Author: "bob", Date: "2024-01-14", CommentCount: 0},
		{Number: 3, TitleText: "Improve performance", Author: "charlie", Date: "2024-01-13", CommentCount: 7},
		{Number: 4, TitleText: "Update documentation", Author: "david", Date: "2024-01-12", CommentCount: 2},
		{Number: 5, TitleText: "Refactor API endpoints", Author: "eve", Date: "2024-01-11", CommentCount: 1},
	}

	// Update issue list
	il.issues = dummyIssues

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