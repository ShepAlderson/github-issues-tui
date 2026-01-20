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
		{Number: 1, Title: "Issue 1", Author: "user1", CreatedAt: now, Comments: 0},
		{Number: 2, Title: "Issue 2", Author: "user2", CreatedAt: now, Comments: 0},
	}

	columns := []Column{
		{Name: "number", Width: 7, Title: "#"},
		{Name: "title", Width: 0, Title: "Title"},
	}

	model := NewIssueList(issues, columns)

	// Select first issue
	model.SelectCurrent()
	if model.Selected == nil {
		t.Fatal("Expected issue to be selected")
	}
	if model.Selected.Number != 1 {
		t.Errorf("Expected selected issue number 1, got %d", model.Selected.Number)
	}

	// Move to second issue and select it
	model.MoveCursor(1)
	model.SelectCurrent()
	if model.Selected.Number != 2 {
		t.Errorf("Expected selected issue number 2, got %d", model.Selected.Number)
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

	if visible[0].Number != 1 {
		t.Errorf("Expected first visible issue to be #1, got #%d", visible[0].Number)
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
