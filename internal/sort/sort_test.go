package sort

import (
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/storage"
)

func TestValidateSortField(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		wantErr bool
	}{
		{
			name:    "valid field - updated",
			field:   "updated",
			wantErr: false,
		},
		{
			name:    "valid field - created",
			field:   "created",
			wantErr: false,
		},
		{
			name:    "valid field - number",
			field:   "number",
			wantErr: false,
		},
		{
			name:    "valid field - comments",
			field:   "comments",
			wantErr: false,
		},
		{
			name:    "invalid field",
			field:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty field",
			field:   "",
			wantErr: true,
		},
		{
			name:    "case sensitive - Updated",
			field:   "Updated",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSortField(tt.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSortField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDefaultSortField(t *testing.T) {
	field := GetDefaultSortField()
	if field != "updated" {
		t.Errorf("GetDefaultSortField() = %v, want %v", field, "updated")
	}
}

func TestGetDefaultSortDescending(t *testing.T) {
	descending := GetDefaultSortDescending()
	if !descending {
		t.Errorf("GetDefaultSortDescending() = %v, want %v", descending, true)
	}
}

func TestSortIssues(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{
			Number:    1,
			Title:     "First issue",
			CreatedAt: now.Add(-2 * 24 * time.Hour),
			UpdatedAt: now.Add(-1 * 24 * time.Hour),
			Comments:  5,
		},
		{
			Number:    2,
			Title:     "Second issue",
			CreatedAt: now.Add(-1 * 24 * time.Hour),
			UpdatedAt: now,
			Comments:  10,
		},
		{
			Number:    3,
			Title:     "Third issue",
			CreatedAt: now,
			UpdatedAt: now.Add(-2 * 24 * time.Hour),
			Comments:  1,
		},
	}

	tests := []struct {
		name       string
		field      string
		descending bool
		wantFirst  int // Issue number that should be first
		wantLast   int // Issue number that should be last
	}{
		{
			name:       "sort by updated descending (most recent first)",
			field:      "updated",
			descending: true,
			wantFirst:  2, // Updated now
			wantLast:   3, // Updated 2 days ago
		},
		{
			name:       "sort by updated ascending (oldest first)",
			field:      "updated",
			descending: false,
			wantFirst:  3, // Updated 2 days ago
			wantLast:   2, // Updated now
		},
		{
			name:       "sort by created descending",
			field:      "created",
			descending: true,
			wantFirst:  3, // Created now
			wantLast:   1, // Created 2 days ago
		},
		{
			name:       "sort by created ascending",
			field:      "created",
			descending: false,
			wantFirst:  1, // Created 2 days ago
			wantLast:   3, // Created now
		},
		{
			name:       "sort by number descending",
			field:      "number",
			descending: true,
			wantFirst:  3,
			wantLast:   1,
		},
		{
			name:       "sort by number ascending",
			field:      "number",
			descending: false,
			wantFirst:  1,
			wantLast:   3,
		},
		{
			name:       "sort by comments descending",
			field:      "comments",
			descending: true,
			wantFirst:  2, // 10 comments
			wantLast:   3, // 1 comment
		},
		{
			name:       "sort by comments ascending",
			field:      "comments",
			descending: false,
			wantFirst:  3, // 1 comment
			wantLast:   2, // 10 comments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sorted := SortIssues(issues, tt.field, tt.descending)

			if len(sorted) != len(issues) {
				t.Fatalf("SortIssues() returned %d issues, want %d", len(sorted), len(issues))
			}

			if sorted[0].Number != tt.wantFirst {
				t.Errorf("SortIssues() first issue = %d, want %d", sorted[0].Number, tt.wantFirst)
			}

			if sorted[len(sorted)-1].Number != tt.wantLast {
				t.Errorf("SortIssues() last issue = %d, want %d", sorted[len(sorted)-1].Number, tt.wantLast)
			}
		})
	}
}

func TestSortIssuesEmpty(t *testing.T) {
	issues := []storage.Issue{}
	sorted := SortIssues(issues, "updated", true)

	if len(sorted) != 0 {
		t.Errorf("SortIssues() on empty slice returned %d issues, want 0", len(sorted))
	}
}

func TestSortIssuesSingle(t *testing.T) {
	issues := []storage.Issue{
		{
			Number:    1,
			Title:     "Only issue",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	sorted := SortIssues(issues, "updated", true)

	if len(sorted) != 1 {
		t.Errorf("SortIssues() on single issue returned %d issues, want 1", len(sorted))
	}

	if sorted[0].Number != 1 {
		t.Errorf("SortIssues() single issue = %d, want 1", sorted[0].Number)
	}
}

func TestCycleSortField(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		expected string
	}{
		{
			name:     "cycle from updated to created",
			current:  "updated",
			expected: "created",
		},
		{
			name:     "cycle from created to number",
			current:  "created",
			expected: "number",
		},
		{
			name:     "cycle from number to comments",
			current:  "number",
			expected: "comments",
		},
		{
			name:     "cycle from comments to updated",
			current:  "comments",
			expected: "updated",
		},
		{
			name:     "empty field defaults to updated",
			current:  "",
			expected: "updated",
		},
		{
			name:     "invalid field defaults to updated",
			current:  "invalid",
			expected: "updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CycleSortField(tt.current)
			if result != tt.expected {
				t.Errorf("CycleSortField(%q) = %q, want %q", tt.current, result, tt.expected)
			}
		})
	}
}

func TestGetSortFieldLabel(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "label for updated",
			field:    "updated",
			expected: "Updated",
		},
		{
			name:     "label for created",
			field:    "created",
			expected: "Created",
		},
		{
			name:     "label for number",
			field:    "number",
			expected: "Number",
		},
		{
			name:     "label for comments",
			field:    "comments",
			expected: "Comments",
		},
		{
			name:     "label for empty",
			field:    "",
			expected: "Unknown",
		},
		{
			name:     "label for invalid",
			field:    "invalid",
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSortFieldLabel(tt.field)
			if result != tt.expected {
				t.Errorf("GetSortFieldLabel(%q) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}
