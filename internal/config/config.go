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
	Auth         AuthConfig         `toml:"auth"`
	Default      DefaultConfig      `toml:"default"`
	Repositories []RepositoryConfig `toml:"repositories"`
	Display      DisplayConfig      `toml:"display"`
	Sort         SortConfig         `toml:"sort"`
	Database     DatabaseConfig     `toml:"database"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Token string `toml:"token"`
}

// DefaultConfig contains default settings
type DefaultConfig struct {
	Repository string `toml:"repository"`
}

// RepositoryConfig represents a configured repository
type RepositoryConfig struct {
	Owner    string `toml:"owner"`
	Name     string `toml:"name"`
	Database string `toml:"database"`
}

// DisplayConfig contains display settings
type DisplayConfig struct {
	Theme   string   `toml:"theme"`
	Columns []string `toml:"columns"`
}

// SortConfig contains sort settings
type SortConfig struct {
	Field      string `toml:"field"`
	Descending bool   `toml:"descending"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	Path string `toml:"path"`
}

// ConfigPath returns the default path to the configuration file
func ConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if we can't get home
		return filepath.Join(".config", "ghissues", "config.toml")
	}
	return filepath.Join(homeDir, ".config", "ghissues", "config.toml")
}

// Exists checks if the configuration file exists
func Exists() bool {
	path := ConfigPath()
	_, err := os.Stat(path)
	return err == nil
}

// Load reads the configuration from the config file
// Returns an empty config if the file doesn't exist
func Load() (*Config, error) {
	path := ConfigPath()

	cfg := &Config{
		Display: DisplayConfig{
			Theme:   "default",
			Columns: []string{"number", "title", "author", "updated", "comments"},
		},
		Sort: SortConfig{
			Field:      "updated",
			Descending: true,
		},
		Database: DatabaseConfig{
			Path: ".ghissues.db",
		},
	}

	// If config doesn't exist, return empty config with defaults
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	// Read and parse the config file
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to the config file
// Creates parent directories if they don't exist
func (c *Config) Save() error {
	path := ConfigPath()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create file with restricted permissions (owner read/write only)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// ValidateRepository checks if a repository string is in valid "owner/repo" format
func ValidateRepository(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	parts := splitRepository(repo)
	if len(parts) != 2 {
		return fmt.Errorf("repository must be in format 'owner/repo'")
	}

	if parts[0] == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}

	if parts[1] == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	// Validate characters (alphanumeric, hyphens, underscores)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(parts[0]) {
		return fmt.Errorf("repository owner contains invalid characters")
	}
	if !validPattern.MatchString(parts[1]) {
		return fmt.Errorf("repository name contains invalid characters")
	}

	return nil
}

// ParseRepository splits a repository string into owner and name
func ParseRepository(repo string) (owner, name string, err error) {
	if err := ValidateRepository(repo); err != nil {
		return "", "", err
	}

	parts := splitRepository(repo)
	return parts[0], parts[1], nil
}

// splitRepository splits a repo string by "/"
func splitRepository(repo string) []string {
	var parts []string
	current := ""
	for _, r := range repo {
		if r == '/' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	parts = append(parts, current)
	return parts
}
