package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDBPath(t *testing.T) {
	// Default should be .ghissues.db in current working directory
	path := DefaultDBPath()
	assert.Equal(t, ".ghissues.db", path)
}

func TestResolveDBPath_DefaultWhenNoOverrides(t *testing.T) {
	cfg := &config.Config{}

	path, err := ResolveDBPath("", cfg)
	require.NoError(t, err)
	assert.Equal(t, ".ghissues.db", path)
}

func TestResolveDBPath_ConfigOverridesDefault(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: "/custom/path/mydb.db",
		},
	}

	path, err := ResolveDBPath("", cfg)
	require.NoError(t, err)
	assert.Equal(t, "/custom/path/mydb.db", path)
}

func TestResolveDBPath_FlagOverridesConfig(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Path: "/config/path/db.db",
		},
	}

	path, err := ResolveDBPath("/flag/path/db.db", cfg)
	require.NoError(t, err)
	assert.Equal(t, "/flag/path/db.db", path)
}

func TestResolveDBPath_FlagOverridesDefault(t *testing.T) {
	cfg := &config.Config{}

	path, err := ResolveDBPath("/flag/path/db.db", cfg)
	require.NoError(t, err)
	assert.Equal(t, "/flag/path/db.db", path)
}

func TestResolveDBPath_NilConfig(t *testing.T) {
	// Should work with nil config (use default)
	path, err := ResolveDBPath("", nil)
	require.NoError(t, err)
	assert.Equal(t, ".ghissues.db", path)
}

func TestResolveDBPath_FlagWithNilConfig(t *testing.T) {
	// Flag should work even with nil config
	path, err := ResolveDBPath("/my/db.db", nil)
	require.NoError(t, err)
	assert.Equal(t, "/my/db.db", path)
}

func TestEnsureDBPath_CreatesParentDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nested", "deep", "dir", "test.db")

	err := EnsureDBPath(dbPath)
	require.NoError(t, err)

	// Verify parent directories were created
	parentDir := filepath.Dir(dbPath)
	info, err := os.Stat(parentDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestEnsureDBPath_WorksWithExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	err := EnsureDBPath(dbPath)
	require.NoError(t, err)
}

func TestEnsureDBPath_ErrorOnUnwritablePath(t *testing.T) {
	// Skip on CI or if running as root (root can write anywhere)
	if os.Getuid() == 0 {
		t.Skip("Test skipped when running as root")
	}

	// Try to use a path that's definitely not writable
	dbPath := "/root/unauthorized/path/test.db"

	err := EnsureDBPath(dbPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not writable")
}

func TestEnsureDBPath_ErrorMessageIsHelpful(t *testing.T) {
	// Skip on CI or if running as root
	if os.Getuid() == 0 {
		t.Skip("Test skipped when running as root")
	}

	dbPath := "/root/cannot/write/here/test.db"

	err := EnsureDBPath(dbPath)
	require.Error(t, err)
	// Error should mention the path that's not writable
	assert.Contains(t, err.Error(), "not writable")
}

func TestEnsureDBPath_CurrentDirRelativePath(t *testing.T) {
	// Test with relative path (current directory)
	// This should work as we can write to temp directory
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	err = EnsureDBPath(".ghissues.db")
	require.NoError(t, err)
}

func TestEnsureDBPath_ExpandsHomeDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with actual path in tmpDir
	dbPath := filepath.Join(tmpDir, "testuser", "ghissues", "data.db")

	err := EnsureDBPath(dbPath)
	require.NoError(t, err)

	// Verify directories were created
	parentDir := filepath.Dir(dbPath)
	_, err = os.Stat(parentDir)
	require.NoError(t, err)
}

func TestIsPathWritable(t *testing.T) {
	tmpDir := t.TempDir()

	// Writable path
	assert.True(t, IsPathWritable(tmpDir))

	// Non-existent parent that can be created
	newDir := filepath.Join(tmpDir, "newdir")
	assert.True(t, IsPathWritable(newDir))
}

func TestIsPathWritable_NonWritablePath(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Test skipped when running as root")
	}

	assert.False(t, IsPathWritable("/root/unauthorized"))
}
