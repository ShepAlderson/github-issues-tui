package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// GetToken retrieves a GitHub authentication token from multiple sources in priority order:
// 1. Environment variable (GITHUB_TOKEN)
// 2. Config file
// 3. gh CLI token
//
// Returns the token, the source it came from, and an error if no valid authentication is found.
func GetToken(configPath, ghConfigDir string) (string, string, error) {
	// Try environment variable first
	if token, ok := GetTokenFromEnv(); ok {
		return token, "environment variable", nil
	}

	// Try config file
	if token, ok := GetTokenFromConfig(configPath); ok {
		return token, "config file", nil
	}

	// Try gh CLI
	if token, ok := GetTokenFromGhCLI(ghConfigDir); ok {
		return token, "gh CLI", nil
	}

	return "", "", fmt.Errorf("no valid GitHub authentication found\n\nPlease set the GITHUB_TOKEN environment variable,\nadd a token to your config file, or authenticate with gh CLI")
}

// GetTokenFromEnv attempts to get a token from the GITHUB_TOKEN environment variable.
// Returns the token and true if found, false otherwise.
func GetTokenFromEnv() (string, bool) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return "", false
	}
	return token, true
}

// GetTokenFromConfig attempts to get a token from the config file.
// Returns the token and true if found, false otherwise.
func GetTokenFromConfig(configPath string) (string, bool) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", false
	}

	var config struct {
		Auth struct {
			Token string `toml:"token"`
		} `toml:"auth"`
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		return "", false
	}

	if config.Auth.Token == "" {
		return "", false
	}

	return config.Auth.Token, true
}

// GetTokenFromGhCLI attempts to get a token from the gh CLI configuration.
// Returns the token and true if found, false otherwise.
func GetTokenFromGhCLI(ghConfigDir string) (string, bool) {
	hostsPath := filepath.Join(ghConfigDir, "hosts.yml")

	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return "", false
	}

	var hostsConfig struct {
		GitHubCom struct {
			OAuthToken string `yaml:"oauth_token"`
		} `yaml:"github.com"`
	}

	err = yaml.Unmarshal(data, &hostsConfig)
	if err != nil {
		return "", false
	}

	if hostsConfig.GitHubCom.OAuthToken == "" {
		return "", false
	}

	return hostsConfig.GitHubCom.OAuthToken, true
}
