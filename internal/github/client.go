package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is a GitHub API client
type Client struct {
	token string
}

// NewClient creates a new GitHub API client with the given token
func NewClient(token string) *Client {
	return &Client{token: token}
}

// ValidateToken validates the token by making an API call to GitHub
// Returns a helpful error if the token is invalid
func (c *Client) ValidateToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghissues-tui")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf(`invalid GitHub token.

Please check your token and update it:
1. For environment variable: export GITHUB_TOKEN=your_token
2. For config file: run 'ghissues config' to update
3. For GitHub CLI: run 'gh auth refresh'`)

	}

	if resp.StatusCode == http.StatusForbidden {
		// Check if it's a rate limit
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return fmt.Errorf(`GitHub API rate limit exceeded.

Rate limit will reset at: %s

To avoid rate limits:
1. Use a token with higher rate limit (authenticated requests have higher limits)
2. Wait for the rate limit to reset`, resetTime)
		}
		return fmt.Errorf(`GitHub API access denied.

This may be due to:
1. Token lacking required permissions (needs 'repo' scope for private repos)
2. Organization restrictions on token usage`)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Parse response to verify we got valid user data
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to parse GitHub response: %w", err)
	}

	if user.Login == "" {
		return fmt.Errorf("invalid GitHub response: user login not found")
	}

	return nil
}

// User represents a GitHub user response
type User struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ErrInvalidToken is returned when the token validation fails
type ErrInvalidToken struct {
	Message string
}

func (e *ErrInvalidToken) Error() string {
	return e.Message
}

// NewErrInvalidToken creates a new invalid token error with a helpful message
func NewErrInvalidToken() *ErrInvalidToken {
	return &ErrInvalidToken{
		Message: `invalid GitHub token.

Please check your token and update it:
1. For environment variable: export GITHUB_TOKEN=your_token
2. For config file: run 'ghissues config' to update
3. For GitHub CLI: run 'gh auth refresh'`,
	}
}

// Issue represents a GitHub issue
type Issue struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	Author      User      `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Comments    int       `json:"comments"`
	Labels      []Label   `json:"labels"`
	Assignees   []User    `json:"assignees"`
	HTMLURL     string    `json:"html_url"`
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	Author    User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

// Label represents a GitHub label
type Label struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
}

// IssuesResponse represents the paginated issues response
type IssuesResponse struct {
	Items      []Issue `json:"items"`
	TotalCount int     `json:"total_count"`
	NextPage   *int    `json:"next_page,omitempty"`
}

// CommentsResponse represents the paginated comments response
type CommentsResponse struct {
	Items      []Comment `json:"items"`
	TotalCount int       `json:"total_count"`
	NextPage   *int      `json:"next_page,omitempty"`
}

// FetchIssues fetches all open issues from a repository with automatic pagination
// Returns all issues and total count
func (c *Client) FetchIssues(ctx context.Context, owner, repo string) ([]Issue, int, error) {
	var allIssues []Issue
	totalCount := 0
	page := 1

	for {
		select {
		case <-ctx.Done():
			return allIssues, totalCount, ctx.Err()
		default:
		}

		issues, count, nextPage, err := c.fetchIssuesPage(ctx, owner, repo, page)
		if err != nil {
			return allIssues, totalCount, err
		}

		totalCount = count
		allIssues = append(allIssues, issues...)

		if nextPage == nil {
			break
		}
		page = *nextPage
	}

	return allIssues, totalCount, nil
}

// fetchIssuesPage fetches a single page of issues
func (c *Client) fetchIssuesPage(ctx context.Context, owner, repo string, page int) ([]Issue, int, *int, error) {
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("state", "open")
	q.Set("per_page", "100")
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghissues-tui")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, nil, fmt.Errorf("invalid GitHub token")
	}

	if resp.StatusCode == http.StatusForbidden {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return nil, 0, nil, fmt.Errorf("GitHub API rate limit exceeded, resets at %s", resetTime)
		}
		return nil, 0, nil, fmt.Errorf("GitHub API access denied")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to parse issues response: %w", err)
	}

	// Get total count from header
	totalStr := resp.Header.Get("X-Total-Count")
	totalCount := 0
	if totalStr != "" {
		if n, err := strconv.Atoi(totalStr); err == nil {
			totalCount = n
		}
	}

	// Determine next page
	var nextPage *int
	if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
		if next := parseNextPage(linkHeader); next != nil {
			nextPage = next
		}
	}

	return issues, totalCount, nextPage, nil
}

