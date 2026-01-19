package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// TestSyncerIntegration tests the full sync flow
func TestSyncerIntegration(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle issues endpoint
		if r.URL.Path == "/repos/owner/repo/issues" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"number": 1,
					"title": "Test Issue",
					"body": "Test body",
					"state": "open",
					"user": {"login": "testuser"},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z",
					"comments": 1,
					"labels": [{"name": "bug"}],
					"assignees": [{"login": "dev1"}]
				}
			]`))
			return
		}

		// Handle comments endpoint
		if r.URL.Path == "/repos/owner/repo/issues/1/comments" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": 123,
					"body": "Test comment",
					"user": {"login": "commenter"},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-01T00:00:00Z"
				}
			]`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create syncer
	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	syncer := &Syncer{
		client: &GitHubClient{
			baseURL: server.URL,
			token:   "test-token",
			client:  &http.Client{},
		},
		store: store,
	}

	// Run sync
	err = syncer.syncWithContext(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify issue was stored
	count, err := store.GetIssueCount()
	if err != nil {
		t.Fatalf("Failed to get issue count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 issue, got %d", count)
	}

	// Verify comment was stored
	var commentCount int
	err = store.db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&commentCount)
	if err != nil {
		t.Fatalf("Failed to query comments: %v", err)
	}
	if commentCount != 1 {
		t.Errorf("Expected 1 comment, got %d", commentCount)
	}
}

// TestSyncCancellation tests that sync can be cancelled
func TestSyncCancellation(t *testing.T) {
	// Create test server that delays
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"number": 1, "title": "Test", "state": "open", "user": {"login": "user"}, "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "comments": 0}]`))
	}))
	defer server.Close()

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	syncer := &Syncer{
		client: &GitHubClient{
			baseURL: server.URL,
			token:   "test-token",
			client:  &http.Client{},
		},
		store: store,
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Sync should fail with context error
	err = syncer.syncWithContext(ctx, "owner/repo")
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}
}

// TestEmptyRepository tests syncing an empty repository
func TestEmptyRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	syncer := &Syncer{
		client: &GitHubClient{
			baseURL: server.URL,
			token:   "test-token",
			client:  &http.Client{},
		},
		store: store,
	}

	err = syncer.syncWithContext(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	count, err := store.GetIssueCount()
	if err != nil {
		t.Fatalf("Failed to get issue count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 issues, got %d", count)
	}
}

// TestClearAllData tests clearing all synced data
func TestClearAllData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewIssueStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Add test data
	issue := &Issue{
		Number:    1,
		Title:     "Test",
		State:     "open",
		Author:    "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.StoreIssue(issue); err != nil {
		t.Fatalf("Failed to store issue: %v", err)
	}

	syncer := &Syncer{
		client: NewGitHubClient("test-token"),
		store:  store,
	}

	// Clear data
	if err := syncer.ClearAllData(); err != nil {
		t.Fatalf("ClearAllData failed: %v", err)
	}

	// Verify data is cleared
	count, err := store.GetIssueCount()
	if err != nil {
		t.Fatalf("Failed to get issue count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 issues after clear, got %d", count)
	}
}
