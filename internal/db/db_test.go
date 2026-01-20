package db

import (
	"testing"
)

func TestNewDB(t *testing.T) {
	tests := []struct {
		name        string
		dbPath      string
		expectError bool
	}{
		{
			name:        "Creates database successfully",
			dbPath:      ":memory:",
			expectError: false,
		},
		{
			name:        "Creates file-based database",
			dbPath:      "/tmp/test.db",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDB(tt.dbPath)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if db == nil {
				t.Error("Expected database instance, got nil")
			}

			// Close the database
			if err := db.Close(); err != nil {
				t.Errorf("Failed to close database: %v", err)
			}
		})
	}
}

func TestDB_StoreAndGetIssues(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create a test issue
	issue := &Issue{
		Number:      1,
		Title:       "Test Issue",
		Body:        "This is a test issue",
		State:       "open",
		Author:      "testuser",
		CreatedAt:   "2026-01-20T10:00:00Z",
		UpdatedAt:   "2026-01-20T10:00:00Z",
		CommentCount: 5,
		Labels:      []string{"bug", "help wanted"},
		Assignees:   []string{"user1", "user2"},
	}

	// Store the issue
	if err := db.StoreIssue(issue); err != nil {
		t.Fatalf("Failed to store issue: %v", err)
	}

	// Retrieve the issue
	retrieved, err := db.GetIssue(1)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}

	if retrieved.Number != issue.Number {
		t.Errorf("Expected number %d, got %d", issue.Number, retrieved.Number)
	}
	if retrieved.Title != issue.Title {
		t.Errorf("Expected title %q, got %q", issue.Title, retrieved.Title)
	}
	if retrieved.Body != issue.Body {
		t.Errorf("Expected body %q, got %q", issue.Body, retrieved.Body)
	}
	if retrieved.Author != issue.Author {
		t.Errorf("Expected author %q, got %q", issue.Author, retrieved.Author)
	}
	if retrieved.CommentCount != issue.CommentCount {
		t.Errorf("Expected comment count %d, got %d", issue.CommentCount, retrieved.CommentCount)
	}

	// Check labels
	if len(retrieved.Labels) != len(issue.Labels) {
		t.Errorf("Expected %d labels, got %d", len(issue.Labels), len(retrieved.Labels))
	}

	// Check assignees
	if len(retrieved.Assignees) != len(issue.Assignees) {
		t.Errorf("Expected %d assignees, got %d", len(issue.Assignees), len(retrieved.Assignees))
	}
}

func TestDB_StoreAndGetComments(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store an issue first
	issue := &Issue{
		Number: 1,
		Title:  "Test Issue",
		Body:   "Test body",
		State:  "open",
		Author: "testuser",
	}
	if err := db.StoreIssue(issue); err != nil {
		t.Fatalf("Failed to store issue: %v", err)
	}

	// Create test comments
	comments := []*Comment{
		{
			ID:        100,
			IssueNum:  1,
			Body:      "First comment",
			Author:    "user1",
			CreatedAt: "2026-01-20T10:00:00Z",
		},
		{
			ID:        101,
			IssueNum:  1,
			Body:      "Second comment",
			Author:    "user2",
			CreatedAt: "2026-01-20T10:05:00Z",
		},
	}

	// Store comments
	for _, comment := range comments {
		if err := db.StoreComment(comment); err != nil {
			t.Fatalf("Failed to store comment: %v", err)
		}
	}

	// Retrieve comments for the issue
	retrieved, err := db.GetComments(1)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(retrieved) != len(comments) {
		t.Errorf("Expected %d comments, got %d", len(comments), len(retrieved))
	}

	// Verify first comment
	if retrieved[0].Body != comments[0].Body {
		t.Errorf("Expected body %q, got %q", comments[0].Body, retrieved[0].Body)
	}
	if retrieved[0].Author != comments[0].Author {
		t.Errorf("Expected author %q, got %q", comments[0].Author, retrieved[0].Author)
	}
}

func TestDB_GetAllOpenIssues(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store multiple issues
	issues := []*Issue{
		{Number: 1, Title: "Open Issue 1", Body: "Body 1", State: "open", Author: "user1"},
		{Number: 2, Title: "Closed Issue", Body: "Body 2", State: "closed", Author: "user2"},
		{Number: 3, Title: "Open Issue 2", Body: "Body 3", State: "open", Author: "user3"},
		{Number: 4, Title: "Open Issue 3", Body: "Body 4", State: "open", Author: "user4"},
	}

	for _, issue := range issues {
		if err := db.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store issue: %v", err)
		}
	}

	// Get all open issues
	openIssues, err := db.GetAllOpenIssues()
	if err != nil {
		t.Fatalf("Failed to get open issues: %v", err)
	}

	// Should only return open issues
	if len(openIssues) != 3 {
		t.Errorf("Expected 3 open issues, got %d", len(openIssues))
	}

	// Verify all returned issues are open
	for _, issue := range openIssues {
		if issue.State != "open" {
			t.Errorf("Expected open state, got %q", issue.State)
		}
	}
}

func TestDB_ClearAllIssues(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store some issues
	issues := []*Issue{
		{Number: 1, Title: "Issue 1", Body: "Body 1", State: "open", Author: "user1"},
		{Number: 2, Title: "Issue 2", Body: "Body 2", State: "open", Author: "user2"},
	}

	for _, issue := range issues {
		if err := db.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store issue: %v", err)
		}
	}

	// Clear all issues
	if err := db.ClearAllIssues(); err != nil {
		t.Fatalf("Failed to clear issues: %v", err)
	}

	// Verify all issues are gone
	openIssues, err := db.GetAllOpenIssues()
	if err != nil {
		t.Fatalf("Failed to get open issues: %v", err)
	}

	if len(openIssues) != 0 {
		t.Errorf("Expected 0 issues after clearing, got %d", len(openIssues))
	}
}
