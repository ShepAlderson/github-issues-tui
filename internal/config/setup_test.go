package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSetupModel(t *testing.T) {
	model := NewSetupModel()

	if model.step != StepRepository {
		t.Errorf("Expected initial step to be StepRepository, got %d", model.step)
	}

	if model.config == nil {
		t.Error("Expected config to be initialized")
	}

	if model.config.Display.Theme != "default" {
		t.Errorf("Expected default theme, got %s", model.config.Display.Theme)
	}
}

func TestSetupModelUpdate(t *testing.T) {
	t.Run("enter valid repository advances to auth step", func(t *testing.T) {
		model := NewSetupModel()
		model.repoInput.SetValue("charmbracelet/bubbletea")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepAuthMethod {
			t.Errorf("Expected step to advance to StepAuthMethod, got %d", m.step)
		}

		if m.config.Default.Repository != "charmbracelet/bubbletea" {
			t.Errorf("Expected repository to be set, got %s", m.config.Default.Repository)
		}
	})

	t.Run("enter invalid repository shows error", func(t *testing.T) {
		model := NewSetupModel()
		model.repoInput.SetValue("invalid-repo")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepRepository {
			t.Errorf("Expected step to stay at StepRepository, got %d", m.step)
		}

		if m.repoError == "" {
			t.Error("Expected repository error to be set")
		}
	})

	t.Run("empty repository shows error", func(t *testing.T) {
		model := NewSetupModel()
		model.repoInput.SetValue("")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepRepository {
			t.Errorf("Expected step to stay at StepRepository, got %d", m.step)
		}

		if m.repoError == "" {
			t.Error("Expected repository error to be set")
		}
	})

	t.Run("select env var auth advances to confirm", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepAuthMethod
		model.authSelection = 0 // env var

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepConfirm {
			t.Errorf("Expected step to advance to StepConfirm, got %d", m.step)
		}
	})

	t.Run("select config file auth advances to token step", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepAuthMethod
		model.authSelection = 1 // config file

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepToken {
			t.Errorf("Expected step to advance to StepToken, got %d", m.step)
		}
	})

	t.Run("enter valid token advances to confirm", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepToken
		model.tokenInput.SetValue("ghp_test123token")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepConfirm {
			t.Errorf("Expected step to advance to StepConfirm, got %d", m.step)
		}

		if m.config.Auth.Token != "ghp_test123token" {
			t.Errorf("Expected token to be set, got %s", m.config.Auth.Token)
		}
	})

	t.Run("enter invalid token shows error", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepToken
		model.tokenInput.SetValue("invalid_token")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepToken {
			t.Errorf("Expected step to stay at StepToken, got %d", m.step)
		}

		if m.tokenError == "" {
			t.Error("Expected token error to be set")
		}
	})

	t.Run("empty token shows error", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepToken
		model.tokenInput.SetValue("")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepToken {
			t.Errorf("Expected step to stay at StepToken, got %d", m.step)
		}

		if m.tokenError == "" {
			t.Error("Expected token error to be set")
		}
	})

	t.Run("config file token with github_pat prefix works", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepToken
		model.tokenInput.SetValue("github_pat_123abc")

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(SetupModel)

		if m.step != StepConfirm {
			t.Errorf("Expected step to advance to StepConfirm, got %d", m.step)
		}
	})

	t.Run("esc from auth returns to repository", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepAuthMethod

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m := newModel.(SetupModel)

		if m.step != StepRepository {
			t.Errorf("Expected step to return to StepRepository, got %d", m.step)
		}
	})

	t.Run("esc from token returns to auth", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepToken

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m := newModel.(SetupModel)

		if m.step != StepAuthMethod {
			t.Errorf("Expected step to return to StepAuthMethod, got %d", m.step)
		}
	})

	t.Run("esc from confirm returns to auth", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepConfirm

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m := newModel.(SetupModel)

		if m.step != StepAuthMethod {
			t.Errorf("Expected step to return to StepAuthMethod, got %d", m.step)
		}
	})

	t.Run("down key cycles auth selection", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepAuthMethod
		model.authSelection = 0

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		m := newModel.(SetupModel)

		if m.authSelection != 1 {
			t.Errorf("Expected authSelection to be 1, got %d", m.authSelection)
		}

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = newModel.(SetupModel)

		// Should wrap back to 0
		if m.authSelection != 0 {
			t.Errorf("Expected authSelection to wrap to 0, got %d", m.authSelection)
		}
	})

	t.Run("up key cycles auth selection", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepAuthMethod
		model.authSelection = 0

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
		m := newModel.(SetupModel)

		// Should wrap to 1
		if m.authSelection != 1 {
			t.Errorf("Expected authSelection to wrap to 1, got %d", m.authSelection)
		}
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		model := NewSetupModel()

		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

		if cmd == nil {
			t.Error("Expected tea.Quit command")
		}
	})
}

func TestSetupModelView(t *testing.T) {
	t.Run("repository step shows input", func(t *testing.T) {
		model := NewSetupModel()
		view := model.View()

		if !strings.Contains(view, "Step 1/3") {
			t.Error("Expected view to contain 'Step 1/3'")
		}

		if !strings.Contains(view, "owner/repo") {
			t.Error("Expected view to contain repository format hint")
		}
	})

	t.Run("auth step shows options", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepAuthMethod
		view := model.View()

		if !strings.Contains(view, "Step 2/3") {
			t.Error("Expected view to contain 'Step 2/3'")
		}

		if !strings.Contains(view, "GITHUB_TOKEN") {
			t.Error("Expected view to mention GITHUB_TOKEN")
		}
	})

	t.Run("token step shows input", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepToken
		view := model.View()

		if !strings.Contains(view, "GitHub Personal Access Token") {
			t.Error("Expected view to mention token")
		}
	})

	t.Run("confirm step shows summary", func(t *testing.T) {
		model := NewSetupModel()
		model.step = StepConfirm
		model.config.Default.Repository = "owner/repo"
		model.authSelection = 0
		view := model.View()

		if !strings.Contains(view, "Step 3/3") {
			t.Error("Expected view to contain 'Step 3/3'")
		}

		if !strings.Contains(view, "owner/repo") {
			t.Error("Expected view to show repository")
		}
	})

	t.Run("error appears in view", func(t *testing.T) {
		model := NewSetupModel()
		model.repoError = "Invalid format"
		view := model.View()

		if !strings.Contains(view, "Invalid format") {
			t.Error("Expected error message to appear in view")
		}
	})
}

func TestRunSetup(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)
	os.Setenv("HOME", tempDir)

	t.Run("setup completed saves config", func(t *testing.T) {
		// Reset config
		configDir := filepath.Join(tempDir, ".config", "ghissues")
		os.RemoveAll(configDir)

		// Create and complete setup model manually since we can't easily
		// simulate the full TUI flow in a unit test
		model := NewSetupModel()
		model.repoInput.SetValue("testowner/testrepo")
		model.config.Default.Repository = "testowner/testrepo"
		model.config.Repositories = []RepositoryConfig{
			{Owner: "testowner", Name: "testrepo", Database: ".ghissues.db"},
		}

		// Simulate completing setup
		err := model.config.Save()
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify file exists
		configPath := filepath.Join(configDir, "config.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Expected config file to be created")
		}

		// Load and verify content
		loaded, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loaded.Default.Repository != "testowner/testrepo" {
			t.Errorf("Expected repository testowner/testrepo, got %s", loaded.Default.Repository)
		}
	})
}
