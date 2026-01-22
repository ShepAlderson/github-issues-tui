package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/config"
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

func TestListIssuesSortedByNumber(t *testing.T) {
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
			Number:    3,
			Title:     "Third Issue",
			State:     "open",
			Author:    github.User{Login: "user1"},
			Comments:  10,
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			Number:    1,
			Title:     "First Issue",
			State:     "open",
			Author:    github.User{Login: "user1"},
			Comments:  5,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			Number:    2,
			Title:     "Second Issue",
			State:     "open",
			Author:    github.User{Login: "user2"},
			Comments:  0,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, issue := range testIssues {
		err := UpsertIssue(db, owner, repo, &issue)
		if err != nil {
			t.Fatalf("Failed to upsert issue: %v", err)
		}
	}

	// List issues sorted by number ascending
	issues, err := ListIssuesSorted(db, owner, repo, config.SortNumber, config.SortOrderAsc)
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Issues should be ordered by number ascending: 1, 2, 3
	if issues[0].Number != 1 {
		t.Errorf("Expected first issue to be #1, got #%d", issues[0].Number)
	}
	if issues[1].Number != 2 {
		t.Errorf("Expected second issue to be #2, got #%d", issues[1].Number)
	}
	if issues[2].Number != 3 {
		t.Errorf("Expected third issue to be #3, got #%d", issues[2].Number)
	}
}

func TestListIssuesSortedByNumberDesc(t *testing.T) {
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
			Number:    3,
			Title:     "Third Issue",
			State:     "open",
			Author:    github.User{Login: "user1"},
			Comments:  10,
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			Number:    1,
			Title:     "First Issue",
			State:     "open",
			Author:    github.User{Login: "user1"},
			Comments:  5,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			Number:    2,
			Title:     "Second Issue",
			State:     "open",
			Author:    github.User{Login: "user2"},
			Comments:  0,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, issue := range testIssues {
		err := UpsertIssue(db, owner, repo, &issue)
		if err != nil {
			t.Fatalf("Failed to upsert issue: %v", err)
		}
	}

	// List issues sorted by number descending
	issues, err := ListIssuesSorted(db, owner, repo, config.SortNumber, config.SortOrderDesc)
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Issues should be ordered by number descending: 3, 2, 1
	if issues[0].Number != 3 {
		t.Errorf("Expected first issue to be #3, got #%d", issues[0].Number)
	}
	if issues[1].Number != 2 {
		t.Errorf("Expected second issue to be #2, got #%d", issues[1].Number)
	}
	if issues[2].Number != 1 {
		t.Errorf("Expected third issue to be #1, got #%d", issues[2].Number)
	}
}

