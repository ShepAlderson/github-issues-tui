package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
)

// RunTUI starts the TUI application
func RunTUI(dbPath string, cfg *config.Config) error {
	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Parse owner/repo from config
	parts := strings.Split(cfg.Repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format: %s (expected owner/repo)", cfg.Repository)
	}
	owner, repo := parts[0], parts[1]

	// Current sort settings (from config)
	currentSort := cfg.Display.Sort
	currentOrder := cfg.Display.SortOrder

	// Create the tview application
	app := tview.NewApplication()

	// Create the issue list view - explicit type to avoid type inference issues
	var issueList *tview.List
	issueList = tview.NewList()
	issueList.SetMainTextColor(tview.Styles.PrimaryTextColor)
	issueList.SetSecondaryTextColor(tview.Styles.SecondaryTextColor)
	issueList.SetSelectedTextColor(tview.Styles.ContrastBackgroundColor)
	issueList.SetSelectedBackgroundColor(tview.Styles.PrimaryTextColor)
	issueList.SetTitle(" Issues ")
	issueList.SetBorder(true)

	// Create detail placeholder
	var detailView *tview.TextView
	detailView = tview.NewTextView()
	detailView.SetText("Select an issue to view details.\n\nPress 'r' to refresh issues from GitHub.")
	detailView.SetTextAlign(tview.AlignCenter)
	detailView.SetTitle(" Details ")
	detailView.SetBorder(true)

	// Create status bar
	statusBar := tview.NewTextView()
	statusBar.SetTextAlign(tview.AlignLeft)

	// Function to update status bar text
	updateStatusBar := func() {
		sortText := FormatSortDisplay(currentSort, currentOrder)
		statusBar.SetText(fmt.Sprintf(" ghissues | %s/%s | %s | j/k or arrows to navigate | q to quit | s to sort | S to reverse | ? for help ",
			owner, repo, sortText))
	}

	// Function to fetch and display issues
	issues := []db.IssueList{}
	displayIssues := func() {
		// Fetch issues from database with current sort settings
		issues, err = db.ListIssuesSorted(database, owner, repo, currentSort, currentOrder)
		if err != nil {
			// Just log the error, don't fail
			_ = fmt.Errorf("failed to list issues: %w", err)
			issues = []db.IssueList{}
		}

		// Clear and repopulate the list
		issueList.Clear()
		columns := cfg.Display.Columns
		for _, issue := range issues {
			text := formatIssueForDisplay(issue, columns)
			secondary := formatIssueSecondary(issue, columns)
			issueList.AddItem(text, secondary, rune('0'+issue.Number%10), nil)
		}

		// Update issue count in status
		issueCount := len(issues)
		updateStatusBar()

		// Update detail view for first selection
		if issueCount > 0 {
			issue := issues[0]
			detailText := formatIssueDetail(issue, owner, repo)
			detailView.SetText(detailText)
		} else {
			detailView.SetText("No issues found.\n\nPress 'r' to sync issues from GitHub.")
		}
	}

	// Create pages for modals
	pages := tview.NewPages()

	// Set up navigation handlers using tcell events
	issueList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				app.Stop()
				return nil
			case '?':
				// Show help modal
				showHelp(app, pages)
				return nil
			case 's':
				// Cycle to next sort option
				currentSort = CycleSortOption(currentSort)
				cfg.Display.Sort = currentSort
				if err := config.Save(cfg); err != nil {
					_ = fmt.Errorf("failed to save config: %w", err)
				}
				displayIssues()
				return nil
			case 'S':
				// Toggle sort order (reverse)
				currentOrder = ToggleSortOrder(currentOrder)
				cfg.Display.SortOrder = currentOrder
				if err := config.Save(cfg); err != nil {
					_ = fmt.Errorf("failed to save config: %w", err)
				}
				displayIssues()
				return nil
			}
		case tcell.KeyEscape:
			// Dismiss help if visible
			if pages.HasPage("help") {
				pages.SwitchToPage("main")
				app.SetFocus(issueList)
			}
			return event
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return event
	})

	// Initial load of issues
	displayIssues()

	// Update detail view when selection changes
	issueList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(issues) {
			issue := issues[index]
			detailText := formatIssueDetail(issue, owner, repo)
			detailView.SetText(detailText)
		}
	})

	// Update detail view on selection confirm
	issueList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(issues) {
			issue := issues[index]
			detailText := formatIssueDetail(issue, owner, repo)
			detailView.SetText(detailText)
		}
	})

	// Add header
	header := tview.NewTextView()
	header.SetText(fmt.Sprintf(" ghissues - %s/%s ", owner, repo))
	header.SetTextAlign(tview.AlignCenter)

	// Create main layout with vertical split
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexColumn)
	flex.AddItem(issueList, 0, 1, true)
	flex.AddItem(detailView, 0, 2, false)

	pages.AddPage("main", flex, true, true)

	// Main layout with header
	mainFlex := tview.NewFlex()
	mainFlex.SetDirection(tview.FlexRow)
	mainFlex.AddItem(header, 1, 0, false)
	mainFlex.AddItem(pages, 0, 1, true)
	mainFlex.AddItem(statusBar, 1, 0, false)

	app.SetRoot(mainFlex, true)
	app.SetFocus(issueList)

	return app.Run()
}

