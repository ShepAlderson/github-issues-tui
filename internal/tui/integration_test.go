package tui

import (
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/storage"
	"time"
)

func TestIntegration_ColumnConfigFromConfigFile(t *testing.T) {
	// Test that column configuration can be read from config
	cfg := &config.Config{
		Display: config.DisplayConfig{
			Columns: []string{"number", "title", "comments"},
		},
	}

	columns := GetDefaultColumns(cfg)
	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	if columns[0].Name != "number" {
		t.Errorf("Expected first column to be 'number', got '%s'", columns[0].Name)
	}
}

func TestIntegration_ModelWithEmptyIssueList(t *testing.T) {
	// Test that the model handles empty issue lists gracefully
	issues := []storage.Issue{}
	columns := []Column{{Name: "number", Width: 7, Title: "#"}}

	model := NewModel(issues, columns)

	if model.IssueList == nil {
		t.Fatal("Expected IssueList to be initialized")
	}

	if len(model.IssueList.Issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(model.IssueList.Issues))
	}
}

func TestIntegration_ModelWithIssues(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{
			Number:    1,
			Title:     "Test Issue",
			Author:    "testuser",
			State:     "open",
			CreatedAt: now,
			Comments:  5,
		},
	}

	columns := GetDefaultColumns(&config.Config{})
	model := NewModel(issues, columns)

	if len(model.IssueList.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(model.IssueList.Issues))
	}

	if model.IssueList.Cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", model.IssueList.Cursor)
	}
}

func TestIntegration_AllDefaultColumnsExist(t *testing.T) {
	// Verify all default columns are valid
	defaultColumns := []string{"number", "title", "author", "date", "comments"}

	for _, colName := range defaultColumns {
		if !validColumns[colName] {
			t.Errorf("Default column '%s' is not in valid columns map", colName)
		}
	}

	if !ValidateColumnConfig(defaultColumns) {
		t.Error("Default column configuration is not valid")
	}
}
