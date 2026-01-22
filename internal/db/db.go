package db

import (
	"os"
	"path/filepath"

	"github.com/shepbook/ghissues/internal/config"
)

// DefaultPath returns the default database path (.ghissues.db in current directory)
func DefaultPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".ghissues.db")
}

// GetPath determines the database path based on flag, config, or default
// Flag takes precedence over config file
func GetPath(dbFlag string, cfg *config.Config) (string, error) {
	// Flag takes precedence over config
	if dbFlag != "" {
		return dbFlag, nil
	}

	// Check config file
	if cfg != nil && cfg.Database.Path != "" {
		return cfg.Database.Path, nil
	}

	// Default to current working directory
	return DefaultPath(), nil
}

// EnsureDir creates parent directories for the database path if they don't exist
func EnsureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

// IsWritable checks if the given path is writable
func IsWritable(path string) error {
	// If path is just a filename in current directory, parent dir is "."
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		dir, _ = os.Getwd()
	}

	// Try to create the directory if it doesn't exist
	if err := EnsureDir(path); err != nil {
		return &PathError{Path: path, Err: err}
	}

	// Check if we can write to the directory
	testFile := filepath.Join(dir, ".write_test")
	defer os.Remove(testFile)

	f, err := os.Create(testFile)
	if err != nil {
		return &PathError{Path: dir, Err: err}
	}
	f.Close()

	return nil
}

// PathError represents an error with the database path
type PathError struct {
	Path string
	Err  error
}

func (e *PathError) Error() string {
	if e.Err != nil {
		return "database path not writable: " + e.Err.Error()
	}
	return "database path not writable: " + e.Path
}

func (e *PathError) Unwrap() error {
	return e.Err
}