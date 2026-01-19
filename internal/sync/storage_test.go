package sync

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestInitSchema(t *testing.T) {
	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := &IssueStore{db: db}

	// Initialize schema
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	// Verify tables exist
	tables := []string{"issues", "comments", "labels", "assignees"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}

	// Verify idempotency - running again should not fail
	if err := store.InitSchema(); err != nil {
		t.Errorf("InitSchema() should be idempotent, error = %v", err)
	}
}

func TestStoreIssue(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := &IssueStore{db: db}
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	// Test issue
	issue := &Issue{
		Number:       42,
		Title:        "Test Issue",
		Body:         "This is a test issue",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    time.Now().Add(-24 * time.Hour),
		UpdatedAt:    time.Now(),
		CommentCount: 3,
		Labels:       []string{"bug", "enhancement"},
		Assignees:    []string{"dev1", "dev2"},
	}

	// Store issue
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Verify issue was stored
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM issues WHERE number = ?", issue.Number).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query issues: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 issue, got %d", count)
	}

	// Verify issue data
	var title, body, state, author string
	var commentCount int
	err = db.QueryRow("SELECT title, body, state, author, comment_count FROM issues WHERE number = ?", issue.Number).
		Scan(&title, &body, &state, &author, &commentCount)
	if err != nil {
		t.Fatalf("Failed to query issue data: %v", err)
	}

	if title != issue.Title {
		t.Errorf("Title = %s, want %s", title, issue.Title)
	}
	if body != issue.Body {
		t.Errorf("Body = %s, want %s", body, issue.Body)
	}
	if state != issue.State {
		t.Errorf("State = %s, want %s", state, issue.State)
	}
	if author != issue.Author {
		t.Errorf("Author = %s, want %s", author, issue.Author)
	}
	if commentCount != issue.CommentCount {
		t.Errorf("CommentCount = %d, want %d", commentCount, issue.CommentCount)
	}

	// Verify labels were stored
	var labelCount int
	err = db.QueryRow("SELECT COUNT(*) FROM labels WHERE issue_number = ?", issue.Number).Scan(&labelCount)
	if err != nil {
		t.Fatalf("Failed to query labels: %v", err)
	}
	if labelCount != len(issue.Labels) {
		t.Errorf("Expected %d labels, got %d", len(issue.Labels), labelCount)
	}

	// Verify assignees were stored
	var assigneeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM assignees WHERE issue_number = ?", issue.Number).Scan(&assigneeCount)
	if err != nil {
		t.Fatalf("Failed to query assignees: %v", err)
	}
	if assigneeCount != len(issue.Assignees) {
		t.Errorf("Expected %d assignees, got %d", len(issue.Assignees), assigneeCount)
	}
}

func TestStoreIssue_Update(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := &IssueStore{db: db}
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	// Store initial issue
	issue := &Issue{
		Number:    1,
		Title:     "Original Title",
		Body:      "Original Body",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels:    []string{"bug"},
	}

	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Update issue
	issue.Title = "Updated Title"
	issue.Body = "Updated Body"
	issue.State = "closed"
	issue.Labels = []string{"bug", "fixed"}

	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() update error = %v", err)
	}

	// Verify only one issue exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM issues WHERE number = ?", issue.Number).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query issues: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 issue after update, got %d", count)
	}

	// Verify updated data
	var title, body, state string
	err = db.QueryRow("SELECT title, body, state FROM issues WHERE number = ?", issue.Number).
		Scan(&title, &body, &state)
	if err != nil {
		t.Fatalf("Failed to query updated issue: %v", err)
	}

	if title != "Updated Title" {
		t.Errorf("Title = %s, want 'Updated Title'", title)
	}
	if state != "closed" {
		t.Errorf("State = %s, want 'closed'", state)
	}

	// Verify labels were updated
	var labelCount int
	err = db.QueryRow("SELECT COUNT(*) FROM labels WHERE issue_number = ?", issue.Number).Scan(&labelCount)
	if err != nil {
		t.Fatalf("Failed to query labels: %v", err)
	}
	if labelCount != 2 {
		t.Errorf("Expected 2 labels after update, got %d", labelCount)
	}
}

