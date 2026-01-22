package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		baseURL     string
		expectError bool
	}{
		{
			name:        "Creates client successfully",
			token:       "test_token",
			baseURL:     "https://api.github.com",
			expectError: false,
		},
		{
			name:        "Handles empty token",
			token:       "",
			baseURL:     "https://api.github.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.token, tt.baseURL)
			if client == nil {
				t.Error("Expected client instance, got nil")
			}

			if client.token != tt.token {
				t.Errorf("Expected token %q, got %q", tt.token, client.token)
			}

			if client.baseURL != tt.baseURL {
				t.Errorf("Expected baseURL %q, got %q", tt.baseURL, client.baseURL)
			}
		})
	}
}

func TestClient_FetchOpenIssues(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token', got %q", authHeader)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("state") != "open" {
			t.Errorf("Expected state=open parameter, got %q", query.Get("state"))
		}
		if query.Get("per_page") != "100" {
			t.Errorf("Expected per_page=100 parameter, got %q", query.Get("per_page"))
		}

		// Return mock response based on page
		page := query.Get("page")
		if page == "2" {
			// Second page - return empty array
			json.NewEncoder(w).Encode([]Issue{})
		} else {
			// First page - return 2 issues
			issues := []Issue{
				{
					Number:    1,
					Title:     "First Issue",
					Body:      "Body of first issue",
					State:     "open",
					User:      User{Login: "user1"},
					CreatedAt: "2026-01-20T10:00:00Z",
					UpdatedAt: "2026-01-20T10:00:00Z",
					Comments:  5,
					Labels:    []Label{{Name: "bug"}, {Name: "help wanted"}},
					Assignees: []User{{Login: "assignee1"}},
				},
				{
					Number:    2,
					Title:     "Second Issue",
					Body:      "Body of second issue",
					State:     "open",
					User:      User{Login: "user2"},
					CreatedAt: "2026-01-20T11:00:00Z",
					UpdatedAt: "2026-01-20T11:00:00Z",
					Comments:  3,
					Labels:    []Label{{Name: "enhancement"}},
					Assignees: []User{{Login: "assignee2"}},
				},
			}
			json.NewEncoder(w).Encode(issues)
		}
	}))
	defer server.Close()

	client := NewClient("test_token", server.URL)
	issues, err := client.FetchOpenIssues("testowner/testrepo")
	if err != nil {
		t.Fatalf("Failed to fetch issues: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}

	// Verify first issue
	if issues[0].Number != 1 {
		t.Errorf("Expected first issue number 1, got %d", issues[0].Number)
	}
	if issues[0].Title != "First Issue" {
		t.Errorf("Expected title 'First Issue', got %q", issues[0].Title)
	}
	if len(issues[0].Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issues[0].Labels))
	}
	if len(issues[0].Assignees) != 1 {
		t.Errorf("Expected 1 assignee, got %d", len(issues[0].Assignees))
	}
}

func TestClient_FetchIssueComments(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL path
		expectedPath := "/repos/testowner/testrepo/issues/1/comments"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}

		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token', got %q", authHeader)
		}

		// Return mock comments
		comments := []Comment{
			{
				ID:        100,
				Body:      "First comment on issue 1",
				User:      User{Login: "commenter1"},
				CreatedAt: "2026-01-20T10:00:00Z",
			},
			{
				ID:        101,
				Body:      "Second comment on issue 1",
				User:      User{Login: "commenter2"},
				CreatedAt: "2026-01-20T10:05:00Z",
			},
		}
		json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := NewClient("test_token", server.URL)
	comments, err := client.FetchIssueComments("testowner/testrepo", 1)
	if err != nil {
		t.Fatalf("Failed to fetch comments: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}

	// Verify first comment
	if comments[0].ID != 100 {
		t.Errorf("Expected first comment ID 100, got %d", comments[0].ID)
	}
	if comments[0].Body != "First comment on issue 1" {
		t.Errorf("Expected comment body 'First comment on issue 1', got %q", comments[0].Body)
	}
}

func TestClient_FetchOpenIssues_Pagination(t *testing.T) {
	pageCount := 0
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++
		query := r.URL.Query()
		page := query.Get("page")

		// Set Link header for first page to indicate next page exists
		if page == "" || page == "1" {
			w.Header().Set("Link", `<http://example.com?page=2>; rel="next"`)
			// Return 2 issues on first page
			issues := []Issue{
				{Number: 1, Title: "Issue 1", State: "open", User: User{Login: "user1"}},
				{Number: 2, Title: "Issue 2", State: "open", User: User{Login: "user2"}},
			}
			json.NewEncoder(w).Encode(issues)
		} else if page == "2" {
			// Return 1 issue on second page
			issues := []Issue{
				{Number: 3, Title: "Issue 3", State: "open", User: User{Login: "user3"}},
			}
			json.NewEncoder(w).Encode(issues)
		}
	}))
	defer server.Close()

	client := NewClient("test_token", server.URL)
	issues, err := client.FetchOpenIssues("testowner/testrepo")
	if err != nil {
		t.Fatalf("Failed to fetch issues: %v", err)
	}

	// Should fetch all 3 issues across 2 pages
	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Should have made 2 API calls
	if pageCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", pageCount)
	}
}

