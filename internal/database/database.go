package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shepbook/github-issues-tui/internal/config"
)

const defaultDatabaseName = ".ghissues.db"

// GetDatabasePath resolves the database path based on precedence:
// 1. Command-line flag (highest priority)
// 2. Config file setting
// 3. Default (.ghissues.db in current working directory)
func GetDatabasePath(cfg *config.Config, flagPath string) (string, error) {
	var path string

	// Precedence: flag > config > default
	if flagPath != "" {
		path = flagPath
	} else if cfg != nil && cfg.Database.Path != "" {
		path = cfg.Database.Path
	} else {
		// Default: .ghissues.db in current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		path = filepath.Join(cwd, defaultDatabaseName)
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		path = absPath
	}

	return path, nil
}

// InitDatabase initializes the database at the given path.
// It creates parent directories if they don't exist and validates writability.
func InitDatabase(path string) error {
	// Check if path is a directory
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return fmt.Errorf("database path is a directory: %s", path)
		}
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Check if we can write to the directory
	// We test this by trying to create a temporary file in the directory
	testFile := filepath.Join(dir, ".ghissues_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("database path is not writable: %w", err)
	}
	// Clean up test file
	os.Remove(testFile)

	return nil
}
