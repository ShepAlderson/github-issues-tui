package github

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIssue_Struct(t *testing.T) {
	now := time.Now()
	issue := Issue{
		Number:      123,
		Title:       "Test Issue",
		Body:        "This is a test issue body",
		State:       "open",
		Author:      User{Login: "testuser", ID: 12345},
		CreatedAt:   now,
		UpdatedAt:   now,
		Comments:    5,
		Labels:      []Label{{Name: "bug"}, {Name: "priority"}},
		Assignees:   []User{{Login: "assignee1"}, {Login: "assignee2"}},
		HTMLURL:     "https://github.com/owner/repo/issues/123",
	}

	if issue.Number != 123 {
		t.Errorf("Issue.Number = %d, want %d", issue.Number, 123)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Issue.Title = %q, want %q", issue.Title, "Test Issue")
	}
	if issue.State != "open" {
		t.Errorf("Issue.State = %q, want %q", issue.State, "open")
	}
	if issue.Author.Login != "testuser" {
		t.Errorf("Issue.Author.Login = %q, want %q", issue.Author.Login, "testuser")
	}
	if issue.Comments != 5 {
		t.Errorf("Issue.Comments = %d, want %d", issue.Comments, 5)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Issue.Labels length = %d, want %d", len(issue.Labels), 2)
	}
	if len(issue.Assignees) != 2 {
		t.Errorf("Issue.Assignees length = %d, want %d", len(issue.Assignees), 2)
	}
}

func TestLabel_Struct(t *testing.T) {
	label := Label{
		ID:     12345,
		Name:   "bug",
		Color:  "ff0000",
	}

	if label.Name != "bug" {
		t.Errorf("Label.Name = %q, want %q", label.Name, "bug")
	}
	if label.Color != "ff0000" {
		t.Errorf("Label.Color = %q, want %q", label.Color, "ff0000")
	}
}

func TestComment_Struct(t *testing.T) {
	now := time.Now()
	comment := Comment{
		ID:        12345,
		Body:      "This is a test comment",
		Author:    User{Login: "commenter", ID: 67890},
		CreatedAt: now,
	}

	if comment.ID != 12345 {
		t.Errorf("Comment.ID = %d, want %d", comment.ID, 12345)
	}
	if comment.Body != "This is a test comment" {
		t.Errorf("Comment.Body = %q, want %q", comment.Body, "This is a test comment")
	}
	if comment.Author.Login != "commenter" {
		t.Errorf("Comment.Author.Login = %q, want %q", comment.Author.Login, "commenter")
	}
}

func TestIssue_JSON(t *testing.T) {
	jsonData := `{
		"number": 123,
		"title": "Test Issue",
		"body": "This is a test issue body",
		"state": "open",
		"user": {"login": "testuser", "id": 12345},
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-02T00:00:00Z",
		"comments": 5,
		"labels": [{"id": 1, "name": "bug", "color": "ff0000"}],
		"assignees": [{"login": "assignee1", "id": 2}],
		"html_url": "https://github.com/owner/repo/issues/123"
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(jsonData), &issue); err != nil {
		t.Fatalf("failed to unmarshal Issue: %v", err)
	}

	if issue.Number != 123 {
		t.Errorf("Issue.Number = %d, want %d", issue.Number, 123)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Issue.Title = %q, want %q", issue.Title, "Test Issue")
	}
	if issue.Body != "This is a test issue body" {
		t.Errorf("Issue.Body = %q, want %q", issue.Body, "This is a test issue body")
	}
	if issue.Author.Login != "testuser" {
		t.Errorf("Issue.Author.Login = %q, want %q", issue.Author.Login, "testuser")
	}
	if issue.Comments != 5 {
		t.Errorf("Issue.Comments = %d, want %d", issue.Comments, 5)
	}
	if len(issue.Labels) != 1 {
		t.Errorf("Issue.Labels length = %d, want %d", len(issue.Labels), 1)
	}
	if issue.Labels[0].Name != "bug" {
		t.Errorf("Issue.Labels[0].Name = %q, want %q", issue.Labels[0].Name, "bug")
	}
}

func TestComment_JSON(t *testing.T) {
	jsonData := `{
		"id": 12345,
		"body": "This is a test comment",
		"user": {"login": "commenter", "id": 67890},
		"created_at": "2024-01-01T00:00:00Z"
	}`

	var comment Comment
	if err := json.Unmarshal([]byte(jsonData), &comment); err != nil {
		t.Fatalf("failed to unmarshal Comment: %v", err)
	}

	if comment.ID != 12345 {
		t.Errorf("Comment.ID = %d, want %d", comment.ID, 12345)
	}
	if comment.Body != "This is a test comment" {
		t.Errorf("Comment.Body = %q, want %q", comment.Body, "This is a test comment")
	}
	if comment.Author.Login != "commenter" {
		t.Errorf("Comment.Author.Login = %q, want %q", comment.Author.Login, "commenter")
	}
}

func TestIssuesResponse_JSON(t *testing.T) {
	jsonData := `{
		"items": [
			{"number": 1, "title": "Issue 1", "body": "Body 1", "state": "open", "user": {"login": "user1"}, "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "comments": 0, "labels": [], "assignees": []},
			{"number": 2, "title": "Issue 2", "body": "Body 2", "state": "open", "user": {"login": "user2"}, "created_at": "2024-01-02T00:00:00Z", "updated_at": "2024-01-02T00:00:00Z", "comments": 3, "labels": [{"name": "feature"}], "assignees": []}
		],
		"total_count": 2,
		"next_page": null
	}`

	var resp IssuesResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal IssuesResponse: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("IssuesResponse.Items length = %d, want %d", len(resp.Items), 2)
	}
	if resp.TotalCount != 2 {
		t.Errorf("IssuesResponse.TotalCount = %d, want %d", resp.TotalCount, 2)
	}
	if resp.NextPage != nil {
		t.Errorf("IssuesResponse.NextPage = %v, want nil", resp.NextPage)
	}
}

func TestIssuesResponse_WithPagination(t *testing.T) {
	jsonData := `{
		"items": [{"number": 1, "title": "Issue 1", "body": "Body 1", "state": "open", "user": {"login": "user1"}, "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "comments": 0, "labels": [], "assignees": []}],
		"total_count": 100,
		"next_page": 2
	}`

	var resp IssuesResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal IssuesResponse: %v", err)
	}

	if resp.TotalCount != 100 {
		t.Errorf("IssuesResponse.TotalCount = %d, want %d", resp.TotalCount, 100)
	}
	if resp.NextPage == nil || *resp.NextPage != 2 {
		t.Errorf("IssuesResponse.NextPage = %v, want 2", resp.NextPage)
	}
}