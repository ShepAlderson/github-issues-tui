package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestRunConfigCommand(t *testing.T) {
	// Create a temporary config file path
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input: repo, then select option 2 (manual token), then enter token
	input := bytes.NewBufferString("testuser/testrepo\n2\nghp_testtoken123\n")
	output := &bytes.Buffer{}

	// Run the config command
	err := RunConfigCommand(configPath, input, output)
	if err != nil {
		t.Fatalf("RunConfigCommand failed: %v", err)
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

	// Verify output contains expected messages
	if !bytes.Contains(output.Bytes(), []byte("Enter GitHub repository")) {
		t.Error("Output should contain repository prompt")
	}
	if !bytes.Contains(output.Bytes(), []byte("Configuration saved")) {
		t.Error("Output should contain success message")
	}
}

func TestRunConfigCommand_MissingGhTokenEnv(t *testing.T) {
	// Ensure GITHUB_TOKEN is not set
	oldToken := os.Getenv("GITHUB_TOKEN")
	os.Unsetenv("GITHUB_TOKEN")
	defer func() {
		if oldToken != "" {
			os.Setenv("GITHUB_TOKEN", oldToken)
		}
	}()

	// Create a temporary config file path
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input: select option 2 (manual token)
	input := bytes.NewBufferString("testuser/testrepo\n2\nghp_testtoken123\n")
	output := &bytes.Buffer{}

	// Run the config command
	err := RunConfigCommand(configPath, input, output)
	if err != nil {
		t.Fatalf("RunConfigCommand failed: %v", err)
	}

	// Verify config was created
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should have been created")
	}
}

func TestRunConfigCommand_InvalidTokenChoice(t *testing.T) {
	// Create a temporary config file path
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Mock user input: try invalid choice first, then valid option 2
	input := bytes.NewBufferString("testuser/testrepo\n99\n2\nghp_testtoken123\n")
	output := &bytes.Buffer{}

	// Run the config command
	err := RunConfigCommand(configPath, input, output)
	if err != nil {
		t.Fatalf("RunConfigCommand failed: %v", err)
	}

	// Verify config was created
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config should have been created")
	}
}
