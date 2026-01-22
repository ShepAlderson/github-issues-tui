package tui

import (
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/storage"
)

func TestNewDetailPanel(t *testing.T) {
	issue := storage.Issue{
		Number:    123,
		Title:     "Test Issue",
		Body:      "This is a test issue with **markdown**",
		Author:    "testuser",
		State:     "open",
		CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
		Comments:  5,
		Labels:    "bug,enhancement",
		Assignees: "user1,user2",
	}

	panel := NewDetailPanel(issue)

	if panel.Issue.Number != issue.Number {
		t.Errorf("Expected issue number %d, got %d", issue.Number, panel.Issue.Number)
	}

	if panel.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be false by default")
	}

	if panel.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to be 0, got %d", panel.ScrollOffset)
	}
}

func TestDetailPanelToggleMarkdown(t *testing.T) {
	issue := storage.Issue{
		Number: 123,
		Title:  "Test",
		Body:   "Test body",
		Author: "user",
		State:  "open",
	}

	panel := NewDetailPanel(issue)

	// Initial state should be false (raw markdown)
	if panel.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be false initially")
	}

	// Toggle to true
	panel.ToggleMarkdown()
	if !panel.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be true after toggle")
	}

	// Toggle back to false
	panel.ToggleMarkdown()
	if panel.RenderMarkdown {
		t.Error("Expected RenderMarkdown to be false after second toggle")
	}
}

func TestDetailPanelScroll(t *testing.T) {
	issue := storage.Issue{
		Number: 123,
		Title:  "Test",
		Body:   "Test body",
		Author: "user",
		State:  "open",
	}

	panel := NewDetailPanel(issue)

	// Test scrolling down
	panel.ScrollDown()
	if panel.ScrollOffset != 1 {
		t.Errorf("Expected ScrollOffset to be 1, got %d", panel.ScrollOffset)
	}

	panel.ScrollDown()
	if panel.ScrollOffset != 2 {
		t.Errorf("Expected ScrollOffset to be 2, got %d", panel.ScrollOffset)
	}

	// Test scrolling up
	panel.ScrollUp()
	if panel.ScrollOffset != 1 {
		t.Errorf("Expected ScrollOffset to be 1 after scroll up, got %d", panel.ScrollOffset)
	}

	// Test that offset doesn't go negative
	panel.ScrollUp()
	if panel.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to be 0, got %d", panel.ScrollOffset)
	}

	panel.ScrollUp()
	if panel.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to still be 0, got %d", panel.ScrollOffset)
	}
}

func TestDetailPanelSetViewport(t *testing.T) {
	issue := storage.Issue{
		Number: 123,
		Title:  "Test",
		Body:   "Test body",
		Author: "user",
		State:  "open",
	}

	panel := NewDetailPanel(issue)
	panel.SetViewport(20)

	if panel.ViewportHeight != 20 {
		t.Errorf("Expected ViewportHeight to be 20, got %d", panel.ViewportHeight)
	}
}

func TestDetailPanelWithEmptyIssue(t *testing.T) {
	issue := storage.Issue{
		Number: 0,
		Title:  "",
		Body:   "",
		Author: "",
		State:  "",
	}

	panel := NewDetailPanel(issue)

	if panel.Issue.Number != 0 {
		t.Errorf("Expected issue number 0, got %d", panel.Issue.Number)
	}

	// Should still be able to toggle
	panel.ToggleMarkdown()
	if !panel.RenderMarkdown {
		t.Error("Should be able to toggle markdown even with empty issue")
	}

	// Should still be able to scroll (offset stays at 0)
	panel.ScrollDown()
	if panel.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset to remain 0 for empty issue, got %d", panel.ScrollOffset)
	}
}

func TestDetailPanelWithLongBody(t *testing.T) {
	// Create a long body with multiple lines
	longBody := ""
	for i := 0; i < 100; i++ {
		longBody += "Line " + formatNumber(i) + "\n"
	}

	issue := storage.Issue{
		Number: 123,
		Title:  "Long Issue",
		Body:   longBody,
		Author: "user",
		State:  "open",
	}

	panel := NewDetailPanel(issue)

	// Scroll through multiple lines
	for i := 0; i < 50; i++ {
		panel.ScrollDown()
	}

	if panel.ScrollOffset != 50 {
		t.Errorf("Expected ScrollOffset to be 50, got %d", panel.ScrollOffset)
	}

	// Scroll back up
	for i := 0; i < 25; i++ {
		panel.ScrollUp()
	}

	if panel.ScrollOffset != 25 {
		t.Errorf("Expected ScrollOffset to be 25, got %d", panel.ScrollOffset)
	}
}
