package tui

import (
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/storage"
)

func TestNewIssueList(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "First issue", Author: "user1", CreatedAt: now, Comments: 2},
		{Number: 2, Title: "Second issue", Author: "user2", CreatedAt: now, Comments: 5},
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)

	if model == nil {
		t.Fatal("NewIssueList() returned nil")
	}

	if len(model.Issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(model.Issues))
	}

	if model.Cursor != 0 {
		t.Errorf("Expected cursor at position 0, got %d", model.Cursor)
	}

	if model.Selected != nil {
		t.Error("Expected no selected issue initially")
	}
}

func TestIssueList_MoveCursor(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "Issue 1", Author: "user1", CreatedAt: now, Comments: 0},
		{Number: 2, Title: "Issue 2", Author: "user2", CreatedAt: now, Comments: 0},
		{Number: 3, Title: "Issue 3", Author: "user3", CreatedAt: now, Comments: 0},
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)

	// Test moving down
	model.MoveCursor(1)
	if model.Cursor != 1 {
		t.Errorf("Expected cursor at position 1 after moving down, got %d", model.Cursor)
	}

	// Test moving up
	model.MoveCursor(-1)
	if model.Cursor != 0 {
		t.Errorf("Expected cursor at position 0 after moving up, got %d", model.Cursor)
	}

	// Test boundary - can't go below 0
	model.MoveCursor(-1)
	if model.Cursor != 0 {
		t.Errorf("Expected cursor at position 0 when trying to move below 0, got %d", model.Cursor)
	}

	// Test boundary - can't go past last item
	model.Cursor = 2
	model.MoveCursor(1)
	if model.Cursor != 2 {
		t.Errorf("Expected cursor at position 2 when trying to move past end, got %d", model.Cursor)
	}
}

func TestIssueList_SelectCurrent(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "Issue 1", Author: "user1", CreatedAt: now, UpdatedAt: now, Comments: 0},
		{Number: 2, Title: "Issue 2", Author: "user2", CreatedAt: now, UpdatedAt: now, Comments: 0},
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)

	// Select first issue (may be #1 or #2 depending on default sort)
	firstIssueNumber := model.Issues[0].Number
	model.SelectCurrent()
	if model.Selected == nil {
		t.Fatal("Expected issue to be selected")
	}
	if model.Selected.Number != firstIssueNumber {
		t.Errorf("Expected selected issue number %d, got %d", firstIssueNumber, model.Selected.Number)
	}

	// Move to second issue and select it
	model.MoveCursor(1)
	secondIssueNumber := model.Issues[1].Number
	model.SelectCurrent()
	if model.Selected.Number != secondIssueNumber {
		t.Errorf("Expected selected issue number %d, got %d", secondIssueNumber, model.Selected.Number)
	}
}

func TestIssueList_SetViewport(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "Issue 1", Author: "user1", CreatedAt: now, Comments: 0},
		{Number: 2, Title: "Issue 2", Author: "user2", CreatedAt: now, Comments: 0},
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)
	model.SetViewport(10)

	if model.ViewportHeight != 10 {
		t.Errorf("Expected viewport height 10, got %d", model.ViewportHeight)
	}
}

func TestIssueList_VisibleRange(t *testing.T) {
	now := time.Now()
	issues := make([]storage.Issue, 20)
	for i := 0; i < 20; i++ {
		issues[i] = storage.Issue{
			Number:    i + 1,
			Title:     "Issue Title",
			Author:    "user",
			CreatedAt: now,
			Comments:  0,
		}
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)
	model.SetViewport(10)

	// Test initial visible range
	start, end := model.VisibleRange()
	if start != 0 || end != 10 {
		t.Errorf("Expected visible range [0, 10), got [%d, %d)", start, end)
	}

	// Move cursor to trigger scrolling
	model.Cursor = 9
	model.MoveCursor(1) // Move to item 10

	start, end = model.VisibleRange()
	if start != 1 || end != 11 {
		t.Errorf("Expected visible range [1, 11) after scrolling, got [%d, %d)", start, end)
	}
}

func TestIssueList_GetVisibleIssues(t *testing.T) {
	now := time.Now()
	issues := make([]storage.Issue, 20)
	for i := 0; i < 20; i++ {
		issues[i] = storage.Issue{
			Number:    i + 1,
			Title:     "Issue Title",
			Author:    "user",
			CreatedAt: now,
			UpdatedAt: now,
			Comments:  0,
		}
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)
	model.SetViewport(10)

	visible := model.GetVisibleIssues()
	if len(visible) != 10 {
		t.Errorf("Expected 10 visible issues, got %d", len(visible))
	}

	// Just verify we got issues, don't check specific numbers since they're sorted
	if len(visible) == 0 {
		t.Error("Expected to get visible issues, got none")
	}
}

func TestIssueList_EmptyList(t *testing.T) {
	issues := []storage.Issue{}
	columns := []Column{{Name: "number", Width: 7, Title: "#"}}

	model := NewIssueList(issues, columns)

	if len(model.Issues) != 0 {
		t.Errorf("Expected empty issue list, got %d issues", len(model.Issues))
	}

	// Test that operations on empty list don't panic
	model.MoveCursor(1)
	model.MoveCursor(-1)
	model.SelectCurrent()

	if model.Cursor != 0 {
		t.Errorf("Expected cursor to remain at 0 for empty list, got %d", model.Cursor)
	}
}

