package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestRunMain_NoConfigFile(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input for setup
	input := bytes.NewBufferString("testuser/testrepo\n2\nghp_testtoken123\n")
	output := &bytes.Buffer{}

	// Run main with no config file (should trigger setup)
	err := runMain([]string{"ghissues"}, configPath, input, output)
	if err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// Verify config was created
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should have been created")
	}

	if cfg.Repository != "testuser/testrepo" {
		t.Errorf("Expected repository 'testuser/testrepo', got %q", cfg.Repository)
	}
	if cfg.Token != "ghp_testtoken123" {
		t.Errorf("Expected token 'ghp_testtoken123', got %q", cfg.Token)
	}

	// Check output contains setup messages
	if !bytes.Contains(output.Bytes(), []byte("Enter GitHub repository")) {
		t.Error("Output should contain setup prompt")
	}
}

func TestRunMain_ConfigCommand(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input for config command
	input := bytes.NewBufferString("anotheruser/anotherrepo\n2\nghp_anothertoken\n")
	output := &bytes.Buffer{}

	// Run with config command
	err := runMain([]string{"ghissues", "config"}, configPath, input, output)
	if err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// Verify config was created
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should have been created")
	}

	if cfg.Repository != "anotheruser/anotherrepo" {
		t.Errorf("Expected repository 'anotheruser/anotherrepo', got %q", cfg.Repository)
	}
}

func TestRunMain_ConfigAlreadyExists(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create an existing config file
	cfg := &config.Config{
		Repository: "existing/repo",
		Token:      "ghp_existing",
	}
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Run main with existing config (should skip setup)
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	err := runMain([]string{"ghissues"}, configPath, input, output)
	if err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// Verify config wasn't changed
	loadedCfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Repository != "existing/repo" {
		t.Errorf("Expected repository 'existing/repo', got %q", loadedCfg.Repository)
	}
	if loadedCfg.Token != "ghp_existing" {
		t.Errorf("Expected token 'ghp_existing', got %q", loadedCfg.Token)
	}

	// Check that setup was skipped (no prompts in output)
	if bytes.Contains(output.Bytes(), []byte("Enter GitHub repository")) {
		t.Error("Output should not contain setup prompt when config exists")
	}
}

func TestRunMain_TooManyArgs(t *testing.T) {
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Run with too many arguments
	input := bytes.NewBufferString("")
	output := &bytes.Buffer{}

	err := runMain([]string{"ghissues", "config", "extra"}, configPath, input, output)
	if err == nil {
		t.Fatal("Expected error for too many arguments")
	}

	if err.Error() != "too many arguments" {
		t.Errorf("Expected 'too many arguments' error, got: %v", err)
	}
}

func TestGetConfigFilePath_FromEnv(t *testing.T) {
	// Test getting config path from environment variable
	oldEnv := os.Getenv("GHISSUES_CONFIG")
	defer os.Setenv("GHISSUES_CONFIG", oldEnv)

	os.Setenv("GHISSUES_CONFIG", "/custom/path/config.toml")

	path := getConfigFilePath()
	if path != "/custom/path/config.toml" {
		t.Errorf("Expected '/custom/path/config.toml', got %q", path)
	}
}

func TestGetConfigFilePath_Default(t *testing.T) {
	// Test getting default config path
	oldEnv := os.Getenv("GHISSUES_CONFIG")
	defer os.Setenv("GHISSUES_CONFIG", oldEnv)

	os.Unsetenv("GHISSUES_CONFIG")

	path := getConfigFilePath()
	expectedSuffix := ".config/ghissues/config.toml"
	if len(path) < len(expectedSuffix) || path[len(path)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("Config path should end with %q, got %q", expectedSuffix, path)
	}
}
