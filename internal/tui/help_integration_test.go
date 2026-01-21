package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/database"
)

func TestAppHelpIntegration(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Display: config.Display{
			Columns: []string{"number", "title", "author"},
		},
	}

	// Create test managers
	cfgMgr := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})
	dbMgr, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbMgr.Close()

	// Create app
	app := NewApp(cfg, dbMgr, cfgMgr)

	// Test that help component is initialized
	if app.help == nil {
		t.Fatal("App should have a help component")
	}

	// Test initial state - help should be hidden
	if app.help.showHelp {
		t.Error("Help should be hidden initially")
	}

	// Test ? key toggles help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updated, cmd := app.Update(msg)

	if cmd != nil {
		t.Error("? key should not return a command")
	}
	appModel, ok := updated.(*App)
	if !ok {
		t.Fatal("Update should return *App")
	}

	// Help should now be shown
	if !appModel.help.showHelp {
		t.Error("? key should show help")
	}

	// Test ? key again toggles help off
	updated, cmd = appModel.Update(msg)
	if cmd != nil {
		t.Error("? key should not return a command")
	}
	appModel, ok = updated.(*App)
	if !ok {
		t.Fatal("Update should return *App")
	}

	if appModel.help.showHelp {
		t.Error("? key should hide help when shown")
	}

	// Test Esc key hides help when shown
	appModel.help.ShowHelp()
	msg = tea.KeyMsg{Type: tea.KeyEsc}
	updated, cmd = appModel.Update(msg)
	if cmd != nil {
		t.Error("Esc key should not return a command")
	}
	appModel, ok = updated.(*App)
	if !ok {
		t.Fatal("Update should return *App")
	}

	if appModel.help.showHelp {
		t.Error("Esc key should hide help when shown")
	}

	// Test Enter key hides help when shown
	appModel.help.ShowHelp()
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd = appModel.Update(msg)
	if cmd != nil {
		t.Error("Enter key should not return a command")
	}
	appModel, ok = updated.(*App)
	if !ok {
		t.Fatal("Update should return *App")
	}

	if appModel.help.showHelp {
		t.Error("Enter key should hide help when shown")
	}

	// Test Space key hides help when shown
	appModel.help.ShowHelp()
	msg = tea.KeyMsg{Type: tea.KeySpace}
	updated, cmd = appModel.Update(msg)
	if cmd != nil {
		t.Error("Space key should not return a command")
	}
	appModel, ok = updated.(*App)
	if !ok {
		t.Fatal("Update should return *App")
	}

	if appModel.help.showHelp {
		t.Error("Space key should hide help when shown")
	}
}

func TestAppViewWithHelp(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		Display: config.Display{
			Columns: []string{"number", "title", "author"},
		},
	}

	// Create test managers
	cfgMgr := config.NewTestManager(func() (string, error) {
		return t.TempDir(), nil
	})
	dbMgr, err := database.NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer dbMgr.Close()

	// Create app
	app := NewApp(cfg, dbMgr, cfgMgr)
	app.width = 80
	app.height = 24
	app.ready = true

	// Test view without help
	view := app.View()
	if view == "" {
		t.Error("View should return non-empty string")
	}

	// Test view with help
	app.help.ShowHelp()
	app.help.width = 80
	app.help.height = 24

	viewWithHelp := app.View()
	if viewWithHelp == "" {
		t.Error("View should return non-empty string when help is shown")
	}

	// Help view should be different from regular view
	if viewWithHelp == view {
		t.Error("View with help should be different from view without help")
	}

	// Help view should contain help-related content (simple check)
	if len(viewWithHelp) < 100 {
		t.Errorf("Help view should be substantial, got length: %d", len(viewWithHelp))
	}
}

