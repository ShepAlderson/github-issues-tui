package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")

	// Test loading non-existent config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig should not error on non-existent file: %v", err)
	}
	if cfg != nil {
		t.Error("LoadConfig should return nil for non-existent config")
	}

	// Create a valid config file
	configContent := `
repository = "testuser/testrepo"
token = "test-token"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading existing config
	cfg, err = LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig should return config for existing file")
	}
	if cfg.Repository != "testuser/testrepo" {
		t.Errorf("Expected repository 'testuser/testrepo', got %q", cfg.Repository)
	}
	if cfg.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", cfg.Token)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")

	// Create a config to save
	cfg := &Config{
		Repository: "testuser/testrepo",
		Token:      "test-token",
	}

	// Save the config
	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify file permissions
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600, got %v", info.Mode().Perm())
	}

	// Load and verify the saved config
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	if loadedCfg.Repository != cfg.Repository {
		t.Errorf("Expected repository %q, got %q", cfg.Repository, loadedCfg.Repository)
	}
	if loadedCfg.Token != cfg.Token {
		t.Errorf("Expected token %q, got %q", cfg.Token, loadedCfg.Token)
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()

	// Should contain expected directory structure
	expectedSuffix := filepath.Join(".config", "ghissues", "config.toml")
	if len(path) < len(expectedSuffix) {
		t.Errorf("Config path too short: %s", path)
	}
	if path[len(path)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("Config path should end with %q, got %q", expectedSuffix, path)
	}
}

func TestConfigExists(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")

	// Test non-existent config
	exists, err := ConfigExists(configPath)
	if err != nil {
		t.Fatalf("ConfigExists failed: %v", err)
	}
	if exists {
		t.Error("ConfigExists should return false for non-existent file")
	}

	// Create config file
	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test existing config
	exists, err = ConfigExists(configPath)
	if err != nil {
		t.Fatalf("ConfigExists failed: %v", err)
	}
	if !exists {
		t.Error("ConfigExists should return true for existing file")
	}
}
