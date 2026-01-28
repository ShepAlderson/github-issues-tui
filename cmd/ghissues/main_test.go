package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shepbook/ghissues/internal/config"
)

func TestMainModel(t *testing.T) {
	t.Run("new main model with config", func(t *testing.T) {
		cfg := &config.Config{
			Default: config.DefaultConfig{
				Repository: "owner/repo",
			},
		}
		model := NewMainModel(cfg)

		if model.config == nil {
			t.Error("Expected config to be set")
		}

		if model.config.Default.Repository != "owner/repo" {
			t.Errorf("Expected repository owner/repo, got %s", model.config.Default.Repository)
		}
	})

	t.Run("view shows repository", func(t *testing.T) {
		cfg := &config.Config{
			Default: config.DefaultConfig{
				Repository: "test/repo",
			},
		}
		model := NewMainModel(cfg)
		view := model.View()

		if !contains(view, "test/repo") {
			t.Error("Expected view to contain repository")
		}
	})

	t.Run("q key quits", func(t *testing.T) {
		model := NewMainModel(&config.Config{})

		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Error("Expected quit command on 'q'")
		}
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		model := NewMainModel(&config.Config{})

		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Error("Expected quit command on Ctrl+C")
		}
	})

	t.Run("init returns nil", func(t *testing.T) {
		model := NewMainModel(&config.Config{})

		cmd := model.Init()
		if cmd != nil {
			t.Error("Expected nil init command")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
