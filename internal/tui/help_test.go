package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/git/github-issues-tui/internal/config"
)

func TestNewHelpModel(t *testing.T) {
	tests := []struct {
		name        string
		viewType    ViewType
		wantBindings int
	}{
		{
			name:        "ListView",
			viewType:    ListView,
			wantBindings: 4, // list, detail, comment, global sections
		},
		{
			name:        "DetailView",
			viewType:    DetailView,
			wantBindings: 4,
		},
		{
			name:        "CommentsView",
			viewType:    CommentsView,
			wantBindings: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpModel := NewHelpModel(tt.viewType, 80, 24, config.DefaultTheme())

			// Verify fields
			if helpModel.viewType != tt.viewType {
				t.Errorf("Expected viewType %v, got %v", tt.viewType, helpModel.viewType)
			}

			if helpModel.width != 80 {
				t.Errorf("Expected width 80, got %d", helpModel.width)
			}

			if helpModel.height != 24 {
				t.Errorf("Expected height 24, got %d", helpModel.height)
			}

			if !helpModel.active {
				t.Error("Expected help model to be active")
			}

			// Verify keybinding groups
			if len(helpModel.keybindings) != tt.wantBindings {
				t.Errorf("Expected %d keybinding groups, got %d", tt.wantBindings, len(helpModel.keybindings))
			}
		})
	}
}

func TestHelpModel_View(t *testing.T) {
	helpModel := NewHelpModel(ListView, 80, 24, config.DefaultTheme())

	view := helpModel.View()

	// Verify help content
	if !strings.Contains(view, "Keybinding Help") {
		t.Error("Expected 'Keybinding Help' title in view")
	}

	if !strings.Contains(view, "?") && !strings.Contains(view, "Esc") {
		t.Error("Expected dismissal instructions in view")
	}

	// Verify different sections exist
	sections := []string{"List View", "Issue Detail", "Comments", "Global"}
	for _, section := range sections {
		if !strings.Contains(view, section) {
			t.Errorf("Expected '%s' section in help view", section)
		}
	}

	// Verify at least some keybindings are present
	keybindings := []string{"j/k", "↑/↓", "Enter", "q", "m", "c"}
	foundBinding := false
	for _, binding := range keybindings {
		if strings.Contains(view, binding) {
			foundBinding = true
			break
		}
	}
	if !foundBinding {
		t.Error("Expected at least one keybinding to be present in view")
	}
}

func TestHelpModel_Update(t *testing.T) {
	helpModel := NewHelpModel(ListView, 80, 24, config.DefaultTheme())

	tests := []struct {
		name       string
		key        tea.KeyMsg
		expectQuit bool
	}{
		{
			name:       "Dismiss with ?",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			expectQuit: true,
		},
		{
			name:       "Dismiss with Esc",
			key:        tea.KeyMsg{Type: tea.KeyEscape},
			expectQuit: true,
		},
		{
			name:       "Dismiss with Enter",
			key:        tea.KeyMsg{Type: tea.KeyEnter},
			expectQuit: true,
		},
		{
			name:       "Dismiss with 'q'",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			expectQuit: true,
		},
		{
			name:       "Ignore other keys",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			expectQuit: false,
		},
		{
			name:       "Ignore arrow key",
			key:        tea.KeyMsg{Type: tea.KeyDown},
			expectQuit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset for each test
			helpModel.active = true

			updated, cmd := helpModel.Update(tt.key)
			helpModel = updated.(*HelpModel)

			if tt.expectQuit {
				if cmd == nil {
					t.Error("Expected quit command")
				}
				if helpModel.active {
					t.Error("Expected help model to be inactive after dismissal")
				}
			} else {
				if cmd != nil {
					t.Error("Did not expect quit command")
				}
				if !helpModel.active {
					t.Error("Expected help model to remain active")
				}
			}
		})
	}
}

func TestHelpModel_IsActive(t *testing.T) {
	helpModel := NewHelpModel(ListView, 80, 24, config.DefaultTheme())

	if !helpModel.IsActive() {
		t.Error("Expected new help model to be active")
	}

	helpModel.active = false

	if helpModel.IsActive() {
		t.Error("Expected inactive help model to return false")
	}
}

