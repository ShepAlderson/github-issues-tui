package database

import (
	"path/filepath"
	"testing"
)

func TestListIssues(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	// Insert test issues
	testIssues := []Issue{
		{Number: 1, Title: "First Issue", Author: "alice", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-02T10:00:00Z", State: "open", CommentCount: 2},
		{Number: 2, Title: "Second Issue", Author: "bob", CreatedAt: "2024-01-03T10:00:00Z", UpdatedAt: "2024-01-04T10:00:00Z", State: "open", CommentCount: 0},
		{Number: 3, Title: "Third Issue", Author: "charlie", CreatedAt: "2024-01-05T10:00:00Z", UpdatedAt: "2024-01-06T10:00:00Z", State: "closed", CommentCount: 5},
	}

	for _, issue := range testIssues {
		if err := SaveIssue(db, "owner/repo", issue); err != nil {
			t.Fatalf("Failed to save issue %d: %v", issue.Number, err)
		}
	}

	t.Run("returns all issues", func(t *testing.T) {
		issues, err := ListIssues(db, "owner/repo")
		if err != nil {
			t.Fatalf("ListIssues failed: %v", err)
		}

		if len(issues) != 3 {
			t.Errorf("expected 3 issues, got %d", len(issues))
		}
	})

	t.Run("returns empty slice for non-existent repo", func(t *testing.T) {
		issues, err := ListIssues(db, "other/repo")
		if err != nil {
			t.Fatalf("ListIssues failed: %v", err)
		}

		if len(issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(issues))
		}
	})

	t.Run("issue data is correctly populated", func(t *testing.T) {
		issues, err := ListIssues(db, "owner/repo")
		if err != nil {
			t.Fatalf("ListIssues failed: %v", err)
		}

		// Find issue #2
		var issue2 *ListIssue
		for _, i := range issues {
			if i.Number == 2 {
				issue2 = &i
				break
			}
		}

		if issue2 == nil {
			t.Fatal("issue #2 not found")
		}

		if issue2.Title != "Second Issue" {
			t.Errorf("expected title 'Second Issue', got '%s'", issue2.Title)
		}

		if issue2.Author != "bob" {
			t.Errorf("expected author 'bob', got '%s'", issue2.Author)
		}

		if issue2.CommentCount != 0 {
			t.Errorf("expected 0 comments, got %d", issue2.CommentCount)
		}
	})
}

