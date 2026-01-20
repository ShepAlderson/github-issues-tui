package main

import (
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/db"
)

func TestParseColumns(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", config.DefaultColumns()},
		{"number,title", []string{"number", "title"}},
		{"number, title, author", []string{"number", "title", "author"}},
		{" number , title , author ", []string{"number", "title", "author"}},
		{"number,title,author,date,comments", []string{"number", "title", "author", "date", "comments"}},
	}

	for _, tt := range tests {
		result := ParseColumns(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("ParseColumns(%q) returned %d columns, expected %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i, col := range result {
			if col != tt.expected[i] {
				t.Errorf("ParseColumns(%q)[%d] = %q, expected %q", tt.input, i, col, tt.expected[i])
			}
		}
	}
}

func TestColumnsToString(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"number", "title"}, "number,title"},
		{[]string{"number", "title", "author", "date", "comments"}, "number,title,author,date,comments"},
		{[]string{}, ""},
		{[]string{"single"}, "single"},
	}

	for _, tt := range tests {
		result := ColumnsToString(tt.input)
		if result != tt.expected {
			t.Errorf("ColumnsToString(%v) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestValidateColumns(t *testing.T) {
	validCols := []string{"number", "title", "author", "date", "comments"}

	if !ValidateColumns(validCols) {
		t.Error("Valid columns should return true")
	}

	if !ValidateColumns([]string{"number", "title"}) {
		t.Error("Subset of valid columns should return true")
	}

	if ValidateColumns([]string{"number", "invalid"}) {
		t.Error("Columns with invalid column should return false")
	}

	if ValidateColumns([]string{"number", "title", "bad"}) {
		t.Error("Columns with bad column should return false")
	}

	if ValidateColumns([]string{}) {
		t.Error("Empty columns should return false")
	}
	if !ValidateColumns([]string{"number", "title", "author", "date", "comments"}) {
		t.Error("All valid columns should return true")
	}
}

func TestGetColumnIndex(t *testing.T) {
	columns := []string{"number", "title", "author", "date", "comments"}

	tests := []struct {
		name     string
		column   string
		expected int
	}{
		{"number", "number", 0},
		{"title", "title", 1},
		{"author", "author", 2},
		{"date", "date", 3},
		{"comments", "comments", 4},
		{"not found", "notfound", -1},
	}

	for _, tt := range tests {
		result := GetColumnIndex(columns, tt.column)
		if result != tt.expected {
			t.Errorf("GetColumnIndex(%q) = %d, expected %d", tt.column, result, tt.expected)
		}
	}
}

func TestColumnWidth(t *testing.T) {
	issue := db.IssueList{
		Number:     123,
		Title:      "Test Title",
		Author:     "test-author",
		CreatedAt:  "2024-01-15T10:30:00Z",
		CommentCnt: 5,
	}

	tests := []struct {
		columnName string
		expected   int
	}{
		{"number", 4},   // "#123"
		{"title", 10},   // "Test Title"
		{"author", 11},  // "test-author"
		{"date", 10},    // "2024-01-15"
		{"comments", 1}, // "5"
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := ColumnWidth(nil, tt.columnName, issue)
		if result != tt.expected {
			t.Errorf("ColumnWidth(%q) = %d, expected %d", tt.columnName, result, tt.expected)
		}
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2024-01-15T10:30:00Z", "2024-01-15"},
		{"2023-12-25T00:00:00Z", "2023-12-25"},
		{"short", "short"},
		{"", ""},
	}

	for _, tt := range tests {
		result := formatDate(tt.input)
		if result != tt.expected {
			t.Errorf("formatDate(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatIssueForDisplay(t *testing.T) {
	issue := db.IssueList{
		Number:     42,
		Title:      "Test Issue",
		Author:     "testuser",
		CreatedAt:  "2024-01-15T10:30:00Z",
		CommentCnt: 5,
	}

	// Test with all columns
	columns := []string{"number", "title", "author", "date", "comments"}
	result := formatIssueForDisplay(issue, columns)
	expected := "#42 | Test Issue | testuser | 2024-01-15 | 5 comments"
	if result != expected {
		t.Errorf("formatIssueForDisplay() = %q, expected %q", result, expected)
	}

	// Test with partial columns
	columns = []string{"number", "title"}
	result = formatIssueForDisplay(issue, columns)
	expected = "#42 | Test Issue"
	if result != expected {
		t.Errorf("formatIssueForDisplay() = %q, expected %q", result, expected)
	}

	// Test with zero comments
	issue.CommentCnt = 0
	columns = []string{"number", "title", "comments"}
	result = formatIssueForDisplay(issue, columns)
	expected = "#42 | Test Issue"
	if result != expected {
		t.Errorf("formatIssueForDisplay() with 0 comments = %q, expected %q", result, expected)
	}
}

func TestFormatIssueSecondary(t *testing.T) {
	issue := db.IssueList{
		Number:     42,
		Title:      "Test Issue",
		Author:     "testuser",
		CreatedAt:  "2024-01-15T10:30:00Z",
		CommentCnt: 5,
	}

	// Test with all columns (excluding number and title)
	columns := []string{"number", "title", "author", "date", "comments"}
	result := formatIssueSecondary(issue, columns)
	expected := "by testuser 2024-01-15 5 comments"
	if result != expected {
		t.Errorf("formatIssueSecondary() = %q, expected %q", result, expected)
	}

	// Test with only number and title
	columns = []string{"number", "title"}
	result = formatIssueSecondary(issue, columns)
	if result != "" {
		t.Errorf("formatIssueSecondary() with only number/title = %q, expected empty string", result)
	}
}

func TestFormatIssueDetail(t *testing.T) {
	issue := db.IssueList{
		Number:     42,
		Title:      "Test Issue Title",
		Author:     "testuser",
		CreatedAt:  "2024-01-15T10:30:00Z",
		CommentCnt: 5,
		State:      "open",
	}

	result := formatIssueDetail(issue, "owner", "repo")

	if !contains(result, "#42") {
		t.Error("formatIssueDetail() should contain issue number")
	}
	if !contains(result, "Test Issue Title") {
		t.Error("formatIssueDetail() should contain title")
	}
	if !contains(result, "testuser") {
		t.Error("formatIssueDetail() should contain author")
	}
	if !contains(result, "open") {
		t.Error("formatIssueDetail() should contain state")
	}
	if !contains(result, "github.com/owner/repo/issues/42") {
		t.Error("formatIssueDetail() should contain URL")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}