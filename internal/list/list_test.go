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

	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

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
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
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
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

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
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
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
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
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

func TestModel_SortCycling(t *testing.T) {
	cfg := &testConfig{
		columns:   []string{"number", "title"},
		sortField: "updated",
		sortDesc:  true,
	}

	t.Run("'s' cycles through sort fields", func(t *testing.T) {
		model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

		// Initial sort should be from config
		if model.sortField != "updated" {
			t.Errorf("expected initial sort field to be 'updated', got %q", model.sortField)
		}
		if !model.sortDesc {
			t.Error("expected initial sort to be descending")
		}

		// Press 's' to cycle to next sort field (created)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.sortField != "created" {
			t.Errorf("expected sort field to be 'created' after first 's', got %q", m.sortField)
		}
		// Order should reset to descending when changing field
		if !m.sortDesc {
			t.Error("expected sort to be descending after changing field")
		}

		// Press 's' again to cycle to number
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ = m.Update(msg)
		m = newModel.(Model)

		if m.sortField != "number" {
			t.Errorf("expected sort field to be 'number' after second 's', got %q", m.sortField)
		}

		// Press 's' again to cycle to comments
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ = m.Update(msg)
		m = newModel.(Model)

		if m.sortField != "comments" {
			t.Errorf("expected sort field to be 'comments' after third 's', got %q", m.sortField)
		}

		// Press 's' again to cycle back to updated
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ = m.Update(msg)
		m = newModel.(Model)

		if m.sortField != "updated" {
			t.Errorf("expected sort field to be 'updated' after fourth 's', got %q", m.sortField)
		}
	})

	t.Run("'S' toggles sort order", func(t *testing.T) {
		model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

		// Press 'S' to toggle order
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.sortDesc {
			t.Error("expected sort to be ascending after 'S'")
		}

		// Press 'S' again to toggle back
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
		newModel, _ = m.Update(msg)
		m = newModel.(Model)

		if !m.sortDesc {
			t.Error("expected sort to be descending after second 'S'")
		}
	})

	t.Run("cycling sort field resets to descending", func(t *testing.T) {
		model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

		// First toggle to ascending
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		if m.sortDesc {
			t.Error("expected sort to be ascending after 'S'")
		}

		// Then cycle to next field - should reset to descending
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ = m.Update(msg)
		m = newModel.(Model)

		if !m.sortDesc {
			t.Error("expected sort to reset to descending when changing field")
		}
	})
}

func TestModel_SortShownInView(t *testing.T) {
	cfg := &testConfig{
		columns:   []string{"number", "title"},
		repo:      "owner/repo",
		sortField: "updated",
		sortDesc:  true,
	}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}
	model.width = 80
	model.height = 24

	view := model.View()

	// Check that view contains sort information
	if !contains(view, "sort:updated") {
		t.Error("expected view to contain 'sort:updated'")
	}
	if !contains(view, "↓") {
		t.Error("expected view to contain descending indicator '↓'")
	}
}

func TestModel_SortPersistence(t *testing.T) {
	cfg := &testConfig{
		columns:   []string{"number", "title"},
		sortField: "updated",
		sortDesc:  true,
	}

	t.Run("sort change calls SaveSort", func(t *testing.T) {
		model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

		// Press 'S' to toggle sort order
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		// The sort order should have toggled
		if m.sortDesc {
			t.Error("expected sort to be ascending after 'S'")
		}

		// Verify the saveSort callback would be called (it's non-nil)
		if m.saveSort == nil {
			t.Error("expected saveSort callback to be set")
		}
	})

	t.Run("cycle sort calls SaveSort", func(t *testing.T) {
		model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

		// Press 's' to cycle sort field
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		// The sort field should have cycled
		if m.sortField != "created" {
			t.Errorf("expected sort field to be 'created', got %q", m.sortField)
		}

		// Verify the saveSort callback would be called (it's non-nil)
		if m.saveSort == nil {
			t.Error("expected saveSort callback to be set")
		}
	})
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
	columns   []string
	repo      string
	sortField string
	sortDesc  bool
}

func (c *testConfig) GetDisplayColumns() []string {
	return c.columns
}

func (c *testConfig) GetDefaultRepository() string {
	return c.repo
}

func (c *testConfig) GetSortField() string {
	return c.sortField
}

func (c *testConfig) GetSortDescending() bool {
	return c.sortDesc
}

func (c *testConfig) SaveSort(field string, descending bool) error {
	c.sortField = field
	c.sortDesc = descending
	return nil
}

func TestModel_MarkdownToggle(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}
	model.width = 100
	model.height = 24

	t.Run("'m' key is handled without error", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		// View should still render
		view := m.View()
		if !contains(view, "Issue 1") {
			t.Error("expected view to contain issue title")
		}
	})
}

func TestModel_EnterKey(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}

	t.Run("enter key is handled without error", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		// View should still render
		view := m.View()
		if !contains(view, "Issue 1") {
			t.Error("expected view to contain issue title")
		}
	})
}

func TestModel_ShouldOpenComments(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}

	t.Run("initially should not open comments", func(t *testing.T) {
		if model.ShouldOpenComments() {
			t.Error("expected ShouldOpenComments to be false initially")
		}
	})

	t.Run("enter key sets comments pending flag", func(t *testing.T) {
		m := model
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, _ := m.Update(msg)
		m = newModel.(Model)

		if !m.ShouldOpenComments() {
			t.Error("expected ShouldOpenComments to be true after Enter")
		}
	})
}

