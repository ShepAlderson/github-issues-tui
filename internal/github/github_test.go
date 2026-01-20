package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFetchIssues(t *testing.T) {
	// Track page requests
	pageRequests := make(map[string]int)

	// Create a test server that returns GitHub API responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track which page was requested
		if strings.Contains(r.URL.String(), "page=2") {
			pageRequests["page2"]++
		} else {
			pageRequests["page1"]++
		}

		// Check authentication
		auth := r.Header.Get("Authorization")
		if auth != "token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check pagination - return first page with link to second
		if !strings.Contains(r.URL.String(), "page=") {
			// Return first page with pagination headers
			nextURL := r.URL.String() + "&page=2"
			w.Header().Set("Link", "<"+nextURL+`>; rel="next"`)
			issues := []map[string]interface{}{
				{
					"number": 1,
					"title":  "First Issue",
					"body":   "First issue body",
					"user": map[string]interface{}{
						"login": "user1",
					},
					"state":       "open",
					"created_at":  "2024-01-01T00:00:00Z",
					"updated_at":  "2024-01-02T00:00:00Z",
					"comments":    5,
					"labels":      []map[string]string{{"name": "bug"}, {"name": "enhancement"}},
					"assignees":   []map[string]string{{"login": "assignee1"}},
				},
			}
			json.NewEncoder(w).Encode(issues)
		} else if strings.Contains(r.URL.String(), "page=2") {
			// Return second page (no next link)
			issues := []map[string]interface{}{
				{
					"number": 2,
					"title":  "Second Issue",
					"body":   "Second issue body",
					"user": map[string]interface{}{
						"login": "user2",
					},
					"state":       "open",
					"created_at":  "2024-01-03T00:00:00Z",
					"updated_at":  "2024-01-04T00:00:00Z",
					"comments":    3,
					"labels":      []map[string]string{{"name": "documentation"}},
					"assignees":   []map[string]string{},
				},
			}
			json.NewEncoder(w).Encode(issues)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient("test-token", "owner/repo", server.URL)

	// Mock progress callback
	progress := make(chan int)
	go func() {
		for range progress {
			// Consume progress updates
		}
	}()

	// Fetch issues
	issues, err := client.FetchIssues(progress, nil)
	close(progress)

	if err != nil {
		t.Fatalf("FetchIssues failed: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}

	// Verify both pages were requested
	if pageRequests["page1"] != 1 {
		t.Errorf("Expected 1 request to page 1, got %d", pageRequests["page1"])
	}
	if pageRequests["page2"] != 1 {
		t.Errorf("Expected 1 request to page 2, got %d", pageRequests["page2"])
	}

	// Verify first issue
	if issues[0].Number != 1 {
		t.Errorf("Expected issue number 1, got %d", issues[0].Number)
	}
	if issues[0].Title != "First Issue" {
		t.Errorf("Expected title 'First Issue', got '%s'", issues[0].Title)
	}
	if issues[0].Author != "user1" {
		t.Errorf("Expected author 'user1', got '%s'", issues[0].Author)
	}
	if issues[0].Comments != 5 {
		t.Errorf("Expected 5 comments, got %d", issues[0].Comments)
	}
	if issues[0].Labels != "bug,enhancement" {
		t.Errorf("Expected labels 'bug,enhancement', got '%s'", issues[0].Labels)
	}
	if issues[0].Assignees != "assignee1" {
		t.Errorf("Expected assignees 'assignee1', got '%s'", issues[0].Assignees)
	}

	// Verify second issue
	if issues[1].Number != 2 {
		t.Errorf("Expected issue number 2, got %d", issues[1].Number)
	}
	if issues[1].Assignees != "" {
		t.Errorf("Expected empty assignees for issue 2, got '%s'", issues[1].Assignees)
	}
}

func TestFetchIssuesUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient("bad-token", "owner/repo", server.URL)
	progress := make(chan int)

	_, err := client.FetchIssues(progress, nil)
	close(progress)

	if err == nil {
		t.Fatal("Expected error for unauthorized request, got nil")
	}

	if !strings.Contains(err.Error(), "authentication") && !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("Expected authentication error, got: %v", err)
	}
}

func TestFetchComments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authentication
		auth := r.Header.Get("Authorization")
		if auth != "token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		comments := []map[string]interface{}{
			{
				"id":   1,
				"body": "First comment",
				"user": map[string]interface{}{
					"login": "commenter1",
				},
				"created_at": "2024-01-01T01:00:00Z",
				"updated_at": "2024-01-01T01:00:00Z",
			},
			{
				"id":   2,
				"body": "Second comment",
				"user": map[string]interface{}{
					"login": "commenter2",
				},
				"created_at": "2024-01-01T02:00:00Z",
				"updated_at": "2024-01-01T02:00:00Z",
			},
		}
		json.NewEncoder(w).Encode(comments)
	}))
	defer server.Close()

	client := NewClient("test-token", "owner/repo", server.URL)

	comments, err := client.FetchComments(1)
	if err != nil {
		t.Fatalf("FetchComments failed: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}

	if comments[0].ID != 1 {
		t.Errorf("Expected comment ID 1, got %d", comments[0].ID)
	}
	if comments[0].Body != "First comment" {
		t.Errorf("Expected body 'First comment', got '%s'", comments[0].Body)
	}
	if comments[0].Author != "commenter1" {
		t.Errorf("Expected author 'commenter1', got '%s'", comments[0].Author)
	}
}

func TestFetchIssuesCancellation(t *testing.T) {
	serverCalled := make(chan bool, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled <- true
		// Delay response to allow cancellation
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient("test-token", "owner/repo", server.URL)
	progress := make(chan int)

	// Create a channel to signal cancellation
	cancelChan := make(chan struct{})
	go func() {
		<-serverCalled // Wait for server to be called
		time.Sleep(10 * time.Millisecond)
		close(cancelChan) // Cancel the request
	}()

	_, err := client.FetchIssues(progress, cancelChan)
	close(progress)

	if err == nil {
		t.Fatal("Expected error for cancelled request, got nil")
	}
}

func TestParseLinkHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		next     string
		last     string
	}{
		{
			name: "next and last links",
			header: `<http://example.com?page=2>; rel="next", <http://example.com?page=10>; rel="last"`,
			next: "http://example.com?page=2",
			last: "http://example.com?page=10",
		},
		{
			name:   "only next link",
			header: `<http://example.com?page=2>; rel="next"`,
			next:   "http://example.com?page=2",
			last:   "",
		},
		{
			name:   "no links",
			header: "",
			next:   "",
			last:   "",
		},
		{
			name:   "malformed links",
			header: "not a link header",
			next:   "",
			last:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, last := parseLinkHeader(tt.header)
			if next != tt.next {
				t.Errorf("Expected next '%s', got '%s'", tt.next, next)
			}
			if last != tt.last {
				t.Errorf("Expected last '%s', got '%s'", tt.last, last)
			}
		})
	}
}

