package auth

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	"github.com/shepbook/ghissues/internal/config"
)

// TokenSource represents where a token was obtained from
type TokenSource string

const (
	TokenSourceEnv    TokenSource = "environment variable (GITHUB_TOKEN)"
	TokenSourceConfig TokenSource = "config file"
	TokenSourceGhCli  TokenSource = "GitHub CLI"
)

// ErrNoAuthFound is returned when no valid authentication is found
var ErrNoAuthFound = errors.New(`no GitHub authentication found.

Please set one of the following:
1. Environment variable: export GITHUB_TOKEN=your_token
2. Config file token: run 'ghissues config' to set a personal access token
3. GitHub CLI: ensure you are logged in with 'gh auth login'`)

// GetToken attempts to get a GitHub token in order of precedence:
// 1. GITHUB_TOKEN environment variable
// 2. Token from config file
// 3. Token from GitHub CLI (gh auth token)
func GetToken() (string, TokenSource, error) {
	// 1. Try GITHUB_TOKEN environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, TokenSourceEnv, nil
	}

	// 2. Try config file
	if config.Exists() {
		cfg, err := config.Load()
		if err == nil && cfg.Token != "" {
			return cfg.Token, TokenSourceConfig, nil
		}
	}

	// 3. Try GitHub CLI
	token, err := getGhAuthToken()
	if err == nil && token != "" {
		return token, TokenSourceGhCli, nil
	}

	return "", "", ErrNoAuthFound
}

// getGhAuthToken runs 'gh auth token' to get the token from GitHub CLI
func getGhAuthToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return stringTrim(output), nil
}

// stringTrim trims whitespace and newlines from the output
func stringTrim(b []byte) string {
	return string(bytes.TrimSpace(b))
}