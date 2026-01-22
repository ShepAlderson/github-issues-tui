package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shepbook/github-issues-tui/internal/config"
)

// GetToken retrieves a GitHub token using the configured authentication method.
// It tries authentication methods in the following order:
// 1. GITHUB_TOKEN environment variable (if auth_method is "env")
// 2. Token from config file (if auth_method is "token")
// 3. GitHub CLI (if auth_method is "gh")
// Returns an error if no valid token is found.
func GetToken(cfg *config.Config) (string, error) {
	switch cfg.GitHub.AuthMethod {
	case "env":
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return "", errors.New("GITHUB_TOKEN environment variable is not set. Please set it or use a different authentication method (run 'ghissues config' to reconfigure)")
		}
		return token, nil

	case "token":
		if cfg.GitHub.Token == "" {
			return "", errors.New("no token found in config file. Please run 'ghissues config' to set up authentication")
		}
		return cfg.GitHub.Token, nil

	case "gh":
		return getGhToken()

	default:
		return "", fmt.Errorf("unknown authentication method: %s", cfg.GitHub.AuthMethod)
	}
}

// getGhToken retrieves a token from the GitHub CLI
func getGhToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get token from gh CLI: %s. Please ensure gh CLI is installed and authenticated (run 'gh auth login')", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute gh CLI: %w. Please ensure gh CLI is installed", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", errors.New("gh CLI returned an empty token. Please run 'gh auth login' to authenticate")
	}

	return token, nil
}

// ValidateToken validates a GitHub token by making an API request to /user endpoint.
// Returns nil if the token is valid, or an error with a helpful message if invalid.
func ValidateToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w. Please check your network connection", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("invalid GitHub token. Please check your token and try again (run 'ghissues config' to reconfigure)")
	}

	if resp.StatusCode == http.StatusForbidden {
		// Try to parse the error message for rate limiting
		var errorResponse struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			if strings.Contains(strings.ToLower(errorResponse.Message), "rate limit") {
				return errors.New("GitHub API rate limit exceeded. Please try again later")
			}
		}
		return errors.New("access forbidden. Your token might lack required permissions")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from GitHub API: %s", resp.Status)
	}

	return nil
}
