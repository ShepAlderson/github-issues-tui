package tui

import (
	"bytes"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/git/github-issues-tui/internal/config"
	"github.com/shepbook/git/github-issues-tui/internal/db"
)

func TestNewListModel(t *testing.T) {
	// Create a test database file
	dbPath := "/tmp/ghissues_test.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		// Clean up test database
		os.Remove(dbPath)
	}()

	// Insert test issues
	testIssues := []*db.Issue{
		{Number: 1, Title: "Test Issue 1", Author: "user1", CreatedAt: "2024-01-01T10:00:00Z", CommentCount: 5, State: "open"},
		{Number: 2, Title: "Test Issue 2", Author: "user2", CreatedAt: "2024-01-02T11:00:00Z", CommentCount: 3, State: "open"},
		{Number: 3, Title: "Test Issue 3", Author: "user1", CreatedAt: "2024-01-03T12:00:00Z", CommentCount: 0, State: "open"},
	}

	for _, issue := range testIssues {
		if err := database.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store test issue: %v", err)
		}
	}

	// Create config with default columns
	cfg := &config.Config{}
	cfg.Display.Columns = config.GetDefaultDisplayColumns()

	// Test NewListModel
	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model: %v", err)
	}
	defer model.Close()

	// Verify issues were loaded
	if len(model.issues) != len(testIssues) {
		t.Errorf("Expected %d issues, got %d", len(testIssues), len(model.issues))
	}

	// Verify table was created
	if model.table.Columns() == nil || len(model.table.Columns()) == 0 {
		t.Error("Table columns not initialized")
	}

	if model.table.Rows() == nil || len(model.table.Rows()) == 0 {
		t.Error("Table rows not initialized")
	}
}

func TestNewListModel_NoIssues(t *testing.T) {
	// Create an empty database
	dbPath := "/tmp/ghissues_test_no_issues.db"
	cfg := &config.Config{}
	cfg.Display.Columns = config.GetDefaultDisplayColumns()
	defer os.Remove(dbPath)

	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model: %v", err)
	}
	defer model.Close()

	// Verify no issues
	if len(model.issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(model.issues))
	}
}

func TestListModel_Navigation(t *testing.T) {
	// Create a test database with issues
	dbPath := "/tmp/ghissues_test_nav.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issues
	for i := 1; i <= 5; i++ {
		issue := &db.Issue{
			Number:       i,
			Title:        "Test Issue",
			Author:       "user",
			CreatedAt:    "2024-01-01T10:00:00Z",
			CommentCount: 0,
			State:        "open",
		}
		if err := database.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store test issue: %v", err)
		}
	}

	cfg := &config.Config{}
	cfg.Display.Columns = config.GetDefaultDisplayColumns()

	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model: %v", err)
	}
	defer model.Close()

	// Test initial selection
	if model.selected != 0 {
		t.Errorf("Expected initial selection to be 0, got %d", model.selected)
	}

	// Test moving down with 'j'
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(*ListModel)
	if model.selected != 1 {
		t.Errorf("Expected selection to be 1 after moving down, got %d", model.selected)
	}

	// Test moving down with arrow key
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(*ListModel)
	if model.selected != 2 {
		t.Errorf("Expected selection to be 2 after moving down, got %d", model.selected)
	}

	// Test moving up with 'k'
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(*ListModel)
	if model.selected != 1 {
		t.Errorf("Expected selection to be 1 after moving up, got %d", model.selected)
	}

	// Test moving up with arrow key
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(*ListModel)
	if model.selected != 0 {
		t.Errorf("Expected selection to be 0 after moving up, got %d", model.selected)
	}
}

func TestListModel_Quit(t *testing.T) {
	dbPath := "/tmp/ghissues_test_quit.db"
	cfg := &config.Config{}
	cfg.Display.Columns = config.GetDefaultDisplayColumns()
	defer os.Remove(dbPath)

	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model: %v", err)
	}
	defer model.Close()

	// Test quitting with 'q'
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(*ListModel)
	if !model.quitting {
		t.Error("Expected model to be quitting after pressing 'q'")
	}
	if cmd == nil {
		t.Error("Expected quit command after pressing 'q'")
	}

	// Reset and test quitting with Ctrl+C
	model.quitting = false
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model = updated.(*ListModel)
	if !model.quitting {
		t.Error("Expected model to be quitting after Ctrl+C")
	}
	if cmd == nil {
		t.Error("Expected quit command after Ctrl+C")
	}
}