func TestGetListKeybindings(t *testing.T) {
	keybindings := getListKeybindings()

	if len(keybindings) == 0 {
		t.Fatal("Expected list keybindings to be non-empty")
	}

	// Verify keybindings have descriptions
	for _, kb := range keybindings {
		if kb.key == "" {
			t.Error("Expected keybinding to have a key")
		}
		if kb.description == "" {
			t.Error("Expected keybinding to have a description")
		}
	}

	// Verify specific keybindings exist
	expectedKeys := map[string]bool{
		"j/k":  false,
		"↑/↓": false,
		"enter": false,
		"s":    false,
		"S":    false,
		"r":    false,
	}

	for _, kb := range keybindings {
		if _, ok := expectedKeys[kb.key]; ok {
			expectedKeys[kb.key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected to find keybinding for '%s'", key)
		}
	}
}

func TestGetDetailKeybindings(t *testing.T) {
	keybindings := getDetailKeybindings()

	if len(keybindings) == 0 {
		t.Fatal("Expected detail keybindings to be non-empty")
	}

	// Verify specific keybindings exist
	expectedKeys := map[string]bool{
		"j/k":   false,
		"↑/↓":  false,
		"m":     false,
		"c":     false,
		"q/esc": false,
	}

	for _, kb := range keybindings {
		if _, ok := expectedKeys[kb.key]; ok {
			expectedKeys[kb.key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected to find keybinding for '%s'", key)
		}
	}
}

func TestGetCommentsKeybindings(t *testing.T) {
	keybindings := getCommentsKeybindings()

	if len(keybindings) == 0 {
		t.Fatal("Expected comments keybindings to be non-empty")
	}

	// Verify specific keybindings exist
	expectedKeys := map[string]bool{
		"j/k":   false,
		"↑/↓":  false,
		"m":     false,
		"q/esc": false,
	}

	for _, kb := range keybindings {
		if _, ok := expectedKeys[kb.key]; ok {
			expectedKeys[kb.key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected to find keybinding for '%s'", key)
		}
	}
}

func TestGetGlobalKeybindings(t *testing.T) {
	keybindings := getGlobalKeybindings()

	if len(keybindings) == 0 {
		t.Fatal("Expected global keybindings to be non-empty")
	}

	// Verify specific keybindings exist
	expectedKeys := map[string]bool{
		"?":       false,
		"ctrl+c": false,
	}

	for _, kb := range keybindings {
		if _, ok := expectedKeys[kb.key]; ok {
			expectedKeys[kb.key] = true
		}
	}

	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Expected to find keybinding for '%s'", key)
		}
	}
}

func TestGetContextSensitiveKeybindings(t *testing.T) {
	tests := []struct {
		viewType ViewType
		expectedSections int
	}{
		{ListView, 4},
		{DetailView, 4},
		{CommentsView, 4},
	}

	for _, tt := range tests {
		t.Run(tt.viewType.String(), func(t *testing.T) {
			keybindings := getContextSensitiveKeybindings(tt.viewType)

			if len(keybindings) != tt.expectedSections {
				t.Errorf("Expected %d keybinding sections for %s, got %d",
					tt.expectedSections, tt.viewType, len(keybindings))
			}

			// Verify all sections have keybindings
			for section, bindings := range keybindings {
				if len(bindings) == 0 {
					t.Errorf("Expected keybindings for section '%s'", section)
				}
			}
		})
	}
}

func TestGetFooter(t *testing.T) {
	tests := []struct {
		name     string
		viewType ViewType
		wantKeys []string
	}{
		{
			name:     "List view footer",
			viewType: ListView,
			wantKeys: []string{"?", "q", "enter"},
		},
		{
			name:     "Detail view footer",
			viewType: DetailView,
			wantKeys: []string{"?", "q", "m", "c"},
		},
		{
			name:     "Comments view footer",
			viewType: CommentsView,
			wantKeys: []string{"?", "q", "m"},
		},
	}


	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := getFooter(tt.viewType)

			if footer == "" {
				t.Error("Expected non-empty footer")
			}

			// Verify all wanted keys are present in footer
			for _, key := range tt.wantKeys {
				if !strings.Contains(footer, key) {
					t.Errorf("Expected footer to contain key '%s'", key)
				}
			}

			// Verify footer has proper styling elements
			if !strings.Contains(footer, "? :") {
				t.Error("Expected footer to have key descriptions")
			}
		})
	}
}
