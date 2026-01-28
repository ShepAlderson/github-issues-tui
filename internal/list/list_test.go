package list

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/ghissues/internal/database"
)

func TestNewModel(t *testing.T) {
	cfg := &testConfig{
		columns: []string{"number", "title", "author"},
		repo:    "owner/repo",
	}

	model := NewModel(cfg, "/tmp/test.db")

	if model.dbPath != "/tmp/test.db" {
		t.Errorf("expected dbPath to be '/tmp/test.db', got %q", model.dbPath)
	}

	if model.selected != 0 {
		t.Errorf("expected selected to be 0, got %d", model.selected)
	}

	if model.repo != "owner/repo" {
		t.Errorf("expected repo to be 'owner/repo', got %q", model.repo)
	}
}

func TestModel_Navigation(t *testing.T) {
	// Create model with test issues
	cfg := &testConfig{columns: []string{"number", "title", "author"}}
	model := NewModel(cfg, "/tmp/test.db")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
		{Number: 2, Title: "Issue 2", Author: "bob"},
		{Number: 3, Title: "Issue 3", Author: "charlie"},
	}

	t.Run("j moves down", func(t *testing.T) {
		m := model
		m.selected = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.selected != 1 {
			t.Errorf("expected selected to be 1 after 'j', got %d", m.selected)
		}
	})

	t.Run("k moves up", func(t *testing.T) {
		m := model
		m.selected = 1

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.selected != 0 {
			t.Errorf("expected selected to be 0 after 'k', got %d", m.selected)
		}
	})

	t.Run("down arrow moves down", func(t *testing.T) {
		m := model
		m.selected = 0

		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.selected != 1 {
			t.Errorf("expected selected to be 1 after down arrow, got %d", m.selected)
		}
	})

	t.Run("up arrow moves up", func(t *testing.T) {
		m := model
		m.selected = 1

		msg := tea.KeyMsg{Type: tea.KeyUp}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.selected != 0 {
			t.Errorf("expected selected to be 0 after up arrow, got %d", m.selected)
		}
	})

	t.Run("j at bottom does not overflow", func(t *testing.T) {
		m := model
		m.selected = 2

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.selected != 2 {
			t.Errorf("expected selected to stay at 2 at bottom, got %d", m.selected)
		}
	})

	t.Run("k at top does not underflow", func(t *testing.T) {
		m := model
		m.selected = 0

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if m.selected != 0 {
			t.Errorf("expected selected to stay at 0 at top, got %d", m.selected)
		}
	})
}

func TestModel_CtrlCQuits(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db")

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := model.Update(msg)

	// Check that the command is tea.Quit
	if cmd == nil {
		t.Error("expected tea.Quit command for Ctrl+C, got nil")
	}
}

func TestRenderColumns(t *testing.T) {
	issue := database.ListIssue{
		Number:       42,
		Title:        "Test Issue",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-16T14:20:00Z",
		CommentCount: 5,
	}

	tests := []struct {
		name    string
		columns []string
		want    []string
	}{
		{
			name:    "number column",
			columns: []string{"number"},
			want:    []string{"#42"},
		},
		{
			name:    "title column",
			columns: []string{"title"},
			want:    []string{"Test Issue"},
		},
		{
			name:    "author column",
			columns: []string{"author"},
			want:    []string{"testuser"},
		},
		{
			name:    "date columns",
			columns: []string{"created", "updated"},
			want:    []string{"2024-01-15", "2024-01-16"},
		},
		{
			name:    "comments column",
			columns: []string{"comments"},
			want:    []string{"💬 5"},
		},
		{
			name:    "multiple columns",
			columns: []string{"number", "title", "author"},
			want:    []string{"#42", "Test Issue", "testuser"},
		},
		{
			name:    "unknown column is skipped",
			columns: []string{"number", "unknown", "title"},
			want:    []string{"#42", "Test Issue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderColumns(issue, tt.columns)
			for i, want := range tt.want {
				if i >= len(result) {
					t.Errorf("missing column %d, want %q", i, want)
					continue
				}
				if result[i] != want {
					t.Errorf("column %d = %q, want %q", i, result[i], want)
				}
			}
		})
	}
}

func TestModel_ViewContainsIssueCount(t *testing.T) {
	cfg := &testConfig{
		columns: []string{"number", "title", "author"},
		repo:    "owner/repo",
	}
	model := NewModel(cfg, "/tmp/test.db")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
		{Number: 2, Title: "Issue 2", Author: "bob"},
	}
	model.width = 80
	model.height = 24

	view := model.View()

	// Check that view contains issue count in status area
	if !contains(view, "2 issues") {
		t.Error("expected view to contain '2 issues'")
	}
}

func TestModel_ViewShowsSelected(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
		{Number: 2, Title: "Issue 2", Author: "bob"},
	}
	model.selected = 1
	model.width = 80
	model.height = 24

	view := model.View()

	// The selected issue should be highlighted (contains special styling)
	// We verify the view contains both issues
	if !contains(view, "Issue 1") {
		t.Error("expected view to contain 'Issue 1'")
	}
	if !contains(view, "Issue 2") {
		t.Error("expected view to contain 'Issue 2'")
	}
}

func TestValidateColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		valid   []string
	}{
		{
			name:    "valid columns",
			columns: []string{"number", "title", "author", "created", "updated", "comments"},
			valid:   []string{"number", "title", "author", "created", "updated", "comments"},
		},
		{
			name:    "filters invalid columns",
			columns: []string{"number", "invalid", "title", "unknown"},
			valid:   []string{"number", "title"},
		},
		{
			name:    "empty returns empty",
			columns: []string{},
			valid:   []string{},
		},
		{
			name:    "all invalid returns empty",
			columns: []string{"invalid", "unknown"},
			valid:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateColumns(tt.columns)
			if len(result) != len(tt.valid) {
				t.Errorf("validateColumns() = %v, want %v", result, tt.valid)
				return
			}
			for i, v := range tt.valid {
				if result[i] != v {
					t.Errorf("validateColumns()[%d] = %v, want %v", i, result[i], v)
				}
			}
		})
	}
}

// testConfig implements a minimal Config interface for testing
type testConfig struct {
	columns []string
	repo    string
}

func (c *testConfig) GetDisplayColumns() []string {
	return c.columns
}

func (c *testConfig) GetDefaultRepository() string {
	return c.repo
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
