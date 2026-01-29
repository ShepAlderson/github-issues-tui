package main

import (
	"testing"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/list"
)

func TestConfigAdapter_GetRepositories(t *testing.T) {
	t.Run("returns empty when no repositories", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{},
		}
		adapter := &ConfigAdapter{cfg: cfg}

		repos := adapter.GetRepositories()
		if len(repos) != 0 {
			t.Errorf("expected 0 repositories, got %d", len(repos))
		}
	})

	t.Run("returns configured repositories", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{
				{Owner: "owner1", Name: "repo1", Database: "/path/to/repo1.db"},
				{Owner: "owner2", Name: "repo2", Database: "/path/to/repo2.db"},
			},
		}
		adapter := &ConfigAdapter{cfg: cfg}

		repos := adapter.GetRepositories()
		if len(repos) != 2 {
			t.Errorf("expected 2 repositories, got %d", len(repos))
		}

		if repos[0].FullName != "owner1/repo1" {
			t.Errorf("expected first repo 'owner1/repo1', got %q", repos[0].FullName)
		}

		if repos[1].FullName != "owner2/repo2" {
			t.Errorf("expected second repo 'owner2/repo2', got %q", repos[1].FullName)
		}
	})

	t.Run("repository info contains correct fields", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{
				{Owner: "testowner", Name: "testrepo", Database: "/path/to/db.db"},
			},
		}
		adapter := &ConfigAdapter{cfg: cfg}

		repos := adapter.GetRepositories()
		if len(repos) != 1 {
			t.Fatal("expected 1 repository")
		}

		repo := repos[0]
		if repo.Owner != "testowner" {
			t.Errorf("expected Owner 'testowner', got %q", repo.Owner)
		}
		if repo.Name != "testrepo" {
			t.Errorf("expected Name 'testrepo', got %q", repo.Name)
		}
		if repo.FullName != "testowner/testrepo" {
			t.Errorf("expected FullName 'testowner/testrepo', got %q", repo.FullName)
		}
	})
}

func TestConfigAdapter_GetRepositoryDatabase(t *testing.T) {
	t.Run("returns empty when repo not found", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{},
		}
		adapter := &ConfigAdapter{cfg: cfg}

		dbPath := adapter.GetRepositoryDatabase("owner/nonexistent")
		if dbPath != "" {
			t.Errorf("expected empty dbPath, got %q", dbPath)
		}
	})

	t.Run("returns database path for repository", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{
				{Owner: "owner1", Name: "repo1", Database: "/path/to/repo1.db"},
				{Owner: "owner2", Name: "repo2", Database: "/path/to/repo2.db"},
			},
		}
		adapter := &ConfigAdapter{cfg: cfg}

		dbPath := adapter.GetRepositoryDatabase("owner1/repo1")
		if dbPath != "/path/to/repo1.db" {
			t.Errorf("expected dbPath '/path/to/repo1.db', got %q", dbPath)
		}
	})

	t.Run("returns empty when database not set for repo", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{
				{Owner: "owner1", Name: "repo1"},
			},
		}
		adapter := &ConfigAdapter{cfg: cfg}

		dbPath := adapter.GetRepositoryDatabase("owner1/repo1")
		if dbPath != "" {
			t.Errorf("expected empty dbPath when not set, got %q", dbPath)
		}
	})
}

func TestRepositoryInfo_Struct(t *testing.T) {
	// Test that RepositoryInfo has the expected fields
	repo := list.RepositoryInfo{
		Owner:    "testowner",
		Name:     "testrepo",
		FullName: "testowner/testrepo",
	}

	if repo.Owner != "testowner" {
		t.Errorf("expected Owner 'testowner', got %q", repo.Owner)
	}
	if repo.Name != "testrepo" {
		t.Errorf("expected Name 'testrepo', got %q", repo.Name)
	}
	if repo.FullName != "testowner/testrepo" {
		t.Errorf("expected FullName 'testowner/testrepo', got %q", repo.FullName)
	}
}

func TestFindRepoDatabase(t *testing.T) {
	t.Run("finds database for configured repository", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{
				{Owner: "owner1", Name: "repo1", Database: "/path/to/repo1.db"},
				{Owner: "owner2", Name: "repo2", Database: "/path/to/repo2.db"},
			},
		}

		dbPath := findRepoDatabase(cfg, "owner1/repo1")
		if dbPath != "/path/to/repo1.db" {
			t.Errorf("expected '/path/to/repo1.db', got %q", dbPath)
		}
	})

	t.Run("returns empty for unknown repository", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{
				{Owner: "owner1", Name: "repo1", Database: "/path/to/repo1.db"},
			},
		}

		dbPath := findRepoDatabase(cfg, "unknown/repo")
		if dbPath != "" {
			t.Errorf("expected empty string, got %q", dbPath)
		}
	})

	t.Run("returns empty when no repositories configured", func(t *testing.T) {
		cfg := &config.Config{
			Repositories: []config.RepositoryConfig{},
		}

		dbPath := findRepoDatabase(cfg, "any/repo")
		if dbPath != "" {
			t.Errorf("expected empty string, got %q", dbPath)
		}
	})
}
