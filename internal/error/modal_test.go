package error

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModalModel(t *testing.T) {
	appErr := AppError{
		Original:   errors.New("database error"),
		Severity:   SeverityCritical,
		Display:    "Database error",
		Guidance:   "Try removing the database file",
		Actionable: true,
	}

	model := NewModalModel(appErr)

	if model.Error.Severity != SeverityCritical {
		t.Errorf("Expected critical severity, got %v", model.Error.Severity)
	}

	if model.Error.Display != "Database error" {
		t.Errorf("Expected display 'Database error', got %s", model.Error.Display)
	}

	if model.Error.Guidance != "Try removing the database file" {
		t.Errorf("Expected guidance 'Try removing the database file', got %s", model.Error.Guidance)
	}
}

func TestModalModel_Init(t *testing.T) {
	appErr := AppError{
		Original: errors.New("test error"),
		Severity: SeverityCritical,
		Display:  "Test error",
	}

	model := NewModalModel(appErr)
	cmd := model.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModalModel_Update(t *testing.T) {
	appErr := AppError{
		Original: errors.New("test error"),
		Severity: SeverityCritical,
		Display:  "Test error",
	}

	tests := []struct {
		name        string
		keyType     tea.KeyType
		keyRunes    string
		wantQuit    bool
		wantHandled bool
	}{
		{
			name:        "Enter acknowledges modal",
			keyType:     tea.KeyEnter,
			wantQuit:    true,
			wantHandled: true,
		},
		{
			name:        "Space acknowledges modal",
			keyRunes:    " ",
			wantQuit:    true,
			wantHandled: true,
		},
		{
			name:        "Any other key is ignored",
			keyRunes:    "a",
			wantQuit:    false,
			wantHandled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModalModel(appErr)

			var msg tea.Msg
			if tt.keyType != 0 {
				msg = tea.KeyMsg{Type: tt.keyType}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.keyRunes)}
			}

			newModel, cmd := model.Update(msg)

			if tt.wantQuit {
				// Check that a quit message was returned
				if cmd == nil {
					t.Error("Expected quit command, got nil")
				}
			}

			// Verify the model type
			_, ok := newModel.(ModalModel)
			if !ok {
				t.Error("Update should return ModalModel")
			}
		})
	}
}

func TestModalModel_Update_WindowSize(t *testing.T) {
	appErr := AppError{
		Original: errors.New("test error"),
		Severity: SeverityCritical,
		Display:  "Test error",
	}

	model := NewModalModel(appErr)

	// Test window size message
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(msg)

	m, ok := newModel.(ModalModel)
	if !ok {
		t.Fatal("Update should return ModalModel")
	}

	if m.Width != 80 {
		t.Errorf("Expected width 80, got %d", m.Width)
	}

	if m.Height != 24 {
		t.Errorf("Expected height 24, got %d", m.Height)
	}
}

func TestModalModel_View(t *testing.T) {
	tests := []struct {
		name           string
		err            AppError
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "Critical error with guidance",
			err: AppError{
				Original:   errors.New("database error"),
				Severity:   SeverityCritical,
				Display:    "Database error",
				Guidance:   "Try removing the database file",
				Actionable: true,
			},
			wantContains: []string{
				"Database error",
				"Try removing the database file",
				"✗",  // Error indicator
				"Press Enter",
			},
		},
		{
			name: "Error without guidance",
			err: AppError{
				Original: errors.New("unknown"),
				Severity: SeverityMinor,
				Display:  "An error occurred",
			},
			wantContains: []string{
				"An error occurred",
				"Press Enter",
			},
			wantNotContain: []string{
				"Try removing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModalModel(tt.err)
			view := model.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf("View should contain %q, got:\n%s", want, view)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(view, notWant) {
					t.Errorf("View should NOT contain %q", notWant)
				}
			}
		})
	}
}

func TestModalModel_View_ContainsAcknowledgment(t *testing.T) {
	appErr := AppError{
		Original:   errors.New("test"),
		Severity:   SeverityCritical,
		Display:    "Test error",
		Guidance:   "Some guidance",
		Actionable: true,
	}

	model := NewModalModel(appErr)
	view := model.View()

	// Should contain acknowledgment instruction
	if !strings.Contains(view, "Press Enter or Space to continue") && !strings.Contains(view, "Press Enter") {
		t.Errorf("View should contain acknowledgment instruction, got:\n%s", view)
	}
}

func TestModalModel_SetDimensions(t *testing.T) {
	appErr := AppError{
		Original: errors.New("test"),
		Severity: SeverityCritical,
		Display:  "Test error",
	}

	model := NewModalModel(appErr)
	model.SetDimensions(120, 40)

	if model.Width != 120 {
		t.Errorf("Expected width 120, got %d", model.Width)
	}

	if model.Height != 40 {
		t.Errorf("Expected height 40, got %d", model.Height)
	}
}

func TestModalModel_WasAcknowledged(t *testing.T) {
	appErr := AppError{
		Original: errors.New("test"),
		Severity: SeverityCritical,
		Display:  "Test error",
	}

	model := NewModalModel(appErr)

	// Initially not acknowledged
	if model.WasAcknowledged() {
		t.Error("Should not be acknowledged initially")
	}

	// Simulate acknowledgment
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)

	m := newModel.(ModalModel)
	if !m.WasAcknowledged() {
		t.Error("Should be acknowledged after Enter key")
	}
}

func TestRunModal(t *testing.T) {
	appErr := AppError{
		Original: errors.New("test"),
		Severity: SeverityCritical,
		Display:  "Test error",
		Guidance: "Test guidance",
	}

	// This just tests that RunModal creates a model correctly
	// We can't actually run the tea.Program in unit tests
	model := RunModal(appErr)

	if model.Error.Display != "Test error" {
		t.Errorf("Expected display 'Test error', got %s", model.Error.Display)
	}
}
