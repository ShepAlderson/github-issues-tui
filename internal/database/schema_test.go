package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitializeSchema(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("creates tables successfully", func(t *testing.T) {
		db, err := InitializeSchema(dbPath)
		if err != nil {
			t.Fatalf("InitializeSchema failed: %v", err)
		}
		defer db.Close()

		// Verify the database file was created
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("Database file was not created")
		}
	})
}

func TestSaveIssue(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	issue := Issue{
		Number:       123,
		Title:        "Test Issue",
		Body:         "This is a test issue body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-16T14:20:00Z",
		ClosedAt:     "",
		CommentCount: 5,
		Labels:       []string{"bug", "help wanted"},
		Assignees:    []string{"user1", "user2"},
	}

	t.Run("saves issue successfully", func(t *testing.T) {
		err := SaveIssue(db, "owner/repo", issue)
		if err != nil {
			t.Errorf("SaveIssue failed: %v", err)
		}
	})

	t.Run("updates existing issue", func(t *testing.T) {
		// Save the same issue again - should update
		issue.Title = "Updated Title"
		err := SaveIssue(db, "owner/repo", issue)
		if err != nil {
			t.Errorf("SaveIssue update failed: %v", err)
		}
	})
}

func TestSaveComment(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	// First save an issue to reference
	issue := Issue{
		Number: 456,
		Title:  "Test Issue with Comments",
		Author: "testuser",
		State:  "open",
	}
	err = SaveIssue(db, "owner/repo", issue)
	if err != nil {
		t.Fatalf("SaveIssue failed: %v", err)
	}

	comment := Comment{
		ID:          1001,
		IssueNumber: 456,
		Body:        "This is a test comment",
		Author:      "commenter",
		CreatedAt:   "2024-01-15T11:00:00Z",
		UpdatedAt:   "2024-01-15T11:00:00Z",
	}

	t.Run("saves comment successfully", func(t *testing.T) {
		err := SaveComment(db, "owner/repo", comment)
		if err != nil {
			t.Errorf("SaveComment failed: %v", err)
		}
	})

	t.Run("updates existing comment", func(t *testing.T) {
		comment.Body = "Updated comment body"
		err := SaveComment(db, "owner/repo", comment)
		if err != nil {
			t.Errorf("SaveComment update failed: %v", err)
		}
	})
}

func TestGetIssueCount(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	t.Run("returns zero for empty database", func(t *testing.T) {
		count, err := GetIssueCount(db, "owner/repo")
		if err != nil {
			t.Errorf("GetIssueCount failed: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected 0 issues, got %d", count)
		}
	})

	t.Run("returns correct count after saving issues", func(t *testing.T) {
		// Save some issues
		for i := 1; i <= 3; i++ {
			issue := Issue{
				Number: i,
				Title:  "Test Issue",
				Author: "testuser",
				State:  "open",
			}
			err := SaveIssue(db, "owner/repo", issue)
			if err != nil {
				t.Fatalf("SaveIssue failed: %v", err)
			}
		}

		count, err := GetIssueCount(db, "owner/repo")
		if err != nil {
			t.Errorf("GetIssueCount failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected 3 issues, got %d", count)
		}
	})
}

func TestParseLabelsAndAssignees(t *testing.T) {
	t.Run("parses labels from JSON array", func(t *testing.T) {
		jsonData := `["bug", "help wanted", "good first issue"]`
		labels := parseLabels(jsonData)

		if len(labels) != 3 {
			t.Errorf("Expected 3 labels, got %d", len(labels))
		}
		if labels[0] != "bug" {
			t.Errorf("Expected first label to be 'bug', got %s", labels[0])
		}
	})

	t.Run("handles empty labels", func(t *testing.T) {
		labels := parseLabels("")
		if len(labels) != 0 {
			t.Errorf("Expected 0 labels, got %d", len(labels))
		}
	})

	t.Run("parses assignees from JSON array", func(t *testing.T) {
		jsonData := `["user1", "user2"]`
		assignees := parseAssignees(jsonData)

		if len(assignees) != 2 {
			t.Errorf("Expected 2 assignees, got %d", len(assignees))
		}
		if assignees[0] != "user1" {
			t.Errorf("Expected first assignee to be 'user1', got %s", assignees[0])
		}
	})
}

func TestSyncMetadata(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	t.Run("returns empty string when no sync metadata exists", func(t *testing.T) {
		lastSync, err := GetLastSyncTime(db, "owner/repo")
		if err != nil {
			t.Errorf("GetLastSyncTime failed: %v", err)
		}
		if lastSync != "" {
			t.Errorf("Expected empty string, got %s", lastSync)
		}
	})

	t.Run("saves and retrieves last sync time", func(t *testing.T) {
		syncTime := "2024-01-20T10:30:00Z"
		err := SaveLastSyncTime(db, "owner/repo", syncTime)
		if err != nil {
			t.Errorf("SaveLastSyncTime failed: %v", err)
		}

		lastSync, err := GetLastSyncTime(db, "owner/repo")
		if err != nil {
			t.Errorf("GetLastSyncTime failed: %v", err)
		}
		if lastSync != syncTime {
			t.Errorf("Expected %s, got %s", syncTime, lastSync)
		}
	})

	t.Run("updates existing sync time", func(t *testing.T) {
		newSyncTime := "2024-01-21T15:45:00Z"
		err := SaveLastSyncTime(db, "owner/repo", newSyncTime)
		if err != nil {
			t.Errorf("SaveLastSyncTime update failed: %v", err)
		}

		lastSync, err := GetLastSyncTime(db, "owner/repo")
		if err != nil {
			t.Errorf("GetLastSyncTime failed: %v", err)
		}
		if lastSync != newSyncTime {
			t.Errorf("Expected %s, got %s", newSyncTime, lastSync)
		}
	})
}

