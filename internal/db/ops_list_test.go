package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/github"
)

func TestListIssues(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	owner := "test-owner"
	repo := "test-repo"

	// Insert some test issues
	testIssues := []github.Issue{
		{
			Number: 1,
			Title:  "First Issue",
			State:  "open",
			Author: github.User{Login: "user1"},
			Comments: 5,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			Number: 2,
			Title:  "Second Issue",
			State:  "open",
			Author: github.User{Login: "user2"},
			Comments: 0,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			Number: 3,
			Title:  "Third Issue",
			State:  "closed",
			Author: github.User{Login: "user1"},
			Comments: 10,
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for _, issue := range testIssues {
		err := UpsertIssue(db, owner, repo, &issue)
		if err != nil {
			t.Fatalf("Failed to upsert issue: %v", err)
		}
	}

	// List issues
	issues, err := ListIssues(db, owner, repo)
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Issues should be ordered by updated_at descending (most recent first)
	if issues[0].Number != 1 {
		t.Errorf("Expected first issue to be #1 (most recently updated), got #%d", issues[0].Number)
	}

	// Verify issue data
	for _, issue := range issues {
		if issue.Number == 1 {
			if issue.Title != "First Issue" {
				t.Errorf("Expected title 'First Issue', got '%s'", issue.Title)
			}
			if issue.Author != "user1" {
				t.Errorf("Expected author 'user1', got '%s'", issue.Author)
			}
			if issue.CommentCnt != 5 {
				t.Errorf("Expected 5 comments, got %d", issue.CommentCnt)
			}
		}
	}
}

func TestListIssuesEmpty(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// List issues for a repo with no issues
	issues, err := ListIssues(db, "empty", "repo")
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestGetIssue(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	owner := "test-owner"
	repo := "test-repo"

	// Insert a test issue
	issue := &github.Issue{
		Number:    42,
		Title:     "Test Issue",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		Comments:  3,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = UpsertIssue(db, owner, repo, issue)
	if err != nil {
		t.Fatalf("Failed to upsert issue: %v", err)
	}

	// Get the issue
	result, err := GetIssue(db, owner, repo, 42)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}

	if result == nil {
		t.Fatal("Expected issue, got nil")
	}

	if result.Number != 42 {
		t.Errorf("Expected issue number 42, got %d", result.Number)
	}
	if result.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", result.Title)
	}
	if result.Author != "testuser" {
		t.Errorf("Expected author 'testuser', got '%s'", result.Author)
	}
}

func TestGetIssueNotFound(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get a non-existent issue
	result, err := GetIssue(db, "owner", "repo", 999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil for non-existent issue, got %+v", result)
	}
}

func TestIssueListStruct(t *testing.T) {
	issue := IssueList{
		Number:     123,
		Title:      "Test Title",
		Author:     "test-author",
		CreatedAt:  "2024-01-01T00:00:00Z",
		CommentCnt: 5,
		State:      "open",
	}

	if issue.Number != 123 {
		t.Errorf("Expected Number 123, got %d", issue.Number)
	}
	if issue.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got '%s'", issue.Title)
	}
	if issue.Author != "test-author" {
		t.Errorf("Expected Author 'test-author', got '%s'", issue.Author)
	}
	if issue.CommentCnt != 5 {
		t.Errorf("Expected CommentCnt 5, got %d", issue.CommentCnt)
	}
	if issue.State != "open" {
		t.Errorf("Expected State 'open', got '%s'", issue.State)
	}
}