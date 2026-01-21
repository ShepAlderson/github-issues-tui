package tui

import (
	"testing"
	"time"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestIssueListSorting(t *testing.T) {
	cfg := config.DefaultConfig()
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Create a test config manager
	cfgMgr := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})

	issueList := NewIssueList(dbManager, cfg, cfgMgr)
	if issueList == nil {
		t.Fatal("NewIssueList returned nil")
	}

	// Create test issues with varied data for sorting
	now := time.Now()
	testIssues := []IssueItem{
		{
			Number:       3,
			TitleText:    "Issue C",
			Author:       "Charlie",
			Created:      now.Add(-72 * time.Hour),
			Updated:      now.Add(-24 * time.Hour),
			CommentCount: 5,
		},
		{
			Number:       1,
			TitleText:    "Issue A",
			Author:       "Alice",
			Created:      now.Add(-48 * time.Hour),
			Updated:      now.Add(-12 * time.Hour),
			CommentCount: 2,
		},
		{
			Number:       2,
			TitleText:    "Issue B",
			Author:       "Bob",
			Created:      now.Add(-96 * time.Hour),
			Updated:      now.Add(-36 * time.Hour),
			CommentCount: 8,
		},
	}

	// Set test issues
	issueList.issues = testIssues

	// Test default sort (updated date, descending - most recent first)
	issueList.sortField = "updated"
	issueList.sortAscending = false
	issueList.sortIssues()

	// Check that issues are sorted by updated date descending
	if issueList.issues[0].Number != 1 {
		t.Errorf("Expected issue #1 first (most recently updated), got #%d", issueList.issues[0].Number)
	}
	if issueList.issues[1].Number != 3 {
		t.Errorf("Expected issue #3 second, got #%d", issueList.issues[1].Number)
	}
	if issueList.issues[2].Number != 2 {
		t.Errorf("Expected issue #2 last (least recently updated), got #%d", issueList.issues[2].Number)
	}

	// Test ascending order
	issueList.sortAscending = true
	issueList.sortIssues()

	if issueList.issues[0].Number != 2 {
		t.Errorf("Expected issue #2 first (least recently updated ascending), got #%d", issueList.issues[0].Number)
	}

	// Test sort by created date
	issueList.sortField = "created"
	issueList.sortAscending = false
	issueList.sortIssues()

	if issueList.issues[0].Number != 1 {
		t.Errorf("Expected issue #1 first (most recently created), got #%d", issueList.issues[0].Number)
	}

	// Test sort by number
	issueList.sortField = "number"
	issueList.sortAscending = true
	issueList.sortIssues()

	if issueList.issues[0].Number != 1 {
		t.Errorf("Expected issue #1 first (number ascending), got #%d", issueList.issues[0].Number)
	}
	if issueList.issues[2].Number != 3 {
		t.Errorf("Expected issue #3 last (number ascending), got #%d", issueList.issues[2].Number)
	}

	// Test sort by comment count
	issueList.sortField = "comments"
	issueList.sortAscending = false
	issueList.sortIssues()

	if issueList.issues[0].Number != 2 {
		t.Errorf("Expected issue #2 first (most comments), got #%d", issueList.issues[0].Number)
	}
	if issueList.issues[2].Number != 1 {
		t.Errorf("Expected issue #1 last (fewest comments), got #%d", issueList.issues[2].Number)
	}
}

func TestIssueListSortCycle(t *testing.T) {
	cfg := config.DefaultConfig()
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Create a test config manager
	cfgMgr := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})

	issueList := NewIssueList(dbManager, cfg, cfgMgr)
	if issueList == nil {
		t.Fatal("NewIssueList returned nil")
	}

	// Test initial state
	if issueList.sortField != "updated" {
		t.Errorf("Expected default sort field 'updated', got %s", issueList.sortField)
	}
	if issueList.sortAscending != false {
		t.Error("Expected default sort ascending false (descending)")
	}

	// Test cycling through sort fields
	// Start with default "updated", first cycle should go to "created"
	expectedCycle := []string{"created", "number", "comments", "updated"}
	for i, expectedField := range expectedCycle {
		// Simulate pressing 's' key
		issueList.cycleSortField()
		if issueList.sortField != expectedField {
			t.Errorf("After %d cycles, expected sort field %s, got %s", i+1, expectedField, issueList.sortField)
		}
	}

	// Should wrap back to "created" (we've cycled 4 times, next is first in cycle)
	issueList.cycleSortField()
	if issueList.sortField != "created" {
		t.Errorf("Expected wrap to 'created', got %s", issueList.sortField)
	}
}

func TestIssueListSortReverse(t *testing.T) {
	cfg := config.DefaultConfig()
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Create a test config manager
	cfgMgr := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})

	issueList := NewIssueList(dbManager, cfg, cfgMgr)
	if issueList == nil {
		t.Fatal("NewIssueList returned nil")
	}

	// Test initial state
	if issueList.sortAscending != false {
		t.Error("Expected default sort ascending false (descending)")
	}

	// Test reverse (toggle ascending)
	issueList.toggleSortOrder()
	if issueList.sortAscending != true {
		t.Error("Expected sort ascending true after toggle")
	}

	// Test reverse again
	issueList.toggleSortOrder()
	if issueList.sortAscending != false {
		t.Error("Expected sort ascending false after second toggle")
	}
}

func TestIssueListSortStatus(t *testing.T) {
	cfg := config.DefaultConfig()
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Create a test config manager
	cfgMgr := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})

	issueList := NewIssueList(dbManager, cfg, cfgMgr)
	if issueList == nil {
		t.Fatal("NewIssueList returned nil")
	}

	// Test status display for different sort configurations
	tests := []struct {
		field     string
		ascending bool
		expected  string
	}{
		{"updated", false, "↑ updated"},
		{"updated", true, "↓ updated"},
		{"created", false, "↑ created"},
		{"created", true, "↓ created"},
		{"number", false, "↑ number"},
		{"number", true, "↓ number"},
		{"comments", false, "↑ comments"},
		{"comments", true, "↓ comments"},
	}

	for _, test := range tests {
		issueList.sortField = test.field
		issueList.sortAscending = test.ascending
		status := issueList.sortStatus()
		if status != test.expected {
			t.Errorf("For field %s, ascending %v: expected %s, got %s", test.field, test.ascending, test.expected, status)
		}
	}
}
