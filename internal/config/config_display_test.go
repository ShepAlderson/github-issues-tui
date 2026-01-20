package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultColumns(t *testing.T) {
	cols := DefaultColumns()
	if len(cols) != 5 {
		t.Errorf("Expected 5 default columns, got %d", len(cols))
	}

	expected := []string{"number", "title", "author", "date", "comments"}
	for i, col := range expected {
		if cols[i] != col {
			t.Errorf("Expected column %d to be %q, got %q", i, col, cols[i])
		}
	}
}

func TestLoadWithDisplayColumns(t *testing.T) {
	// Create a temp config file with display columns
	tempDir := t.TempDir()
	home := filepath.Join(tempDir, "home")
	configDir := filepath.Join(home, ".config", "ghissues")
	configFile := filepath.Join(configDir, "config.toml")

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	// Create config directory
	os.MkdirAll(configDir, 0755)

	// Write config with display columns
	configContent := `
repository = "test/repo"
auth_method = "env"
token = "test-token"

[database]
path = "test.db"

[display]
columns = ["number", "title", "author"]
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check display columns
	if len(cfg.Display.Columns) != 3 {
		t.Errorf("Expected 3 display columns, got %d", len(cfg.Display.Columns))
	}

	expectedCols := []string{"number", "title", "author"}
	for i, col := range expectedCols {
		if cfg.Display.Columns[i] != col {
			t.Errorf("Expected column %d to be %q, got %q", i, col, cfg.Display.Columns[i])
		}
	}
}

func TestLoadWithDefaultColumns(t *testing.T) {
	// Create a temp config file without display columns
	tempDir := t.TempDir()
	home := filepath.Join(tempDir, "home")
	configDir := filepath.Join(home, ".config", "ghissues")
	configFile := filepath.Join(configDir, "config.toml")

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	// Create config directory
	os.MkdirAll(configDir, 0755)

	// Write config without display columns
	configContent := `
repository = "test/repo"
auth_method = "env"
token = "test-token"

[database]
path = "test.db"
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check that default columns are applied
	if len(cfg.Display.Columns) != 5 {
		t.Errorf("Expected 5 default display columns, got %d", len(cfg.Display.Columns))
	}

	// Verify it's the same as DefaultColumns()
	defaultCols := DefaultColumns()
	for i, col := range defaultCols {
		if cfg.Display.Columns[i] != col {
			t.Errorf("Expected column %d to be %q, got %q", i, col, cfg.Display.Columns[i])
		}
	}
}