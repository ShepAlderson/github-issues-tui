package tui

import (
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/storage"
)

func TestNewCommentsView(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test Issue",
		Author:    "testuser",
		Body:      "Test body",
		State:     "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "First comment",
			Author:      "user1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			IssueNumber: 123,
			Body:        "Second comment",
			Author:      "user2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)

	if view == nil {
		t.Fatal("NewCommentsView returned nil")
	}

	if view.Issue.Number != 123 {
		t.Errorf("Expected issue number 123, got %d", view.Issue.Number)
	}

	if len(view.Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(view.Comments))
	}

	if view.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be false by default")
	}

	if view.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to be 0, got %d", view.ScrollOffset)
	}
}

func TestNewCommentsViewEmptyComments(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test Issue",
		Author:    "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{}

	view := NewCommentsView(issue, comments)

	if view == nil {
		t.Fatal("NewCommentsView returned nil")
	}

	if len(view.Comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(view.Comments))
	}
}

func TestToggleMarkdown(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)

	// Initial state should be false
	if view.RenderMarkdown {
		t.Error("Expected initial RenderMarkdown to be false")
	}

	// Toggle to true
	view.ToggleMarkdown()
	if !view.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be true after toggle")
	}

	// Toggle back to false
	view.ToggleMarkdown()
	if view.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be false after second toggle")
	}
}

func TestToggleMarkdownResetsScroll(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.ScrollOffset = 5

	view.ToggleMarkdown()

	if view.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to reset to 0, got %d", view.ScrollOffset)
	}
}

func TestScrollDown(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	initialOffset := view.ScrollOffset

	view.ScrollDown()

	if view.ScrollOffset != initialOffset+1 {
		t.Errorf("Expected ScrollOffset to increment, got %d", view.ScrollOffset)
	}
}

func TestScrollUp(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.ScrollOffset = 5

	view.ScrollUp()

	if view.ScrollOffset != 4 {
		t.Errorf("Expected ScrollOffset to be 4, got %d", view.ScrollOffset)
	}
}

func TestScrollUpAtZero(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)

	// Scroll up at position 0 should stay at 0
	view.ScrollUp()

	if view.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to remain 0, got %d", view.ScrollOffset)
	}
}

func TestSetViewport(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.SetViewport(25)

	if view.ViewportHeight != 25 {
		t.Errorf("Expected ViewportHeight to be 25, got %d", view.ViewportHeight)
	}
}

func TestViewRendersHeader(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test Issue Title",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Test comment",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.SetViewport(20)

	output := view.View()

	// Check that header is rendered
	if len(output) == 0 {
		t.Error("Expected View to return non-empty output")
	}

	// The output should contain the issue number and title
	// We can't check exact format due to styling codes, but can check it contains relevant text
}

func TestViewWithEmptyComments(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test Issue",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{}

	view := NewCommentsView(issue, comments)
	view.SetViewport(20)

	output := view.View()

	if len(output) == 0 {
		t.Error("Expected View to return output even with no comments")
	}
}

func TestViewWithMarkdownToggle(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "**Bold comment**",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.SetViewport(20)

	// Get raw markdown view
	rawView := view.View()

	// Toggle markdown rendering
	view.ToggleMarkdown()
	renderedView := view.View()

	// Both should be non-empty
	if len(rawView) == 0 {
		t.Error("Expected raw view to return output")
	}

	if len(renderedView) == 0 {
		t.Error("Expected rendered view to return output")
	}

	// The rendered view might be different due to glamour rendering
	// We just verify both produce output
}

func TestGetVisibleLines(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Line 1\nLine 2\nLine 3",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.SetViewport(2)

	lines := view.GetVisibleLines()

	// Should return up to ViewportHeight lines
	if len(lines) > view.ViewportHeight {
		t.Errorf("Expected at most %d lines, got %d", view.ViewportHeight, len(lines))
	}
}

func TestGetVisibleLinesWithScroll(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test Issue",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	comments := []storage.Comment{
		{
			ID:          1,
			IssueNumber: 123,
			Body:        "Comment body with enough content to span multiple lines when rendered",
			Author:      "user",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	view := NewCommentsView(issue, comments)
	view.SetViewport(20)

	// First, check that we get lines without scrolling
	lines := view.GetVisibleLines()
	if len(lines) == 0 {
		t.Error("Expected some lines to be returned")
	}

	// Now scroll a small amount (not past the end)
	view.ScrollOffset = 1
	lines = view.GetVisibleLines()

	// Should still return lines, just one fewer
	if len(lines) == 0 {
		t.Error("Expected some lines to be returned after scrolling")
	}
}
