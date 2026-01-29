package database

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolvePath determines the database path based on priority:
// 1. --db flag (if provided)
// 2. database.path from config file
// 3. Default: .ghissues.db in current working directory
func ResolvePath(flagPath, configPath string) string {
	// Priority 1: --db flag
	if flagPath != "" {
		return flagPath
	}

	// Priority 2: config file path
	if configPath != "" {
		return configPath
	}

	// Priority 3: default
	return ".ghissues.db"
}

// EnsureWritable checks if the database path is writable
// Creates parent directories if they don't exist
// Returns an error if the path is not writable
func EnsureWritable(dbPath string) error {
	// Get the absolute path
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Get parent directory
	dir := filepath.Dir(absPath)

	// Create parent directories if they don't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Check if directory is writable by trying to create a temp file
	testFile := filepath.Join(dir, ".write_test")
	file, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("database path is not writable: %w", err)
	}
	file.Close()
	os.Remove(testFile)

	return nil
}