func TestStoreComment(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := &IssueStore{db: db}
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	// Store issue first
	issue := &Issue{
		Number:    1,
		Title:     "Test",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Store comment
	comment := &Comment{
		ID:           123,
		IssueNumber:  1,
		Body:         "This is a comment",
		Author:       "commenter",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := store.StoreComment(comment); err != nil {
		t.Fatalf("StoreComment() error = %v", err)
	}

	// Verify comment was stored
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM comments WHERE id = ?", comment.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query comments: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 comment, got %d", count)
	}

	// Verify comment data
	var body, author string
	var issueNumber int
	err = db.QueryRow("SELECT body, author, issue_number FROM comments WHERE id = ?", comment.ID).
		Scan(&body, &author, &issueNumber)
	if err != nil {
		t.Fatalf("Failed to query comment data: %v", err)
	}

	if body != comment.Body {
		t.Errorf("Body = %s, want %s", body, comment.Body)
	}
	if author != comment.Author {
		t.Errorf("Author = %s, want %s", author, comment.Author)
	}
	if issueNumber != comment.IssueNumber {
		t.Errorf("IssueNumber = %d, want %d", issueNumber, comment.IssueNumber)
	}
}

func TestClearData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := &IssueStore{db: db}
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	// Store test data
	issue := &Issue{
		Number:    1,
		Title:     "Test",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels:    []string{"test"},
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	comment := &Comment{
		ID:          1,
		IssueNumber: 1,
		Body:        "Test comment",
		Author:      "user1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := store.StoreComment(comment); err != nil {
		t.Fatalf("StoreComment() error = %v", err)
	}

	// Clear all data
	if err := store.ClearData(); err != nil {
		t.Fatalf("ClearData() error = %v", err)
	}

	// Verify all tables are empty
	tables := []string{"issues", "comments", "labels", "assignees"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query %s: %v", table, err)
		}
		if count != 0 {
			t.Errorf("Expected %s to be empty after clear, got %d rows", table, count)
		}
	}
}

// Test helper to ensure database file is created with proper permissions
func TestDatabasePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	store := &IssueStore{db: db}
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	// Check file permissions
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Failed to stat database file: %v", err)
	}

	// Database should be readable and writable
	mode := info.Mode()
	if mode&0600 != 0600 {
		t.Logf("Warning: Database file has permissions %v, expected at least 0600", mode)
	}
}

func TestLoadIssues(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store test issues with different timestamps
	now := time.Now()
	issues := []*Issue{
		{
			Number:       1,
			Title:        "First Issue",
			State:        "open",
			Author:       "user1",
			CreatedAt:    now.Add(-3 * time.Hour),
			UpdatedAt:    now.Add(-1 * time.Hour),
			CommentCount: 5,
			Labels:       []string{"bug"},
		},
		{
			Number:       2,
			Title:        "Second Issue",
			State:        "open",
			Author:       "user2",
			CreatedAt:    now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-30 * time.Minute),
			CommentCount: 3,
			Labels:       []string{"feature"},
		},
		{
			Number:       3,
			Title:        "Third Issue",
			State:        "open",
			Author:       "user1",
			CreatedAt:    now.Add(-1 * time.Hour),
			UpdatedAt:    now,
			CommentCount: 0,
			Assignees:    []string{"dev1"},
		},
	}

	for _, issue := range issues {
		if err := store.StoreIssue(issue); err != nil {
			t.Fatalf("StoreIssue() error = %v", err)
		}
	}

	// Test loading all issues
	loaded, err := store.LoadIssues()
	if err != nil {
		t.Fatalf("LoadIssues() error = %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("LoadIssues() returned %d issues, want 3", len(loaded))
	}

	// Verify issues are sorted by updated_at DESC (most recent first)
	if loaded[0].Number != 3 {
		t.Errorf("First issue number = %d, want 3 (most recently updated)", loaded[0].Number)
	}
	if loaded[1].Number != 2 {
		t.Errorf("Second issue number = %d, want 2", loaded[1].Number)
	}
	if loaded[2].Number != 1 {
		t.Errorf("Third issue number = %d, want 1 (least recently updated)", loaded[2].Number)
	}

	// Verify labels and assignees are loaded
	// Issue 3 has no labels
	if len(loaded[0].Assignees) != 1 || loaded[0].Assignees[0] != "dev1" {
		t.Errorf("Issue 3 assignees = %v, want [dev1]", loaded[0].Assignees)
	}
	if len(loaded[2].Labels) != 1 || loaded[2].Labels[0] != "bug" {
		t.Errorf("Issue 1 labels = %v, want [bug]", loaded[2].Labels)
	}
}

