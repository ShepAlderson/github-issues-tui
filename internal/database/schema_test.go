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
