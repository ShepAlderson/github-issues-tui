package tui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestSortPersistence(t *testing.T) {
	// Create a temporary directory for config
	tempDir := t.TempDir()

	// Create a test config manager pointing to temp directory
	cfgMgr := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	// Create default config
	cfg := config.DefaultConfig()

	// Modify sort settings in config
	cfg.Display.SortField = "number"
	cfg.Display.SortAscending = true

	// Save initial config
	if err := cfgMgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Create database manager
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Create issue list - should read sort settings from config
	issueList := NewIssueList(dbManager, cfg, cfgMgr)
	if issueList == nil {
		t.Fatal("NewIssueList returned nil")
	}

	// Verify issue list read config values
	if issueList.sortField != "number" {
		t.Errorf("Expected sortField 'number' from config, got %s", issueList.sortField)
	}
	if issueList.sortAscending != true {
		t.Error("Expected sortAscending true from config")
	}

	// Change sort settings through issue list
	issueList.sortField = "comments"
	issueList.sortAscending = false
	issueList.updateConfigAndSave()

	// Load config again to verify it was saved
	loadedCfg, err := cfgMgr.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config was updated
	if loadedCfg.Display.SortField != "comments" {
		t.Errorf("Expected saved SortField 'comments', got %s", loadedCfg.Display.SortField)
	}
	if loadedCfg.Display.SortAscending != false {
		t.Error("Expected saved SortAscending false")
	}

	// Test that config file actually exists and has correct content
	configPath, err := cfgMgr.ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file doesn't exist: %v", err)
	}

	// Read raw config file to verify TOML serialization
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Basic check that file contains our sort settings
	content := string(data)
	// The TOML might have formatting differences, just check key parts
	if !strings.Contains(content, "sort_field") {
		t.Error("Config file should contain 'sort_field'")
	}
	if !strings.Contains(content, "comments") {
		t.Error("Config file should contain 'comments' value")
	}
	if !strings.Contains(content, "sort_ascending") {
		t.Error("Config file should contain 'sort_ascending'")
	}
}

func TestSortPersistenceThroughUI(t *testing.T) {
	// Create a temporary directory for config
	tempDir := t.TempDir()

	// Create a test config manager pointing to temp directory
	cfgMgr := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	// Create default config
	cfg := config.DefaultConfig()

	// Save initial config
	if err := cfgMgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	// Create database manager
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Create issue list
	issueList := NewIssueList(dbManager, cfg, cfgMgr)
	if issueList == nil {
		t.Fatal("NewIssueList returned nil")
	}

	// Load dummy issues for testing
	now := time.Now()
	issueList.issues = []IssueItem{
		{Number: 1, TitleText: "Test A", Author: "alice", Created: now.Add(-24 * time.Hour), Updated: now.Add(-2 * time.Hour), CommentCount: 3},
		{Number: 2, TitleText: "Test B", Author: "bob", Created: now.Add(-48 * time.Hour), Updated: now.Add(-4 * time.Hour), CommentCount: 5},
	}
	issueList.sortIssues()

	// Simulate pressing 's' to cycle sort field (should save to config)
	issueList.cycleSortField()

	// Load config to verify it was saved
	loadedCfg, err := cfgMgr.Load()
	if err != nil {
		t.Fatalf("Failed to load config after cycle: %v", err)
	}

	// Should have cycled from default "updated" to "created"
	if loadedCfg.Display.SortField != "created" {
		t.Errorf("After cycling, expected SortField 'created', got %s", loadedCfg.Display.SortField)
	}

	// Simulate pressing 'S' to toggle sort order
	issueList.toggleSortOrder()

	// Load config again
	loadedCfg2, err := cfgMgr.Load()
	if err != nil {
		t.Fatalf("Failed to load config after toggle: %v", err)
	}

	if loadedCfg2.Display.SortAscending != true {
		t.Error("After toggling, expected SortAscending true")
	}
}
