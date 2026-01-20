package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/shepbook/git/github-issues-tui/internal/config"
)

// defaultGetDefaultConfigPath and defaultGetDefaultGhHostsPath are variables that can be overridden in tests
var (
	defaultGetDefaultConfigPath  = getDefaultConfigPath
	defaultGetDefaultGhHostsPath = getDefaultGhHostsPath
	defaultGetUserHomeDir        = os.UserHomeDir
	defaultReadFile              = os.ReadFile
	defaultGetenv                = os.Getenv
)

// GetGitHubToken returns a GitHub token and its source by checking in priority order:
// 1. GITHUB_TOKEN environment variable
// 2. Config file token (from provided config struct)
// 3. gh CLI token from ~/.config/gh/hosts.yml
func GetGitHubToken(cfg *config.Config) (token string, source string, err error) {
	// Check GITHUB_TOKEN environment variable first
	if token := getTokenFromEnv(); token != "" {
		return token, "environment", nil
	}

	// Check config file token
	if cfg != nil && cfg.Token != "" {
		return cfg.Token, "config", nil
	}

	// Check gh CLI token
	if token, err := getTokenFromGhCLI(); err != nil {
		// Don't fail if gh CLI config doesn't exist
		if !os.IsNotExist(err) {
			return "", "", fmt.Errorf("error reading gh CLI config: %w", err)
		}
	} else if token != "" {
		return token, "gh cli", nil
	}

	return "", "", fmt.Errorf("no GitHub authentication token found. Please set GITHUB_TOKEN environment variable, configure a token in ~/.config/ghissues/config.toml, or authenticate with gh CLI (gh auth login)")
}

// getTokenFromEnv returns the GITHUB_TOKEN from environment if set
func getTokenFromEnv() string {
	return defaultGetenv("GITHUB_TOKEN")
}

// getTokenFromConfig returns the token from the config file
func getTokenFromConfig() (string, error) {
	configPath := defaultGetDefaultConfigPath()
	data, err := defaultReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	// Parse as a simple map first to check structure
	var rawConfig map[string]interface{}
	if err := toml.Unmarshal(data, &rawConfig); err != nil {
		return "", fmt.Errorf("failed to parse config: %w", err)
	}

	if token, ok := rawConfig["token"].(string); ok {
		return token, nil
	}

	return "", nil
}

// getTokenFromGhCLI returns the token from gh CLI configuration
func getTokenFromGhCLI() (string, error) {
	hostsPath := defaultGetDefaultGhHostsPath()
	data, err := defaultReadFile(hostsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	// Use a simple line-by-line parser to extract the oauth_token value
	lines := string(data)
	inGitHubSection := false
	for _, line := range splitLines(lines) {
		trimmed := line
		if len(line) > 0 {
			trimmed = string(line)
		}

		// Check if we're entering the github.com section
		if trimmed == "github.com:" {
			inGitHubSection = true
			continue
		}

		// If we're in the github.com section, look for oauth_token
		if inGitHubSection {
			// Parse oauth_token line
			if idx := find(trimmed, "oauth_token:"); idx >= 0 {
				// Extract the token value
				valuePart := trimmed[idx+len("oauth_token:"):]
				valuePart = trimSpace(valuePart)
				// Remove quotes if present
				if len(valuePart) >= 2 && valuePart[0] == '"' && valuePart[len(valuePart)-1] == '"' {
					valuePart = valuePart[1 : len(valuePart)-1]
				}
				return valuePart, nil
			}

			// If we hit another top-level key (no indentation), we're out of github.com section
			if len(trimmed) > 0 && trimmed[0] != ' ' && trimmed[0] != '\t' && trimmed[len(trimmed)-1] == ':' {
				if trimmed != "github.com:" {
					break
				}
			}
		}
	}

	return "", nil
}

// Helper functions for string operations
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func find(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}
	for start < end && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}
	if start > end {
		return ""
	}
	return s[start:end]
}

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	homeDir, _ := defaultGetUserHomeDir()
	if homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, ".config", "ghissues", "config.toml")
}

// getDefaultGhHostsPath returns the default gh CLI hosts.yml path
func getDefaultGhHostsPath() string {
	homeDir, _ := defaultGetUserHomeDir()
	if homeDir == "" {
		return ""
	}
	return filepath.Join(homeDir, ".config", "gh", "hosts.yml")
}

// ValidateToken checks if a GitHub token is valid
// This is a basic validation that checks if the token is non-empty
// A full implementation would make an API call to GitHub to verify the token
func ValidateToken(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("no GitHub token provided")
	}

	// Token is non-empty, consider it potentially valid
	// In a real implementation, this would make an API call to GitHub
	// to verify the token is actually valid and has the required permissions

	return true, nil
}
