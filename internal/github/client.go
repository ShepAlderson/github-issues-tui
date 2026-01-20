package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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