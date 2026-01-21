package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()

	// Should be in user's home directory under .config/ghissues
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(homeDir, ".config", "ghissues", "config.toml")
	assert.Equal(t, expected, path)
}

func TestNewConfig(t *testing.T) {
	cfg := New()

	// Should have default values
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Repository)
	assert.Empty(t, cfg.Auth.Token)
	assert.Equal(t, "env", cfg.Auth.Method) // Default to env method
}

func TestConfigExists(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Should return false when file doesn't exist
	assert.False(t, Exists(configPath))

	// Create the file
	err := os.WriteFile(configPath, []byte("[auth]\nmethod = \"env\"\n"), 0600)
	require.NoError(t, err)

	// Should return true when file exists
	assert.True(t, Exists(configPath))
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &Config{
		Repository: "owner/repo",
		Auth: AuthConfig{
			Method: "env",
		},
	}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	// Verify file was created
	assert.True(t, Exists(configPath))

	// Verify file permissions are secure (0600)
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify contents
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `repository = "owner/repo"`)
	assert.Contains(t, string(data), `method = "env"`)
}

func TestSaveConfigCreatesParentDirs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "deep", "config.toml")

	cfg := &Config{
		Repository: "owner/repo",
		Auth: AuthConfig{
			Method: "env",
		},
	}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	// Verify file was created in nested directory
	assert.True(t, Exists(configPath))
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file
	content := `repository = "myorg/myrepo"

[auth]
method = "token"
token = "ghp_secret123"
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "myorg/myrepo", cfg.Repository)
	assert.Equal(t, "token", cfg.Auth.Method)
	assert.Equal(t, "ghp_secret123", cfg.Auth.Token)
}

func TestLoadConfigFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.toml")

	_, err := Load(configPath)
	assert.Error(t, err)
}

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		name       string
		repo       string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "valid repository",
			repo:    "owner/repo",
			wantErr: false,
		},
		{
			name:    "valid repository with dashes",
			repo:    "my-org/my-repo",
			wantErr: false,
		},
		{
			name:    "valid repository with numbers",
			repo:    "org123/repo456",
			wantErr: false,
		},
		{
			name:       "missing slash",
			repo:       "ownerrepo",
			wantErr:    true,
			errMessage: "must be in owner/repo format",
		},
		{
			name:       "empty string",
			repo:       "",
			wantErr:    true,
			errMessage: "repository cannot be empty",
		},
		{
			name:       "too many slashes",
			repo:       "owner/repo/extra",
			wantErr:    true,
			errMessage: "must be in owner/repo format",
		},
		{
			name:       "empty owner",
			repo:       "/repo",
			wantErr:    true,
			errMessage: "owner cannot be empty",
		},
		{
			name:       "empty repo name",
			repo:       "owner/",
			wantErr:    true,
			errMessage: "repository name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepository(tt.repo)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAuthMethod(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		wantErr bool
	}{
		{name: "env method", method: "env", wantErr: false},
		{name: "token method", method: "token", wantErr: false},
		{name: "gh method", method: "gh", wantErr: false},
		{name: "invalid method", method: "invalid", wantErr: true},
		{name: "empty method", method: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthMethod(tt.method)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadConfigWithDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file with database section
	content := `repository = "myorg/myrepo"

[auth]
method = "env"

[database]
path = "/custom/path/issues.db"
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, "/custom/path/issues.db", cfg.Database.Path)
}

func TestSaveConfigWithDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &Config{
		Repository: "owner/repo",
		Auth: AuthConfig{
			Method: "env",
		},
		Database: DatabaseConfig{
			Path: "/my/custom/db.db",
		},
	}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	// Verify contents
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `path = "/my/custom/db.db"`)
}

func TestLoadConfigWithoutDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file without database section
	content := `repository = "myorg/myrepo"

[auth]
method = "env"
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Database path should be empty (will use default)
	assert.Empty(t, cfg.Database.Path)
}
