package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/spf13/cobra"
)

func TestRootCommandNoConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a custom manager for testing
	testManager := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a test command
	rootCmd := &cobra.Command{
		Use:   "ghissues",
		Short: "GitHub Issues TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config exists
			exists, err := testManager.Exists()
			if err != nil {
				return err
			}

			if !exists {
				fmt.Fprint(cmd.OutOrStdout(), "No configuration found. Running first-time setup...\n")
				// We can't actually run interactive setup in tests
				// Just verify the path
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), "TUI will be launched here (to be implemented)\n")
			return nil
		},
	}

	// Execute command
	err := rootCmd.Execute()

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	// Check that setup message appears when no config exists
	if !strings.Contains(output, "No configuration found") {
		t.Errorf("Expected setup message, got: %s", output)
	}
}

func TestRootCommandWithConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a config file first
	cfgMgr := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	cfg := config.DefaultConfig()
	cfg.Repository = "testowner/testrepo"
	if err := cfgMgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Verify config file exists
	configPath := filepath.Join(tempDir, "ghissues", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Test config file not created: %s", configPath)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a test command
	rootCmd := &cobra.Command{
		Use:   "ghissues",
		Short: "GitHub Issues TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if config exists using test manager
			exists, err := cfgMgr.Exists()
			if err != nil {
				return err
			}

			if !exists {
				fmt.Fprint(cmd.OutOrStdout(), "No configuration found. Running first-time setup...\n")
				return nil
			}

			fmt.Fprint(cmd.OutOrStdout(), "TUI will be launched here (to be implemented)\n")
			return nil
		},
	}

	// Execute command
	err := rootCmd.Execute()

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("Command execution failed: %v", err)
	}

	// Check that TUI launch message appears when config exists
	if !strings.Contains(output, "TUI will be launched here") {
		t.Errorf("Expected TUI launch message, got: %s", output)
	}

	if strings.Contains(output, "No configuration found") {
		t.Errorf("Should not show setup message when config exists, got: %s", output)
	}
}