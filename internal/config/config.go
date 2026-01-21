package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	Repository string         `toml:"repository"`
	Auth       AuthConfig     `toml:"auth"`
	Database   DatabaseConfig `toml:"database"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `toml:"path,omitempty"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Method string `toml:"method"` // "env", "token", or "gh"
	Token  string `toml:"token,omitempty"`
}

// DefaultConfigPath returns the default path for the config file
// (~/.config/ghissues/config.toml)
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home can't be determined
		return filepath.Join(".config", "ghissues", "config.toml")
	}
	return filepath.Join(homeDir, ".config", "ghissues", "config.toml")
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		Auth: AuthConfig{
			Method: "env", // Default to environment variable method
		},
	}
}

// Exists checks if a config file exists at the given path
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Save writes the config to the specified path with secure permissions (0600)
func Save(cfg *Config, path string) error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create or truncate the file with secure permissions
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode the config as TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Load reads and parses a config file from the specified path
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

// ValidateRepository validates that a repository string is in owner/repo format
func ValidateRepository(repo string) error {
	if repo == "" {
		return errors.New("repository cannot be empty")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return errors.New("repository must be in owner/repo format")
	}

	if parts[0] == "" {
		return errors.New("owner cannot be empty")
	}

	if parts[1] == "" {
		return errors.New("repository name cannot be empty")
	}

	return nil
}

// ValidateAuthMethod validates that an auth method is one of the supported methods
func ValidateAuthMethod(method string) error {
	switch method {
	case "env", "token", "gh":
		return nil
	default:
		return fmt.Errorf("invalid auth method: %q, must be one of: env, token, gh", method)
	}
}
