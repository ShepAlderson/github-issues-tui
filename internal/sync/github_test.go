package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestFetchIssues tests fetching issues from GitHub API
func TestFetchIssues(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		statusCode    int
		wantIssues    int
		wantErr       bool
		linkHeader    string // For pagination testing
	}{
		{
			name:       "successful fetch single page",
			statusCode: http.StatusOK,
			response: `[
				{
					"number": 1,
					"title": "Test Issue",
					"body": "Test body",
					"state": "open",
					"user": {"login": "testuser"},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z",
					"comments": 5,
					"labels": [{"name": "bug"}],
					"assignees": [{"login": "assignee1"}]
				}
			]`,
			wantIssues: 1,
			wantErr:    false,
		},
		{
			name:       "empty repository",
			statusCode: http.StatusOK,
			response:   `[]`,
			wantIssues: 0,
			wantErr:    false,
		},
		{
			name:       "API error",
			statusCode: http.StatusInternalServerError,
			response:   `{"message": "Internal server error"}`,
			wantIssues: 0,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"message": "Bad credentials"}`,
			wantIssues: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Expected Authorization header with Bearer token")
				}
				if r.URL.Query().Get("state") != "open" {
					t.Errorf("Expected state=open query parameter")
				}
				if r.URL.Query().Get("per_page") != "100" {
					t.Errorf("Expected per_page=100 query parameter")
				}

				// Set headers
				if tt.linkHeader != "" {
					w.Header().Set("Link", tt.linkHeader)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// Create client with test server
			client := &GitHubClient{
				baseURL: server.URL,
				token:   "test-token",
				client:  &http.Client{},
			}

			// Fetch issues
			issues, err := client.FetchIssues(context.Background(), "owner/repo")

			// Verify results
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchIssues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(issues) != tt.wantIssues {
				t.Errorf("FetchIssues() got %d issues, want %d", len(issues), tt.wantIssues)
			}

			// Verify issue data if we got issues
			if len(issues) > 0 && !tt.wantErr {
				issue := issues[0]
				if issue.Number != 1 {
					t.Errorf("Issue number = %d, want 1", issue.Number)
				}
				if issue.Title != "Test Issue" {
					t.Errorf("Issue title = %s, want 'Test Issue'", issue.Title)
				}
				if issue.Author != "testuser" {
					t.Errorf("Issue author = %s, want 'testuser'", issue.Author)
				}
				if issue.CommentCount != 5 {
					t.Errorf("Issue comment count = %d, want 5", issue.CommentCount)
				}
				if len(issue.Labels) != 1 || issue.Labels[0] != "bug" {
					t.Errorf("Issue labels = %v, want ['bug']", issue.Labels)
				}
				if len(issue.Assignees) != 1 || issue.Assignees[0] != "assignee1" {
					t.Errorf("Issue assignees = %v, want ['assignee1']", issue.Assignees)
				}
			}
		})
	}
}

// TestFetchIssuesPagination tests pagination handling
func TestFetchIssuesPagination(t *testing.T) {
	pageCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++

		// First page returns link to next page
		if r.URL.Query().Get("page") == "" || r.URL.Query().Get("page") == "1" {
			w.Header().Set("Link", `<http://example.com?page=2>; rel="next"`)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"number": 1, "title": "Issue 1", "state": "open", "user": {"login": "user1"}, "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "comments": 0}]`))
			return
		}

		// Second page (no next link)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"number": 2, "title": "Issue 2", "state": "open", "user": {"login": "user2"}, "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "comments": 0}]`))
	}))
	defer server.Close()

	client := &GitHubClient{
		baseURL: server.URL,
		token:   "test-token",
		client:  &http.Client{},
	}

	issues, err := client.FetchIssues(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("FetchIssues() error = %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues from pagination, got %d", len(issues))
	}

	if pageCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", pageCount)
	}
}

// TestFetchComments tests fetching comments for an issue
func TestFetchComments(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		statusCode   int
		wantComments int
		wantErr      bool
	}{
		{
			name:       "successful fetch",
			statusCode: http.StatusOK,
			response: `[
				{
					"id": 1,
					"body": "Test comment",
					"user": {"login": "commenter"},
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			]`,
			wantComments: 1,
			wantErr:      false,
		},
		{
			name:         "no comments",
			statusCode:   http.StatusOK,
			response:     `[]`,
			wantComments: 0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := &GitHubClient{
				baseURL: server.URL,
				token:   "test-token",
				client:  &http.Client{},
			}

			comments, err := client.FetchComments(context.Background(), "owner/repo", 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("FetchComments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(comments) != tt.wantComments {
				t.Errorf("FetchComments() got %d comments, want %d", len(comments), tt.wantComments)
			}

			if len(comments) > 0 && !tt.wantErr {
				comment := comments[0]
				if comment.Body != "Test comment" {
					t.Errorf("Comment body = %s, want 'Test comment'", comment.Body)
				}
				if comment.Author != "commenter" {
					t.Errorf("Comment author = %s, want 'commenter'", comment.Author)
				}
			}
		})
	}
}

// TestContextCancellation tests that operations respect context cancellation
func TestContextCancellation(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := &GitHubClient{
		baseURL: server.URL,
		token:   "test-token",
		client:  &http.Client{},
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.FetchIssues(ctx, "owner/repo")
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}
}
