package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// RepositoryConfig holds configuration for a single repository
type RepositoryConfig struct {
	DatabasePath string `toml:"database_path"`
}

// Config represents the application configuration
type Config struct {
	// Backward compatibility fields
	Repository string `toml:"repository"`
	Token      string `toml:"token"`
	Database   struct {
		Path string `toml:"path"`
	} `toml:"database"`

	// Multi-repository support
	DefaultRepo  string                      `toml:"default_repo"`
	Repositories map[string]RepositoryConfig `toml:"repositories"`

	Display struct {
		Columns []string `toml:"columns"`
		Sort    Sort     `toml:"sort"`
		Theme   string   `toml:"theme"`
	} `toml:"display"`
}

// Sort represents sort preferences
type Sort struct {
	Field      string `toml:"field"`
	Descending bool   `toml:"descending"`
}

// LoadConfig loads configuration from the specified path
// Returns nil if config file doesn't exist
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig saves configuration to the specified path
func SaveConfig(configPath string, cfg *Config) error {
	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Initialize Repositories map if it's nil to avoid TOML encoding issues
	if cfg.Repositories == nil {
		cfg.Repositories = make(map[string]RepositoryConfig)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write with restricted permissions (only owner can read/write)
	return os.WriteFile(configPath, data, 0600)
}

// GetDefaultConfigPath returns the default configuration file path
// ~/.config/ghissues/config.toml
func GetDefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "ghissues", "config.toml")
}

// ConfigExists checks if a config file exists at the specified path
func ConfigExists(configPath string) (bool, error) {
	_, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetDefaultDisplayColumns returns the default display columns
func GetDefaultDisplayColumns() []string {
	return []string{"number", "title", "author", "created_at", "comment_count"}
}

// GetDefaultSort returns the default sort preferences
func GetDefaultSort() Sort {
	return Sort{
		Field:      "updated_at",
		Descending: true,
	}
}

// GetDefaultTheme returns the default theme name
func GetDefaultTheme() string {
	return "default"
}

// GetRepoDatabasePath returns the database path for a specific repository
func GetRepoDatabasePath(cfg *Config, repoName string) (string, error) {
	// Check if repository exists in multi-repo config
	if repo, exists := cfg.Repositories[repoName]; exists {
		return repo.DatabasePath, nil
	}

	// Backward compatibility: if looking for the single repository, use its database path
	if repoName == cfg.Repository && cfg.Database.Path != "" {
		return cfg.Database.Path, nil
	}

	return "", fmt.Errorf("repository not found: %s", repoName)
}

// GetDefaultRepo returns the default repository to use
func GetDefaultRepo(cfg *Config) string {
	// Check if default_repo is set
	if cfg.DefaultRepo != "" {
		return cfg.DefaultRepo
	}

	// Fall back to single repository field (backward compatibility)
	if cfg.Repository != "" {
		return cfg.Repository
	}

	// As a last resort, return the first repository in the map
	for repoName := range cfg.Repositories {
		return repoName
	}

	return ""
}

// ListRepositories returns a list of all configured repository names
func ListRepositories(cfg *Config) []string {
	var repos []string

	// Add repositories from multi-repo config
	for repoName := range cfg.Repositories {
		repos = append(repos, repoName)
	}

	// If there are no multi-repo entries but there's a single repository, include it
	if len(repos) == 0 && cfg.Repository != "" {
		repos = append(repos, cfg.Repository)
	}

	return repos
}
