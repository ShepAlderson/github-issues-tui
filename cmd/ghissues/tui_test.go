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

func TestGetSortOptionInfo(t *testing.T) {
	tests := []struct {
		option   config.SortOption
		expected string
	}{
		{config.SortUpdated, "Updated"},
		{config.SortCreated, "Created"},
		{config.SortNumber, "Number"},
		{config.SortComments, "Comments"},
		{"invalid", "Updated"}, // Default for unknown options
	}

	for _, tt := range tests {
		result := GetSortOptionInfo(tt.option)
		if result.Name != tt.expected {
			t.Errorf("GetSortOptionInfo(%q).Name = %q, expected %q", tt.option, result.Name, tt.expected)
		}
		if result.Option != tt.option && tt.option != "invalid" {
			t.Errorf("GetSortOptionInfo(%q).Option = %q, expected %q", tt.option, result.Option, tt.option)
		}
	}
}

func TestCycleSortOption(t *testing.T) {
	// Test cycling through all options
	options := config.AllSortOptions()

	// Start with the first option and cycle through all
	current := options[0]
	for i := 0; i < len(options); i++ {
		next := CycleSortOption(current)
		expectedNext := options[(i+1)%len(options)]
		if next != expectedNext {
			t.Errorf("CycleSortOption(%q) = %q, expected %q", current, next, expectedNext)
		}
		current = next
	}

	// Test wrapping around
	lastOption := options[len(options)-1]
	firstOption := options[0]
	next := CycleSortOption(lastOption)
	if next != firstOption {
		t.Errorf("CycleSortOption(%q) = %q, expected %q (wrap around)", lastOption, next, firstOption)
	}

	// Test with invalid option - should return first option
	invalidResult := CycleSortOption("invalid")
	if invalidResult != firstOption {
		t.Errorf("CycleSortOption(%q) = %q, expected %q (default)", "invalid", invalidResult, firstOption)
	}
}

func TestToggleSortOrder(t *testing.T) {
	// Test toggling from desc to asc
	if ToggleSortOrder(config.SortOrderDesc) != config.SortOrderAsc {
		t.Errorf("ToggleSortOrder(%q) should return %q", config.SortOrderDesc, config.SortOrderAsc)
	}

	// Test toggling from asc to desc
	if ToggleSortOrder(config.SortOrderAsc) != config.SortOrderDesc {
		t.Errorf("ToggleSortOrder(%q) should return %q", config.SortOrderAsc, config.SortOrderDesc)
	}

	// Test with invalid option - should return desc (anything not desc becomes desc)
	if ToggleSortOrder("invalid") != config.SortOrderDesc {
		t.Errorf("ToggleSortOrder(%q) should return %q", "invalid", config.SortOrderDesc)
	}
}

func TestFormatSortDisplay(t *testing.T) {
	tests := []struct {
		sort     config.SortOption
		order    config.SortOrder
		expected string
	}{
		{config.SortUpdated, config.SortOrderDesc, "Sort: Updated ↓"},
		{config.SortUpdated, config.SortOrderAsc, "Sort: Updated ↑"},
		{config.SortCreated, config.SortOrderDesc, "Sort: Created ↓"},
		{config.SortCreated, config.SortOrderAsc, "Sort: Created ↑"},
		{config.SortNumber, config.SortOrderDesc, "Sort: Number ↓"},
		{config.SortNumber, config.SortOrderAsc, "Sort: Number ↑"},
		{config.SortComments, config.SortOrderDesc, "Sort: Comments ↓"},
		{config.SortComments, config.SortOrderAsc, "Sort: Comments ↑"},
	}

	for _, tt := range tests {
		result := FormatSortDisplay(tt.sort, tt.order)
		if result != tt.expected {
			t.Errorf("FormatSortDisplay(%q, %q) = %q, expected %q", tt.sort, tt.order, result, tt.expected)
		}
	}
}

func TestFormatComments(t *testing.T) {
	comments := []db.Comment{
		{ID: 1, Body: "Test comment", Author: "user1", CreatedAt: "2024-01-15T10:30:00Z"},
		{ID: 2, Body: "Another comment", Author: "user2", CreatedAt: "2024-01-16T14:20:00Z"},
	}

	result := formatComments(comments, true)

	if !contains(result, "2 Comment(s)") {
		t.Error("formatComments should show comment count")
	}
	if !contains(result, "user1") {
		t.Error("formatComments should show first comment author")
	}
	if !contains(result, "user2") {
		t.Error("formatComments should show second comment author")
	}
	if !contains(result, "2024-01-15") {
		t.Error("formatComments should show first comment date")
	}
	if !contains(result, "Test comment") {
		t.Error("formatComments should show first comment body")
	}
}

func TestFormatCommentsEmpty(t *testing.T) {
	comments := []db.Comment{}

	result := formatComments(comments, true)

	if result != "No comments yet." {
		t.Errorf("formatComments with no comments = %q, expected %q", result, "No comments yet.")
	}
}

func TestFormatCommentsMarkdown(t *testing.T) {
	comments := []db.Comment{
		{ID: 1, Body: "**bold** and *italic*", Author: "user1", CreatedAt: "2024-01-15T10:30:00Z"},
	}

	// Test with markdown rendered
	renderedResult := formatComments(comments, true)
	if !contains(renderedResult, "bold") || !contains(renderedResult, "italic") {
		t.Error("formatComments should render markdown when enabled")
	}

	// Test with raw text
	rawResult := formatComments(comments, false)
	if !contains(rawResult, "**bold**") {
		t.Error("formatComments should show raw markdown when disabled")
	}
}