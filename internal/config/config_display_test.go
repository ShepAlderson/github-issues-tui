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

func TestDefaultSort(t *testing.T) {
	sort := DefaultSort()
	if sort != SortUpdated {
		t.Errorf("Expected default sort to be %q, got %q", SortUpdated, sort)
	}
}

func TestDefaultSortOrder(t *testing.T) {
	order := DefaultSortOrder()
	if order != SortOrderDesc {
		t.Errorf("Expected default sort order to be %q, got %q", SortOrderDesc, order)
	}
}

func TestAllSortOptions(t *testing.T) {
	options := AllSortOptions()
	if len(options) != 4 {
		t.Errorf("Expected 4 sort options, got %d", len(options))
	}

	expected := []SortOption{SortUpdated, SortCreated, SortNumber, SortComments}
	for i, opt := range expected {
		if options[i] != opt {
			t.Errorf("Expected sort option %d to be %q, got %q", i, opt, options[i])
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

func TestLoadWithSortOptions(t *testing.T) {
	// Create a temp config file with sort options
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

	// Write config with sort options
	configContent := `
repository = "test/repo"
auth_method = "env"
token = "test-token"

[database]
path = "test.db"

[display]
sort = "created"
sort_order = "asc"
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

	// Check sort options
	if cfg.Display.Sort != SortCreated {
		t.Errorf("Expected sort to be %q, got %q", SortCreated, cfg.Display.Sort)
	}
	if cfg.Display.SortOrder != SortOrderAsc {
		t.Errorf("Expected sort order to be %q, got %q", SortOrderAsc, cfg.Display.SortOrder)
	}
}

func TestLoadWithDefaultSortOptions(t *testing.T) {
	// Create a temp config file without sort options
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

	// Write config without sort options
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

	// Check that default sort options are applied
	if cfg.Display.Sort != DefaultSort() {
		t.Errorf("Expected default sort %q, got %q", DefaultSort(), cfg.Display.Sort)
	}
	if cfg.Display.SortOrder != DefaultSortOrder() {
		t.Errorf("Expected default sort order %q, got %q", DefaultSortOrder(), cfg.Display.SortOrder)
	}
}

func TestSaveAndLoadSortOptions(t *testing.T) {
	// Create a temp config file
	tempDir := t.TempDir()
	home := filepath.Join(tempDir, "home")
	configDir := filepath.Join(home, ".config", "ghissues")

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", oldHome)

	// Create config directory
	os.MkdirAll(configDir, 0755)

	// Create and save config with sort options
	cfg := &Config{
		Repository: "test/repo",
		AuthMethod: AuthMethodEnv,
		Token:      "test-token",
		Database:   Database{Path: "test.db"},
		Display: Display{
			Columns:   []string{"number", "title"},
			Sort:      SortComments,
			SortOrder: SortOrderAsc,
		},
	}

	err := Save(cfg)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify saved values
	if loadedCfg.Display.Sort != SortComments {
		t.Errorf("Expected sort %q, got %q", SortComments, loadedCfg.Display.Sort)
	}
	if loadedCfg.Display.SortOrder != SortOrderAsc {
		t.Errorf("Expected sort order %q, got %q", SortOrderAsc, loadedCfg.Display.SortOrder)
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme != "default" {
		t.Errorf("Expected default theme to be %q, got %q", "default", theme)
	}
}

func TestLoadWithTheme(t *testing.T) {
	// Create a temp config file with theme
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

	// Write config with theme
	configContent := `
repository = "test/repo"
auth_method = "env"
token = "test-token"

[database]
path = "test.db"

[display]
theme = "dracula"
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

	// Check theme
	if cfg.Display.Theme != "dracula" {
		t.Errorf("Expected theme to be %q, got %q", "dracula", cfg.Display.Theme)
	}
}

func TestLoadWithDefaultTheme(t *testing.T) {
	// Create a temp config file without theme
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

	// Write config without theme
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

	// Check that default theme is applied
	if cfg.Display.Theme != DefaultTheme() {
		t.Errorf("Expected default theme %q, got %q", DefaultTheme(), cfg.Display.Theme)
	}
}