func TestListIssuesSortByUpdated(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	// Insert test issues with different updated_at times
	testIssues := []Issue{
		{Number: 1, Title: "Oldest", Author: "alice", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open"},
		{Number: 2, Title: "Middle", Author: "bob", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-02T10:00:00Z", State: "open"},
		{Number: 3, Title: "Newest", Author: "charlie", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-03T10:00:00Z", State: "open"},
	}

	for _, issue := range testIssues {
		if err := SaveIssue(db, "owner/repo", issue); err != nil {
			t.Fatalf("Failed to save issue %d: %v", issue.Number, err)
		}
	}

	t.Run("sorts by updated_at descending by default", func(t *testing.T) {
		issues, err := ListIssuesSorted(db, "owner/repo", "updated", true)
		if err != nil {
			t.Fatalf("ListIssuesSorted failed: %v", err)
		}

		if len(issues) != 3 {
			t.Fatalf("expected 3 issues, got %d", len(issues))
		}

		// Should be in descending order (newest first)
		if issues[0].Number != 3 {
			t.Errorf("expected first issue to be #3 (newest), got #%d", issues[0].Number)
		}
		if issues[2].Number != 1 {
			t.Errorf("expected last issue to be #1 (oldest), got #%d", issues[2].Number)
		}
	})

	t.Run("sorts by updated_at ascending", func(t *testing.T) {
		issues, err := ListIssuesSorted(db, "owner/repo", "updated", false)
		if err != nil {
			t.Fatalf("ListIssuesSorted failed: %v", err)
		}

		// Should be in ascending order (oldest first)
		if issues[0].Number != 1 {
			t.Errorf("expected first issue to be #1 (oldest), got #%d", issues[0].Number)
		}
		if issues[2].Number != 3 {
			t.Errorf("expected last issue to be #3 (newest), got #%d", issues[2].Number)
		}
	})
}

func TestListIssuesSortByNumber(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	// Insert test issues out of order
	testIssues := []Issue{
		{Number: 3, Title: "Third", Author: "charlie", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open"},
		{Number: 1, Title: "First", Author: "alice", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open"},
		{Number: 2, Title: "Second", Author: "bob", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open"},
	}

	for _, issue := range testIssues {
		if err := SaveIssue(db, "owner/repo", issue); err != nil {
			t.Fatalf("Failed to save issue %d: %v", issue.Number, err)
		}
	}

	t.Run("sorts by number ascending", func(t *testing.T) {
		issues, err := ListIssuesSorted(db, "owner/repo", "number", false)
		if err != nil {
			t.Fatalf("ListIssuesSorted failed: %v", err)
		}

		if issues[0].Number != 1 {
			t.Errorf("expected first issue to be #1, got #%d", issues[0].Number)
		}
		if issues[1].Number != 2 {
			t.Errorf("expected second issue to be #2, got #%d", issues[1].Number)
		}
		if issues[2].Number != 3 {
			t.Errorf("expected third issue to be #3, got #%d", issues[2].Number)
		}
	})
}

func TestListIssuesSortByCommentCount(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	// Insert test issues with different comment counts
	testIssues := []Issue{
		{Number: 1, Title: "No comments", Author: "alice", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open", CommentCount: 0},
		{Number: 2, Title: "Most comments", Author: "bob", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open", CommentCount: 10},
		{Number: 3, Title: "Few comments", Author: "charlie", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open", CommentCount: 3},
	}

	for _, issue := range testIssues {
		if err := SaveIssue(db, "owner/repo", issue); err != nil {
			t.Fatalf("Failed to save issue %d: %v", issue.Number, err)
		}
	}

	t.Run("sorts by comment count descending", func(t *testing.T) {
		issues, err := ListIssuesSorted(db, "owner/repo", "comments", true)
		if err != nil {
			t.Fatalf("ListIssuesSorted failed: %v", err)
		}

		if issues[0].CommentCount != 10 {
			t.Errorf("expected first issue to have 10 comments, got %d", issues[0].CommentCount)
		}
		if issues[1].CommentCount != 3 {
			t.Errorf("expected second issue to have 3 comments, got %d", issues[1].CommentCount)
		}
		if issues[2].CommentCount != 0 {
			t.Errorf("expected third issue to have 0 comments, got %d", issues[2].CommentCount)
		}
	})

	t.Run("sorts by comment count ascending", func(t *testing.T) {
		issues, err := ListIssuesSorted(db, "owner/repo", "comments", false)
		if err != nil {
			t.Fatalf("ListIssuesSorted failed: %v", err)
		}

		if issues[0].CommentCount != 0 {
			t.Errorf("expected first issue to have 0 comments, got %d", issues[0].CommentCount)
		}
		if issues[1].CommentCount != 3 {
			t.Errorf("expected second issue to have 3 comments, got %d", issues[1].CommentCount)
		}
		if issues[2].CommentCount != 10 {
			t.Errorf("expected third issue to have 10 comments, got %d", issues[2].CommentCount)
		}
	})
}

func TestListIssuesFiltersByState(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	// Insert test issues with different states
	testIssues := []Issue{
		{Number: 1, Title: "Open 1", Author: "alice", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open"},
		{Number: 2, Title: "Closed 1", Author: "bob", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "closed"},
		{Number: 3, Title: "Open 2", Author: "charlie", CreatedAt: "2024-01-01T10:00:00Z", UpdatedAt: "2024-01-01T10:00:00Z", State: "open"},
	}

	for _, issue := range testIssues {
		if err := SaveIssue(db, "owner/repo", issue); err != nil {
			t.Fatalf("Failed to save issue %d: %v", issue.Number, err)
		}
	}

	t.Run("filters by open state", func(t *testing.T) {
		issues, err := ListIssuesByState(db, "owner/repo", "open")
		if err != nil {
			t.Fatalf("ListIssuesByState failed: %v", err)
		}

		if len(issues) != 2 {
			t.Errorf("expected 2 open issues, got %d", len(issues))
		}

		for _, issue := range issues {
			if issue.State != "open" {
				t.Errorf("expected only open issues, found state '%s'", issue.State)
			}
		}
	})

	t.Run("filters by closed state", func(t *testing.T) {
		issues, err := ListIssuesByState(db, "owner/repo", "closed")
		if err != nil {
			t.Fatalf("ListIssuesByState failed: %v", err)
		}

		if len(issues) != 1 {
			t.Errorf("expected 1 closed issue, got %d", len(issues))
		}

		if issues[0].Number != 2 {
			t.Errorf("expected closed issue #2, got #%d", issues[0].Number)
		}
	})
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "formats RFC3339 date",
			input:    "2024-01-15T10:30:00Z",
			expected: "2024-01-15",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles invalid date",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetIssueDetail(t *testing.T) {
	// Create a temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	defer db.Close()

	testIssue := Issue{
		Number:       42,
		Title:        "Test Issue with Full Details",
		Body:         "This is the body of the issue with **markdown** support.",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:00:00Z",
		UpdatedAt:    "2024-01-16T14:00:00Z",
		ClosedAt:     "",
		CommentCount: 3,
		Labels:       []string{"bug", "enhancement"},
		Assignees:    []string{"user1", "user2"},
	}

	if err := SaveIssue(db, "owner/repo", testIssue); err != nil {
		t.Fatalf("Failed to save issue: %v", err)
	}

	t.Run("returns issue with all fields", func(t *testing.T) {
		issue, err := GetIssueDetail(db, "owner/repo", 42)
		if err != nil {
			t.Fatalf("GetIssueDetail failed: %v", err)
		}

		if issue.Number != 42 {
			t.Errorf("expected number 42, got %d", issue.Number)
		}
		if issue.Title != "Test Issue with Full Details" {
			t.Errorf("expected title 'Test Issue with Full Details', got '%s'", issue.Title)
		}
		if issue.Body != "This is the body of the issue with **markdown** support." {
			t.Errorf("expected body 'This is the body of the issue with **markdown** support.', got '%s'", issue.Body)
		}
		if issue.State != "open" {
			t.Errorf("expected state 'open', got '%s'", issue.State)
		}
		if issue.Author != "testuser" {
			t.Errorf("expected author 'testuser', got '%s'", issue.Author)
		}
		if issue.CreatedAt != "2024-01-15T10:00:00Z" {
			t.Errorf("expected created_at '2024-01-15T10:00:00Z', got '%s'", issue.CreatedAt)
		}
		if issue.UpdatedAt != "2024-01-16T14:00:00Z" {
			t.Errorf("expected updated_at '2024-01-16T14:00:00Z', got '%s'", issue.UpdatedAt)
		}
		if issue.ClosedAt != "" {
			t.Errorf("expected closed_at to be empty, got '%s'", issue.ClosedAt)
		}
		if issue.CommentCount != 3 {
			t.Errorf("expected comment count 3, got %d", issue.CommentCount)
		}
	})

	t.Run("includes labels", func(t *testing.T) {
		issue, err := GetIssueDetail(db, "owner/repo", 42)
		if err != nil {
			t.Fatalf("GetIssueDetail failed: %v", err)
		}

		if len(issue.Labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(issue.Labels))
		}
		if issue.Labels[0] != "bug" || issue.Labels[1] != "enhancement" {
			t.Errorf("expected labels ['bug', 'enhancement'], got %v", issue.Labels)
		}
	})

	t.Run("includes assignees", func(t *testing.T) {
		issue, err := GetIssueDetail(db, "owner/repo", 42)
		if err != nil {
			t.Fatalf("GetIssueDetail failed: %v", err)
		}

		if len(issue.Assignees) != 2 {
			t.Errorf("expected 2 assignees, got %d", len(issue.Assignees))
		}
		if issue.Assignees[0] != "user1" || issue.Assignees[1] != "user2" {
			t.Errorf("expected assignees ['user1', 'user2'], got %v", issue.Assignees)
		}
	})

	t.Run("returns error for non-existent issue", func(t *testing.T) {
		_, err := GetIssueDetail(db, "owner/repo", 999)
		if err == nil {
			t.Error("expected error for non-existent issue, got nil")
		}
	})

	t.Run("returns error for non-existent repo", func(t *testing.T) {
		_, err := GetIssueDetail(db, "other/repo", 42)
		if err == nil {
			t.Error("expected error for non-existent repo, got nil")
		}
	})

	t.Run("returns closed_at for closed issues", func(t *testing.T) {
		closedIssue := Issue{
			Number:    101,
			Title:     "Closed Issue",
			State:     "closed",
			Author:    "testuser",
			CreatedAt: "2024-01-15T10:00:00Z",
			UpdatedAt: "2024-01-17T10:00:00Z",
			ClosedAt:  "2024-01-17T09:00:00Z",
		}
		if err := SaveIssue(db, "owner/repo", closedIssue); err != nil {
			t.Fatalf("Failed to save closed issue: %v", err)
		}

		issue, err := GetIssueDetail(db, "owner/repo", 101)
		if err != nil {
			t.Fatalf("GetIssueDetail failed: %v", err)
		}

		if issue.ClosedAt != "2024-01-17T09:00:00Z" {
			t.Errorf("expected closed_at '2024-01-17T09:00:00Z', got '%s'", issue.ClosedAt)
		}
	})
}

func TestListIssue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		issue   ListIssue
		wantErr bool
	}{
		{
			name:    "valid issue",
			issue:   ListIssue{Number: 1, Title: "Test", Author: "user"},
			wantErr: false,
		},
		{
			name:    "zero number is invalid",
			issue:   ListIssue{Number: 0, Title: "Test", Author: "user"},
			wantErr: true,
		},
		{
			name:    "empty title is invalid",
			issue:   ListIssue{Number: 1, Title: "", Author: "user"},
			wantErr: true,
		},
		{
			name:    "empty author is invalid",
			issue:   ListIssue{Number: 1, Title: "Test", Author: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.issue.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
