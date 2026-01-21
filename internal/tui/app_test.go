package tui

import (
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestNewApp(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create a temporary database for testing
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	app := NewApp(cfg, dbManager)
	if app == nil {
		t.Error("NewApp returned nil")
	}
}

func TestNewIssueList(t *testing.T) {
	cfg := config.DefaultConfig()

	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	issueList := NewIssueList(dbManager, cfg)
	if issueList == nil {
		t.Error("NewIssueList returned nil")
	}

	// Check that column renderer was created
	if issueList.columnRenderer == nil {
		t.Error("IssueList columnRenderer is nil")
	}
}

func TestColumnRenderer(t *testing.T) {
	columnKeys := []string{"number", "title", "author"}
	renderer := NewColumnRenderer(columnKeys)

	if renderer == nil {
		t.Fatal("NewColumnRenderer returned nil")
	}

	if len(renderer.columns) != len(columnKeys) {
		t.Errorf("Expected %d columns, got %d", len(columnKeys), len(renderer.columns))
	}

	// Test header rendering
	header := renderer.RenderHeader()
	if header == "" {
		t.Error("RenderHeader returned empty string")
	}
}