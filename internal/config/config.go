package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration
type Config struct {
	GitHub   GitHubConfig   `toml:"github"`
	Database DatabaseConfig `toml:"database,omitempty"`
	Display  DisplayConfig  `toml:"display,omitempty"`
}

// GitHubConfig contains GitHub-related settings
type GitHubConfig struct {
	Repository string `toml:"repository"`
	AuthMethod string `toml:"auth_method"`
	Token      string `toml:"token,omitempty"`
}

// DatabaseConfig contains database-related settings
type DatabaseConfig struct {
	Path string `toml:"path,omitempty"`
}

// DisplayConfig contains display-related settings
type DisplayConfig struct {
	Theme   string   `toml:"theme,omitempty"`
	Columns []string `toml:"columns,omitempty"`
}

// ConfigPath returns the default configuration file path
func ConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "ghissues", "config.toml")
}

// ConfigExists checks if a config file exists at the given path
func ConfigExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// LoadConfig loads configuration from the specified path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves configuration to the specified path
func SaveConfig(cfg *Config, path string) error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Save with secure permissions (0600) as per US-002 requirements
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig validates the configuration
func ValidateConfig(cfg *Config) error {
	if cfg.GitHub.Repository == "" {
		return errors.New("repository is required")
	}

	// Validate repository format (owner/repo)
	parts := strings.Split(cfg.GitHub.Repository, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("repository must be in owner/repo format")
	}

	// Validate auth method
	if cfg.GitHub.AuthMethod == "" {
		return errors.New("auth_method is required")
	}

	validAuthMethods := map[string]bool{
		"token": true,
		"env":   true,
		"gh":    true,
	}
	if !validAuthMethods[cfg.GitHub.AuthMethod] {
		return fmt.Errorf("invalid auth_method: %s (must be one of: token, env, gh)", cfg.GitHub.AuthMethod)
	}

	// If auth method is token, token must be provided
	if cfg.GitHub.AuthMethod == "token" && cfg.GitHub.Token == "" {
		return errors.New("token is required when auth_method is 'token'")
	}

	return nil
}

// GetDisplayColumns returns the columns to display in the issue list
// Returns defaults if not configured or empty
func GetDisplayColumns(cfg *Config) []string {
	defaultColumns := []string{"number", "title", "author", "date", "comments"}

	if len(cfg.Display.Columns) == 0 {
		return defaultColumns
	}

	return cfg.Display.Columns
}
