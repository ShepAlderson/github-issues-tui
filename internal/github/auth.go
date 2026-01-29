package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shepbook/ghissues/internal/config"
)

// TokenSource represents where the token was obtained from
type TokenSource string

const (
	// TokenSourceEnvVar indicates the token came from GITHUB_TOKEN environment variable
	TokenSourceEnvVar TokenSource = "GITHUB_TOKEN environment variable"
	// TokenSourceConfig indicates the token came from the config file
	TokenSourceConfig TokenSource = "config file"
	// TokenSourceGhCLI indicates the token came from gh CLI
	TokenSourceGhCLI TokenSource = "gh CLI"
)

// ResolveToken attempts to resolve a GitHub authentication token using the following priority:
// 1. GITHUB_TOKEN environment variable
// 2. Config file token
// 3. gh CLI authentication
//
// Returns an error if no valid authentication is found.
func ResolveToken() (string, error) {
	// Priority 1: Environment variable
	if token, found := getEnvToken(); found {
		return token, nil
	}

	// Priority 2: Config file
	if token, found := getConfigToken(); found {
		return token, nil
	}

	// Priority 3: gh CLI
	if token, found := getGhCliToken(); found {
		return token, nil
	}

	// No authentication found
	return "", fmt.Errorf(`no GitHub authentication found

Please configure authentication using one of these methods:
1. Set GITHUB_TOKEN environment variable
2. Run 'ghissues config' to save a token to your config file
3. Login with 'gh auth login' to use gh CLI authentication`)
}

// GetTokenWithSource attempts to resolve a GitHub authentication token and returns
// the token along with information about where it was sourced from.
//
// This is useful for debugging authentication issues.
func GetTokenWithSource() (string, string, error) {
	// Priority 1: Environment variable
	if token, found := getEnvToken(); found {
		return token, string(TokenSourceEnvVar), nil
	}

	// Priority 2: Config file
	if token, found := getConfigToken(); found {
		return token, string(TokenSourceConfig), nil
	}

	// Priority 3: gh CLI
	if token, found := getGhCliToken(); found {
		return token, string(TokenSourceGhCLI), nil
	}

	// No authentication found
	return "", "", fmt.Errorf(`no GitHub authentication found

Please configure authentication using one of these methods:
1. Set GITHUB_TOKEN environment variable
2. Run 'ghissues config' to save a token to your config file
3. Login with 'gh auth login' to use gh CLI authentication`)
}

// getEnvToken retrieves the token from the GITHUB_TOKEN environment variable
func getEnvToken() (string, bool) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", false
	}
	return token, true
}

// getConfigToken retrieves the token from the config file
func getConfigToken() (string, bool) {
	cfg, err := config.Load()
	if err != nil {
		return "", false
	}

	if cfg.Auth.Token == "" {
		return "", false
	}

	return cfg.Auth.Token, true
}

// getGhCliToken attempts to retrieve a token from the gh CLI
// by running 'gh auth token'
func getGhCliToken() (string, bool) {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return "", false
	}

	// Run gh auth token
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", false
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", false
	}

	return token, true
}

// ValidateToken validates that a token is not empty
// In a real implementation, this would make an API call to verify the token
func ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	return nil
}
