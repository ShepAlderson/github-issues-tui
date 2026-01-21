package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/db"
	"github.com/shepbook/ghissues/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncer_Sync_BasicFetch(t *testing.T) {
	// Set up mock server
	server := setupMockServer(t, []github.Issue{
		{Number: 1, Title: "Issue 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
		{Number: 2, Title: "Issue 2", Author: github.User{Login: "user2"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
	})
	defer server.Close()

	// Set up database
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Create syncer with mock client
	client := github.NewClient("test-token")
	client.SetBaseURL(server.URL)

	syncer := NewSyncer(client, store)

	// Run sync
	ctx := context.Background()
	result, err := syncer.Sync(ctx, "owner", "repo", nil)

	require.NoError(t, err)
	assert.Equal(t, 2, result.IssuesFetched)

	// Verify issues in database
	issues, err := store.GetAllIssues(ctx)
	require.NoError(t, err)
	assert.Len(t, issues, 2)
}

func TestSyncer_Sync_FetchesComments(t *testing.T) {
	// Set up mock server that returns issues with comments
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var reqBody struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)

		if containsIssuesQuery(reqBody.Query) {
			writeIssuesResponse(w, []github.Issue{
				{Number: 1, Title: "Issue 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z", CommentCount: 2},
			}, false, 1)
		} else if containsCommentsQuery(reqBody.Query) {
			writeCommentsResponse(w, []github.Comment{
				{ID: "c1", Body: "Comment 1", Author: github.User{Login: "commenter1"}, CreatedAt: "2024-01-15T11:00:00Z"},
				{ID: "c2", Body: "Comment 2", Author: github.User{Login: "commenter2"}, CreatedAt: "2024-01-15T12:00:00Z"},
			})
		}
	}))
	defer server.Close()

	// Set up database
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Create syncer with mock client
	client := github.NewClient("test-token")
	client.SetBaseURL(server.URL)

	syncer := NewSyncer(client, store)

	// Run sync
	ctx := context.Background()
	result, err := syncer.Sync(ctx, "owner", "repo", nil)

	require.NoError(t, err)
	assert.Equal(t, 1, result.IssuesFetched)
	assert.Equal(t, 2, result.CommentsFetched)

	// Verify comments in database
	comments, err := store.GetComments(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, comments, 2)
}

func TestSyncer_Sync_ProgressCallback(t *testing.T) {
	server := setupMockServer(t, []github.Issue{
		{Number: 1, Title: "Issue 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
		{Number: 2, Title: "Issue 2", Author: github.User{Login: "user2"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
	})
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	client := github.NewClient("test-token")
	client.SetBaseURL(server.URL)

	syncer := NewSyncer(client, store)

	var progressCalls []Progress
	ctx := context.Background()
	_, err = syncer.Sync(ctx, "owner", "repo", func(p Progress) {
		progressCalls = append(progressCalls, p)
	})

	require.NoError(t, err)
	assert.NotEmpty(t, progressCalls)

	// Check that progress reports fetched/total
	finalProgress := progressCalls[len(progressCalls)-1]
	assert.Equal(t, 2, finalProgress.IssuesFetched)
	assert.Equal(t, 2, finalProgress.TotalIssues)
}

func TestSyncer_Sync_Cancellation(t *testing.T) {
	// Server that returns pagination, giving us time to cancel
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeIssuesResponse(w, []github.Issue{
			{Number: 1, Title: "Issue 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
		}, true, 100) // hasNextPage=true, simulating a large repo
	}))
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	client := github.NewClient("test-token")
	client.SetBaseURL(server.URL)

	syncer := NewSyncer(client, store)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = syncer.Sync(ctx, "owner", "repo", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestSyncer_Sync_EmptyRepository(t *testing.T) {
	server := setupMockServer(t, []github.Issue{})
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	client := github.NewClient("test-token")
	client.SetBaseURL(server.URL)

	syncer := NewSyncer(client, store)

	ctx := context.Background()
	result, err := syncer.Sync(ctx, "owner", "repo", nil)

	require.NoError(t, err)
	assert.Equal(t, 0, result.IssuesFetched)

	issues, err := store.GetAllIssues(ctx)
	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestSyncer_Sync_UpdatesLastSyncTime(t *testing.T) {
	server := setupMockServer(t, []github.Issue{
		{Number: 1, Title: "Issue 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
	})
	defer server.Close()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	client := github.NewClient("test-token")
	client.SetBaseURL(server.URL)

	syncer := NewSyncer(client, store)

	ctx := context.Background()

	// Initially no sync time
	lastSync, err := store.GetLastSyncTime(ctx)
	require.NoError(t, err)
	assert.True(t, lastSync.IsZero())

	// Run sync
	_, err = syncer.Sync(ctx, "owner", "repo", nil)
	require.NoError(t, err)

	// Verify sync time was updated
	lastSync, err = store.GetLastSyncTime(ctx)
	require.NoError(t, err)
	assert.False(t, lastSync.IsZero())
}

// Helper functions

func setupMockServer(t *testing.T, issues []github.Issue) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)

		if containsIssuesQuery(reqBody.Query) {
			writeIssuesResponse(w, issues, false, len(issues))
		} else if containsCommentsQuery(reqBody.Query) {
			writeCommentsResponse(w, []github.Comment{})
		}
	}))
}

func containsIssuesQuery(query string) bool {
	return len(query) > 0 && (query[0:50] != "" && containsString(query, "issues("))
}

func containsCommentsQuery(query string) bool {
	return containsString(query, "comments(")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func writeIssuesResponse(w http.ResponseWriter, issues []github.Issue, hasNextPage bool, totalCount int) {
	nodes := make([]map[string]any, len(issues))
	for i, issue := range issues {
		nodes[i] = map[string]any{
			"number":    issue.Number,
			"title":     issue.Title,
			"body":      issue.Body,
			"createdAt": issue.CreatedAt,
			"updatedAt": issue.UpdatedAt,
			"author": map[string]any{
				"login": issue.Author.Login,
			},
			"labels": map[string]any{
				"nodes": []map[string]any{},
			},
			"assignees": map[string]any{
				"nodes": []map[string]any{},
			},
			"comments": map[string]any{
				"totalCount": issue.CommentCount,
			},
		}
	}

	response := map[string]any{
		"data": map[string]any{
			"repository": map[string]any{
				"issues": map[string]any{
					"totalCount": totalCount,
					"pageInfo": map[string]any{
						"hasNextPage": hasNextPage,
						"endCursor":   "cursor1",
					},
					"nodes": nodes,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func writeCommentsResponse(w http.ResponseWriter, comments []github.Comment) {
	nodes := make([]map[string]any, len(comments))
	for i, comment := range comments {
		nodes[i] = map[string]any{
			"id":        comment.ID,
			"body":      comment.Body,
			"createdAt": comment.CreatedAt,
			"author": map[string]any{
				"login": comment.Author.Login,
			},
		}
	}

	response := map[string]any{
		"data": map[string]any{
			"repository": map[string]any{
				"issue": map[string]any{
					"comments": map[string]any{
						"pageInfo": map[string]any{
							"hasNextPage": false,
							"endCursor":   "",
						},
						"nodes": nodes,
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