func TestClient_FetchOpenIssues_HandlesError(t *testing.T) {
	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Bad credentials",
		})
	}))
	defer server.Close()

	client := NewClient("invalid_token", server.URL)
	_, err := client.FetchOpenIssues("testowner/testrepo")
	if err == nil {
		t.Error("Expected error for unauthorized request, got nil")
	}
}

func TestClient_GetRepo(t *testing.T) {
	// GetRepo should parse owner/repo format
	tests := []struct {
		repo        string
		expectOwner string
		expectName  string
		expectError bool
	}{
		{
			repo:        "owner/repo",
			expectOwner: "owner",
			expectName:  "repo",
			expectError: false,
		},
		{
			repo:        "user/my-repo",
			expectOwner: "user",
			expectName:  "my-repo",
			expectError: false,
		},
		{
			repo:        "invalid",
			expectError: true,
		},
		{
			repo:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			owner, name, err := getRepo(tt.repo)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if owner != tt.expectOwner {
				t.Errorf("Expected owner %q, got %q", tt.expectOwner, owner)
			}

			if name != tt.expectName {
				t.Errorf("Expected name %q, got %q", tt.expectName, name)
			}
		})
	}
}

func TestClient_FetchIssuesSince(t *testing.T) {
	// Verify that since parameter is correctly sent
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true

		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token', got %q", authHeader)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("state") != "open" {
			t.Errorf("Expected state=open parameter, got %q", query.Get("state"))
		}
		if query.Get("per_page") != "100" {
			t.Errorf("Expected per_page=100 parameter, got %q", query.Get("per_page"))
		}
		if query.Get("page") != "1" {
			t.Errorf("Expected page=1 parameter, got %q", query.Get("page"))
		}
		if query.Get("since") != "2026-01-20T10:00:00Z" {
			t.Errorf("Expected since=2026-01-20T10:00:00Z parameter, got %q", query.Get("since"))
		}

		// Return mock response - only issues updated since the date
		issues := []Issue{
			{
				Number:    3,
				Title:     "Updated Issue",
				Body:      "Body of updated issue",
				State:     "open",
				User:      User{Login: "user3"},
				CreatedAt: "2026-01-20T09:00:00Z",
				UpdatedAt: "2026-01-20T12:00:00Z", // Updated after since date
				Comments:  2,
				Labels:    []Label{{Name: "bug"}},
				Assignees: []User{{Login: "user3"}},
			},
		}
		json.NewEncoder(w).Encode(issues)
	}))
	defer server.Close()

	client := NewClient("test_token", server.URL)
	issues, err := client.FetchIssuesSince("testowner/testrepo", "2026-01-20T10:00:00Z")
	if err != nil {
		t.Fatalf("Failed to fetch issues: %v", err)
	}

	if !serverCalled {
		t.Error("Server was not called")
	}

	// Verify we got the updated issue
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Number != 3 {
		t.Errorf("Expected issue #3, got #%d", issues[0].Number)
	}
}

func TestClient_FetchIssueCommentsSince(t *testing.T) {
	// Verify that since parameter is correctly sent for comments
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true

		// Verify URL path
		expectedPath := "/repos/testowner/testrepo/issues/1/comments"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %q, got %q", expectedPath, r.URL.Path)
		}

		// Verify authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token', got %q", authHeader)
		}

		// Verify since parameter
		query := r.URL.Query()
		if query.Get("since") != "2026-01-20T14:30:00Z" {
			t.Errorf("Expected since=2026-01-20T14:30:00Z parameter, got %q", query.Get("since"))
		}

		// Return mock comments - only new comments since the date
		comments := []Comment{
			{
				ID:        102,
				Body:      "New comment since timestamp",
				User:      User{Login: "commenter3"},
				CreatedAt: "2026-01-20T15:00:00Z", // Created after since date
			},
		}
		json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := NewClient("test_token", server.URL)
	comments, err := client.FetchIssueCommentsSince("testowner/testrepo", 1, "2026-01-20T14:30:00Z")
	if err != nil {
		t.Fatalf("Failed to fetch comments: %v", err)
	}

	if !serverCalled {
		t.Error("Server was not called")
	}

	// Verify we got only the new comment
	if len(comments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(comments))
	}

	if comments[0].ID != 102 {
		t.Errorf("Expected comment #102, got #%d", comments[0].ID)
	}

	if comments[0].Body != "New comment since timestamp" {
		t.Errorf("Expected specific comment body, got %q", comments[0].Body)
	}
}
