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
		Number:       1,
		Title:        "Test Issue",
		Body:         "This is a test issue",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2026-01-20T10:00:00Z",
		UpdatedAt:    "2026-01-20T10:00:00Z",
		CommentCount: 5,
		Labels:       []string{"bug", "help wanted"},
		Assignees:    []string{"user1", "user2"},
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

func TestDB_GetIssuesForDisplaySorted(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store issues with different dates and comment counts
	issues := []*Issue{
		{
			Number:       1,
			Title:        "Issue 1",
			Body:         "Body 1",
			State:        "open",
			Author:       "user1",
			CreatedAt:    "2026-01-20T10:00:00Z",
			UpdatedAt:    "2026-01-22T15:30:00Z",
			CommentCount: 5,
		},
		{
			Number:       2,
			Title:        "Issue 2",
			Body:         "Body 2",
			State:        "open",
			Author:       "user2",
			CreatedAt:    "2026-01-19T09:00:00Z",
			UpdatedAt:    "2026-01-23T12:00:00Z",
			CommentCount: 10,
		},
		{
			Number:       3,
			Title:        "Issue 3",
			Body:         "Body 3",
			State:        "open",
			Author:       "user3",
			CreatedAt:    "2026-01-21T11:00:00Z",
			UpdatedAt:    "2026-01-21T08:00:00Z",
			CommentCount: 2,
		},
	}

	for _, issue := range issues {
		if err := db.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store issue: %v", err)
		}
	}

	tests := []struct {
		name          string
		sortField     string
		descending    bool
		expectedOrder []int // Expected issue numbers in order
	}{
		{
			name:          "Sort by updated_at descending (default)",
			sortField:     "updated_at",
			descending:    true,
			expectedOrder: []int{2, 1, 3}, // Most recent: #2 (2026-01-23), then #1 (2026-01-22), then #3 (2026-01-21)
		},
		{
			name:          "Sort by updated_at ascending",
			sortField:     "updated_at",
			descending:    false,
			expectedOrder: []int{3, 1, 2}, // Oldest: #3 (2026-01-21), then #1 (2026-01-22), then #2 (2026-01-23)
		},
		{
			name:          "Sort by created_at descending",
			sortField:     "created_at",
			descending:    true,
			expectedOrder: []int{3, 1, 2}, // Most recent: #3 (2026-01-21), then #1 (2026-01-20), then #2 (2026-01-19)
		},
		{
			name:          "Sort by created_at ascending",
			sortField:     "created_at",
			descending:    false,
			expectedOrder: []int{2, 1, 3}, // Oldest: #2 (2026-01-19), then #1 (2026-01-20), then #3 (2026-01-21)
		},
		{
			name:          "Sort by number descending",
			sortField:     "number",
			descending:    true,
			expectedOrder: []int{3, 2, 1}, // #3, #2, #1
		},
		{
			name:          "Sort by number ascending",
			sortField:     "number",
			descending:    false,
			expectedOrder: []int{1, 2, 3}, // #1, #2, #3
		},
		{
			name:          "Sort by comment_count descending",
			sortField:     "comment_count",
			descending:    true,
			expectedOrder: []int{2, 1, 3}, // Most comments: #2 (10), then #1 (5), then #3 (2)
		},
		{
			name:          "Sort by comment_count ascending",
			sortField:     "comment_count",
			descending:    false,
			expectedOrder: []int{3, 1, 2}, // Fewest comments: #3 (2), then #1 (5), then #2 (10)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, err := db.GetIssuesForDisplaySorted(tt.sortField, tt.descending)
			if err != nil {
				t.Fatalf("Failed to get issues for display: %v", err)
			}

			if len(issues) != len(tt.expectedOrder) {
				t.Fatalf("Expected %d issues, got %d", len(tt.expectedOrder), len(issues))
			}

			for i, issue := range issues {
				if issue.Number != tt.expectedOrder[i] {
					t.Errorf("Expected issue #%d at index %d, got issue #%d", tt.expectedOrder[i], i, issue.Number)
				}
			}
		})
	}
}

