package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	Default      DefaultConfig  `toml:"default"`
	Repositories []Repository   `toml:"repositories"`
	Display      DisplayConfig  `toml:"display"`
	Sort         SortConfig     `toml:"sort"`
	Database     DatabaseConfig `toml:"database"`
	Auth         AuthConfig     `toml:"auth"`
}

// DefaultConfig contains default settings
type DefaultConfig struct {
	Repository string `toml:"repository"`
}

// Repository represents a GitHub repository configuration
type Repository struct {
	Owner    string `toml:"owner"`
	Name     string `toml:"name"`
	Database string `toml:"database"`
}

// DisplayConfig contains display settings
type DisplayConfig struct {
	Theme   string   `toml:"theme"`
	Columns []string `toml:"columns"`
}

// SortConfig contains sorting settings
type SortConfig struct {
	Field      string `toml:"field"`
	Descending bool   `toml:"descending"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Path string `toml:"path"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Token string `toml:"token"`
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "ghissues", "config.toml")
}

// ConfigExists checks if the config file exists
func ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return err == nil
}

// LoadConfig loads the configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(configPath string, cfg *Config) error {
	// Ensure directory exists
	err := EnsureConfigDir(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with secure permissions
	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir(configPath string) error {
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return os.MkdirAll(configDir, 0755)
	}
	return nil
}

// ValidateRepository validates a repository string in owner/repo format
func ValidateRepository(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	// Check format: owner/repo
	repoRegex := regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)
	if !repoRegex.MatchString(repo) {
		return fmt.Errorf("repository must be in 'owner/repo' format")
	}

	return nil
}