func TestDeleteIssue(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	// Save an issue with comments
	issue := Issue{
		Number:       789,
		Title:        "Issue to Delete",
		Author:       "testuser",
		State:        "open",
		CommentCount: 2,
	}
	err = SaveIssue(db, "owner/repo", issue)
	if err != nil {
		t.Fatalf("SaveIssue failed: %v", err)
	}

	// Add comments
	for i := 1; i <= 2; i++ {
		comment := Comment{
			ID:          i,
			IssueNumber: 789,
			Body:        "Comment",
			Author:      "user",
			CreatedAt:   "2024-01-15T10:00:00Z",
			UpdatedAt:   "2024-01-15T10:00:00Z",
		}
		err = SaveComment(db, "owner/repo", comment)
		if err != nil {
			t.Fatalf("SaveComment failed: %v", err)
		}
	}

	t.Run("deletes issue and its comments", func(t *testing.T) {
		// Verify issue exists
		count, _ := GetIssueCount(db, "owner/repo")
		if count != 1 {
			t.Errorf("Expected 1 issue before deletion, got %d", count)
		}

		// Delete the issue
		err := DeleteIssue(db, "owner/repo", 789)
		if err != nil {
			t.Errorf("DeleteIssue failed: %v", err)
		}

		// Verify issue is deleted
		count, _ = GetIssueCount(db, "owner/repo")
		if count != 0 {
			t.Errorf("Expected 0 issues after deletion, got %d", count)
		}

		// Verify comments are also deleted
		commentCount, _ := GetCommentCount(db, "owner/repo")
		if commentCount != 0 {
			t.Errorf("Expected 0 comments after deletion, got %d", commentCount)
		}
	})
}

func TestDeleteCommentsForIssue(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	// Save an issue with comments
	issue := Issue{
		Number:       100,
		Title:        "Issue with Comments",
		Author:       "testuser",
		State:        "open",
		CommentCount: 3,
	}
	err = SaveIssue(db, "owner/repo", issue)
	if err != nil {
		t.Fatalf("SaveIssue failed: %v", err)
	}

	// Add comments
	for i := 1; i <= 3; i++ {
		comment := Comment{
			ID:          i,
			IssueNumber: 100,
			Body:        "Comment",
			Author:      "user",
			CreatedAt:   "2024-01-15T10:00:00Z",
			UpdatedAt:   "2024-01-15T10:00:00Z",
		}
		err = SaveComment(db, "owner/repo", comment)
		if err != nil {
			t.Fatalf("SaveComment failed: %v", err)
		}
	}

	t.Run("deletes all comments for an issue", func(t *testing.T) {
		// Verify comments exist
		comments, _ := GetCommentsForIssue(db, "owner/repo", 100)
		if len(comments) != 3 {
			t.Errorf("Expected 3 comments before deletion, got %d", len(comments))
		}

		// Delete comments for the issue
		err := DeleteCommentsForIssue(db, "owner/repo", 100)
		if err != nil {
			t.Errorf("DeleteCommentsForIssue failed: %v", err)
		}

		// Verify comments are deleted
		comments, _ = GetCommentsForIssue(db, "owner/repo", 100)
		if len(comments) != 0 {
			t.Errorf("Expected 0 comments after deletion, got %d", len(comments))
		}

		// Issue should still exist
		count, _ := GetIssueCount(db, "owner/repo")
		if count != 1 {
			t.Errorf("Expected 1 issue after comment deletion, got %d", count)
		}
	})
}

func TestGetAllIssueNumbers(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitializeSchema(dbPath)
	if err != nil {
		t.Fatalf("InitializeSchema failed: %v", err)
	}
	defer db.Close()

	t.Run("returns empty slice for empty database", func(t *testing.T) {
		numbers, err := GetAllIssueNumbers(db, "owner/repo")
		if err != nil {
			t.Errorf("GetAllIssueNumbers failed: %v", err)
		}
		if len(numbers) != 0 {
			t.Errorf("Expected 0 numbers, got %d", len(numbers))
		}
	})

	t.Run("returns all issue numbers", func(t *testing.T) {
		// Save multiple issues
		for i := 1; i <= 3; i++ {
			issue := Issue{
				Number: i * 10, // 10, 20, 30
				Title:  "Test Issue",
				Author: "testuser",
				State:  "open",
			}
			err := SaveIssue(db, "owner/repo", issue)
			if err != nil {
				t.Fatalf("SaveIssue failed: %v", err)
			}
		}

		numbers, err := GetAllIssueNumbers(db, "owner/repo")
		if err != nil {
			t.Errorf("GetAllIssueNumbers failed: %v", err)
		}
		if len(numbers) != 3 {
			t.Errorf("Expected 3 numbers, got %d", len(numbers))
		}

		// Check that we got the right numbers
		expected := map[int]bool{10: true, 20: true, 30: true}
		for _, num := range numbers {
			if !expected[num] {
				t.Errorf("Unexpected issue number: %d", num)
			}
		}
	})
}
