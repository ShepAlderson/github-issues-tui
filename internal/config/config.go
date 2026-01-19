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
	GitHub       GitHubConfig        `toml:"github"`
	Database     DatabaseConfig      `toml:"database,omitempty"`
	Display      DisplayConfig       `toml:"display,omitempty"`
	Repositories []RepositoryConfig  `toml:"repositories,omitempty"`
}

// GitHubConfig contains GitHub-related settings
type GitHubConfig struct {
	Repository        string `toml:"repository,omitempty"`        // Legacy single repository support
	AuthMethod        string `toml:"auth_method"`
	Token             string `toml:"token,omitempty"`
	DefaultRepository string `toml:"default_repository,omitempty"`
}

// RepositoryConfig represents a configured repository
type RepositoryConfig struct {
	Name string `toml:"name"`
}

// DatabaseConfig contains database-related settings
type DatabaseConfig struct {
	Path string `toml:"path,omitempty"`
}

// DisplayConfig contains display-related settings
type DisplayConfig struct {
	Theme         string   `toml:"theme,omitempty"`
	Columns       []string `toml:"columns,omitempty"`
	SortBy        string   `toml:"sort_by,omitempty"`
	SortAscending bool     `toml:"sort_ascending,omitempty"`
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

	// Check if we have any repositories configured
	if len(cfg.Repositories) == 0 && cfg.GitHub.Repository == "" {
		return errors.New("at least one repository must be configured")
	}

	// Validate legacy single repository if set
	if cfg.GitHub.Repository != "" {
		parts := strings.Split(cfg.GitHub.Repository, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return errors.New("repository must be in owner/repo format")
		}
	}

	// Validate all repositories in the list
	for _, repo := range cfg.Repositories {
		parts := strings.Split(repo.Name, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("repository '%s' must be in owner/repo format", repo.Name)
		}
	}

	// Validate default_repository is in the configured list
	if cfg.GitHub.DefaultRepository != "" && len(cfg.Repositories) > 0 {
		found := false
		for _, repo := range cfg.Repositories {
			if repo.Name == cfg.GitHub.DefaultRepository {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default_repository '%s' is not in the configured repositories list", cfg.GitHub.DefaultRepository)
		}
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

// GetSortBy returns the sort field to use
// Returns "updated" (default) if not configured or invalid
func GetSortBy(cfg *Config) string {
	if cfg.Display.SortBy == "" {
		return "updated"
	}

	// Validate sort field
	validSortFields := map[string]bool{
		"updated":  true,
		"created":  true,
		"number":   true,
		"comments": true,
	}

	if !validSortFields[cfg.Display.SortBy] {
		return "updated"
	}

	return cfg.Display.SortBy
}

// GetSortAscending returns whether to sort in ascending order
// Returns false (descending) by default if not configured
func GetSortAscending(cfg *Config) bool {
	return cfg.Display.SortAscending
}

// GetTheme returns the theme name to use
// Returns "default" if not configured or if unknown theme specified
func GetTheme(cfg *Config) string {
	if cfg.Display.Theme == "" {
		return "default"
	}

	// List of valid themes
	validThemes := map[string]bool{
		"default":         true,
		"dracula":         true,
		"gruvbox":         true,
		"nord":            true,
		"solarized-dark":  true,
		"solarized-light": true,
	}

	theme := strings.ToLower(strings.TrimSpace(cfg.Display.Theme))
	if !validThemes[theme] {
		return "default"
	}

	return theme
}

// GetRepository resolves which repository to use based on precedence:
// 1. Command-line --repo flag (highest priority)
// 2. Default repository in config
// 3. First repository in list
// 4. Legacy single repository field
func GetRepository(cfg *Config, repoFlag string) (string, error) {
	// If flag is provided, use it (must be in configured list)
	if repoFlag != "" {
		// Check if it's in the configured repositories
		if len(cfg.Repositories) > 0 {
			for _, repo := range cfg.Repositories {
				if repo.Name == repoFlag {
					return repoFlag, nil
				}
			}
			return "", fmt.Errorf("repository '%s' is not configured", repoFlag)
		}
		// For legacy single repository, allow flag if it matches
		if cfg.GitHub.Repository == repoFlag {
			return repoFlag, nil
		}
		return "", fmt.Errorf("repository '%s' is not configured", repoFlag)
	}

	// If default repository is set, use it
	if cfg.GitHub.DefaultRepository != "" {
		return cfg.GitHub.DefaultRepository, nil
	}

	// Use first repository in list
	if len(cfg.Repositories) > 0 {
		return cfg.Repositories[0].Name, nil
	}

	// Fall back to legacy single repository
	if cfg.GitHub.Repository != "" {
		return cfg.GitHub.Repository, nil
	}

	return "", errors.New("no repository configured")
}

// ListRepositories returns a list of all configured repositories
func ListRepositories(cfg *Config) []string {
	if len(cfg.Repositories) > 0 {
		repos := make([]string, len(cfg.Repositories))
		for i, repo := range cfg.Repositories {
			repos[i] = repo.Name
		}
		return repos
	}

	// Fall back to legacy single repository
	if cfg.GitHub.Repository != "" {
		return []string{cfg.GitHub.Repository}
	}

	return []string{}
}

// GetDatabasePathForRepository returns the database path for a specific repository
// Databases are stored in ~/.local/share/ghissues/<owner_repo>.db
func GetDatabasePathForRepository(repo string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to current directory if home dir not available
		return strings.ReplaceAll(repo, "/", "_") + ".db"
	}

	// Convert owner/repo to owner_repo for filename
	dbName := strings.ReplaceAll(repo, "/", "_") + ".db"
	return filepath.Join(homeDir, ".local", "share", "ghissues", dbName)
}