func TestFormatIssueField(t *testing.T) {
	tests := []struct {
		name     string
		issue    *db.Issue
		field    string
		expected string
	}{
		{
			name:     "Format number",
			issue:    &db.Issue{Number: 123},
			field:    "number",
			expected: "#123",
		},
		{
			name:     "Format author",
			issue:    &db.Issue{Author: "testuser"},
			field:    "author",
			expected: "testuser",
		},
		{
			name:     "Format comment count",
			issue:    &db.Issue{CommentCount: 42},
			field:    "comment_count",
			expected: "42",
		},
		{
			name:     "Format created date",
			issue:    &db.Issue{CreatedAt: "2024-01-15T10:30:00Z"},
			field:    "created_at",
			expected: "2024-01-15",
		},
		{
			name:     "Truncate long title",
			issue:    &db.Issue{Title: "This is a very long title that should be truncated to fit in the display"},
			field:    "title",
			expected: "This is a very long title that should be truncat...",
		},
		{
			name:     "Short title unchanged",
			issue:    &db.Issue{Title: "Short title"},
			field:    "title",
			expected: "Short title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatIssueField(tt.issue, tt.field)
			if result != tt.expected {
				t.Errorf("formatIssueField() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatColumnTitle(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected string
	}{
		{"Number column", "number", "#"},
		{"Title column", "title", "Title"},
		{"Author column", "author", "Author"},
		{"Created column", "created_at", "Created"},
		{"Comments column", "comment_count", "Comments"},
		{"Unknown column", "unknown_field", "Unknown Field"},
		{"Multi-word column", "very_long_field_name", "Very Long Field Name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatColumnTitle(tt.column)
			if result != tt.expected {
				t.Errorf("formatColumnTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetColumnWidth(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected int
	}{
		{"Number width", "number", 6},
		{"Author width", "author", 15},
		{"Created width", "created_at", 12},
		{"Comments width", "comment_count", 10},
		{"Default width", "title", 30},
		{"Unknown column width", "unknown", 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColumnWidth(tt.column)
			if result != tt.expected {
				t.Errorf("getColumnWidth() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestListModel_View(t *testing.T) {
	// Create an empty model
	dbPath := "/tmp/ghissues_test_view1.db"
	cfg := &config.Config{}
	cfg.Display.Columns = config.GetDefaultDisplayColumns()
	defer os.Remove(dbPath)

	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model: %v", err)
	}
	defer model.Close()

	// Test view with no issues
	view := model.View()
	if !bytes.Contains([]byte(view), []byte("No issues found")) {
		t.Error("Expected 'No issues found' message in view")
	}

	// Create a model with issues
	dbPath2 := "/tmp/ghissues_test_view2.db"
	database, err := db.NewDB(dbPath2)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath2)
	}()

	issue := &db.Issue{
		Number:       1,
		Title:        "Test Issue",
		Author:       "testuser",
		CreatedAt:    "2024-01-01T10:00:00Z",
		CommentCount: 5,
		State:        "open",
	}
	if err := database.StoreIssue(issue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	modelWithIssues, err := NewListModel(dbPath2, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model with issues: %v", err)
	}
	defer modelWithIssues.Close()

	// Test view with issues
	view = modelWithIssues.View()
	if !bytes.Contains([]byte(view), []byte("#1")) {
		t.Error("Expected issue number in view")
	}
	if !bytes.Contains([]byte(view), []byte("Test Issue")) {
		t.Error("Expected issue title in view")
	}
	if !bytes.Contains([]byte(view), []byte("1/1 issues")) {
		t.Error("Expected issue count in status bar")
	}
}

func TestListModel_Sort(t *testing.T) {
	// Create a test database with issues having different dates
	dbPath := "/tmp/ghissues_test_sort.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issues with different dates and comment counts
	testIssues := []*db.Issue{
		{
			Number:       1,
			Title:        "Old Issue",
			Author:       "user1",
			CreatedAt:    "2024-01-01T10:00:00Z",
			UpdatedAt:    "2024-01-01T10:00:00Z",
			CommentCount: 5,
			State:        "open",
		},
		{
			Number:       2,
			Title:        "Recent Issue",
			Author:       "user2",
			CreatedAt:    "2024-01-03T12:00:00Z",
			UpdatedAt:    "2024-01-03T12:00:00Z",
			CommentCount: 10,
			State:        "open",
		},
		{
			Number:       3,
			Title:        "Middle Issue",
			Author:       "user3",
			CreatedAt:    "2024-01-02T11:00:00Z",
			UpdatedAt:    "2024-01-02T11:00:00Z",
			CommentCount: 2,
			State:        "open",
		},
	}

	for _, issue := range testIssues {
		if err := database.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store test issue: %v", err)
		}
	}

	// Create config with default sort (updated_at descending)
	cfg := &config.Config{}
	cfg.Display.Columns = config.GetDefaultDisplayColumns()
	cfg.Display.Sort = config.Sort{
		Field:      "updated_at",
		Descending: true,
	}

	model, err := NewListModel(dbPath, cfg)
	if err != nil {
		t.Fatalf("Failed to create list model: %v", err)
	}
	defer model.Close()

	// Verify initial sort (updated_at descending)
	if model.sortField != "updated_at" {
		t.Errorf("Expected initial sort field 'updated_at', got %q", model.sortField)
	}
	if !model.sortDescending {
		t.Error("Expected initial sort direction to be descending")
	}

	// Test cycle sort (should change to next option)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updated.(*ListModel)

	if model.sortField != "created_at" {
		t.Errorf("Expected sort field 'created_at' after cycling, got %q", model.sortField)
	}

	// Test cycle sort again
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updated.(*ListModel)

	if model.sortField != "number" {
		t.Errorf("Expected sort field 'number' after second cycle, got %q", model.sortField)
	}

	// Test toggle direction (S)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	model = updated.(*ListModel)

	if model.sortField != "number" {
		t.Errorf("Expected sort field to remain 'number' after toggling direction, got %q", model.sortField)
	}
	if model.sortDescending {
		t.Error("Expected sort direction to be ascending after toggle")
	}

	// Test status bar shows sort indicator
	view := model.View()
	if !bytes.Contains([]byte(view), []byte("number")) {
		t.Error("Expected sort field in status bar")
	}
	if !bytes.Contains([]byte(view), []byte("↑")) {
		t.Error("Expected ascending indicator in status bar")
	}
}
