package detail

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/database"
)

func TestNewModel(t *testing.T) {
	issue := database.IssueDetail{
		Number:    42,
		Title:     "Test Issue",
		Author:    "testuser",
		State:     "open",
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-16T14:00:00Z",
	}

	model := NewModel(issue, 60, 20, "default")

	if model.Issue.Number != 42 {
		t.Errorf("expected issue number 42, got %d", model.Issue.Number)
	}

	if model.Width != 60 {
		t.Errorf("expected width 60, got %d", model.Width)
	}

	if model.Height != 20 {
		t.Errorf("expected height 20, got %d", model.Height)
	}

	// Default should be rendered mode
	if !model.RenderedMode {
		t.Error("expected rendered mode to be true by default")
	}
}

func TestModel_ToggleRenderedMode(t *testing.T) {
	issue := database.IssueDetail{
		Number:    1,
		Title:     "Test",
		Body:      "Body content",
		Author:    "user",
		State:     "open",
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-15T10:00:00Z",
	}

	model := NewModel(issue, 60, 20, "default")

	// Initially in rendered mode
	if !model.RenderedMode {
		t.Error("expected rendered mode initially")
	}

	// Toggle to raw mode
	model.ToggleRenderedMode()
	if model.RenderedMode {
		t.Error("expected raw mode after toggle")
	}

	// Toggle back to rendered mode
	model.ToggleRenderedMode()
	if !model.RenderedMode {
		t.Error("expected rendered mode after second toggle")
	}
}

func TestRenderHeader(t *testing.T) {
	issue := database.IssueDetail{
		Number:    42,
		Title:     "Test Issue Title",
		Author:    "testuser",
		State:     "open",
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-16T14:00:00Z",
		Body:      "Issue body",
	}

	model := NewModel(issue, 60, 20, "default")
	header := model.renderHeader()

	if header == "" {
		t.Error("expected non-empty header")
	}

	// Check that header contains issue number
	if !contains(header, "#42") {
		t.Error("expected header to contain issue number '#42'")
	}

	// Check that header contains title
	if !contains(header, "Test Issue Title") {
		t.Error("expected header to contain title")
	}

	// Check that header contains author
	if !contains(header, "testuser") {
		t.Error("expected header to contain author")
	}

	// Check that header contains state
	if !contains(header, "open") {
		t.Error("expected header to contain state")
	}
}

func TestRenderHeader_WithClosedAt(t *testing.T) {
	issue := database.IssueDetail{
		Number:    42,
		Title:     "Closed Issue",
		Author:    "testuser",
		State:     "closed",
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-16T14:00:00Z",
		ClosedAt:  "2024-01-17T09:00:00Z",
		Body:      "Issue body",
	}

	model := NewModel(issue, 60, 20, "default")
	header := model.renderHeader()

	// Check that header contains closed date
	if !contains(header, "closed") {
		t.Error("expected header to contain 'closed' state")
	}
}

func TestRenderLabels(t *testing.T) {
	t.Run("renders labels when present", func(t *testing.T) {
		issue := database.IssueDetail{
			Number:    1,
			Title:     "Test",
			Author:    "user",
			State:     "open",
			CreatedAt: "2024-01-15T10:00:00Z",
			UpdatedAt: "2024-01-15T10:00:00Z",
			Labels:    []string{"bug", "enhancement"},
			Body:      "Body",
		}

		model := NewModel(issue, 60, 20, "default")
		labels := model.renderLabels()

		if labels == "" {
			t.Error("expected non-empty labels section")
		}

		if !contains(labels, "bug") {
			t.Error("expected labels to contain 'bug'")
		}

		if !contains(labels, "enhancement") {
			t.Error("expected labels to contain 'enhancement'")
		}
	})

	t.Run("returns empty string when no labels", func(t *testing.T) {
		issue := database.IssueDetail{
			Number:    1,
			Title:     "Test",
			Author:    "user",
			State:     "open",
			CreatedAt: "2024-01-15T10:00:00Z",
			UpdatedAt: "2024-01-15T10:00:00Z",
			Labels:    []string{},
			Body:      "Body",
		}

		model := NewModel(issue, 60, 20, "default")
		labels := model.renderLabels()

		if labels != "" {
			t.Errorf("expected empty string when no labels, got: %s", labels)
		}
	})
}

