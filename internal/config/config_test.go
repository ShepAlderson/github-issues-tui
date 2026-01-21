package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if cfg.Repository != "" {
		t.Errorf("DefaultConfig.Repository should be empty, got %s", cfg.Repository)
	}

	if cfg.Auth.Method != "gh" {
		t.Errorf("DefaultConfig.Auth.Method should be 'gh', got %s", cfg.Auth.Method)
	}

	if cfg.Database.Path != ".ghissues.db" {
		t.Errorf("DefaultConfig.Database.Path should be '.ghissues.db', got %s", cfg.Database.Path)
	}

	if cfg.Display.Theme != "default" {
		t.Errorf("DefaultConfig.Display.Theme should be 'default', got %s", cfg.Display.Theme)
	}
}

func TestManagerConfigDir(t *testing.T) {
	tempDir := t.TempDir()
	m := NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	dir, err := m.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir failed: %v", err)
	}

	expected := filepath.Join(tempDir, "ghissues")
	if dir != expected {
		t.Errorf("ConfigDir returned %s, want %s", dir, expected)
	}
}

func TestManagerConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	m := NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	path, err := m.ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}

	expected := filepath.Join(tempDir, "ghissues", "config.toml")
	if path != expected {
		t.Errorf("ConfigPath returned %s, want %s", path, expected)
	}
}

func TestManagerSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	m := NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	cfg := DefaultConfig()
	cfg.Repository = "testowner/testrepo"
	cfg.Auth.Method = "token"
	cfg.Auth.Token = "test-token-123"

	// Save the config
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load the config
	loadedCfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded values match saved values
	if loadedCfg.Repository != cfg.Repository {
		t.Errorf("Loaded Repository = %s, want %s", loadedCfg.Repository, cfg.Repository)
	}

	if loadedCfg.Auth.Method != cfg.Auth.Method {
		t.Errorf("Loaded Auth.Method = %s, want %s", loadedCfg.Auth.Method, cfg.Auth.Method)
	}

	if loadedCfg.Auth.Token != cfg.Auth.Token {
		t.Errorf("Loaded Auth.Token = %s, want %s", loadedCfg.Auth.Token, cfg.Auth.Token)
	}
}

func TestManagerLoadNonExistentConfig(t *testing.T) {
	tempDir := t.TempDir()
	m := NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	// Try to load non-existent config
	_, err := m.Load()
	if err == nil {
		t.Fatal("Load should fail for non-existent config")
	}

	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

func TestManagerExists(t *testing.T) {
	tempDir := t.TempDir()
	m := NewTestManager(func() (string, error) {
		return tempDir, nil
	})

	// Config should not exist initially
	exists, err := m.Exists()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Exists should return false for non-existent config")
	}

	// Save a config
	cfg := DefaultConfig()
	if err := m.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Config should exist now
	exists, err = m.Exists()
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Exists should return true for existing config")
	}
}