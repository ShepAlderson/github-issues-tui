package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/db"
)

func TestRunSyncCommand(t *testing.T) {
	// Create a mock GitHub server
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		query := r.URL.Query()

		// Handle pagination
		page := query.Get("page")
		if r.URL.Path == "/repos/testowner/testrepo/issues" {
			if page == "" || page == "1" {
				// First page - return 2 issues
				issues := []struct {
					Number    int      `json:"number"`
					Title     string   `json:"title"`
					Body      string   `json:"body"`
					State     string   `json:"state"`
					User      struct{ Login string `json:"login"` } `json:"user"`
					CreatedAt string   `json:"created_at"`
					UpdatedAt string   `json:"updated_at"`
					Comments  int      `json:"comments"`
					Labels    []struct{ Name string `json:"name"` } `json:"labels"`
					Assignees []struct{ Login string `json:"login"` } `json:"assignees"`
				}{
					{
						Number:    1,
						Title:     "First Issue",
						Body:      "Body of first issue",
						State:     "open",
						User:      struct{ Login string `json:"login"` }{Login: "user1"},
						CreatedAt: "2026-01-20T10:00:00Z",
						UpdatedAt: "2026-01-20T10:00:00Z",
						Comments:  2,
						Labels:    []struct{ Name string `json:"name"` }{{Name: "bug"}, {Name: "help wanted"}},
						Assignees: []struct{ Login string `json:"login"` }{{Login: "assignee1"}},
					},
					{
						Number:    2,
						Title:     "Second Issue",
						Body:      "Body of second issue",
						State:     "open",
						User:      struct{ Login string `json:"login"` }{Login: "user2"},
						CreatedAt: "2026-01-20T11:00:00Z",
						UpdatedAt: "2026-01-20T11:00:00Z",
						Comments:  0,
						Labels:    []struct{ Name string `json:"name"` }{{Name: "enhancement"}},
						Assignees: []struct{ Login string `json:"login"` }{},
					},
				}
				// Set Link header to indicate no next page
				w.Header().Set("Link", ``)
				json.NewEncoder(w).Encode(issues)
			}
		} else if r.URL.Path == "/repos/testowner/testrepo/issues/1/comments" {
			// Return comments for issue 1
			comments := []struct {
				ID        int64  `json:"id"`
				Body      string `json:"body"`
				User      struct{ Login string `json:"login"` } `json:"user"`
				CreatedAt string `json:"created_at"`
			}{
				{
					ID:        100,
					Body:      "First comment on issue 1",
					User:      struct{ Login string `json:"login"` }{Login: "commenter1"},
					CreatedAt: "2026-01-20T10:05:00Z",
				},
				{
					ID:        101,
					Body:      "Second comment on issue 1",
					User:      struct{ Login string `json:"login"` }{Login: "commenter2"},
					CreatedAt: "2026-01-20T10:10:00Z",
				},
			}
			json.NewEncoder(w).Encode(comments)
		}
	}))
	defer server.Close()

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize config with mock server URL
	config := &SyncConfig{
		Token:      "test_token",
		Repository: "testowner/testrepo",
		GitHubURL:  server.URL, // Use mock server
	}

	output := &bytes.Buffer{}

	// Run sync command
	err := RunSyncCommand(dbPath, config, output)
	if err != nil {
		t.Fatalf("RunSyncCommand failed: %v", err)
	}

	// Verify output contains progress and success messages
	outputStr := output.String()
	if !bytes.Contains(output.Bytes(), []byte("Fetching issues")) {
		t.Error("Output should contain 'Fetching issues' message")
	}
	if !bytes.Contains(output.Bytes(), []byte("Sync complete")) {
		t.Error("Output should contain 'Sync complete' message")
	}

	// Verify database contents
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Check that both issues were stored
	openIssues, err := database.GetAllOpenIssues()
	if err != nil {
		t.Fatalf("Failed to get open issues: %v", err)
	}

	if len(openIssues) != 2 {
		t.Errorf("Expected 2 open issues, got %d", len(openIssues))
	}

	// Verify first issue details
	issue1, err := database.GetIssue(1)
	if err != nil {
		t.Fatalf("Failed to get issue 1: %v", err)
	}

	if issue1.Title != "First Issue" {
		t.Errorf("Expected title 'First Issue', got %q", issue1.Title)
	}
	if issue1.CommentCount != 2 {
		t.Errorf("Expected comment count 2, got %d", issue1.CommentCount)
	}
	if len(issue1.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issue1.Labels))
	}

	// Verify comments were stored for issue 1
	comments, err := database.GetComments(1)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments for issue 1, got %d", len(comments))
	}

	// Verify second issue has no comments (since Comments count was 0)
	issue2, err := database.GetIssue(2)
	if err != nil {
		t.Fatalf("Failed to get issue 2: %v", err)
	}

	if issue2.CommentCount != 0 {
		t.Errorf("Expected comment count 0, got %d", issue2.CommentCount)
	}

	// Verify progress was shown (check for percentage in output)
	if !containsProgress(outputStr) {
		t.Error("Progress bar should be displayed in output")
	}
}

func TestRunSyncCommand_Cancelled(t *testing.T) {
	// Create a mock server that will be slow to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		// In real scenario, Ctrl+C would trigger context cancellation
		// For testing, we simulate by checking interrupt signal handling
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := &SyncConfig{
		Token:      "test_token",
		Repository: "testowner/testrepo",
		GitHubURL:  server.URL,
	}

	output := &bytes.Buffer{}

	// For now, just test that the command runs without actual cancellation
	// Real Ctrl+C handling would require signal handling in the sync command
	err := RunSyncCommand(dbPath, config, output)
	if err != nil {
		t.Fatalf("RunSyncCommand failed: %v", err)
	}
}

func TestRunSyncCommand_HandlesError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		response := map[string]string{"message": "Bad credentials"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := &SyncConfig{
		Token:      "invalid_token",
		Repository: "testowner/testrepo",
		GitHubURL:  server.URL,
	}

	output := &bytes.Buffer{}

	err := RunSyncCommand(dbPath, config, output)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}

	// Error is returned, not written to output in this case since it fails early
	// Just check that an error was returned
}

func TestRunSyncCommand_EmptyRepository(t *testing.T) {
	// Create a mock server that returns empty issues
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]interface{}{})
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := &SyncConfig{
		Token:      "test_token",
		Repository: "empty/repo",
		GitHubURL:  server.URL,
	}

	output := &bytes.Buffer{}

	err := RunSyncCommand(dbPath, config, output)
	if err != nil {
		t.Fatalf("RunSyncCommand failed: %v", err)
	}

	if !bytes.Contains(output.Bytes(), []byte("No issues found")) {
		t.Error("Output should indicate no issues found")
	}
}

// Helper function to check if output contains progress bar
func containsProgress(output string) bool {
	// Look for percentage indicators commonly used in progress bars
	return bytes.Contains([]byte(output), []byte("%")) ||
		bytes.Contains([]byte(output), []byte("[=")) // Progress bar style
}