func TestGetEnvToken(t *testing.T) {
	// Save original env value
	original := os.Getenv("GITHUB_TOKEN")
	defer os.Setenv("GITHUB_TOKEN", original)

	// Test with token set
	os.Setenv("GITHUB_TOKEN", "test-token-from-env")
	token := GetEnvToken()
	if token != "test-token-from-env" {
		t.Errorf("Expected token 'test-token-from-env', got '%s'", token)
	}

	// Test without token
	os.Unsetenv("GITHUB_TOKEN")
	token = GetEnvToken()
	if token != "" {
		t.Errorf("Expected empty token, got '%s'", token)
	}
}

func TestFetchIssuesSince(t *testing.T) {
	// Track requests to verify since parameter is used
	var receivedSince string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if since parameter is present
		receivedSince = r.URL.Query().Get("since")
		if receivedSince == "" {
			t.Errorf("Expected 'since' parameter in request")
		}

		// Check authentication
		auth := r.Header.Get("Authorization")
		if auth != "token test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Return updated issue
		issues := []map[string]interface{}{
			{
				"number":      1,
				"title":       "Updated Issue",
				"body":        "Updated issue body",
				"user":        map[string]interface{}{"login": "user1"},
				"state":       "open",
				"created_at":  "2024-01-01T00:00:00Z",
				"updated_at":  "2024-01-05T00:00:00Z",
				"comments":    10,
				"labels":      []map[string]string{{"name": "bug"}},
				"assignees":   []map[string]string{{"login": "assignee1"}},
			},
		}
		json.NewEncoder(w).Encode(issues)
	}))
	defer server.Close()

	client := NewClient("test-token", "owner/repo", server.URL)
	since := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

	progress := make(chan int)
	go func() {
		for range progress {
		}
	}()

	issues, err := client.FetchIssuesSince(since, progress, nil)
	close(progress)

	if err != nil {
		t.Fatalf("FetchIssuesSince failed: %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Number != 1 {
		t.Errorf("Expected issue number 1, got %d", issues[0].Number)
	}

	if issues[0].Title != "Updated Issue" {
		t.Errorf("Expected title 'Updated Issue', got '%s'", issues[0].Title)
	}

	if issues[0].Comments != 10 {
		t.Errorf("Expected 10 comments (updated), got %d", issues[0].Comments)
	}
}

func TestFetchIssuesSinceEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty list (no updates)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer server.Close()

	client := NewClient("test-token", "owner/repo", server.URL)
	since := time.Now()

	progress := make(chan int)
	go func() {
		for range progress {
		}
	}()

	issues, err := client.FetchIssuesSince(since, progress, nil)
	close(progress)

	if err != nil {
		t.Fatalf("FetchIssuesSince failed: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}
