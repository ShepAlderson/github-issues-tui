package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/github"
)

func TestOpen_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Verify we can query the database
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count); err != nil {
		t.Fatalf("failed to query issues table: %v", err)
	}
}

func TestOpen_CreatesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Check that all tables exist
	tables := []string{"issues", "labels", "assignees", "comments"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		if err := db.QueryRow(query, table).Scan(&count); err != nil {
			t.Fatalf("failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("table %s not found", table)
		}
	}
}

func TestUpsertIssue_Insert(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "This is a test issue body",
		State:     "open",
		Author:    github.User{Login: "testuser", ID: 12345},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Comments:  5,
		HTMLURL:   "https://github.com/owner/repo/issues/1",
	}

	if err := UpsertIssue(db, "owner", "repo", issue); err != nil {
		t.Fatalf("UpsertIssue() error = %v", err)
	}

	// Verify the issue was inserted
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM issues WHERE number = 1").Scan(&count); err != nil {
		t.Fatalf("failed to count issues: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 issue, got %d", count)
	}
}

func TestUpsertIssue_Update(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	issue := &github.Issue{
		Number:    1,
		Title:     "Original Title",
		Body:      "Original body",
		State:     "open",
		Author:    github.User{Login: "testuser", ID: 12345},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Comments:  0,
		HTMLURL:   "https://github.com/owner/repo/issues/1",
	}

	// Insert first
	if err := UpsertIssue(db, "owner", "repo", issue); err != nil {
		t.Fatalf("UpsertIssue() error = %v", err)
	}

	// Update with new title
	issue.Title = "Updated Title"
	if err := UpsertIssue(db, "owner", "repo", issue); err != nil {
		t.Fatalf("UpsertIssue() error = %v", err)
	}

	// Verify only one issue exists and it has the updated title
	var count int
	var title string
	if err := db.QueryRow("SELECT COUNT(*), title FROM issues WHERE number = 1").Scan(&count, &title); err != nil {
		t.Fatalf("failed to query issues: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 issue, got %d", count)
	}
	if title != "Updated Title" {
		t.Errorf("expected 'Updated Title', got %q", title)
	}
}