// formatIssueForDisplay formats the main text of an issue for the list view
func formatIssueForDisplay(issue db.IssueList, columns []string) string {
	var parts []string
	for _, col := range columns {
		switch col {
		case "number":
			parts = append(parts, fmt.Sprintf("#%d", issue.Number))
		case "title":
			parts = append(parts, issue.Title)
		case "author":
			parts = append(parts, issue.Author)
		case "date":
			parts = append(parts, formatDate(issue.CreatedAt))
		case "comments":
			if issue.CommentCnt > 0 {
				parts = append(parts, fmt.Sprintf("%d comments", issue.CommentCnt))
			}
		}
	}
	return strings.Join(parts, " | ")
}

// formatIssueSecondary formats the secondary text of an issue
func formatIssueSecondary(issue db.IssueList, columns []string) string {
	var parts []string
	for _, col := range columns {
		if col == "number" || col == "title" {
			continue
		}
		switch col {
		case "author":
			parts = append(parts, "by "+issue.Author)
		case "date":
			parts = append(parts, formatDate(issue.CreatedAt))
		case "comments":
			if issue.CommentCnt > 0 {
				parts = append(parts, fmt.Sprintf("%d comments", issue.CommentCnt))
			}
		}
	}
	return strings.Join(parts, " ")
}

// formatIssueDetail formats an issue for the detail view
func formatIssueDetail(issue db.IssueList, owner, repo string) string {
	return fmt.Sprintf(`
 Issue #%d - %s

 State: %s
 Author: %s
 Created: %s
 Comments: %d

 URL: https://github.com/%s/%s/issues/%d
`,
		issue.Number,
		issue.Title,
		issue.State,
		issue.Author,
		formatDate(issue.CreatedAt),
		issue.CommentCnt,
		owner, repo,
		issue.Number,
	)
}

// formatDate formats a date string for display
func formatDate(dateStr string) string {
	// Try to parse and format the date nicely
	// The date comes from database as RFC3339 format
	// For now, just return it as-is or a simple version
	if len(dateStr) >= 10 {
		return dateStr[:10]
	}
	return dateStr
}