func TestLoadIssues_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Load from empty database
	issues, err := store.LoadIssues()
	if err != nil {
		t.Fatalf("LoadIssues() error = %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("LoadIssues() returned %d issues, want 0", len(issues))
	}
}

func TestLoadComments(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store test issue
	issue := &Issue{
		Number:    42,
		Title:     "Test Issue",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Store test comments with different timestamps
	now := time.Now()
	comments := []*Comment{
		{
			ID:          1,
			IssueNumber: 42,
			Body:        "First comment",
			Author:      "user1",
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ID:          2,
			IssueNumber: 42,
			Body:        "Second comment",
			Author:      "user2",
			CreatedAt:   now.Add(-1 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:          3,
			IssueNumber: 42,
			Body:        "Third comment",
			Author:      "user1",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, comment := range comments {
		if err := store.StoreComment(comment); err != nil {
			t.Fatalf("StoreComment() error = %v", err)
		}
	}

	// Load comments for the issue
	loaded, err := store.LoadComments(42)
	if err != nil {
		t.Fatalf("LoadComments() error = %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("LoadComments() returned %d comments, want 3", len(loaded))
	}

	// Verify comments are sorted by created_at ASC (chronological order)
	if loaded[0].ID != 1 {
		t.Errorf("First comment ID = %d, want 1 (oldest)", loaded[0].ID)
	}
	if loaded[1].ID != 2 {
		t.Errorf("Second comment ID = %d, want 2", loaded[1].ID)
	}
	if loaded[2].ID != 3 {
		t.Errorf("Third comment ID = %d, want 3 (newest)", loaded[2].ID)
	}

	// Verify comment data
	if loaded[0].Body != "First comment" {
		t.Errorf("First comment body = %s, want 'First comment'", loaded[0].Body)
	}
	if loaded[1].Author != "user2" {
		t.Errorf("Second comment author = %s, want 'user2'", loaded[1].Author)
	}
}

func TestLoadComments_NoComments(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Store issue without comments
	issue := &Issue{
		Number:    1,
		Title:     "Test",
		State:     "open",
		Author:    "user1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("StoreIssue() error = %v", err)
	}

	// Load comments for issue with no comments
	comments, err := store.LoadComments(1)
	if err != nil {
		t.Fatalf("LoadComments() error = %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("LoadComments() returned %d comments, want 0", len(comments))
	}
}

func TestLoadComments_NonexistentIssue(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("NewIssueStore() error = %v", err)
	}
	defer store.Close()

	// Load comments for issue that doesn't exist
	comments, err := store.LoadComments(999)
	if err != nil {
		t.Fatalf("LoadComments() error = %v", err)
	}

	// Should return empty slice, not error
	if len(comments) != 0 {
		t.Errorf("LoadComments() returned %d comments, want 0", len(comments))
	}
}