func TestListIssuesSortedByComments(t *testing.T) {
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

	// Insert some test issues with different comment counts
	testIssues := []github.Issue{
		{
			Number:    1,
			Title:     "First Issue",
			State:     "open",
			Author:    github.User{Login: "user1"},
			Comments:  5,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		},
		{
			Number:    2,
			Title:     "Second Issue",
			State:     "open",
			Author:    github.User{Login: "user2"},
			Comments:  0,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			Number:    3,
			Title:     "Third Issue",
			State:     "closed",
			Author:    github.User{Login: "user1"},
			Comments:  10,
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

	// List issues sorted by comment count descending
	issues, err := ListIssuesSorted(db, owner, repo, config.SortComments, config.SortOrderDesc)
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Issues should be ordered by comment count descending: 10, 5, 0
	if issues[0].CommentCnt != 10 {
		t.Errorf("Expected first issue to have 10 comments, got %d", issues[0].CommentCnt)
	}
	if issues[1].CommentCnt != 5 {
		t.Errorf("Expected second issue to have 5 comments, got %d", issues[1].CommentCnt)
	}
	if issues[2].CommentCnt != 0 {
		t.Errorf("Expected third issue to have 0 comments, got %d", issues[2].CommentCnt)
	}
}

func TestListIssuesSortedByCreated(t *testing.T) {
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
			Number:    1,
			Title:     "First Issue",
			State:     "open",
			Author:    github.User{Login: "user1"},
			Comments:  5,
			CreatedAt: time.Now().Add(-72 * time.Hour), // Oldest
			UpdatedAt: time.Now(),
		},
		{
			Number:    2,
			Title:     "Second Issue",
			State:     "open",
			Author:    github.User{Login: "user2"},
			Comments:  0,
			CreatedAt: time.Now().Add(-24 * time.Hour), // Newest
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			Number:    3,
			Title:     "Third Issue",
			State:     "closed",
			Author:    github.User{Login: "user1"},
			Comments:  10,
			CreatedAt: time.Now().Add(-48 * time.Hour), // Middle
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for _, issue := range testIssues {
		err := UpsertIssue(db, owner, repo, &issue)
		if err != nil {
			t.Fatalf("Failed to upsert issue: %v", err)
		}
	}

	// List issues sorted by created_at descending (newest first)
	issues, err := ListIssuesSorted(db, owner, repo, config.SortCreated, config.SortOrderDesc)
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Issues should be ordered by created_at descending: 2 (newest), 3, 1 (oldest)
	if issues[0].Number != 2 {
		t.Errorf("Expected first issue to be #2 (newest), got #%d", issues[0].Number)
	}
	if issues[1].Number != 3 {
		t.Errorf("Expected second issue to be #3, got #%d", issues[1].Number)
	}
	if issues[2].Number != 1 {
		t.Errorf("Expected third issue to be #1 (oldest), got #%d", issues[2].Number)
	}
}

func TestListIssuesSortedInvalidOption(t *testing.T) {
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
		Number:    1,
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

	// List issues with an invalid sort option - should default to updated
	issues, err := ListIssuesSorted(db, owner, repo, "invalid", config.SortOrderDesc)
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
}

func TestGetIssueDetail(t *testing.T) {
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
		Body:      "This is the **body** of the issue",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		Comments:  3,
		HTMLURL:   "https://github.com/test-owner/test-repo/issues/42",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = UpsertIssue(db, owner, repo, issue)
	if err != nil {
		t.Fatalf("Failed to upsert issue: %v", err)
	}

	// Add labels
	label1 := &github.Label{Name: "bug", Color: "ff0000"}
	label2 := &github.Label{Name: "enhancement", Color: "00ff00"}
	err = InsertLabel(db, 42, label1)
	if err != nil {
		t.Fatalf("Failed to insert label: %v", err)
	}
	err = InsertLabel(db, 42, label2)
	if err != nil {
		t.Fatalf("Failed to insert label: %v", err)
	}

	// Add assignees
	assignee1 := &github.User{Login: "assignee1"}
	assignee2 := &github.User{Login: "assignee2"}
	err = InsertAssignee(db, 42, assignee1)
	if err != nil {
		t.Fatalf("Failed to insert assignee: %v", err)
	}
	err = InsertAssignee(db, 42, assignee2)
	if err != nil {
		t.Fatalf("Failed to insert assignee: %v", err)
	}

	// Get the issue detail
	detail, err := GetIssueDetail(db, owner, repo, 42)
	if err != nil {
		t.Fatalf("Failed to get issue detail: %v", err)
	}

	if detail == nil {
		t.Fatal("Expected issue detail, got nil")
	}

	if detail.Number != 42 {
		t.Errorf("Expected issue number 42, got %d", detail.Number)
	}
	if detail.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", detail.Title)
	}
	if detail.Body != "This is the **body** of the issue" {
		t.Errorf("Expected body mismatch, got '%s'", detail.Body)
	}
	if detail.State != "open" {
		t.Errorf("Expected state 'open', got '%s'", detail.State)
	}
	if detail.Author != "testuser" {
		t.Errorf("Expected author 'testuser', got '%s'", detail.Author)
	}
	if len(detail.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(detail.Labels))
	}
	if len(detail.Assignees) != 2 {
		t.Errorf("Expected 2 assignees, got %d", len(detail.Assignees))
	}
}

