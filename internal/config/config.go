package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration
type Config struct {
	Repository string `toml:"repository"`
	Token      string `toml:"token"`
	Database   struct {
		Path string `toml:"path"`
	} `toml:"database"`
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
