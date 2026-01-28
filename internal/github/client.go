package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shepbook/ghissues/internal/database"
)

const (
	GitHubAPIBase = "https://api.github.com"
	PerPage       = 100 // Maximum allowed by GitHub API
)

// Client handles GitHub API requests
type Client struct {
	token   string
	client  *http.Client
	BaseURL string
}

// FetchProgress represents the progress of fetching issues
type FetchProgress struct {
	Fetched int
	Total   int
	Current string // Current operation (e.g., "Fetching page 3")
}

// GitHubIssue represents the GitHub API response for an issue
type GitHubIssue struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	ClosedAt  string `json:"closed_at"`
	Comments  int    `json:"comments"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
}

// GitHubComment represents the GitHub API response for a comment
type GitHubComment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	return &Client{
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
		BaseURL: GitHubAPIBase,
	}
}

// FetchIssues fetches all open issues from a repository
// Supports cancellation through the cancel channel
func (c *Client) FetchIssues(repo string, progress chan<- FetchProgress) ([]database.Issue, error) {
	return c.FetchIssuesSince(repo, "", progress)
}

// FetchIssuesSince fetches issues updated since a given timestamp
// If since is empty, fetches all open issues
func (c *Client) FetchIssuesSince(repo string, since string, progress chan<- FetchProgress) ([]database.Issue, error) {
	owner, name, err := ParseGitHubRepoURL(repo)
	if err != nil {
		return nil, err
	}

	var allIssues []database.Issue
	page := 1

	for {
		// Update progress
		if progress != nil {
			progress <- FetchProgress{
				Fetched: len(allIssues),
				Total:   -1, // Unknown until we fetch
				Current: fmt.Sprintf("Fetching issues page %d", page),
			}
		}

		// Build URL with optional since parameter
		urlStr := fmt.Sprintf("%s/repos/%s/%s/issues?state=open&per_page=%d&page=%d",
			c.BaseURL, owner, name, PerPage, page)
		if since != "" {
			// URL encode the timestamp
			urlStr = urlStr + "&since=" + url.QueryEscape(since)
		}

		issues, hasMore, err := c.fetchIssuesPage(urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		allIssues = append(allIssues, issues...)

		// Update progress with accurate count
		if progress != nil {
			progress <- FetchProgress{
				Fetched: len(allIssues),
				Total:   len(allIssues), // At least this many
				Current: fmt.Sprintf("Fetched %d issues", len(allIssues)),
			}
		}

		if !hasMore {
			break
		}

		page++

		// Rate limiting - be respectful to the API
		time.Sleep(100 * time.Millisecond)
	}

	// Final progress update
	if progress != nil {
		progress <- FetchProgress{
			Fetched: len(allIssues),
			Total:   len(allIssues),
			Current: fmt.Sprintf("Fetched all %d issues", len(allIssues)),
		}
	}

	return allIssues, nil
}

// fetchIssuesPage fetches a single page of issues
func (c *Client) fetchIssuesPage(url string) ([]database.Issue, bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, false, fmt.Errorf("API error: %s (status %d)", errResp.Message, resp.StatusCode)
	}

	var ghIssues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&ghIssues); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for next page in Link header
	hasMore := hasNextPage(resp.Header.Get("Link"))

	// Convert to our Issue type
	issues := make([]database.Issue, len(ghIssues))
	for i, gi := range ghIssues {
		issues[i] = convertGitHubIssue(gi)
	}

	return issues, hasMore, nil
}

// FetchComments fetches all comments for a specific issue
func (c *Client) FetchComments(repo string, issueNumber int, progress chan<- string) ([]database.Comment, error) {
	owner, name, err := ParseGitHubRepoURL(repo)
	if err != nil {
		return nil, err
	}

	var allComments []database.Comment
	page := 1

	for {
		if progress != nil {
			progress <- fmt.Sprintf("Fetching comments for issue #%d (page %d)", issueNumber, page)
		}

		url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?per_page=%d&page=%d",
			c.BaseURL, owner, name, issueNumber, PerPage, page)

		comments, hasMore, err := c.fetchCommentsPage(url, issueNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments for issue #%d: %w", issueNumber, err)
		}

		allComments = append(allComments, comments...)

		if !hasMore {
			break
		}

		page++
		time.Sleep(50 * time.Millisecond)
	}

	return allComments, nil
}

// fetchCommentsPage fetches a single page of comments
func (c *Client) fetchCommentsPage(url string, issueNumber int) ([]database.Comment, bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, false, fmt.Errorf("API error: %s (status %d)", errResp.Message, resp.StatusCode)
	}

	var ghComments []GitHubComment
	if err := json.NewDecoder(resp.Body).Decode(&ghComments); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	hasMore := hasNextPage(resp.Header.Get("Link"))

	comments := make([]database.Comment, len(ghComments))
	for i, gc := range ghComments {
		comments[i] = database.Comment{
			ID:          gc.ID,
			IssueNumber: issueNumber,
			Body:        gc.Body,
			Author:      gc.User.Login,
			CreatedAt:   gc.CreatedAt,
			UpdatedAt:   gc.UpdatedAt,
		}
	}

	return comments, hasMore, nil
}

// hasNextPage checks if there's a next page in the Link header
func hasNextPage(linkHeader string) bool {
	if linkHeader == "" {
		return false
	}

	// Parse the Link header: <url>; rel="next", <url>; rel="last"
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) >= 2 {
			rel := strings.TrimSpace(parts[1])
			if strings.Contains(rel, `rel="next"`) {
				return true
			}
		}
	}
	return false
}

// convertGitHubIssue converts a GitHubIssue to an Issue
func convertGitHubIssue(gi GitHubIssue) database.Issue {
	labels := make([]string, len(gi.Labels))
	for i, l := range gi.Labels {
		labels[i] = l.Name
	}

	assignees := make([]string, len(gi.Assignees))
	for i, a := range gi.Assignees {
		assignees[i] = a.Login
	}

	return database.Issue{
		Number:       gi.Number,
		Title:        gi.Title,
		Body:         gi.Body,
		State:        gi.State,
		Author:       gi.User.Login,
		CreatedAt:    gi.CreatedAt,
		UpdatedAt:    gi.UpdatedAt,
		ClosedAt:     gi.ClosedAt,
		CommentCount: gi.Comments,
		Labels:       labels,
		Assignees:    assignees,
	}
}

// ParseGitHubRepoURL parses an "owner/repo" string into its components
func ParseGitHubRepoURL(repo string) (owner, name string, err error) {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return "", "", fmt.Errorf("repository cannot be empty")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("repository must be in format 'owner/repo'")
	}

	owner = strings.TrimSpace(parts[0])
	name = strings.TrimSpace(parts[1])

	if owner == "" {
		return "", "", fmt.Errorf("repository owner cannot be empty")
	}

	if name == "" {
		return "", "", fmt.Errorf("repository name cannot be empty")
	}

	// Validate characters (alphanumeric, hyphens, underscores)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(owner) {
		return "", "", fmt.Errorf("repository owner contains invalid characters")
	}
	if !validPattern.MatchString(name) {
		return "", "", fmt.Errorf("repository name contains invalid characters")
	}

	return owner, name, nil
}

// IssueCount fetches the total count of open issues from the API
// This is used to show accurate progress
func (c *Client) IssueCount(repo string) (int, error) {
	owner, name, err := ParseGitHubRepoURL(repo)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s?per_page=1", c.BaseURL, owner, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error (status %d)", resp.StatusCode)
	}

	var repoInfo struct {
		OpenIssuesCount int `json:"open_issues_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return repoInfo.OpenIssuesCount, nil
}

// sanitizeURL creates a safe URL string for display (removes tokens)
func sanitizeURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	// Remove token from query string if present
	q := u.Query()
	q.Del("token")
	q.Del("access_token")
	u.RawQuery = q.Encode()

	return u.String()
}

// parseLastPage extracts the last page number from Link header
func parseLastPage(linkHeader string) int {
	if linkHeader == "" {
		return 1
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		parts := strings.Split(link, ";")
		if len(parts) >= 2 {
			rel := strings.TrimSpace(parts[1])
			if strings.Contains(rel, `rel="last"`) {
				// Extract page number from URL
				urlPart := strings.TrimSpace(parts[0])
				urlPart = strings.TrimPrefix(urlPart, "<")
				urlPart = strings.TrimSuffix(urlPart, ">")

				u, err := url.Parse(urlPart)
				if err != nil {
					continue
				}

				pageStr := u.Query().Get("page")
				if pageStr != "" {
					page, _ := strconv.Atoi(pageStr)
					if page > 0 {
						return page
					}
				}
			}
		}
	}
	return 1
}
