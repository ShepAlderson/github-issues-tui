package sync

import "time"

// Issue represents a GitHub issue
type Issue struct {
	Number       int
	Title        string
	Body         string
	State        string
	Author       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CommentCount int
	Labels       []string
	Assignees    []string
}

// Comment represents a comment on an issue
type Comment struct {
	ID          int64
	IssueNumber int
	Body        string
	Author      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GitHubIssueResponse represents the JSON structure from GitHub API
type GitHubIssueResponse struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Comments  int       `json:"comments"`
	Labels    []Label   `json:"labels"`
	Assignees []User    `json:"assignees"`
}

// GitHubCommentResponse represents the JSON structure for comments from GitHub API
type GitHubCommentResponse struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// SyncProgress represents the current state of a sync operation
type SyncProgress struct {
	TotalIssues    int
	FetchedIssues  int
	TotalComments  int
	FetchedComments int
}
