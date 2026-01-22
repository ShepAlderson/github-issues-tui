package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")
	assert.NotNil(t, client)
}

func TestClient_FetchIssues(t *testing.T) {
	tests := []struct {
		name           string
		responses      []issuesResponse
		wantIssues     int
		wantErr        bool
		wantErrMessage string
	}{
		{
			name: "single page of issues",
			responses: []issuesResponse{
				{
					issues: []Issue{
						{Number: 1, Title: "Issue 1"},
						{Number: 2, Title: "Issue 2"},
					},
					hasNextPage: false,
				},
			},
			wantIssues: 2,
		},
		{
			name: "multiple pages of issues",
			responses: []issuesResponse{
				{
					issues: []Issue{
						{Number: 1, Title: "Issue 1"},
						{Number: 2, Title: "Issue 2"},
					},
					hasNextPage: true,
					endCursor:   "cursor1",
				},
				{
					issues: []Issue{
						{Number: 3, Title: "Issue 3"},
					},
					hasNextPage: false,
				},
			},
			wantIssues: 3,
		},
		{
			name: "empty repository",
			responses: []issuesResponse{
				{
					issues:      []Issue{},
					hasNextPage: false,
				},
			},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageIdx := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				if pageIdx >= len(tt.responses) {
					t.Fatalf("unexpected request for page %d", pageIdx)
				}

				resp := tt.responses[pageIdx]
				pageIdx++

				writeGraphQLResponse(t, w, resp)
			}))
			defer server.Close()

			client := NewClient("test-token")
			client.baseURL = server.URL

			ctx := context.Background()
			issues, err := client.FetchIssues(ctx, "owner", "repo", nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMessage != "" {
					assert.Contains(t, err.Error(), tt.wantErrMessage)
				}
				return
			}

			require.NoError(t, err)
			assert.Len(t, issues, tt.wantIssues)
		})
	}
}

func TestClient_FetchIssues_Progress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeGraphQLResponse(t, w, issuesResponse{
			issues: []Issue{
				{Number: 1, Title: "Issue 1"},
				{Number: 2, Title: "Issue 2"},
			},
			hasNextPage: false,
			totalCount:  2,
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	ctx := context.Background()
	var progressCalls []FetchProgress

	issues, err := client.FetchIssues(ctx, "owner", "repo", func(p FetchProgress) {
		progressCalls = append(progressCalls, p)
	})

	require.NoError(t, err)
	assert.Len(t, issues, 2)
	assert.NotEmpty(t, progressCalls)
	assert.Equal(t, 2, progressCalls[len(progressCalls)-1].Total)
}

func TestClient_FetchIssues_Cancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if context is cancelled
		if r.Context().Err() != nil {
			return
		}
		writeGraphQLResponse(t, w, issuesResponse{
			issues: []Issue{
				{Number: 1, Title: "Issue 1"},
			},
			hasNextPage: true,
			endCursor:   "cursor1",
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.FetchIssues(ctx, "owner", "repo", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestClient_FetchIssueComments(t *testing.T) {
	tests := []struct {
		name         string
		responses    []commentsResponse
		wantComments int
		wantErr      bool
	}{
		{
			name: "single page of comments",
			responses: []commentsResponse{
				{
					comments: []Comment{
						{ID: "1", Body: "Comment 1"},
						{ID: "2", Body: "Comment 2"},
					},
					hasNextPage: false,
				},
			},
			wantComments: 2,
		},
		{
			name: "multiple pages of comments",
			responses: []commentsResponse{
				{
					comments: []Comment{
						{ID: "1", Body: "Comment 1"},
					},
					hasNextPage: true,
					endCursor:   "cursor1",
				},
				{
					comments: []Comment{
						{ID: "2", Body: "Comment 2"},
					},
					hasNextPage: false,
				},
			},
			wantComments: 2,
		},
		{
			name: "no comments",
			responses: []commentsResponse{
				{
					comments:    []Comment{},
					hasNextPage: false,
				},
			},
			wantComments: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageIdx := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if pageIdx >= len(tt.responses) {
					t.Fatalf("unexpected request for page %d", pageIdx)
				}

				resp := tt.responses[pageIdx]
				pageIdx++

				writeCommentsGraphQLResponse(t, w, resp)
			}))
			defer server.Close()

			client := NewClient("test-token")
			client.baseURL = server.URL

			ctx := context.Background()
			comments, err := client.FetchIssueComments(ctx, "owner", "repo", 1)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, comments, tt.wantComments)
		})
	}
}

func TestIssue_ParsedDates(t *testing.T) {
	issue := Issue{
		Number:    1,
		Title:     "Test Issue",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-16T12:00:00Z",
	}

	created, err := issue.CreatedAtTime()
	require.NoError(t, err)
	assert.Equal(t, 2024, created.Year())
	assert.Equal(t, time.January, created.Month())
	assert.Equal(t, 15, created.Day())

	updated, err := issue.UpdatedAtTime()
	require.NoError(t, err)
	assert.Equal(t, 2024, updated.Year())
	assert.Equal(t, time.January, updated.Month())
	assert.Equal(t, 16, updated.Day())
}

// Test helpers

type issuesResponse struct {
	issues      []Issue
	hasNextPage bool
	endCursor   string
	totalCount  int
}

type commentsResponse struct {
	comments    []Comment
	hasNextPage bool
	endCursor   string
}

func writeGraphQLResponse(t *testing.T, w http.ResponseWriter, resp issuesResponse) {
	t.Helper()

	// Build nodes from issues
	nodes := make([]map[string]interface{}, len(resp.issues))
	for i, issue := range resp.issues {
		nodes[i] = map[string]interface{}{
			"number":    issue.Number,
			"title":     issue.Title,
			"body":      issue.Body,
			"createdAt": issue.CreatedAt,
			"updatedAt": issue.UpdatedAt,
			"author": map[string]interface{}{
				"login": issue.Author.Login,
			},
			"labels": map[string]interface{}{
				"nodes": []map[string]interface{}{},
			},
			"assignees": map[string]interface{}{
				"nodes": []map[string]interface{}{},
			},
			"comments": map[string]interface{}{
				"totalCount": issue.CommentCount,
			},
		}
	}

	totalCount := resp.totalCount
	if totalCount == 0 {
		totalCount = len(resp.issues)
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"repository": map[string]interface{}{
				"issues": map[string]interface{}{
					"totalCount": totalCount,
					"pageInfo": map[string]interface{}{
						"hasNextPage": resp.hasNextPage,
						"endCursor":   resp.endCursor,
					},
					"nodes": nodes,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func writeCommentsGraphQLResponse(t *testing.T, w http.ResponseWriter, resp commentsResponse) {
	t.Helper()

	nodes := make([]map[string]interface{}, len(resp.comments))
	for i, comment := range resp.comments {
		nodes[i] = map[string]interface{}{
			"id":        comment.ID,
			"body":      comment.Body,
			"createdAt": comment.CreatedAt,
			"author": map[string]interface{}{
				"login": comment.Author.Login,
			},
		}
	}

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"repository": map[string]interface{}{
				"issue": map[string]interface{}{
					"comments": map[string]interface{}{
						"pageInfo": map[string]interface{}{
							"hasNextPage": resp.hasNextPage,
							"endCursor":   resp.endCursor,
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
