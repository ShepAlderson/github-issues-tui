package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration
type Config struct {
	Repository string `toml:"repository"`
	Auth       Auth   `toml:"auth"`
	Database   Database `toml:"database"`
	Display    Display  `toml:"display"`
}

// Auth represents authentication configuration
type Auth struct {
	Method string `toml:"method"`
	Token  string `toml:"token,omitempty"`
}

// Database represents database configuration
type Database struct {
	Path string `toml:"path"`
}

// Display represents display configuration
type Display struct {
	Theme string `toml:"theme"`
}

// Manager handles configuration operations
type Manager struct {
	configDirFunc func() (string, error)
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		configDirFunc: os.UserConfigDir,
	}
}

// NewTestManager creates a new configuration manager with a custom config directory function
// This is useful for testing
func NewTestManager(configDirFunc func() (string, error)) *Manager {
	return &Manager{
		configDirFunc: configDirFunc,
	}
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		Repository: "",
		Auth: Auth{
			Method: "gh",
		},
		Database: Database{
			Path: ".ghissues.db",
		},
		Display: Display{
			Theme: "default",
		},
	}
}

// ConfigDir returns the configuration directory path
func (m *Manager) ConfigDir() (string, error) {
	configDir, err := m.configDirFunc()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	return filepath.Join(configDir, "ghissues"), nil
}

// ConfigPath returns the path to the config file
func (m *Manager) ConfigPath() (string, error) {
	configDir, err := m.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.toml"), nil
}

// Load loads the configuration from file
func (m *Manager) Load() (*Config, error) {
	configPath, err := m.ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %s", configPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to file
func (m *Manager) Save(cfg *Config) error {
	configDir, err := m.ConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := m.ConfigPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the config file exists
func (m *Manager) Exists() (bool, error) {
	configPath, err := m.ConfigPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}