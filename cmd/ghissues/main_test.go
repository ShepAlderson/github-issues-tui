package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
)

func TestConfigCommandFlow(t *testing.T) {
	// This test verifies the 'ghissues config' command flow
	// We're testing the integration but not running the actual interactive setup

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Create a valid config
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Repository: "owner/repo",
			AuthMethod: "env",
		},
	}

	// Save config
	if err := config.SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Verify config exists
	exists, err := config.ConfigExists(configPath)
	if err != nil {
		t.Fatalf("Failed to check config existence: %v", err)
	}
	if !exists {
		t.Error("Config file was not created")
	}

	// Verify config can be loaded
	loaded, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config is valid
	if err := config.ValidateConfig(loaded); err != nil {
		t.Errorf("Config validation failed: %v", err)
	}
}

func TestConfigFilePermissions(t *testing.T) {
	// Test that config file is created with secure permissions
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Repository: "owner/repo",
			AuthMethod: "token",
			Token:      "ghp_secret123",
		},
	}

	if err := config.SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	mode := info.Mode()
	// Config should be readable/writable only by owner (0600)
	if mode.Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", mode.Perm())
	}
}

func TestFirstTimeSetupFlow(t *testing.T) {
	// Test that setup is triggered when no config exists
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Verify config doesn't exist initially
	exists, err := config.ConfigExists(configPath)
	if err != nil {
		t.Fatalf("Failed to check config existence: %v", err)
	}
	if exists {
		t.Error("Config should not exist initially")
	}

	// After setup runs (simulated by creating config)
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Repository: "owner/repo",
			AuthMethod: "env",
		},
	}

	if err := config.SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify config now exists
	exists, err = config.ConfigExists(configPath)
	if err != nil {
		t.Fatalf("Failed to check config existence: %v", err)
	}
	if !exists {
		t.Error("Config should exist after setup")
	}

	// On subsequent runs, setup should not be triggered
	// (config already exists)
	exists, err = config.ConfigExists(configPath)
	if err != nil {
		t.Fatalf("Failed to check config existence: %v", err)
	}
	if !exists {
		t.Error("Config should still exist")
	}
}

func TestConfigPathCreation(t *testing.T) {
	// Test that config directory is created if it doesn't exist
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "path", "config.toml")

	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Repository: "owner/repo",
			AuthMethod: "gh",
		},
	}

	// Save should create parent directories
	if err := config.SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("Failed to save config with nested path: %v", err)
	}

	// Verify parent directories were created
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Parent directories were not created")
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}
