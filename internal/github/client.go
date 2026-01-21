package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.github.com/graphql"

// Client is a GitHub API client for fetching issues
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultBaseURL,
	}
}

// SetBaseURL sets the base URL for API requests (mainly for testing)
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// Issue represents a GitHub issue
type Issue struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Author       User      `json:"author"`
	CreatedAt    string    `json:"createdAt"`
	UpdatedAt    string    `json:"updatedAt"`
	CommentCount int       `json:"commentCount"`
	Labels       []Label   `json:"labels"`
	Assignees    []User    `json:"assignees"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
}

// Label represents a GitHub label
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Comment represents a comment on a GitHub issue
type Comment struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	Author    User   `json:"author"`
	CreatedAt string `json:"createdAt"`
}

// FetchProgress contains progress information during issue fetching
type FetchProgress struct {
	Fetched int
	Total   int
}

// ProgressCallback is called during fetch operations to report progress
type ProgressCallback func(FetchProgress)

// CreatedAtTime parses the CreatedAt string into a time.Time
func (i *Issue) CreatedAtTime() (time.Time, error) {
	return time.Parse(time.RFC3339, i.CreatedAt)
}

// UpdatedAtTime parses the UpdatedAt string into a time.Time
func (i *Issue) UpdatedAtTime() (time.Time, error) {
	return time.Parse(time.RFC3339, i.UpdatedAt)
}

// CreatedAtTime parses the CreatedAt string into a time.Time
func (c *Comment) CreatedAtTime() (time.Time, error) {
	return time.Parse(time.RFC3339, c.CreatedAt)
}

const issuesQuery = `
query($owner: String!, $repo: String!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    issues(first: 100, after: $cursor, states: OPEN, orderBy: {field: UPDATED_AT, direction: DESC}) {
      totalCount
      pageInfo {
        hasNextPage
        endCursor
      }
      nodes {
        number
        title
        body
        createdAt
        updatedAt
        author {
          login
        }
        labels(first: 100) {
          nodes {
            name
            color
          }
        }
        assignees(first: 20) {
          nodes {
            login
          }
        }
        comments {
          totalCount
        }
      }
    }
  }
}
`

const commentsQuery = `
query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      comments(first: 100, after: $cursor) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          id
          body
          createdAt
          author {
            login
          }
        }
      }
    }
  }
}
`

// graphqlRequest represents a GraphQL request
type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// FetchIssues fetches all open issues from a repository
func (c *Client) FetchIssues(ctx context.Context, owner, repo string, progress ProgressCallback) ([]Issue, error) {
	var allIssues []Issue
	var cursor *string
	totalCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		issues, pageInfo, total, err := c.fetchIssuesPage(ctx, owner, repo, cursor)
		if err != nil {
			return nil, err
		}

		if totalCount == 0 {
			totalCount = total
		}

		allIssues = append(allIssues, issues...)

		if progress != nil {
			progress(FetchProgress{
				Fetched: len(allIssues),
				Total:   totalCount,
			})
		}

		if !pageInfo.HasNextPage {
			break
		}

		cursor = &pageInfo.EndCursor
	}

	return allIssues, nil
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

func (c *Client) fetchIssuesPage(ctx context.Context, owner, repo string, cursor *string) ([]Issue, pageInfo, int, error) {
	variables := map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	}
	if cursor != nil {
		variables["cursor"] = *cursor
	}

	req := graphqlRequest{
		Query:     issuesQuery,
		Variables: variables,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, pageInfo{}, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, pageInfo{}, 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, pageInfo{}, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, pageInfo{}, 0, fmt.Errorf("GitHub API error: %d %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Repository struct {
				Issues struct {
					TotalCount int      `json:"totalCount"`
					PageInfo   pageInfo `json:"pageInfo"`
					Nodes      []struct {
						Number    int    `json:"number"`
						Title     string `json:"title"`
						Body      string `json:"body"`
						CreatedAt string `json:"createdAt"`
						UpdatedAt string `json:"updatedAt"`
						Author    *struct {
							Login string `json:"login"`
						} `json:"author"`
						Labels struct {
							Nodes []struct {
								Name  string `json:"name"`
								Color string `json:"color"`
							} `json:"nodes"`
						} `json:"labels"`
						Assignees struct {
							Nodes []struct {
								Login string `json:"login"`
							} `json:"nodes"`
						} `json:"assignees"`
						Comments struct {
							TotalCount int `json:"totalCount"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"issues"`
			} `json:"repository"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, pageInfo{}, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, pageInfo{}, 0, fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
	}

	// Convert to Issue structs
	issues := make([]Issue, len(result.Data.Repository.Issues.Nodes))
	for i, node := range result.Data.Repository.Issues.Nodes {
		issue := Issue{
			Number:       node.Number,
			Title:        node.Title,
			Body:         node.Body,
			CreatedAt:    node.CreatedAt,
			UpdatedAt:    node.UpdatedAt,
			CommentCount: node.Comments.TotalCount,
		}

		if node.Author != nil {
			issue.Author = User{Login: node.Author.Login}
		}

		issue.Labels = make([]Label, len(node.Labels.Nodes))
		for j, label := range node.Labels.Nodes {
			issue.Labels[j] = Label{
				Name:  label.Name,
				Color: label.Color,
			}
		}

		issue.Assignees = make([]User, len(node.Assignees.Nodes))
		for j, assignee := range node.Assignees.Nodes {
			issue.Assignees[j] = User{Login: assignee.Login}
		}

		issues[i] = issue
	}

	return issues, result.Data.Repository.Issues.PageInfo, result.Data.Repository.Issues.TotalCount, nil
}

// FetchIssueComments fetches all comments for an issue
func (c *Client) FetchIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]Comment, error) {
	var allComments []Comment
	var cursor *string

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		comments, pageInfo, err := c.fetchCommentsPage(ctx, owner, repo, issueNumber, cursor)
		if err != nil {
			return nil, err
		}

		allComments = append(allComments, comments...)

		if !pageInfo.HasNextPage {
			break
		}

		cursor = &pageInfo.EndCursor
	}

	return allComments, nil
}

func (c *Client) fetchCommentsPage(ctx context.Context, owner, repo string, issueNumber int, cursor *string) ([]Comment, pageInfo, error) {
	variables := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": issueNumber,
	}
	if cursor != nil {
		variables["cursor"] = *cursor
	}

	req := graphqlRequest{
		Query:     commentsQuery,
		Variables: variables,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, pageInfo{}, fmt.Errorf("GitHub API error: %d %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Repository struct {
				Issue struct {
					Comments struct {
						PageInfo pageInfo `json:"pageInfo"`
						Nodes    []struct {
							ID        string `json:"id"`
							Body      string `json:"body"`
							CreatedAt string `json:"createdAt"`
							Author    *struct {
								Login string `json:"login"`
							} `json:"author"`
						} `json:"nodes"`
					} `json:"comments"`
				} `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, pageInfo{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, pageInfo{}, fmt.Errorf("GraphQL error: %s", result.Errors[0].Message)
	}

	// Convert to Comment structs
	comments := make([]Comment, len(result.Data.Repository.Issue.Comments.Nodes))
	for i, node := range result.Data.Repository.Issue.Comments.Nodes {
		comment := Comment{
			ID:        node.ID,
			Body:      node.Body,
			CreatedAt: node.CreatedAt,
		}

		if node.Author != nil {
			comment.Author = User{Login: node.Author.Login}
		}

		comments[i] = comment
	}

	return comments, result.Data.Repository.Issue.Comments.PageInfo, nil
}
