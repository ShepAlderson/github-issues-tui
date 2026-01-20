package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
)

func TestDefaultPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(cwd, ".ghissues.db")
	got := DefaultPath()

	if got != want {
		t.Errorf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestGetPath_FlagTakesPrecedence(t *testing.T) {
	cfg := &config.Config{
		Database: config.Database{Path: "/config/path.db"},
	}

	got, err := GetPath("/flag/path.db", cfg)
	if err != nil {
		t.Fatalf("GetPath() error = %v", err)
	}

	want := "/flag/path.db"
	if got != want {
		t.Errorf("GetPath() = %q, want %q", got, want)
	}
}

func TestGetPath_ConfigFallback(t *testing.T) {
	cfg := &config.Config{
		Database: config.Database{Path: "/config/path.db"},
	}

	got, err := GetPath("", cfg)
	if err != nil {
		t.Fatalf("GetPath() error = %v", err)
	}

	want := "/config/path.db"
	if got != want {
		t.Errorf("GetPath() = %q, want %q", got, want)
	}
}

func TestGetPath_DefaultWhenNoConfig(t *testing.T) {
	cfg := &config.Config{
		Database: config.Database{Path: ""},
	}

	got, err := GetPath("", cfg)
	if err != nil {
		t.Fatalf("GetPath() error = %v", err)
	}

	want := DefaultPath()
	if got != want {
		t.Errorf("GetPath() = %q, want %q", got, want)
	}
}

func TestGetPath_NilConfig(t *testing.T) {
	got, err := GetPath("", nil)
	if err != nil {
		t.Fatalf("GetPath() error = %v", err)
	}

	want := DefaultPath()
	if got != want {
		t.Errorf("GetPath() = %q, want %q", got, want)
	}
}

func TestEnsureDir_CreatesParent(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "subdir", "subdir2", "test.db")

	err := EnsureDir(testPath)
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(filepath.Dir(testPath))
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory, got file")
	}
}

func TestEnsureDir_EmptyPath(t *testing.T) {
	err := EnsureDir("test.db")
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}
}

func TestEnsureDir_DotPath(t *testing.T) {
	err := EnsureDir("./test.db")
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}
}

func TestIsWritable_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, ".ghissues.db")

	err := IsWritable(dbPath)
	if err != nil {
		t.Errorf("IsWritable() error = %v", err)
	}
}

func TestIsWritable_CreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "nested", ".ghissues.db")

	err := IsWritable(dbPath)
	if err != nil {
		t.Errorf("IsWritable() error = %v", err)
	}

	// Verify the nested directory was created
	nestedDir := filepath.Dir(dbPath)
	if _, err := os.Stat(nestedDir); err != nil {
		t.Errorf("expected directory to be created, got error: %v", err)
	}
}

func TestIsWritable_UnwritablePath(t *testing.T) {
	// Create a read-only directory to test unwritable path
	tmpDir := t.TempDir()
	readonlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readonlyDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(readonlyDir, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(readonlyDir, 0755) // Restore for cleanup

	dbPath := filepath.Join(readonlyDir, ".ghissues.db")

	err := IsWritable(dbPath)
	if err == nil {
		t.Error("IsWritable() expected error for read-only directory, got nil")
	}
}

func TestPathError(t *testing.T) {
	err := &PathError{Path: "/bad/path", Err: os.ErrPermission}
	want := "database path not writable: permission denied"
	if got := err.Error(); got != want {
		t.Errorf("PathError.Error() = %q, want %q", got, want)
	}
}