// FetchComments fetches all comments for a specific issue with automatic pagination
func (c *Client) FetchComments(ctx context.Context, owner, repo string, issueNumber int) ([]Comment, error) {
	var allComments []Comment
	page := 1

	for {
		select {
		case <-ctx.Done():
			return allComments, ctx.Err()
		default:
		}

		comments, nextPage, err := c.fetchCommentsPage(ctx, owner, repo, issueNumber, page)
		if err != nil {
			return allComments, err
		}

		allComments = append(allComments, comments...)

		if nextPage == nil {
			break
		}
		page = *nextPage
	}

	return allComments, nil
}

// fetchCommentsPage fetches a single page of comments for an issue
func (c *Client) fetchCommentsPage(ctx context.Context, owner, repo string, issueNumber int, page int) ([]Comment, *int, error) {
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, issueNumber)
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("per_page", "100")
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghissues-tui")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, nil, fmt.Errorf("invalid GitHub token")
	}

	if resp.StatusCode == http.StatusForbidden {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return nil, nil, fmt.Errorf("GitHub API rate limit exceeded, resets at %s", resetTime)
		}
		return nil, nil, fmt.Errorf("GitHub API access denied")
	}

	if resp.StatusCode == http.StatusOK {
		var comments []Comment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			return nil, nil, fmt.Errorf("failed to parse comments response: %w", err)
		}

		// Determine next page
		var nextPage *int
		if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
			if next := parseNextPage(linkHeader); next != nil {
				nextPage = next
			}
		}

		return comments, nextPage, nil
	}

	// 404 means comments are disabled or issue doesn't exist
	body, _ := io.ReadAll(resp.Body)
	return nil, nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

// parseNextPage extracts the next page number from the Link header
// Link header format: <url>; rel="next", <url>; rel="last"
func parseNextPage(linkHeader string) *int {
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) != 2 {
			continue
		}

		urlStr := strings.TrimSpace(parts[0])
		rel := strings.TrimSpace(parts[1])

		if !strings.Contains(rel, `rel="next"`) {
			continue
		}

		// Extract page from URL
		urlStr = strings.Trim(urlStr, "<>")
		u, err := url.Parse(urlStr)
		if err != nil {
			continue
		}

		pageStr := u.Query().Get("page")
		if pageStr == "" {
			continue
		}

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			continue
		}

		return &page
	}
	return nil
}

// FetchIssuesSince fetches all open issues updated since the given timestamp with automatic pagination
// The since parameter should be in RFC3339 format (e.g., "2006-01-02T15:04:05Z07:00")
func (c *Client) FetchIssuesSince(ctx context.Context, owner, repo, since string) ([]Issue, int, error) {
	var allIssues []Issue
	totalCount := 0
	page := 1

	for {
		select {
		case <-ctx.Done():
			return allIssues, totalCount, ctx.Err()
		default:
		}

		issues, count, nextPage, err := c.fetchIssuesPageSince(ctx, owner, repo, page, since)
		if err != nil {
			return allIssues, totalCount, err
		}

		totalCount = count
		allIssues = append(allIssues, issues...)

		if nextPage == nil {
			break
		}
		page = *nextPage
	}

	return allIssues, totalCount, nil
}

// fetchIssuesPageSince fetches a single page of issues with since filter
func (c *Client) fetchIssuesPageSince(ctx context.Context, owner, repo string, page int, since string) ([]Issue, int, *int, error) {
	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	q.Set("state", "open")
	q.Set("per_page", "100")
	q.Set("page", strconv.Itoa(page))
	if since != "" {
		q.Set("since", since)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghissues-tui")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, nil, fmt.Errorf("invalid GitHub token")
	}

	if resp.StatusCode == http.StatusForbidden {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			return nil, 0, nil, fmt.Errorf("GitHub API rate limit exceeded, resets at %s", resetTime)
		}
		return nil, 0, nil, fmt.Errorf("GitHub API access denied")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, 0, nil, fmt.Errorf("failed to parse issues response: %w", err)
	}

	// Get total count from header
	totalStr := resp.Header.Get("X-Total-Count")
	totalCount := 0
	if totalStr != "" {
		if n, err := strconv.Atoi(totalStr); err == nil {
			totalCount = n
		}
	}

	// Determine next page
	var nextPage *int
	if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
		if next := parseNextPage(linkHeader); next != nil {
			nextPage = next
		}
	}

	return issues, totalCount, nextPage, nil
}