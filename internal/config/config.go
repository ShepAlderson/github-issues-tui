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
}

// Database represents database configuration
type Database struct {
	Path string `toml:"path"`
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
