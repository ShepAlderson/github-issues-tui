package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/errors"
	"github.com/shepbook/ghissues/internal/themes"
)

// runTUIWithRefresh starts the TUI with automatic refresh on launch
func RunTUIWithRefresh(dbPath string, cfg *config.Config) error {
	// Perform initial sync in background
	go func() {
		_ = RefreshSync(dbPath, cfg, nil)
	}()

	return RunTUI(dbPath, cfg)
}

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
	currentSortOrder := cfg.Display.SortOrder

	// Track current selection index
	currentIndex := 0

	// Markdown mode state
	markdownRendered := true

	// Refresh state - must be declared before updateStatusBar
	isRefreshing := false
	refreshStatus := ""

	// Last synced timestamp - must be declared before updateStatusBar
	lastSynced := ""

	// Get the current theme
	themeName := cfg.Display.Theme
	if themeName == "" {
		themeName = config.DefaultTheme()
	}
	currentTheme := themes.Get(themeName)

	// Create the tview application
	app := tview.NewApplication()

	// Create the issue list view - explicit type to avoid type inference issues
	var issueList *tview.List
	issueList = tview.NewList()
	issueList.SetMainTextColor(getThemeColor(currentTheme.PrimaryTextColor))
	issueList.SetSecondaryTextColor(getThemeColor(currentTheme.SecondaryTextColor))
	issueList.SetSelectedTextColor(getThemeColor(currentTheme.PrimaryTextColor))
	issueList.SetSelectedBackgroundColor(getThemeColor(currentTheme.ContrastBackgroundColor))
	issueList.SetTitle(" Issues ")
	issueList.SetBorder(true)
	issueList.SetBorderColor(getThemeColor(currentTheme.BorderColor))

	// Create scrollable detail view
	var detailView *tview.TextView
	detailView = tview.NewTextView()
	detailView.SetText("Select an issue to view details.\n\nPress 'r' to refresh issues from GitHub.")
	detailView.SetTextAlign(tview.AlignLeft)
	detailView.SetTitle(" Details ")
	detailView.SetBorder(true)
	detailView.SetBorderColor(getThemeColor(currentTheme.BorderColor))
	detailView.SetScrollable(true)

	// Create status bar
	statusBar := tview.NewTextView()
	statusBar.SetTextAlign(tview.AlignLeft)

	// Create footer for context-sensitive keybindings
	footer := tview.NewTextView()
	footer.SetTextAlign(tview.AlignCenter)
	footer.SetDynamicColors(true)

	// Comments view state - must be declared before updateStatusBar
	commentsMarkdownRendered := true
	inCommentsView := false

	// Error state - must be declared before updateStatusBar
	var minorError string

	// Function to update status bar text
	updateStatusBar := func() {
		if inCommentsView {
			// Comments view status bar
			markdownText := ""
			if commentsMarkdownRendered {
				markdownText = " [Markdown]"
			} else {
				markdownText = " [Raw]"
			}
			statusBar.SetText(fmt.Sprintf(" ghissues | %s/%s | Comments View%s | q or Esc to return | m for markdown toggle ",
				owner, repo, markdownText))
		} else if minorError != "" {
			// Show minor error in status bar
			statusBar.SetText(fmt.Sprintf(" ghissues | %s/%s | [ERROR] %s | j/k or arrows to navigate | r to refresh | q to quit ",
				owner, repo, minorError))
		} else if isRefreshing {
			// Show refresh status during refresh
			statusBar.SetText(fmt.Sprintf(" ghissues | %s/%s | Refreshing... %s | j/k or arrows to navigate | r to refresh | q to quit ",
				owner, repo, refreshStatus))
		} else {
			// Issue list status bar
			sortText := FormatSortDisplay(currentSort, currentSortOrder)
			markdownText := ""
			if markdownRendered {
				markdownText = " [Markdown]"
			} else {
				markdownText = " [Raw]"
			}
			lastSyncedText := GetLastSyncedDisplay(lastSynced)
			statusBar.SetText(fmt.Sprintf(" ghissues | %s/%s | %s | %s | %sj/k or arrows to navigate | q to quit | s to sort | S to reverse | ? for help | m for markdown toggle | r to refresh ",
				owner, repo, lastSyncedText, sortText, markdownText))
		}
	}

	// Function to update footer with context-sensitive keybindings
	updateFooter := func() {
		footer.SetText(GetFooterDisplay(inCommentsView, commentsMarkdownRendered))
	}

	// Function to format issue detail with full information
	formatIssueDetailFull := func(issue *db.IssueDetail) string {
		if issue == nil {
			return "Select an issue to view details.\n\nPress 'r' to refresh issues from GitHub."
		}

		var sb strings.Builder

		// Header
		sb.WriteString(fmt.Sprintf(" #%d %s\n\n", issue.Number, issue.Title))

		// Status badge
		stateIcon := "○"
		if issue.State == "closed" {
			stateIcon = "●"
		}
		sb.WriteString(fmt.Sprintf("%s **%s**  |  ", stateIcon, strings.ToUpper(issue.State)))

		// Author
		sb.WriteString(fmt.Sprintf("by **%s**  |  ", issue.Author))

		// Dates
		sb.WriteString(fmt.Sprintf("Created: %s  |  Updated: %s\n\n", formatDate(issue.CreatedAt), formatDate(issue.UpdatedAt)))

		// Labels
		if len(issue.Labels) > 0 {
			sb.WriteString("Labels: ")
			for i, label := range issue.Labels {
				if i > 0 {
					sb.WriteString("  ")
				}
				sb.WriteString(fmt.Sprintf("[%s]", label))
			}
			sb.WriteString("\n")
		}

		// Assignees
		if len(issue.Assignees) > 0 {
			sb.WriteString("Assignees: ")
			for i, assignee := range issue.Assignees {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(assignee)
			}
			sb.WriteString("\n")
		}

		// Comments count
		if issue.CommentCnt > 0 {
			sb.WriteString(fmt.Sprintf("%d comment(s)\n\n", issue.CommentCnt))
		} else {
			sb.WriteString("No comments\n\n")
		}

		// Body
		if markdownRendered {
			sb.WriteString(renderMarkdown(issue.Body))
		} else {
			sb.WriteString("--- Body ---\n")
			sb.WriteString(issue.Body)
		}

		// URL at the bottom
		sb.WriteString(fmt.Sprintf("\n\n%s", issue.HTMLURL))

		return sb.String()
	}

	// Function to fetch and display issues
	issues := []db.IssueList{}
	displayIssues := func() {
		// Fetch issues from database with current sort settings
		issues, err = db.ListIssuesSorted(database, owner, repo, currentSort, currentSortOrder)
		if err != nil {
			// Just log the error, don't fail
			_ = fmt.Errorf("failed to list issues: %w", err)
			issues = []db.IssueList{}
		}

		// Fetch last sync time for status bar
		lastSynced, _ = db.GetLastSyncTime(database, owner, repo)

		// Clear and repopulate the list
		issueList.Clear()
		columns := cfg.Display.Columns
		for _, issue := range issues {
			text := formatIssueForDisplay(issue, columns)
			secondary := formatIssueSecondary(issue, columns)
			issueList.AddItem(text, secondary, rune('0'+issue.Number%10), nil)
		}

		// Update status bar
		updateStatusBar()
		// Update footer
		updateFooter()

		// Update detail view for first selection
		if len(issues) > 0 {
			issueNum := issues[0].Number
			detail, err := db.GetIssueDetail(database, owner, repo, issueNum)
			if err == nil && detail != nil {
				detailView.SetText(formatIssueDetailFull(detail))
			}
		} else {
			detailView.SetText("No issues found.\n\nPress 'r' to sync issues from GitHub.")
		}
	}

	// Create pages for modals
	pages := tview.NewPages()

	// Comments view TextView - state variables are declared earlier
	var commentsView *tview.TextView

	// Return to issue list from comments view
	returnToIssueList := func() {
		pages.SwitchToPage("main")
		inCommentsView = false
		commentsMarkdownRendered = true // Reset markdown state for next time
		updateStatusBar()
		updateFooter()
		app.SetFocus(issueList)
	}

	// Show comments view (drill-down)
	showComments := func(issueNum int) {
		// Get issue details for header
		detail, err := db.GetIssueDetail(database, owner, repo, issueNum)
		if err != nil {
			detail = nil
		}

		comments, err := db.GetComments(database, issueNum)
		if err != nil {
			commentsText := fmt.Sprintf("Error loading comments: %v", err)
			if commentsView == nil {
				commentsView = tview.NewTextView()
			}
			commentsView.SetText(commentsText)
			commentsView.SetTextAlign(tview.AlignLeft)
			commentsView.SetScrollable(true)
			pages.SwitchToPage("comments")
			return
		}

		issueTitle := ""
		if detail != nil {
			issueTitle = detail.Title
		}

		// Build comments text with issue header
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf(" #%d %s\n\n", issueNum, issueTitle))
		sb.WriteString(formatComments(comments, commentsMarkdownRendered))
		commentsText := sb.String()

		if commentsView == nil {
			commentsView = tview.NewTextView()
		}
		commentsView.SetText(commentsText)
		commentsView.SetTextAlign(tview.AlignLeft)
		commentsView.SetScrollable(true)
		commentsView.SetBorder(true)
		commentsView.SetTitle(" Comments ")

		// Set up input capture for comments view
		commentsView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'q', 'Q':
					returnToIssueList()
					return nil
				case 'm', 'M':
					// Toggle markdown rendering in comments
					commentsMarkdownRendered = !commentsMarkdownRendered
					updateStatusBar()
					updateFooter()
					// Refresh comments view
					var refreshSb strings.Builder
					refreshSb.WriteString(fmt.Sprintf(" #%d %s\n\n", issueNum, issueTitle))
					refreshSb.WriteString(formatComments(comments, commentsMarkdownRendered))
					commentsView.SetText(refreshSb.String())
					return nil
				}
			case tcell.KeyEscape:
				returnToIssueList()
				return nil
			case tcell.KeyCtrlC:
				app.Stop()
				return nil
			}
			return event
		})

		// Add comments page (replaces main view)
		pages.AddPage("comments", commentsView, true, true)
		inCommentsView = true
		updateStatusBar()
		updateFooter()
		pages.SwitchToPage("comments")
		app.SetFocus(commentsView)
	}

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
				currentSortOrder = ToggleSortOrder(currentSortOrder)
				cfg.Display.SortOrder = currentSortOrder
				if err := config.Save(cfg); err != nil {
					_ = fmt.Errorf("failed to save config: %w", err)
				}
				displayIssues()
				return nil
			case 'm', 'M':
				// Toggle markdown rendering
				markdownRendered = !markdownRendered
				updateStatusBar()
				// Refresh detail view with current issue
				if currentIndex >= 0 && currentIndex < len(issues) {
					issueNum := issues[currentIndex].Number
					detail, err := db.GetIssueDetail(database, owner, repo, issueNum)
					if err == nil && detail != nil {
						detailView.SetText(formatIssueDetailFull(detail))
					}
				}
				return nil
			case 'r', 'R':
				// Manual refresh - only if not already refreshing and not in comments view
				if !isRefreshing && !inCommentsView {
					isRefreshing = true
					minorError = "" // Clear previous minor error
					updateStatusBar()

					// Run refresh in a goroutine
					go func() {
						progress := func(current, total int, status string) {
							refreshStatus = status
							app.QueueUpdateDraw(func() {
								updateStatusBar()
							})
						}

						err := RefreshSync(dbPath, cfg, progress)

						app.QueueUpdateDraw(func() {
							isRefreshing = false
							refreshStatus = ""
							if err != nil {
								uiErr := errors.NewUIError(err)
								if uiErr.Category == errors.CategoryCritical {
									// Show critical error as modal
									showCriticalErrorModal(app, pages, uiErr)
								} else {
									// Show minor error in status bar
									minorError = uiErr.Hint.Message
								}
							}
							updateStatusBar()
							// Refresh the issue list
							displayIssues()
						})
					}()
				} else if minorError != "" {
					// If there's a minor error, pressing 'r' clears it and retries
					minorError = ""
					// Trigger refresh
					isRefreshing = true
					updateStatusBar()

					go func() {
						progress := func(current, total int, status string) {
							refreshStatus = status
							app.QueueUpdateDraw(func() {
								updateStatusBar()
							})
						}

						err := RefreshSync(dbPath, cfg, progress)

						app.QueueUpdateDraw(func() {
							isRefreshing = false
							refreshStatus = ""
							if err != nil {
								uiErr := errors.NewUIError(err)
								if uiErr.Category == errors.CategoryCritical {
									showCriticalErrorModal(app, pages, uiErr)
								} else {
									minorError = uiErr.Hint.Message
								}
							}
							updateStatusBar()
							displayIssues()
						})
					}()
				}
				return nil
			}
		case tcell.KeyEscape:
			// Dismiss modal if visible, or return from comments view
			if pages.HasPage("help") {
				pages.SwitchToPage("main")
				app.SetFocus(app.GetFocus())
			} else if pages.HasPage("error") {
				pages.SwitchToPage("main")
				app.SetFocus(app.GetFocus())
			} else if pages.HasPage("comments") {
				returnToIssueList()
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
		currentIndex = index
		if index >= 0 && index < len(issues) {
			issueNum := issues[index].Number
			detail, err := db.GetIssueDetail(database, owner, repo, issueNum)
			if err == nil && detail != nil {
				detailView.SetText(formatIssueDetailFull(detail))
			}
		}
	})

	// Update detail view and show comments on selection confirm
	issueList.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		currentIndex = index
		if index >= 0 && index < len(issues) {
			issueNum := issues[index].Number
			detail, err := db.GetIssueDetail(database, owner, repo, issueNum)
			if err == nil && detail != nil {
				detailView.SetText(formatIssueDetailFull(detail))
			}
			// Show comments view
			showComments(issueNum)
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
	mainFlex.AddItem(footer, 1, 0, false)

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

// formatIssueDetail formats an issue for the detail view (for backward compatibility with tests)
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

// FormatRelativeTime formats a timestamp as a relative time string
// e.g., "5 minutes ago", "2 hours ago", "3 days ago"
// Returns "never" for empty or invalid timestamps
func FormatRelativeTime(timestamp string) string {
	if timestamp == "" {
		return "never"
	}

	// Try to parse the timestamp
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "never"
	}

	now := time.Now()
	diff := now.Sub(t)

	// Handle future timestamps (show as "just now")
	if diff < 0 {
		return "just now"
	}

	// Just now (less than 1 minute)
	if diff < time.Minute {
		return "just now"
	}

	// Minutes
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}

	// Hours
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// Days
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

