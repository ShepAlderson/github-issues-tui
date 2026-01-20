package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client represents a GitHub API client
type Client struct {
	token   string
	baseURL string
	client  *http.Client
}

// Issue represents a GitHub issue response
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	User      User      `json:"user"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  int       `json:"comments"`
	Labels    []Label   `json:"labels"`
	Assignees []User    `json:"assignees"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	User      User   `json:"user"`
	CreatedAt string `json:"created_at"`
}

// NewClient creates a new GitHub API client
func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}

	return &Client{
		token:   token,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// FetchOpenIssues fetches all open issues from a repository with pagination
func (c *Client) FetchOpenIssues(repo string) ([]Issue, error) {
	var allIssues []Issue
	page := 1

	for {
		issues, hasNext, err := c.fetchIssuesPage(repo, page)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues page %d: %w", page, err)
		}

		allIssues = append(allIssues, issues...)

		if !hasNext || len(issues) == 0 {
			break
		}

		page++
	}

	return allIssues, nil
}

// fetchIssuesPage fetches a single page of issues
func (c *Client) fetchIssuesPage(repo string, page int) ([]Issue, bool, error) {
	return c.fetchIssuesPageWithSince(repo, page, "")
}

// fetchIssuesPageWithSince fetches a single page of issues with optional since parameter
func (c *Client) fetchIssuesPageWithSince(repo string, page int, since string) ([]Issue, bool, error) {
	owner, name, err := getRepo(repo)
	if err != nil {
		return nil, false, fmt.Errorf("invalid repository format: %w", err)
	}

	// Build URL
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, name)
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	query := u.Query()
	query.Set("state", "open")
	query.Set("per_page", "100")
	query.Set("page", fmt.Sprintf("%d", page))
	if since != "" {
		query.Set("since", since)
	}
	u.RawQuery = query.Encode()

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for next page
	hasNext := hasNextPage(resp.Header.Get("Link"))

	return issues, hasNext, nil
}

// FetchIssueComments fetches all comments for an issue
func (c *Client) FetchIssueComments(repo string, issueNum int) ([]Comment, error) {
	owner, name, err := getRepo(repo)
	if err != nil {
		return nil, fmt.Errorf("invalid repository format: %w", err)
	}

	// Build URL
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, name, issueNum)
	u := c.baseURL + path

	// Make request
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var comments []Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return comments, nil
}

// hasNextPage checks if there's a next page based on Link header
func hasNextPage(linkHeader string) bool {
	// Parse Link header to check for next page
	// Format: <url>; rel="next", <url>; rel="last"
	if linkHeader == "" {
		return false
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) >= 2 {
			rel := strings.TrimSpace(parts[1])
			if strings.HasPrefix(rel, `rel="next"`) {
				return true
			}
		}
	}

	return false
}

// getRepo parses owner/repo format
func getRepo(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository format: expected owner/repo, got %s", repo)
	}

	owner = strings.TrimSpace(parts[0])
	name = strings.TrimSpace(parts[1])

	if owner == "" || name == "" {
		return "", "", fmt.Errorf("invalid repository format: owner and name cannot be empty")
	}

	return owner, name, nil
}

// FetchIssuesSince fetches issues updated since the given timestamp
func (c *Client) FetchIssuesSince(repo string, since string) ([]Issue, error) {
	var allIssues []Issue
	page := 1

	for {
		issues, hasNext, err := c.fetchIssuesPageWithSince(repo, page, since)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues page %d: %w", page, err)
		}

		allIssues = append(allIssues, issues...)

		if !hasNext || len(issues) == 0 {
			break
		}

		page++
	}

	return allIssues, nil
}

// FetchIssueCommentsSince fetches comments created since the given timestamp
func (c *Client) FetchIssueCommentsSince(repo string, issueNum int, since string) ([]Comment, error) {
	owner, name, err := getRepo(repo)
	if err != nil {
		return nil, fmt.Errorf("invalid repository format: %w", err)
	}

	// Build URL
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, name, issueNum)
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add since parameter if provided
	query := u.Query()
	if since != "" {
		query.Set("since", since)
	}
	u.RawQuery = query.Encode()

	// Make request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var comments []Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return comments, nil
}
