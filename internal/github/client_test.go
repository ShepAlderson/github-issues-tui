package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchIssues(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify path
		if r.URL.Path == "/repos/owner/repo/issues" {
			issues := []map[string]interface{}{
				{
					"number":     1,
					"title":      "Test Issue 1",
					"body":       "This is issue 1",
					"state":      "open",
					"created_at": "2024-01-15T10:00:00Z",
					"updated_at": "2024-01-16T14:00:00Z",
					"comments":   2,
					"user":       map[string]string{"login": "author1"},
					"labels":     []map[string]string{{"name": "bug"}},
					"assignees":  []map[string]string{{"login": "user1"}},
				},
				{
					"number":     2,
					"title":      "Test Issue 2",
					"body":       "This is issue 2",
					"state":      "open",
					"created_at": "2024-01-14T09:00:00Z",
					"updated_at": "2024-01-15T13:00:00Z",
					"comments":   0,
					"user":       map[string]string{"login": "author2"},
					"labels":     []map[string]string{{"name": "feature"}, {"name": "help wanted"}},
					"assignees":  []map[string]string{},
				},
			}
			json.NewEncoder(w).Encode(issues)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	t.Run("fetches issues with pagination", func(t *testing.T) {
		client := NewClient("test_token")
		client.BaseURL = testServer.URL

		issues, err := client.FetchIssues("owner/repo", nil)
		if err != nil {
			t.Fatalf("FetchIssues failed: %v", err)
		}

		if len(issues) != 2 {
			t.Errorf("Expected 2 issues, got %d", len(issues))
		}

		if issues[0].Number != 1 {
			t.Errorf("Expected first issue number 1, got %d", issues[0].Number)
		}

		if issues[0].Title != "Test Issue 1" {
			t.Errorf("Expected title 'Test Issue 1', got %s", issues[0].Title)
		}

		if issues[0].Author != "author1" {
			t.Errorf("Expected author 'author1', got %s", issues[0].Author)
		}

		if len(issues[0].Labels) != 1 || issues[0].Labels[0] != "bug" {
			t.Errorf("Expected label 'bug', got %v", issues[0].Labels)
		}

		if len(issues[0].Assignees) != 1 || issues[0].Assignees[0] != "user1" {
			t.Errorf("Expected assignee 'user1', got %v", issues[0].Assignees)
		}
	})

	t.Run("handles pagination correctly", func(t *testing.T) {
		pageCount := 0
		paginatedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := r.URL.Query().Get("page")
			if page == "" || page == "1" {
				pageCount++
				w.Header().Set("Link", `<`+r.URL.Path+`?page=2>; rel="next"`)
				issues := []map[string]interface{}{
					{
						"number": 1,
						"title":  "Page 1 Issue",
						"state":  "open",
						"user":   map[string]string{"login": "author"},
					},
				}
				json.NewEncoder(w).Encode(issues)
			} else if page == "2" {
				pageCount++
				w.Header().Set("Link", `<`+r.URL.Path+`?page=3>; rel="next"`)
				issues := []map[string]interface{}{
					{
						"number": 2,
						"title":  "Page 2 Issue",
						"state":  "open",
						"user":   map[string]string{"login": "author"},
					},
				}
				json.NewEncoder(w).Encode(issues)
			} else {
				// No more pages
				pageCount++
				json.NewEncoder(w).Encode([]map[string]interface{}{})
			}
		}))
		defer paginatedServer.Close()

		client := NewClient("test_token")
		client.BaseURL = paginatedServer.URL

		// Create a channel to receive progress updates
		progressChan := make(chan FetchProgress, 10)

		go func() {
			for range progressChan {
				// Consume progress updates
			}
		}()

		issues, err := client.FetchIssues("owner/repo", progressChan)
		if err != nil {
			t.Fatalf("FetchIssues failed: %v", err)
		}

		// Should have fetched issues from both pages
		if len(issues) != 2 {
			t.Errorf("Expected 2 issues total, got %d", len(issues))
		}

		close(progressChan)
	})

	t.Run("returns error on authentication failure", func(t *testing.T) {
		authFailureServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Bad credentials",
			})
		}))
		defer authFailureServer.Close()

		client := NewClient("bad_token")
		client.BaseURL = authFailureServer.URL

		_, err := client.FetchIssues("owner/repo", nil)
		if err == nil {
			t.Error("Expected error on authentication failure")
		}
	})
}