func TestRenderAssignees(t *testing.T) {
	t.Run("renders assignees when present", func(t *testing.T) {
		issue := database.IssueDetail{
			Number:    1,
			Title:     "Test",
			Author:    "user",
			State:     "open",
			CreatedAt: "2024-01-15T10:00:00Z",
			UpdatedAt: "2024-01-15T10:00:00Z",
			Assignees: []string{"alice", "bob"},
			Body:      "Body",
		}

		model := NewModel(issue, 60, 20, "default")
		assignees := model.renderAssignees()

		if assignees == "" {
			t.Error("expected non-empty assignees section")
		}

		if !contains(assignees, "alice") {
			t.Error("expected assignees to contain 'alice'")
		}

		if !contains(assignees, "bob") {
			t.Error("expected assignees to contain 'bob'")
		}
	})

	t.Run("returns empty string when no assignees", func(t *testing.T) {
		issue := database.IssueDetail{
			Number:    1,
			Title:     "Test",
			Author:    "user",
			State:     "open",
			CreatedAt: "2024-01-15T10:00:00Z",
			UpdatedAt: "2024-01-15T10:00:00Z",
			Assignees: []string{},
			Body:      "Body",
		}

		model := NewModel(issue, 60, 20, "default")
		assignees := model.renderAssignees()

		if assignees != "" {
			t.Errorf("expected empty string when no assignees, got: %s", assignees)
		}
	})
}

func TestModel_View(t *testing.T) {
	issue := database.IssueDetail{
		Number:       42,
		Title:        "Test Issue",
		Author:       "testuser",
		State:        "open",
		CreatedAt:    "2024-01-15T10:00:00Z",
		UpdatedAt:    "2024-01-16T14:00:00Z",
		Body:         "This is the issue body with markdown **bold** and _italic_ text.",
		CommentCount: 3,
		Labels:       []string{"bug"},
		Assignees:    []string{"alice"},
	}

	model := NewModel(issue, 60, 20, "default")
	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// Check that view contains issue number
	if !contains(view, "#42") {
		t.Error("expected view to contain issue number")
	}

	// Check that view contains title
	if !contains(view, "Test Issue") {
		t.Error("expected view to contain title")
	}

	// Check that view contains body (in raw or rendered form)
	if !contains(view, "body") && !contains(view, "bold") {
		t.Error("expected view to contain body content")
	}
}

func TestModel_View_RawMode(t *testing.T) {
	issue := database.IssueDetail{
		Number:    1,
		Title:     "Test",
		Author:    "user",
		State:     "open",
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-15T10:00:00Z",
		Body:      "Raw **markdown** body",
	}

	model := NewModel(issue, 60, 20, "default")
	model.RenderedMode = false

	view := model.View()

	// In raw mode, markdown should appear as-is
	if !contains(view, "**markdown**") {
		t.Error("expected raw mode to show unparsed markdown")
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "formats RFC3339 date",
			input:    "2024-01-15T10:30:00Z",
			expected: "2024-01-15",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSetDimensions(t *testing.T) {
	issue := database.IssueDetail{
		Number:    1,
		Title:     "Test",
		Author:    "user",
		State:     "open",
		CreatedAt: "2024-01-15T10:00:00Z",
		UpdatedAt: "2024-01-15T10:00:00Z",
		Body:      "Body",
	}

	model := NewModel(issue, 60, 20, "default")
	model.SetDimensions(80, 30)

	if model.Width != 80 {
		t.Errorf("expected width 80, got %d", model.Width)
	}

	if model.Height != 30 {
		t.Errorf("expected height 30, got %d", model.Height)
	}
}

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		maxLines int
		want     string
	}{
		{
			name:     "short body unchanged",
			body:     "Line 1\nLine 2",
			maxLines: 5,
			want:     "Line 1\nLine 2",
		},
		{
			name:     "truncates long body",
			body:     "1\n2\n3\n4\n5\n6\n7",
			maxLines: 3,
			want:     "1\n2\n3\n...",
		},
		{
			name:     "empty body",
			body:     "",
			maxLines: 5,
			want:     "",
		},
		{
			name:     "single line",
			body:     "Only line",
			maxLines: 5,
			want:     "Only line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateBody(tt.body, tt.maxLines)
			if got != tt.want {
				t.Errorf("truncateBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssueKey(t *testing.T) {
	key := IssueKey(42)
	if key != "detail_42" {
		t.Errorf("expected key 'detail_42', got %q", key)
	}
}

func contains(s, substr string) bool {
	return lipgloss.Width(s) > 0 && len(substr) > 0 && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
