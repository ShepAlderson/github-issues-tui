package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ConfigPath returns the path to the config directory (~/.config/ghissues)
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".config", "ghissues")
}

// ConfigFilePath returns the path to the config file (~/.config/ghissues/config.toml)
func ConfigFilePath() string {
	return filepath.Join(ConfigPath(), "config.toml")
}

// Exists returns true if the config file exists
func Exists() bool {
	_, err := os.Stat(ConfigFilePath())
	return err == nil
}

// AuthMethod represents the authentication method preference
type AuthMethod string

const (
	AuthMethodEnv   AuthMethod = "env"   // GITHUB_TOKEN environment variable
	AuthMethodGhCli AuthMethod = "gh"    // GitHub CLI
	AuthMethodToken AuthMethod = "token" // Personal access token
)

// Config represents the application configuration
type Config struct {
	Repository string     `toml:"repository"`
	AuthMethod AuthMethod `toml:"auth_method"`
	Token      string     `toml:"token,omitempty"`
	Database   Database   `toml:"database"`
	Display    Display    `toml:"display"`
}

// Database represents database configuration
type Database struct {
	Path string `toml:"path"`
}

// SortOption represents the sort criteria for issues
type SortOption string

const (
	SortUpdated  SortOption = "updated"  // Sort by updated date
	SortCreated  SortOption = "created"  // Sort by created date
	SortNumber   SortOption = "number"   // Sort by issue number
	SortComments SortOption = "comments" // Sort by comment count
)

// SortOrder represents the sort order direction
type SortOrder string

const (
	SortOrderDesc SortOrder = "desc" // Descending (default)
	SortOrderAsc  SortOrder = "asc"  // Ascending
)

// DefaultSort returns the default sort option
func DefaultSort() SortOption {
	return SortUpdated
}

// DefaultSortOrder returns the default sort order
func DefaultSortOrder() SortOrder {
	return SortOrderDesc
}

// DefaultTheme returns the default theme name
func DefaultTheme() string {
	return "default"
}

// AllSortOptions returns all available sort options
func AllSortOptions() []SortOption {
	return []SortOption{SortUpdated, SortCreated, SortNumber, SortComments}
}

// Display represents display configuration
type Display struct {
	Columns    []string   `toml:"columns"`
	Sort       SortOption `toml:"sort"`
	SortOrder  SortOrder  `toml:"sort_order"`
	Theme      string     `toml:"theme"`
}

// DefaultColumns returns the default columns to display in the issue list
func DefaultColumns() []string {
	return []string{"number", "title", "author", "date", "comments"}
}

// Load reads the configuration from the config file
func Load() (*Config, error) {
	path := ConfigFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := parseConfig(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults
	if len(cfg.Display.Columns) == 0 {
		cfg.Display.Columns = DefaultColumns()
	}
	if cfg.Display.Sort == "" {
		cfg.Display.Sort = DefaultSort()
	}
	if cfg.Display.SortOrder == "" {
		cfg.Display.SortOrder = DefaultSortOrder()
	}
	if cfg.Display.Theme == "" {
		cfg.Display.Theme = DefaultTheme()
	}

	return &cfg, nil
}

// Save writes the configuration to the config file
func Save(cfg *Config) error {
	dir := ConfigPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := formatConfig(cfg)
	if err != nil {
		return err
	}

	path := ConfigFilePath()
	return os.WriteFile(path, data, 0600)
}

// parseConfig parses TOML data into Config struct
func parseConfig(data []byte, cfg *Config) error {
	_, err := toml.Decode(string(data), cfg)
	return err
}

// formatConfig formats Config struct to TOML bytes
func formatConfig(cfg *Config) ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	if err := enc.Encode(cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
