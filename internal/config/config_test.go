package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ghissues")
	if got := ConfigPath(); got != expected {
		t.Errorf("ConfigPath() = %q, want %q", got, expected)
	}
}

func TestConfigFilePath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ghissues", "config.toml")
	if got := ConfigFilePath(); got != expected {
		t.Errorf("ConfigFilePath() = %q, want %q", got, expected)
	}
}

func TestExists(t *testing.T) {
	// Create a temporary config directory and file
	tmpDir := t.TempDir()

	// Override for test by setting HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Calculate what ConfigPath would return with the new HOME
	testConfigPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	testConfigFile := filepath.Join(testConfigPath, "config.toml")

	// Test when file doesn't exist
	if Exists() {
		t.Error("Exists() = true, want false (file doesn't exist)")
	}

	// Create the config directory and file
	if err := os.MkdirAll(testConfigPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}
	if err := os.WriteFile(testConfigFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test when file exists
	if !Exists() {
		t.Error("Exists() = false, want true (file exists)")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	// Create a temp directory that won't have the config file
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestAuthMethod_Constants(t *testing.T) {
	if AuthMethodEnv != "env" {
		t.Errorf("AuthMethodEnv = %q, want %q", AuthMethodEnv, "env")
	}
	if AuthMethodGhCli != "gh" {
		t.Errorf("AuthMethodGhCli = %q, want %q", AuthMethodGhCli, "gh")
	}
	if AuthMethodToken != "token" {
		t.Errorf("AuthMethodToken = %q, want %q", AuthMethodToken, "token")
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		Repository: "owner/repo",
		AuthMethod: AuthMethodGhCli,
		Token:      "",
	}

	if cfg.Repository != "owner/repo" {
		t.Errorf("Config.Repository = %q, want %q", cfg.Repository, "owner/repo")
	}
	if cfg.AuthMethod != AuthMethodGhCli {
		t.Errorf("Config.AuthMethod = %q, want %q", cfg.AuthMethod, AuthMethodGhCli)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	// Create and save a config
	cfg := &Config{
		Repository: "anthropics/claude-code",
		AuthMethod: AuthMethodGhCli,
		Token:      "",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify file was created with correct permissions
	configFile := filepath.Join(configPath, "config.toml")
	info, err := os.Stat(configFile)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}
	if info.Mode() != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", info.Mode())
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loaded.Repository != cfg.Repository {
		t.Errorf("Loaded Repository = %q, want %q", loaded.Repository, cfg.Repository)
	}
	if loaded.AuthMethod != cfg.AuthMethod {
		t.Errorf("Loaded AuthMethod = %q, want %q", loaded.AuthMethod, cfg.AuthMethod)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Create config directory and invalid config file
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create test config directory: %v", err)
	}

	configFile := filepath.Join(configPath, "config.toml")
	if err := os.WriteFile(configFile, []byte("invalid = toml content ["), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid TOML")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	// Create a temporary base directory
	tmpDir := t.TempDir()

	// Override HOME to point to our temp directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Dir(tmpDir))
	defer os.Setenv("HOME", oldHome)

	// Ensure the config directory doesn't exist
	configPath := filepath.Join(filepath.Dir(tmpDir), ".config", "ghissues")
	if _, err := os.Stat(configPath); err == nil {
		os.RemoveAll(configPath)
	}

	// Save should create the directory
	cfg := &Config{
		Repository: "test/repo",
		AuthMethod: AuthMethodEnv,
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config directory not created: %v", err)
	}
}
