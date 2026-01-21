package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Column represents a display column configuration
type Column struct {
	Key   string
	Label string
	Width int
}

// ColumnRenderer handles rendering of issue data into columns
type ColumnRenderer struct {
	columns []Column
	styles  map[string]lipgloss.Style
}

// NewColumnRenderer creates a new column renderer from config
func NewColumnRenderer(columnKeys []string) *ColumnRenderer {
	// Define available columns
	availableColumns := map[string]Column{
		"number": {
			Key:   "number",
			Label: "#",
			Width: 6,
		},
		"title": {
			Key:   "title",
			Label: "Title",
			Width: 30,
		},
		"author": {
			Key:   "author",
			Label: "Author",
			Width: 15,
		},
		"date": {
			Key:   "date",
			Label: "Date",
			Width: 12,
		},
		"comments": {
			Key:   "comments",
			Label: "Comments",
			Width: 10,
		},
	}

	// Build columns list from config keys
	var columns []Column
	for _, key := range columnKeys {
		if col, ok := availableColumns[key]; ok {
			columns = append(columns, col)
		}
	}

	// If no valid columns specified, use defaults
	if len(columns) == 0 {
		columns = []Column{
			availableColumns["number"],
			availableColumns["title"],
			availableColumns["author"],
			availableColumns["date"],
			availableColumns["comments"],
		}
	}

	// Create styles
	styles := map[string]lipgloss.Style{
		"header": lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("240")),
		"cell": lipgloss.NewStyle(),
		"selected": lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")),
	}

	return &ColumnRenderer{
		columns: columns,
		styles:  styles,
	}
}

// RenderHeader renders the column headers
func (cr *ColumnRenderer) RenderHeader() string {
	var headers []string
	for _, col := range cr.columns {
		header := cr.styles["header"].Width(col.Width).Render(col.Label)
		headers = append(headers, header)
	}
	return strings.Join(headers, " ")
}

// RenderIssue renders an issue row
func (cr *ColumnRenderer) RenderIssue(issue IssueItem, selected bool) string {
	var cells []string
	style := cr.styles["cell"]
	if selected {
		style = cr.styles["selected"]
	}

	for _, col := range cr.columns {
		value := cr.getColumnValue(issue, col.Key)
		cell := style.Width(col.Width).Render(value)
		cells = append(cells, cell)
	}

	return strings.Join(cells, " ")
}

// getColumnValue extracts the value for a specific column from an issue
func (cr *ColumnRenderer) getColumnValue(issue IssueItem, columnKey string) string {
	switch columnKey {
	case "number":
		return fmt.Sprintf("#%d", issue.Number)
	case "title":
		return issue.TitleText
	case "author":
		return issue.Author
	case "date":
		return issue.Date
	case "comments":
		if issue.CommentCount == 0 {
			return ""
		}
		return fmt.Sprintf("💬%d", issue.CommentCount)
	default:
		return ""
	}
}

// FormatDate formats a date for display
func FormatDate(t time.Time) string {
	// Format as YYYY-MM-DD
	return t.Format("2006-01-02")
}

// TotalWidth returns the total width of all columns
func (cr *ColumnRenderer) TotalWidth() int {
	total := 0
	for _, col := range cr.columns {
		total += col.Width
	}
	// Add spacing between columns
	total += len(cr.columns) - 1
	return total
}