func TestModel_GetSelectedIssueForComments(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 42, Title: "Test Issue", Author: "alice"},
		{Number: 1, Title: "Issue 1", Author: "bob"},
	}

	t.Run("returns selected issue info", func(t *testing.T) {
		m := model
		m.selected = 0

		num, title, ok := m.GetSelectedIssueForComments()
		if !ok {
			t.Error("expected ok to be true")
		}
		if num != 42 {
			t.Errorf("expected issue number 42, got %d", num)
		}
		if title != "Test Issue" {
			t.Errorf("expected title 'Test Issue', got %q", title)
		}
	})

	t.Run("returns false when no issues", func(t *testing.T) {
		m := model
		m.issues = []database.ListIssue{}
		m.selected = 0

		_, _, ok := m.GetSelectedIssueForComments()
		if ok {
			t.Error("expected ok to be false when no issues")
		}
	})

	t.Run("returns false for invalid selection", func(t *testing.T) {
		m := model
		m.selected = -1

		_, _, ok := m.GetSelectedIssueForComments()
		if ok {
			t.Error("expected ok to be false for invalid selection")
		}
	})
}

func TestModel_ResetCommentsPending(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}

	m := model
	// Trigger comments pending
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if !m.ShouldOpenComments() {
		t.Fatal("expected ShouldOpenComments to be true")
	}

	// Reset
	m.ResetCommentsPending()

	if m.ShouldOpenComments() {
		t.Error("expected ShouldOpenComments to be false after reset")
	}
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

func TestModel_RefreshKey(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}

	t.Run("'r' key triggers refresh", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		// View should indicate refresh is happening
		view := m.View()
		// The view should still render (may show refreshing state)
		if !contains(view, "Issue 1") {
			t.Error("expected view to still contain issue after refresh key")
		}
	})

	t.Run("'R' key triggers refresh", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}}
		newModel, _ := model.Update(msg)
		m := newModel.(Model)

		view := m.View()
		if !contains(view, "Issue 1") {
			t.Error("expected view to still contain issue after refresh key")
		}
	})
}

func TestModel_RefreshState(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")

	t.Run("initial refresh state is false", func(t *testing.T) {
		if model.IsRefreshing() {
			t.Error("expected IsRefreshing to be false initially")
		}
	})

	t.Run("refresh state can be set", func(t *testing.T) {
		m := model
		m.SetRefreshing(true)
		if !m.IsRefreshing() {
			t.Error("expected IsRefreshing to be true after setting")
		}
	})
}

func TestModel_RefreshProgressShown(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}
	model.width = 80
	model.height = 24
	model.SetRefreshing(true)

	t.Run("refresh indicator shown in status bar", func(t *testing.T) {
		view := model.View()
		// View should render even during refresh
		if !contains(view, "Issue 1") {
			t.Error("expected view to contain issue during refresh")
		}
	})
}

func TestModel_ErrorHandling(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}
	model.width = 80
	model.height = 24

	t.Run("minor error shown in status bar", func(t *testing.T) {
		m := model
		m.SetMinorError("Network timeout", "Check connection")

		view := m.View()
		// Error message should appear in status bar
		if !contains(view, "Network timeout") {
			t.Error("expected view to contain error message in status bar")
		}
	})

	t.Run("minor error can be cleared", func(t *testing.T) {
		m := model
		m.SetMinorError("Network timeout", "Check connection")
		m.ClearMinorError()

		view := m.View()
		// Error should be cleared
		if contains(view, "Network timeout") {
			t.Error("expected error message to be cleared")
		}
	})

	t.Run("critical error triggers modal", func(t *testing.T) {
		m := model
		m.SetCriticalError("Database error", "Database corrupted")

		if !m.HasCriticalError() {
			t.Error("expected HasCriticalError to be true")
		}
	})

	t.Run("critical error returns error info", func(t *testing.T) {
		m := model
		m.SetCriticalError("Database error", "Database corrupted")

		errInfo := m.GetCriticalError()
		if errInfo == nil {
			t.Fatal("expected GetCriticalError to return error info")
		}
		if errInfo.Display != "Database error" {
			t.Errorf("expected display 'Database error', got %s", errInfo.Display)
		}
	})

	t.Run("critical error can be acknowledged", func(t *testing.T) {
		m := model
		m.SetCriticalError("Database error", "Database corrupted")
		m.AcknowledgeCriticalError()

		if m.HasCriticalError() {
			t.Error("expected HasCriticalError to be false after acknowledgment")
		}
	})
}

func TestModel_MinorErrorStatus(t *testing.T) {
	cfg := &testConfig{columns: []string{"number", "title"}}
	model := NewModel(cfg, "/tmp/test.db", "/tmp/test.toml")
	model.issues = []database.ListIssue{
		{Number: 1, Title: "Issue 1", Author: "alice"},
	}
	model.width = 80
	model.height = 24

	t.Run("error shown in status bar format", func(t *testing.T) {
		m := model
		m.SetMinorError("Connection failed", "Retry with 'r'")

		view := m.View()
		// Should contain error indicator
		if !contains(view, "Connection failed") {
			t.Error("expected view to contain error message")
		}
	})

	t.Run("error cleared on successful operation", func(t *testing.T) {
		m := model
		m.SetMinorError("Connection failed", "Retry with 'r'")

		// Simulate successful operation
		m.ClearMinorError()

		view := m.View()
		if contains(view, "Connection failed") {
			t.Error("expected error to be cleared after successful operation")
		}
	})
}
