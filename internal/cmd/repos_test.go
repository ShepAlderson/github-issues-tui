package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReposCommandExists(t *testing.T) {
	cmd := NewRootCmd()

	// Look for repos subcommand
	var reposCmd *bool
	for _, c := range cmd.Commands() {
		if c.Use == "repos" {
			reposCmd = new(bool)
			*reposCmd = true
			break
		}
	}

	require.NotNil(t, reposCmd, "repos subcommand should exist")
}

func TestReposCommandListsRepositories(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with multiple repositories
	cfg := &config.Config{
		DefaultRepository: "owner/default",
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/default", DatabasePath: "/db/default.db"},
			{Name: "owner/other", DatabasePath: "/db/other.db"},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "owner/default")
	assert.Contains(t, output, "owner/other")
}

func TestReposCommandShowsDefaultIndicator(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with default repository
	cfg := &config.Config{
		DefaultRepository: "owner/default",
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/default", DatabasePath: "/db/default.db"},
			{Name: "owner/other", DatabasePath: "/db/other.db"},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Default repo should be marked somehow (e.g., with asterisk or "(default)")
	assert.Contains(t, output, "default")
}

func TestReposCommandShowsEmptyMessage(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with no repositories
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No repositories configured")
}

func TestReposCommandAddFlag(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create initial config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/existing", DatabasePath: "/db/existing.db"},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos", "--add", "owner/new"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Reload config and verify new repo was added
	cfg, err = config.Load(cfgPath)
	require.NoError(t, err)
	require.Len(t, cfg.Repositories, 2)
	assert.Equal(t, "owner/new", cfg.Repositories[1].Name)
}

func TestReposCommandAddFlagWithDBPath(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create initial config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos", "--add", "owner/repo", "--db-path", "/custom/path.db"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Reload config and verify
	cfg, err = config.Load(cfgPath)
	require.NoError(t, err)
	require.Len(t, cfg.Repositories, 1)
	assert.Equal(t, "/custom/path.db", cfg.Repositories[0].DatabasePath)
}

func TestReposCommandSetDefaultFlag(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with multiple repositories but no default
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/repo1", DatabasePath: "/db/repo1.db"},
			{Name: "owner/repo2", DatabasePath: "/db/repo2.db"},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos", "--set-default", "owner/repo2"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	// Reload config and verify default was set
	cfg, err = config.Load(cfgPath)
	require.NoError(t, err)
	assert.Equal(t, "owner/repo2", cfg.DefaultRepository)
}

func TestReposCommandSetDefaultFlagInvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with one repository
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/existing", DatabasePath: "/db/existing.db"},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"repos", "--set-default", "owner/nonexistent"})

	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRootCommandRepoFlag(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with multiple repositories
	cfg := &config.Config{
		DefaultRepository: "owner/default",
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/default", DatabasePath: filepath.Join(tmpDir, "default.db")},
			{Name: "owner/selected", DatabasePath: filepath.Join(tmpDir, "selected.db")},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")
	SetDisableTUI(true)
	defer SetDisableTUI(false)
	defer SetDBPath("")
	defer SetRepoFlag("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--repo", "owner/selected"})

	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should show selected repo, not default
	assert.Contains(t, output, "owner/selected")
}

func TestRootCommandRepoFlagInvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create config with one repository
	cfg := &config.Config{
		Auth: config.AuthConfig{
			Method: "env",
		},
		Repositories: []config.RepositoryConfig{
			{Name: "owner/existing", DatabasePath: filepath.Join(tmpDir, "existing.db")},
		},
	}
	err := config.Save(cfg, cfgPath)
	require.NoError(t, err)

	SetConfigPath(cfgPath)
	defer SetConfigPath("")
	SetDisableTUI(true)
	defer SetDisableTUI(false)
	defer SetRepoFlag("")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--repo", "owner/nonexistent"})

	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
