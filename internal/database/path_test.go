package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name       string
		flagPath   string
		configPath string
		want       string
	}{
		{
			name:       "default path when neither flag nor config provided",
			flagPath:   "",
			configPath: "",
			want:       ".ghissues.db",
		},
		{
			name:       "config path used when no flag provided",
			flagPath:   "",
			configPath: "/custom/path/issues.db",
			want:       "/custom/path/issues.db",
		},
		{
			name:       "flag takes precedence over config",
			flagPath:   "/flag/path/flag.db",
			configPath: "/config/path/config.db",
			want:       "/flag/path/flag.db",
		},
		{
			name:       "flag alone used when config empty",
			flagPath:   "/flag/path/flag.db",
			configPath: "",
			want:       "/flag/path/flag.db",
		},
		{
			name:       "flag takes precedence even if config set",
			flagPath:   "./custom.db",
			configPath: "~/.config/ghissues/issues.db",
			want:       "./custom.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePath(tt.flagPath, tt.configPath)
			if got != tt.want {
				t.Errorf("ResolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureWritable(t *testing.T) {
	t.Run("creates parent directories if they don't exist", func(t *testing.T) {
		// Create a temp directory
		tempDir := t.TempDir()
		nestedPath := filepath.Join(tempDir, "nested", "deep", "test.db")

		err := EnsureWritable(nestedPath)
		if err != nil {
			t.Errorf("EnsureWritable() error = %v, want nil", err)
		}

		// Verify directory was created
		statDir := filepath.Join(tempDir, "nested", "deep")
		if _, err := os.Stat(statDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist, but it doesn't", statDir)
		}
	})

	t.Run("succeeds when path already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "existing.db")

		// Create the file first
		file, err := os.Create(dbPath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		file.Close()

		err = EnsureWritable(dbPath)
		if err != nil {
			t.Errorf("EnsureWritable() error = %v, want nil", err)
		}
	})

	t.Run("returns error when path is not writable", func(t *testing.T) {
		// This test is tricky on different platforms
		// On Unix, we can test with a read-only directory
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")

		// Create the directory
		if err := os.MkdirAll(readOnlyDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Try to make it read-only (may not work on all platforms)
		if err := os.Chmod(readOnlyDir, 0555); err != nil {
			t.Skip("Cannot set read-only permissions, skipping test")
		}

		// Restore permissions after test
		defer os.Chmod(readOnlyDir, 0755)

		dbPath := filepath.Join(readOnlyDir, "test.db")
		err := EnsureWritable(dbPath)
		if err == nil {
			t.Errorf("EnsureWritable() error = nil, want error for non-writable path")
		}
	})

	t.Run("succeeds with relative path", func(t *testing.T) {
		tempDir := t.TempDir()

		// Change to temp directory temporarily
		origDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origDir)

		err := EnsureWritable("relative.db")
		if err != nil {
			t.Errorf("EnsureWritable() error = %v, want nil", err)
		}

		// Verify directory exists
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			t.Errorf("Directory should exist")
		}
	})
}
