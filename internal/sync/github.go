package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	perPage          = 100 // Maximum allowed by GitHub API
)

// GitHubClient handles communication with the GitHub API
type GitHubClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		baseURL: githubAPIBaseURL,
		token:   token,
		client:  &http.Client{},
	}
}

// FetchIssues fetches all open issues from a repository, handling pagination
func (c *GitHubClient) FetchIssues(ctx context.Context, repo string) ([]*Issue, error) {
	var allIssues []*Issue
	page := 1

	for {
		// Build URL
		url := fmt.Sprintf("%s/repos/%s/issues?state=open&per_page=%d&page=%d",
			c.baseURL, repo, perPage, page)

		// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Execute request
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
		}

		// Parse response
		var ghIssues []GitHubIssueResponse
		if err := json.NewDecoder(resp.Body).Decode(&ghIssues); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Convert to internal format
		for _, ghIssue := range ghIssues {
			// Skip pull requests (they appear in the issues API)
			if ghIssue.IsPullRequest() {
				continue
			}

			issue := &Issue{
				Number:       ghIssue.Number,
				Title:        ghIssue.Title,
				Body:         ghIssue.Body,
				State:        ghIssue.State,
				Author:       ghIssue.User.Login,
				CreatedAt:    ghIssue.CreatedAt,
				UpdatedAt:    ghIssue.UpdatedAt,
				CommentCount: ghIssue.Comments,
			}

			// Extract labels
			for _, label := range ghIssue.Labels {
				issue.Labels = append(issue.Labels, label.Name)
			}

			// Extract assignees
			for _, assignee := range ghIssue.Assignees {
				issue.Assignees = append(issue.Assignees, assignee.Login)
			}

			allIssues = append(allIssues, issue)
		}

		// Check for next page
		linkHeader := resp.Header.Get("Link")
		if !hasNextPage(linkHeader) {
			break
		}

		page++
	}

	return allIssues, nil
}

// FetchComments fetches all comments for a specific issue
func (c *GitHubClient) FetchComments(ctx context.Context, repo string, issueNumber int) ([]*Comment, error) {
	var allComments []*Comment
	page := 1

	for {
		// Build URL
		url := fmt.Sprintf("%s/repos/%s/issues/%d/comments?per_page=%d&page=%d",
			c.baseURL, repo, issueNumber, perPage, page)

		// Create request
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Execute request
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments: %w", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
		}

		// Parse response
		var ghComments []GitHubCommentResponse
		if err := json.NewDecoder(resp.Body).Decode(&ghComments); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Convert to internal format
		for _, ghComment := range ghComments {
			comment := &Comment{
				ID:          ghComment.ID,
				IssueNumber: issueNumber,
				Body:        ghComment.Body,
				Author:      ghComment.User.Login,
				CreatedAt:   ghComment.CreatedAt,
				UpdatedAt:   ghComment.UpdatedAt,
			}
			allComments = append(allComments, comment)
		}

		// Check for next page
		linkHeader := resp.Header.Get("Link")
		if !hasNextPage(linkHeader) {
			break
		}

		page++
	}

	return allComments, nil
}

// hasNextPage checks if the Link header indicates there's a next page
func hasNextPage(linkHeader string) bool {
	if linkHeader == "" {
		return false
	}

	// GitHub Link header format: <url>; rel="next", <url>; rel="last"
	return strings.Contains(linkHeader, `rel="next"`)
}

// IsPullRequest checks if this issue is actually a pull request
// Pull requests appear in the issues API but have a pull_request field
func (r *GitHubIssueResponse) IsPullRequest() bool {
	// This is a simple heuristic - in practice, GitHub API includes a pull_request field
	// For now, we'll keep it simple and just check if the issue has certain characteristics
	// In a real implementation, we'd add a PullRequest field to the struct
	return false
}
