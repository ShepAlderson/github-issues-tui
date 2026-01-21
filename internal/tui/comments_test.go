package tui

import (
	"testing"

	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestNewCommentsComponent(t *testing.T) {
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

	component := NewCommentsComponent(dbManager)
	if component == nil {
		t.Fatal("NewCommentsComponent returned nil")
	}

	if component.dbManager != dbManager {
		t.Error("dbManager not set correctly")
	}

	if component.comments != nil {
		t.Error("comments should be nil initially")
	}

	if component.currentIssueNumber != 0 {
		t.Error("currentIssueNumber should be 0 initially")
	}

	if component.showRawMarkdown {
		t.Error("showRawMarkdown should be false initially")
	}

	if component.scrollOffset != 0 {
		t.Error("scrollOffset should be 0 initially")
	}
}

func TestCommentsComponent_View_NoComments(t *testing.T) {
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	component := NewCommentsComponent(dbManager)

	view := component.View()
	if view == "" {
		t.Error("View returned empty string")
	}

	// Should contain placeholder text
	if len(view) < 10 {
		t.Error("View should render placeholder text")
	}
}

func TestCommentsComponent_SetIssue(t *testing.T) {
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Initialize schema
	if err := dbManager.InitializeSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	component := NewCommentsComponent(dbManager)

	// Test with invalid database manager (nil)
	component.dbManager = nil
	err = component.SetIssue(1, "Test Issue")
	if err == nil {
		t.Error("SetIssue should return error when dbManager is nil")
	}
}

func TestCommentsComponent_ToggleMarkdownView(t *testing.T) {
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	// Initialize schema
	if err := dbManager.InitializeSchema(); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	component := NewCommentsComponent(dbManager)

	// Initially should be false
	if component.showRawMarkdown {
		t.Error("showRawMarkdown should be false initially")
	}

	// Toggle with no comments should not change (function returns early)
	component.toggleMarkdownView()
	if component.showRawMarkdown {
		t.Error("showRawMarkdown should remain false when no comments")
	}

	// Add a dummy comment to test toggle
	component.comments = []database.Comment{{ID: 1, Author: "test", Body: "test"}}

	// Now toggle should work
	component.toggleMarkdownView()
	if !component.showRawMarkdown {
		t.Error("showRawMarkdown should be true after toggle when comments exist")
	}

	// Toggle again should set back to false
	component.toggleMarkdownView()
	if component.showRawMarkdown {
		t.Error("showRawMarkdown should be false after second toggle")
	}
}

func TestCommentsComponent_ScrollMethods(t *testing.T) {
	dbManager, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbManager.Close()

	component := NewCommentsComponent(dbManager)

	// Initial scroll offset should be 0
	if component.scrollOffset != 0 {
		t.Error("scrollOffset should be 0 initially")
	}

	// Test scrollDown with no comments (should not panic)
	component.scrollDown()
	if component.scrollOffset != 0 {
		t.Error("scrollOffset should remain 0 when no comments")
	}

	// Test scrollUp with no comments (should not panic)
	component.scrollOffset = 5
	component.scrollUp()
	if component.scrollOffset != 4 {
		t.Error("scrollOffset should decrement by 1 when scrolling up")
	}

	// Test scrollToTop
	component.scrollOffset = 10
	component.scrollToTop()
	if component.scrollOffset != 0 {
		t.Error("scrollToTop should set scrollOffset to 0")
	}

	// Test scrollToBottom with no comments (should not panic)
	component.scrollToBottom()
	if component.scrollOffset != 0 {
		t.Error("scrollOffset should be 0 when no comments")
	}
}