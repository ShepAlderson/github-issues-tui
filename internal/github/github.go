package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shepbook/ghissues/internal/storage"
)

const (
	defaultGitHubAPIURL = "https://api.github.com"
)

// Issue represents a GitHub issue (extended from storage.Issue with API-specific fields)
type Issue struct {
	Number    int
	Title     string
	Body      string
	Author    string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	Comments  int
	Labels    string // Comma-separated label names
	Assignees string // Comma-separated assignee usernames
}

// Comment represents a GitHub issue comment (extended from storage.Comment)
type Comment struct {
	ID          int
	IssueNumber int
	Body        string
	Author      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Client represents a GitHub API client
type Client struct {
	token      string
	repo       string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client
func NewClient(token, repo, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultGitHubAPIURL
	}

	return &Client{
		token:   token,
		repo:    repo,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchIssues fetches all open issues for the repository, handling pagination
// It sends progress updates on the progress channel and can be cancelled via cancelChan
func (c *Client) FetchIssues(progress chan<- int, cancelChan <-chan struct{}) ([]storage.Issue, error) {
	var allIssues []storage.Issue
	url := fmt.Sprintf("%s/repos/%s/issues?state=open&per_page=100", c.baseURL, c.repo)
	page := 1

	for url != "" {
		// Check for cancellation
		select {
		case <-cancelChan:
			return nil, fmt.Errorf("sync cancelled by user")
		default:
		}

		// Create request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Make request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}

		// Check response status
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			return nil, fmt.Errorf("authentication failed: invalid GitHub token")
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("repository not found: %s", c.repo)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Parse response
		var apiIssues []struct {
			Number      int       `json:"number"`
			Title       string    `json:"title"`
			Body        string    `json:"body"`
			User        struct {
				Login string `json:"login"`
			} `json:"user"`
			State      string    `json:"state"`
			CreatedAt  time.Time `json:"created_at"`
			UpdatedAt  time.Time `json:"updated_at"`
			ClosedAt   *time.Time `json:"closed_at"`
			Comments   int       `json:"comments"`
			Labels     []struct {
				Name string `json:"name"`
			} `json:"labels"`
			Assignees []struct {
				Login string `json:"login"`
			} `json:"assignees"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiIssues); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		// Convert to storage.Issue
		for _, apiIssue := range apiIssues {
			labelNames := make([]string, len(apiIssue.Labels))
			for i, label := range apiIssue.Labels {
				labelNames[i] = label.Name
			}

			assigneeLogins := make([]string, len(apiIssue.Assignees))
			for i, assignee := range apiIssue.Assignees {
				assigneeLogins[i] = assignee.Login
			}

			issue := storage.Issue{
				Number:    apiIssue.Number,
				Title:     apiIssue.Title,
				Body:      apiIssue.Body,
				Author:    apiIssue.User.Login,
				State:     apiIssue.State,
				CreatedAt: apiIssue.CreatedAt,
				UpdatedAt: apiIssue.UpdatedAt,
				ClosedAt:  apiIssue.ClosedAt,
				Comments:  apiIssue.Comments,
				Labels:    strings.Join(labelNames, ","),
				Assignees: strings.Join(assigneeLogins, ","),
			}

			allIssues = append(allIssues, issue)
		}

		// Send progress update
		progress <- len(allIssues)

		// Get next page from Link header
		linkHeader := resp.Header.Get("Link")
		_, url = parseLinkHeader(linkHeader)
		page++
	}

	return allIssues, nil
}

// FetchComments fetches all comments for a specific issue
func (c *Client) FetchComments(issueNumber int) ([]storage.Comment, error) {
	url := fmt.Sprintf("%s/repos/%s/issues/%d/comments?per_page=100", c.baseURL, c.repo, issueNumber)
	var allComments []storage.Comment

	for url != "" {
		// Create request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Make request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments: %w", err)
		}

		// Check response status
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			return nil, fmt.Errorf("authentication failed: invalid GitHub token")
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// Parse response
		var apiComments []struct {
			ID        int       `json:"id"`
			Body      string    `json:"body"`
			User      struct {
				Login string `json:"login"`
			} `json:"user"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiComments); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		// Convert to storage.Comment
		for _, apiComment := range apiComments {
			comment := storage.Comment{
				ID:          apiComment.ID,
				IssueNumber: issueNumber,
				Body:        apiComment.Body,
				Author:      apiComment.User.Login,
				CreatedAt:   apiComment.CreatedAt,
				UpdatedAt:   apiComment.UpdatedAt,
			}

			allComments = append(allComments, comment)
		}

		// Get next page from Link header
		linkHeader := resp.Header.Get("Link")
		_, url = parseLinkHeader(linkHeader)
	}

	return allComments, nil
}

// parseLinkHeader parses GitHub's Link header and returns the next and last URLs
func parseLinkHeader(header string) (next, last string) {
	if header == "" {
		return "", ""
	}

	// Parse links using regex
	linkRegex := regexp.MustCompile(`<([^>]+)>;\s*rel="(\w+)"`)
	matches := linkRegex.FindAllStringSubmatch(header, -1)

	for _, match := range matches {
		if len(match) == 3 {
			url := match[1]
			rel := match[2]

			switch rel {
			case "next":
				next = url
			case "last":
				last = url
			}
		}
	}

	return next, last
}

// ValidateRepo validates that a repository string is in the correct format (owner/repo)
func ValidateRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	// Check format: owner/repo
	repoRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)
	if !repoRegex.MatchString(repo) {
		return fmt.Errorf("repository must be in 'owner/repo' format")
	}

	return nil
}

// GetEnvToken returns the GitHub token from the environment variable
func GetEnvToken() string {
	return os.Getenv("GITHUB_TOKEN")
}