func TestDB_SyncDateOperations(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test getting last sync date when none exists
	date, err := db.GetLastSyncDate()
	if err != nil {
		t.Fatalf("Failed to get last sync date: %v", err)
	}

	// Should return default date when no sync has happened
	expected := "1970-01-01T00:00:00Z"
	if date != expected {
		t.Errorf("Expected default date %s, got %s", expected, date)
	}

	// Test setting last sync date for the first time (full sync needed)
	testDate := "2026-01-20T15:30:00Z"
	fullSync, err := db.SetLastSyncDate(testDate)
	if err != nil {
		t.Fatalf("Failed to set last sync date: %v", err)
	}

	if !fullSync {
		t.Error("Expected full sync to be true on first sync")
	}

	// Test getting the sync date we just set
	date, err = db.GetLastSyncDate()
	if err != nil {
		t.Fatalf("Failed to get last sync date: %v", err)
	}

	if date != testDate {
		t.Errorf("Expected date %s, got %s", testDate, date)
	}

	// Test updating sync date (incremental sync this time)
	newTestDate := "2026-01-20T16:45:00Z"
	fullSync, err = db.SetLastSyncDate(newTestDate)
	if err != nil {
		t.Fatalf("Failed to update last sync date: %v", err)
	}

	if fullSync {
		t.Error("Expected full sync to be false on subsequent syncs")
	}

	// Verify the date was updated
	date, err = db.GetLastSyncDate()
	if err != nil {
		t.Fatalf("Failed to get last sync date: %v", err)
	}

	if date != newTestDate {
		t.Errorf("Expected updated date %s, got %s", newTestDate, date)
	}
}

func TestDB_RemoveIssues(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store some issues
	issues := []*Issue{
		{Number: 1, Title: "Issue 1", Body: "Body 1", State: "open", Author: "user1"},
		{Number: 2, Title: "Issue 2", Body: "Body 2", State: "open", Author: "user2"},
		{Number: 3, Title: "Issue 3", Body: "Body 3", State: "open", Author: "user3"},
	}

	for _, issue := range issues {
		if err := db.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store issue: %v", err)
		}
	}

	// Remove middle issue
	err = db.RemoveIssues([]int{2})
	if err != nil {
		t.Fatalf("Failed to remove issue: %v", err)
	}

	// Check that only two issues remain
	openIssues, err := db.GetAllOpenIssues()
	if err != nil {
		t.Fatalf("Failed to get open issues: %v", err)
	}

	if len(openIssues) != 2 {
		t.Errorf("Expected 2 issues after removal, got %d", len(openIssues))
	}

	// Verify the removed issue is gone
	for _, issue := range openIssues {
		if issue.Number == 2 {
			t.Error("Issue 2 should have been removed")
		}
	}

	// Remove remaining issues
	err = db.RemoveIssues([]int{1, 3})
	if err != nil {
		t.Fatalf("Failed to remove issues: %v", err)
	}

	// Check that no issues remain
	openIssues, err = db.GetAllOpenIssues()
	if err != nil {
		t.Fatalf("Failed to get open issues: %v", err)
	}

	if len(openIssues) != 0 {
		t.Errorf("Expected 0 issues after removing all, got %d", len(openIssues))
	}

	// Test removing empty list (should not error)
	err = db.RemoveIssues([]int{})
	if err != nil {
		t.Errorf("Removing empty list should not error: %v", err)
	}
}

func TestDB_GetAllIssueNumbers(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Store some issues with mixed states
	issues := []*Issue{
		{Number: 1, Title: "Open Issue 1", Body: "Body 1", State: "open", Author: "user1"},
		{Number: 2, Title: "Closed Issue", Body: "Body 2", State: "closed", Author: "user2"},
		{Number: 3, Title: "Open Issue 2", Body: "Body 3", State: "open", Author: "user3"},
	}

	for _, issue := range issues {
		if err := db.StoreIssue(issue); err != nil {
			t.Fatalf("Failed to store issue: %v", err)
		}
	}

	// Get all issue numbers
	numbers, err := db.GetAllIssueNumbers()
	if err != nil {
		t.Fatalf("Failed to get issue numbers: %v", err)
	}

	// Should only return open issues
	if len(numbers) != 2 {
		t.Errorf("Expected 2 open issue numbers, got %d", len(numbers))
	}

	// Verify we got the correct numbers (open issues)
	expected := map[int]bool{1: true, 3: true}
	for _, num := range numbers {
		if !expected[num] {
			t.Errorf("Unexpected issue number %d returned", num)
		}
	}
}
