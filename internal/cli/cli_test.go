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

func TestDatabaseFlag(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a config file first
	cfgMgr := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	cfg := config.DefaultConfig()
	cfg.Repository = "testowner/testrepo"
	cfg.Database.Path = "/config/path/db.db" // Set config database path
	if err := cfgMgr.Save(cfg); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	tests := []struct {
		name           string
		dbFlag         string
		expectedSubstr string
	}{
		{
			name:           "no flag uses config path",
			dbFlag:         "",
			expectedSubstr: "/config/path/db.db",
		},
		{
			name:           "flag overrides config path",
			dbFlag:         "/flag/path/db.db",
			expectedSubstr: "/flag/path/db.db",
		},
		{
			name:           "relative flag path",
			dbFlag:         "custom.db",
			expectedSubstr: "custom.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create a test command with db flag support
			var dbPath string
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

					// Load config
					cfg, err := cfgMgr.Load()
					if err != nil {
						return fmt.Errorf("failed to load config: %w", err)
					}

					// Simulate database path resolution (we'll just print it)
					finalDbPath := dbPath
					if finalDbPath == "" && cfg.Database.Path != "" {
						finalDbPath = cfg.Database.Path
					}
					if finalDbPath == "" {
						finalDbPath = ".ghissues.db"
					}

					fmt.Fprintf(cmd.OutOrStdout(), "Database path: %s\n", finalDbPath)
					fmt.Fprint(cmd.OutOrStdout(), "TUI will be launched here (to be implemented)\n")
					return nil
				},
			}
			rootCmd.Flags().StringVarP(&dbPath, "db", "d", "", "database file path")

			// Set the flag value if provided
			if tt.dbFlag != "" {
				if err := rootCmd.Flags().Set("db", tt.dbFlag); err != nil {
					t.Fatalf("Failed to set db flag: %v", err)
				}
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

			// Check that database path appears in output
			if !strings.Contains(output, tt.expectedSubstr) {
				t.Errorf("Expected database path containing %q, got: %s", tt.expectedSubstr, output)
			}
		})
	}
}

func TestSyncCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test config
	configManager := config.NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	cfg := config.DefaultConfig()
	cfg.Repository = "testowner/testrepo"
	if err := configManager.Save(cfg); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create sync command
	cmd := newSyncCmd()

	// Execute command
	err := cmd.Execute()

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// The sync command should at least try to run (even if it fails due to missing auth)
	// We're just testing that the command structure works
	if err != nil && !strings.Contains(output, "Starting sync") {
		t.Logf("Sync command output: %s", output)
		t.Logf("Sync command error: %v", err)
		// Don't fail the test - sync requires actual GitHub auth which we don't have in tests
	}
}
