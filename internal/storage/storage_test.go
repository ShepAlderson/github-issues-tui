package storage

import (
	"os"
	"testing"
	"time"
)

func TestInitializeDatabase(t *testing.T) {
	// Create a temporary database file
	dbPath := t.TempDir() + "/test.db"

	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Verify tables exist
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='issues'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Issues table not created: %v", err)
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='comments'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Comments table not created: %v", err)
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='metadata'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Metadata table not created: %v", err)
	}
}

func TestStoreIssue(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Create a test issue
	issue := Issue{
		Number:      123,
		Title:       "Test Issue",
		Body:        "This is a test issue",
		Author:      "testuser",
		State:       "open",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		ClosedAt:    nil,
		Comments:    5,
		Labels:      "bug,enhancement",
		Assignees:   "user1,user2",
	}

	// Store the issue
	err = StoreIssue(db, &issue)
	if err != nil {
		t.Fatalf("StoreIssue failed: %v", err)
	}

	// Verify the issue was stored
	var storedTitle string
	var storedComments int
	err = db.QueryRow("SELECT title, comments FROM issues WHERE number = ?", 123).Scan(&storedTitle, &storedComments)
	if err != nil {
		t.Fatalf("Failed to retrieve stored issue: %v", err)
	}

	if storedTitle != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", storedTitle)
	}

	if storedComments != 5 {
		t.Errorf("Expected 5 comments, got %d", storedComments)
	}
}

func TestStoreIssueUpdate(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Create and store an initial issue
	issue := Issue{
		Number:    123,
		Title:     "Test Issue",
		Body:      "This is a test issue",
		Author:    "testuser",
		State:     "open",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
		Comments:  5,
	}

	err = StoreIssue(db, &issue)
	if err != nil {
		t.Fatalf("StoreIssue failed: %v", err)
	}

	// Update the issue
	issue.Title = "Updated Test Issue"
	issue.Comments = 10
	err = StoreIssue(db, &issue)
	if err != nil {
		t.Fatalf("StoreIssue (update) failed: %v", err)
	}

	// Verify the issue was updated
	var storedTitle string
	var storedComments int
	err = db.QueryRow("SELECT title, comments FROM issues WHERE number = ?", 123).Scan(&storedTitle, &storedComments)
	if err != nil {
		t.Fatalf("Failed to retrieve stored issue: %v", err)
	}

	if storedTitle != "Updated Test Issue" {
		t.Errorf("Expected title 'Updated Test Issue', got '%s'", storedTitle)
	}

	if storedComments != 10 {
		t.Errorf("Expected 10 comments, got %d", storedComments)
	}
}

func TestStoreComment(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Store an issue first
	issue := Issue{
		Number:    123,
		Title:     "Test Issue",
		Author:    "testuser",
		State:     "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = StoreIssue(db, &issue)
	if err != nil {
		t.Fatalf("StoreIssue failed: %v", err)
	}

	// Create a test comment
	comment := Comment{
		ID:        456,
		IssueNumber: 123,
		Body:      "This is a test comment",
		Author:    "commenter",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store the comment
	err = StoreComment(db, &comment)
	if err != nil {
		t.Fatalf("StoreComment failed: %v", err)
	}

	// Verify the comment was stored
	var storedBody string
	err = db.QueryRow("SELECT body FROM comments WHERE id = ?", 456).Scan(&storedBody)
	if err != nil {
		t.Fatalf("Failed to retrieve stored comment: %v", err)
	}

	if storedBody != "This is a test comment" {
		t.Errorf("Expected body 'This is a test comment', got '%s'", storedBody)
	}
}

func TestGetIssues(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Store multiple issues
	issues := []Issue{
		{
			Number:    1,
			Title:     "First Issue",
			Author:    "user1",
			State:     "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Number:    2,
			Title:     "Second Issue",
			Author:    "user2",
			State:     "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, issue := range issues {
		err = StoreIssue(db, &issue)
		if err != nil {
			t.Fatalf("StoreIssue failed: %v", err)
		}
	}

	// Retrieve all issues
	retrieved, err := GetIssues(db)
	if err != nil {
		t.Fatalf("GetIssues failed: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(retrieved))
	}
}

func TestGetCommentsForIssue(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Store an issue
	issue := Issue{
		Number:    123,
		Title:     "Test Issue",
		Author:    "testuser",
		State:     "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = StoreIssue(db, &issue)
	if err != nil {
		t.Fatalf("StoreIssue failed: %v", err)
	}

	// Store comments for the issue
	comments := []Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "First comment",
			Author:      "user1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			IssueNumber: 123,
			Body:        "Second comment",
			Author:      "user2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, comment := range comments {
		err = StoreComment(db, &comment)
		if err != nil {
			t.Fatalf("StoreComment failed: %v", err)
		}
	}

	// Retrieve comments for the issue
	retrieved, err := GetCommentsForIssue(db, 123)
	if err != nil {
		t.Fatalf("GetCommentsForIssue failed: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(retrieved))
	}
}

func TestUpdateLastSync(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Update last sync time
	now := time.Now()
	err = UpdateLastSync(db, now)
	if err != nil {
		t.Fatalf("UpdateLastSync failed: %v", err)
	}

	// Retrieve the last sync time
	var storedTimeStr string
	err = db.QueryRow("SELECT value FROM metadata WHERE key = 'last_sync'").Scan(&storedTimeStr)
	if err != nil {
		t.Fatalf("Failed to retrieve last sync time: %v", err)
	}

	storedTime, err := time.Parse(time.RFC3339, storedTimeStr)
	if err != nil {
		t.Fatalf("Failed to parse stored time: %v", err)
	}

	// Allow for some time difference due to rounding
	if storedTime.Sub(now).Seconds() > 1 {
		t.Errorf("Expected time close to %v, got %v", now, storedTime)
	}
}

func TestGetLastSync(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	db, err := InitializeDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitializeDatabase failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	// Test when no sync has occurred
	syncTime, err := GetLastSync(db)
	if err != nil {
		t.Fatalf("GetLastSync failed: %v", err)
	}
	if !syncTime.IsZero() {
		t.Errorf("Expected zero time for no sync, got %v", syncTime)
	}

	// Update last sync time
	now := time.Now()
	err = UpdateLastSync(db, now)
	if err != nil {
		t.Fatalf("UpdateLastSync failed: %v", err)
	}

	// Retrieve the last sync time
	syncTime, err = GetLastSync(db)
	if err != nil {
		t.Fatalf("GetLastSync failed: %v", err)
	}

	// Allow for some time difference due to rounding
	if syncTime.Sub(now).Seconds() > 1 {
		t.Errorf("Expected time close to %v, got %v", now, syncTime)
	}
}