func TestFetchComments(t *testing.T) {
	// Create a test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify path
		if r.URL.Path == "/repos/owner/repo/issues/1/comments" {
			comments := []map[string]interface{}{
				{
					"id":         1001,
					"body":       "Test comment 1",
					"created_at": "2024-01-15T11:00:00Z",
					"updated_at": "2024-01-15T11:00:00Z",
					"user":       map[string]string{"login": "commenter1"},
				},
				{
					"id":         1002,
					"body":       "Test comment 2",
					"created_at": "2024-01-15T12:00:00Z",
					"updated_at": "2024-01-15T12:00:00Z",
					"user":       map[string]string{"login": "commenter2"},
				},
			}
			json.NewEncoder(w).Encode(comments)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	t.Run("fetches comments for an issue", func(t *testing.T) {
		client := NewClient("test_token")
		client.BaseURL = testServer.URL

		comments, err := client.FetchComments("owner/repo", 1, nil)
		if err != nil {
			t.Fatalf("FetchComments failed: %v", err)
		}

		if len(comments) != 2 {
			t.Errorf("Expected 2 comments, got %d", len(comments))
		}

		if comments[0].ID != 1001 {
			t.Errorf("Expected first comment ID 1001, got %d", comments[0].ID)
		}

		if comments[0].Body != "Test comment 1" {
			t.Errorf("Expected body 'Test comment 1', got %s", comments[0].Body)
		}

		if comments[0].Author != "commenter1" {
			t.Errorf("Expected author 'commenter1', got %s", comments[0].Author)
		}

		if comments[0].IssueNumber != 1 {
			t.Errorf("Expected issue number 1, got %d", comments[0].IssueNumber)
		}
	})
}

func TestParseGitHubRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "valid owner/repo format",
			input:     "owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "valid owner/repo with hyphens",
			input:     "my-org/my-repo",
			wantOwner: "my-org",
			wantName:  "my-repo",
			wantErr:   false,
		},
		{
			name:      "valid with numbers",
			input:     "org123/repo456",
			wantOwner: "org123",
			wantName:  "repo456",
			wantErr:   false,
		},
		{
			name:      "invalid - missing slash",
			input:     "ownerrepo",
			wantOwner: "",
			wantName:  "",
			wantErr:   true,
		},
		{
			name:      "invalid - empty",
			input:     "",
			wantOwner: "",
			wantName:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := ParseGitHubRepoURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGitHubRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner {
				t.Errorf("ParseGitHubRepoURL() owner = %v, want %v", owner, tt.wantOwner)
			}
			if name != tt.wantName {
				t.Errorf("ParseGitHubRepoURL() name = %v, want %v", name, tt.wantName)
			}
		})
	}
}

func TestFetchIssuesSince(t *testing.T) {
	// Create a test server that checks for the 'since' parameter
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check for since parameter
		sinceParam := r.URL.Query().Get("since")
		if sinceParam == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Missing since parameter",
			})
			return
		}

		// Verify path
		if r.URL.Path == "/repos/owner/repo/issues" {
			// Return only issues updated since the given time
			issues := []map[string]interface{}{
				{
					"number":     3,
					"title":      "Recently Updated Issue",
					"body":       "This issue was updated",
					"state":      "open",
					"created_at": "2024-01-15T10:00:00Z",
					"updated_at": "2024-01-20T16:00:00Z",
					"comments":   1,
					"user":       map[string]string{"login": "author3"},
					"labels":     []map[string]string{},
					"assignees":  []map[string]string{},
				},
			}
			json.NewEncoder(w).Encode(issues)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testServer.Close()

	t.Run("fetches only issues updated since given timestamp", func(t *testing.T) {
		client := NewClient("test_token")
		client.BaseURL = testServer.URL

		sinceTime := "2024-01-20T00:00:00Z"
		issues, err := client.FetchIssuesSince("owner/repo", sinceTime, nil)
		if err != nil {
			t.Fatalf("FetchIssuesSince failed: %v", err)
		}

		if len(issues) != 1 {
			t.Errorf("Expected 1 issue, got %d", len(issues))
		}

		if issues[0].Number != 3 {
			t.Errorf("Expected issue number 3, got %d", issues[0].Number)
		}

		if issues[0].Title != "Recently Updated Issue" {
			t.Errorf("Expected title 'Recently Updated Issue', got %s", issues[0].Title)
		}
	})

	t.Run("returns empty when no issues updated since time", func(t *testing.T) {
		emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Return empty array
			json.NewEncoder(w).Encode([]map[string]interface{}{})
		}))
		defer emptyServer.Close()

		client := NewClient("test_token")
		client.BaseURL = emptyServer.URL

		sinceTime := "2024-01-25T00:00:00Z"
		issues, err := client.FetchIssuesSince("owner/repo", sinceTime, nil)
		if err != nil {
			t.Fatalf("FetchIssuesSince failed: %v", err)
		}

		if len(issues) != 0 {
			t.Errorf("Expected 0 issues, got %d", len(issues))
		}
	})
}