// GetLastSyncedDisplay returns a formatted string for the status bar
// showing when the data was last synced
func GetLastSyncedDisplay(timestamp string) string {
	return fmt.Sprintf("Last synced: %s", FormatRelativeTime(timestamp))
}

// GetFooterDisplay returns the footer text based on the current view context
func GetFooterDisplay(inCommentsView, commentsMarkdownRendered bool) string {
	if inCommentsView {
		// Comments view footer
		markdownText := ""
		if commentsMarkdownRendered {
			markdownText = "[yellow]m[-] markdown"
		} else {
			markdownText = "[yellow]m[-] raw"
		}
		return fmt.Sprintf(" [yellow]q/Esc[-] back  %s ", markdownText)
	}
	// Issue list view footer
	return " [yellow]?[-] help  [yellow]Enter[-] comments  [yellow]r[-] refresh "
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
   Enter          - View comments for selected issue
   r              - Refresh issues from GitHub

 View:
   m              - Toggle markdown rendered/raw

 Comments View:
   m              - Toggle markdown rendered/raw
   q / Esc        - Return to issue list

 Sorting:
   s              - Cycle sort (updated → created → number → comments)
   S              - Reverse sort order

 Other:
   ?              - Show this help / Dismiss help
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

	// Add input capture to dismiss help with ? or Esc
	helpView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			pages.SwitchToPage("main")
			app.SetFocus(app.GetFocus())
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case '?':
				pages.SwitchToPage("main")
				app.SetFocus(app.GetFocus())
				return nil
			}
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
		}
		return event
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