// showHelp shows a help modal
func showHelp(app *tview.Application, pages *tview.Pages) {
	helpText := `
 Keyboard Shortcuts

 Navigation:
   j / Down Arrow  - Move down
   k / Up Arrow    - Move up
   g / Home       - Go to first item
   G / End        - Go to last item

 Actions:
   Enter          - Select/refresh details
   r              - Refresh issues from GitHub
   c              - Toggle columns configuration

 Sorting:
   s              - Cycle sort (updated → created → number → comments)
   S              - Reverse sort order

 Other:
   ?              - Show this help
   q / Esc        - Quit help / Quit application
   Ctrl+C         - Force quit

`
	helpView := tview.NewModal()
	helpView.SetText(helpText)
	helpView.AddButtons([]string{"Close"})
	helpView.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.SwitchToPage("main")
		app.SetFocus(app.GetFocus())
	})

	pages.AddPage("help", helpView, true, true)
	pages.SwitchToPage("help")
}

// ParseColumns parses a comma-separated list of column names
func ParseColumns(s string) []string {
	if s == "" {
		return config.DefaultColumns()
	}
	parts := strings.Split(s, ",")
	var columns []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			columns = append(columns, p)
		}
	}
	if len(columns) == 0 {
		return config.DefaultColumns()
	}
	return columns
}

// ColumnsToString converts a column slice to a comma-separated string
func ColumnsToString(columns []string) string {
	return strings.Join(columns, ",")
}

// ValidateColumns validates that all column names are valid
func ValidateColumns(columns []string) bool {
	if len(columns) == 0 {
		return false
	}
	validColumns := map[string]bool{
		"number":   true,
		"title":    true,
		"author":   true,
		"date":     true,
		"comments": true,
	}
	for _, col := range columns {
		if !validColumns[col] {
			return false
		}
	}
	return true
}

// GetColumnIndex returns the index of a column in the list, or -1 if not found
func GetColumnIndex(columns []string, name string) int {
	for i, col := range columns {
		if col == name {
			return i
		}
	}
	return -1
}

// ColumnWidth calculates the display width for a column value
func ColumnWidth(columns []string, columnName string, issue db.IssueList) int {
	switch columnName {
	case "number":
		return len(fmt.Sprintf("#%d", issue.Number))
	case "title":
		return len(issue.Title)
	case "author":
		return len(issue.Author)
	case "date":
		return len(formatDate(issue.CreatedAt))
	case "comments":
		return len(strconv.Itoa(issue.CommentCnt))
	}
	return 0
}

// SortOptionInfo contains display information for a sort option
type SortOptionInfo struct {
	Option config.SortOption
	Name   string
}

// GetSortOptionInfo returns display information for a sort option
func GetSortOptionInfo(option config.SortOption) SortOptionInfo {
	switch option {
	case config.SortUpdated:
		return SortOptionInfo{Option: config.SortUpdated, Name: "Updated"}
	case config.SortCreated:
		return SortOptionInfo{Option: config.SortCreated, Name: "Created"}
	case config.SortNumber:
		return SortOptionInfo{Option: config.SortNumber, Name: "Number"}
	case config.SortComments:
		return SortOptionInfo{Option: config.SortComments, Name: "Comments"}
	default:
		return SortOptionInfo{Option: config.SortUpdated, Name: "Updated"}
	}
}

// CycleSortOption returns the next sort option in the cycle
func CycleSortOption(current config.SortOption) config.SortOption {
	options := config.AllSortOptions()
	for i, opt := range options {
		if opt == current {
			// Return the next option, wrapping around to the first
			nextIndex := (i + 1) % len(options)
			return options[nextIndex]
		}
	}
	// If current is not found, return the first option
	return options[0]
}

// ToggleSortOrder returns the opposite sort order
func ToggleSortOrder(current config.SortOrder) config.SortOrder {
	if current == config.SortOrderDesc {
		return config.SortOrderAsc
	}
	return config.SortOrderDesc
}

// FormatSortDisplay returns a formatted string for the status bar
func FormatSortDisplay(sort config.SortOption, order config.SortOrder) string {
	info := GetSortOptionInfo(sort)
	orderStr := "↓"
	if order == config.SortOrderAsc {
		orderStr = "↑"
	}
	return fmt.Sprintf("Sort: %s %s", info.Name, orderStr)
}