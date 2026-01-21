package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	gh "github.com/google/go-github/v62/github"
	"github.com/shepbook/github-issues-tui/internal/config"
	"golang.org/x/oauth2"
)

// AuthManager handles GitHub authentication and token resolution
type AuthManager struct {
	configManager *config.Manager
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(configManager *config.Manager) *AuthManager {
	return &AuthManager{
		configManager: configManager,
	}
}

// GetAuthenticatedClient returns an authenticated GitHub client
// It follows the authentication priority order:
// 1. GITHUB_TOKEN environment variable
// 2. Token stored in config file
// 3. GitHub CLI token
func (am *AuthManager) GetAuthenticatedClient(ctx context.Context) (*gh.Client, error) {
	token, err := am.resolveToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve authentication token: %w", err)
	}

	if token == "" {
		return nil, fmt.Errorf("no valid authentication found. Please set GITHUB_TOKEN environment variable, configure a token in ghissues config, or ensure GitHub CLI is authenticated")
	}

	// Validate the token by making a simple API call
	if err := am.validateToken(ctx, token); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Create authenticated client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return gh.NewClient(tc), nil
}

// resolveToken resolves the authentication token following priority order
func (am *AuthManager) resolveToken(ctx context.Context) (string, error) {
	// 1. Check GITHUB_TOKEN environment variable
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		return envToken, nil
	}

	// 2. Check config file
	cfg, err := am.configManager.Load()
	if err == nil && cfg.Auth.Method == "token" && cfg.Auth.Token != "" {
		return cfg.Auth.Token, nil
	}

	// 3. Check GitHub CLI
	ghToken, err := getGitHubCLIToken(ctx)
	if err == nil && ghToken != "" {
		return ghToken, nil
	}

	// No valid token found
	return "", nil
}

// getGitHubCLIToken retrieves token from GitHub CLI
func getGitHubCLIToken(ctx context.Context) (string, error) {
	// First check if gh CLI is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return "", fmt.Errorf("GitHub CLI not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, "gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get token from GitHub CLI: %w", err)
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("GitHub CLI returned empty token")
	}

	return token, nil
}

// validateToken validates the token by making a simple API call
func (am *AuthManager) validateToken(ctx context.Context, token string) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := gh.NewClient(tc)

	// Make a simple API call to validate the token
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	if user.Login == nil || *user.Login == "" {
		return fmt.Errorf("token validation succeeded but no user login returned")
	}

	return nil
}
