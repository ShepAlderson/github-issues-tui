package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestRunReposCommand(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a multi-repo config
	cfg := &config.Config{
		DefaultRepo: "owner/repo2",
		Repositories: map[string]config.RepositoryConfig{
			"owner/repo1": {DatabasePath: "/path/to/db1.db"},
			"owner/repo2": {DatabasePath: "/path/to/db2.db"},
			"owner/repo3": {DatabasePath: "/path/to/db3.db"},
		},
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Run the repos command
	output := &bytes.Buffer{}
	err := RunReposCommand(configPath, output)
	if err != nil {
		t.Fatalf("RunReposCommand failed: %v", err)
	}

	// Verify output contains all repositories
	if !bytes.Contains(output.Bytes(), []byte("owner/repo1")) {
		t.Error("Output should contain owner/repo1")
	}
	if !bytes.Contains(output.Bytes(), []byte("owner/repo2")) {
		t.Error("Output should contain owner/repo2")
	}
	if !bytes.Contains(output.Bytes(), []byte("owner/repo3")) {
		t.Error("Output should contain owner/repo3")
	}

	// Verify default repository is marked
	if !bytes.Contains(output.Bytes(), []byte("(default)")) {
		t.Error("Output should mark the default repository")
	}
}

func TestRunReposCommand_BackwardCompatibility(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a single-repo config (old format)
	cfg := &config.Config{
		Repository: "testuser/testrepo",
		Database: struct {
			Path string `toml:"path"`
		}{
			Path: "/path/to/single.db",
		},
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Run the repos command
	output := &bytes.Buffer{}
	err := RunReposCommand(configPath, output)
	if err != nil {
		t.Fatalf("RunReposCommand failed: %v", err)
	}

	// Verify output contains the single repository
	if !bytes.Contains(output.Bytes(), []byte("testuser/testrepo")) {
		t.Error("Output should contain testuser/testrepo")
	}

	// Verify it's marked as default
	if !bytes.Contains(output.Bytes(), []byte("(default)")) {
		t.Error("Output should mark the repository as default")
	}
}

func TestRunReposCommand_EmptyConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create an empty config
	cfg := &config.Config{}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Run the repos command
	output := &bytes.Buffer{}
	err := RunReposCommand(configPath, output)
	if err != nil {
		t.Fatalf("RunReposCommand failed: %v", err)
	}

	// Verify output indicates no repositories configured
	if !bytes.Contains(output.Bytes(), []byte("No repositories")) {
		t.Error("Output should indicate no repositories configured")
	}
}

func TestRunReposCommand_NoConfigFile(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.toml")

	// Run the repos command
	output := &bytes.Buffer{}
	err := RunReposCommand(configPath, output)

	// Should return an error
	if err == nil {
		t.Fatal("RunReposCommand should fail when config file doesn't exist")
	}

	if !bytes.Contains(output.Bytes(), []byte("not found")) {
		t.Error("Output should indicate config file not found")
	}
}
