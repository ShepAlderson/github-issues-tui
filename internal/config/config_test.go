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

func TestLoadConfigWithDisplayColumns(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file with display section
	content := `repository = "myorg/myrepo"

[auth]
method = "env"

[display]
columns = ["number", "title", "author"]
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, []string{"number", "title", "author"}, cfg.Display.Columns)
}

func TestSaveConfigWithDisplayColumns(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &Config{
		Repository: "owner/repo",
		Auth: AuthConfig{
			Method: "env",
		},
		Display: DisplayConfig{
			Columns: []string{"number", "title", "date", "comments"},
		},
	}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	// Verify contents
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `columns = ["number", "title", "date", "comments"]`)
}

func TestLoadConfigWithoutDisplayColumns(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file without display section
	content := `repository = "myorg/myrepo"

[auth]
method = "env"
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Display columns should be nil (will use defaults)
	assert.Nil(t, cfg.Display.Columns)
}

func TestDefaultDisplayColumns(t *testing.T) {
	expected := []string{"number", "title", "author", "date", "comments"}
	assert.Equal(t, expected, DefaultDisplayColumns())
}

func TestValidateDisplayColumn(t *testing.T) {
	tests := []struct {
		name    string
		column  string
		wantErr bool
	}{
		{name: "number column", column: "number", wantErr: false},
		{name: "title column", column: "title", wantErr: false},
		{name: "author column", column: "author", wantErr: false},
		{name: "date column", column: "date", wantErr: false},
		{name: "comments column", column: "comments", wantErr: false},
		{name: "invalid column", column: "invalid", wantErr: true},
		{name: "empty column", column: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDisplayColumn(tt.column)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSortFieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		field   SortField
		wantErr bool
	}{
		{name: "updated", field: SortByUpdated, wantErr: false},
		{name: "created", field: SortByCreated, wantErr: false},
		{name: "number", field: SortByNumber, wantErr: false},
		{name: "comments", field: SortByComments, wantErr: false},
		{name: "invalid", field: "invalid", wantErr: true},
		{name: "empty", field: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSortField(tt.field)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSortOrderValidation(t *testing.T) {
	tests := []struct {
		name    string
		order   SortOrder
		wantErr bool
	}{
		{name: "desc", order: SortDesc, wantErr: false},
		{name: "asc", order: SortAsc, wantErr: false},
		{name: "invalid", order: "invalid", wantErr: true},
		{name: "empty", order: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSortOrder(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultSortConfig(t *testing.T) {
	sortField, sortOrder := DefaultSortConfig()

	// Default should be updated descending (most recently updated first)
	assert.Equal(t, SortByUpdated, sortField)
	assert.Equal(t, SortDesc, sortOrder)
}

func TestAllSortFields(t *testing.T) {
	fields := AllSortFields()

	expected := []SortField{SortByUpdated, SortByCreated, SortByNumber, SortByComments}
	assert.Equal(t, expected, fields)
}

func TestLoadConfigWithSortSettings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file with sort settings
	content := `repository = "myorg/myrepo"

[auth]
method = "env"

[display]
sort_field = "created"
sort_order = "asc"
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	assert.Equal(t, SortByCreated, cfg.Display.SortField)
	assert.Equal(t, SortAsc, cfg.Display.SortOrder)
}

func TestSaveConfigWithSortSettings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &Config{
		Repository: "owner/repo",
		Auth: AuthConfig{
			Method: "env",
		},
		Display: DisplayConfig{
			SortField: SortByNumber,
			SortOrder: SortDesc,
		},
	}

	err := Save(cfg, configPath)
	require.NoError(t, err)

	// Verify contents
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `sort_field = "number"`)
	assert.Contains(t, string(data), `sort_order = "desc"`)
}

func TestLoadConfigWithoutSortSettings(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a config file without sort settings
	content := `repository = "myorg/myrepo"

[auth]
method = "env"
`
	err := os.WriteFile(configPath, []byte(content), 0600)
	require.NoError(t, err)

	// Load the config
	cfg, err := Load(configPath)
	require.NoError(t, err)

	// Sort settings should be empty (will use defaults)
	assert.Empty(t, cfg.Display.SortField)
	assert.Empty(t, cfg.Display.SortOrder)
}

func TestNextSortField(t *testing.T) {
	tests := []struct {
		name     string
		current  SortField
		expected SortField
	}{
		{name: "updated to created", current: SortByUpdated, expected: SortByCreated},
		{name: "created to number", current: SortByCreated, expected: SortByNumber},
		{name: "number to comments", current: SortByNumber, expected: SortByComments},
		{name: "comments wraps to updated", current: SortByComments, expected: SortByUpdated},
		{name: "empty defaults to updated then created", current: "", expected: SortByCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NextSortField(tt.current)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToggleSortOrder(t *testing.T) {
	tests := []struct {
		name     string
		current  SortOrder
		expected SortOrder
	}{
		{name: "desc to asc", current: SortDesc, expected: SortAsc},
		{name: "asc to desc", current: SortAsc, expected: SortDesc},
		{name: "empty defaults to desc then asc", current: "", expected: SortAsc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToggleSortOrder(tt.current)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortFieldDisplayName(t *testing.T) {
	tests := []struct {
		field    SortField
		expected string
	}{
		{SortByUpdated, "Updated"},
		{SortByCreated, "Created"},
		{SortByNumber, "Number"},
		{SortByComments, "Comments"},
		{"unknown", "Updated"}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.field), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.field.DisplayName())
		})
	}
}
