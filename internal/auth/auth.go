package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shepbook/ghissues/internal/config"
)

// TokenSource indicates where a token was retrieved from
type TokenSource int

const (
	SourceEnvVar TokenSource = iota
	SourceConfig
	SourceGhCLI
)

// String returns a human-readable description of the token source
func (s TokenSource) String() string {
	switch s {
	case SourceEnvVar:
		return "environment variable (GITHUB_TOKEN)"
	case SourceConfig:
		return "config file"
	case SourceGhCLI:
		return "gh CLI"
	default:
		return "unknown"
	}
}

// ErrNoAuth is returned when no valid authentication method is found
var ErrNoAuth = errors.New("no valid authentication found")

// ghCLITokenFunc is the function used to get tokens from gh CLI.
// It can be overridden in tests.
var ghCLITokenFunc = getTokenFromGhCLI

// GetToken attempts to retrieve a GitHub token using the following priority:
// 1. GITHUB_TOKEN environment variable
// 2. Token from config file (if auth.method is "token")
// 3. GitHub CLI (gh auth token)
//
// Returns the token, its source, and any error.
// If no valid authentication is found, returns a helpful error message.
func GetToken(cfg *config.Config) (string, TokenSource, error) {
	// 1. Try environment variable first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, SourceEnvVar, nil
	}

	// 2. Try config file token (only if method is "token")
	if cfg != nil && cfg.Auth.Method == "token" && cfg.Auth.Token != "" {
		return cfg.Auth.Token, SourceConfig, nil
	}

	// 3. Try gh CLI
	if token, err := ghCLITokenFunc(); err == nil && token != "" {
		return token, SourceGhCLI, nil
	}

	// No authentication found - return helpful error
	return "", 0, fmt.Errorf("%w: tried GITHUB_TOKEN env var, config file token, and gh CLI. "+
		"Please set GITHUB_TOKEN environment variable, run 'ghissues config' to set up authentication, "+
		"or install and authenticate the GitHub CLI (gh auth login)", ErrNoAuth)
}

// getTokenFromGhCLI attempts to get a token using the gh CLI
func getTokenFromGhCLI() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get token from gh CLI: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", errors.New("gh CLI returned empty token")
	}

	return token, nil
}

// ErrInvalidToken is returned when a token fails validation
var ErrInvalidToken = errors.New("invalid GitHub token")

// ValidateToken checks if a GitHub token is valid by making a test API call.
// Returns nil if the token is valid, or an error with a helpful message if not.
func ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("%w: token is empty", ErrInvalidToken)
	}

	// Make a simple API call to verify the token
	// GET /user is a good endpoint because it requires authentication
	// and returns 401 for invalid tokens
	return validateTokenWithAPI(token)
}

// validateTokenWithAPI makes an actual API call to validate the token
func validateTokenWithAPI(token string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: authentication failed (401 Unauthorized). Please check that your token is correct and has not expired", ErrInvalidToken)
	case http.StatusForbidden:
		return fmt.Errorf("%w: access forbidden (403 Forbidden). Your token may have insufficient permissions or be rate limited", ErrInvalidToken)
	default:
		return fmt.Errorf("%w: unexpected response status %d", ErrInvalidToken, resp.StatusCode)
	}
}
