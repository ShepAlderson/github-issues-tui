package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
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

func TestDatabasePath_Default(t *testing.T) {
	// Test that default database path is .ghissues.db in current directory
	cfg := &config.Config{}

	dbPath, err := database.GetDatabasePath(cfg, "")
	if err != nil {
		t.Fatalf("GetDatabasePath() failed: %v", err)
	}

	// Should be absolute path
	if !filepath.IsAbs(dbPath) {
		t.Errorf("Expected absolute path, got: %s", dbPath)
	}

	// Should end with .ghissues.db
	expectedSuffix := ".ghissues.db"
	if len(dbPath) < len(expectedSuffix) || dbPath[len(dbPath)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("Expected path to end with %s, got: %s", expectedSuffix, dbPath)
	}
}

func TestDatabasePath_ConfigOverridesDefault(t *testing.T) {
	// Test that config file path takes precedence over default
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: "/custom/location/issues.db",
		},
	}

	dbPath, err := database.GetDatabasePath(cfg, "")
	if err != nil {
		t.Fatalf("GetDatabasePath() failed: %v", err)
	}

	if dbPath != "/custom/location/issues.db" {
		t.Errorf("Expected /custom/location/issues.db, got: %s", dbPath)
	}
}

func TestDatabasePath_FlagOverridesConfig(t *testing.T) {
	// Test that --db flag takes precedence over config file
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: "/config/location/issues.db",
		},
	}

	flagPath := "/flag/location/issues.db"
	dbPath, err := database.GetDatabasePath(cfg, flagPath)
	if err != nil {
		t.Fatalf("GetDatabasePath() failed: %v", err)
	}

	if dbPath != flagPath {
		t.Errorf("Expected flag path %s, got: %s", flagPath, dbPath)
	}
}

func TestDatabaseInit_CreatesDirectories(t *testing.T) {
	// Test that parent directories are created if they don't exist
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "nested", "path", "test.db")

	err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitDatabase() failed: %v", err)
	}

	// Verify parent directories were created
	parentDir := filepath.Dir(dbPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Error("Parent directories were not created")
	}
}

func TestDatabaseInit_Writable(t *testing.T) {
	// Test that database location is writable
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("InitDatabase() failed: %v", err)
	}

	// Test writability by creating the database file
	testData := []byte("test")
	if err := os.WriteFile(dbPath, testData, 0644); err != nil {
		t.Errorf("Cannot write to database path: %v", err)
	}
}

func TestDatabaseInit_ErrorOnNonWritable(t *testing.T) {
	// Test that initialization fails if path is not writable
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	tempDir := t.TempDir()
	readonlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readonlyDir, 0555); err != nil {
		t.Fatalf("Failed to create readonly directory: %v", err)
	}

	dbPath := filepath.Join(readonlyDir, "test.db")
	err := database.InitDatabase(dbPath)
	if err == nil {
		t.Error("Expected error for non-writable path, got nil")
	}
}

func TestDatabaseInit_ErrorOnDirectory(t *testing.T) {
	// Test that initialization fails if path points to a directory
	tempDir := t.TempDir()
	dirPath := filepath.Join(tempDir, "is-a-dir")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	err := database.InitDatabase(dirPath)
	if err == nil {
		t.Error("Expected error when database path is a directory, got nil")
	}
}