func TestIssueList_DefaultSort(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "Old issue", Author: "user1", CreatedAt: now.Add(-2 * 24 * time.Hour), UpdatedAt: now.Add(-2 * 24 * time.Hour), Comments: 1},
		{Number: 2, Title: "New issue", Author: "user2", CreatedAt: now, UpdatedAt: now, Comments: 5},
		{Number: 3, Title: "Medium issue", Author: "user3", CreatedAt: now.Add(-1 * 24 * time.Hour), UpdatedAt: now.Add(-1 * 24 * time.Hour), Comments: 3},
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)

	// Default should be updated descending (most recent first)
	if model.SortField != "updated" {
		t.Errorf("Expected default sort field 'updated', got '%s'", model.SortField)
	}

	if !model.SortDescending {
		t.Error("Expected default sort order to be descending")
	}

	// First issue should be #2 (most recently updated)
	if model.Issues[0].Number != 2 {
		t.Errorf("Expected first issue to be #2 (most recent), got #%d", model.Issues[0].Number)
	}
}

func TestIssueList_SetSort(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 3, Title: "Third", Author: "user1", CreatedAt: now, UpdatedAt: now, Comments: 10},
		{Number: 1, Title: "First", Author: "user2", CreatedAt: now, UpdatedAt: now, Comments: 5},
		{Number: 2, Title: "Second", Author: "user3", CreatedAt: now, UpdatedAt: now, Comments: 1},
	}

	columns := []Column{{Name: "number", Width: 7, Title: "#"}}
	model := NewIssueList(issues, columns)

	// Sort by number ascending
	model.SetSort("number", false)

	if model.SortField != "number" {
		t.Errorf("Expected sort field 'number', got '%s'", model.SortField)
	}

	if model.SortDescending {
		t.Error("Expected ascending sort order")
	}

	if model.Issues[0].Number != 1 {
		t.Errorf("Expected first issue to be #1, got #%d", model.Issues[0].Number)
	}

	if model.Issues[2].Number != 3 {
		t.Errorf("Expected last issue to be #3, got #%d", model.Issues[2].Number)
	}
}

func TestIssueList_CycleSortField(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "Issue", Author: "user", CreatedAt: now, UpdatedAt: now, Comments: 1},
	}

	columns := []Column{{Name: "number", Width: 7, Title: "#"}}
	model := NewIssueList(issues, columns)

	// Default is updated
	if model.SortField != "updated" {
		t.Errorf("Expected initial sort field 'updated', got '%s'", model.SortField)
	}

	// Cycle: updated -> created
	model.CycleSortField()
	if model.SortField != "created" {
		t.Errorf("Expected sort field 'created' after first cycle, got '%s'", model.SortField)
	}

	// Cycle: created -> number
	model.CycleSortField()
	if model.SortField != "number" {
		t.Errorf("Expected sort field 'number' after second cycle, got '%s'", model.SortField)
	}

	// Cycle: number -> comments
	model.CycleSortField()
	if model.SortField != "comments" {
		t.Errorf("Expected sort field 'comments' after third cycle, got '%s'", model.SortField)
	}

	// Cycle: comments -> updated (back to start)
	model.CycleSortField()
	if model.SortField != "updated" {
		t.Errorf("Expected sort field 'updated' after fourth cycle, got '%s'", model.SortField)
	}
}

func TestIssueList_ToggleSortOrder(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "Issue", Author: "user", CreatedAt: now, UpdatedAt: now, Comments: 1},
	}

	columns := []Column{{Name: "number", Width: 7, Title: "#"}}
	model := NewIssueList(issues, columns)

	// Default is descending
	if !model.SortDescending {
		t.Error("Expected initial sort order to be descending")
	}

	// Toggle to ascending
	model.ToggleSortOrder()
	if model.SortDescending {
		t.Error("Expected sort order to be ascending after toggle")
	}

	// Toggle back to descending
	model.ToggleSortOrder()
	if !model.SortDescending {
		t.Error("Expected sort order to be descending after second toggle")
	}
}

func TestIssueList_SortPreservesUnsortedIssues(t *testing.T) {
	now := time.Now()
	issues := []storage.Issue{
		{Number: 1, Title: "First", Author: "user1", CreatedAt: now, UpdatedAt: now.Add(-2 * 24 * time.Hour), Comments: 1},
		{Number: 2, Title: "Second", Author: "user2", CreatedAt: now, UpdatedAt: now, Comments: 5},
	}

	columns := []Column{{Name: "number", Width: 7, Title: "#"}}
	model := NewIssueList(issues, columns)

	// UnsortedIssues should preserve original order
	if len(model.UnsortedIssues) != 2 {
		t.Fatalf("Expected 2 unsorted issues, got %d", len(model.UnsortedIssues))
	}

	if model.UnsortedIssues[0].Number != 1 {
		t.Errorf("Expected first unsorted issue to be #1, got #%d", model.UnsortedIssues[0].Number)
	}

	// Sorted issues should be in updated descending order
	if model.Issues[0].Number != 2 {
		t.Errorf("Expected first sorted issue to be #2 (most recent), got #%d", model.Issues[0].Number)
	}
}

