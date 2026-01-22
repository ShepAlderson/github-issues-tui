package tui

import (
	"time"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/storage"
)

// Column represents a column in the issue list view
type Column struct {
	Name  string // Internal name (number, title, author, etc.)
	Width int    // Fixed width, or 0 for flexible width
	Title string // Display title
}

// Valid column names
var validColumns = map[string]bool{
	"number":    true,
	"title":     true,
	"author":    true,
	"state":     true,
	"date":      true,
	"updated":   true,
	"comments":  true,
	"labels":    true,
	"assignees": true,
}

// Default column configuration
var defaultColumnConfig = []string{"number", "title", "author", "date", "comments"}

// Column metadata for defaults
var columnMetadata = map[string]Column{
	"number":    {Name: "number", Width: 7, Title: "#"},
	"title":     {Name: "title", Width: 0, Title: "Title"},
	"author":    {Name: "author", Width: 15, Title: "Author"},
	"state":     {Name: "state", Width: 8, Title: "State"},
	"date":      {Name: "date", Width: 0, Title: "Date"},
	"updated":   {Name: "updated", Width: 0, Title: "Updated"},
	"comments":  {Name: "comments", Width: 8, Title: "Comments"},
	"labels":    {Name: "labels", Width: 0, Title: "Labels"},
	"assignees": {Name: "assignees", Width: 0, Title: "Assignees"},
}

// ParseColumnConfig parses column configuration into Column structs
func ParseColumnConfig(columnNames []string) []Column {
	if len(columnNames) == 0 {
		return []Column{}
	}

	columns := make([]Column, 0, len(columnNames))
	for _, name := range columnNames {
		if meta, ok := columnMetadata[name]; ok {
			columns = append(columns, meta)
		}
	}

	return columns
}

// GetColumnValue retrieves the value for a specific column from an issue
func GetColumnValue(issue storage.Issue, columnName string) string {
	switch columnName {
	case "number":
		return formatNumber(issue.Number)
	case "title":
		return issue.Title
	case "author":
		return issue.Author
	case "state":
		return issue.State
	case "date":
		return formatDate(issue.CreatedAt)
	case "updated":
		return formatDate(issue.UpdatedAt)
	case "comments":
		return formatNumber(issue.Comments)
	case "labels":
		return issue.Labels
	case "assignees":
		return issue.Assignees
	default:
		return ""
	}
}

// GetDefaultColumns returns column configuration from config or defaults
func GetDefaultColumns(cfg *config.Config) []Column {
	if cfg == nil || len(cfg.Display.Columns) == 0 {
		return ParseColumnConfig(defaultColumnConfig)
	}

	return ParseColumnConfig(cfg.Display.Columns)
}

// ValidateColumnConfig checks if all column names are valid
func ValidateColumnConfig(columnNames []string) bool {
	if len(columnNames) == 0 {
		return true
	}

	for _, name := range columnNames {
		if !validColumns[name] {
			return false
		}
	}

	return true
}

// formatNumber formats a number for display
func formatNumber(n int) string {
	// Use fmt.Sprintf but avoid import by using simple conversion
	if n == 0 {
		return "0"
	}

	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	return sign + string(digits)
}

// formatDate formats a date for display (e.g., "Jan 15")
func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2")
}
