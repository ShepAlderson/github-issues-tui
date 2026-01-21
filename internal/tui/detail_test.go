package tui

import (
	"testing"

	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestNewIssueDetailComponent(t *testing.T) {
	// Create an in-memory database for testing
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Initialize schema
	if err := dbManager.InitializeSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	component := NewIssueDetailComponent(dbManager)
	if component == nil {
		t.Fatal("NewIssueDetailComponent returned nil")
	}

	if component.dbManager != dbManager {
		t.Error("dbManager not set correctly")
	}

	if component.currentIssue != nil {
		t.Error("currentIssue should be nil initially")
	}

	if component.showRawMarkdown {
		t.Error("showRawMarkdown should be false initially")
	}

	if component.scrollOffset != 0 {
		t.Error("scrollOffset should be 0 initially")
	}
}

func TestIssueDetailComponent_View_NoIssue(t *testing.T) {
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	component := NewIssueDetailComponent(dbManager)

	view := component.View()
	if view == "" {
		t.Error("View returned empty string")
	}

	// Should contain placeholder text
	if len(view) < 10 {
		t.Error("View should render placeholder text")
	}
}

// Note: More comprehensive tests would require:
// 1. Inserting test data into the database
// 2. Testing SetIssue with actual data
// 3. Testing keybindings and rendering
// However, those are more integration tests and would require
// a more complex test setup. For unit tests, we're testing
// the basic structure and initialization.