func TestGetIssueDetailNotFound(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get a non-existent issue detail
	detail, err := GetIssueDetail(db, "owner", "repo", 999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if detail != nil {
		t.Errorf("Expected nil for non-existent issue, got %+v", detail)
	}
}

func TestGetLabels(t *testing.T) {
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

	// First insert the issue (required for foreign key)
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = UpsertIssue(db, owner, repo, issue)
	if err != nil {
		t.Fatalf("Failed to upsert issue: %v", err)
	}

	// Insert labels
	label1 := &github.Label{Name: "bug", Color: "ff0000"}
	label2 := &github.Label{Name: "enhancement", Color: "00ff00"}
	label3 := &github.Label{Name: "help wanted", Color: "0000ff"}
	err = InsertLabel(db, 1, label1)
	if err != nil {
		t.Fatalf("Failed to insert label: %v", err)
	}
	err = InsertLabel(db, 1, label2)
	if err != nil {
		t.Fatalf("Failed to insert label: %v", err)
	}
	err = InsertLabel(db, 1, label3)
	if err != nil {
		t.Fatalf("Failed to insert label: %v", err)
	}

	// Get labels
	labels, err := GetLabels(db, 1)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}
}

func TestGetLabelsEmpty(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get labels for non-existent issue
	labels, err := GetLabels(db, 999)
	if err != nil {
		t.Fatalf("Failed to get labels: %v", err)
	}

	if len(labels) != 0 {
		t.Errorf("Expected 0 labels, got %d", len(labels))
	}
}

func TestGetAssignees(t *testing.T) {
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

	// First insert the issue (required for foreign key)
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = UpsertIssue(db, owner, repo, issue)
	if err != nil {
		t.Fatalf("Failed to upsert issue: %v", err)
	}

	// Insert assignees
	assignee1 := &github.User{Login: "user1"}
	assignee2 := &github.User{Login: "user2"}
	err = InsertAssignee(db, 1, assignee1)
	if err != nil {
		t.Fatalf("Failed to insert assignee: %v", err)
	}
	err = InsertAssignee(db, 1, assignee2)
	if err != nil {
		t.Fatalf("Failed to insert assignee: %v", err)
	}

	// Get assignees
	assignees, err := GetAssignees(db, 1)
	if err != nil {
		t.Fatalf("Failed to get assignees: %v", err)
	}

	if len(assignees) != 2 {
		t.Errorf("Expected 2 assignees, got %d", len(assignees))
	}
}

func TestGetAssigneesEmpty(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get assignees for non-existent issue
	assignees, err := GetAssignees(db, 999)
	if err != nil {
		t.Fatalf("Failed to get assignees: %v", err)
	}

	if len(assignees) != 0 {
		t.Errorf("Expected 0 assignees, got %d", len(assignees))
	}
}

func TestGetComments(t *testing.T) {
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

	// First insert the issue (required for foreign key)
	issue := &github.Issue{
		Number:    42,
		Title:     "Test Issue",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = UpsertIssue(db, owner, repo, issue)
	if err != nil {
		t.Fatalf("Failed to upsert issue: %v", err)
	}

	// Insert comments
	comment1 := &github.Comment{
		ID:        1,
		Body:      "First comment",
		Author:    github.User{Login: "commenter1"},
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	comment2 := &github.Comment{
		ID:        2,
		Body:      "Second comment",
		Author:    github.User{Login: "commenter2"},
		CreatedAt: time.Now(),
	}
	err = UpsertComment(db, 42, comment1)
	if err != nil {
		t.Fatalf("Failed to insert comment: %v", err)
	}
	err = UpsertComment(db, 42, comment2)
	if err != nil {
		t.Fatalf("Failed to insert comment: %v", err)
	}

	// Get comments
	comments, err := GetComments(db, 42)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}

	// Comments should be ordered by created_at ascending
	if comments[0].ID != 1 {
		t.Errorf("Expected first comment to be ID 1, got %d", comments[0].ID)
	}
	if comments[1].ID != 2 {
		t.Errorf("Expected second comment to be ID 2, got %d", comments[1].ID)
	}
}

func TestGetCommentsEmpty(t *testing.T) {
	// Create a temp database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Get comments for issue with no comments
	comments, err := GetComments(db, 999)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(comments))
	}
}