// renderMarkdown renders markdown text for terminal display
func renderMarkdown(body string) string {
	if body == "" {
		return "_No description provided._"
	}
	// Use glamour to render markdown with terminal styling
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	if err != nil {
		// Fall back to raw text if rendering fails
		return body
	}
	rendered, err := renderer.Render(body)
	if err != nil {
		return body
	}
	return rendered
}

// formatComments returns a formatted string for displaying comments
func formatComments(comments []db.Comment, markdownRendered bool) string {
	if len(comments) == 0 {
		return "No comments yet."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d Comment(s)\n\n", len(comments)))

	for i, comment := range comments {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("--- Comment #%d ---\n", i+1))
		sb.WriteString(fmt.Sprintf("By: %s  |  Date: %s\n\n", comment.Author, formatDate(comment.CreatedAt)))
		if markdownRendered {
			sb.WriteString(renderMarkdown(comment.Body))
		} else {
			sb.WriteString(comment.Body)
		}
	}

	return sb.String()
}

// showCriticalErrorModal displays a critical error in a modal dialog
// The modal requires user acknowledgment before continuing
func showCriticalErrorModal(app *tview.Application, pages *tview.Pages, err *errors.UIError) {
	// Build error message with hint
	var errorText string
	if err.Hint != nil {
		errorText = fmt.Sprintf("[red]ERROR[white]\n\n%s\n\n%s", err.Hint.Message, err.Hint.Action)
	} else {
		errorText = fmt.Sprintf("[red]ERROR[white]\n\n%s", err.Err.Error())
	}

	modal := tview.NewModal()
	modal.SetText(errorText)
	modal.AddButtons([]string{"OK"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		pages.SwitchToPage("main")
		app.SetFocus(app.GetFocus())
	})

	pages.AddPage("error", modal, true, true)
	pages.SwitchToPage("error")
}

// getThemeColor converts a hex color string to tcell.Color
// Returns tcell.ColorDefault if the color cannot be parsed
func getThemeColor(hexColor string) tcell.Color {
	if hexColor == "" {
		return tcell.ColorDefault
	}

	// Remove # prefix if present
	colorStr := strings.TrimPrefix(hexColor, "#")
	if len(colorStr) != 6 {
		return tcell.ColorDefault
	}

	// Parse RGB values
	r, err1 := strconv.ParseUint(colorStr[0:2], 16, 8)
	g, err2 := strconv.ParseUint(colorStr[2:4], 16, 8)
	b, err3 := strconv.ParseUint(colorStr[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return tcell.ColorDefault
	}

	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}