func TestInsertLabel(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert the parent issue first
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	label := &github.Label{
		ID:    12345,
		Name:  "bug",
		Color: "ff0000",
	}

	if err := InsertLabel(db, 1, label); err != nil {
		t.Fatalf("InsertLabel() error = %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM labels WHERE issue_number = 1 AND name = 'bug'").Scan(&count); err != nil {
		t.Fatalf("failed to count labels: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 label, got %d", count)
	}
}

func TestDeleteLabels(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert the parent issue first
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	// Insert some labels
	label := &github.Label{Name: "bug", Color: "ff0000"}
	if err := InsertLabel(db, 1, label); err != nil {
		t.Fatalf("InsertLabel() error = %v", err)
	}
	if err := InsertLabel(db, 1, &github.Label{Name: "priority", Color: "00ff00"}); err != nil {
		t.Fatalf("InsertLabel() error = %v", err)
	}

	// Delete them
	if err := DeleteLabels(db, 1); err != nil {
		t.Fatalf("DeleteLabels() error = %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM labels WHERE issue_number = 1").Scan(&count); err != nil {
		t.Fatalf("failed to count labels: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 labels, got %d", count)
	}
}

func TestInsertAssignee(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert the parent issue first
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	assignee := &github.User{Login: "assignee1", ID: 12345}

	if err := InsertAssignee(db, 1, assignee); err != nil {
		t.Fatalf("InsertAssignee() error = %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM assignees WHERE issue_number = 1 AND login = 'assignee1'").Scan(&count); err != nil {
		t.Fatalf("failed to count assignees: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 assignee, got %d", count)
	}
}

func TestUpsertComment(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert the parent issue first
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	comment := &github.Comment{
		ID:        12345,
		Body:      "This is a test comment",
		Author:    github.User{Login: "commenter", ID: 67890},
		CreatedAt: time.Now(),
	}

	if err := UpsertComment(db, 1, comment); err != nil {
		t.Fatalf("UpsertComment() error = %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE id = 12345").Scan(&count); err != nil {
		t.Fatalf("failed to count comments: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 comment, got %d", count)
	}
}

func TestIssueExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Should not exist yet
	exists, err := IssueExists(db, "owner", "repo", 1)
	if err != nil {
		t.Fatalf("IssueExists() error = %v", err)
	}
	if exists {
		t.Error("IssueExists() = true, want false")
	}

	// Insert the issue
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	// Should exist now
	exists, err = IssueExists(db, "owner", "repo", 1)
	if err != nil {
		t.Fatalf("IssueExists() error = %v", err)
	}
	if !exists {
		t.Error("IssueExists() = false, want true")
	}
}

func TestGetIssueCount(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Should be 0
	count, err := GetIssueCount(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetIssueCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 issues, got %d", count)
	}

	// Insert some issues
	for i := 1; i <= 5; i++ {
		issue := &github.Issue{
			Number:    i,
			Title:     "Test Issue",
			Body:      "",
			State:     "open",
			Author:    github.User{Login: "testuser"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		UpsertIssue(db, "owner", "repo", issue)
	}

	// Should be 5
	count, err = GetIssueCount(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetIssueCount() error = %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 issues, got %d", count)
	}
}

func TestIssueLabelsCascade(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert issue and labels
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)
	InsertLabel(db, 1, &github.Label{Name: "bug"})
	InsertLabel(db, 1, &github.Label{Name: "priority"})

	// Delete the issue (labels should cascade delete)
	_, err = db.Exec("DELETE FROM issues WHERE number = 1")
	if err != nil {
		t.Fatalf("failed to delete issue: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM labels WHERE issue_number = 1").Scan(&count); err != nil {
		t.Fatalf("failed to count labels: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 labels after cascade delete, got %d", count)
	}
}

func TestOpen_FileCreated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// File should not exist yet
	if _, err := os.Stat(dbPath); err == nil {
		t.Error("database file already exists before Open()")
	}

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// File should exist now
	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file not created: %v", err)
	}
}

func TestGetLastSyncTime_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	lastSync, err := GetLastSyncTime(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetLastSyncTime() error = %v", err)
	}
	if lastSync != "" {
		t.Errorf("expected empty string for no issues, got %q", lastSync)
	}
}

func TestGetLastSyncTime_WithIssues(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert an issue
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	lastSync, err := GetLastSyncTime(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetLastSyncTime() error = %v", err)
	}
	if lastSync == "" {
		t.Error("expected non-empty sync time after inserting issue")
	}
}

func TestGetIssueNumbers_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	numbers, err := GetIssueNumbers(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetIssueNumbers() error = %v", err)
	}
	if len(numbers) != 0 {
		t.Errorf("expected 0 numbers, got %d", len(numbers))
	}
}

func TestGetIssueNumbers_WithIssues(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert some issues
	for i := 1; i <= 5; i++ {
		issue := &github.Issue{
			Number:    i,
			Title:     fmt.Sprintf("Test Issue %d", i),
			Body:      "",
			State:     "open",
			Author:    github.User{Login: "testuser"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		UpsertIssue(db, "owner", "repo", issue)
	}

	numbers, err := GetIssueNumbers(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetIssueNumbers() error = %v", err)
	}
	if len(numbers) != 5 {
		t.Errorf("expected 5 numbers, got %d", len(numbers))
	}
	// Check they're the expected values
	for i, num := range numbers {
		if num != i+1 {
			t.Errorf("expected %d, got %d", i+1, num)
		}
	}
}

func TestDeleteIssue(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert an issue
	issue := &github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "",
		State:     "open",
		Author:    github.User{Login: "testuser"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	UpsertIssue(db, "owner", "repo", issue)

	// Delete it
	if err := DeleteIssue(db, "owner", "repo", 1); err != nil {
		t.Fatalf("DeleteIssue() error = %v", err)
	}

	// Verify it's gone
	count, err := GetIssueCount(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetIssueCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 issues after delete, got %d", count)
	}
}

func TestDeleteIssues(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Insert some issues
	for i := 1; i <= 5; i++ {
		issue := &github.Issue{
			Number:    i,
			Title:     fmt.Sprintf("Test Issue %d", i),
			Body:      "",
			State:     "open",
			Author:    github.User{Login: "testuser"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		UpsertIssue(db, "owner", "repo", issue)
	}

	// Delete some issues
	if err := DeleteIssues(db, "owner", "repo", []int{1, 3, 5}); err != nil {
		t.Fatalf("DeleteIssues() error = %v", err)
	}

	// Verify remaining issues
	numbers, err := GetIssueNumbers(db, "owner", "repo")
	if err != nil {
		t.Fatalf("GetIssueNumbers() error = %v", err)
	}
	if len(numbers) != 2 {
		t.Errorf("expected 2 remaining issues, got %d", len(numbers))
	}
}

func TestDeleteIssues_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	// Delete with empty list should not error
	if err := DeleteIssues(db, "owner", "repo", []int{}); err != nil {
		t.Errorf("DeleteIssues() with empty list should not error: %v", err)
	}
}