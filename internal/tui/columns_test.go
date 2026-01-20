package tui

import (
	"reflect"
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/storage"
)

func TestParseColumnConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []Column
	}{
		{
			name:  "default columns",
			input: []string{"number", "title", "author", "date", "comments"},
			expected: []Column{
				{Name: "number", Width: 7, Title: "#"},
				{Name: "title", Width: 0, Title: "Title"},
				{Name: "author", Width: 15, Title: "Author"},
				{Name: "date", Width: 0, Title: "Date"},
				{Name: "comments", Width: 8, Title: "Comments"},
			},
		},
		{
			name:  "custom column set",
			input: []string{"number", "title", "labels"},
			expected: []Column{
				{Name: "number", Width: 7, Title: "#"},
				{Name: "title", Width: 0, Title: "Title"},
				{Name: "labels", Width: 0, Title: "Labels"},
			},
		},
		{
			name:  "single column",
			input: []string{"title"},
			expected: []Column{
				{Name: "title", Width: 0, Title: "Title"},
			},
		},
		{
			name:     "empty config returns defaults",
			input:    []string{},
			expected: []Column{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseColumnConfig(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseColumnConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetColumnValue(t *testing.T) {
	createdAt := testDate("2024-01-15T10:30:00Z")
	updatedAt := testDate("2024-01-16T14:20:00Z")

	issue := storage.Issue{
		Number:    123,
		Title:     "Test issue title",
		Author:    "testuser",
		State:     "open",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Comments:  5,
		Labels:    "bug,enhancement",
		Assignees: "user1,user2",
	}

	tests := []struct {
		name       string
		columnName string
		expected   string
	}{
		{
			name:       "number column",
			columnName: "number",
			expected:   "123",
		},
		{
			name:       "title column",
			columnName: "title",
			expected:   "Test issue title",
		},
		{
			name:       "author column",
			columnName: "author",
			expected:   "testuser",
		},
		{
			name:       "state column",
			columnName: "state",
			expected:   "open",
		},
		{
			name:       "date column",
			columnName: "date",
			expected:   "Jan 15",
		},
		{
			name:       "comments column",
			columnName: "comments",
			expected:   "5",
		},
		{
			name:       "labels column",
			columnName: "labels",
			expected:   "bug,enhancement",
		},
		{
			name:       "assignees column",
			columnName: "assignees",
			expected:   "user1,user2",
		},
		{
			name:       "unknown column returns empty",
			columnName: "unknown",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetColumnValue(issue, tt.columnName)
			if result != tt.expected {
				t.Errorf("GetColumnValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetDefaultColumns(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{
			Columns: []string{},
		},
	}

	columns := GetDefaultColumns(cfg)
	if len(columns) == 0 {
		t.Error("GetDefaultColumns() should return default columns when config is empty")
	}

	expectedColumns := []string{"number", "title", "author", "date", "comments"}
	for i, col := range columns {
		if col.Name != expectedColumns[i] {
			t.Errorf("Column %d name = %q, want %q", i, col.Name, expectedColumns[i])
		}
	}
}

func TestGetDefaultColumns_WithConfig(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{
			Columns: []string{"number", "title", "labels"},
		},
	}

	columns := GetDefaultColumns(cfg)
	if len(columns) != 3 {
		t.Errorf("GetDefaultColumns() returned %d columns, want 3", len(columns))
	}

	expectedColumns := []string{"number", "title", "labels"}
	for i, col := range columns {
		if col.Name != expectedColumns[i] {
			t.Errorf("Column %d name = %q, want %q", i, col.Name, expectedColumns[i])
		}
	}
}

func TestValidateColumnConfig(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		expected bool
	}{
		{
			name:     "valid columns",
			columns:  []string{"number", "title", "author", "date", "comments"},
			expected: true,
		},
		{
			name:     "valid with labels and assignees",
			columns:  []string{"number", "title", "labels", "assignees"},
			expected: true,
		},
		{
			name:     "invalid column name",
			columns:  []string{"number", "invalid", "title"},
			expected: false,
		},
		{
			name:     "empty column list",
			columns:  []string{},
			expected: true,
		},
		{
			name:     "partial valid columns",
			columns:  []string{"number", "unknown", "title"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateColumnConfig(tt.columns)
			if result != tt.expected {
				t.Errorf("ValidateColumnConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to create test dates
func testDate(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
